package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/nahue/setlist_manager/internal/app/shared/types"
	"github.com/nahue/setlist_manager/internal/services"
	"github.com/nahue/setlist_manager/internal/store"
	"github.com/nahue/setlist_manager/templates"
)

// Handler handles song-related requests
type SongHandler struct {
	songsDB         *store.SQLiteSongsStore
	bandsDB         *store.SQLiteBandsStore
	authService     *services.AuthService
	authStore       *store.SQLiteAuthStore
	markdownService *services.MarkdownService
	aiService       *services.AIService
	pdfService      *services.PDFService
}

// NewHandler creates a new songs handler
func NewSongHandler(songsDB *store.SQLiteSongsStore, bandsDB *store.SQLiteBandsStore, authService *services.AuthService, authStore *store.SQLiteAuthStore, markdownService *services.MarkdownService, aiService *services.AIService, pdfService *services.PDFService) *SongHandler {
	return &SongHandler{
		songsDB:         songsDB,
		bandsDB:         bandsDB,
		authService:     authService,
		authStore:       authStore,
		markdownService: markdownService,
		aiService:       aiService,
		pdfService:      pdfService,
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
	content := r.FormValue("content")

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
	_, err = h.songsDB.CreateSong(bandID, title, artist, key, notes, content, user.ID, tempo)
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

// ServeSongDetails handles GET /song
func (h *SongHandler) ServeSongDetails(w http.ResponseWriter, r *http.Request) {
	songID := r.URL.Query().Get("id")
	if songID == "" {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get song details
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

	// Get band details
	band, err := h.bandsDB.GetBandByID(song.BandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}
	if band == nil {
		http.Error(w, "Band not found", http.StatusNotFound)
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

	// Get user info for the song creator
	if song.User == nil {
		userInfo, err := h.authStore.GetUserByID(song.CreatedBy)
		if err == nil && userInfo != nil {
			song.User = userInfo
		}
	}

	// Convert store.Band to types.Band
	bandType := &types.Band{
		ID:          band.ID,
		Name:        band.Name,
		Description: band.Description,
		CreatedBy:   band.CreatedBy,
		CreatedAt:   band.CreatedAt,
		UpdatedAt:   band.UpdatedAt,
		IsActive:    band.IsActive,
	}

	// Store original markdown content for editing
	originalMarkdown := song.Content

	// Process song content to convert markdown to HTML for display
	if song.Content != "" {
		htmlContent := h.markdownService.ParseMarkdown(song.Content)
		song.Content = string(htmlContent)
	}

	// Render the song details page
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongDetailsPage(song, bandType, user, originalMarkdown).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song details page: %v", err)
		http.Error(w, "Failed to render song details page", http.StatusInternalServerError)
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

// EditSong handles PUT /api/bands/songs/{songID}
func (h *SongHandler) EditSong(w http.ResponseWriter, r *http.Request) {
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
	content := r.FormValue("content")

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

	// Update song
	err = h.songsDB.UpdateSong(songID, title, artist, key, notes, content, tempo)
	if err != nil {
		log.Printf("Error updating song: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongDetailsError("Failed to update song", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Redirect to song details page
	http.Redirect(w, r, "/song?id="+songID, http.StatusSeeOther)
}

// UpdateSongContent handles POST /api/songs/{songID}/update-content
func (h *SongHandler) UpdateSongContent(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}
	songID := pathParts[3]

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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract content
	content := r.FormValue("content")

	// Update song content
	err = h.songsDB.UpdateSong(songID, song.Title, song.Artist, song.Key, song.Notes, content, song.Tempo)
	if err != nil {
		log.Printf("Error updating song content: %v", err)
		http.Error(w, "Failed to update song content", http.StatusInternalServerError)
		return
	}

	// Get the updated song with processed content
	updatedSong, err := h.songsDB.GetSongByID(songID)
	if err != nil {
		log.Printf("Error getting updated song: %v", err)
		http.Error(w, "Failed to get updated song", http.StatusInternalServerError)
		return
	}

	// Prepare original markdown and processed HTML
	originalMarkdown := updatedSong.Content
	if updatedSong.Content != "" {
		htmlContent := h.markdownService.ParseMarkdown(updatedSong.Content)
		updatedSong.Content = string(htmlContent)
	}

	// Return HTML response with the updated song content
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongContent(updatedSong, originalMarkdown).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song content: %v", err)
		http.Error(w, "Failed to render song content", http.StatusInternalServerError)
		return
	}
}

// ServeEditSong handles GET /song/edit
func (h *SongHandler) ServeEditSong(w http.ResponseWriter, r *http.Request) {
	songID := r.URL.Query().Get("id")
	if songID == "" {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.authService.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get song details
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

	// Get band details
	band, err := h.bandsDB.GetBandByID(song.BandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}
	if band == nil {
		http.Error(w, "Band not found", http.StatusNotFound)
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

	// Convert store.Band to types.Band
	bandType := &types.Band{
		ID:          band.ID,
		Name:        band.Name,
		Description: band.Description,
		CreatedBy:   band.CreatedBy,
		CreatedAt:   band.CreatedAt,
		UpdatedAt:   band.UpdatedAt,
		IsActive:    band.IsActive,
	}

	// Render the edit song page
	w.Header().Set("Content-Type", "text/html")
	err = templates.EditSongPage(song, bandType, user).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering edit song page: %v", err)
		http.Error(w, "Failed to render edit song page", http.StatusInternalServerError)
		return
	}
}

// ExportSongPDF handles GET /api/songs/{songID}/export-pdf
func (h *SongHandler) ExportSongPDF(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}
	songID := pathParts[3]

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

	// Get band details for the PDF (for future use)
	_, err = h.bandsDB.GetBandByID(song.BandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}

	// Create PDF request with original markdown content
	pdfReq := &services.SongContentPDFRequest{
		SongTitle: song.Title,
		Artist:    song.Artist,
		Key:       song.Key,
		Tempo:     song.Tempo,
		Content:   song.Content, // This is the original markdown content from the database
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateSongPDF(pdfReq)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Set response headers for PDF download
	filename := fmt.Sprintf("%s - %s.pdf", song.Title, song.Artist)
	if filename == " - .pdf" {
		filename = fmt.Sprintf("song_%s.pdf", songID)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Write PDF content
	w.Write(pdfBytes)
}

// GenerateSongContent handles POST /api/songs/{songID}/generate-content
func (h *SongHandler) GenerateSongContent(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}
	songID := pathParts[3]

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

	// Generate content using AI service
	aiReq := &services.SongContentRequest{
		SongTitle: song.Title,
		Artist:    song.Artist,
		Key:       song.Key,
		Tempo:     song.Tempo,
	}

	aiResponse, err := h.aiService.GenerateSongContent(aiReq)
	if err != nil {
		log.Printf("Error generating song content: %v", err)
		http.Error(w, "Failed to generate song content", http.StatusInternalServerError)
		return
	}

	// Update the song with the generated content
	err = h.songsDB.UpdateSong(songID, song.Title, song.Artist, song.Key, song.Notes, aiResponse.Content, song.Tempo)
	if err != nil {
		log.Printf("Error updating song with generated content: %v", err)
		http.Error(w, "Failed to update song with generated content", http.StatusInternalServerError)
		return
	}

	// Get the updated song with processed content
	updatedSong, err := h.songsDB.GetSongByID(songID)
	if err != nil {
		log.Printf("Error getting updated song: %v", err)
		http.Error(w, "Failed to get updated song", http.StatusInternalServerError)
		return
	}

	// Prepare original markdown and processed HTML
	originalMarkdown := updatedSong.Content
	if updatedSong.Content != "" {
		htmlContent := h.markdownService.ParseMarkdown(updatedSong.Content)
		updatedSong.Content = string(htmlContent)
	}

	// Return HTML response with the updated song content
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongContent(updatedSong, originalMarkdown).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song content: %v", err)
		http.Error(w, "Failed to render song content", http.StatusInternalServerError)
		return
	}
}
