package testhelpers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("..", "db", "schema.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

// Connect to the database and run `DELETE FROM tasks`
func CleanupTasksTable(ctx context.Context, t *testing.T, connectionString string) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		t.Fatalf("unable to connect to the database: %s", err)
	}
	defer conn.Close(ctx)

	t.Log("cleaning up tasks table")
	cleanupProjects := "DELETE FROM tasks"
	rows, err := conn.Query(ctx, cleanupProjects)
	if err != nil {
		t.Fatalf("failed to clean up tasks table: %s", err)
	}
	rows.Close()
}

// Connect to the database and run `DELETE FROM projects`
func CleanupProjectsTable(ctx context.Context, t *testing.T, connectionString string) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		t.Fatalf("unable to connect to the database: %s", err)
	}
	defer conn.Close(ctx)

	t.Log("cleaning up projects table")
	cleanupProjects := "DELETE FROM projects"
	rows, err := conn.Query(ctx, cleanupProjects)
	if err != nil {
		t.Fatalf("failed to clean up projects table: %s", err)
	}
	rows.Close()
}
