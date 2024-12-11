package task

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdateTaskStatusTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *UpdateTaskStatusTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = pgContainer
	pgPool, err := pgxpool.New(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	repository, err := NewTaskRepositoryPostgres(suite.ctx, pgPool)
	if err != nil {
		log.Fatal(err)
	}

	projectRepository, err := project.NewProjectRepositoryPostgres(suite.ctx,
		pgPool,
	)

	suite.taskService = NewTaskService(repository, projectRepository)
}

// Setup database before each test
func (suite *UpdateTaskStatusTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *UpdateTaskStatusTestSuite) TestCompleteWithoutSubtasks() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, err = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, TaskStatusCompleted, task.Status)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestCompleteWithSubtasks() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	nestedSubtask, err := suite.taskService.CreateTask("Nested subtask", suite.projectID, &subtask.ID)
	require.NoError(t, err)

	nestedSiblingSubstask, err := suite.taskService.CreateTask("Nested sibling subtask", suite.projectID, &subtask.ID)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			subtask.Status,
			"subtask was not marked completed successfully",
		)
	}

	nestedSubtask, err = suite.taskService.FindTaskByID(nestedSubtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			nestedSubtask.Status,
			"nestedSubtask was not marked completed successfully",
		)
	}

	nestedSiblingSubstask, err = suite.taskService.FindTaskByID(nestedSiblingSubstask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			nestedSiblingSubstask.Status,
			"nestedSiblingSubtask was not marked completed successfully",
		)
	}

	task, err = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestCompleteAlsoCompletesTheParentTask() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	nestedSubtask, err := suite.taskService.CreateTask("Nested subtask", suite.projectID, &subtask.ID)
	require.NoError(t, err)

	nestedSiblingSubstask, err := suite.taskService.CreateTask("Nested sibling subtask", suite.projectID, &subtask.ID)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(nestedSubtask.ID)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(nestedSiblingSubstask.ID)
	require.NoError(t, err)

	nestedSubtask, err = suite.taskService.FindTaskByID(nestedSubtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			nestedSubtask.Status,
			"nestedSubtask was not marked completed successfully",
		)
	}

	nestedSiblingSubstask, err = suite.taskService.FindTaskByID(nestedSiblingSubstask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			nestedSiblingSubstask.Status,
			"nestedSiblingSubtask was not marked completed successfully",
		)
	}

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			subtask.Status,
			"subtask was not marked completed successfully",
		)
	}

	task, err = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestCompleteTaskIsAlreadyCompleted() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestPending() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}

	err = suite.taskService.MarkTaskAsPending(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusPending,
			task.Status,
			"task was not marked as pending successfully",
		)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestPendingDoesNotMarkSubtasksAsPending() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	require.NoError(t, err)

	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			subtask.Status,
			"subtask was not marked completed successfully",
		)
	}

	err = suite.taskService.MarkTaskAsPending(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusPending,
			task.Status,
			"task was not marked as pending successfully",
		)
	}

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	require.NoError(t, err)

	if assert.NoError(t, err) {
		assert.NotEqual(t,
			TaskStatusPending,
			subtask.Status,
			"subtask was accidentally marked as pending",
		)
	}
}

func (suite *UpdateTaskStatusTestSuite) TestPendingMarksParentAsPending() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	nestedSubtask, err := suite.taskService.CreateTask("Nested subtask", suite.projectID, &subtask.ID)
	require.NoError(t, err)

	err = suite.taskService.CompleteTask(task.ID)
	require.NoError(t, err)

	task, _ = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			task.Status,
			"task was not marked completed successfully",
		)
	}

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			subtask.Status,
			"subtask was not marked completed successfully",
		)
	}

	nestedSubtask, err = suite.taskService.FindTaskByID(nestedSubtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusCompleted,
			nestedSubtask.Status,
			"nestedSubtask was not marked completed successfully",
		)
	}

	err = suite.taskService.MarkTaskAsPending(nestedSubtask.ID)
	require.NoError(t, err)

	nestedSubtask, err = suite.taskService.FindTaskByID(nestedSubtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusPending,
			nestedSubtask.Status,
			"nestedSubtask was not marked as pending successfully",
		)
	}

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusPending,
			subtask.Status,
			"subtask was not marked as pending successfully",
		)
	}

	task, err = suite.taskService.FindTaskByID(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			TaskStatusPending,
			task.Status,
			"task was not marked as pending successfully",
		)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	suite.Run(t, new(UpdateTaskStatusTestSuite))
}
