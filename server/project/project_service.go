package project

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
)

type ProjectService struct {
	repository ProjectRepository
}

func NewProjectService(db ProjectRepository) *ProjectService {
	return &ProjectService{db}
}

// Creates and persists a project to the repository.
func (p *ProjectService) CreateProject(name string) (Project, error) {
	project, err := p.repository.GetByName(name)
	if !errors.Is(err, internal.ErrNotFound) {
		return Project{}, internal.NewAlreadyExistsError(fmt.Sprintf("Project \"%s\"", name))
	}

	project = NewProject(name)

	err = p.repository.Create(project)

	return project, err
}

func (p *ProjectService) DeleteProject(id uuid.UUID) (Project, error) {
	_, err := p.repository.Get(id)
	if err != nil {
		return Project{}, err
	}

	return p.repository.Delete(id)
}

func (p *ProjectService) RenameProject(id uuid.UUID, newName string) error {
	project, err := p.repository.Get(id)
	if err != nil {
		fmt.Printf("Project \"%s\" does not exist\n", id)
		return err
	}
	if project.Name == newName {
		fmt.Printf("Project already has name \"%s\", nothing to do.\n", newName)
		return nil
	}

	// Check if another project with the new name already exists
	_, err = p.repository.GetByName(newName)
	if err == nil {
		return internal.NewAlreadyExistsError(fmt.Sprintf("Project with name \"%s\"", newName))
	}
	if !errors.Is(err, internal.ErrNotFound) {
		return err
	}

	return p.repository.Rename(project.ID, newName)
}
