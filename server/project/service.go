package project

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/murasakiwano/todoctian/server/internal"
)

// This code indicates that a duplicate constraint was violated by the query
var ErrPgDuplicate = "23505"

type ProjectService struct {
	repository ProjectRepository
	logger     slog.Logger
}

func NewProjectService(db ProjectRepository) *ProjectService {
	return &ProjectService{repository: db, logger: *internal.NewLogger("ProjectService")}
}

// Creates and persists a project to the repository.
func (p *ProjectService) CreateProject(name string) (Project, error) {
	project, err := p.repository.GetByName(name)
	if err != nil && !errors.Is(err, internal.ErrNotFound) {
		p.logger.Error("failed to create project", slog.String("err", err.Error()))

		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ErrPgDuplicate {
			err = internal.NewAlreadyExistsError(fmt.Sprintf("Project \"%s\"", name))
		}

		return Project{}, err
	}

	project = NewProject(name)

	err = p.repository.Create(project)

	return project, err
}

func (p *ProjectService) GetProject(id uuid.UUID) (Project, error) {
	return p.repository.Get(id)
}

func (p *ProjectService) DeleteProject(id uuid.UUID) (Project, error) {
	_, err := p.repository.Get(id)
	if err != nil {
		p.logger.Error("failed to delete project", slog.String("err", err.Error()))
		return Project{}, err
	}

	return p.repository.Delete(id)
}

func (p *ProjectService) RenameProject(id uuid.UUID, newName string) (Project, error) {
	project, err := p.repository.Get(id)
	if err != nil {
		p.logger.Error("failed to rename project", slog.String("err", err.Error()))
		return Project{}, err
	}
	if project.Name == newName {
		return project, nil
	}

	// Check if another project with the new name already exists
	_, err = p.repository.GetByName(newName)
	if err == nil {
		p.logger.Error(
			"could not rename project - project with requested name already exists",
			slog.String("name", newName),
		)
		return Project{}, internal.NewAlreadyExistsError(fmt.Sprintf("Project with name \"%s\"", newName))
	}
	if !errors.Is(err, internal.ErrNotFound) {
		return Project{}, err
	}

	return p.repository.Rename(project.ID, newName)
}

func (p *ProjectService) ListProjects() ([]Project, error) {
	return p.repository.ListProjects()
}
