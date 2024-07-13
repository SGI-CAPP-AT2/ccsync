package utils_test

import (
	"ccsync_backend/utils"
	"fmt"
	"testing"
)

func TestSetTaskwarriorConfig(t *testing.T) {
	encryptionSecret := ""
	container_origin := ""
	clientId := ""
	err := utils.SetTaskwarriorConfig("./", encryptionSecret, container_origin, clientId)
	if err != nil {
		t.Errorf("SetTaskwarriorConfig() failed: %v", err)
	} else {
		fmt.Println("SetTaskwarriorConfig test passed")
	}
}
func TestSyncTaskwarrior(t *testing.T) {
	err := utils.SyncTaskwarrior("./")
	if err != nil {
		t.Errorf("SyncTaskwarrior failed: %v", err)
	} else {
		fmt.Println("Sync Dir test passed")
	}
}

func TestEditTaskInATaskwarrior(t *testing.T) {
	encryptionSecret := ""
	clientId := ""
	err := utils.EditTaskInTaskwarrior(clientId, "description", "email", encryptionSecret, "")
	if err != nil {
		t.Errorf("EditTaskInTaskwarrior() failed: %v", err)
	} else {
		fmt.Println("Edit test passed")
	}
}

func TestExportTasks(t *testing.T) {
	task, err := utils.ExportTasks("./")
	if task != nil && err == nil {
		fmt.Println("Task export test passed")
	} else {
		t.Errorf("ExportTasks() failed: %v", err)
	}
}

func TestAddTaskToTaskwarrior(t *testing.T) {
	encryptionSecret := ""
	clientId := ""
	err := utils.AddTaskToTaskwarrior("email", encryptionSecret, clientId, "description", "project", "H", "2025-03-03")
	if err != nil {
		t.Errorf("AddTaskToTaskwarrior failed: %v", err)
	} else {
		fmt.Println("Add task passed")
	}
}

func TestCompleteTaskInTaskwarrior(t *testing.T) {
	encryptionSecret := ""
	clientId := ""
	err := utils.CompleteTaskInTaskwarrior("email", encryptionSecret, clientId, "")
	if err != nil {
		t.Errorf("CompleteTaskInTaskwarrior failed: %v", err)
	} else {
		fmt.Println("Complete task passed")
	}
}

func TestDeleteTaskInTaskwarrior(t *testing.T) {
	encryptionSecret := ""
	clientId := ""
	err := utils.DeleteTaskInTaskwarrior("email", encryptionSecret, clientId, "")
	if err != nil {
		t.Errorf("DeleteTaskInTaskwarrior failed: %v", err)
	} else {
		fmt.Println("Delete task passed")
	}
}
