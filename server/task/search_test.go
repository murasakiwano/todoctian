package task

import (
	"context"
	"log"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SearchTaskTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	taskService *TaskService
	projectID   uuid.UUID
}

func (suite *SearchTaskTestSuite) SetupSuite() {
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
func (suite *SearchTaskTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)

	projectIDs := insertTestProjectsInTheDatabase(suite.ctx, t, suite.pgContainer.ConnectionString)
	suite.projectID = projectIDs[0]
}

func (suite *SearchTaskTestSuite) TestSearchTaskByProject() {
	t := suite.T()

	otherProject := project.NewProject("other project")
	err := suite.taskService.projectDB.Create(otherProject)
	require.NoError(t, err)

	firstTask, err := suite.taskService.CreateTask("test task", suite.projectID, nil)
	require.NoError(t, err)

	_, err = suite.taskService.CreateTask("second task", otherProject.ID, nil)
	require.NoError(t, err)

	tasks, err := suite.taskService.SearchTaskByProject(suite.projectID)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 1)
		assert.Equal(t, firstTask.Name, tasks[0].Name)
	}
}

func (suite *SearchTaskTestSuite) TestAcrossAllProjects() {
	t := suite.T()

	firstTask, err := suite.taskService.CreateTask("Test task", suite.projectID, nil)
	require.NoError(t, err)

	secondTask, err := suite.taskService.CreateTask("Test task for second project", suite.projectID, nil)
	require.NoError(t, err)

	unmatching := "unmatching"
	_, err = suite.taskService.CreateTask(unmatching, suite.projectID, nil)
	require.NoError(t, err)

	tasks, err := suite.taskService.SearchTaskName("tsk", suite.projectID)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 2)
		assert.True(t, slices.ContainsFunc(tasks, func(t Task) bool {
			return t.Name == firstTask.Name
		}))
		assert.True(t, slices.ContainsFunc(tasks, func(t Task) bool {
			return t.Name == secondTask.Name
		}))
		assert.False(t, slices.ContainsFunc(tasks, func(t Task) bool {
			return t.Name == unmatching
		}))
	}
}

func TestSearchTask(t *testing.T) {
	suite.Run(t, new(SearchTaskTestSuite))
}
