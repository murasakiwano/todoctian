package task

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/project"
)

func TestCreateTask_Success(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}

	_, err = taskService.CreateTask("My test task", testProject.ID, nil)
	if err != nil {
		t.Fatalf("expected task creation to succeed, but got error: %v", err)
	}
}

func TestCreateTask_UpdatesParentTask(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}

	task, err := taskService.CreateTask("My test task", testProject.ID, nil)
	if err != nil {
		t.Fatalf("expected task creation to succeed, but got error: %v", err)
	}

	subtask, err := taskService.CreateTask("My subtask", testProject.ID, &task.ID)
	if err != nil {
		t.Fatalf("expected subtask creation to succeed, but it failed: %v", err)
	}

	task, err = taskService.taskDB.Get(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(task.SubtaskIDs) == 0 || task.SubtaskIDs[0] != subtask.ID {
		t.Fatal("adding a subtask to a task did not successfully update the parent task")
	}

	subtask, err = taskService.CreateTask("My second subtask", testProject.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, err = taskService.taskDB.Get(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(task.SubtaskIDs) == 1 || task.SubtaskIDs[1] != subtask.ID {
		t.Fatal("adding a second subtask to a task did not successfully update the parent task")
	}
}

func TestCreateTask_ProjectDoesNotExist(t *testing.T) {
	taskService := NewTaskService(NewTaskRepositoryInMemory(), project.NewProjectRepositoryInMemory())

	_, err := taskService.CreateTask("My test task", uuid.New(), nil)
	if err == nil {
		t.Fatal("expected task creation without an existing project to fail, which didn't happen.")
	}
}

func TestCreateTask_ParentTaskIsInvalid(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}

	parentID := uuid.New()
	fmt.Printf("ParentID: %s\n", parentID)
	_, err = taskService.CreateTask("My test task", testProject.ID, &parentID)
	if err == nil {
		t.Fatal("expected task creation with an invalid parent task to fail")
	}
}

func TestCreateTask_SetsOrderCorrectly(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if task.Order != 0 {
		t.Fatalf("expected first task to have order 0, it actually had %d", task.Order)
	}

	task, err = taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if task.Order != 1 {
		t.Fatalf("expected second task to have order 1, it actually had %d", task.Order)
	}
}

func TestCreateTask_SubtaskDoesNotAffectParentTaskOrder(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	parentTask, err := taskService.CreateTask("Parent task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if parentTask.Order != 0 {
		t.Fatalf("expected parent task to have order 0, it actually had %d", parentTask.Order)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &parentTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Order != 0 {
		t.Fatalf("expected subtask to have order 0, it actually had %d", parentTask.Order)
	}

	parentTask, err = taskService.taskDB.Get(parentTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if parentTask.Order != 0 {
		t.Fatalf("expected parent task's order to remain 0, it actually was %d", parentTask.Order)
	}
}
