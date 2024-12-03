package main

import "testing"

func TestTaskCreation(t *testing.T) {
	project := CreateProject("Test Project")
	CreateTask("Test task", project)
}

func TestAddSubtask(t *testing.T) {
	project := CreateProject("Test Project")
	task := CreateTask("Test task", project)
	subtask := CreateTask("Test subtask", project)
	task.AddSubtask(&subtask)

	if task.subtasks[0].TaskUUID != subtask.TaskUUID {
		t.Errorf("Tasks UUID do not match. Expected %s, got %s.", task.subtasks[0].TaskUUID, subtask.TaskUUID)
	}
}

func TestTaskCompletion(t *testing.T) {
	project := CreateProject("Test Project")
	task := CreateTask("Test task", project)
	task.CompleteTask()

	if task.Status != Completed {
		t.Errorf("Could not complete task! It's in %s status", task.Status)
	}
}

func TestTaskCompletionWithSubtasks(t *testing.T) {
	project := CreateProject("Test Project")
	parentTask := CreateTask("Test task", project)
	firstSubtask := CreateTask("First subtask", project)
	secondSubtask := CreateTask("Second subtask", project)
	parentTask.AddSubtask(&firstSubtask)
	parentTask.AddSubtask(&secondSubtask)

	parentTask.CompleteTask()

	if firstSubtask.Status != Completed {
		t.Errorf("Failed to complete first subtask! Task status is %s", firstSubtask.Status)
	}

	if secondSubtask.Status != Completed {
		t.Errorf("Failed to complete second subtask! Task status is %s", secondSubtask.Status)
	}
}

func TestParentTaskCompletion(t *testing.T) {
	project := CreateProject("Test Project")
	parentTask := CreateTask("Parent task", project)
	subtask := CreateTask("Subtask", project)

	parentTask.AddSubtask(&subtask)
	subtask.CompleteTask()

	if parentTask.Status != Completed {
		t.Errorf("Failed to complete parent task! Its status is %s", parentTask.Status)
	}
}
