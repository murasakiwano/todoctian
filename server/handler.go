package todoctian

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/murasakiwano/todoctian/server/internal/openapi"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/murasakiwano/todoctian/server/task"
)

func Handler(taskService *task.TaskService, projectService *project.ProjectService) http.Handler {
	server := NewServer(taskService, projectService)
	return openapi.Handler(server, openapi.ServerOption(func(so *openapi.ServerOptions) {
		so.BaseRouter.Use(middleware.Logger)
	}))
}
