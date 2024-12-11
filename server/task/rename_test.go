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

type RenameTaskTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *RenameTaskTestSuite) SetupSuite() {
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
func (suite *RenameTaskTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *RenameTaskTestSuite) TestSuccess() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("My test task", suite.projectID, nil)
	require.NoError(t, err)

	newTaskName := "My new test task"
	err = suite.taskService.RenameTask(task.ID, newTaskName)
	if assert.NoError(t, err) {
		task, _ = suite.taskService.repository.Get(task.ID)
		assert.Equal(t, newTaskName, task.Name)
	}
}

func (suite *RenameTaskTestSuite) TestTaskDoesNotExist() {
	t := suite.T()

	taskID := uuid.New()
	err := suite.taskService.RenameTask(taskID, "New task name")
	if err == nil {
		t.Fatal("expected renaming an inexistent task to fail, but it did not")
	}
}

func TestRenameTask(t *testing.T) {
	suite.Run(t, new(RenameTaskTestSuite))
}
