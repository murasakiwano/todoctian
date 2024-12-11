package task

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/murasakiwano/todoctian/server/project"
)

func insertTestProjectsInTheDatabase(ctx context.Context, t *testing.T, connectionString string) []uuid.UUID {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		t.Fatalf("unable to connect to the database: %s", err)
	}
	defer conn.Close(ctx)

	t.Log("inserting sample projects")
	testProject := project.NewProject("Test project")
	otherTestProject := project.NewProject("Other test project")
	createProject := `
INSERT INTO projects (
  id, name, created_at
) VALUES (
  $1, $2, $3
)
`
	_, err = conn.Exec(ctx, createProject, testProject.ID, testProject.Name, testProject.CreatedAt)
	if err != nil {
		t.Fatalf("failed to insert project into the database: %s", err)
	}

	_, err = conn.Exec(ctx, createProject, otherTestProject.ID, otherTestProject.Name, otherTestProject.CreatedAt)
	if err != nil {
		t.Fatalf("failed to insert project into the database: %s", err)
	}

	return []uuid.UUID{testProject.ID, otherTestProject.ID}
}
