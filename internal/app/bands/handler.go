package bands

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nahue/setlist_manager/internal/app/bands/database"
	"github.com/nahue/setlist_manager/internal/app/shared"
	"github.com/nahue/setlist_manager/internal/app/shared/types"
	"github.com/nahue/setlist_manager/internal/store"
	"github.com/nahue/setlist_manager/templates"
)

// Handler handles band-related requests
type Handler struct {
	bandsDB     *database.Database
	songsDB     *store.SQLiteSongsStore
	authService *shared.AuthService
}

// NewHandler creates a new bands handler
func NewHandler(bandsDB *database.Database, songsDB *store.SQLiteSongsStore, authService *shared.AuthService) *Handler {
	return &Handler{
		bandsDB:     bandsDB,
		songsDB:     songsDB,
		authService: authService,
	}
}

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

// ServeBands handles GET /bands
func (h *Handler) ServeBands(w http.ResponseWriter, r *http.Request) {
	component := templates.BandsPage()
	component.Render(r.Context(), w)
}

// ServeCreateBand handles GET /bands/create
func (h *Handler) ServeCreateBand(w http.ResponseWriter, r *http.Request) {
	component := templates.CreateBandPage()
	component.Render(r.Context(), w)
}

// ServeBand handles GET /bands/{bandID}
func (h *Handler) ServeBand(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.getCurrentUser(r)
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

	// Get band details
	band, err := h.bandsDB.GetBandByIDShared(bandID)
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
	members, err := h.bandsDB.GetBandMembersShared(bandID)
	if err != nil {
		log.Printf("Error getting band members: %v", err)
		http.Error(w, "Failed to get band members", http.StatusInternalServerError)
		return
	}
	log.Printf("Band members: %v", members)

	// Get songs for the band
	songs, err := h.songsDB.GetSongsByBand(bandID)
	if err != nil {
		log.Printf("Error getting songs: %v", err)
		http.Error(w, "Failed to get songs", http.StatusInternalServerError)
		return
	}

	// Determine user role
	userRole := "member"
	if member.Role == "owner" {
		userRole = "owner"
	} else if member.Role == "admin" {
		userRole = "admin"
	}

	// Render band details page
	component := templates.BandDetailsContent(band, members, songs, userRole)
	component.Render(r.Context(), w)
}

// GetBands handles GET /api/bands
func (h *Handler) GetBands(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get bands for the user
	bands, err := h.bandsDB.GetBandsByUserShared(user.ID)
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

// CreateBand handles POST /api/bands
func (h *Handler) CreateBand(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := h.getCurrentUser(r)
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

	// Create band
	band, err := h.bandsDB.CreateBand(req.Name, req.Description, user.ID)
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
	})
}

// GetBand handles GET /api/bands/band
func (h *Handler) GetBand(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		http.Error(w, "Band ID is required", http.StatusBadRequest)
		return
	}

	// Get current user from session
	user := h.getCurrentUser(r)
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

	// Get band details
	band, err := h.bandsDB.GetBandByIDShared(bandID)
	if err != nil {
		log.Printf("Error getting band: %v", err)
		http.Error(w, "Failed to get band", http.StatusInternalServerError)
		return
	}
	if band == nil {
		http.Error(w, "Band not found", http.StatusNotFound)
		return
	}

	// Return band
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"band":    band,
	})
}

// getCurrentUser gets the current user from the session
func (h *Handler) getCurrentUser(r *http.Request) *types.User {
	return h.authService.GetCurrentUser(r)
}

// InviteMember handles POST /api/bands/invite
func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err := templates.MembersSectionError("Band ID is required", "").Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get current user from session
	user := h.getCurrentUser(r)
	if user == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err := templates.MembersSectionError("Unauthorized", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if user is a member of the band
	member, err := h.bandsDB.GetBandMember(bandID, user.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to check band membership", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}
	if member == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Access denied", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Error parsing form data", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Extract form values
	email := r.FormValue("email")
	name := r.FormValue("name")
	role := r.FormValue("role")

	log.Printf("Received form data: email=%s, name=%s, role=%s", email, name, role)

	if email == "" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Email is required", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if the email exists in the users table
	invitedUser, err := h.bandsDB.GetUserByEmail(email)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to check if user exists", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}
	if invitedUser == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("User with this email does not exist. They must sign up first.", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if user is already a member of this band
	existingMember, err := h.bandsDB.GetBandMember(bandID, invitedUser.ID)
	if err != nil {
		log.Printf("Error checking if user is already a member: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to check if user is already a member", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}
	if existingMember != nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("User is already a member of this band", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Set default role if not provided
	if role == "" {
		role = "member"
	}

	// Validate role
	if role != "member" && role != "admin" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Invalid role", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Add member directly to the band
	_, err = h.bandsDB.AddBandMember(bandID, invitedUser.ID, role)
	if err != nil {
		log.Printf("Error adding member to band: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to add member to band", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated band members
	members, err := h.bandsDB.GetBandMembersShared(bandID)
	if err != nil {
		log.Printf("Error getting updated band members: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to get updated band members", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
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

// RemoveMember handles DELETE /api/bands/members/remove
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	bandID := r.URL.Query().Get("id")
	if bandID == "" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err := templates.MembersSectionError("Band ID is required", "").Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err := templates.MembersSectionError("User ID is required", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get current user from session
	currentUser := h.getCurrentUser(r)
	if currentUser == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err := templates.MembersSectionError("Unauthorized", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if current user is a member of the band
	currentMember, err := h.bandsDB.GetBandMember(bandID, currentUser.ID)
	if err != nil {
		log.Printf("Error checking band membership: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to check band membership", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}
	if currentMember == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Access denied", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if user is trying to remove themselves
	if currentUser.ID == userID {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("You cannot remove yourself from the band", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if the user to be removed is a member of this band
	targetMember, err := h.bandsDB.GetBandMember(bandID, userID)
	if err != nil {
		log.Printf("Error checking target user membership: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to check target user membership", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}
	if targetMember == nil {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("User is not a member of this band", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Check if the user being removed is the owner
	if targetMember.Role == "owner" {
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("The owner cannot be removed from the band", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Remove the member from the band
	err = h.bandsDB.RemoveBandMember(bandID, userID)
	if err != nil {
		log.Printf("Error removing band member: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to remove band member", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
		return
	}

	// Get updated band members
	members, err := h.bandsDB.GetBandMembersShared(bandID)
	if err != nil {
		log.Printf("Error getting updated band members: %v", err)
		// Return HTML error response
		w.Header().Set("Content-Type", "text/html")
		err = templates.MembersSectionError("Failed to get updated band members", bandID).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			http.Error(w, "Failed to render error template", http.StatusInternalServerError)
		}
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

// GetInvitations handles GET /api/invitations
func (h *Handler) GetInvitations(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get pending invitations for the user
	invitations, err := h.bandsDB.GetPendingInvitationsByEmail(user.Email)
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

// AcceptInvitation handles POST /api/invitations/accept
func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user := h.getCurrentUser(r)
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
	err := h.bandsDB.AcceptBandInvitation(req.InvitationID, user.ID)
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

// DeclineInvitation handles POST /api/invitations/decline
func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
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
	err := h.bandsDB.DeclineBandInvitation(req.InvitationID)
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
