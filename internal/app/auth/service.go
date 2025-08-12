package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/nahue/setlist_manager/internal/app/auth/database"
)

// AuthService handles authentication logic
type AuthService struct {
	db *database.Database
}

// NewAuthService creates a new auth service
func NewAuthService(db *database.Database) *AuthService {
	return &AuthService{
		db: db,
	}
}

// GenerateMagicLink generates a magic link for the given email
func (s *AuthService) GenerateMagicLink(email string) (string, error) {
	// Check if user exists, create if not
	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user
		user, err = s.db.CreateUser(email)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("Created new user: %s", email)
	}

	// Generate random token
	token := generateRandomToken()

	// Hash the token for storage
	tokenHash := hashToken(token)

	// Store magic link in database
	expiresAt := time.Now().Add(15 * time.Minute) // 15 minutes expiry
	_, err = s.db.CreateMagicLink(user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create magic link: %w", err)
	}

	return token, nil
}

// VerifyMagicLink verifies a magic link token and returns the user
func (s *AuthService) VerifyMagicLink(token string) (*database.User, error) {
	// Hash the token for comparison
	tokenHash := hashToken(token)

	// Find and validate magic link
	magicLink, err := s.db.GetMagicLinkByTokenHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get magic link: %w", err)
	}

	if magicLink == nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Check if token is expired
	if time.Now().After(magicLink.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	// Check if token has already been used
	if magicLink.UsedAt != nil {
		return nil, fmt.Errorf("token already used")
	}

	// Mark token as used
	err = s.db.MarkMagicLinkAsUsed(magicLink.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Get user
	user, err := s.db.GetUserByID(magicLink.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update last login
	err = s.db.UpdateUserLastLogin(user.ID)
	if err != nil {
		log.Printf("Warning: failed to update last login for user %s: %v", user.ID, err)
		// Don't fail the authentication for this
	}

	return user, nil
}

// CreateSession creates a new session for the user
func (s *AuthService) CreateSession(userID string) (string, error) {
	// Generate session token
	sessionToken := generateRandomToken()

	// Hash the session token for storage
	sessionTokenHash := hashToken(sessionToken)

	// Store session in database
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	_, err := s.db.CreateSession(userID, sessionTokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// GetUserFromSession gets the user from a session token
func (s *AuthService) GetUserFromSession(sessionToken string) (*database.User, error) {
	// Hash the session token for comparison
	sessionTokenHash := hashToken(sessionToken)

	// Find session
	session, err := s.db.GetSessionByToken(sessionTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Get user
	user, err := s.db.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// generateRandomToken generates a random token
func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// hashToken hashes a token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
