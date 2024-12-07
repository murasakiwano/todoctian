package task

import (
	"fmt"

	"github.com/google/uuid"
)

func (ts *TaskService) SearchTaskName(partial string, projectID *uuid.UUID) ([]Task, error) {
	tasks, err := ts.taskDB.SearchFuzzy(partial)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search with string %s: %w", partial, err)
	}

	if projectID != nil {
		relatedTasks := []Task{}
		for _, t := range tasks {
			if t.ProjectID == *projectID {
				relatedTasks = append(relatedTasks, t)
			}
		}
		tasks = relatedTasks
	}

	return tasks, nil
}

// Returns the tasks with given status in a specific project.
func (ts *TaskService) SearchTaskByStatus(status TaskStatus, projectID uuid.UUID) []Task {
	return ts.taskDB.GetTasksByStatus(projectID, status)
}
