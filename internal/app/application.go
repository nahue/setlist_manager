package app

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nahue/setlist_manager/internal/api"
	"github.com/nahue/setlist_manager/internal/database"
	"github.com/nahue/setlist_manager/internal/services"
	"github.com/nahue/setlist_manager/internal/store"
)

// Application holds all the services and dependencies
type Application struct {
	db            *database.Database
	router        *chi.Mux
	useAuth       bool
	authHandler   *api.AuthHandler
	bandsHandler  *api.BandHandler
	songsHandler  *api.SongHandler
	healthHandler *api.HealthHandler
	authService   *services.AuthService
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

	// Create feature-specific database instances
	authDatabase := store.NewSQLiteAuthStore(db.GetDB())
	bandsDatabase := store.NewSQLiteBandsStore(db.GetDB())
	songsDatabase := store.NewSQLiteSongsStore(db.GetDB())

	// Create shared services
	authService := services.NewAuthService(authDatabase)

	// Create handlers
	authHandler := api.NewAuthHandler(authDatabase, bandsDatabase)
	bandsHandler := api.NewBandHandler(bandsDatabase, songsDatabase, authService)
	songsHandler := api.NewSongHandler(songsDatabase, bandsDatabase, authService)
	healthHandler := api.NewHealthHandler(db)

	app := &Application{
		db:            db,
		router:        chi.NewRouter(),
		useAuth:       useAuth,
		authHandler:   authHandler,
		bandsHandler:  bandsHandler,
		songsHandler:  songsHandler,
		healthHandler: healthHandler,
		authService:   authService,
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
	app.router.Get("/health", app.healthHandler.HandleHealth)
	app.router.Get("/ready", app.healthHandler.HandleReadiness)
	app.router.Get("/live", app.healthHandler.HandleLiveness)

	// Authentication routes (public)
	app.router.Get("/auth/login", app.authHandler.HandleLogin)
	app.router.Post("/auth/magic-link", app.authHandler.HandleMagicLinkRequest)
	app.router.Get("/auth/verify", app.authHandler.HandleMagicLinkVerification)
	app.router.Post("/auth/logout", app.authHandler.HandleLogout)
	app.router.Get("/auth/me", app.authHandler.HandleCurrentUser)

	// Apply auth middleware to protected routes
	app.router.Group(func(r chi.Router) {
		r.Use(app.authMiddleware)

		// Protected routes
		r.Get("/", app.serveWelcome)

		// Band routes
		r.Get("/bands", app.bandsHandler.ServeBands)
		r.Get("/bands/create", app.bandsHandler.ServeCreateBand)
		r.Get("/band", app.bandsHandler.ServeBand)

		// Band API routes
		r.Get("/api/bands", app.bandsHandler.GetBands)
		r.Post("/api/bands", app.bandsHandler.CreateBand)
		r.Get("/api/bands/band", app.bandsHandler.GetBand)
		r.Post("/api/bands/invite", app.bandsHandler.InviteMember)
		r.Delete("/api/bands/members/remove", app.bandsHandler.RemoveMember)

		// Song API routes
		r.Get("/api/bands/songs", app.songsHandler.GetSongs)
		r.Post("/api/bands/songs", app.songsHandler.CreateSong)
		r.Post("/api/bands/songs/reorder", app.songsHandler.ReorderSongs)
		r.Delete("/api/bands/songs/{songID}", app.songsHandler.DeleteSong)

		// Invitation routes
		r.Get("/api/invitations", app.bandsHandler.GetInvitations)
		r.Post("/api/invitations/accept", app.bandsHandler.AcceptInvitation)
		r.Post("/api/invitations/decline", app.bandsHandler.DeclineInvitation)
	})
}

// serveWelcome handles GET /
func (app *Application) serveWelcome(w http.ResponseWriter, r *http.Request) {
	// Redirect to bands page
	http.Redirect(w, r, "/bands", http.StatusSeeOther)
}

// authMiddleware checks if the user is authenticated
func (app *Application) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.useAuth {
			next.ServeHTTP(w, r)
			return
		}

		// Get current user from session
		user := app.authService.GetCurrentUser(r)
		if user == nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server on the specified port
func (app *Application) Start(port string) error {
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, app.router)
}
