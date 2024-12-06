package project

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/internal"
)

type ProjectRepositoryInMemory struct {
	projects map[uuid.UUID]Project
}

func NewProjectRepositoryInMemory() *ProjectRepositoryInMemory {
	return &ProjectRepositoryInMemory{projects: make(map[uuid.UUID]Project)}
}

func (r *ProjectRepositoryInMemory) Create(project Project) error {
	p, exists := r.projects[project.ID]

	if exists {
		return fmt.Errorf("Project %v already exists", p)
	}

	r.projects[project.ID] = project

	return nil
}

func (r *ProjectRepositoryInMemory) Get(id uuid.UUID) (Project, error) {
	p, exists := r.projects[id]
	if !exists {
		return Project{}, fmt.Errorf("Project with id %s does not exist", id)
	}

	return p, nil
}

func (r *ProjectRepositoryInMemory) Delete(id uuid.UUID) (Project, error) {
	p, exists := r.projects[id]
	if !exists {
		return Project{}, internal.NewNotFoundError(fmt.Sprintf("Project with UUID %s", id))
	}

	delete(r.projects, id)

	return p, nil
}

func (r *ProjectRepositoryInMemory) Rename(id uuid.UUID, newName string) error {
	p, exists := r.projects[id]
	if !exists {
		return internal.NewNotFoundError(fmt.Sprintf("Project with UUID %s", id))
	}

	p.Name = newName
	r.projects[p.ID] = p

	return nil
}

func (r *ProjectRepositoryInMemory) GetByName(name string) (Project, error) {
	for _, project := range r.projects {
		if project.Name == name {
			return project, nil
		}
	}

	return Project{}, internal.NewNotFoundError(fmt.Sprintf("Project \"%s\"", name))
}
