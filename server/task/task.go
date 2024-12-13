package task

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Task struct    Represents a to-do item. A task is either complete or incomplete, so we
// represent this with "pending" and "completed".
//
// A task may or may not have a parent task and, in turn, a list of subtasks.
type Task struct {
	// Time of task creation
	CreatedAt time.Time
	// The ID of the parent task
	ParentTaskID *uuid.UUID
	// The name of the task
	Name string
	// Task status (whether it's completed or not)
	Status TaskStatus
	// Subtasks of the this task. It is not necessarily present
	Subtasks []Task
	// The ID of the Project this task belongs to
	ProjectID uuid.UUID
	// A unique identifier for the task
	ID uuid.UUID
	// The position of the task relative to its siblings. It starts from 1 (first), 0 means "unset".
	Order int
}

func (t Task) String() string {
	return fmt.Sprintf(
		"{ID: %s Name: %s Status: %s ProjectID: %s ParentTaskID: %s Order: %d}",
		t.ID,
		t.Name,
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
		Status:       TaskStatusPending,
		ParentTaskID: parentTaskID,
		Order:        0, // 0 means the order is unset
		CreatedAt:    now,
		Subtasks:     []Task{},
	}

	return task
}

func (t Task) IsInProjectRoot() bool {
	return t.ParentTaskID == nil
}

type TaskStatus struct {
	value string
}

func (t TaskStatus) String() string {
	return t.value
}

func (t *TaskStatus) FromString(value string) error {
	switch value {
	case TaskStatusCompleted.value:
		t.value = value
		return nil

	case TaskStatusPending.value:
		t.value = value
		return nil
	}

	return fmt.Errorf("unknown enum value: %v", value)
}

var (
	TaskStatusCompleted = TaskStatus{value: "completed"}
	TaskStatusPending   = TaskStatus{value: "pending"}
)

func (t Task) LogValue() slog.Value {
	subtaskIDs := []uuid.UUID{}

	return slog.GroupValue(
		slog.String("ID", t.ID.String()),
		slog.String("Name", t.Name),
		slog.String("Status", t.Status.String()),
		slog.String("ProjectID", t.ProjectID.String()),
		slog.Time("CreatedAt", t.CreatedAt),
		slog.Int("Order", t.Order),
		slog.Any("SubtaskIDs", subtaskIDs),
		slog.Any("ParentTaskID", t.ParentTaskID),
	)
}
