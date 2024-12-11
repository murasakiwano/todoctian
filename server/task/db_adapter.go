package task

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/murasakiwano/todoctian/server/db"
	"github.com/murasakiwano/todoctian/server/internal"
)

// Transforms a Task as seen by the db package to a task as seen by the task package
func TaskDBToTaskModel(taskDB db.Task) (Task, error) {
	taskID, err := internal.EncodeUUID(taskDB.ID.Bytes)
	if err != nil {
		return Task{}, err
	}

	createdAt := taskDB.CreatedAt.Time
	projectID, err := internal.EncodeUUID(taskDB.ProjectID.Bytes)
	if err != nil {
		return Task{}, err
	}

	var parentTaskID *uuid.UUID = nil
	if taskDB.ParentTaskID.Valid {
		pTaskID, err := internal.EncodeUUID(taskDB.ParentTaskID.Bytes)
		if err != nil {
			return Task{}, err
		}

		parentTaskID = &pTaskID
	}

	// NOTE: the database guarantees that "status" is either "pending" or "completed"
	taskStatus := TaskStatusPending
	if taskDB.Status == TaskStatusCompleted.String() {
		taskStatus = TaskStatusCompleted
	}

	return Task{
		ID:           taskID,
		CreatedAt:    createdAt,
		ParentTaskID: parentTaskID,
		ProjectID:    projectID,
		Status:       taskStatus,
		Order:        int(taskDB.Order),
		Name:         taskDB.Name,
	}, nil
}

// Transforms a Task as seen by the task package to a task as seen by the db package
func TaskModelToTaskDB(task Task) (db.Task, error) {
	// Step 1: convert the task, project, and (possibly) parent task IDs to pgtype.UUID
	pgTaskUUID, err := internal.ScanUUID(task.ID)
	if err != nil {
		return db.Task{}, err
	}

	pgProjectUUID, err := internal.ScanUUID(task.ProjectID)
	if err != nil {
		return db.Task{}, err
	}

	pgParentTaskUUID := pgtype.UUID{}
	if task.ParentTaskID != nil {
		pgParentTaskUUID, err = internal.ScanUUID(*task.ParentTaskID)
		if err != nil {
			return db.Task{}, err
		}
	}

	// Step 2: convert task.CreatedAt to pgtype.Timestamp
	pgCreatedAt := pgtype.Timestamp{}
	err = pgCreatedAt.Scan(task.CreatedAt)
	if err != nil {
		return db.Task{}, err
	}

	// Step 3: convert task status to text
	pgTaskStatus := task.Status.String()

	// Step 3: return task data as defined by the db
	return db.Task{
		ID:           pgTaskUUID,
		CreatedAt:    pgCreatedAt,
		ParentTaskID: pgParentTaskUUID,
		ProjectID:    pgProjectUUID,
		Status:       pgTaskStatus,
		Order:        int32(task.Order),
		Name:         task.Name,
	}, nil
}
