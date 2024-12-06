package project

import (
	"github.com/google/uuid"
)

// A Project is a collection of tasks. The only attribute releant to the user is
// its name.
type Project struct {
	Name  string      // Name of the project
	Tasks []uuid.UUID // IDs of the project's tasks
	ID    uuid.UUID   // ID of the project
}

// Create a new instance of a project.
func NewProject(name string) Project {
	return Project{
		ID:    uuid.New(),
		Name:  name,
		Tasks: []uuid.UUID{},
	}
}
