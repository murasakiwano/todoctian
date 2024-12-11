package project

import (
	"context"
	"log"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectRepoPostgresTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	repository  *ProjectRepositoryPostgres
	ctx         context.Context
}

func (suite *ProjectRepoPostgresTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = pgContainer
	pgPool, err := pgxpool.New(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	repository := NewProjectRepositoryPostgres(suite.ctx, pgPool)

	suite.repository = repository
}

// Cleanup database before each test
func (suite *ProjectRepoPostgresTestSuite) SetupTest() {
	t := suite.T()
	t.Log("Cleaning up database before test...")
	conn, err := pgx.Connect(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v\n", err)
	}
	_, err = conn.Query(suite.ctx, "DELETE FROM projects;") // Cleanup everything before each test
	if err != nil {
		t.Fatalf("Failed to cleanup database: %v\n", err)
	}
	conn.Close(suite.ctx)
}

func (suite *ProjectRepoPostgresTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *ProjectRepoPostgresTestSuite) TestCreateProject() {
	t := suite.T()

	err := suite.repository.Create(NewProject("test project"))
	assert.NoError(t, err)
}

func (suite *ProjectRepoPostgresTestSuite) TestGetProject() {
	t := suite.T()

	projectName := "test project"
	project := NewProject(projectName)
	assert.NoError(t, suite.repository.Create(project))

	projectID := project.ID

	project, err := suite.repository.Get(project.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, projectName, project.Name)
		assert.Equal(t, projectID, project.ID)
	}
}

func (suite *ProjectRepoPostgresTestSuite) TestGetProjectByName() {
	t := suite.T()

	projectName := "test project"
	project := NewProject(projectName)
	assert.NoError(t, suite.repository.Create(project))

	projectID := project.ID

	project, err := suite.repository.GetByName(project.Name)
	if assert.NoError(t, err) {
		assert.Equal(t, projectName, project.Name)
		assert.Equal(t, projectID.String(), project.ID.String())
	}
}

func (suite *ProjectRepoPostgresTestSuite) TestListProjects() {
	t := suite.T()

	firstProjectName := "test project"
	firstProject := NewProject(firstProjectName)
	assert.NoError(t, suite.repository.Create(firstProject))
	secondProjectName := "second test project"
	secondProject := NewProject(secondProjectName)
	assert.NoError(t, suite.repository.Create(secondProject))

	projects, err := suite.repository.ListProjects()

	if assert.NoError(t, err) {
		assert.Len(t, projects, 2)
		assert.True(t, slices.ContainsFunc(projects, func(p Project) bool {
			return p.ID == firstProject.ID && p.Name == firstProjectName
		}))
		assert.True(t, slices.ContainsFunc(projects, func(p Project) bool {
			return p.ID == secondProject.ID && p.Name == secondProjectName
		}))
	}
}

func (suite *ProjectRepoPostgresTestSuite) TestRenameProject() {
	t := suite.T()

	project := NewProject("test project")
	assert.NoError(t, suite.repository.Create(project))

	newName := "legit project"
	assert.NoError(t, suite.repository.Rename(project.ID, newName))

	project, err := suite.repository.Get(project.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, project.Name, newName)
	}
}

func (suite *ProjectRepoPostgresTestSuite) TestDeleteProject() {
	t := suite.T()

	project := NewProject("test project")
	assert.NoError(t, suite.repository.Create(project))

	deletedProject, err := suite.repository.Delete(project.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, project.ID, deletedProject.ID)
		assert.Equal(t, project.Name, deletedProject.Name)
	}
}

func TestProjectRepositoryPostgresSuite(t *testing.T) {
	suite.Run(t, new(ProjectRepoPostgresTestSuite))
}
