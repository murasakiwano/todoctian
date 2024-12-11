package task

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
)

// DeleteTask method    Deletes a task if it exists. If it does not, it is a no-op.
func (ts *TaskService) DeleteTask(id uuid.UUID) (Task, error) {
	task, err := ts.repository.Get(id)
	if err != nil {
		if !errors.Is(err, internal.ErrNotFound) {
			return Task{}, fmt.Errorf("Could not get task from DB: %w", err)
		}

		ts.logger.Warn("task does not exist, nothing to do", slog.String("taskID", id.String()))
		return Task{}, nil
	}

	err = ts.maybeDeleteSubtasks(task)
	if err != nil {
		return Task{}, fmt.Errorf("Failed to delete subtasks of task %s: %w", task.ID, err)
	}

	err = ts.rearrangeTaskSiblings(task)
	if err != nil {
		return Task{}, fmt.Errorf("Failed to rearrange siblings of task %s: %w", task.ID, err)
	}

	return ts.repository.Delete(id)
}

func (ts *TaskService) maybeDeleteSubtasks(task Task) error {
	// Does the task have any subtasks?
	subtasks, err := ts.repository.GetSubtasksDeep(task.ID)
	if err != nil {
		return err
	}
	if len(subtasks) == 0 {
		return nil
	}

	// Delete the subtasks recursively, bottom-up
	for _, subtask := range subtasks {
		_, err := ts.DeleteTask(subtask.ID)
		if err != nil {
			return fmt.Errorf("Failed to delete subtask %s of task %s: %w", subtask, task.ID, err)
		}
	}

	return nil
}

func (ts *TaskService) rearrangeTaskSiblings(task Task) error {
	siblings, err := ts.FetchTaskSiblings(task)
	if err != nil {
		return fmt.Errorf("Failed to fetch siblings for task %s: %w", task.ID, err)
	}

	slog.Debug(
		"found task siblings",
		slog.Any("task", task),
		slog.Any("siblings", siblings),
	)

	if len(siblings) <= 1 {
		return nil
	}
	slices.SortFunc(siblings, cmpTasks)

	siblingsAfterTaskToDelete := siblings[task.Order+1:]
	for i, s := range siblingsAfterTaskToDelete {
		s.Order -= 1
		siblings[task.Order+1-i] = s
	}

	_ = slices.Delete(siblings, task.Order, task.Order+1)
	err = ts.repository.BatchUpdateOrder(siblings[:len(siblings)-1])
	if err != nil {
		return fmt.Errorf("Failed to update siblings of task %s: %w", task.ID, err)
	}

	slog.Debug("task siblings after rearrangement", slog.Any("siblings", siblings))

	return nil
}
