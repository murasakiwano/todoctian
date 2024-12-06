package task

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Represents a to-do item. A task is either complete or incomplete, so we
// represent this with "To-do" and "Completed".
//
// A task may or may not have a parent task and, in turn, a list of subtasks.
type Task struct {
	// Time of task creation
	CreatedAt time.Time
	// Las time the task was updated
	UpdatedAt time.Time
	// The ID of the parent task
	ParentTaskID *uuid.UUID
	// The name of the task
	Name string
	// The IDs of each of the subtasks
	SubtaskIDs []uuid.UUID
	// Task status (whether it's completed or not)
	Status TaskStatus
	// The ID of the Project this task belongs to
	ProjectID uuid.UUID
	// A unique identifier for the task
	ID uuid.UUID
	// The position of the task relative to its siblings. It starts from 1 (first), 0 means "unset".
	Order int
}

func (t Task) String() string {
	return fmt.Sprintf(
		"{ID: %s Name: %s SubtaskIDs: %v Status: %s ProjectID: %s ParentTaskID: %s Order: %d}",
		t.ID,
		t.Name,
		t.SubtaskIDs,
		t.Status,
		t.ProjectID,
		t.ParentTaskID,
		t.Order,
	)
}

// NewTask returns a new instance of a task. It does not explicitly add the
// task to the project it belongs to! Also, the order is unset at first. It should be explicitly
// set by a service.
func NewTask(name string, projectID uuid.UUID, parentTaskID *uuid.UUID) Task {
	now := time.Now()
	task := Task{
		ID:           uuid.New(),
		Name:         name,
		ProjectID:    projectID,
		Status:       Todo,
		ParentTaskID: parentTaskID,
		SubtaskIDs:   []uuid.UUID{},
		Order:        0, // 0 means the order is unset
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return task
}

func (t Task) IsInProjectRoot() bool {
	return t.ParentTaskID == nil
}

type TaskStatus int

func (t TaskStatus) String() string {
	return [...]string{"To-do", "Completed"}[t]
}

const (
	Todo TaskStatus = iota
	Completed
)
