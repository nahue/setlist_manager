# People Management System

A Go REST API built with the Chi library featuring a modern web interface using Alpine.js for dynamic interactions.

## Features

- **REST API**: Built with Chi router for clean, fast routing
- **Health Check**: `/health` endpoint for monitoring system status
- **Welcome Page**: Simple welcome page for the Setlist Manager application
- **Modern UI**: Beautiful, responsive interface with Alpine.js
- **Authentication**: Magic link authentication system

## API Endpoints

### Health Check
```
GET /health
```
Returns the system health status and timestamp.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-06T19:58:16.706922-03:00"
}
```



## Pages

### Home Page
```
GET /
```
Serves the welcome page for the Setlist Manager application.

## Frontend Features

The web interface includes:

- **Modern Design**: Clean, responsive design with Tailwind CSS
- **Welcome Page**: Simple and elegant welcome message
- **Navigation**: Clean navigation with authentication support
- **Error Handling**: Graceful error display and loading states

## Getting Started

### Prerequisites
- Go 1.24.4 or higher

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd pr-toolbox-go
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables (optional):
   
   **Option A: Using .env file (recommended)**
   ```bash
   cp .env.example .env
   # Edit .env file and add your configuration
   ```
   
   **Option B: Using environment variables**
   ```bash
   # No required environment variables for basic functionality
   ```

4. Run the application:

   **Option A: Using go run**
   ```bash
   go run .
   ```

   **Option B: Using Makefile (recommended)**
   ```bash
   make run
   ```

   **Option C: Development mode (with templ generation)**
   ```bash
   make dev
   ```

   **Option D: Hot reload development with Air**
   ```bash
   make air
   ```
   This will automatically reload the application when you modify any Go or templ files.

5. Open your browser and navigate to:
```
http://localhost:9090
```

## Usage

1. **Welcome Page**: View the main welcome page with application information
2. **Health Check**: Visit `/health` to verify the API is running properly
3. **Authentication**: Use the authentication system to log in and access protected routes

## Project Structure

```
setlist_manager/
├── main.go              # Main application file with server setup
├── go.mod               # Go module dependencies
├── Taskfile.yml         # Build and development tasks
├── .env.example         # Example environment variables file
├── templates/           # Templ templates
│   ├── layout.templ      # Base layout template
│   └── index.templ       # Welcome page template
└── README.md            # This documentation
```

## Technologies Used

- **Backend**: Go with Chi router
- **Frontend**: Alpine.js for reactive UI
- **Templating**: Templ for type-safe HTML templates
- **Configuration**: Godotenv for environment variable management
- **Styling**: Tailwind CSS with modern design patterns
- **Middleware**: CORS support, logging, and error recovery

## Development

The application is structured for easy extension:

- Add new API endpoints in the main router
- Create new templ templates for additional pages using the base layout
- Add new Alpine.js functions for additional frontend functionality
- Build setlist management features as needed

## Environment Variables

The application supports loading environment variables from a `.env` file. Copy `.env.example` to `.env` and configure the following variables:

### Required Variables
- None (all variables are optional)

### Optional Variables
- `PORT`: Server port (default: 9090)
- `LOG_LEVEL`: Logging level (default: info)

## Running the Application

The application has multiple Go files, so you need to run it using:

```bash
go run .
```

This will compile and run all Go files in the project.

## API Testing

Test the API endpoints using curl:

```bash
# Health check
curl http://localhost:8080/health

# Health check
curl http://localhost:9090/health
```

## License

This project is open source and available under the MIT License. 