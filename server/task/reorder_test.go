package task

import (
	"testing"

	"github.com/google/uuid"
)

func TestReorderTask_KeepOrder(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	firstTask, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.ReorderTask(firstTask, 0)
	if err != nil {
		t.Fatalf("failed to reorder task with the same order it had: %v", err)
	}

	firstTask, err = taskService.taskDB.Get(firstTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if firstTask.Order != 0 {
		t.Fatalf("task should not have changed order, but it was moved to %d", firstTask.Order)
	}
}

func TestReorderTask_IncreaseOrder(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	firstTask, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.ReorderTask(firstTask, 1)
	if err != nil {
		t.Fatalf("failed to increase task's order: %v", err)
	}

	firstTask, err = taskService.taskDB.Get(firstTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if firstTask.Order != 1 {
		t.Fatalf("task should have increased order to 1, but it was %d", firstTask.Order)
	}
}

func TestReorderTask_DecreaseOrder(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	secondTask, err := taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.ReorderTask(secondTask, 0)
	if err != nil {
		t.Fatalf("failed to increase task's order: %v", err)
	}

	secondTask, err = taskService.taskDB.Get(secondTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if secondTask.Order != 0 {
		t.Fatalf("task should have decreased order to 0, but it was %d", secondTask.Order)
	}
}

func TestReorderTask_OrderOutOfBounds(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	secondTask, err := taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.ReorderTask(secondTask, -10)
	if err != nil {
		t.Fatalf("failed to increase task's order: %v", err)
	}

	secondTask, err = taskService.taskDB.Get(secondTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if secondTask.Order != 0 {
		t.Fatalf("task should have decreased order to 0, but it was %d", secondTask.Order)
	}
}

func TestReorderTask_TaskDoesNotExist(t *testing.T) {
	taskService, _ := setupServices()

	task := NewTask("Test task", uuid.New(), nil)
	err := taskService.ReorderTask(task, 0)
	if err == nil {
		t.Fatal("expected task reorder to fail, but it succeeded")
	}
}
