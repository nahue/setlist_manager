package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/nahue/setlist_manager/internal/app"
	"github.com/nahue/setlist_manager/internal/database"
	"github.com/nahue/setlist_manager/internal/store"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create feature-specific database instances
	authStore := store.NewSQLiteAuthStore(db.GetDB())
	bandsStore := store.NewSQLiteBandsStore(db.GetDB())
	songsStore := store.NewSQLiteSongsStore(db.GetDB())
	sectionsStore := store.NewSQLiteSongSectionsStore(db.GetDB())

	// Create application with all dependencies - always use authentication
	application := app.NewApplication(db, authStore, bandsStore, songsStore, sectionsStore)

	// Start server
	log.Fatal(application.Start("9090"))
}
