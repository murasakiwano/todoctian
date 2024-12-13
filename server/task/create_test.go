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

type CreateTaskTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *CreateTaskTestSuite) SetupSuite() {
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
func (suite *CreateTaskTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *CreateTaskTestSuite) TestSuccess() {
	_, err := suite.taskService.CreateTask("My test task", suite.projectID, nil)
	require.NoError(suite.T(), err, "expected task creation to succeed, but got error: %v", err)
}

func (suite *CreateTaskTestSuite) TestUpdatesParentTask() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("My test task", suite.projectID, nil)
	require.NoError(t, err, "expected task creation to succeed, but got error: %v", err)

	subtask, err := suite.taskService.CreateTask("My subtask", suite.projectID, &task.ID)
	require.NoError(t, err, "expected subtask creation to succeed, but it failed: %v", err)

	subtasks, err := suite.taskService.repository.GetSubtasksDirect(task.ID)
	require.NoError(t, err)

	require.NotEmpty(t, subtasks, "adding a subtask to a task did not successfully update the parent task")
	assert.Equal(t, subtasks[0].ID, subtask.ID, "adding a subtask to a task did not successfully update the parent task")

	subtask, err = suite.taskService.CreateTask("My second subtask", suite.projectID, &task.ID)
	require.NoError(t, err)

	subtasks, err = suite.taskService.repository.GetSubtasksDirect(task.ID)
	require.NoError(t, err)

	require.NotEmpty(t, subtasks, "adding a subtask to a task did not successfully update the parent task")
	assert.Equal(t, subtasks[1].ID, subtask.ID, "adding a subtask to a task did not successfully update the parent task")
}

func (suite *CreateTaskTestSuite) TestProjectDoesNotExist() {
	_, err := suite.taskService.CreateTask("My test task", uuid.New(), nil)
	assert.Error(suite.T(), err, "expected task creation without an existing project to fail, which didn't happen.")
}

func (suite *CreateTaskTestSuite) TestParentTaskIsInvalid() {
	t := suite.T()

	parentID := uuid.New()
	_, err := suite.taskService.CreateTask("My test task", suite.projectID, &parentID)
	assert.Error(t, err, "expected task creation with an invalid parent task to fail")
}

func (suite *CreateTaskTestSuite) TestSetsOrderCorrectly() {
	t := suite.T()

	task, err := suite.taskService.CreateTask("First task", suite.projectID, nil)
	require.NoError(t, err)

	assert.Equal(
		t, 0, task.Order,
		"expected first task to have order 0, it actually had %d", task.Order,
	)

	task, err = suite.taskService.CreateTask("Second task", suite.projectID, nil)
	require.NoError(t, err)

	assert.Equal(
		t, 1, task.Order,
		"expected second task to have order 1, it actually had %d", task.Order,
	)
}

func (suite *CreateTaskTestSuite) TestSubtaskDoesNotAffectParentTaskOrder() {
	t := suite.T()

	parentTask, err := suite.taskService.CreateTask("Parent task", suite.projectID, nil)
	require.NoError(t, err)

	assert.Equal(
		t, 0, parentTask.Order,
		"expected parent task to have order 0, it actually had %d", parentTask.Order,
	)

	subtask, err := suite.taskService.CreateTask("Subtask", suite.projectID, &parentTask.ID)
	require.NoError(t, err)

	assert.Equal(
		t, 0, subtask.Order,
		"expected subtask to have order 0, it actually had %d", parentTask.Order,
	)

	parentTask, err = suite.taskService.repository.Get(parentTask.ID)
	require.NoError(t, err)

	assert.Equal(
		t, 0, parentTask.Order,
		"expected parent task's order to remain 0, it actually was %d", parentTask.Order,
	)
}

func (suite *CreateTaskTestSuite) TestDoesNotAllowParentTaskFromAnotherProject() {
	t := suite.T()

	parentTask, err := suite.taskService.CreateTask("Parent task", suite.projectID, nil)
	require.NoError(t, err)

	otherProject := project.NewProject("Other project")
	err = suite.taskService.projectDB.Create(otherProject)
	require.NoError(t, err)
	_, err = suite.taskService.CreateTask("Subtask", otherProject.ID, &parentTask.ID)
	require.Error(t, err)
}

func TestCreateTask(t *testing.T) {
	suite.Run(t, new(CreateTaskTestSuite))
}
