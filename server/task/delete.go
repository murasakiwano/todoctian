package task

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

func (ts *TaskService) DeleteTask(id uuid.UUID) (Task, error) {
	task, err := ts.taskDB.Get(id)
	if err != nil {
		return Task{}, fmt.Errorf("Could not get task from DB: %w", err)
	}

	err = ts.maybeDeleteSubtasks(task)
	if err != nil {
		return Task{}, fmt.Errorf("Failed to delete subtasks of task %s: %w", task.ID, err)
	}

	err = ts.rearrangeTaskSiblings(task)
	if err != nil {
		return Task{}, fmt.Errorf("Failed to rearrange siblings of task %s: %w", task.ID, err)
	}

	return ts.taskDB.Delete(id)
}

func (ts *TaskService) maybeDeleteSubtasks(task Task) error {
	// Does the task have any subtasks?
	if len(task.SubtaskIDs) == 0 {
		return nil
	}
	// Delete the subtasks recursively, bottom-up
	for _, subtaskID := range task.SubtaskIDs {
		_, err := ts.DeleteTask(subtaskID)
		if err != nil {
			return fmt.Errorf("Failed to delete subtask %s of task %s: %w", subtaskID, task.ID, err)
		}
	}

	return nil
}

func (ts *TaskService) rearrangeTaskSiblings(task Task) error {
	siblings, err := ts.FetchTaskSiblings(task)
	if err != nil {
		return fmt.Errorf("Failed to fetch siblings for task %s: %w", task.ID, err)
	}

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
	_, err = ts.taskDB.BatchUpdate(siblings[:len(siblings)-1])
	if err != nil {
		return fmt.Errorf("Failed to update siblings of task %s: %w", task.ID, err)
	}

	return nil
}
