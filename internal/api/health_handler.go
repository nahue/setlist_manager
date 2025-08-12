package api

import (
	"encoding/json"
	"net/http"

	"github.com/nahue/setlist_manager/internal/database"
)

// Handler handles health check requests
type HealthHandler struct {
	db *database.Database
}

// NewHandler creates a new health handler
func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version,omitempty"`
}

// HandleHealth handles GET /health
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	// Check database connectivity
	if err := h.db.Ping(); err != nil {
		response.Status = "error"
		response.Database = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		response.Database = "connected"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleReadiness handles GET /ready
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	// Check if all dependencies are ready
	if err := h.db.Ping(); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}

	// Add other readiness checks here (Redis, external APIs, etc.)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleLiveness handles GET /live
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
