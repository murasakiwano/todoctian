package todoctian

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/internal/openapi"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/task"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	TestProjectName   = "Test project"
	SampleProjectName = "Sample project"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *HandlerTestSuite) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode is a simple utility to check the response code
// of the response
func checkResponseCode(t *testing.T, expected, actual int) {
	require.Equal(t, expected, actual, "status code was different than expected\n")
}

type HandlerTestSuite struct {
	suite.Suite
	pgContainer       *testhelpers.PostgresContainer
	taskRepository    *task.TaskRepositoryPostgres
	taskService       *task.TaskService
	projectRepository *project.ProjectRepositoryPostgres
	projectService    *project.ProjectService
	ctx               context.Context
	handler           http.Handler
	pool              *pgxpool.Pool
	router            *chi.Mux
}

func (suite *HandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := postgres.Run(suite.ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("db", "schema.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
	connStr, err := pgContainer.ConnectionString(suite.ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = &testhelpers.PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}

	pgPool, err := pgxpool.New(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	suite.pool = pgPool

	suite.projectRepository = project.NewProjectRepositoryPostgres(suite.ctx, pgPool)
	suite.taskRepository = task.NewTaskRepositoryPostgres(suite.ctx, pgPool)

	suite.projectService = project.NewProjectService(suite.projectRepository)
	suite.taskService = task.NewTaskService(suite.taskRepository, suite.projectRepository)

	suite.handler = Handler(suite.taskService, suite.projectService)

	r := chi.NewRouter()
	r.Mount("/", suite.handler)
	suite.router = r
}

// Setup database before each test
func (suite *HandlerTestSuite) SetupTest() {
	t := suite.T()
	t.Log("cleaning up database before test...")
	testhelpers.CleanupTasksTable(suite.ctx, t, suite.pgContainer.ConnectionString)
	testhelpers.CleanupProjectsTable(suite.ctx, t, suite.pgContainer.ConnectionString)
}

func (suite *HandlerTestSuite) insertTestProjectsInTheDatabase() []uuid.UUID {
	t := suite.T()
	t.Log("inserting sample projects")
	testProject := project.NewProject(TestProjectName)
	sampleProject := project.NewProject(SampleProjectName)

	insertProject := `
INSERT INTO projects (
  id, name, created_at
) VALUES (
  $1, $2, $3
)
`
	_, err := suite.pool.Exec(suite.ctx, insertProject, testProject.ID, testProject.Name, testProject.CreatedAt)
	require.NoError(t, err, "failed to insert project into the database: %s\n", err)

	_, err = suite.pool.Exec(suite.ctx, insertProject, sampleProject.ID, sampleProject.Name, sampleProject.CreatedAt)
	require.NoError(t, err, "failed to insert project into the database: %s\n", err)

	return []uuid.UUID{testProject.ID, sampleProject.ID}
}

func (suite *HandlerTestSuite) TestGetProjects_NoProjects() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/projects", nil)
	rr := executeRequest(req, suite)

	checkResponseCode(t, 200, rr.Code)

	var projects []openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &projects)
	if assert.NoError(t, err) {
		assert.Len(t, projects, 0)
	}
}

func (suite *HandlerTestSuite) TestGetProjects_ReturnsExistingProjects() {
	t := suite.T()

	_ = suite.insertTestProjectsInTheDatabase()

	req, _ := http.NewRequest("GET", "/projects", nil)
	rr := executeRequest(req, suite)

	checkResponseCode(t, 200, rr.Code)

	var projects []openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &projects)
	if assert.NoError(t, err) {
		assert.Len(t, projects, 2)
	}
}

func (suite *HandlerTestSuite) TestPostProjects_NoBody() {
	t := suite.T()

	req, _ := http.NewRequest("POST", "/projects", nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusBadRequest, rr.Code)
}

func (suite *HandlerTestSuite) TestPostProjects_PostOneProject() {
	t := suite.T()

	projectName := "test project"
	body := openapi.PostProjectsJSONRequestBody{
		Name: &projectName,
	}
	buff := bodyInBytes(t, body)
	req, _ := http.NewRequest("POST", "/projects", buff)
	rr := executeRequest(req, suite)

	checkResponseCode(t, http.StatusCreated, rr.Code)

	var project openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &project)
	if assert.NoError(t, err) {
		require.NotNil(t, project.Name)
		assert.Equal(t, projectName, *project.Name)
	}
}

func (suite *HandlerTestSuite) TestPostProject_FailsIfDuplicate() {
	t := suite.T()

	projectName := "test project"
	body := openapi.PostTasksJSONRequestBody{
		Name: &projectName,
	}
	buff := bodyInBytes(t, body)

	req, _ := http.NewRequest("POST", "/projects", buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusCreated, rr.Code)

	buff = bodyInBytes(t, body)
	req, _ = http.NewRequest("POST", "/projects", buff)
	rr = executeRequest(req, suite)
	checkResponseCode(t, http.StatusConflict, rr.Code)
}

func (suite *HandlerTestSuite) TestDeleteProject_DeletesAProject() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[0])
	req, _ := http.NewRequest("DELETE", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNoContent, rr.Code)

	var project openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &project)
	if assert.NoError(t, err) {
		assert.Equal(t, projectIDs[0].String(), *project.ID)
	}
}

func (suite *HandlerTestSuite) TestDeleteProject_FailsIfProjectDoesNotExist() {
	t := suite.T()

	projectID := uuid.New()
	reqPath := fmt.Sprintf("/projects/%s", projectID)
	req, _ := http.NewRequest("DELETE", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestGetProjectsProjectID_ProjectExists() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[0])
	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var project openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &project)
	if assert.NoError(t, err) {
		assert.Equal(t, projectIDs[0].String(), *project.ID)
	}
}

func (suite *HandlerTestSuite) TestGetProjectsProjectID_ProjectDoesNotExist() {
	t := suite.T()

	projectID := uuid.New()
	reqPath := fmt.Sprintf("/projects/%s", projectID)
	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestPatchProjectsProjectID_NoBody() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[0])
	req, _ := http.NewRequest("PATCH", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusBadRequest, rr.Code)
}

func (suite *HandlerTestSuite) TestPatchProjectsProjectID_RenamesProject() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[0])

	newName := "new project name"
	body := openapi.PatchProjectsProjectIDJSONRequestBody{
		Name: &newName,
	}
	buff := bodyInBytes(t, body)
	req, _ := http.NewRequest("PATCH", reqPath, buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var respBody openapi.Project
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	if assert.NoError(t, err) {
		require.NotNil(t, respBody.Name)
		assert.Equal(t, newName, *respBody.Name)
	}
}

func (suite *HandlerTestSuite) TestPatchProjectsProjectID_BadRequest() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[0])
	body := "{\"noname\":\"foo\"}"
	buff := bytes.NewBufferString(body)
	req, _ := http.NewRequest("PATCH", reqPath, buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusBadRequest, rr.Code)
}

func (suite *HandlerTestSuite) TestPatchProjectsProjectID_NameIsAlreadyTaken() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s", projectIDs[1]) // sample project

	newName := TestProjectName
	buff := bodyInBytes(t, openapi.PatchProjectsProjectIDJSONRequestBody{
		Name: &newName,
	})
	req, _ := http.NewRequest("PATCH", reqPath, buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusConflict, rr.Code)
}

func (suite *HandlerTestSuite) TestGetProjectsProjectIDTasks_NoTasksInProject() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s/tasks", projectIDs[0])

	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var tasks []openapi.Task
	err := json.Unmarshal(rr.Body.Bytes(), &tasks)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 0)
	}
}

func (suite *HandlerTestSuite) TestGetProjectsProjectIDTasks_ReturnsTasks() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	reqPath := fmt.Sprintf("/projects/%s/tasks", projectIDs[0])

	suite.taskService.CreateTask("first test task", projectIDs[0], nil)
	suite.taskService.CreateTask("second test task", projectIDs[0], nil)
	suite.taskService.CreateTask("third test task", projectIDs[0], nil)

	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var tasks []openapi.Task
	err := json.Unmarshal(rr.Body.Bytes(), &tasks)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 3)
	}
}

func (suite *HandlerTestSuite) TestGetProjectsProjectIDTasks_ProjectDoesNotExist() {
	t := suite.T()

	projectID := uuid.New()
	reqPath := fmt.Sprintf("/projects/%s/tasks", projectID)

	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestGetTasks_NoTasks() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/tasks", nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var tasks []openapi.Task
	err := json.Unmarshal(rr.Body.Bytes(), &tasks)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 0)
	}
}

func (suite *HandlerTestSuite) TestGetTasks_ReturnsTasks() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	suite.taskService.CreateTask("first test task", projectIDs[0], nil)
	suite.taskService.CreateTask("second test task", projectIDs[0], nil)
	suite.taskService.CreateTask("third test task", projectIDs[1], nil)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var tasks []openapi.Task
	err := json.Unmarshal(rr.Body.Bytes(), &tasks)
	if assert.NoError(t, err) {
		assert.Len(t, tasks, 3)
	}
}

func (suite *HandlerTestSuite) TestPostTasks_NoBody() {
	req, _ := http.NewRequest("POST", "/tasks", nil)
	rr := executeRequest(req, suite)
	checkResponseCode(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *HandlerTestSuite) TestPostTasks_CreatesATask() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	projectID := projectIDs[0].String()
	taskName := "test task"
	body := openapi.PostTasksJSONBody{
		Name:      &taskName,
		ProjectID: &projectID,
	}
	buff := bodyInBytes(t, body)

	req, _ := http.NewRequest("POST", "/tasks", buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusCreated, rr.Code)

	var task openapi.Task
	err := json.Unmarshal(rr.Body.Bytes(), &task)
	if assert.NoError(t, err) {
		require.NotNil(t, task.Name)
		assert.Equal(t, taskName, *task.Name)
	}
}

func (suite *HandlerTestSuite) TestPostTasks_ProjectDoesNotExist() {
	t := suite.T()

	projectID := uuid.New().String()
	taskName := "test task"
	body := openapi.PostTasksJSONBody{
		Name:      &taskName,
		ProjectID: &projectID,
	}
	buff := bodyInBytes(t, body)

	req, _ := http.NewRequest("POST", "/tasks", buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestPostTasks_BadRequest() {
	t := suite.T()

	_ = suite.insertTestProjectsInTheDatabase()
	taskName := "test task"
	body := openapi.PostTasksJSONBody{
		Name: &taskName,
	}
	buff := bodyInBytes(t, body)

	req, _ := http.NewRequest("POST", "/tasks", buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusBadRequest, rr.Code)
}

func (suite *HandlerTestSuite) TestGetTasksTaskID_TaskExists() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskName := "test task"
	task, err := suite.taskService.CreateTask(taskName, projectIDs[0], nil)
	require.NoError(t, err)

	reqPath := fmt.Sprintf("/tasks/%s", task.ID)
	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var taskOAPI openapi.Task
	err = json.Unmarshal(rr.Body.Bytes(), &taskOAPI)
	if assert.NoError(t, err) {
		require.NotNil(t, taskOAPI.ID)
		require.NotNil(t, taskOAPI.Name)
		require.NotNil(t, taskOAPI.ProjectID)
		assert.Equal(t, taskName, *taskOAPI.Name)
	}
}

func (suite *HandlerTestSuite) TestGetTasksTaskID_TaskExists_ReturnsSubtasks() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskName := "test task"
	task, err := suite.taskService.CreateTask(taskName, projectIDs[0], nil)
	require.NoError(t, err)

	subtask, err := suite.taskService.CreateTask("subtask", projectIDs[0], &task.ID)
	require.NoError(t, err)

	nestedSubtask, err := suite.taskService.CreateTask("subtask", projectIDs[0], &subtask.ID)
	require.NoError(t, err)

	reqPath := fmt.Sprintf("/tasks/%s?withSubtasks=true", task.ID)
	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	var taskOAPI openapi.Task
	err = json.Unmarshal(rr.Body.Bytes(), &taskOAPI)
	if assert.NoError(t, err) {
		require.NotEmpty(t, taskOAPI.Subtasks)
		require.NotNil(t, *taskOAPI.Name)
		assert.Equal(t, task.Name, *taskOAPI.Name)

		subtaskOAPI := taskOAPI.Subtasks[0]
		require.NotEmpty(t, subtaskOAPI.Subtasks)
		require.NotNil(t, *subtaskOAPI.Name)
		assert.Equal(t, subtask.Name, *subtaskOAPI.Name)

		nestedSubtaskOAPI := subtaskOAPI.Subtasks[0]
		require.Empty(t, nestedSubtaskOAPI.Subtasks)
		require.NotNil(t, *nestedSubtaskOAPI.Name)
		assert.Equal(t, nestedSubtask.Name, *nestedSubtaskOAPI.Name)
	}
}

func (suite *HandlerTestSuite) TestGetTasksTaskID_TaskDoesNotExist() {
	reqPath := fmt.Sprintf("/tasks/%s", uuid.New())
	req, _ := http.NewRequest("GET", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(suite.T(), http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestDeleteTasksTaskID_TaskIsDeleted() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	task, err := suite.taskService.CreateTask("test task", projectIDs[0], nil)
	require.NoError(t, err)

	reqPath := fmt.Sprintf("/tasks/%s", task.ID)
	req, _ := http.NewRequest("DELETE", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNoContent, rr.Code)

	var taskOAPI openapi.Task
	err = json.Unmarshal(rr.Body.Bytes(), &taskOAPI)
	require.NoError(t, err)
	require.NotNil(t, taskOAPI.Name)
	require.Equal(t, task.Name, *taskOAPI.Name)
	require.NotNil(t, taskOAPI.ID)
	require.Equal(t, task.ID.String(), *taskOAPI.ID)

	_, err = suite.taskService.FindTaskByID(task.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, internal.ErrNotFound))
}

func (suite *HandlerTestSuite) TestDeleteTasksTaskID_TaskDoesNotExist() {
	reqPath := fmt.Sprintf("/tasks/%s", uuid.New())
	req, _ := http.NewRequest("DELETE", reqPath, nil)
	rr := executeRequest(req, suite)
	checkResponseCode(suite.T(), http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestPatchTasksTaskIDStatus_TaskDoesNotExist() {
	t := suite.T()

	body := openapi.PatchTasksTaskIDStatusJSONRequestBody{Status: &openapi.TaskStatusCompleted}
	reqPath := fmt.Sprintf("/tasks/%s/status", uuid.New())
	req, _ := http.NewRequest("PATCH", reqPath, bodyInBytes(t, body))
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestPatchTasksTaskIDStatus_MarksPendingTaskAsCompleted() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskModel, err := suite.taskService.CreateTask("test task", projectIDs[0], nil)
	require.NoError(t, err)

	suite.checkMarkTaskAsCompleted(taskModel)
}

func (suite *HandlerTestSuite) TestPatchTasksTaskIDStatus_MarksPendingTaskAsPending() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskModel, err := suite.taskService.CreateTask("test task", projectIDs[0], nil)
	require.NoError(t, err)

	suite.checkMarkTaskAsPending(taskModel)
}

func (suite *HandlerTestSuite) TestPatchTasksTaskIDStatus_MarksCompletedTaskAsPending() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskModel, err := suite.taskService.CreateTask("test task", projectIDs[0], nil)
	require.NoError(t, err)
	err = suite.taskService.UpdateTaskStatus(taskModel.ID, task.TaskStatusCompleted.String())
	require.NoError(t, err)

	suite.checkMarkTaskAsPending(taskModel)
}

func (suite *HandlerTestSuite) TestPatchTasksTaskIDStatus_MarksCompletedTaskAsCompleted() {
	t := suite.T()

	projectIDs := suite.insertTestProjectsInTheDatabase()
	taskModel, err := suite.taskService.CreateTask("test task", projectIDs[0], nil)
	require.NoError(t, err)
	err = suite.taskService.UpdateTaskStatus(taskModel.ID, task.TaskStatusCompleted.String())
	require.NoError(t, err)

	suite.checkMarkTaskAsCompleted(taskModel)
}

func (suite *HandlerTestSuite) checkMarkTaskAsCompleted(taskModel task.Task) {
	suite.checkMarkTaskStatus(taskModel, openapi.TaskStatusCompleted)
}

func (suite *HandlerTestSuite) checkMarkTaskAsPending(taskModel task.Task) {
	suite.checkMarkTaskStatus(taskModel, openapi.TaskStatusPending)
}

func (suite *HandlerTestSuite) checkMarkTaskStatus(taskModel task.Task, taskStatus openapi.TaskStatus) {
	t := suite.T()
	reqPath := fmt.Sprintf("/tasks/%s/status", taskModel.ID)
	body := openapi.PatchTasksTaskIDStatusJSONRequestBody{
		Status: &taskStatus,
	}
	buff := bodyInBytes(t, body)
	req, _ := http.NewRequest("PATCH", reqPath, buff)
	rr := executeRequest(req, suite)
	checkResponseCode(t, http.StatusOK, rr.Code)

	slog.Info("request body", slog.Any("body", rr.Body))

	taskModel, err := suite.taskService.FindTaskByID(taskModel.ID)
	require.NoError(t, err)
	require.Equal(t, taskStatus.ToValue(), taskModel.Status.String())
}

func bodyInBytes(t *testing.T, body interface{}) *bytes.Buffer {
	bodystr, err := json.Marshal(body)
	require.NoError(t, err)

	return bytes.NewBuffer(bodystr)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
