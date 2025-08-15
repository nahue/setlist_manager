package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/nahue/setlist_manager/internal/services"
	"github.com/nahue/setlist_manager/internal/store"
	"github.com/nahue/setlist_manager/templates"
)

// SongSectionsHandler handles song section-related requests
type SongSectionsHandler struct {
	sectionsDB      *store.SQLiteSongSectionsStore
	songsDB         *store.SQLiteSongsStore
	bandsDB         *store.SQLiteBandsStore
	authService     *services.AuthService
	authStore       *store.SQLiteAuthStore
	markdownService *services.MarkdownService
}

// NewSongSectionsHandler creates a new song sections handler
func NewSongSectionsHandler(sectionsDB *store.SQLiteSongSectionsStore, songsDB *store.SQLiteSongsStore, bandsDB *store.SQLiteBandsStore, authService *services.AuthService, authStore *store.SQLiteAuthStore, markdownService *services.MarkdownService) *SongSectionsHandler {
	return &SongSectionsHandler{
		sectionsDB:      sectionsDB,
		songsDB:         songsDB,
		bandsDB:         bandsDB,
		authService:     authService,
		authStore:       authStore,
		markdownService: markdownService,
	}
}

// Request/Response structs
type CreateSongSectionRequest struct {
	Title string `json:"title"`
	Key   string `json:"key"`
	Body  string `json:"body"`
}

type ReorderSongSectionsRequest struct {
	SectionOrder []string `json:"section_order"`
}

// GetSongSections handles GET /api/songs/{songID}/sections
func (h *SongSectionsHandler) GetSongSections(w http.ResponseWriter, r *http.Request) {
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

	// Get sections for the song
	sections, err := h.sectionsDB.GetSongSectionsBySongID(songID)
	if err != nil {
		log.Printf("Error getting song sections: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to get song sections", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the sections
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongSections(sections, songID).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song sections: %v", err)
		http.Error(w, "Failed to render song sections", http.StatusInternalServerError)
		return
	}
}

// CreateSongSection handles POST /api/songs/{songID}/sections
func (h *SongSectionsHandler) CreateSongSection(w http.ResponseWriter, r *http.Request) {
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

	// Extract form fields
	title := r.FormValue("title")
	key := r.FormValue("key")
	body := r.FormValue("body")

	if title == "" {
		http.Error(w, "Section title is required", http.StatusBadRequest)
		return
	}

	// Create song section
	_, err = h.sectionsDB.CreateSongSection(songID, title, key, body, user.ID)
	if err != nil {
		log.Printf("Error creating song section: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to create song section", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated sections list to return
	sections, err := h.sectionsDB.GetSongSectionsBySongID(songID)
	if err != nil {
		log.Printf("Error getting updated sections: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to get updated sections", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated sections
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongSections(sections, songID).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song sections: %v", err)
		http.Error(w, "Failed to render song sections", http.StatusInternalServerError)
		return
	}
}

// ReorderSongSections handles POST /api/songs/{songID}/sections/reorder
func (h *SongSectionsHandler) ReorderSongSections(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
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

	var req ReorderSongSectionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reorder sections
	err = h.sectionsDB.ReorderSongSections(songID, req.SectionOrder)
	if err != nil {
		log.Printf("Error reordering song sections: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to reorder song sections", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated sections list to return
	sections, err := h.sectionsDB.GetSongSectionsBySongID(songID)
	if err != nil {
		log.Printf("Error getting updated sections: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to get updated sections", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated sections
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongSections(sections, songID).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song sections: %v", err)
		http.Error(w, "Failed to render song sections", http.StatusInternalServerError)
		return
	}
}

// DeleteSongSection handles DELETE /api/songs/{songID}/sections/{sectionID}
func (h *SongSectionsHandler) DeleteSongSection(w http.ResponseWriter, r *http.Request) {
	// Extract song ID and section ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Song ID and Section ID are required", http.StatusBadRequest)
		return
	}
	songID := pathParts[3]
	sectionID := pathParts[5]

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

	// Delete song section
	err = h.sectionsDB.DeleteSongSection(sectionID)
	if err != nil {
		log.Printf("Error deleting song section: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to delete song section", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated sections list to return
	sections, err := h.sectionsDB.GetSongSectionsBySongID(songID)
	if err != nil {
		log.Printf("Error getting updated sections: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.SongSectionsError("Failed to get updated sections", songID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML response with the updated sections
	w.Header().Set("Content-Type", "text/html")
	err = templates.SongSections(sections, songID).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering song sections: %v", err)
		http.Error(w, "Failed to render song sections", http.StatusInternalServerError)
		return
	}
}
