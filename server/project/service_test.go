package project

import (
	"context"
	"log"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProjectServiceTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	service     *ProjectService
	ctx         context.Context
}

func (suite *ProjectServiceTestSuite) SetupSuite() {
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

	suite.service = NewProjectService(repository)
}

func (suite *ProjectServiceTestSuite) SetupTest() {
	t := suite.T()
	t.Log("Cleaning up database before test...")
	conn, err := pgx.Connect(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v\n", err)
	}
	conn.Query(suite.ctx, "DELETE FROM projects;") // Cleanup everything before each test
	conn.Close(suite.ctx)
}

func (suite *ProjectServiceTestSuite) TestCreateProject_Success() {
	t := suite.T()

	_, err := suite.service.CreateProject("My test project")
	assert.NoError(t, err)
}

func (suite *ProjectServiceTestSuite) TestCreateProject_AlreadyExists() {
	t := suite.T()

	_, err := suite.service.CreateProject("My test project")
	assert.NoError(t, err)

	_, err = suite.service.CreateProject("My test project")
	assert.Error(t, err)
}

func (suite *ProjectServiceTestSuite) TestDeleteProject_Success() {
	t := suite.T()

	project, err := suite.service.CreateProject("My test project")
	assert.NoError(t, err)

	_, err = suite.service.DeleteProject(project.ID)
	assert.NoError(t, err)
}

func (suite *ProjectServiceTestSuite) TestDeleteProject_NonExistentProject() {
	t := suite.T()

	// Attempt to delete a non-existent project
	_, err := suite.service.DeleteProject(uuid.New())
	assert.Error(t, err)
}

func (suite *ProjectServiceTestSuite) TestRenameProject_SuccessfulRename() {
	t := suite.T()
	oldName := "My test project"
	newName := "New test project name"

	project, _ := suite.service.CreateProject(oldName)
	project, err := suite.service.RenameProject(project.ID, newName)
	if assert.NoError(t, err) {
		assert.Equal(t, newName, project.Name)
	}
}

func (suite *ProjectServiceTestSuite) TestRenameProject_ReuseOldNameAfterRename() {
	t := suite.T()
	oldName := "My test project"
	newName := "New test project name"

	project, _ := suite.service.CreateProject(oldName)
	project, err := suite.service.RenameProject(project.ID, newName)
	if assert.NoError(t, err) {
		assert.Equal(t, newName, project.Name)
	}

	// Should allow creating a new project with the old name
	_, err = suite.service.CreateProject(oldName)
	assert.NoError(t, err)
}

func (suite *ProjectServiceTestSuite) TestRenameProject_FailOnDuplicateName() {
	t := suite.T()
	name := "My test project"

	_, _ = suite.service.CreateProject(name)
	project2, _ := suite.service.CreateProject("Another project")

	// Renaming project2 to the same name as project1 should fail
	_, err := suite.service.RenameProject(project2.ID, name)
	assert.Error(t, err)
}

func (suite *ProjectServiceTestSuite) TestRenameProject_NoOpForSameName() {
	t := suite.T()
	name := "My test project"

	project, _ := suite.service.CreateProject(name)

	// Renaming to the same name should succeed and be a NOOP
	project, err := suite.service.RenameProject(project.ID, name)
	if assert.NoError(t, err) {
		assert.Equal(t, name, project.Name)
	}
}

func (suite *ProjectServiceTestSuite) TestListProjects() {
	t := suite.T()
	firstProject, err := suite.service.CreateProject("test project")
	require.NoError(t, err)

	secondProject, err := suite.service.CreateProject("second test project")
	require.NoError(t, err)

	projectList, err := suite.service.ListProjects()
	if assert.NoError(t, err) {
		assert.Len(t, projectList, 2)
		assert.True(t, slices.ContainsFunc(projectList, func(p Project) bool {
			return p.ID == firstProject.ID
		}))
		assert.True(t, slices.ContainsFunc(projectList, func(p Project) bool {
			return p.ID == secondProject.ID
		}))
	}
}

func TestProjectService(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}
