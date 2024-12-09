package task

import (
	"testing"
)

func TestUpdateTaskStatus_CompleteWithoutSubtasks(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("expected task to be marked completed, but an error occurred: %v", err)
	}

	task, err = taskService.FindTaskByID(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}
}

func TestUpdateTaskStatus_CompleteWithSubtasks(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSubtask, err := taskService.CreateTask("Nested subtask", project.ID, &subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSiblingSubstask, err := taskService.CreateTask("Nested sibling subtask", project.ID, &subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("expected task to be marked completed, but an error occurred: %v", err)
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusCompleted {
		t.Fatal("subtask was not marked completed successfully")
	}

	nestedSubtask, err = taskService.FindTaskByID(nestedSubtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSubtask.Status != TaskStatusCompleted {
		t.Fatal("nestedSubtask was not marked completed successfully")
	}

	nestedSiblingSubstask, err = taskService.FindTaskByID(nestedSiblingSubstask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSiblingSubstask.Status != TaskStatusCompleted {
		t.Fatal("nestedSiblingSubtask was not marked completed successfully")
	}

	task, err = taskService.FindTaskByID(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}
}

func TestUpdateTaskStatus_CompleteAlsoCompletesTheParentTask(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSubtask, err := taskService.CreateTask("Nested subtask", project.ID, &subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSiblingSubstask, err := taskService.CreateTask("Nested sibling subtask", project.ID, &subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(nestedSubtask.ID)
	if err != nil {
		t.Fatalf("expected nestedSubtask to be marked completed, but an error occurred: %v", err)
	}

	err = taskService.CompleteTask(nestedSiblingSubstask.ID)
	if err != nil {
		t.Fatalf("expected nestedSiblingSubtask to be marked completed, but an error occurred: %v", err)
	}

	nestedSubtask, err = taskService.FindTaskByID(nestedSubtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSubtask.Status != TaskStatusCompleted {
		t.Fatal("nestedSubtask was not marked completed successfully")
	}

	nestedSiblingSubstask, err = taskService.FindTaskByID(nestedSiblingSubstask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSiblingSubstask.Status != TaskStatusCompleted {
		t.Fatal("nestedSiblingSubtask was not marked completed successfully")
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusCompleted {
		t.Fatal("subtask was not marked completed successfully")
	}

	task, err = taskService.FindTaskByID(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}
}

func TestUpdateTaskStatus_CompleteTaskIsAlreadyCompleted(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}
}

func TestUpdateTaskStatus_Pending(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}

	err = taskService.MarkTaskAsPending(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusPending {
		t.Fatal("task was not marked as pending successfully")
	}
}

func TestUpdateTaskStatus_PendingDoesNotMarkSubtasksAsPending(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusCompleted {
		t.Fatal("subtask was not completed successfully")
	}

	err = taskService.MarkTaskAsPending(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusPending {
		t.Fatal("task was not marked as pending successfully")
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusCompleted {
		t.Fatal("subtask was accidentally marked as pending")
	}
}

func TestUpdateTaskStatus_PendingMarksParentAsPending(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	task, err := taskService.CreateTask("First task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	subtask, err := taskService.CreateTask("Subtask", project.ID, &task.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSubtask, err := taskService.CreateTask("Nested subtask", project.ID, &subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = taskService.CompleteTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusCompleted {
		t.Fatal("task was not completed successfully")
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusCompleted {
		t.Fatal("subtask was not completed successfully")
	}

	nestedSubtask, err = taskService.FindTaskByID(nestedSubtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSubtask.Status != TaskStatusCompleted {
		t.Fatal("nestedSubtask was not completed successfully")
	}

	err = taskService.MarkTaskAsPending(nestedSubtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	nestedSubtask, err = taskService.FindTaskByID(nestedSubtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if nestedSubtask.Status != TaskStatusPending {
		t.Fatal("nestedSubtask was not marked as pending successfully")
	}

	subtask, err = taskService.FindTaskByID(subtask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if subtask.Status != TaskStatusPending {
		t.Fatal("subtask was not marked as pending successfully")
	}

	task, _ = taskService.FindTaskByID(task.ID)
	if task.Status != TaskStatusPending {
		t.Fatal("task was not marked as pending successfully")
	}
}
