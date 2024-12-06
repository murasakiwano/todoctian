package task

import (
	"fmt"

	"github.com/google/uuid"
)

// CreateTask instantiates a new Task and persists it to the TaskRepository, while performing
// validations.
func (t *TaskService) CreateTask(taskName string, projectID uuid.UUID, parentTaskID *uuid.UUID) (Task, error) {
	task := NewTask(taskName, projectID, parentTaskID)
	err := t.ValidateTask(task)
	if err != nil {
		return Task{}, fmt.Errorf("Could not create task \"%s\": %w", taskName, err)
	}

	task, err = t.setInitialTaskOrder(task)
	if err != nil {
		return Task{}, err
	}

	return task, t.taskDB.Save(task)
}

// Sets the initial order of the task relative to its siblings. The order is an integer starting at 0
// (first task to be performed). The initial order is equivalent to the number of siblings.
//
// NOTE: this is before saving the task to the database!
func (ts *TaskService) setInitialTaskOrder(task Task) (Task, error) {
	siblings, err := ts.FetchTaskSiblings(task)
	if err != nil {
		return Task{}, err
	}

	task.Order = len(siblings)

	return task, nil
}
