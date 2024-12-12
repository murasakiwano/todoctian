package project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/db"
	"github.com/murasakiwano/todoctian/server/internal"
)

type ProjectRepositoryPostgres struct {
	Queries *db.Queries
	ctx     context.Context
	logger  slog.Logger
}

func NewProjectRepositoryPostgres(ctx context.Context, pool *pgxpool.Pool) *ProjectRepositoryPostgres {
	return &ProjectRepositoryPostgres{
		Queries: db.New(pool),
		ctx:     ctx,
		logger:  *internal.NewLogger("ProjectRepositoryPostgres"),
	}
}

func (p *ProjectRepositoryPostgres) Create(project Project) error {
	pgUUID, err := internal.ScanUUID(project.ID)
	if err != nil {
		return err
	}

	pgCreatedAt := pgtype.Timestamp{}
	err = pgCreatedAt.Scan(project.CreatedAt)
	if err != nil {
		return err
	}

	p.logger.Info("Creating project", slog.Any("project", project))
	err = p.Queries.CreateProject(p.ctx, db.CreateProjectParams{
		ID:        pgUUID,
		Name:      project.Name,
		CreatedAt: pgCreatedAt,
	})
	if err != nil {
		p.logger.Error("failed to insert project in the database", slog.String("err", err.Error()))
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ErrPgDuplicate {
			err = internal.NewAlreadyExistsError(fmt.Sprintf("Project \"%s\"", project.ID))
		}
	}
	return err
}

func (p *ProjectRepositoryPostgres) Get(id uuid.UUID) (Project, error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		p.logger.Error("failed to scan id to a postgres uuid", slog.String("err", err.Error()))
		return Project{}, err
	}

	projectDB, err := p.Queries.GetProject(p.ctx, pgUUID)
	if err != nil {
		p.logger.Error("failed to retrieve project from database", slog.String("err", err.Error()))

		if errors.Is(err, pgx.ErrNoRows) {
			err = internal.NewNotFoundError(fmt.Sprintf("Project with id %s", id.String()))
		}
		return Project{}, err
	}

	return ProjectDBToProjectModel(projectDB)
}

func (p *ProjectRepositoryPostgres) GetByName(name string) (Project, error) {
	projectDB, err := p.Queries.GetProjectByName(p.ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = internal.NewNotFoundError(fmt.Sprintf("Project %s", name))
		}

		p.logger.Error("failed to get project by name",
			slog.String("project_name", name),
			slog.String("err", err.Error()),
		)

		return Project{}, err
	}

	return ProjectDBToProjectModel(projectDB)
}

func (prepo *ProjectRepositoryPostgres) ListProjects() ([]Project, error) {
	projectsDB, err := prepo.Queries.ListProjects(prepo.ctx)
	if err != nil {
		prepo.logger.Error("failed to list projects from the database", slog.String("err", err.Error()))

		return nil, err
	}

	projects := []Project{}
	for _, pDB := range projectsDB {
		p, err := ProjectDBToProjectModel(pDB)
		if err != nil {
			prepo.logger.Error("could not adapt DB project to the project model", slog.String("err", err.Error()))
			return nil, err
		}

		projects = append(projects, p)
	}

	return projects, nil
}

func (p *ProjectRepositoryPostgres) Rename(id uuid.UUID, newName string) (Project, error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return Project{}, err
	}

	projectDB, err := p.Queries.RenameProject(p.ctx, db.RenameProjectParams{
		ID:   pgUUID,
		Name: newName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, internal.NewNotFoundError(fmt.Sprintf("project %s", id))
		}

		return Project{}, err
	}

	return ProjectDBToProjectModel(projectDB)
}

func (p *ProjectRepositoryPostgres) Delete(id uuid.UUID) (Project, error) {
	pgUUID, err := internal.ScanUUID(id)
	if err != nil {
		return Project{}, err
	}

	project, err := p.Queries.DeleteProject(p.ctx, pgUUID)
	if err != nil {
		return Project{}, err
	}

	return ProjectDBToProjectModel(project)
}
