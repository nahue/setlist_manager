package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nahue/setlist_manager/internal/app/auth/database"
	bands "github.com/nahue/setlist_manager/internal/app/bands/database"
	"github.com/nahue/setlist_manager/templates"
)

// Handler handles authentication-related requests
type Handler struct {
	authDB  *database.Database
	bandsDB *bands.Database
}

// NewHandler creates a new auth handler
func NewHandler(authDB *database.Database, bandsDB *bands.Database) *Handler {
	return &Handler{
		authDB:  authDB,
		bandsDB: bandsDB,
	}
}

// MagicLinkRequest represents a magic link request
type MagicLinkRequest struct {
	Email string `json:"email"`
}

// MagicLinkResponse represents a magic link response
type MagicLinkResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// HandleLogin handles GET /auth/login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	errorMsg := r.URL.Query().Get("error")
	component := templates.LoginPage(errorMsg)
	component.Render(r.Context(), w)
}

// HandleMagicLinkRequest handles POST /auth/magic-link
func (h *Handler) HandleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var req MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Rate limiting would go here in production

	// Generate magic link
	authService := NewAuthService(h.authDB)
	token, err := authService.GenerateMagicLink(req.Email)
	if err != nil {
		log.Printf("Failed to generate magic link: %v", err)
		http.Error(w, "Failed to send magic link", http.StatusInternalServerError)
		return
	}

	// In production, send email with magic link
	// For now, we'll log it (NEVER do this in production)
	baseURL := getBaseURL(r)
	magicLink := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)
	log.Printf("Magic link for %s: %s", req.Email, magicLink)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MagicLinkResponse{
		Message: "Magic link sent to your email",
		Success: true,
	})
}

// HandleMagicLinkVerification handles GET /auth/verify
func (h *Handler) HandleMagicLinkVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	authService := NewAuthService(h.authDB)

	// Verify magic link
	user, err := authService.VerifyMagicLink(token)
	if err != nil {
		log.Printf("Magic link verification failed: %v", err)
		// Redirect to login with error
		http.Redirect(w, r, "/auth/login?error=invalid_token", http.StatusSeeOther)
		return
	}

	// Create session
	sessionToken, err := authService.CreateSession(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Only secure in HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	log.Printf("User authenticated successfully: %s", user.ID)

	// Check if user has any bands, create default band if not
	bands, err := h.bandsDB.GetBandsByUser(user.ID)
	if err != nil {
		log.Printf("Error checking user bands: %v", err)
		// Continue anyway, don't fail the login
	} else if len(bands) == 0 {
		// Create a default band for the user
		defaultBandName := "My Band"
		defaultBandDescription := "Your personal band for managing songs and setlists"

		band, err := h.bandsDB.CreateBand(defaultBandName, defaultBandDescription, user.ID)
		if err != nil {
			log.Printf("Error creating default band: %v", err)
			// Continue anyway, don't fail the login
		} else {
			log.Printf("Created default band '%s' for user: %s", band.Name, user.Email)
		}
	}

	// Redirect to bands page
	http.Redirect(w, r, "/bands", http.StatusSeeOther)
}

// HandleLogout handles POST /auth/logout
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete the cookie
	})

	// Redirect to login page
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// HandleCurrentUser handles GET /auth/me
func (h *Handler) HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

// getCurrentUser gets the current user from the session
func (h *Handler) getCurrentUser(r *http.Request) *database.User {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	authService := NewAuthService(h.authDB)
	user, err := authService.GetUserFromSession(cookie.Value)
	if err != nil {
		return nil
	}

	return user
}

// getBaseURL gets the base URL for the application
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
