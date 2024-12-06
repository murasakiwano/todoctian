package task

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/project"
)

type TaskService struct {
	taskDB    TaskRepository
	projectDB project.ProjectRepository
	logger    slog.Logger
}

func NewTaskService(taskRepository TaskRepository, projectRepository project.ProjectRepository) *TaskService {
	return &TaskService{
		taskDB:    taskRepository,
		projectDB: projectRepository,
		logger: *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

// ValidateTask checks if some conditions are true for a given task:
// - The project it references must exist
// - If there is a parent task, it must exist
func (ts TaskService) ValidateTask(task Task) error {
	// Check if the project exists
	_, err := ts.projectDB.Get(task.ProjectID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return fmt.Errorf("Task %s is invalid: the project %s does not exist", task.ID, task.ProjectID)
		}

		return fmt.Errorf("Failed to fetch project %s from repository: %w", task.ProjectID, err)
	}
	// Check if the task parent is valid
	if task.ParentTaskID != nil {
		_, err = ts.taskDB.Get(*task.ParentTaskID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Fetches every task in a given level of the task tree.
//
// The siblings of a task are the ones that are found in the same level of the task tree.
// This means that it is a set of tasks that has the same parent task. If there is no parent task,
// then they are the tasks in the same project with a nil parent task ID.
func (ts TaskService) FetchTaskSiblings(task Task) ([]Task, error) {
	siblings := []Task{}
	if task.IsInProjectRoot() {
		siblings = ts.taskDB.GetTasksInProjectRoot(task.ProjectID)
	} else {
		s, err := ts.taskDB.GetSubtasks(*task.ParentTaskID)
		if err != nil {
			return siblings, err
		}

		siblings = s
	}

	return siblings, nil
}

func (ts TaskService) FindTaskByID(taskID uuid.UUID) (Task, error) {
	return ts.taskDB.Get(taskID)
}
