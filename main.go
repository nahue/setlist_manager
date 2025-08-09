package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/nahue/setlist_manager/internal/app"
	"github.com/nahue/setlist_manager/internal/database"
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

	// Create application with all dependencies
	application := app.NewApplication(db)

	// Start server
	log.Fatal(application.Start("9090"))
}
