package todoctian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/internal"
	"github.com/murasakiwano/todoctian/server/internal/openapi"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/task"
)

type Server struct {
	TaskService    *task.TaskService
	ProjectService *project.ProjectService
	logger         slog.Logger
}

func NewServer(connString string) *Server {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("could not connect to PostgreSQL: %s", err)
	}

	projectRepository := project.NewProjectRepositoryPostgres(ctx, pool)
	taskRepository := task.NewTaskRepositoryPostgres(ctx, pool)

	projectService := project.NewProjectService(projectRepository)
	taskService := task.NewTaskService(taskRepository, projectRepository)

	return &Server{
		TaskService:    taskService,
		ProjectService: projectService,
		logger:         *internal.NewLogger("Server"),
	}
}

// TODO: LOG ERRORS

// Get all projects
// (GET /projects)
func (s *Server) GetProjects(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	s.logger.Info("received request to GET /projects")
	projectList, err := s.ProjectService.ListProjects()
	if err != nil {
		s.logger.Error("could not list projects", slog.Any("err", err.Error()))

		internalServerError(w)
		return
	}

	s.logger.Debug("got project list", slog.Any("projects", projectList))

	projects := []openapi.Project{}
	for _, p := range projectList {
		pID := p.ID.String()
		projects = append(projects, openapi.Project{
			CreatedAt: &p.CreatedAt,
			ID:        &pID,
			Name:      &p.Name,
		})
	}

	resp := openapi.GetProjectsJSON200Response(projects)
	s.logger.Info("got response", slog.Any("response", resp))
	return resp
}

// Create a project.
// (POST /projects)
func (s *Server) PostProjects(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	if r.Body == nil {
		http.Error(w, "request body is required for this operation", http.StatusBadRequest)
		return
	}

	body := openapi.PostProjectsJSONBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil || body.Name == nil {
		slog.Error("failed to decode request body", slog.Any("err", err))
		http.Error(w, "body must be a json object with a \"name\" field", http.StatusBadRequest)
		return
	}

	projectName := *body.Name
	project, err := s.ProjectService.CreateProject(projectName)
	if err != nil {
		if errors.Is(err, internal.ErrAlreadyExists) {
			http.Error(w, fmt.Sprintf("project \"%s\" already exists", projectName), http.StatusConflict)
			return
		}

		s.logger.Error("an error occurred", slog.Any("err", err))
		internalServerError(w)
		return
	}

	projectOAPI := projectModelToProjectOAPI(project)
	return openapi.PostProjectsJSON201Response(projectOAPI)
}

// Delete a project. Also deletes the project's tasks.
// (DELETE /projects/{projectID})
func (s *Server) DeleteProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := s.ProjectService.DeleteProject(projectUUID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.Error(w, fmt.Sprintf("Could not find project %s", project.Name), http.StatusNotFound)
			return
		}
	}

	projectOAPI := projectModelToProjectOAPI(project)
	return openapi.DeleteProjectsProjectIDJSON204Response(projectOAPI)
}

// Get a single project
// (GET /projects/{projectID})
func (s *Server) GetProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "malformed project ID", http.StatusBadRequest)
		return
	}

	project, err := s.ProjectService.GetProject(projectUUID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	projectOAPI := projectModelToProjectOAPI(project)
	return openapi.GetProjectsProjectIDJSON200Response(projectOAPI)
}

// Rename a project
// (PATCH /projects/{projectID})
func (s *Server) PatchProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "malformed project ID", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "request body is required for this operation", http.StatusBadRequest)
		return
	}

	var params openapi.PatchProjectsProjectIDJSONRequestBody
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil || params.Name == nil {
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}

	project, err := s.ProjectService.RenameProject(projectUUID, *params.Name)
	if err != nil {
		if errors.Is(err, internal.ErrAlreadyExists) {
			http.Error(w, "project name already taken", http.StatusConflict)
			return
		}

		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	projectOAPI := projectModelToProjectOAPI(project)
	return openapi.PatchProjectsProjectIDJSON200Response(projectOAPI)
}

// Get all project's tasks.
// (GET /projects/{projectID}/tasks)
func (s *Server) GetProjectsProjectIDTasks(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "malformed project ID", http.StatusBadRequest)
		return
	}

	tasks, err := s.TaskService.SearchTaskByProject(projectUUID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	tasksOAPI := []openapi.Task{}
	for _, task := range tasks {
		taskOAPI, err := taskModelToTaskOAPI(task)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert task %s to OAPI model", task.ID), http.StatusInternalServerError)
			return
		}

		tasksOAPI = append(tasksOAPI, taskOAPI)
	}

	return openapi.GetProjectsProjectIDTasksJSON200Response(tasksOAPI)
}

// Get all tasks
// (GET /tasks)
func (s *Server) GetTasks(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	tasks, err := s.TaskService.ListTasks()
	if err != nil {
		internalServerError(w)
		return
	}

	tasksOAPI := []openapi.Task{}
	for _, task := range tasks {
		taskOAPI, err := taskModelToTaskOAPI(task)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert task %s to OAPI model", task.ID), http.StatusInternalServerError)
			return
		}

		tasksOAPI = append(tasksOAPI, taskOAPI)
	}

	return openapi.GetTasksJSON200Response(tasksOAPI)
}

// Create a new task.
// (POST /tasks)
func (s *Server) PostTasks(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	if r.Body == nil {
		http.Error(w, "request body is required for this operation", http.StatusBadRequest)
		return
	}

	var body openapi.PostTasksJSONRequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		http.Error(w, "malformed request body", http.StatusBadRequest)
		return
	}

	if body.Name == nil || *body.Name == "" {
		http.Error(w, "request body is missing task name", http.StatusBadRequest)
		return
	}
	if body.ProjectID == nil || *body.ProjectID == "" {
		http.Error(w, "request body is missing a project ID", http.StatusBadRequest)
		return
	}

	taskModel := task.Task{}
	taskModel.Name = *body.Name
	projectID, err := uuid.Parse(*body.ProjectID)
	if err != nil {
		http.Error(w, "malformed project id", http.StatusBadRequest)
		return
	}
	taskModel.ProjectID = projectID
	if body.ParentTaskID != nil && *body.ParentTaskID != "" {
		parentTaskID, err := uuid.Parse(*body.ParentTaskID)
		if err != nil {
			http.Error(w, "malformed parent task ID", http.StatusBadRequest)
			return
		}
		taskModel.ParentTaskID = &parentTaskID
	}

	task, err := s.TaskService.CreateTask(taskModel.Name, taskModel.ProjectID, taskModel.ParentTaskID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return openapi.PostTasksJSON404Response(openapi.Project{ID: body.ProjectID})
		}

		internalServerError(w)
		return
	}

	taskOAPI, err := taskModelToTaskOAPI(task)
	if err != nil {
		internalServerError(w)
		return
	}

	return openapi.PostTasksJSON201Response(taskOAPI)
}

// Delete a task.
// (DELETE /tasks/{taskID})
func (s *Server) DeleteTasksTaskID(w http.ResponseWriter, r *http.Request, taskID string) (_ *openapi.Response) {
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		http.Error(w, "malformed task ID", http.StatusBadRequest)
		return
	}

	deletedTask, err := s.TaskService.DeleteTask(taskUUID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	deletedTaskOAPI, err := taskModelToTaskOAPI(deletedTask)
	if err != nil {
		internalServerError(w)
		return
	}

	return openapi.DeleteTasksTaskIDJSON204Response(deletedTaskOAPI)
}

// Get a single task.
// (GET /tasks/{taskID})
func (s *Server) GetTasksTaskID(w http.ResponseWriter, r *http.Request, taskID string, params openapi.GetTasksTaskIDParams) (_ *openapi.Response) {
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		http.Error(w, "malformed task ID", http.StatusBadRequest)
		return
	}

	task, err := s.TaskService.FindTaskByID(taskUUID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	if params.WithSubtasks != nil && *params.WithSubtasks {
		subtasks, err := s.buildSubtasksStructure(taskUUID)
		if err != nil {
			s.logger.Error("failed to build subtasks structure", slog.String("taskID", taskUUID.String()), slog.Any("err", err.Error()))
			internalServerError(w)
			return
		}
		task.Subtasks = subtasks
	}

	taskOAPI, err := taskModelToTaskOAPI(task)
	if err != nil {
		internalServerError(w)
		return
	}

	return openapi.GetTasksTaskIDJSON200Response(taskOAPI)
}

// Update a task's status.
// (PATCH /tasks/{taskID}/status)
func (s *Server) PatchTasksTaskIDStatus(w http.ResponseWriter, r *http.Request, taskID string) (_ *openapi.Response) {
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		http.Error(w, "malformed task ID", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "request body is required for this operation", http.StatusBadRequest)
		return
	}

	var body openapi.PatchTasksTaskIDStatusJSONBody
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&body)
	if err != nil || body.Status == nil {
		http.Error(w, "malformed request body", http.StatusBadRequest)
		return
	}

	err = s.TaskService.UpdateTaskStatus(taskUUID, body.Status.ToValue())
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("task status updated successfully"))
	return
}

func projectOAPIToProjectModel(projectOAPI openapi.Project) (project.Project, error) {
	projectID, err := uuid.Parse(*projectOAPI.ID)
	if err != nil {
		return project.Project{}, err
	}

	return project.Project{
		ID:        projectID,
		Name:      *projectOAPI.Name,
		CreatedAt: *projectOAPI.CreatedAt,
	}, nil
}

func projectModelToProjectOAPI(projectModel project.Project) openapi.Project {
	projectIDString := projectModel.ID.String()

	return openapi.Project{
		ID:        &projectIDString,
		Name:      &projectModel.Name,
		CreatedAt: &projectModel.CreatedAt,
	}
}

func taskModelToTaskOAPI(taskModel task.Task) (openapi.Task, error) {
	taskID := taskModel.ID.String()
	projectID := taskModel.ProjectID.String()
	parentTaskID := ""
	if taskModel.ParentTaskID != nil {
		parentTaskID = taskModel.ParentTaskID.String()
	}

	taskStatus := openapi.TaskStatus{}
	err := taskStatus.FromValue(taskModel.Status.String())
	if err != nil {
		return openapi.Task{}, err
	}

	subtasks := []openapi.Task{}
	for _, st := range taskModel.Subtasks {
		stOAPI, err := taskModelToTaskOAPI(st)
		if err != nil {
			return openapi.Task{}, err
		}
		subtasks = append(subtasks, stOAPI)
	}

	return openapi.Task{
		CreatedAt:    &taskModel.CreatedAt,
		ID:           &taskID,
		Name:         &taskModel.Name,
		ParentTaskID: &parentTaskID,
		ProjectID:    &projectID,
		Status:       &taskStatus,
		Subtasks:     subtasks,
	}, nil
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func (s *Server) buildSubtasksStructure(taskUUID uuid.UUID) ([]task.Task, error) {
	subtasks, err := s.TaskService.FetchSubtasksDirect(taskUUID)
	if err != nil {
		return nil, err
	}

	for i, st := range subtasks {
		ssubtasks, err := s.buildSubtasksStructure(st.ID)
		if err != nil {
			return nil, err
		}

		st.Subtasks = ssubtasks
		subtasks[i] = st
	}

	return subtasks, nil
}
