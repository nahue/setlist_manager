package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Database handles auth-related database operations
type SQLiteAuthStore struct {
	db *sql.DB
}

// NewDatabase creates a new auth database instance
func NewSQLiteAuthStore(db *sql.DB) *SQLiteAuthStore {
	return &SQLiteAuthStore{db: db}
}

// User represents a user in the system
type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  bool       `json:"is_active"`
}

// MagicLink represents a magic link for authentication
type MagicLink struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUser creates a new user
func (d *SQLiteAuthStore) CreateUser(email string) (*User, error) {
	userID := generateUUID()

	query := `INSERT INTO users (id, email) VALUES (?, ?)`
	_, err := d.db.Exec(query, userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:        userID,
		Email:     email,
		CreatedAt: time.Now(),
		IsActive:  true,
	}, nil
}

// GetUserByEmail gets a user by email
func (d *SQLiteAuthStore) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, email, created_at, last_login, is_active FROM users WHERE email = ?`

	var user User
	var lastLogin sql.NullTime

	err := d.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&lastLogin,
		&user.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// GetUserByID gets a user by ID
func (d *SQLiteAuthStore) GetUserByID(userID string) (*User, error) {
	query := `SELECT id, email, created_at, last_login, is_active FROM users WHERE id = ?`

	var user User
	var lastLogin sql.NullTime

	err := d.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&lastLogin,
		&user.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// UpdateUserLastLogin updates the user's last login time
func (d *SQLiteAuthStore) UpdateUserLastLogin(userID string) error {
	query := `UPDATE users SET last_login = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// CreateMagicLink creates a new magic link
func (d *SQLiteAuthStore) CreateMagicLink(userID, tokenHash string, expiresAt time.Time) (*MagicLink, error) {
	magicLinkID := generateUUID()

	query := `INSERT INTO magic_links (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, magicLinkID, userID, tokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create magic link: %w", err)
	}

	return &MagicLink{
		ID:        magicLinkID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// GetMagicLinkByTokenHash gets a magic link by token hash
func (d *SQLiteAuthStore) GetMagicLinkByTokenHash(tokenHash string) (*MagicLink, error) {
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM magic_links WHERE token_hash = ?`

	var magicLink MagicLink
	var usedAt sql.NullTime

	err := d.db.QueryRow(query, tokenHash).Scan(
		&magicLink.ID,
		&magicLink.UserID,
		&magicLink.TokenHash,
		&magicLink.ExpiresAt,
		&usedAt,
		&magicLink.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get magic link: %w", err)
	}

	if usedAt.Valid {
		magicLink.UsedAt = &usedAt.Time
	}

	return &magicLink, nil
}

// MarkMagicLinkAsUsed marks a magic link as used
func (d *SQLiteAuthStore) MarkMagicLinkAsUsed(magicLinkID string) error {
	query := `UPDATE magic_links SET used_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), magicLinkID)
	if err != nil {
		return fmt.Errorf("failed to mark magic link as used: %w", err)
	}
	return nil
}

// CleanupExpiredMagicLinks removes expired magic links
func (d *SQLiteAuthStore) CleanupExpiredMagicLinks() error {
	query := `DELETE FROM magic_links WHERE expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired magic links: %w", err)
	}
	return nil
}

// CreateSession creates a new session
func (d *SQLiteAuthStore) CreateSession(userID, sessionToken string, expiresAt time.Time) (*Session, error) {
	sessionID := generateUUID()

	query := `INSERT INTO sessions (id, user_id, session_token, expires_at) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, sessionID, userID, sessionToken, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:           sessionID,
		UserID:       userID,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}, nil
}

// GetSessionByToken gets a session by token
func (d *SQLiteAuthStore) GetSessionByToken(sessionToken string) (*Session, error) {
	query := `SELECT id, user_id, session_token, expires_at, created_at FROM sessions WHERE session_token = ?`

	var session Session

	err := d.db.QueryRow(query, sessionToken).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session
func (d *SQLiteAuthStore) DeleteSession(sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err := d.db.Exec(query, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (d *SQLiteAuthStore) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}
