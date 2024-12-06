package task

import (
	"github.com/murasakiwano/todoctian/server/project"
)

func setupServices() (*TaskService, *project.ProjectService) {
	projectRepository := project.NewProjectRepositoryInMemory()
	taskService := NewTaskService(NewTaskRepositoryInMemory(), projectRepository)
	projectService := project.NewProjectService(projectRepository)

	return taskService, projectService
}
