package main

import (
	"fmt"

	"github.com/google/uuid"
)

// Represents a to-do item. A task is either complete or incomplete, so we
// represent this with "To-do" and "Completed".
//
// A task may or may not have a parent task and, in turn, a list of subtasks.
type Task struct {
	parentTask *Task
	Name       string  // The name of the task
	Project    Project // The project this task belongs to
	subtasks   []*Task
	Status     taskStatus // Task status (whether it's completed or not)
	TaskUUID   uuid.UUID  // A unique identifier for the task
}

type taskStatus int

func (t taskStatus) String() string {
	return [...]string{"To-do", "Completed"}[t]
}

const (
	Todo taskStatus = iota
	Completed
)

// CreateTask returns a new instance of a task. It does not explicitly add the
// task to the project it belongs to!
func CreateTask(name string, project Project) Task {
	task := Task{
		Name:       name,
		Project:    project,
		Status:     Todo,
		parentTask: nil,
		subtasks:   []*Task{},
		TaskUUID:   uuid.New(),
	}

	return task
}

func (t *Task) ChangeTaskName(name string) {
	t.Name = name
}

// CompleteTask sets a task status to "Complete".
//
// When a task gets completed, all its subtasks are automatically completed too.
// Conversely, when all the subtasks are complete, the parent task also gets completed.
func (t *Task) CompleteTask() {
	t.Status = Completed

	for _, st := range t.subtasks {
		if st.Status != Completed {
			st.CompleteTask()
		}
	}

	if t.parentTask == nil {
		return
	}

	fmt.Println(t.parentTask)
	if t.parentTask.Status != Completed && t.parentTask.subtasksAreComplete() {
		fmt.Println("Completing parent task...", t.parentTask)
		t.parentTask.CompleteTask()
		fmt.Println("Completed parent task:", t.parentTask)
	}
}

func (t Task) subtasksAreComplete() bool {
	for _, st := range t.subtasks {
		if st.Status == Todo {
			return false
		}
	}

	return true
}

// ChangeTaskOrder rearranges a task in its parent's subtasks array.
//
// If the new index is greater then the subtasks' length, then the task is set as
// the last task in the list.
func (t *Task) ChangeTaskOrder(newIndex int) {
	// subtasks := t.parentTask.subtasks
	// n := len(subtasks)

	if newIndex <= 0 {
	}
}

// AddSubtask inserts a new subtask to the current task. It is always appended to
// the subtask list.
func (t *Task) AddSubtask(newTask *Task) {
	t.subtasks = append(t.subtasks, newTask)
	newTask.parentTask = t
}

// RemoveSubtask removes a specific task from the subtask list, if it exists.
//
// The method returns `true` and a Task if the subtask existed, otherwise it
// returns `false` and `nil`. Keep in mind that this only searches for direct
// subtasks, no further than 1 level deep.
func (t *Task) RemoveSubtask(taskUUID uuid.UUID) (bool, *Task) {
	for i, st := range t.subtasks {
		if st.TaskUUID == taskUUID {
			// TODO: does it work for the last element??
			t.subtasks = removeTask(t.subtasks, i)

			return true, st
		}
	}

	return false, nil
}

func removeTask(subtasks []*Task, i int) []*Task {
	if i >= len(subtasks) || i < 0 {
		return subtasks
	}

	copy(subtasks[i:], subtasks[i+1:])

	return subtasks[:len(subtasks)-1]
}
