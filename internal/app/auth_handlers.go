package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nahue/setlist_manager/templates"
)

// handleMagicLinkRequest handles POST /auth/magic-link
func (app *Application) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var req MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Rate limiting would go here in production

	// Generate magic link
	authService := NewAuthService(app.db)
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

// handleMagicLinkVerification handles GET /auth/verify
func (app *Application) handleMagicLinkVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	authService := NewAuthService(app.db)

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
	bands, err := app.db.GetBandsByUser(user.ID)
	if err != nil {
		log.Printf("Error checking user bands: %v", err)
		// Continue anyway, don't fail the login
	} else if len(bands) == 0 {
		// Create a default band for the user
		defaultBandName := "My Band"
		defaultBandDescription := "Your personal band for managing songs and setlists"

		band, err := app.db.CreateBand(defaultBandName, defaultBandDescription, user.ID)
		if err != nil {
			log.Printf("Error creating default band: %v", err)
			// Continue anyway, don't fail the login
		} else {
			log.Printf("Created default band '%s' for user: %s", band.Name, user.Email)
		}
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleLogout handles POST /auth/logout
func (app *Application) handleLogout(w http.ResponseWriter, r *http.Request) {
	// If authentication is disabled, redirect to main page
	if !app.useAuth {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		authService := NewAuthService(app.db)
		// Delete session from database
		_ = authService.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
	})

	// Redirect to login
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// handleLogin handles GET /auth/login
func (app *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	// If authentication is disabled, redirect to main page
	if !app.useAuth {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Check if already authenticated
	if user := app.getCurrentUser(r); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get error message if any
	errorMsg := r.URL.Query().Get("error")

	// Render login page
	component := templates.LoginPage(errorMsg)
	component.Render(r.Context(), w)
}

// handleCurrentUser handles GET /auth/me
func (app *Application) handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
