package task

import (
	"github.com/google/uuid"
)

type TaskRepository interface {
	// Create a task in the database
	Create(task Task) error

	// Retrieve a task by its ID
	Get(id uuid.UUID) (Task, error)

	// Retrieve all direct children/subtasks of a specific task
	GetSubtasksDirect(id uuid.UUID) ([]Task, error)

	// Recursively retrieve all subtasks of a specific task
	GetSubtasksDeep(id uuid.UUID) ([]Task, error)

	// Retrieve all tasks in a specific project
	GetTasksByProject(projectID uuid.UUID) ([]Task, error)

	// Retrieve all tasks in a project
	GetTasksInProjectRoot(projectID uuid.UUID) ([]Task, error)

	// Filter tasks in a project by their status
	GetTasksByStatus(projectID uuid.UUID, status TaskStatus) ([]Task, error)

	// Rename a single task
	Rename(taskID uuid.UUID, newName string) error

	// Update a single task's order
	UpdateOrder(taskID uuid.UUID, newTaskOrder int) error

	// Batch update the order a collection of tasks
	BatchUpdateOrder(tasks []Task) error

	// Update task status to Pending or Completed
	UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) error

	// Delete the task with the specified ID
	Delete(id uuid.UUID) (Task, error)
}
