package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/nahue/setlist_manager/internal/services"
	"github.com/nahue/setlist_manager/internal/store"
	"github.com/nahue/setlist_manager/templates"
)

// Handler handles song-related requests
type SongHandler struct {
	songsDB     *store.SQLiteSongsStore
	bandsDB     *store.SQLiteBandsStore
	authService *services.AuthService
}

// NewHandler creates a new songs handler
func NewSongHandler(songsDB *store.SQLiteSongsStore, bandsDB *store.SQLiteBandsStore, authService *services.AuthService) *SongHandler {
	return &SongHandler{
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
func (h *SongHandler) GetSongs(w http.ResponseWriter, r *http.Request) {
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
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to get songs", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the songs section
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}

// CreateSong handles POST /api/bands/songs
func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract form fields
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	key := r.FormValue("key")
	tempoStr := r.FormValue("tempo")
	notes := r.FormValue("notes")

	if title == "" {
		http.Error(w, "Song title is required", http.StatusBadRequest)
		return
	}

	// Parse tempo if provided
	var tempo *int
	if tempoStr != "" {
		if tempoVal, err := strconv.Atoi(tempoStr); err == nil {
			tempo = &tempoVal
		}
	}

	// Create song
	_, err = h.songsDB.CreateSong(bandID, title, artist, key, notes, user.ID, tempo)
	if err != nil {
		log.Printf("Error creating song: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to create song", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated songs list to return
	songs, err := h.songsDB.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting updated songs: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to get updated songs", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}

// ReorderSongs handles POST /api/bands/songs/reorder
func (h *SongHandler) ReorderSongs(w http.ResponseWriter, r *http.Request) {
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
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to reorder songs", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated songs list to return
	songs, err := h.songsDB.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting updated songs: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to get updated songs", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}

// DeleteSong handles DELETE /api/bands/songs/{songID}
func (h *SongHandler) DeleteSong(w http.ResponseWriter, r *http.Request) {
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
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to delete song", song.BandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated songs list to return
	songs, err := h.songsDB.GetSongsByBand(song.BandID)
	if err != nil {
		log.Printf("Error getting updated songs: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongsSectionError("Failed to get updated songs", song.BandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}
