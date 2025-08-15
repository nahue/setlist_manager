package store

import (
	"database/sql"
	"fmt"
	"time"
)

// SQLiteSongSectionsStore handles song section-related database operations
type SQLiteSongSectionsStore struct {
	db *sql.DB
}

// NewSQLiteSongSectionsStore creates a new song sections store instance
func NewSQLiteSongSectionsStore(db *sql.DB) *SQLiteSongSectionsStore {
	return &SQLiteSongSectionsStore{db: db}
}

// SongSection represents a song section
type SongSection struct {
	ID        string    `json:"id"`
	SongID    string    `json:"song_id"`
	Title     string    `json:"title"`
	Key       string    `json:"key"`
	Body      string    `json:"body"`
	Position  int       `json:"position"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
	User      *User     `json:"user,omitempty"`
}

// CreateSongSection creates a new song section
func (s *SQLiteSongSectionsStore) CreateSongSection(songID, title, key, body, createdBy string) (*SongSection, error) {
	sectionID := generateUUID()

	// Get the next position for this song
	var maxPosition int
	err := s.db.QueryRow("SELECT COALESCE(MAX(position), 0) FROM song_sections WHERE song_id = ? AND is_active = 1", songID).Scan(&maxPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to get max position: %w", err)
	}
	nextPosition := maxPosition + 1

	query := `INSERT INTO song_sections (id, song_id, title, key, body, position, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, sectionID, songID, title, key, body, nextPosition, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create song section: %w", err)
	}

	return &SongSection{
		ID:        sectionID,
		SongID:    songID,
		Title:     title,
		Key:       key,
		Body:      body,
		Position:  nextPosition,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}, nil
}

// GetSongSectionsBySongID gets all sections for a song
func (s *SQLiteSongSectionsStore) GetSongSectionsBySongID(songID string) ([]*SongSection, error) {
	query := `
		SELECT ss.id, ss.song_id, ss.title, ss.key, ss.body, ss.position, ss.created_by, ss.created_at, ss.updated_at, ss.is_active,
		       u.id, u.email, u.created_at, u.last_login, u.is_active
		FROM song_sections ss
		INNER JOIN users u ON ss.created_by = u.id
		WHERE ss.song_id = ? AND ss.is_active = 1
		ORDER BY ss.position ASC
	`

	rows, err := s.db.Query(query, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song sections: %w", err)
	}
	defer rows.Close()

	var sections []*SongSection
	for rows.Next() {
		var section SongSection
		var user User
		var lastLogin sql.NullTime

		err := rows.Scan(
			&section.ID,
			&section.SongID,
			&section.Title,
			&section.Key,
			&section.Body,
			&section.Position,
			&section.CreatedBy,
			&section.CreatedAt,
			&section.UpdatedAt,
			&section.IsActive,
			&user.ID,
			&user.Email,
			&user.CreatedAt,
			&lastLogin,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan song section: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		section.User = &user
		sections = append(sections, &section)
	}

	return sections, nil
}

// GetSongSectionByID gets a song section by ID
func (s *SQLiteSongSectionsStore) GetSongSectionByID(sectionID string) (*SongSection, error) {
	query := `
		SELECT id, song_id, title, key, body, position, created_by, created_at, updated_at, is_active
		FROM song_sections
		WHERE id = ? AND is_active = 1
	`

	var section SongSection
	err := s.db.QueryRow(query, sectionID).Scan(
		&section.ID,
		&section.SongID,
		&section.Title,
		&section.Key,
		&section.Body,
		&section.Position,
		&section.CreatedBy,
		&section.CreatedAt,
		&section.UpdatedAt,
		&section.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get song section: %w", err)
	}

	return &section, nil
}

// UpdateSongSection updates a song section
func (s *SQLiteSongSectionsStore) UpdateSongSection(sectionID, title, key, body string) error {
	query := `UPDATE song_sections SET title = ?, key = ?, body = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, title, key, body, time.Now(), sectionID)
	if err != nil {
		return fmt.Errorf("failed to update song section: %w", err)
	}
	return nil
}

// DeleteSongSection deletes a song section (soft delete)
func (s *SQLiteSongSectionsStore) DeleteSongSection(sectionID string) error {
	query := `UPDATE song_sections SET is_active = 0, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now(), sectionID)
	if err != nil {
		return fmt.Errorf("failed to delete song section: %w", err)
	}
	return nil
}

// ReorderSongSections updates the positions of sections in a song
func (s *SQLiteSongSectionsStore) ReorderSongSections(songID string, sectionOrder []string) error {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update positions for each section
	for i, sectionID := range sectionOrder {
		_, err := tx.Exec("UPDATE song_sections SET position = ?, updated_at = ? WHERE id = ? AND song_id = ?",
			i+1, time.Now(), sectionID, songID)
		if err != nil {
			return fmt.Errorf("failed to update section position: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
