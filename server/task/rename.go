package task

import (
	"fmt"

	"github.com/google/uuid"
)

// RenameTask changes the name of a previously existing task.
func (ts *TaskService) RenameTask(id uuid.UUID, newTaskName string) error {
	// The task must exist first
	task, err := ts.taskDB.Get(id)
	if err != nil {
		return fmt.Errorf("Could not rename task %s: %w", id, err)
	}

	task.Name = newTaskName
	err = ts.taskDB.Save(task)
	if err != nil {
		return err
	}

	return nil
}
