package task

import (
	"fmt"

	"github.com/google/uuid"
)

// RenameTask changes the name of a previously existing task.
func (ts *TaskService) RenameTask(id uuid.UUID, newTaskName string) (Task, error) {
	// The task must exist first
	task, err := ts.repository.Get(id)
	if err != nil {
		return Task{}, fmt.Errorf("Could not rename task %s: %w", id, err)
	}

	task, err = ts.repository.Rename(task.ID, newTaskName)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}
