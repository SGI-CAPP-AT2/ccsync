package main

import (
	"ccsync_backend/auth"
	"ccsync_backend/handlers"
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// OAuth2 client credentials
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SEC")
	redirectURL := os.Getenv("REDIRECT_URL_DEV")

	// OAuth2 configuration
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// Create a session store
	sessionKey := []byte(os.Getenv("SESSION_KEY"))
	if len(sessionKey) == 0 {
		log.Fatal("SESSION_KEY environment variable is not set or empty")
	}
	store := sessions.NewCookieStore(sessionKey)
	gob.Register(map[string]interface{}{})

	app := auth.App{Config: conf, SessionStore: store}
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/auth/oauth", app.OAuthHandler)
	mux.HandleFunc("/auth/callback", app.OAuthCallbackHandler)
	mux.HandleFunc("/api/user", app.UserInfoHandler)
	mux.HandleFunc("/auth/logout", app.LogoutHandler)
	mux.HandleFunc("/tasks", handlers.TasksHandler)
	mux.HandleFunc("/add-task", handlers.AddTaskHandler)
	mux.HandleFunc("/edit-task", handlers.EditTaskHandler)
	mux.HandleFunc("/modify-task", handlers.ModifyTaskHandler)
	mux.HandleFunc("/complete-task", handlers.CompleteTaskHandler)
	mux.HandleFunc("/delete-task", handlers.DeleteTaskHandler)

	log.Println("Server started at :8000")
	if err := http.ListenAndServe(":8000", app.EnableCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
