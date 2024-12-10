package project

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/murasakiwano/todoctian/server/testhelpers"
	"github.com/stretchr/testify/assert"
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
	repository, err := NewProjectRepositoryPostgres(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

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
	err := suite.service.RenameProject(project.ID, newName)
	if err != nil {
		t.Fatalf("expected rename to succeed, but got error: %v", err)
	}
}

func (suite *ProjectServiceTestSuite) TestRenameProject_ReuseOldNameAfterRename() {
	t := suite.T()
	oldName := "My test project"
	newName := "New test project name"

	project, _ := suite.service.CreateProject(oldName)
	err := suite.service.RenameProject(project.ID, newName)
	assert.NoError(t, err)

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
	err := suite.service.RenameProject(project2.ID, name)
	if err == nil {
		t.Fatal("expected renaming to a duplicate name to fail, but it succeeded")
	}
}

func (suite *ProjectServiceTestSuite) TestRenameProject_NoOpForSameName() {
	t := suite.T()
	name := "My test project"

	project, _ := suite.service.CreateProject(name)

	// Renaming to the same name should succeed and be a NOOP
	err := suite.service.RenameProject(project.ID, name)
	assert.NoError(t, err)
}

func TestProjectService(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}
