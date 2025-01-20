package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"text/template"

	"go-website-framework/auth" // import the auth package

	"github.com/google/uuid"
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/openidConnect"
)

type Server struct {
	Clients     map[string]*websocket.Conn // Exported field, capitalized
	ClientsLock sync.Mutex                 // Exported field, capitalized
	Router      *pat.Router                // Exported field, capitalized
	Auth        *auth.AuthHandler          // Exported field, capitalized
}

// NewServer creates a new Server instance
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

// sendWebsocket sends a WebSocket message to a specific session identified by session_uuid
func (s *Server) SendWebsocket(session_uuid, message string) error {
	s.ClientsLock.Lock()
	defer s.ClientsLock.Unlock()

	// Find the WebSocket connection for the session_uuid
	conn, ok := s.Clients[session_uuid]
	if !ok {
		return fmt.Errorf("no active WebSocket connection found for session_uuid: %s", session_uuid)
	}

	// Send the message to the WebSocket connection
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message to session %s: %v", session_uuid, err)
	}

	return nil
}

// websocketBroadcast sends a message to all connected WebSocket clients
func (s *Server) WebsocketBroadcast(message string) {
	s.ClientsLock.Lock()
	defer s.ClientsLock.Unlock()

	// Iterate over all connected clients and send the message
	for session_uuid, conn := range s.Clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Failed to send message to session %s: %v\n", session_uuid, err)
		}
	}
}

// onReceiveWebsocket sets up a handler function that gets called whenever a WebSocket message is received
func (s *Server) OnReceiveWebsocket(handler func(session_uuid string, msg string) error) { // Exported method, capitalized
	// WebSocket route handling
	s.Router.Get("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Generate a unique session_uuid for the new connection
		session_uuid := generateSessionUUID()

		// Log that a new user has connected
		fmt.Printf("New WebSocket connection from %s\n", session_uuid)

		// Store the connection in the clients map
		s.ClientsLock.Lock()
		s.Clients[session_uuid] = conn
		s.ClientsLock.Unlock()

		// Continuously listen for messages from the WebSocket
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading WebSocket message:", err)
				break
			}

			// Call the handler function when a message is received
			if handler != nil {
				if err := handler(session_uuid, string(msg)); err != nil {
					log.Println("Handler error:", err)
				}
			}
		}

		// Clean up the connection after the session ends
		s.ClientsLock.Lock()
		delete(s.Clients, session_uuid)
		s.ClientsLock.Unlock()

		// Log that the user has disconnected
		fmt.Printf("WebSocket connection closed for %s\n", session_uuid)
	}))
}

func (s *Server) AuthRoutes() {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:3000/auth/google/callback"),
	)

	// OpenID Connect is based on OpenID Connect Auto Discovery URL (https://openid.net/specs/openid-connect-discovery-1_0-17.html)
	// because the OpenID Connect provider initialize itself in the New(), it can return an error which should be handled or ignored
	// ignore the error for now
	openidConnect, _ := openidConnect.New(os.Getenv("OPENID_CONNECT_KEY"), os.Getenv("OPENID_CONNECT_SECRET"), "http://localhost:3000/auth/openid-connect/callback", os.Getenv("OPENID_CONNECT_DISCOVERY_URL"))
	if openidConnect != nil {
		goth.UseProviders(openidConnect)
	}

	m := map[string]string{
		"google": "Google",
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	s.Router.Get("/auth/{provider}/callback", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(res, user)
	}))

	s.Router.Get("/logout/{provider}", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	}))

	s.Router.Get("/auth/{provider}", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(res, gothUser)
		} else {
			gothic.BeginAuthHandler(res, req)
		}
	}))

	s.Router.Get("/", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.New("foo").Parse(indexTemplate)
		t.Execute(res, providerIndex)
	}))

	log.Println("listening on localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", s.Router))
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

// StaticDirectory serves all files from a directory
func (s *Server) StaticDirectory(directory string) {
	// This converts the http.Handler returned by http.FileServer into an http.HandlerFunc
	s.Router.Get("/static/{file}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", http.FileServer(http.Dir(directory))).ServeHTTP(w, r)
	}))
}

// Serve starts the HTTP server
func (s *Server) Serve(port string) {
	log.Fatal(http.ListenAndServe(port, s.Router))
}

// ProviderIndex is a struct for holding the provider information
type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
