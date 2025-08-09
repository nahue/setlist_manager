package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Role  string `json:"role"`
}

type AcceptInvitationRequest struct {
	InvitationID string `json:"invitation_id"`
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

	// TODO: Create BandPage template
	http.Error(w, "Band page not implemented yet", http.StatusNotImplemented)
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

	// Check if user has permission to invite (owner or admin)
	member, err := app.db.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		http.Error(w, "Failed to check band membership", http.StatusInternalServerError)
		return
	}
	if member == nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "member"
	}

	// Validate role
	if req.Role != "member" && req.Role != "admin" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Create invitation (expires in 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation, err := app.db.CreateBandInvitation(bandID, req.Email, user.ID, req.Role, expiresAt)
	if err != nil {
		log.Printf("Error creating invitation: %v", err)
		http.Error(w, "Failed to create invitation", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"invitation": invitation,
		"message":    fmt.Sprintf("Invitation sent to %s", req.Email),
	})
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
