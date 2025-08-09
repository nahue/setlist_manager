package app

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nahue/setlist_manager/internal/database"
)

// Application holds all the services and dependencies
type Application struct {
	db      *database.Database
	router  *chi.Mux
	useAuth bool
}

// NewApplication creates a new application instance with all dependencies
func NewApplication(db *database.Database) *Application {
	// Check if authentication is enabled via environment variable
	useAuth := true // default to true for security
	if useAuthStr := os.Getenv("USE_AUTH"); useAuthStr != "" {
		if parsed, err := strconv.ParseBool(useAuthStr); err == nil {
			useAuth = parsed
		} else {
			log.Printf("Warning: Invalid USE_AUTH value '%s', defaulting to true", useAuthStr)
		}
	}

	if !useAuth {
		log.Println("Warning: Authentication is DISABLED. This should only be used for development.")
	}

	app := &Application{
		db:      db,
		router:  chi.NewRouter(),
		useAuth: useAuth,
	}

	app.setupMiddleware()
	app.setupRoutes()

	return app
}

// setupMiddleware configures all middleware for the application
func (app *Application) setupMiddleware() {
	app.router.Use(middleware.Logger)
	app.router.Use(middleware.Recoverer)
	app.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

// setupRoutes configures all routes for the application
func (app *Application) setupRoutes() {
	// Health check routes (public)
	app.router.Get("/health", app.handleHealth)
	app.router.Get("/ready", app.handleReadiness)
	app.router.Get("/live", app.handleLiveness)

	// Authentication routes (public)
	app.router.Get("/auth/login", app.handleLogin)
	app.router.Post("/auth/magic-link", app.handleMagicLinkRequest)
	app.router.Get("/auth/verify", app.handleMagicLinkVerification)
	app.router.Post("/auth/logout", app.handleLogout)
	app.router.Get("/auth/me", app.handleCurrentUser)

	// Apply auth middleware to protected routes
	app.router.Group(func(r chi.Router) {
		r.Use(app.authMiddleware)

		// Protected routes
		r.Get("/", app.serveWelcome)

		// Band routes
		r.Get("/bands", app.serveBands)
		r.Get("/bands/create", app.serveCreateBand)
		r.Get("/band", app.serveBand)

		// Band API routes
		r.Get("/api/bands", app.getBands)
		r.Post("/api/bands", app.createBand)
		r.Get("/api/bands/band", app.getBand)
		r.Post("/api/bands/invite", app.inviteMember)

		// Invitation routes
		r.Get("/api/invitations", app.getInvitations)
		r.Post("/api/invitations/accept", app.acceptInvitation)
		r.Post("/api/invitations/decline", app.declineInvitation)
	})
}

// Start starts the HTTP server on the specified port
func (app *Application) Start(port string) error {
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, app.router)
}

// HTTP Handlers moved to separate files:
// - auth_handlers.go for authentication routes
// - pr_handlers.go for PR description routes
// - health_handlers.go for health check routes
