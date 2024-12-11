package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/internal/openapi"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/task"
)

type Server struct {
	DB             pgxpool.Pool
	TaskService    task.TaskService
	ProjectService project.ProjectService
}

// Get all projects
// (GET /projects)
func (s *Server) GetProjects(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Create a project.
// (POST /projects)
func (s *Server) PostProjects(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Delete a project. Also deletes the project's tasks.
// (DELETE /projects/{projectID})
func (s *Server) DeleteProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Get a single project
// (GET /projects/{projectID})
func (s *Server) GetProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Rename a project
// (PATCH /projects/{projectID})
func (s *Server) PatchProjectsProjectID(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Get all project's tasks.
// (GET /projects/{projectID}/tasks)
func (s *Server) GetProjectsProjectIDTasks(w http.ResponseWriter, r *http.Request, projectID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Get all tasks
// (GET /tasks)
func (s *Server) GetTasks(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Create a new task.
// (POST /tasks)
func (s *Server) PostTasks(w http.ResponseWriter, r *http.Request) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Delete a task.
// (DELETE /tasks/{taskID})
func (s *Server) DeleteTasksTaskID(w http.ResponseWriter, r *http.Request, taskID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Get a single task.
// (GET /tasks/{taskID})
func (s *Server) GetTasksTaskID(w http.ResponseWriter, r *http.Request, taskID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}

// Update a task's status.
// (PATCH /tasks/{taskID}/status)
func (s *Server) PatchTasksTaskIDStatus(w http.ResponseWriter, r *http.Request, taskID string) (_ *openapi.Response) {
	panic("not implemented") // TODO: Implement
}
