package task

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/murasakiwano/todoctian/server/internal"
)

type TaskRepositoryInMemory struct {
	tasks map[uuid.UUID]Task
}

func NewTaskRepositoryInMemory() *TaskRepositoryInMemory {
	return &TaskRepositoryInMemory{tasks: make(map[uuid.UUID]Task)}
}

func (tr *TaskRepositoryInMemory) Save(task Task) error {
	valid, err := tr.TaskIsValidParent(task.ParentTaskID)
	if err != nil || !valid {
		return errors.New("Task has an invalid parent task")
	}

	tr.tasks[task.ID] = task

	if task.ParentTaskID != nil {
		parentTask, exists := tr.tasks[*task.ParentTaskID]
		if !exists {
			return errors.New("Parent task does not exist")
		}

		parentTask.SubtaskIDs = append(parentTask.SubtaskIDs, task.ID)
		tr.tasks[parentTask.ID] = parentTask
	}

	return nil
}

func (tr *TaskRepositoryInMemory) BatchSave(tasks []Task) (int, error) {
	n := len(tasks)

	for _, t := range tasks {
		err := tr.Save(t)
		if err != nil {
			return 0, err
		}
	}

	return n, nil
}

// Update reads a task from the database, then updates the information. For simplicity, it overwrites
// all the task info.
func (tr *TaskRepositoryInMemory) Update(task Task) error {
	_, exists := tr.tasks[task.ID]
	if !exists {
		return internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", task.ID))
	}

	tr.tasks[task.ID] = task

	return nil
}

func (tr *TaskRepositoryInMemory) BatchUpdate(tasks []Task) (int, error) {
	n := len(tasks)

	for _, task := range tasks {
		err := tr.Update(task)
		if err != nil {
			return 0, err
		}
	}

	return n, nil
}

func (tr *TaskRepositoryInMemory) Get(id uuid.UUID) (Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return Task{}, internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	return task, nil
}

func (tr *TaskRepositoryInMemory) Delete(id uuid.UUID) (Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return Task{}, internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	delete(tr.tasks, id)

	return task, nil
}

func (tr *TaskRepositoryInMemory) UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) error {
	task, exists := tr.tasks[id]
	if !exists {
		return internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	task.Status = newStatus
	tr.tasks[id] = task

	return nil
}

func (tr TaskRepositoryInMemory) TaskIsValidParent(parentTaskID *uuid.UUID) (bool, error) {
	// If there is no parent task (i.e., task.ParentTask == nil), then it is valid
	if parentTaskID == nil {
		return true, nil
	}

	// If there is a parent task, then we need to check if it exists in the database
	_, err := tr.Get(*parentTaskID)
	// If it does not exist, it is invalid
	if err != nil {
		// We need to check if the parent task was not found or if another error occurred
		if errors.Is(err, internal.ErrNotFound) {
			return false, nil
		}

		return false, err
	}

	// If it exists, then the parent task is valid
	return true, nil
}

func (tr *TaskRepositoryInMemory) GetSubtasks(taskID uuid.UUID) ([]Task, error) {
	subtasks := []Task{}
	task, err := tr.Get(taskID)
	if err != nil {
		return subtasks, err
	}

	for _, subtaskID := range task.SubtaskIDs {
		subtask, err := tr.Get(subtaskID)
		if err != nil {
			return subtasks, err
		}
		subtasks = append(subtasks, subtask)
	}

	return subtasks, nil
}

func (tr *TaskRepositoryInMemory) GetTasksByProject(projectID uuid.UUID) []Task {
	projectTasks := []Task{}

	for _, task := range tr.tasks {
		if task.ProjectID == projectID {
			projectTasks = append(projectTasks, task)
		}
	}

	return projectTasks
}

func (tr *TaskRepositoryInMemory) GetTasksInProjectRoot(projectID uuid.UUID) []Task {
	projectTasks := tr.GetTasksByProject(projectID)

	root := []Task{}
	for _, task := range projectTasks {
		if task.IsInProjectRoot() {
			root = append(root, task)
		}
	}

	return root
}

func (tr *TaskRepositoryInMemory) SearchFuzzy(partialTaskName string) ([]Task, error) {
	tasks := []Task{}
	for _, t := range tr.tasks {
		if fuzzy.MatchFold(partialTaskName, t.Name) {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}
