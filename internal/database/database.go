package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  bool       `json:"is_active"`
}

type MagicLink struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Band struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

type BandMember struct {
	ID       string    `json:"id"`
	BandID   string    `json:"band_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
	User     *User     `json:"user,omitempty"`
}

type BandInvitation struct {
	ID            string     `json:"id"`
	BandID        string     `json:"band_id"`
	InvitedEmail  string     `json:"invited_email"`
	InvitedBy     string     `json:"invited_by"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	AcceptedAt    *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt    *time.Time `json:"declined_at,omitempty"`
	Band          *Band      `json:"band,omitempty"`
	InvitedByUser *User      `json:"invited_by_user,omitempty"`
}

func NewDatabase() (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", "./data/setlist_manager.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	log.Println("Database connected successfully")
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Ping() error {
	return d.db.Ping()
}

// User operations
func (d *Database) CreateUser(email string) (*User, error) {
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

func (d *Database) GetUserByEmail(email string) (*User, error) {
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

func (d *Database) GetUserByID(userID string) (*User, error) {
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

func (d *Database) UpdateUserLastLogin(userID string) error {
	query := `UPDATE users SET last_login = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Magic link operations
func (d *Database) CreateMagicLink(userID, tokenHash string, expiresAt time.Time) (*MagicLink, error) {
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

func (d *Database) GetMagicLinkByTokenHash(tokenHash string) (*MagicLink, error) {
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

func (d *Database) MarkMagicLinkAsUsed(magicLinkID string) error {
	query := `UPDATE magic_links SET used_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), magicLinkID)
	if err != nil {
		return fmt.Errorf("failed to mark magic link as used: %w", err)
	}
	return nil
}

func (d *Database) CleanupExpiredMagicLinks() error {
	query := `DELETE FROM magic_links WHERE expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired magic links: %w", err)
	}
	return nil
}

// Session operations
func (d *Database) CreateSession(userID, sessionToken string, expiresAt time.Time) (*Session, error) {
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

func (d *Database) GetSessionByToken(sessionToken string) (*Session, error) {
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

func (d *Database) DeleteSession(sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err := d.db.Exec(query, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (d *Database) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// Band operations
func (d *Database) CreateBand(name, description, createdBy string) (*Band, error) {
	bandID := generateUUID()

	query := `INSERT INTO bands (id, name, description, created_by) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, bandID, name, description, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create band: %w", err)
	}

	// Add the creator as the owner
	_, err = d.AddBandMember(bandID, createdBy, "owner")
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as band owner: %w", err)
	}

	return &Band{
		ID:          bandID,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}, nil
}

func (d *Database) GetBandByID(bandID string) (*Band, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at, is_active FROM bands WHERE id = ?`

	var band Band
	err := d.db.QueryRow(query, bandID).Scan(
		&band.ID,
		&band.Name,
		&band.Description,
		&band.CreatedBy,
		&band.CreatedAt,
		&band.UpdatedAt,
		&band.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get band: %w", err)
	}

	return &band, nil
}

func (d *Database) GetBandsByUser(userID string) ([]*Band, error) {
	query := `
		SELECT b.id, b.name, b.description, b.created_by, b.created_at, b.updated_at, b.is_active 
		FROM bands b
		INNER JOIN band_members bm ON b.id = bm.band_id
		WHERE bm.user_id = ? AND bm.is_active = 1 AND b.is_active = 1
		ORDER BY b.updated_at DESC
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bands: %w", err)
	}
	defer rows.Close()

	var bands []*Band
	for rows.Next() {
		var band Band
		err := rows.Scan(
			&band.ID,
			&band.Name,
			&band.Description,
			&band.CreatedBy,
			&band.CreatedAt,
			&band.UpdatedAt,
			&band.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan band: %w", err)
		}
		bands = append(bands, &band)
	}

	return bands, nil
}

// Band member operations
func (d *Database) AddBandMember(bandID, userID, role string) (*BandMember, error) {
	memberID := generateUUID()

	query := `INSERT INTO band_members (id, band_id, user_id, role) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, memberID, bandID, userID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to add band member: %w", err)
	}

	return &BandMember{
		ID:       memberID,
		BandID:   bandID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
		IsActive: true,
	}, nil
}

func (d *Database) GetBandMembers(bandID string) ([]*BandMember, error) {
	query := `
		SELECT bm.id, bm.band_id, bm.user_id, bm.role, bm.joined_at, bm.is_active,
		       u.id, u.email, u.created_at, u.last_login, u.is_active
		FROM band_members bm
		INNER JOIN users u ON bm.user_id = u.id
		WHERE bm.band_id = ? AND bm.is_active = 1
		ORDER BY bm.joined_at ASC
	`

	rows, err := d.db.Query(query, bandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get band members: %w", err)
	}
	defer rows.Close()

	var members []*BandMember
	for rows.Next() {
		var member BandMember
		var user User
		var lastLogin sql.NullTime

		err := rows.Scan(
			&member.ID,
			&member.BandID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.IsActive,
			&user.ID,
			&user.Email,
			&user.CreatedAt,
			&lastLogin,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan band member: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		member.User = &user
		members = append(members, &member)
	}

	return members, nil
}

func (d *Database) GetBandMember(bandID, userID string) (*BandMember, error) {
	query := `
		SELECT id, band_id, user_id, role, joined_at, is_active
		FROM band_members
		WHERE band_id = ? AND user_id = ? AND is_active = 1
	`

	var member BandMember
	err := d.db.QueryRow(query, bandID, userID).Scan(
		&member.ID,
		&member.BandID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get band member: %w", err)
	}

	return &member, nil
}

// Band invitation operations
func (d *Database) CreateBandInvitation(bandID, invitedEmail, invitedBy, role string, expiresAt time.Time) (*BandInvitation, error) {
	invitationID := generateUUID()

	query := `INSERT INTO band_invitations (id, band_id, invited_email, invited_by, role, expires_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, invitationID, bandID, invitedEmail, invitedBy, role, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create band invitation: %w", err)
	}

	return &BandInvitation{
		ID:           invitationID,
		BandID:       bandID,
		InvitedEmail: invitedEmail,
		InvitedBy:    invitedBy,
		Role:         role,
		Status:       "pending",
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}, nil
}

func (d *Database) GetBandInvitationByID(invitationID string) (*BandInvitation, error) {
	query := `
		SELECT bi.id, bi.band_id, bi.invited_email, bi.invited_by, bi.role, bi.status, 
		       bi.expires_at, bi.created_at, bi.accepted_at, bi.declined_at,
		       b.name, b.description,
		       u.email
		FROM band_invitations bi
		INNER JOIN bands b ON bi.band_id = b.id
		INNER JOIN users u ON bi.invited_by = u.id
		WHERE bi.id = ?
	`

	var invitation BandInvitation
	var band Band
	var invitedByUser User
	var acceptedAt, declinedAt sql.NullTime

	err := d.db.QueryRow(query, invitationID).Scan(
		&invitation.ID,
		&invitation.BandID,
		&invitation.InvitedEmail,
		&invitation.InvitedBy,
		&invitation.Role,
		&invitation.Status,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
		&acceptedAt,
		&declinedAt,
		&band.Name,
		&band.Description,
		&invitedByUser.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get band invitation: %w", err)
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	if declinedAt.Valid {
		invitation.DeclinedAt = &declinedAt.Time
	}

	invitation.Band = &band
	invitation.InvitedByUser = &invitedByUser

	return &invitation, nil
}

func (d *Database) GetPendingInvitationsByEmail(email string) ([]*BandInvitation, error) {
	query := `
		SELECT bi.id, bi.band_id, bi.invited_email, bi.invited_by, bi.role, bi.status, 
		       bi.expires_at, bi.created_at, bi.accepted_at, bi.declined_at,
		       b.name, b.description,
		       u.email
		FROM band_invitations bi
		INNER JOIN bands b ON bi.band_id = b.id
		INNER JOIN users u ON bi.invited_by = u.id
		WHERE bi.invited_email = ? AND bi.status = 'pending' AND bi.expires_at > ?
		ORDER BY bi.created_at DESC
	`

	rows, err := d.db.Query(query, email, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*BandInvitation
	for rows.Next() {
		var invitation BandInvitation
		var band Band
		var invitedByUser User
		var acceptedAt, declinedAt sql.NullTime

		err := rows.Scan(
			&invitation.ID,
			&invitation.BandID,
			&invitation.InvitedEmail,
			&invitation.InvitedBy,
			&invitation.Role,
			&invitation.Status,
			&invitation.ExpiresAt,
			&invitation.CreatedAt,
			&acceptedAt,
			&declinedAt,
			&band.Name,
			&band.Description,
			&invitedByUser.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		if acceptedAt.Valid {
			invitation.AcceptedAt = &acceptedAt.Time
		}
		if declinedAt.Valid {
			invitation.DeclinedAt = &declinedAt.Time
		}

		invitation.Band = &band
		invitation.InvitedByUser = &invitedByUser
		invitations = append(invitations, &invitation)
	}

	return invitations, nil
}

func (d *Database) AcceptBandInvitation(invitationID, userID string) error {
	// Get the invitation
	invitation, err := d.GetBandInvitationByID(invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err)
	}
	if invitation == nil {
		return fmt.Errorf("invitation not found")
	}

	// Check if invitation is still valid
	if invitation.Status != "pending" || time.Now().After(invitation.ExpiresAt) {
		return fmt.Errorf("invitation is no longer valid")
	}

	// Start a transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update invitation status
	_, err = tx.Exec("UPDATE band_invitations SET status = 'accepted', accepted_at = ? WHERE id = ?", time.Now(), invitationID)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Add user to band
	_, err = tx.Exec("INSERT INTO band_members (id, band_id, user_id, role) VALUES (?, ?, ?, ?)",
		generateUUID(), invitation.BandID, userID, invitation.Role)
	if err != nil {
		return fmt.Errorf("failed to add band member: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *Database) DeclineBandInvitation(invitationID string) error {
	query := `UPDATE band_invitations SET status = 'declined', declined_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), invitationID)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}
	return nil
}

func (d *Database) CleanupExpiredInvitations() error {
	query := `UPDATE band_invitations SET status = 'expired' WHERE status = 'pending' AND expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired invitations: %w", err)
	}
	return nil
}

// Helper function to generate UUID (simplified for SQLite)
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
