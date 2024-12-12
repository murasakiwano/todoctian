package task

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (ts *TaskService) SearchTaskByProject(projectID uuid.UUID) ([]Task, error) {
	return ts.repository.GetTasksByProject(projectID)
}

func (ts *TaskService) SearchTaskName(partial string, projectID uuid.UUID) ([]Task, error) {
	tasks, err := ts.repository.GetTasksByProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf(
			"failed to perform search with string %s: there are no tasks for project %s",
			partial,
			projectID,
		)
	}

	type taskAndDistance struct {
		task     Task
		distance int
	}
	matchingTasks := []taskAndDistance{}
	for _, t := range tasks {
		distance := fuzzy.RankMatchFold(partial, t.Name)
		ts.logger.Debug("fuzzy.RankMatchFold result", slog.Int("result", distance), slog.String("taskName", t.Name))

		if distance > -1 {
			matchingTasks = append(matchingTasks, taskAndDistance{task: t, distance: distance})
		}
	}

	slices.SortFunc(matchingTasks, func(a, b taskAndDistance) int {
		if a.distance > b.distance {
			return -1
		}
		if a.distance < b.distance {
			return 1
		}

		return 0
	})

	sortedTasks := []Task{}
	for _, t := range matchingTasks {
		sortedTasks = append(sortedTasks, t.task)
	}

	return sortedTasks, nil
}

// Returns the tasks with given status in a specific project.
func (ts *TaskService) SearchTaskByStatus(status TaskStatus, projectID uuid.UUID) ([]Task, error) {
	return ts.repository.GetTasksByStatus(projectID, status)
}
