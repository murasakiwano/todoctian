package main

import (
	"github.com/google/uuid"
)

type TaskRepository interface {
	SaveTask(task Task) error
	GetTask(id uuid.UUID) (*Task, error)
	UpdateTask(id uuid.UUID, newTaskInfo Task) (*Task, error)
	DeleteTask(id uuid.UUID) (*Task, error)
	CompleteTask(id uuid.UUID) error
}
