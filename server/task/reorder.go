package task

import (
	"fmt"
	"log/slog"
	"slices"
)

// ReorderTask changes the order of the task accordingly with the given number.
//
// If the order is less than 0, it will bring the task to the beginning. If it is greater than
// the number of tasks in its level, than the task will be brought to the end. If the new order
// is equal to the current order, then nothing is done.
//
// The algorithm is as follows:
// - Find the interval between the current order and the new order
// - Slice the siblings array in this interval
// - Update the order of each task accordingly:
//   - If the current order is less than the new order, then we need to subtract 1 from all other siblings
//   - If the current order is greater than the new order, then we need to add 1 to all other siblings
func (ts *TaskService) ReorderTask(task Task, newOrder int) error {
	// Check if the task exists
	_, err := ts.taskDB.Get(task.ID)
	if err != nil {
		return err
	}

	siblings, err := ts.FetchTaskSiblings(task)
	if err != nil {
		return fmt.Errorf("Failed to fetch task siblings for %s: %w", task.ID, err)
	}
	if newOrder < 0 {
		newOrder = 0
	} else if newOrder > len(siblings)-1 {
		newOrder = len(siblings) - 1
	}

	if newOrder == task.Order {
		ts.logger.Debug("newOrder and task.Order are equal")
		return nil
	}

	slices.SortFunc(siblings, cmpTasks) // make the order match the index

	if newOrder > task.Order {
		reorderSlice := siblings[task.Order : newOrder+1]
		ts.logger.Debug(
			"newOrder is greater than task.Order",
			slog.Int("newOrder", newOrder),
			slog.Int("task.Order", task.Order),
			slog.Any("siblings", siblings),
			slog.Any("reorderSlice", reorderSlice),
		)

		for _, t := range reorderSlice[1:] {
			ts.logger.Debug("reordering tasks...", slog.Any("task", t))
			siblings[t.Order].Order -= 1
		}
		siblings[task.Order].Order = newOrder // Need to alter it directly in the slice
	} else {
		reorderSlice := siblings[newOrder : task.Order+1]
		ts.logger.Debug(
			"newOrder is less than task.Order",
			slog.Int("newOrder", newOrder),
			slog.Int("task.Order", task.Order),
			slog.Any("siblings", siblings),
			slog.Any("reorderSlice", reorderSlice),
		)

		for _, t := range reorderSlice[:len(reorderSlice)-1] {
			ts.logger.Debug("reordering tasks...", slog.Any("task", t))
			siblings[t.Order].Order += 1
		}
		siblings[task.Order].Order = newOrder
	}

	slices.SortFunc(siblings, cmpTasks)
	ts.logger.Debug("Siblings are now like this", slog.Any("siblings", siblings))

	ts.taskDB.BatchUpdate(siblings)
	return nil
}

func cmpTasks(taskA, taskB Task) int {
	if taskA.Order < taskB.Order {
		return -1
	}
	if taskA.Order > taskB.Order {
		return 1
	}

	return 0 // taskA.Order == taskB.Order
}
