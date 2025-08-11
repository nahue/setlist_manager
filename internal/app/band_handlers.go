package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nahue/setlist_manager/templates"
)

// Request/Response structs
type CreateBandRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type InviteMemberRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type AcceptInvitationRequest struct {
	InvitationID string `json:"invitation_id"`
}

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

// serveBands handles GET /bands
func (app *Application) serveBands(w http.ResponseWriter, r *http.Request) {
	component := templates.BandsPage()
	component.Render(r.Context(), w)
}

// serveCreateBand handles GET /bands/create
func (app *Application) serveCreateBand(w http.ResponseWriter, r *http.Request) {
	component := templates.CreateBandPage()
	component.Render(r.Context(), w)
}

// serveBand handles GET /bands/{bandID}
func (app *Application) serveBand(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get band details
	band, err := app.db.GetBandByID(bandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}
	if band == nil {
		http.Error(w, "Band not found", http.StatusNotFound)
		return
	}

	// Get band members
	members, err := app.db.GetBandMembers(bandID)
	if err != nil {
		log.Printf("Error getting band members: %v", err)
		http.Error(w, "Failed to get band members", http.StatusInternalServerError)
		return
	}

	// Get songs for the band
	songs, err := app.db.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting songs: %v", err)
		http.Error(w, "Failed to get songs", http.StatusInternalServerError)
		return
	}

	component := templates.BandDetailsPage(band, members, songs, member.Role)
	component.Render(r.Context(), w)
}

// createBand handles POST /api/bands
func (app *Application) createBand(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateBandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Band name is required", http.StatusBadRequest)
		return
	}

	// Create the band
	band, err := app.db.CreateBand(req.Name, req.Description, user.ID)
	if err != nil {
		log.Printf("Error creating band: %v", err)
		http.Error(w, "Failed to create band", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"band":    band,
		"message": "Band created successfully",
	})
}

// getBands handles GET /api/bands
func (app *Application) getBands(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's bands
	bands, err := app.db.GetBandsByUser(user.ID)
	if err != nil {
		log.Printf("Error getting bands: %v", err)
		http.Error(w, "Failed to get bands", http.StatusInternalServerError)
		return
	}

	// Return bands
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"bands":   bands,
	})
}

// getBand handles GET /api/bands/{bandID}
func (app *Application) getBand(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get band details
	band, err := app.db.GetBandByID(bandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}
	if band == nil {
		http.Error(w, "Band not found", http.StatusNotFound)
		return
	}

	// Get band members
	members, err := app.db.GetBandMembers(bandID)
	if err != nil {
		log.Printf("Error getting band members: %v", err)
		http.Error(w, "Failed to get band members", http.StatusInternalServerError)
		return
	}

	// Return band data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"band":     band,
		"members":  members,
		"userRole": member.Role,
	})
}

// inviteMember handles POST /api/bands/{bandID}/invite
func (app *Application) inviteMember(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
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
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	email := r.FormValue("email")
	name := r.FormValue("name")
	role := r.FormValue("role")

	log.Printf("Received form data: email=%s, name=%s, role=%s", email, name, role)

	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if the email exists in the users table
	invitedUser, err := app.db.GetUserByEmail(email)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		http.Error(w, "Failed to check if user exists", http.StatusInternalServerError)
		return
	}
	if invitedUser == nil {
		http.Error(w, "User with this email does not exist. They must sign up first.", http.StatusBadRequest)
		return
	}

	// Check if user is already a member of this band
	existingMember, err := app.db.GetBandMember(bandID, invitedUser.ID)
	if err != nil {
		log.Printf("Error checking if user is already a member: %v", err)
		http.Error(w, "Failed to check if user is already a member", http.StatusInternalServerError)
		return
	}
	if existingMember != nil {
		http.Error(w, "User is already a member of this band", http.StatusBadRequest)
		return
	}

	// Set default role if not provided
	if role == "" {
		role = "member"
	}

	// Validate role
	if role != "member" && role != "admin" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Create invitation (expires in 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = app.db.CreateBandInvitation(bandID, email, user.ID, role, expiresAt)
	if err != nil {
		log.Printf("Error creating invitation: %v", err)
		http.Error(w, "Failed to create invitation", http.StatusInternalServerError)
		return
	}

	// Get updated band members
	members, err := app.db.GetBandMembers(bandID)
	if err != nil {
		log.Printf("Error getting updated band members: %v", err)
		http.Error(w, "Failed to get updated band members", http.StatusInternalServerError)
		return
	}

	// Return HTML response with the updated members section
	w.Header().Set("Content-Type", "text/html")

	// Render the members section directly to the response
	err = templates.MembersSection(members, bandID).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering members section: %v", err)
		http.Error(w, "Failed to render members section", http.StatusInternalServerError)
		return
	}
}

// getInvitations handles GET /api/invitations
func (app *Application) getInvitations(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get pending invitations for the user
	invitations, err := app.db.GetPendingInvitationsByEmail(user.Email)
	if err != nil {
		log.Printf("Error getting invitations: %v", err)
		http.Error(w, "Failed to get invitations", http.StatusInternalServerError)
		return
	}

	// Return invitations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"invitations": invitations,
	})
}

// acceptInvitation handles POST /api/invitations/accept
func (app *Application) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.InvitationID == "" {
		http.Error(w, "Invitation ID is required", http.StatusBadRequest)
		return
	}

	// Accept the invitation
	err := app.db.AcceptBandInvitation(req.InvitationID, user.ID)
	if err != nil {
		log.Printf("Error accepting invitation: %v", err)
		http.Error(w, "Failed to accept invitation", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Invitation accepted successfully",
	})
}

// declineInvitation handles POST /api/invitations/decline
func (app *Application) declineInvitation(w http.ResponseWriter, r *http.Request) {
	var req AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.InvitationID == "" {
		http.Error(w, "Invitation ID is required", http.StatusBadRequest)
		return
	}

	// Decline the invitation
	err := app.db.DeclineBandInvitation(req.InvitationID)
	if err != nil {
		log.Printf("Error declining invitation: %v", err)
		http.Error(w, "Failed to decline invitation", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Invitation declined successfully",
	})
}

// getSongs handles GET /api/bands/songs
func (app *Application) getSongs(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
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
	songs, err := app.db.GetSongsByBand(bandID)
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

// createSong handles POST /api/bands/songs
func (app *Application) createSong(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Debug: verify user exists
	userCheck, err := app.db.GetUserByID(user.ID)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}
	if userCheck == nil {
		log.Printf("User not found: %s", user.ID)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Debug: verify band exists
	bandCheck, err := app.db.GetBandByID(bandID)
	if err != nil {
		log.Printf("Error checking band existence: %v", err)
		http.Error(w, "Failed to verify band", http.StatusInternalServerError)
		return
	}
	if bandCheck == nil {
		log.Printf("Band not found: %s", bandID)
		http.Error(w, "Band not found", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	key := r.FormValue("key")
	tempoStr := r.FormValue("tempo")
	notes := r.FormValue("notes")

	// Convert tempo string to int pointer
	var tempo *int
	if tempoStr != "" {
		if tempoInt, err := strconv.Atoi(tempoStr); err == nil {
			tempo = &tempoInt
		}
	}

	log.Printf("Received form data: title=%s, artist=%s, key=%s, tempo=%s, notes=%s", title, artist, key, tempoStr, notes)

	if title == "" {
		http.Error(w, "Song title is required", http.StatusBadRequest)
		return
	}

	// Debug: log the values being passed to CreateSong
	log.Printf("Creating song with: bandID=%s, title=%s, artist=%s, key=%s, notes=%s, userID=%s, tempo=%v",
		bandID, title, artist, key, notes, user.ID, tempo)

	// Create the song
	_, err = app.db.CreateSong(bandID, title, artist, key, notes, user.ID, tempo)
	if err != nil {
		log.Printf("Error creating song: %v", err)
		http.Error(w, "Failed to create song", http.StatusInternalServerError)
		return
	}

	// Get updated songs list for the band
	songs, err := app.db.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting updated songs: %v", err)
		http.Error(w, "Failed to get updated songs", http.StatusInternalServerError)
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")

	// Render the songs section directly to the response
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}

// deleteSong handles DELETE /api/bands/songs/{songID}
func (app *Application) deleteSong(w http.ResponseWriter, r *http.Request) {
	// Extract song ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Song ID is required", http.StatusBadRequest)
		return
	}
	songID := pathParts[len(pathParts)-1]

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the song to check ownership
	song, err := app.db.GetSongByID(songID)
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
	member, err := app.db.GetBandMember(song.BandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Only allow song creator or band owner/admin to delete
	if song.CreatedBy != user.ID && member.Role != "owner" && member.Role != "admin" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete the song
	err = app.db.DeleteSong(songID)
	if err != nil {
		log.Printf("Error deleting song: %v", err)
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}

	// Get updated songs list for the band
	songs, err := app.db.GetSongsByBand(song.BandID)
	if err != nil {
		log.Printf("Error getting updated songs: %v", err)
		http.Error(w, "Failed to get updated songs", http.StatusInternalServerError)
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")

	// Render the songs section directly to the response
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}

// reorderSongs handles POST /api/bands/songs/reorder
func (app *Application) reorderSongs(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := app.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the band
	member, err := app.db.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req ReorderSongsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.SongOrder) == 0 {
		http.Error(w, "Song order is required", http.StatusBadRequest)
		return
	}

	// Reorder the songs
	err = app.db.ReorderSongs(bandID, req.SongOrder)
	if err != nil {
		log.Printf("Error reordering songs: %v", err)
		http.Error(w, "Failed to reorder songs", http.StatusInternalServerError)
		return
	}

	// Get updated songs list
	songs, err := app.db.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting songs after reorder: %v", err)
		http.Error(w, "Failed to get songs", http.StatusInternalServerError)
		return
	}

	// Return HTML response with the updated songs section
	w.Header().Set("Content-Type", "text/html")

	// Render the songs section directly to the response
	err = templates.SongsSection(songs).Render(r.Context(), w)
	if err != nil {
		log.Printf("Error rendering songs section: %v", err)
		http.Error(w, "Failed to render songs section", http.StatusInternalServerError)
		return
	}
}
