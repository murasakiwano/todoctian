package task

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/project"
)

type TaskService struct {
	repository TaskRepository
	projectDB  project.ProjectRepository
	logger     slog.Logger
}

func NewTaskService(taskRepository TaskRepository, projectRepository project.ProjectRepository) *TaskService {
	return &TaskService{
		repository: taskRepository,
		projectDB:  projectRepository,
		logger:     *internal.NewLogger("TaskService"),
	}
}

// ValidateTask checks if some conditions are true for a given task:
// - The project it references must exist
// - If there is a parent task, it must exist
func (ts TaskService) ValidateTask(task Task) error {
	// Check if the project exists
	_, err := ts.projectDB.Get(task.ProjectID)
	if err != nil {
		return fmt.Errorf("Failed to fetch project %s from repository: %w", task.ProjectID, err)
	}
	// Check if the task parent is valid
	if task.ParentTaskID != nil {
		parentTask, err := ts.repository.Get(*task.ParentTaskID)
		if err != nil {
			return err
		}

		if parentTask.ProjectID != task.ProjectID {
			return errors.New("task and parent task must belong to the same project")
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
		taskSiblings, err := ts.repository.GetTasksInProjectRoot(task.ProjectID)
		if err != nil {
			return nil, err
		}
		siblings = taskSiblings
	} else {
		s, err := ts.repository.GetSubtasksDirect(*task.ParentTaskID)
		if err != nil {
			return siblings, err
		}

		siblings = s
	}

	return siblings, nil
}

func (ts TaskService) FindTaskByID(taskID uuid.UUID) (Task, error) {
	return ts.repository.Get(taskID)
}

func (ts TaskService) ListTasks() ([]Task, error) {
	return ts.repository.List()
}

func (ts TaskService) FetchSubtasksDirect(taskID uuid.UUID) ([]Task, error) {
	return ts.repository.GetSubtasksDirect(taskID)
}

func (ts TaskService) FetchSubtasksDeep(taskID uuid.UUID) ([]Task, error) {
	return ts.repository.GetSubtasksDeep(taskID)
}
