package task

import (
	"errors"
	"fmt"
	"slices"
	"time"

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

// Save persists a task to the tasks map.
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

		parentTask.Subtasks = append(parentTask.Subtasks, task)
		tr.tasks[parentTask.ID] = parentTask
	}

	return nil
}

// Update reads a task from the map, then updates its information. For simplicity, it overwrites
// all the task info.
func (tr *TaskRepositoryInMemory) Update(task Task) error {
	_, exists := tr.tasks[task.ID]
	if !exists {
		return internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", task.ID))
	}

	task.UpdatedAt = time.Now()
	tr.tasks[task.ID] = task

	return nil
}

// BatchUpdate updates a collection of tasks in batch.
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

// Read a specific task by its ID in the map.
func (tr *TaskRepositoryInMemory) Get(id uuid.UUID) (Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return Task{}, internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	return task, nil
}

// Delete a task with the specified ID.
func (tr *TaskRepositoryInMemory) Delete(id uuid.UUID) (Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return Task{}, internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	delete(tr.tasks, id)

	return task, nil
}

// Update a task status to Pending or Completed.
func (tr *TaskRepositoryInMemory) UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) error {
	task, exists := tr.tasks[id]
	if !exists {
		return internal.NewNotFoundError(fmt.Sprintf("Task with ID %s", id))
	}

	task.Status = newStatus
	err := tr.Update(task)
	if err != nil {
		return fmt.Errorf("Failed to update task status: %w", err)
	}

	return nil
}

// Check if the ID belongs to a valid task.
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

// Get all direct children of a task.
func (tr *TaskRepositoryInMemory) GetSubtasksDirect(taskID uuid.UUID) ([]Task, error) {
	subtasks := []Task{}
	task, err := tr.Get(taskID)
	if err != nil {
		return subtasks, err
	}

	for _, subtask := range task.Subtasks {
		subtask, err := tr.Get(subtask.ID)
		if err != nil {
			return subtasks, err
		}
		subtasks = append(subtasks, subtask)
	}

	return subtasks, nil
}

// Recursively retrieve all subtasks of a specific task
func (r *TaskRepositoryInMemory) GetSubtasksDeep(taskID uuid.UUID) ([]Task, error) {
	subtasks := []Task{}
	task, err := r.Get(taskID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get subtasks of task %s: %w", taskID, err)
	}

	for _, subtask := range task.Subtasks {
		ssubtasks, err := r.GetSubtasksDeep(subtask.ID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get subtasks of task %s: %w", subtask.ID, err)
		}

		subtasks = slices.Concat(subtasks, ssubtasks)
	}

	return subtasks, nil
}

// Filter tasks by project.
func (tr *TaskRepositoryInMemory) GetTasksByProject(projectID uuid.UUID) []Task {
	projectTasks := []Task{}

	for _, task := range tr.tasks {
		if task.ProjectID == projectID {
			projectTasks = append(projectTasks, task)
		}
	}

	return projectTasks
}

// Filter tasks in a project by status.
func (tr *TaskRepositoryInMemory) GetTasksByStatus(projectID uuid.UUID, status TaskStatus) []Task {
	projectTasks := tr.GetTasksByProject(projectID)

	tasks := []Task{}
	for _, t := range projectTasks {
		if t.Status == status {
			tasks = append(tasks, t)
		}
	}

	return tasks
}

// Get all tasks that are in the project root, i.e., that have no parent task.
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

// Fuzzy search a task by its name.
func (tr *TaskRepositoryInMemory) SearchFuzzy(partialTaskName string) ([]Task, error) {
	tasks := []Task{}
	for _, t := range tr.tasks {
		if fuzzy.MatchFold(partialTaskName, t.Name) {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}