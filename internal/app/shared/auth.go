package shared

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/nahue/setlist_manager/internal/app/auth/database"
	"github.com/nahue/setlist_manager/internal/app/shared/types"
)

// AuthService provides authentication utilities
type AuthService struct {
	db *database.Database
}

// NewAuthService creates a new auth service
func NewAuthService(db *database.Database) *AuthService {
	return &AuthService{
		db: db,
	}
}

// GetCurrentUser gets the current user from the session
func (s *AuthService) GetCurrentUser(r *http.Request) *types.User {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	// Hash the session token for comparison
	sessionTokenHash := hashToken(cookie.Value)

	// Find session
	session, err := s.db.GetSessionByToken(sessionTokenHash)
	if err != nil {
		return nil
	}

	if session == nil {
		return nil
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil
	}

	// Get user
	user, err := s.db.GetUserByID(session.UserID)
	if err != nil {
		return nil
	}

	// Convert database.User to types.User
	return &types.User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		IsActive:  user.IsActive,
	}
}

// hashToken hashes a token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
