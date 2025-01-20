package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions" // Import the gorilla sessions package
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var (
	// Store is the session store that gothic uses to save sessions.
	// Replace with a secure key in a production environment.
	store = sessions.NewCookieStore([]byte("your-secret-key"))
)

// AuthHandler stores the configuration for authentication.
type AuthHandler struct {
	Url string
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

// NewAuth creates a new AuthHandler instance.
func NewAuth(url string) *AuthHandler {
	a := &AuthHandler{
		Url: url,
	}
	a.SetupProviders()
	return a
}

// SetupProviders sets up OAuth providers (Google and Discord).
func (a *AuthHandler) SetupProviders() {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), a.Url+"/auth/google/callback"),
		// discord.New(os.Getenv("DISCORD_KEY"), os.Getenv("DISCORD_SECRET"), a.Url+"/auth/discord/callback", discord.ScopeIdentify, discord.ScopeEmail),
	)

	fmt.Printf("callback google: %s\n", a.Url+"/auth/google/callback")
	// fmt.Printf("callback discord: %s\n", a.Url+"/auth/discord/callback")
}

// AuthRoutes sets up the authentication routes for the server.
func (a *AuthHandler) AuthRoutes(router *pat.Router) {
	// Set the session store for gothic
	gothic.Store = store

	// Callback route for Google
	router.Get("/auth/google/callback", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		// Serve profile page from static file (similar logic as server.go)
		profileTemplate, err := template.ParseFiles("static/profile.html")
		if err != nil {
			log.Println("Error loading profile template:", err)
			http.Error(res, "Error loading template", http.StatusInternalServerError)
			return
		}
		profileTemplate.Execute(res, user)
	}))

	// Login page route
	router.Get("/auth/{provider}", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			// Serve profile page from static file
			profileTemplate, err := template.ParseFiles("static/profile.html")
			if err != nil {
				log.Println("Error loading profile template:", err)
				http.Error(res, "Error loading template", http.StatusInternalServerError)
				return
			}
			profileTemplate.Execute(res, gothUser)
		} else {
			gothic.BeginAuthHandler(res, req)
		}
	}))

	// Logout handler
	router.Get("/logout/{provider}", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	}))

	// Login options page
	router.Get("/login", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		indexTemplate, err := template.ParseFiles("static/login.html")
		if err != nil {
			log.Println("Error loading index template:", err)
			http.Error(res, "Error loading template", http.StatusInternalServerError)
			return
		}
		// Provide available OAuth providers to the page
		providerIndex := &ProviderIndex{
			Providers:    []string{"google", "discord"},
			ProvidersMap: map[string]string{"google": "Google", "discord": "Discord"},
		}
		indexTemplate.Execute(res, providerIndex)
	}))

	// Main index route
	router.Get("/", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Check if the user is logged in
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			// Render profile info if logged in
			profileTemplate, err := template.ParseFiles("static/index.html")
			if err != nil {
				log.Println("Error loading profile template:", err)
				http.Error(res, "Error loading template", http.StatusInternalServerError)
				return
			}
			profileTemplate.Execute(res, gothUser)
		} else {
			// Serve the page as usual if not logged in
			indexTemplate, err := template.ParseFiles("static/index.html")
			if err != nil {
				log.Println("Error loading index template:", err)
				http.Error(res, "Error loading template", http.StatusInternalServerError)
				return
			}
			indexTemplate.Execute(res, nil)
		}
	}))
}
