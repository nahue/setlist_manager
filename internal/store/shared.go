package store

import (
	"fmt"
	"time"
)

// Helper function to generate UUID (simplified for SQLite)
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
