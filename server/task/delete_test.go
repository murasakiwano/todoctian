package task

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
)

func TestDeleteTask_Success(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("Test task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.DeleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, err = taskService.FindTaskByID(task.ID)
	if err == nil || !errors.Is(err, internal.ErrNotFound) {
		t.Fatalf("task was returned when it should have been deleted: %v", task)
	}
}

func TestDeleteTask_TaskDoesNotExist(t *testing.T) {
	taskService, _ := setupServices()

	id := uuid.New()
	_, err := taskService.DeleteTask(id)
	if err != nil {
		t.Fatalf(
			"deleting a nonexistent task should be a no-op, but it returned an error: %v",
			err,
		)
	}
}

func TestDeleteTask_AlsoDeletesSubtasks(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("Test task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.DeleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, err = taskService.FindTaskByID(task.ID)
	if err == nil || !errors.Is(err, internal.ErrNotFound) {
		t.Fatalf("task was returned when it should have been deleted: %v", task)
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err == nil {
		t.Fatalf("subtask was returned when it should have been deleted: %v", task)
	}

	if !errors.Is(err, internal.ErrNotFound) {
		t.Fatalf("an unexpected error occurred: %v", err)
	}
}

func TestDeleteTask_RearrangesSiblingsOrders(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test project")
	if err != nil {
		t.Fatal(err)
	}

	firstTask, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	secondTask, err := taskService.CreateTask("Second task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	thirdTask, err := taskService.CreateTask("Third task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.DeleteTask(secondTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	firstTask, _ = taskService.FindTaskByID(firstTask.ID)
	thirdTask, _ = taskService.FindTaskByID(thirdTask.ID)

	if firstTask.Order != 0 {
		t.Fatalf("firstTask's order changed when it should not have changed: %v", firstTask.Order)
	}

	if thirdTask.Order != 1 {
		t.Fatalf("thirdTasks's order should have decreased to 1, but it's %v", thirdTask.Order)
	}
}

func TestDeleteTask_KeepsParentTaskTheSame(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("Test task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.DeleteTask(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.FindTaskByID(task.ID)
	if err != nil {
		t.Fatalf("Parent task was supposed to remain intact, but an error occurred: %v", err)
	}
}
