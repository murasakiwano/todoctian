package task

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DeleteTaskTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *DeleteTaskTestSuite) SetupSuite() {
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

	repository := NewTaskRepositoryPostgres(suite.ctx, pgPool)
	projectRepository := project.NewProjectRepositoryPostgres(suite.ctx, pgPool)

	suite.taskService = NewTaskService(repository, projectRepository)
}

// Setup database before each test
func (suite *DeleteTaskTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *DeleteTaskTestSuite) TestSuccess() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("Test task", suite.projectID, nil)
	require.NoError(t, err)

	_, err = suite.taskService.DeleteTask(task.ID)
	require.NoError(t, err)

	task, err = suite.taskService.repository.Get(task.ID)
	require.Error(t, err)

	assert.True(t, errors.Is(err, internal.ErrNotFound))
}

func (suite *DeleteTaskTestSuite) TestTaskDoesNotExist() {
	t := suite.T()

	id := uuid.New()
	_, err := suite.taskService.DeleteTask(id)
	assert.Error(t, err)
}

func (suite *DeleteTaskTestSuite) TestAlsoDeletesSubtasks() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("Test task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	_, err = suite.taskService.DeleteTask(task.ID)
	require.NoError(t, err)

	task, err = suite.taskService.FindTaskByID(task.ID)
	require.Error(t, err)
	require.True(t,
		errors.Is(err, internal.ErrNotFound),
		"an unexpected error occurred: %s",
		err,
	)

	subtask, err = suite.taskService.FindTaskByID(subtask.ID)
	if assert.Error(t, err) {
		assert.True(t,
			errors.Is(err, internal.ErrNotFound),
			"an unexpected error occurred: %s",
			err,
		)
	}
}

func (suite *DeleteTaskTestSuite) TestRearrangesSiblingsOrders() {
	t := suite.T()

	firstTask, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	secondTask, err := suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	thirdTask, err := suite.taskService.CreateTask("Third task", suite.projectID, nil)
	require.NoError(t, err)

	_, err = suite.taskService.DeleteTask(secondTask.ID)
	require.NoError(t, err)

	firstTask, err = suite.taskService.FindTaskByID(firstTask.ID)
	require.NoError(t, err)

	thirdTask, err = suite.taskService.FindTaskByID(thirdTask.ID)
	require.NoError(t, err)

	assert.Equalf(t,
		0,
		firstTask.Order,
		"firstTask's order changed when it should not have changed: %d",
		firstTask.Order,
	)
	assert.Equalf(t,
		1,
		thirdTask.Order,
		"thirdTasks's order should have decreased to 1, but it's %d",
		thirdTask.Order,
	)
}

func (suite *DeleteTaskTestSuite) TestKeepsParentTaskTheSame() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("Test task", suite.projectID, nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	_, err = suite.taskService.DeleteTask(subtask.ID)
	require.NoError(t, err)

	_, err = suite.taskService.FindTaskByID(task.ID)
	assert.NoErrorf(t, err, "Parent task was supposed to remain intact, but an error occurred: %v", err)
}

func TestDeleteTask(t *testing.T) {
	suite.Run(t, new(DeleteTaskTestSuite))
}
