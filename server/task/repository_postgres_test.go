package task

import (
	"context"
	"errors"
	"log"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TaskRepoPostgresTestSuite struct {
	suite.Suite
	pgContainer    *testhelpers.PostgresContainer
	repository     *TaskRepositoryPostgres
	ctx            context.Context
	projectID      uuid.UUID
	otherProjectID uuid.UUID
}

func (suite *TaskRepoPostgresTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = pgContainer
	repository, err := NewTaskRepositoryPostgres(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	suite.repository = repository
}

// Setup database before each test
func (suite *TaskRepoPostgresTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	conn, err := pgx.Connect(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		t.Fatalf("unable to connect to database: %s\n", err)
	}
	defer conn.Close(suite.ctx)
	cleanupDB := `DELETE FROM tasks`
	r, err := conn.Query(suite.ctx, cleanupDB)
	if err != nil {
		t.Fatalf("failed to cleanup database: %s\n", err)
	}
	r.Close()

	cleanupDB = `DELETE FROM projects`
	r, err = conn.Query(suite.ctx, cleanupDB)
	if err != nil {
		t.Fatalf("failed to cleanup database: %s\n", err)
	}
	r.Close()

	suiteProject := project.NewProject("Test project")
	suite.projectID = suiteProject.ID
	otherProject := project.NewProject("Another test project")
	suite.otherProjectID = otherProject.ID
	createProject := `-- name: CreateProject :exec
INSERT INTO projects (
  id, name, created_at
) VALUES (
  $1, $2, $3
)
`
	_, err = conn.Exec(suite.ctx, createProject, suite.projectID, suiteProject.Name, suiteProject.CreatedAt)
	if err != nil {
		t.Fatalf("failed to create project: %s\n", err)
	}
	_, err = conn.Exec(suite.ctx, createProject, suite.otherProjectID, otherProject.Name, otherProject.CreatedAt)
	if err != nil {
		t.Fatalf("failed to create project: %s\n", err)
	}
}

func (suite *TaskRepoPostgresTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestCreateTask() {
	t := suite.T()

	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)

	assert.NoError(t, err)
}

func (suite *TaskRepoPostgresTestSuite) TestCreateDuplicateTaskShouldReturnAlreadyExistsError() {
	t := suite.T()

	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	err = suite.repository.Create(task)
	if assert.Error(t, err) {
		assert.True(t, errors.Is(err, internal.ErrAlreadyExists))
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetTask() {
	t := suite.T()

	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)

	require.NoError(t, err)
	retrievedTask, err := suite.repository.Get(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, task.ID, retrievedTask.ID)
		assert.Equal(t, task.Name, retrievedTask.Name)
		assert.Equal(t, task.ProjectID, retrievedTask.ProjectID)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetUnexistentTaskShouldReturnNotFoundError() {
	t := suite.T()

	_, err := suite.repository.Get(uuid.New())
	require.Error(t, err)

	assert.True(t, errors.Is(err, internal.ErrNotFound))
}

func (suite *TaskRepoPostgresTestSuite) TestGetSubtasksDirect() {
	t := suite.T()
	task := NewTask("Test parent task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	directSubtask := NewTask("Direct test subtask", suite.projectID, &task.ID)
	err = suite.repository.Create(directSubtask)
	require.NoError(t, err)

	secondDirectSubtask := NewTask("Second test direct subtask", suite.projectID, &task.ID)
	secondDirectSubtask.Order = 1
	err = suite.repository.Create(secondDirectSubtask)
	require.NoError(t, err)

	deepSubtask := NewTask("Test deep subtask", suite.projectID, &directSubtask.ID)
	err = suite.repository.Create(deepSubtask)
	require.NoError(t, err)

	subtasks, err := suite.repository.GetSubtasksDirect(task.ID)
	if assert.NoError(t, err) {
		assert.Len(t, subtasks, 2)
		assert.True(t, slices.ContainsFunc(subtasks, func(t Task) bool {
			return t.ID == directSubtask.ID
		}))
		assert.True(t, slices.ContainsFunc(subtasks, func(t Task) bool {
			return t.ID == secondDirectSubtask.ID
		}))
		assert.False(t, slices.ContainsFunc(subtasks, func(t Task) bool {
			return t.ID == deepSubtask.ID
		}))
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetSubtasksDeep() {
	t := suite.T()
	task := NewTask("Test parent task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	directSubtask := NewTask("Direct test subtask", suite.projectID, &task.ID)
	err = suite.repository.Create(directSubtask)
	require.NoError(t, err)

	secondDirectSubtask := NewTask("Second test direct subtask", suite.projectID, &task.ID)
	secondDirectSubtask.Order = 1
	err = suite.repository.Create(secondDirectSubtask)
	require.NoError(t, err)

	deepSubtask := NewTask("Test deep subtask", suite.projectID, &directSubtask.ID)
	err = suite.repository.Create(deepSubtask)
	require.NoError(t, err)

	deepSubtasks, err := suite.repository.GetSubtasksDeep(task.ID)
	if assert.NoError(t, err) {
		assert.Len(t, deepSubtasks, 3)
		assert.True(t, slices.ContainsFunc(deepSubtasks, func(t Task) bool {
			return t.ID == deepSubtask.ID
		}))
		assert.True(t, slices.ContainsFunc(deepSubtasks, func(t Task) bool {
			return t.ID == directSubtask.ID
		}))
		assert.True(t, slices.ContainsFunc(deepSubtasks, func(t Task) bool {
			return t.ID == secondDirectSubtask.ID
		}))
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetTasksByProject() {
	t := suite.T()

	firstProjectTask := NewTask("First project task", suite.projectID, nil)
	err := suite.repository.Create(firstProjectTask)
	require.NoError(t, err)

	secondProjectTask := NewTask("Second project task", suite.otherProjectID, nil)
	err = suite.repository.Create(secondProjectTask)
	require.NoError(t, err)

	projectTasks, err := suite.repository.GetTasksByProject(suite.projectID)
	require.NoError(t, err)
	if assert.True(t, len(projectTasks) == 1) {
		assert.Equal(t, firstProjectTask.ID, projectTasks[0].ID)
	}

	otherProjectTasks, err := suite.repository.GetTasksByProject(suite.otherProjectID)
	require.NoError(t, err)
	if assert.True(t, len(otherProjectTasks) == 1) {
		assert.Equal(t, secondProjectTask.ID, otherProjectTasks[0].ID)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetTasksInProjectRoot() {
	t := suite.T()
	task := NewTask("Test parent task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	rootSibling := NewTask("Test root sibling task", suite.projectID, nil)
	rootSibling.Order = 1
	err = suite.repository.Create(rootSibling)
	require.NoError(t, err)

	directSubtask := NewTask("Direct test subtask", suite.projectID, &task.ID)
	err = suite.repository.Create(directSubtask)
	require.NoError(t, err)

	secondDirectSubtask := NewTask("Second test direct subtask", suite.projectID, &task.ID)
	secondDirectSubtask.Order = 1
	err = suite.repository.Create(secondDirectSubtask)
	require.NoError(t, err)

	deepSubtask := NewTask("Test deep subtask", suite.projectID, &directSubtask.ID)
	err = suite.repository.Create(deepSubtask)
	require.NoError(t, err)

	rootTasks, err := suite.repository.GetTasksInProjectRoot(suite.projectID)
	if assert.NoError(t, err) {
		assert.Len(t, rootTasks, 2)
		assert.True(t, slices.ContainsFunc(rootTasks, func(t Task) bool {
			return t.ID == task.ID
		}))
		assert.True(t, slices.ContainsFunc(rootTasks, func(t Task) bool {
			return t.ID == rootSibling.ID
		}))
	}
}

func (suite *TaskRepoPostgresTestSuite) TestGetTasksByStatus() {
	t := suite.T()
	completedTask := NewTask("Test parent task", suite.projectID, nil)
	completedTask.Status = TaskStatusCompleted
	err := suite.repository.Create(completedTask)
	require.NoError(t, err)

	rootSibling := NewTask("Test root sibling task", suite.projectID, nil)
	rootSibling.Order = 1
	err = suite.repository.Create(rootSibling)
	require.NoError(t, err)

	completedTasks, err := suite.repository.GetTasksByStatus(suite.projectID, TaskStatusCompleted)
	if assert.NoError(t, err) {
		assert.Len(t, completedTasks, 1)
		assert.Equal(t, completedTasks[0].ID, completedTask.ID)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestRenameTask() {
	t := suite.T()
	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	newName := "New test task name"
	err = suite.repository.Rename(task.ID, newName)
	require.NoError(t, err)

	renamedTask, err := suite.repository.Get(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, task.ID, renamedTask.ID)
		assert.Equal(t, newName, renamedTask.Name)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestUpdateTaskOrder() {
	t := suite.T()
	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	assert.NoError(t, suite.repository.UpdateOrder(task.ID, 1), "order should be freely changed when there is no conflict")
}

func (suite *TaskRepoPostgresTestSuite) TestBatchUpdateTaskOrder() {
	t := suite.T()
	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	secondTask := NewTask("Second test task", suite.projectID, nil)
	secondTask.Order = 1
	err = suite.repository.Create(secondTask)
	require.NoError(t, err)

	task.Order = 1
	secondTask.Order = 0
	err = suite.repository.BatchUpdateOrder([]Task{task, secondTask})
	require.NoError(t, err)

	task, err = suite.repository.Get(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, task.Order)
	}

	secondTask, err = suite.repository.Get(secondTask.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, 0, secondTask.Order)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestUpdateTaskStatus() {
	t := suite.T()
	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	err = suite.repository.UpdateTaskStatus(task.ID, TaskStatusCompleted)
	require.NoError(t, err)

	completedTask, err := suite.repository.Get(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, TaskStatusCompleted, completedTask.Status)
	}
}

func (suite *TaskRepoPostgresTestSuite) TestDeleteTask() {
	t := suite.T()
	task := NewTask("Test task", suite.projectID, nil)
	err := suite.repository.Create(task)
	require.NoError(t, err)

	deletedTask, err := suite.repository.Delete(task.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, task.ID, deletedTask.ID)
	}

	_, err = suite.repository.Get(deletedTask.ID)
	require.Error(t, err)
}

func TestTaskRepositoryPostgres(t *testing.T) {
	suite.Run(t, new(TaskRepoPostgresTestSuite))
}
