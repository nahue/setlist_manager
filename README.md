# Setlist Manager

A modern web application for managing band setlists, built with Go, Chi router, and Alpine.js. Features magic link authentication, band management, song organization, and collaborative setlist creation.

## Features

- **Authentication**: Magic link authentication system
- **Band Management**: Create and manage bands with member invitations
- **Song Management**: Add, edit, and organize songs within bands
- **Setlist Organization**: Drag-and-drop song reordering
- **Collaborative**: Multiple band members can contribute
- **Modern UI**: Responsive design with Alpine.js and Tailwind CSS
- **Type-safe Templates**: Templ for server-side rendering

## Project Structure

```
setlist_manager/
├── main.go                    # Application entry point
├── go.mod                     # Go module dependencies
├── Taskfile.yml               # Build and development tasks
├── .air.toml                  # Hot reload configuration
├── goose.yaml                 # Database migration configuration
├── data/                      # SQLite database files
├── migrations/                # Database migration files
├── components/                # Reusable UI components
│   ├── health.templ           # Health check component
│   └── health_templ.go        # Generated component code
├── templates/                 # Page templates
│   ├── layout.templ           # Base layout with navigation
│   ├── index.templ            # Welcome page
│   ├── login.templ            # Login page
│   ├── bands.templ            # Bands listing page
│   ├── create_band.templ      # Create band form
│   ├── band_details.templ     # Band details and songs
│   └── *_templ.go             # Generated template code
└── internal/                  # Application internals
    ├── app/                   # Application setup
    │   ├── application.go     # Main application configuration
    │   ├── context.go         # User context utilities
    │   └── shared/            # Shared types and utilities
    │       └── types/         # Data types
    │           └── types.go   # User, Band, Song, etc.
    ├── api/                   # HTTP handlers
    │   ├── auth_handler.go    # Authentication endpoints
    │   ├── band_handler.go    # Band management endpoints
    │   ├── song_handler.go    # Song management endpoints
    │   └── health_handler.go  # Health check endpoints
    ├── services/              # Business logic
    │   └── auth_service.go    # Authentication service
    ├── store/                 # Data access layer
    │   ├── auth_store.go      # User and session storage
    │   ├── bands_store.go     # Band and member storage
    │   ├── songs_store.go     # Song storage
    │   └── shared.go          # Shared database utilities
    └── database/              # Database connection
        └── database.go        # SQLite connection setup
```

## Getting Started

### Prerequisites
- Go 1.24.4 or higher
- SQLite3

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd setlist_manager
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up the database:
```bash
task db:reset
```

4. Run the application:
```bash
task dev
```

5. Open your browser and navigate to:
```
http://localhost:9090
```

## Development Workflow

### Available Tasks

```bash
# Development with hot reload
task dev

# Build the application
task build

# Run tests
task test

# Clean build artifacts
task clean

# Generate templ files
task templ

# Database operations
task db:migrate    # Run migrations
task db:rollback   # Rollback last migration
task db:status     # Show migration status
task db:reset      # Reset database (drop all tables and re-run migrations)
```

## Creating New Features

This section provides detailed instructions for extending the application with new features.

### 1. Creating New Data Types

**Location**: `internal/app/shared/types/types.go`

Add new structs to define your data models:

```go
// Example: Adding a new "Event" type
type Event struct {
    ID          string    `json:"id"`
    BandID      string    `json:"band_id"`
    Name        string    `json:"name"`
    Date        time.Time `json:"date"`
    Location    string    `json:"location"`
    Description string    `json:"description"`
    CreatedBy   string    `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    IsActive    bool      `json:"is_active"`
}
```

### 2. Creating Database Migrations

**Location**: `migrations/`

Create new migration files with the naming convention: `YYYYMMDDHHMMSS_description.sql`

```sql
-- Example: 20250808000000_add_events.sql
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    name TEXT NOT NULL,
    date DATETIME NOT NULL,
    location TEXT,
    description TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    FOREIGN KEY (band_id) REFERENCES bands(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_events_band_id ON events(band_id);
CREATE INDEX idx_events_date ON events(date);
```

Run the migration:
```bash
task db:migrate
```

### 3. Creating Database Stores

**Location**: `internal/store/`

Create a new store file following the existing pattern:

```go
// events_store.go
package store

import (
    "database/sql"
    "fmt"
    "time"
    
    "github.com/nahue/setlist_manager/internal/app/shared/types"
)

type SQLiteEventsStore struct {
    db *sql.DB
}

func NewSQLiteEventsStore(db *sql.DB) *SQLiteEventsStore {
    return &SQLiteEventsStore{db: db}
}

// CreateEvent creates a new event
func (s *SQLiteEventsStore) CreateEvent(event *types.Event) (*types.Event, error) {
    event.ID = generateID()
    event.CreatedAt = time.Now()
    event.UpdatedAt = time.Now()
    event.IsActive = true
    
    query := `
        INSERT INTO events (id, band_id, name, date, location, description, created_by, created_at, updated_at, is_active)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := s.db.Exec(query, 
        event.ID, event.BandID, event.Name, event.Date, 
        event.Location, event.Description, event.CreatedBy,
        event.CreatedAt, event.UpdatedAt, event.IsActive)
    
    if err != nil {
        return nil, fmt.Errorf("failed to create event: %w", err)
    }
    
    return event, nil
}

// GetEventsByBand gets all events for a band
func (s *SQLiteEventsStore) GetEventsByBand(bandID string) ([]*types.Event, error) {
    query := `
        SELECT id, band_id, name, date, location, description, created_by, created_at, updated_at, is_active
        FROM events
        WHERE band_id = ? AND is_active = 1
        ORDER BY date ASC
    `
    
    rows, err := s.db.Query(query, bandID)
    if err != nil {
        return nil, fmt.Errorf("failed to query events: %w", err)
    }
    defer rows.Close()
    
    var events []*types.Event
    for rows.Next() {
        event := &types.Event{}
        err := rows.Scan(
            &event.ID, &event.BandID, &event.Name, &event.Date,
            &event.Location, &event.Description, &event.CreatedBy,
            &event.CreatedAt, &event.UpdatedAt, &event.IsActive)
        if err != nil {
            return nil, fmt.Errorf("failed to scan event: %w", err)
        }
        events = append(events, event)
    }
    
    return events, nil
}

// Additional methods: GetEventByID, UpdateEvent, DeleteEvent, etc.
```

### 4. Creating Services

**Location**: `internal/services/`

Create business logic services:

```go
// events_service.go
package services

import (
    "fmt"
    "github.com/nahue/setlist_manager/internal/app/shared/types"
    "github.com/nahue/setlist_manager/internal/store"
)

type EventsService struct {
    eventsStore *store.SQLiteEventsStore
    bandsStore  *store.SQLiteBandsStore
}

func NewEventsService(eventsStore *store.SQLiteEventsStore, bandsStore *store.SQLiteBandsStore) *EventsService {
    return &EventsService{
        eventsStore: eventsStore,
        bandsStore:  bandsStore,
    }
}

func (s *EventsService) CreateEvent(bandID, name, location, description string, date time.Time, userID string) (*types.Event, error) {
    // Validate band membership
    member, err := s.bandsStore.GetBandMember(bandID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to check band membership: %w", err)
    }
    if member == nil {
        return nil, fmt.Errorf("user is not a member of this band")
    }
    
    event := &types.Event{
        BandID:      bandID,
        Name:        name,
        Date:        date,
        Location:    location,
        Description: description,
        CreatedBy:   userID,
    }
    
    return s.eventsStore.CreateEvent(event)
}

func (s *EventsService) GetEventsForBand(bandID, userID string) ([]*types.Event, error) {
    // Validate band membership
    member, err := s.bandsStore.GetBandMember(bandID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to check band membership: %w", err)
    }
    if member == nil {
        return nil, fmt.Errorf("user is not a member of this band")
    }
    
    return s.eventsStore.GetEventsByBand(bandID)
}
```

### 5. Creating API Handlers

**Location**: `internal/api/`

Create HTTP handlers following the existing pattern:

```go
// events_handler.go
package api

import (
    "encoding/json"
    "log"
    "net/http"
    "time"
    
    "github.com/nahue/setlist_manager/internal/app/shared/types"
    "github.com/nahue/setlist_manager/internal/services"
    "github.com/nahue/setlist_manager/internal/store"
    "github.com/nahue/setlist_manager/templates"
)

type EventsHandler struct {
    eventsStore *store.SQLiteEventsStore
    eventsService *services.EventsService
}

func NewEventsHandler(eventsStore *store.SQLiteEventsStore, eventsService *services.EventsService) *EventsHandler {
    return &EventsHandler{
        eventsStore:   eventsStore,
        eventsService: eventsService,
    }
}

// Request/Response structs
type CreateEventRequest struct {
    Name        string    `json:"name"`
    Date        time.Time `json:"date"`
    Location    string    `json:"location"`
    Description string    `json:"description"`
}

// ServeEvents handles GET /events
func (h *EventsHandler) ServeEvents(w http.ResponseWriter, r *http.Request) {
    bandID := r.URL.Query().Get("band_id")
    if bandID == "" {
        http.Error(w, "Band ID is required", http.StatusBadRequest)
        return
    }
    
    user := GetUserFromContext(r.Context())
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    events, err := h.eventsService.GetEventsForBand(bandID, user.ID)
    if err != nil {
        log.Printf("Error getting events: %v", err)
        http.Error(w, "Failed to get events", http.StatusInternalServerError)
        return
    }
    
    component := templates.EventsPage(events, bandID, user)
    component.Render(r.Context(), w)
}

// CreateEvent handles POST /api/events
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
    bandID := r.URL.Query().Get("band_id")
    if bandID == "" {
        http.Error(w, "Band ID is required", http.StatusBadRequest)
        return
    }
    
    user := GetUserFromContext(r.Context())
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    var req CreateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    event, err := h.eventsService.CreateEvent(bandID, req.Name, req.Location, req.Description, req.Date, user.ID)
    if err != nil {
        log.Printf("Error creating event: %v", err)
        http.Error(w, "Failed to create event", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "event":   event,
    })
}

// GetEvents handles GET /api/events
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
    bandID := r.URL.Query().Get("band_id")
    if bandID == "" {
        http.Error(w, "Band ID is required", http.StatusBadRequest)
        return
    }
    
    user := GetUserFromContext(r.Context())
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    events, err := h.eventsService.GetEventsForBand(bandID, user.ID)
    if err != nil {
        log.Printf("Error getting events: %v", err)
        http.Error(w, "Failed to get events", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "events":  events,
    })
}
```

### 6. Creating Templates

**Location**: `templates/`

Create new template files:

```templ
// events.templ
package templates

import "github.com/nahue/setlist_manager/internal/app/shared/types"

templ EventsPage(events []*types.Event, bandID string, user *types.User) {
    @BaseLayout(PageData{
        Title: "Events",
        Description: "Manage band events and performances",
        Content: EventsContent(events, bandID),
        User: user,
    })
}

templ EventsContent(events []*types.Event, bandID string) {
    <div class="max-w-6xl mx-auto">
        <div class="mb-8">
            <div class="flex justify-between items-center">
                <h1 class="text-3xl font-bold text-gray-900">Band Events</h1>
                <button 
                    @click="showAddEventModal = true"
                    class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">
                    <svg class="-ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    Add Event
                </button>
            </div>
        </div>

        <div class="bg-white shadow rounded-lg">
            <div class="px-6 py-4 border-b border-gray-200">
                <h2 class="text-lg font-medium text-gray-900">Upcoming Events</h2>
            </div>
            <div class="p-6">
                if len(events) == 0 {
                    <div class="text-center py-8">
                        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <p class="mt-2 text-sm text-gray-500">No events scheduled</p>
                        <p class="text-xs text-gray-400">Add your first event to get started</p>
                    </div>
                } else {
                    <div class="space-y-4">
                        for _, event := range events {
                            <div class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors">
                                <div class="flex justify-between items-start">
                                    <div class="flex-1">
                                        <h3 class="text-lg font-medium text-gray-900">{ event.Name }</h3>
                                        <p class="text-sm text-gray-600 mt-1">{ event.Location }</p>
                                        <p class="text-xs text-gray-500 mt-2">
                                            { event.Date.Format("Monday, January 2, 2006 at 3:04 PM") }
                                        </p>
                                        if event.Description != "" {
                                            <p class="text-sm text-gray-600 mt-2">{ event.Description }</p>
                                        }
                                    </div>
                                </div>
                            </div>
                        }
                    </div>
                }
            </div>
        </div>
    </div>

    <!-- Add Event Modal -->
    <div x-show="showAddEventModal" 
         x-transition:enter="transition ease-out duration-300"
         x-transition:enter-start="opacity-0"
         x-transition:enter-end="opacity-100"
         x-transition:leave="transition ease-in duration-200"
         x-transition:leave-start="opacity-100"
         x-transition:leave-end="opacity-0"
         class="fixed inset-0 z-50 overflow-y-auto"
         style="display: none;">
        <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
            
            <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
                <form id="add-event-form" class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                    <div class="mb-4">
                        <label for="event-name" class="block text-sm font-medium text-gray-700">Event Name</label>
                        <input type="text" id="event-name" name="name" required 
                               class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>
                    
                    <div class="mb-4">
                        <label for="event-date" class="block text-sm font-medium text-gray-700">Date & Time</label>
                        <input type="datetime-local" id="event-date" name="date" required 
                               class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>
                    
                    <div class="mb-4">
                        <label for="event-location" class="block text-sm font-medium text-gray-700">Location</label>
                        <input type="text" id="event-location" name="location" 
                               class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    </div>
                    
                    <div class="mb-4">
                        <label for="event-description" class="block text-sm font-medium text-gray-700">Description</label>
                        <textarea id="event-description" name="description" rows="3" 
                                  class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"></textarea>
                    </div>
                </form>
                
                <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                    <button type="submit" form="add-event-form"
                            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
                        Add Event
                    </button>
                    <button @click="showAddEventModal = false"
                            class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
                        Cancel
                    </button>
                </div>
            </div>
        </div>
    </div>

    <script>
        document.getElementById('add-event-form').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = {
                name: formData.get('name'),
                date: new Date(formData.get('date')).toISOString(),
                location: formData.get('location'),
                description: formData.get('description')
            };

            fetch(`/api/events?band_id=${bandID}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    window.location.reload();
                } else {
                    alert('Failed to create event: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(error => {
                console.error('Error creating event:', error);
                alert('Error creating event');
            });
        });
    </script>
}
```

### 7. Creating Reusable Components

**Location**: `components/`

Create reusable UI components:

```templ
// event_card.templ
package components

import "github.com/nahue/setlist_manager/internal/app/shared/types"

templ EventCard(event *types.Event) {
    <div class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors">
        <div class="flex justify-between items-start">
            <div class="flex-1">
                <h3 class="text-lg font-medium text-gray-900">{ event.Name }</h3>
                <p class="text-sm text-gray-600 mt-1">{ event.Location }</p>
                <p class="text-xs text-gray-500 mt-2">
                    { event.Date.Format("Monday, January 2, 2006 at 3:04 PM") }
                </p>
                if event.Description != "" {
                    <p class="text-sm text-gray-600 mt-2">{ event.Description }</p>
                }
            </div>
            <div class="flex space-x-2">
                <button class="text-indigo-600 hover:text-indigo-900 text-sm font-medium">
                    Edit
                </button>
                <button class="text-red-600 hover:text-red-900 text-sm font-medium">
                    Delete
                </button>
            </div>
        </div>
    </div>
}
```

### 8. Updating Application Configuration

**Location**: `internal/app/application.go`

Add your new components to the application:

```go
// Add to Application struct
type Application struct {
    // ... existing fields ...
    eventsHandler *api.EventsHandler
}

// Add to NewApplication function
func NewApplication(
    db *database.Database,
    authStore *store.SQLiteAuthStore,
    bandsStore *store.SQLiteBandsStore,
    songsStore *store.SQLiteSongsStore,
    eventsStore *store.SQLiteEventsStore, // Add this
) *Application {
    // ... existing initialization ...
    
    eventsService := services.NewEventsService(eventsStore, bandsStore)
    eventsHandler := api.NewEventsHandler(eventsStore, eventsService)
    
    app := &Application{
        // ... existing fields ...
        eventsHandler: eventsHandler,
    }
    
    // ... rest of function
}

// Add to setupRoutes function
func (app *Application) setupRoutes() {
    // ... existing routes ...
    
    // Event routes
    r.Get("/events", app.eventsHandler.ServeEvents)
    r.Get("/api/events", app.eventsHandler.GetEvents)
    r.Post("/api/events", app.eventsHandler.CreateEvent)
}
```

**Location**: `main.go`

Update main.go to include your new store:

```go
func main() {
    // ... existing setup ...
    
    eventsStore := store.NewSQLiteEventsStore(db.GetDB())
    
    application := app.NewApplication(db, authStore, bandsStore, songsStore, eventsStore)
    
    // ... rest of function
}
```

### 9. Generate Templates

After creating new templates, regenerate the templ files:

```bash
task templ
```

### 10. Testing Your New Feature

1. **Start the development server**:
```bash
task dev
```

2. **Test the database migration**:
```bash
task db:migrate
```

3. **Test the API endpoints**:
```bash
# Test creating an event
curl -X POST http://localhost:9090/api/events?band_id=YOUR_BAND_ID \
  -H "Content-Type: application/json" \
  -d '{"name":"Gig at The Pub","date":"2025-08-15T20:00:00Z","location":"The Pub","description":"Friday night gig"}'

# Test getting events
curl http://localhost:9090/api/events?band_id=YOUR_BAND_ID
```

## Best Practices

### Database Design
- Always include `id`, `created_at`, `updated_at`, and `is_active` fields
- Use foreign keys for relationships
- Create appropriate indexes for performance
- Use consistent naming conventions

### API Design
- Follow RESTful conventions
- Always validate user permissions
- Return consistent JSON responses
- Handle errors gracefully

### Template Design
- Use the base layout for all pages
- Pass user context to all templates
- Use Alpine.js for interactive components
- Keep components reusable

### Security
- Always validate user authentication
- Check band membership for band-related operations
- Sanitize user inputs
- Use prepared statements for database queries

## Troubleshooting

### Common Issues

1. **Template not found**: Run `task templ` to regenerate template files
2. **Database errors**: Check migration status with `task db:status`
3. **Build errors**: Clean and rebuild with `task clean && task build`
4. **Hot reload not working**: Check `.air.toml` configuration

### Debugging

1. **Check logs**: The application logs to stdout
2. **Database inspection**: Use SQLite CLI to inspect the database
3. **Network requests**: Use browser dev tools to inspect API calls

## Contributing

1. Follow the established patterns for new features
2. Add appropriate tests for new functionality
3. Update documentation for new features
4. Ensure all templates are properly generated

## License

This project is open source and available under the MIT License. 