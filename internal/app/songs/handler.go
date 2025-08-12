package songs

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	bands "github.com/nahue/setlist_manager/internal/app/bands/database"
	"github.com/nahue/setlist_manager/internal/app/shared"
	"github.com/nahue/setlist_manager/internal/store"
)

// Handler handles song-related requests
type Handler struct {
	songsDB     *store.SQLiteSongsStore
	bandsDB     *bands.Database
	authService *shared.AuthService
}

// NewHandler creates a new songs handler
func NewHandler(songsDB *store.SQLiteSongsStore, bandsDB *bands.Database, authService *shared.AuthService) *Handler {
	return &Handler{
		songsDB:     songsDB,
		bandsDB:     bandsDB,
		authService: authService,
	}
}

// Request/Response structs
type CreateSongRequest struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Key    string `json:"key"`
	Tempo  *int   `json:"tempo"`
	Notes  string `json:"notes"`
}

type ReorderSongsRequest struct {
	SongOrder []string `json:"song_order"`
}

// GetSongs handles GET /api/bands/songs
func (h *Handler) GetSongs(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := h.bandsDB.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get songs for the band
	songs, err := h.songsDB.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting songs: %v", err)
		http.Error(w, "Failed to get songs", http.StatusInternalServerError)
		return
	}

	// Return songs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"songs":   songs,
	})
}

// CreateSong handles POST /api/bands/songs
func (h *Handler) CreateSong(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := h.bandsDB.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Song title is required", http.StatusBadRequest)
		return
	}

	// Create song
	song, err := h.songsDB.CreateSong(bandID, req.Title, req.Artist, req.Key, req.Notes, user.ID, req.Tempo)
	if err != nil {
		log.Printf("Error creating song: %v", err)
		http.Error(w, "Failed to create song", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"song":    song,
	})
}

// ReorderSongs handles POST /api/bands/songs/reorder
func (h *Handler) ReorderSongs(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := h.bandsDB.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req ReorderSongsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reorder songs
	err = h.songsDB.ReorderSongs(bandID, req.SongOrder)
	if err != nil {
		log.Printf("Error reordering songs: %v", err)
		http.Error(w, "Failed to reorder songs", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Songs reordered successfully",
	})
}

// DeleteSong handles DELETE /api/bands/songs/{songID}
func (h *Handler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}
	songID := pathParts[len(pathParts)-1]

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get song to check band membership
	song, err := h.songsDB.GetSongByID(songID)
	if err != nil {
		log.Printf("Error getting song: %v", err)
		http.Error(w, "Failed to get song", http.StatusInternalServerError)
		return
	}
	if song == nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	// Check if user is a member of the band
	member, err := h.bandsDB.GetBandMember(song.BandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete song
	err = h.songsDB.DeleteSong(songID)
	if err != nil {
		log.Printf("Error deleting song: %v", err)
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Song deleted successfully",
	})
}
