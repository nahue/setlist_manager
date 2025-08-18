package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nahue/setlist_manager/internal/store"
)

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "./data/setlist_manager.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🌱 Starting database seeding...")

	// Initialize stores
	authStore := store.NewSQLiteAuthStore(db)
	bandsStore := store.NewSQLiteBandsStore(db)
	songsStore := store.NewSQLiteSongsStore(db)

	// Seed users
	fmt.Println("👥 Creating users...")
	users := seedUsers(authStore)
	if len(users) == 0 {
		log.Fatal("Failed to create users")
	}

	// Seed bands
	fmt.Println("🎸 Creating bands...")
	bands := seedBands(bandsStore, users)
	if len(bands) == 0 {
		log.Fatal("Failed to create bands")
	}

	// Seed songs
	fmt.Println("🎵 Creating songs...")
	songs := seedSongs(songsStore, bands, users)
	if len(songs) == 0 {
		log.Fatal("Failed to create songs")
	}

	fmt.Println("✅ Database seeding completed successfully!")
}

func seedUsers(authStore *store.SQLiteAuthStore) []*store.User {
	userEmails := []string{
		"john@example.com",
		"sarah@example.com",
		"mike@example.com",
		"lisa@example.com",
		"david@example.com",
	}

	var users []*store.User
	for _, email := range userEmails {
		user, err := authStore.CreateUser(email)
		if err != nil {
			fmt.Printf("Warning: Failed to create user %s: %v\n", email, err)
			continue
		}
		users = append(users, user)
		fmt.Printf("  ✅ Created user: %s\n", email)
	}

	return users
}

func seedBands(bandsStore *store.SQLiteBandsStore, users []*store.User) []*store.Band {
	if len(users) == 0 {
		return nil
	}

	bandData := []struct {
		name        string
		description string
		creatorIdx  int
		members     []int // indices of users to add as members
	}{
		{
			name:        "The Rock Band",
			description: "A classic rock cover band playing hits from the 70s and 80s",
			creatorIdx:  0,
			members:     []int{0, 1, 2},
		},
		{
			name:        "Jazz Collective",
			description: "Modern jazz ensemble exploring contemporary compositions",
			creatorIdx:  1,
			members:     []int{1, 3, 4},
		},
		{
			name:        "Acoustic Duo",
			description: "Intimate acoustic performances with guitar and vocals",
			creatorIdx:  2,
			members:     []int{2, 4},
		},
	}

	var bands []*store.Band
	for _, data := range bandData {
		if data.creatorIdx >= len(users) {
			continue
		}

		band, err := bandsStore.CreateBand(data.name, data.description, users[data.creatorIdx].ID)
		if err != nil {
			fmt.Printf("Warning: Failed to create band %s: %v\n", data.name, err)
			continue
		}

		// Add additional members
		for _, memberIdx := range data.members {
			if memberIdx < len(users) && memberIdx != data.creatorIdx {
				_, err := bandsStore.AddBandMember(band.ID, users[memberIdx].ID, "member")
				if err != nil {
					fmt.Printf("Warning: Failed to add member to band %s: %v\n", data.name, err)
				}
			}
		}

		bands = append(bands, band)
		fmt.Printf("  ✅ Created band: %s\n", data.name)
	}

	return bands
}

func seedSongs(songsStore *store.SQLiteSongsStore, bands []*store.Band, users []*store.User) []*store.Song {
	if len(bands) == 0 || len(users) == 0 {
		return nil
	}

	songData := []struct {
		title      string
		artist     string
		key        string
		tempo      *int
		notes      string
		bandIdx    int
		creatorIdx int
	}{
		// Rock Band songs
		{
			title:      "Still Loving You",
			artist:     "Scorpions",
			key:        "D",
			tempo:      intPtr(75),
			notes:      "Power ballad with emotional guitar work",
			bandIdx:    0,
			creatorIdx: 0,
		},
		{
			title:      "Don't Dream It's Over",
			artist:     "Crowded House",
			key:        "C",
			tempo:      intPtr(85),
			notes:      "Melodic pop-rock with memorable chorus",
			bandIdx:    0,
			creatorIdx: 1,
		},
		{
			title:      "Is This Love",
			artist:     "Whitesnake",
			key:        "A",
			tempo:      intPtr(95),
			notes:      "Classic rock ballad with powerful vocals",
			bandIdx:    0,
			creatorIdx: 2,
		},
		{
			title:      "Everybody Wants to Rule the World",
			artist:     "Tears for Fears",
			key:        "D",
			tempo:      intPtr(120),
			notes:      "80s synth-pop with driving rhythm",
			bandIdx:    0,
			creatorIdx: 0,
		},
		{
			title:      "Otro dia mas sin verte",
			artist:     "Jon Secada",
			key:        "G",
			tempo:      intPtr(80),
			notes:      "Latin pop ballad with emotional delivery",
			bandIdx:    0,
			creatorIdx: 1,
		},
		{
			title:      "Money for Nothing",
			artist:     "Dire Straits",
			key:        "D",
			tempo:      intPtr(135),
			notes:      "Rock anthem with iconic guitar riff",
			bandIdx:    0,
			creatorIdx: 2,
		},
		// Jazz Collective songs
		{
			title:      "Take Five",
			artist:     "Dave Brubeck",
			key:        "Eb",
			tempo:      intPtr(176),
			notes:      "Jazz standard in 5/4 time signature",
			bandIdx:    1,
			creatorIdx: 1,
		},
		{
			title:      "So What",
			artist:     "Miles Davis",
			key:        "D",
			tempo:      intPtr(160),
			notes:      "Modal jazz composition from Kind of Blue",
			bandIdx:    1,
			creatorIdx: 3,
		},
		// Acoustic Duo songs
		{
			title:      "Wonderwall",
			artist:     "Oasis",
			key:        "F#m",
			tempo:      intPtr(87),
			notes:      "Acoustic version with simplified arrangement",
			bandIdx:    2,
			creatorIdx: 2,
		},
		{
			title:      "Hallelujah",
			artist:     "Jeff Buckley",
			key:        "C",
			tempo:      intPtr(72),
			notes:      "Emotional ballad with fingerpicking",
			bandIdx:    2,
			creatorIdx: 4,
		},
	}

	var songs []*store.Song
	for _, data := range songData {
		if data.bandIdx >= len(bands) || data.creatorIdx >= len(users) {
			continue
		}

		song, err := songsStore.CreateSong(
			bands[data.bandIdx].ID,
			data.title,
			data.artist,
			data.key,
			data.notes,
			"", // Empty content field - will be generated by AI if needed
			users[data.creatorIdx].ID,
			data.tempo,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to create song %s: %v\n", data.title, err)
			continue
		}

		songs = append(songs, song)
		fmt.Printf("  ✅ Created song: %s by %s\n", data.title, data.artist)
	}

	return songs
}

func intPtr(i int) *int {
	return &i
}
