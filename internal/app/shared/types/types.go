package types

import "time"

// User represents a user in the system
type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  bool       `json:"is_active"`
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

// Song represents a song
type Song struct {
	ID        string    `json:"id"`
	BandID    string    `json:"band_id"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Key       string    `json:"key"`
	Tempo     *int      `json:"tempo,omitempty"`
	Notes     string    `json:"notes"`
	Position  int       `json:"position"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
	User      *User     `json:"user,omitempty"`
}
