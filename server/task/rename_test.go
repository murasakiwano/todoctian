package task

import (
	"testing"

	"github.com/google/uuid"
)

func TestRenameTask_Success(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}

	task, err := taskService.CreateTask("My test task", testProject.ID, nil)
	if err != nil {
		t.Fatalf("expected task creation to succeed, but an error occurred: %s", err)
	}

	newTaskName := "My new test task"
	err = taskService.RenameTask(task.ID, newTaskName)
	if err != nil {
		t.Fatalf("expected task renaming to work, but an error occurred: %v", err)
	}

	task, _ = taskService.taskDB.Get(task.ID)
	if task.Name != newTaskName {
		t.Fatal("task was not renamed as expected")
	}
}

func TestRenameTask_TaskDoesNotExist(t *testing.T) {
	taskService, _ := setupServices()

	taskID := uuid.New()
	err := taskService.RenameTask(taskID, "New task name")
	if err == nil {
		t.Fatal("expected renaming an inexistent task to fail, but it did not")
	}
}
