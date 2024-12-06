package task

import (
	"slices"
	"testing"
)

func TestSearchTask_AcrossAllProjects(t *testing.T) {
	taskService, projectService := setupServices()

	project, err := projectService.CreateProject("Test project")
	if err != nil {
		t.Fatal(err)
	}

	firstTask, err := taskService.CreateTask("Test task", project.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	secondProject, err := projectService.CreateProject("Second project")
	if err != nil {
		t.Fatal(err)
	}

	secondTask, err := taskService.CreateTask("Test task for second project", secondProject.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := taskService.SearchTaskName("test", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !(slices.ContainsFunc(tasks, func(t Task) bool {
		return t.Name == firstTask.Name
	}) && slices.ContainsFunc(tasks, func(t Task) bool {
		return t.Name == secondTask.Name
	})) {
		t.Fatalf("expected slice to contain both tasks that were added: %v", tasks)
	}
}
