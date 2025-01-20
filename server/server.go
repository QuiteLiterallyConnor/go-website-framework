package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/QuiteLiterallyConnor/go-website-framework/auth"

	"github.com/google/uuid"
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
)

// Server struct remains the same
type Server struct {
	Clients     map[string]*websocket.Conn // Tracks WebSocket connections by session_uuid
	ClientsLock sync.Mutex                 // Mutex to manage concurrent access to Clients map
	Router      *pat.Router                // Pat router for HTTP serving
	Auth        *auth.AuthHandler          // Auth instance for handling authentication
}

func NewServer() *Server {
	return &Server{
		Clients: make(map[string]*websocket.Conn),
		Router:  pat.New(),
	}
}

// Function to generate a new unique session_uuid (UUID)
func generateSessionUUID() string {
	return uuid.New().String()
}

// WebSocket communication and broadcast logic remains the same...

// OnReceiveWebsocket sets up a handler function that gets called whenever a WebSocket message is received
func (s *Server) OnReceiveWebsocket(handler func(session_uuid string, msg string) error) {
	// WebSocket route handling remains the same...
}

// AuthRoutes now uses the AuthHandler to set up the routes
func (s *Server) AuthRoutes(port string) {
	// Setup OAuth providers (Google and Discord)
	s.Auth.SetupProviders()

	// Set up authentication-related routes
	s.Auth.AuthRoutes(s.Router)

	// Serve the authentication routes
	log.Printf("listening on localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, s.Router))
}

// StaticFile serves a single file from the server with a specific route
func (s *Server) StaticFile(route, filePath string) {
	s.Router.Get(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}))
}

// StaticFiles serves multiple files by mapping routes to file paths
func (s *Server) StaticFiles(fileMapping map[string]string) {
	for route, filePath := range fileMapping {
		s.StaticFile(route, filePath)
	}
}

func (s *Server) StaticDirectory(directory string) {
	s.Router.Get("/static/{file:.*}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the file path from the request
		requestedFile := r.URL.Path[len("/static/"):]
		fullPath := directory + "/" + requestedFile

		// Ensure the path is not treated as a directory
		if len(requestedFile) > 0 && requestedFile[len(requestedFile)-1] == '/' {
			http.NotFound(w, r) // Return 404 for trailing slash
			return
		}

		// Serve the file
		http.ServeFile(w, r, fullPath)
	}))
}

// Serve starts the HTTP server
func (s *Server) Serve(port string) {
	log.Fatal(http.ListenAndServe(port, s.Router))
}
