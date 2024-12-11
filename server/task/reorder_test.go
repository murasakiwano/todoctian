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

type ReorderTaskTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *ReorderTaskTestSuite) SetupSuite() {
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
func (suite *ReorderTaskTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *ReorderTaskTestSuite) TestKeepOrder() {
	t := suite.T()

	firstTask, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	_, err = suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.ReorderTask(firstTask, 0)
	require.NoError(t, err, "failed to reorder task with the same order it had: %s", err)

	firstTask, err = suite.taskService.repository.Get(firstTask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			0,
			firstTask.Order,
			"task should not have changed order, but it was moved to %d",
			firstTask.Order,
		)
	}
}

func (suite *ReorderTaskTestSuite) TestIncreaseOrder() {
	t := suite.T()

	firstTask, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	_, err = suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.ReorderTask(firstTask, 1)
	require.NoError(t, err)

	firstTask, err = suite.taskService.repository.Get(firstTask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			1,
			firstTask.Order,
			"task should have increased order to 1, but it was %d",
			firstTask.Order,
		)
	}
}

func (suite *ReorderTaskTestSuite) TestDecreaseOrder() {
	t := suite.T()

	_, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	secondTask, err := suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.ReorderTask(secondTask, 0)
	require.NoError(t, err)

	secondTask, err = suite.taskService.repository.Get(secondTask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t,
			0,
			secondTask.Order,
			"task should have decreased order to 0, but it was %d",
			secondTask.Order,
		)
	}
}

func (suite *ReorderTaskTestSuite) TestOrderOutOfBounds() {
	t := suite.T()

	_, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	secondTask, err := suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	err = suite.taskService.ReorderTask(secondTask, -10)
	require.NoError(t, err)

	secondTask, err = suite.taskService.repository.Get(secondTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if secondTask.Order != 0 {
		t.Fatalf("task should have decreased order to 0, but it was %d", secondTask.Order)
	}

	if assert.NoError(t, err) {
		assert.Equal(t,
			0,
			secondTask.Order,
			"task should have decreased order to 0, but it was %d",
			secondTask.Order,
		)
	}
}

func (suite *ReorderTaskTestSuite) TestTaskDoesNotExist() {
	task := NewTask("Test task", uuid.New(), nil)
	err := suite.taskService.ReorderTask(task, 0)
	require.Error(suite.T(), err)
}

func TestReorderTask(t *testing.T) {
	suite.Run(t, new(ReorderTaskTestSuite))
}
