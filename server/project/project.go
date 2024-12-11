package project

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// A Project is a collection of tasks. The only attribute releant to the user is
// its name.
type Project struct {
	CreatedAt time.Time
	Name      string      // Name of the project
	Tasks     []uuid.UUID // IDs of the project's tasks
	ID        uuid.UUID   // ID of the project
}

// Create a new instance of a project.
func NewProject(name string) Project {
	now := time.Now().UTC()
	return Project{
		ID:        uuid.New(),
		Name:      name,
		Tasks:     []uuid.UUID{},
		CreatedAt: now,
	}
}

// Will not log tasks in order to save log space.
func (p Project) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("ID", p.ID.String()),
		slog.String("Name", p.Name),
		slog.Time("CreatedAt", p.CreatedAt),
	)
}
