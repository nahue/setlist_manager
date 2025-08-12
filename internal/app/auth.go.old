package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nahue/setlist_manager/internal/database"
)

// Context key type for user data
type contextKey string

const userContextKey contextKey = "user"

// AuthService handles authentication logic
type AuthService struct {
	db *database.Database
}

// NewAuthService creates a new authentication service
func NewAuthService(db *database.Database) *AuthService {
	return &AuthService{db: db}
}

// Request/Response types
type MagicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type MagicLinkResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type AuthUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// Magic link generation and verification
func (a *AuthService) GenerateMagicLink(email string) (string, error) {
	// Validate email format
	if !isValidEmail(email) {
		return "", fmt.Errorf("invalid email format")
	}

	// Get or create user
	user, err := a.db.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user
		user, err = a.db.CreateUser(email)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("Created new user: %s", email)
	}

	if !user.IsActive {
		return "", fmt.Errorf("user account is inactive")
	}

	// Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash token for storage
	tokenHash := hashToken(token)

	// Store magic link with expiration (15 minutes)
	expiresAt := time.Now().Add(15 * time.Minute)
	_, err = a.db.CreateMagicLink(user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create magic link: %w", err)
	}

	log.Printf("Generated magic link for user: %s", email)
	return token, nil
}

func (a *AuthService) VerifyMagicLink(token string) (*AuthUser, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Hash the provided token
	tokenHash := hashToken(token)

	// Get magic link from database
	magicLink, err := a.db.GetMagicLinkByTokenHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get magic link: %w", err)
	}

	if magicLink == nil {
		return nil, fmt.Errorf("invalid or expired magic link")
	}

	// Check if already used
	if magicLink.UsedAt != nil {
		return nil, fmt.Errorf("magic link has already been used")
	}

	// Check if expired
	if time.Now().After(magicLink.ExpiresAt) {
		return nil, fmt.Errorf("magic link has expired")
	}

	// Get user by ID
	user, err := a.db.GetUserByID(magicLink.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark magic link as used
	err = a.db.MarkMagicLinkAsUsed(magicLink.ID)
	if err != nil {
		log.Printf("Warning: failed to mark magic link as used: %v", err)
	}

	// Update user last login
	err = a.db.UpdateUserLastLogin(magicLink.UserID)
	if err != nil {
		log.Printf("Warning: failed to update last login: %v", err)
	}

	return &AuthUser{
		ID:       user.ID,
		Email:    user.Email,
		IsActive: user.IsActive,
	}, nil
}

// Session management
func (a *AuthService) CreateSession(userID string) (string, error) {
	// Generate session token
	sessionToken, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Session expires in 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Store session
	_, err = a.db.CreateSession(userID, sessionToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

func (a *AuthService) VerifySession(sessionToken string) (*AuthUser, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("session token is required")
	}

	// Get session from database
	session, err := a.db.GetSessionByToken(sessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("invalid session")
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		_ = a.db.DeleteSession(sessionToken)
		return nil, fmt.Errorf("session has expired")
	}

	// Get user data
	user, err := a.db.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Clean up invalid session
		_ = a.db.DeleteSession(sessionToken)
		return nil, fmt.Errorf("user not found")
	}

	return &AuthUser{
		ID:       user.ID,
		Email:    user.Email,
		IsActive: user.IsActive,
	}, nil
}

func (a *AuthService) DeleteSession(sessionToken string) error {
	return a.db.DeleteSession(sessionToken)
}

// Middleware
func (app *Application) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public routes
		if app.isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// If authentication is disabled, create a mock user for development
		if !app.useAuth {
			log.Println("Authentication is disabled, creating mock user")
			mockUser := &AuthUser{
				ID:       "dev-user",
				Email:    "dev@example.com",
				IsActive: true,
			}
			ctx := context.WithValue(r.Context(), userContextKey, mockUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Get current user
		user := app.getCurrentUser(r)
		if user == nil {
			// Redirect to login for browser requests
			if strings.Contains(r.Header.Get("Accept"), "text/html") {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}
			// Return 401 for API requests
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions
func (app *Application) getCurrentUser(r *http.Request) *AuthUser {
	// If authentication is disabled, return a mock user
	if !app.useAuth {
		return &AuthUser{
			ID:       "dev-user",
			Email:    "dev@example.com",
			IsActive: true,
		}
	}

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	authService := NewAuthService(app.db)
	user, err := authService.VerifySession(cookie.Value)
	if err != nil {
		return nil
	}

	return user
}

func (app *Application) isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/auth/login",
		"/auth/magic-link",
		"/auth/verify",
		"/health",
		"/public/",
	}

	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func isValidEmail(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// GetUserFromContext extracts the authenticated user from the request context
func GetUserFromContext(ctx context.Context) *AuthUser {
	if user, ok := ctx.Value(userContextKey).(*AuthUser); ok {
		return user
	}
	return nil
}
