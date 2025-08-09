package app

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version,omitempty"`
}

// handleHealth handles GET /health
func (app *Application) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	// Check database connectivity
	if err := app.db.Ping(); err != nil {
		response.Status = "error"
		response.Database = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		response.Database = "connected"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleReadiness handles GET /ready
func (app *Application) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Check if all dependencies are ready
	if err := app.db.Ping(); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}

	// Add other readiness checks here (Redis, external APIs, etc.)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleLiveness handles GET /live
func (app *Application) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
