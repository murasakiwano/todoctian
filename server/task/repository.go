package task

import (
	"github.com/google/uuid"
)

type TaskRepository interface {
	Save(task Task) error
	BatchSave(tasks []Task) (int, error)
	Get(id uuid.UUID) (Task, error)
	Delete(id uuid.UUID) (Task, error)
	Update(task Task) error
	BatchUpdate(tasks []Task) (int, error)
	UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) error
	GetSubtasks(id uuid.UUID) ([]Task, error)
	GetTasksByProject(projectID uuid.UUID) []Task
	GetTasksInProjectRoot(projectID uuid.UUID) []Task
	SearchFuzzy(partialTaskName string) ([]Task, error)
}
