package task

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

func (ts *TaskService) UpdateTaskStatus(id uuid.UUID, status string) error {
	task, err := ts.repository.Get(id)
	if err != nil {
		return err
	}

	switch status {
	case TaskStatusPending.value:
		return ts.markTaskAsPending(task)

	case TaskStatusCompleted.value:
		return ts.markTaskAsCompleted(task)
	}

	return fmt.Errorf("invalid task status: %s", status)
}

// markTaskAsPending either does nothing (if the task is already marked "Todo") or updates the
// task status to "Todo". When a subtask is marked as pending, its parent must also be marked as
// pending, since it does not make sense to have a collection of tasks be marked as completed
// when not all steps have been done.
func (ts *TaskService) markTaskAsPending(task Task) error {
	if task.Status == TaskStatusPending {
		return nil
	}

	task.Status = TaskStatusPending
	err := ts.repository.UpdateTaskStatus(task.ID, TaskStatusPending)
	if err != nil {
		return err
	}

	if task.ParentTaskID == nil {
		return nil
	}

	parentTask, err := ts.repository.Get(*task.ParentTaskID)
	if err != nil {
		return err
	}

	if parentTask.Status == TaskStatusPending {
		return nil
	}

	return ts.markTaskAsPending(parentTask)
}

// completeTask sets a task as completed. When a task is completed, all its subtasks must also be
// completed. Here we do a tree traversal downwards and then upwards. We stop the traversal whenever
// we find an already completed task, since it means that the work has already been done for it.
func (ts *TaskService) markTaskAsCompleted(task Task) error {
	err := ts.completeTask(task)
	if err != nil {
		return err
	}

	// Tree traversal downwards
	err = ts.completeSubtasks(task)
	if err != nil {
		return err
	}

	// Tree traversal upwards
	err = ts.completeParentTask(task)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskService) completeTask(task Task) error {
	ts.logger.Debug("marked task as completed", slog.String("taskID", task.ID.String()))
	err := ts.repository.UpdateTaskStatus(task.ID, TaskStatusCompleted)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskService) completeSubtasks(task Task) error {
	subtasks, err := ts.repository.GetSubtasksDeep(task.ID)
	if err != nil {
		return err
	}
	if len(subtasks) == 0 {
		return nil
	}

	ts.logger.Debug("task has subtasks, completing them...", slog.String("taskID", task.ID.String()))
	for _, subtask := range subtasks {
		err := ts.completeTask(subtask)
		if err != nil {
			ts.logger.Error(
				"Failed to complete task",
				slog.Group("task",
					slog.String("id", task.ID.String()),
					slog.String("parentTaskID", task.ParentTaskID.String()),
					slog.String("projectID", task.ProjectID.String()),
				),
			)
			return err
		}
	}

	return nil
}

func (ts *TaskService) completeParentTask(task Task) error {
	if task.ParentTaskID == nil {
		return nil
	}

	ts.logger.Debug(
		"task has parent task, checking if it needs completion...",
		slog.String("taskID", task.ID.String()),
		slog.String("parentTaskID", task.ParentTaskID.String()),
	)
	parentTask, err := ts.repository.Get(*task.ParentTaskID)
	if err != nil {
		return err
	}

	// Suppose that this current task is the only child that's left to be done. Then, we need to
	// mark the parent as completed.
	siblings, err := ts.FetchTaskSiblings(task)
	if err != nil {
		return err
	}

	ts.logger.Debug("found task siblings", slog.Any("siblings", siblings))
	if ts.allSiblingsCompleted(siblings) && parentTask.Status != TaskStatusCompleted {
		err := ts.completeTask(parentTask)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ts *TaskService) allSiblingsCompleted(tasks []Task) bool {
	for _, s := range tasks {
		if s.Status != TaskStatusCompleted {
			ts.logger.Debug(
				"sibling is not marked completed, will not complete parent task",
				slog.String("siblingID", s.ID.String()),
			)
			return false
		}
	}

	return true
}
