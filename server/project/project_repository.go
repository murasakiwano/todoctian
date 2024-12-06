package project

import (
	"errors"

	"github.com/google/uuid"
)

var ErrProjectDoesNotExist = errors.New("Project does not exist")

type ProjectRepository interface {
	Create(project Project) error
	Get(id uuid.UUID) (Project, error)
	GetByName(name string) (Project, error)
	Delete(id uuid.UUID) (Project, error)
	Rename(id uuid.UUID, newName string) error
}
