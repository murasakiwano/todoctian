package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/db"
	"github.com/murasakiwano/todoctian/server/internal"
)

// This code indicates that a duplicate constraint was violated by the query
var ErrPgDuplicate = "23505"

type TaskRepositoryPostgres struct {
	Queries *db.Queries
	ctx     context.Context
	logger  slog.Logger
}

func NewTaskRepositoryPostgres(ctx context.Context, pool *pgxpool.Pool) (*TaskRepositoryPostgres, error) {
	slog.Debug("Connected to the database")
	return &TaskRepositoryPostgres{
		Queries: db.New(pool),
		ctx:     ctx,
		logger:  *internal.NewLogger("TaskRepositoryPostgres"),
	}, nil
}

func (t *TaskRepositoryPostgres) Create(task Task) error {
	taskDB, err := TaskModelToTaskDB(task)
	if err != nil {
		t.logger.Error("failed to adapt task model to task db",
			slog.Any("task", task),
			slog.String("err", err.Error()),
		)
		return err
	}

	err = t.Queries.CreateTask(t.ctx, db.CreateTaskParams{
		ID:           taskDB.ID,
		ProjectID:    taskDB.ProjectID,
		Name:         taskDB.Name,
		Status:       taskDB.Status,
		Order:        taskDB.Order,
		ParentTaskID: taskDB.ParentTaskID,
		CreatedAt:    taskDB.CreatedAt,
	})
	if err != nil {
		t.logger.Info("failed to create task", slog.Any("task", task), slog.String("err", err.Error()))

		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ErrPgDuplicate {
			err = internal.NewAlreadyExistsError(fmt.Sprintf("Task \"%s\"", task.ID))
		}

		return err
	}

	t.logger.Debug("Successfully created task", slog.Any("task", task))

	return nil
}

func (t *TaskRepositoryPostgres) Get(id uuid.UUID) (Task, error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return Task{}, err
	}

	taskDB, err := t.Queries.GetTask(t.ctx, pgUUID)
	if err != nil {
		t.logger.Error("failed to retrieve task from database",
			slog.String("taskID", id.String()),
			slog.String("err", err.Error()),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			err = internal.NewNotFoundError(fmt.Sprintf("Task %s", id))
		}
		return Task{}, err
	}

	task, err := TaskDBToTaskModel(taskDB)
	if err != nil {
		return Task{}, err
	}

	t.logger.Debug("retrieved task from database", slog.Any("Task", task))

	return task, nil
}

// Retrieve all direct children/subtasks of a specific task
func (t *TaskRepositoryPostgres) GetSubtasksDirect(id uuid.UUID) (_ []Task, _ error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return nil, err
	}

	subtasksDB, err := t.Queries.GetSubtasksDirect(t.ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	subtasks := []Task{}
	subtaskIDs := []string{}
	for _, subtaskDB := range subtasksDB {
		s, err := TaskDBToTaskModel(subtaskDB)
		if err != nil {
			return nil, err
		}
		subtasks = append(subtasks, s)
		subtaskIDs = append(subtaskIDs, s.ID.String())
	}

	t.logger.Info("successfully retrieved subtasks", slog.String("ParentTaskID", id.String()), slog.Any("SubtaskIDs", subtaskIDs))

	return subtasks, nil
}

// Recursively retrieve all subtasks of a specific task
func (t *TaskRepositoryPostgres) GetSubtasksDeep(id uuid.UUID) (_ []Task, _ error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return nil, err
	}

	subtasksDeepDB, err := t.Queries.GetSubtasksDeep(t.ctx, pgUUID)
	if err != nil {
		t.logger.Error("failed to retrieve deep subtasks", slog.Any("ParentTaskID", id.String()), slog.String("err", err.Error()))
		return nil, err
	}

	subtasksDeep := []Task{}
	subtaskIDs := []string{}
	for _, subtaskDB := range subtasksDeepDB {
		subtask, err := TaskDBToTaskModel(db.Task(subtaskDB))
		if err != nil {
			return nil, err
		}

		subtasksDeep = append(subtasksDeep, subtask)
		subtaskIDs = append(subtaskIDs, subtask.ID.String())
	}

	t.logger.Info("successfully retrieved subtasks", slog.Any("subtaskIDs", subtaskIDs))

	return subtasksDeep, nil
}

// Retrieve all tasks in a specific project
func (t *TaskRepositoryPostgres) GetTasksByProject(projectID uuid.UUID) (_ []Task, _ error) {
	pgUUID, err := internal.ScanUUID(projectID)
	if err != nil {
		return nil, err
	}

	projectTasksDB, err := t.Queries.GetTasksByProject(t.ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	projectTasks := []Task{}
	for _, pTaskDB := range projectTasksDB {
		pTask, err := TaskDBToTaskModel(pTaskDB)
		if err != nil {
			return nil, err
		}
		projectTasks = append(projectTasks, pTask)
	}

	return projectTasks, nil
}

// Retrieve all tasks in a project
func (t *TaskRepositoryPostgres) GetTasksInProjectRoot(projectID uuid.UUID) (_ []Task, _ error) {
	pgUUID, err := internal.ScanUUID(projectID)
	if err != nil {
		return nil, err
	}

	projectRootDB, err := t.Queries.GetTasksInProjectRoot(t.ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	projectRoot := []Task{}
	for _, taskDB := range projectRootDB {
		task, err := TaskDBToTaskModel(taskDB)
		if err != nil {
			return nil, err
		}

		projectRoot = append(projectRoot, task)
	}

	return projectRoot, nil
}

// Filter tasks in a project by their status
func (t *TaskRepositoryPostgres) GetTasksByStatus(projectID uuid.UUID, status TaskStatus) (_ []Task, _ error) {
	pgUUID, err := internal.ScanUUID(projectID)
	if err != nil {
		return nil, err
	}

	tasksDB, err := t.Queries.GetTasksByStatus(t.ctx, db.GetTasksByStatusParams{
		ProjectID: pgUUID,
		Status:    status.String(),
	})
	if err != nil {
		return nil, err
	}

	tasks := []Task{}
	for _, taskDB := range tasksDB {
		task, err := TaskDBToTaskModel(taskDB)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Rename a single task
func (t *TaskRepositoryPostgres) Rename(taskID uuid.UUID, newName string) (_ error) {
	pgUUID, err := internal.ScanUUID(taskID)
	if err != nil {
		return err
	}

	return t.Queries.RenameTask(t.ctx, db.RenameTaskParams{ID: pgUUID, Name: newName})
}

// Update a single task's order
func (t *TaskRepositoryPostgres) UpdateOrder(taskID uuid.UUID, newTaskOrder int) (_ error) {
	pgUUID, err := internal.ScanUUID(taskID)
	if err != nil {
		return err
	}

	return t.Queries.UpdateTaskOrder(t.ctx, db.UpdateTaskOrderParams{
		ID:    pgUUID,
		Order: int32(newTaskOrder),
	})
}

// Batch update the order a collection of tasks
func (t *TaskRepositoryPostgres) BatchUpdateOrder(tasks []Task) (_ error) {
	batchUpdateTaskOrderParams := []db.BatchUpdateTaskOrdersParams{}

	for _, task := range tasks {
		pgUUID, err := internal.ScanUUID(task.ID)
		if err != nil {
			return err
		}

		batchUpdateTaskOrderParams = append(batchUpdateTaskOrderParams,
			db.BatchUpdateTaskOrdersParams{
				ID:    pgUUID,
				Order: int32(task.Order),
			},
		)
	}

	errs := []error{}
	br := t.Queries.BatchUpdateTaskOrders(t.ctx, batchUpdateTaskOrderParams)
	br.Exec(func(i int, err error) {
		if err != nil {
			t.logger.Error("failed to execute query in batch", slog.Int("queryNumber", i), slog.String("err", err.Error()))
			errs = append(errs, err)
			br.Close()
		}
	})

	if len(errs) == 0 {
		return nil
	}

	return errs[len(errs)-1]
}

// Update task status to Pending or Completed
func (t *TaskRepositoryPostgres) UpdateTaskStatus(id uuid.UUID, newStatus TaskStatus) (_ error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return err
	}

	return t.Queries.UpdateTaskStatus(t.ctx, db.UpdateTaskStatusParams{
		ID: pgUUID, Status: newStatus.String(),
	})
}

// Delete the task with the specified ID
func (t *TaskRepositoryPostgres) Delete(id uuid.UUID) (_ Task, _ error) {
	task, err := t.Get(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = internal.NewNotFoundError(fmt.Sprintf("task %s", id.String()))
		}
		return Task{}, err
	}

	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		t.logger.Error("failed to delete task",
			slog.String("taskID", id.String()),
			slog.String("err", err.Error()),
		)
		return Task{}, err
	}

	return task, t.Queries.DeleteTask(t.ctx, pgUUID)
}
