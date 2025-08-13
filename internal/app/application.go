package app

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nahue/setlist_manager/internal/api"
	"github.com/nahue/setlist_manager/internal/database"
	"github.com/nahue/setlist_manager/internal/services"
	"github.com/nahue/setlist_manager/internal/store"
)

// Application represents the main application
type Application struct {
	router        *chi.Mux
	authService   *services.AuthService
	authHandler   *api.AuthHandler
	bandsHandler  *api.BandHandler
	songsHandler  *api.SongHandler
	healthHandler *api.HealthHandler
}

// NewApplication creates a new application instance
func NewApplication(
	db *database.Database,
	authStore *store.SQLiteAuthStore,
	bandsStore *store.SQLiteBandsStore,
	songsStore *store.SQLiteSongsStore,
) *Application {
	// Initialize services
	authService := services.NewAuthService(authStore)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authStore, bandsStore)
	bandsHandler := api.NewBandHandler(bandsStore, songsStore, authService)
	songsHandler := api.NewSongHandler(songsStore, bandsStore, authService)
	healthHandler := api.NewHealthHandler(db)

	// Initialize router
	router := chi.NewRouter()

	app := &Application{
		router:        router,
		authService:   authService,
		authHandler:   authHandler,
		bandsHandler:  bandsHandler,
		songsHandler:  songsHandler,
		healthHandler: healthHandler,
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
		// Get current user from session
		user := app.authService.GetCurrentUser(r)
		if user == nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Store user in request context
		ctx := context.WithValue(r.Context(), api.UserContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Start starts the HTTP server on the specified port
func (app *Application) Start(port string) error {
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, app.router)
}
