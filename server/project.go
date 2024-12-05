package main

import (
	"github.com/google/uuid"
)

// A Project is a collection of tasks. The only attribute releant to the user is
// its name.
type Project struct {
	Name  string
	Tasks []Task
	ID    uuid.UUID
}

// Create a new instance of a project.
func NewProject(name string) Project {
	return Project{
		ID:    uuid.New(),
		Name:  name,
		Tasks: []Task{},
	}
}

// Adds a task to the project.
func (p *Project) AddTask(task Task) {
	p.Tasks = append(p.Tasks, task)
}
