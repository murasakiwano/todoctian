package task

import (
	"github.com/google/uuid"
)

type TaskRepository interface {
	// Save a task to the database
	Save(task Task) error

	// Retrieve a task by its ID
	Get(id uuid.UUID) (Task, error)

	// Retrieve all direct children/subtasks of a specific task
	GetSubtasksDirect(id uuid.UUID) ([]Task, error)

	// Recursively retrieve all subtasks of a specific task
	GetSubtasksDeep(id uuid.UUID) ([]Task, error)

	// Retrieve all tasks in a specific project
	GetTasksByProject(projectID uuid.UUID) []Task

	// Retrieve all tasks in a project
	GetTasksInProjectRoot(projectID uuid.UUID) []Task

	// Filter tasks in a project by their status
	GetTasksByStatus(projectID uuid.UUID, status TaskStatus) []Task

	// Fuzzy search for task names
	SearchFuzzy(partialTaskName string) ([]Task, error)

	// Update a single task
	Update(task Task) error

	// Batch update a collection of tasks
	BatchUpdate(tasks []Task) (int, error)

	// Update task status to Pending or Completed
	UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) error

	// Delete the task with the specified ID
	Delete(id uuid.UUID) (Task, error)
}
