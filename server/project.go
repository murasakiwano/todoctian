package main

import (
	"github.com/google/uuid"
)

// A Project is a collection of tasks. The only attribute releant to the user is
// its name.
type Project struct {
	UUID  uuid.UUID // unique identifier for the project
	Name  string    // the project name
	Tasks []Task    // tasks associated with the project
}

// Create a new instance of a project.
func CreateProject(name string) Project {
	return Project{
		Name:  name,
		Tasks: []Task{},
	}
}

// Adds a task to the project.
func (p *Project) AddTask(task Task) {
	p.Tasks = append(p.Tasks, task)
}
