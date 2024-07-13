package handlers

import (
	"ccsync_backend/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// helps to fetch tasks using '/tasks' route
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	encryptionSecret := r.URL.Query().Get("encryptionSecret")
	UUID := r.URL.Query().Get("UUID")
	origin := os.Getenv("CONTAINER_ORIGIN")

	if email == "" || encryptionSecret == "" || UUID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		tasks, _ := utils.FetchTasksFromTaskwarrior(email, encryptionSecret, origin, UUID)
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

		if err := utils.DeleteTaskInTaskwarrior(email, encryptionSecret, uuid, taskuuid); err != nil {
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

		if err := utils.CompleteTaskInTaskwarrior(email, encryptionSecret, uuid, taskuuid); err != nil {
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

		if err := utils.EditTaskInTaskwarrior(uuid, description, email, encryptionSecret, taskuuid); err != nil {
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

		if err := utils.ModifyTaskInTaskwarrior(uuid, description, project, priority, status, due, email, encryptionSecret, taskuuid); err != nil {
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

		if err := utils.AddTaskToTaskwarrior(email, encryptionSecret, uuid, description, project, priority, dueDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
