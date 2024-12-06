package project

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateProject_Success(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())

	_, err := service.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}
}

func TestCreateProject_AlreadyExists(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())

	_, err := service.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected first project creation to succeed, but got error: %v", err)
	}

	_, err = service.CreateProject("My test project")
	if err == nil {
		t.Fatal("expected second project creation to fail due to duplicate name, but it succeeded")
	}
}

func TestDeleteProject_Success(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())

	project, err := service.CreateProject("My test project")
	if err != nil {
		t.Fatalf("expected project creation to succeed, but got error: %v", err)
	}

	_, err = service.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("expected project deletion to succeed, but got error: %v", err)
	}
}

func TestDeleteProject_NonExistentProject(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())

	// Attempt to delete a non-existent project
	_, err := service.DeleteProject(uuid.New())
	if err == nil {
		t.Fatal("expected deleting a non-existent project to fail, but it succeeded")
	}
}

func TestRenameProject_SuccessfulRename(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())
	oldName := "My test project"
	newName := "New test project name"

	project, _ := service.CreateProject(oldName)
	err := service.RenameProject(project.ID, newName)
	if err != nil {
		t.Fatalf("expected rename to succeed, but got error: %v", err)
	}
}

func TestRenameProject_ReuseOldNameAfterRename(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())
	oldName := "My test project"
	newName := "New test project name"

	project, _ := service.CreateProject(oldName)
	err := service.RenameProject(project.ID, newName)
	if err != nil {
		t.Fatalf("expected rename to succeed, but got error: %v", err)
	}

	// Should allow creating a new project with the old name
	_, err = service.CreateProject(oldName)
	if err != nil {
		t.Fatalf("expected to create a project with the old name, but got error: %v", err)
	}
}

func TestRenameProject_FailOnDuplicateName(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())
	name := "My test project"

	_, _ = service.CreateProject(name)
	project2, _ := service.CreateProject("Another project")

	// Renaming project2 to the same name as project1 should fail
	err := service.RenameProject(project2.ID, name)
	if err == nil {
		t.Fatal("expected renaming to a duplicate name to fail, but it succeeded")
	}
}

func TestRenameProject_NoOpForSameName(t *testing.T) {
	service := NewProjectService(NewProjectRepositoryInMemory())
	name := "My test project"

	project, _ := service.CreateProject(name)

	// Renaming to the same name should succeed and be a NOOP
	err := service.RenameProject(project.ID, name)
	if err != nil {
		t.Fatalf("expected renaming to the same name to succeed, but got error: %v", err)
	}
}
