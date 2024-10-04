package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type App struct {
	Config           *oauth2.Config
	SessionStore     *sessions.CookieStore
	UserEmail        string
	EncryptionSecret string
	UUID             string
}

func main() {
	// OAuth2 client credentials
	clientID, err := RetrieveSecret("projects/your-project-id/secrets/CLIENT_ID/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve CLIENT_ID: %v", err)
	}
	clientSecret, err := RetrieveSecret("projects/your-project-id/secrets/CLIENT_SEC/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve CLIENT_SEC: %v", err)
	}
	redirectURL, err := RetrieveSecret("projects/your-project-id/secrets/REDIRECT_URL_DEV/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve REDIRECT_URL_DEV: %v", err)
	}
	sessionKey, err := RetrieveSecret("projects/your-project-id/secrets/SESSION_KEY/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve SESSION_KEY: %v", err)
	}
	// OAuth2 configuration
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	if len(sessionKey) == 0 {
		log.Fatal("SESSION_KEY variable is not set or empty")
	}
	store := sessions.NewCookieStore([]byte(sessionKey))
	gob.Register(map[string]interface{}{})

	app := App{Config: conf, SessionStore: store}
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/auth/oauth", app.OAuthHandler)
	mux.HandleFunc("/auth/callback", app.OAuthCallbackHandler)
	mux.HandleFunc("/api/user", app.UserInfoHandler)
	mux.HandleFunc("/auth/logout", app.LogoutHandler)
	mux.HandleFunc("/tasks", app.TasksHandler)
	mux.HandleFunc("/add-task", AddTaskHandler)
	mux.HandleFunc("/edit-task", EditTaskHandler)
	mux.HandleFunc("/modify-task", ModifyTaskHandler)
	mux.HandleFunc("/complete-task", CompleteTaskHandler)
	mux.HandleFunc("/delete-task", DeleteTaskHandler)

	log.Println("Server started at :8000")
	if err := http.ListenAndServe(":8000", app.EnableCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// RetrieveSecret gets the secret value from Secret Manager.
func RetrieveSecret(name string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx, option.WithCredentialsFile("path-to-your-service-account-key.json"))
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Access the secret.
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	// Return the secret payload.
	return string(result.Payload.Data), nil
}

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	url := a.Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// fetching the info
func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching user info...")

	code := r.URL.Query().Get("code")

	t, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := a.Config.Client(context.Background(), t)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email, okEmail := userInfo["email"].(string)
	id, okId := userInfo["id"].(string)
	if !okEmail || !okId {
		http.Error(w, "Unable to retrieve user info", http.StatusInternalServerError)
		return
	}
	uuidStr := GenerateUUID(email, id)
	encryptionSecret := GenerateEncryptionSecret(uuidStr, email, id)

	userInfo["uuid"] = uuidStr
	userInfo["encryption_secret"] = encryptionSecret
	session, _ := a.SessionStore.Get(r, "session-name")
	session.Values["user"] = userInfo
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("User Info: %v", userInfo)

	frontendOriginDev, err := RetrieveSecret("projects/your-project-id/secrets/FRONTEND_ORIGIN_DEV/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve FRONTEND_ORIGIN_DEV: %v", err)
	}
	http.Redirect(w, r, frontendOriginDev+"/home", http.StatusSeeOther)
}

func (a *App) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.SessionStore.Get(r, "session-name")
	userInfo, ok := session.Values["user"].(map[string]interface{})
	if !ok || userInfo == nil {
		http.Error(w, "No user info available", http.StatusUnauthorized)
		return
	}

	log.Printf("Sending User Info: %v", userInfo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

func (a *App) EnableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin, err := RetrieveSecret("projects/your-project-id/secrets/FRONTEND_ORIGIN_DEV/versions/latest")
		if err != nil {
			log.Fatalf("Failed to retrieve FRONTEND_ORIGIN_DEV: %v", err)
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // to allow credentials
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// helps to fetch tasks using '/tasks' route
func (a *App) TasksHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	encryptionSecret := r.URL.Query().Get("encryptionSecret")
	UUID := r.URL.Query().Get("UUID")
	origin, err := RetrieveSecret("projects/your-project-id/secrets/CONTAINER_ORIGIN/versions/latest")
	if err != nil {
		log.Fatalf("Failed to retrieve CONTAINER_ORIGIN: %v", err)
	}
	if email == "" || encryptionSecret == "" || UUID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		tasks, _ := FetchTasksFromTaskwarrior(email, encryptionSecret, origin, UUID)
		if tasks == nil {
			http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var requestBody struct {
			Email            string `json:"email"`
			EncryptionSecret string `json:"encryptionSecret"`
			UUID             string `json:"UUID"`
			TaskUUID         string `json:"taskuuid"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		encryptionSecret := requestBody.EncryptionSecret
		uuid := requestBody.UUID
		taskuuid := requestBody.TaskUUID

		if taskuuid == "" {
			http.Error(w, "taskuuid is required", http.StatusBadRequest)
			return
		}

		if err := DeleteTaskInTaskwarrior(email, encryptionSecret, uuid, taskuuid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func CompleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		fmt.Printf("Raw request body: %s\n", string(body))

		var requestBody struct {
			Email            string `json:"email"`
			EncryptionSecret string `json:"encryptionSecret"`
			UUID             string `json:"UUID"`
			TaskUUID         string `json:"taskuuid"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		encryptionSecret := requestBody.EncryptionSecret
		uuid := requestBody.UUID
		taskuuid := requestBody.TaskUUID

		if taskuuid == "" {
			http.Error(w, "taskuuid is required", http.StatusBadRequest)
			return
		}

		if err := CompleteTaskInTaskwarrior(email, encryptionSecret, uuid, taskuuid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func EditTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		fmt.Printf("Raw request body: %s\n", string(body))

		var requestBody struct {
			Email            string `json:"email"`
			EncryptionSecret string `json:"encryptionSecret"`
			UUID             string `json:"UUID"`
			TaskUUID         string `json:"taskuuid"`
			Description      string `json:"description"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		encryptionSecret := requestBody.EncryptionSecret
		uuid := requestBody.UUID
		taskuuid := requestBody.TaskUUID
		description := requestBody.Description

		if taskuuid == "" {
			http.Error(w, "taskuuid is required", http.StatusBadRequest)
			return
		}

		if err := EditTaskInTaskwarrior(uuid, description, email, encryptionSecret, taskuuid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func ModifyTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		fmt.Printf("Raw request body: %s\n", string(body))

		var requestBody struct {
			Email            string `json:"email"`
			EncryptionSecret string `json:"encryptionSecret"`
			UUID             string `json:"UUID"`
			TaskUUID         string `json:"taskuuid"`
			Description      string `json:"description"`
			Project          string `json:"project"`
			Priority         string `json:"priority"`
			Status           string `json:"status"`
			Due              string `json:"due"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
			return
		}

		email := requestBody.Email
		encryptionSecret := requestBody.EncryptionSecret
		uuid := requestBody.UUID
		taskuuid := requestBody.TaskUUID
		description := requestBody.Description
		project := requestBody.Project
		priority := requestBody.Priority
		status := requestBody.Status
		due := requestBody.Due

		if taskuuid == "" {
			http.Error(w, "taskuuid is required", http.StatusBadRequest)
			return
		}

		if err := ModifyTaskInTaskwarrior(uuid, description, project, priority, status, due, email, encryptionSecret, taskuuid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		fmt.Printf("Raw request body: %s\n", string(body))

		var requestBody struct {
			Email            string `json:"email"`
			EncryptionSecret string `json:"encryptionSecret"`
			UUID             string `json:"UUID"`
			Description      string `json:"description"`
			Project          string `json:"project"`
			Priority         string `json:"priority"`
			DueDate          string `json:"due"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
			return
		}
		email := requestBody.Email
		encryptionSecret := requestBody.EncryptionSecret
		uuid := requestBody.UUID
		description := requestBody.Description
		project := requestBody.Project
		priority := requestBody.Priority
		dueDate := requestBody.DueDate

		if description == "" {
			http.Error(w, "description is required", http.StatusBadRequest)
			return
		}

		if err := AddTaskToTaskwarrior(email, encryptionSecret, uuid, description, project, priority, dueDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

// logout and delete session
func (a *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.SessionStore.Get(r, "session-name")
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Print("User has logged out")
}
