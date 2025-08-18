package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Database handles song-related database operations
type SQLiteSongsStore struct {
	db *sql.DB
}

// NewDatabase creates a new songs database instance
func NewSQLiteSongsStore(db *sql.DB) *SQLiteSongsStore {
	return &SQLiteSongsStore{db: db}
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
	Content   string    `json:"content"`
	Position  int       `json:"position"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
	User      *User     `json:"user,omitempty"`
}

// CreateSong creates a new song
func (d *SQLiteSongsStore) CreateSong(bandID, title, artist, key, notes, content, createdBy string, tempo *int) (*Song, error) {
	songID := generateUUID()

	// Get the next position for this band
	var maxPosition int
	err := d.db.QueryRow("SELECT COALESCE(MAX(position), 0) FROM songs WHERE band_id = ? AND is_active = 1", bandID).Scan(&maxPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to get max position: %w", err)
	}
	nextPosition := maxPosition + 1

	query := `INSERT INTO songs (id, band_id, title, artist, key, tempo, notes, content, created_by, position) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = d.db.Exec(query, songID, bandID, title, artist, key, tempo, notes, content, createdBy, nextPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to create song: %w", err)
	}

	return &Song{
		ID:        songID,
		BandID:    bandID,
		Title:     title,
		Artist:    artist,
		Key:       key,
		Tempo:     tempo,
		Notes:     notes,
		Content:   content,
		Position:  nextPosition,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}, nil
}

// GetSongsByBand gets all songs for a band
func (d *SQLiteSongsStore) GetSongsByBand(bandID string) ([]*Song, error) {
	query := `
		SELECT s.id, s.band_id, s.title, s.artist, s.key, s.tempo, s.notes, s.content, s.position, s.created_by, s.created_at, s.updated_at, s.is_active,
		       u.id, u.email, u.created_at, u.last_login, u.is_active
		FROM songs s
		INNER JOIN users u ON s.created_by = u.id
		WHERE s.band_id = ? AND s.is_active = 1
		ORDER BY s.position ASC
	`

	rows, err := d.db.Query(query, bandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get songs: %w", err)
	}
	defer rows.Close()

	var songs []*Song
	for rows.Next() {
		var song Song
		var user User
		var lastLogin sql.NullTime
		var tempo sql.NullInt32
		var content sql.NullString

		err := rows.Scan(
			&song.ID,
			&song.BandID,
			&song.Title,
			&song.Artist,
			&song.Key,
			&tempo,
			&song.Notes,
			&content,
			&song.Position,
			&song.CreatedBy,
			&song.CreatedAt,
			&song.UpdatedAt,
			&song.IsActive,
			&user.ID,
			&user.Email,
			&user.CreatedAt,
			&lastLogin,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan song: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}
		if tempo.Valid {
			tempoInt := int(tempo.Int32)
			song.Tempo = &tempoInt
		}
		if content.Valid {
			song.Content = content.String
		}

		song.User = &user
		songs = append(songs, &song)
	}

	return songs, nil
}

// GetSongByID gets a song by ID
func (d *SQLiteSongsStore) GetSongByID(songID string) (*Song, error) {
	query := `
		SELECT s.id, s.band_id, s.title, s.artist, s.key, s.tempo, s.notes, s.content, s.position, s.created_by, s.created_at, s.updated_at, s.is_active
		FROM songs s
		WHERE s.id = ? AND s.is_active = 1
	`

	var song Song
	var tempo sql.NullInt32
	var content sql.NullString

	err := d.db.QueryRow(query, songID).Scan(
		&song.ID,
		&song.BandID,
		&song.Title,
		&song.Artist,
		&song.Key,
		&tempo,
		&song.Notes,
		&content,
		&song.Position,
		&song.CreatedBy,
		&song.CreatedAt,
		&song.UpdatedAt,
		&song.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if tempo.Valid {
		tempoInt := int(tempo.Int32)
		song.Tempo = &tempoInt
	}
	if content.Valid {
		song.Content = content.String
	}

	return &song, nil
}

// UpdateSong updates a song
func (d *SQLiteSongsStore) UpdateSong(songID, title, artist, key, notes, content string, tempo *int) error {
	query := `UPDATE songs SET title = ?, artist = ?, key = ?, tempo = ?, notes = ?, content = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, title, artist, key, tempo, notes, content, time.Now(), songID)
	if err != nil {
		return fmt.Errorf("failed to update song: %w", err)
	}
	return nil
}

// DeleteSong deletes a song (soft delete)
func (d *SQLiteSongsStore) DeleteSong(songID string) error {
	query := `UPDATE songs SET is_active = 0, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), songID)
	if err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}
	return nil
}

// ReorderSongs updates the positions of songs in a band
func (d *SQLiteSongsStore) ReorderSongs(bandID string, songOrder []string) error {
	// Start a transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update positions for each song
	for i, songID := range songOrder {
		_, err := tx.Exec("UPDATE songs SET position = ?, updated_at = ? WHERE id = ? AND band_id = ?",
			i+1, time.Now(), songID, bandID)
		if err != nil {
			return fmt.Errorf("failed to update song position: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
