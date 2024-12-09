package task

import (
	"testing"

	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/project"
	"github.com/stretchr/testify/assert"
)

func TestCreateTask_Success(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	assert.Nil(t, err, "expected project creation to succeed, but got error: %v", err)

	_, err = taskService.CreateTask("My test task", testProject.ID, nil)
	assert.Nil(t, err, "expected task creation to succeed, but got error: %v", err)
}

func TestCreateTask_UpdatesParentTask(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	assert.Nil(t, err, "expected project creation to succeed, but got error: %v", err)

	task, err := taskService.CreateTask("My test task", testProject.ID, nil)
	assert.Nil(t, err, "expected task creation to succeed, but got error: %v", err)

	subtask, err := taskService.CreateTask("My subtask", testProject.ID, &task.ID)
	assert.Nil(t, err, "expected subtask creation to succeed, but it failed: %v", err)

	task, err = taskService.taskDB.Get(task.ID)
	assert.Nil(t, err)

	assert.NotEmpty(t, task.Subtasks, "adding a subtask to a task did not successfully update the parent task")
	assert.Equal(t, task.Subtasks[0].ID, subtask.ID, "adding a subtask to a task did not successfully update the parent task")

	subtask, err = taskService.CreateTask("My second subtask", testProject.ID, &task.ID)
	assert.Nil(t, err)

	task, err = taskService.taskDB.Get(task.ID)
	assert.Nil(t, err)

	assert.NotEmpty(t, task.Subtasks, "adding a subtask to a task did not successfully update the parent task")
	assert.Equal(t, task.Subtasks[1].ID, subtask.ID, "adding a subtask to a task did not successfully update the parent task")
}

func TestCreateTask_ProjectDoesNotExist(t *testing.T) {
	taskService := NewTaskService(NewTaskRepositoryInMemory(), project.NewProjectRepositoryInMemory())

	_, err := taskService.CreateTask("My test task", uuid.New(), nil)
	assert.NotNil(t, err, "expected task creation without an existing project to fail, which didn't happen.")
}

func TestCreateTask_ParentTaskIsInvalid(t *testing.T) {
	taskService, projectService := setupServices()

	testProject, err := projectService.CreateProject("My test project")
	assert.Nil(t, err, "expected project creation to succeed, but got error: %v", err)

	parentID := uuid.New()
	_, err = taskService.CreateTask("My test task", testProject.ID, &parentID)
	assert.NotNil(t, err, "expected task creation with an invalid parent task to fail")
}

func TestCreateTask_SetsOrderCorrectly(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	assert.Nil(t, err)

	task, err := taskService.CreateTask("First task", project.ID, nil)
	assert.Nil(t, err)

	assert.Equal(
		t, 0, task.Order,
		"expected first task to have order 0, it actually had %d", task.Order,
	)

	task, err = taskService.CreateTask("Second task", project.ID, nil)
	assert.Nil(t, err)

	assert.Equal(
		t, 1, task.Order,
		"expected second task to have order 1, it actually had %d", task.Order,
	)
}

func TestCreateTask_SubtaskDoesNotAffectParentTaskOrder(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test Project")
	assert.Nil(t, err)

	parentTask, err := taskService.CreateTask("Parent task", project.ID, nil)
	assert.Nil(t, err)

	assert.Equal(
		t, 0, parentTask.Order,
		"expected parent task to have order 0, it actually had %d", parentTask.Order,
	)

	subtask, err := taskService.CreateTask("Subtask", project.ID, &parentTask.ID)
	assert.Nil(t, err)

	assert.Equal(
		t, 0, subtask.Order,
		"expected subtask to have order 0, it actually had %d", parentTask.Order,
	)

	parentTask, err = taskService.taskDB.Get(parentTask.ID)
	assert.Nil(t, err)

	assert.Equal(
		t, 0, parentTask.Order,
		"expected parent task's order to remain 0, it actually was %d", parentTask.Order,
	)
}
