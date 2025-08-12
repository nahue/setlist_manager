package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nahue/setlist_manager/internal/app/shared/types"
)

// Database handles band-related database operations
type SQLiteBandsStore struct {
	db *sql.DB
}

// NewDatabase creates a new bands database instance
func NewSQLiteBandsStore(db *sql.DB) *SQLiteBandsStore {
	return &SQLiteBandsStore{db: db}
}

// Band represents a band
type Band struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// BandMember represents a band member
type BandMember struct {
	ID       string    `json:"id"`
	BandID   string    `json:"band_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
	User     *User     `json:"user,omitempty"`
}

// BandInvitation represents a band invitation
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

// CreateBand creates a new band
func (d *SQLiteBandsStore) CreateBand(name, description, createdBy string) (*Band, error) {
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

// GetBandByID gets a band by ID
func (d *SQLiteBandsStore) GetBandByID(bandID string) (*Band, error) {
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

// GetBandsByUser gets all bands for a user
func (d *SQLiteBandsStore) GetBandsByUser(userID string) ([]*Band, error) {
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

// AddBandMember adds a member to a band
func (d *SQLiteBandsStore) AddBandMember(bandID, userID, role string) (*BandMember, error) {
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

// GetBandMembers gets all members of a band
func (d *SQLiteBandsStore) GetBandMembers(bandID string) ([]*BandMember, error) {
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

// GetBandMember gets a specific band member
func (d *SQLiteBandsStore) GetBandMember(bandID, userID string) (*BandMember, error) {
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

// RemoveBandMember removes a member from a band
func (d *SQLiteBandsStore) RemoveBandMember(bandID, userID string) error {
	query := `DELETE FROM band_members WHERE band_id = ? AND user_id = ?`
	_, err := d.db.Exec(query, bandID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove band member: %w", err)
	}
	return nil
}

// GetUserByEmail gets a user by email
func (d *SQLiteBandsStore) GetUserByEmail(email string) (*User, error) {
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

// CreateBandInvitation creates a new band invitation
func (d *SQLiteBandsStore) CreateBandInvitation(bandID, invitedEmail, invitedBy, role string, expiresAt time.Time) (*BandInvitation, error) {
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

// GetBandInvitationByID gets a band invitation by ID
func (d *SQLiteBandsStore) GetBandInvitationByID(invitationID string) (*BandInvitation, error) {
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

// GetPendingInvitationsByEmail gets pending invitations for a user
func (d *SQLiteBandsStore) GetPendingInvitationsByEmail(email string) ([]*BandInvitation, error) {
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

// AcceptBandInvitation accepts a band invitation
func (d *SQLiteBandsStore) AcceptBandInvitation(invitationID, userID string) error {
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

// DeclineBandInvitation declines a band invitation
func (d *SQLiteBandsStore) DeclineBandInvitation(invitationID string) error {
	query := `UPDATE band_invitations SET status = 'declined', declined_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), invitationID)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}
	return nil
}

// CleanupExpiredInvitations marks expired invitations as expired
func (d *SQLiteBandsStore) CleanupExpiredInvitations() error {
	query := `UPDATE band_invitations SET status = 'expired' WHERE status = 'pending' AND expires_at < ?`
	_, err := d.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired invitations: %w", err)
	}
	return nil
}

// Convert database types to shared types
func (d *SQLiteBandsStore) GetBandMembersShared(bandID string) ([]*types.BandMember, error) {
	members, err := d.GetBandMembers(bandID)
	if err != nil {
		return nil, err
	}

	var sharedMembers []*types.BandMember
	for _, member := range members {
		sharedMember := &types.BandMember{
			ID:       member.ID,
			BandID:   member.BandID,
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
			IsActive: member.IsActive,
		}
		if member.User != nil {
			sharedMember.User = &types.User{
				ID:        member.User.ID,
				Email:     member.User.Email,
				CreatedAt: member.User.CreatedAt,
				LastLogin: member.User.LastLogin,
				IsActive:  member.User.IsActive,
			}
		}
		sharedMembers = append(sharedMembers, sharedMember)
	}

	return sharedMembers, nil
}

func (d *SQLiteBandsStore) GetBandsByUserShared(userID string) ([]*types.Band, error) {
	bands, err := d.GetBandsByUser(userID)
	if err != nil {
		return nil, err
	}

	var sharedBands []*types.Band
	for _, band := range bands {
		sharedBand := &types.Band{
			ID:          band.ID,
			Name:        band.Name,
			Description: band.Description,
			CreatedBy:   band.CreatedBy,
			CreatedAt:   band.CreatedAt,
			UpdatedAt:   band.UpdatedAt,
			IsActive:    band.IsActive,
		}
		sharedBands = append(sharedBands, sharedBand)
	}

	return sharedBands, nil
}

func (d *SQLiteBandsStore) GetBandByIDShared(bandID string) (*types.Band, error) {
	band, err := d.GetBandByID(bandID)
	if err != nil {
		return nil, err
	}
	if band == nil {
		return nil, nil
	}

	return &types.Band{
		ID:          band.ID,
		Name:        band.Name,
		Description: band.Description,
		CreatedBy:   band.CreatedBy,
		CreatedAt:   band.CreatedAt,
		UpdatedAt:   band.UpdatedAt,
		IsActive:    band.IsActive,
	}, nil
}
