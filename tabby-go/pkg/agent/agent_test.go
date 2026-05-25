package agent

import (
	"context"
	"testing"
	"time"
)

func TestManager_RunTask(t *testing.T) {
	mgr := NewManager(nil)
	ctx := context.Background()

	task, err := mgr.RunTask(ctx, "Test Task")
	if err != nil {
		t.Fatalf("Failed to run task: %v", err)
	}

	if task.ID == "" {
		t.Error("Task ID should not be empty")
	}

	if task.Description != "Test Task" {
		t.Errorf("Expected description 'Test Task', got '%s'", task.Description)
	}

	if task.Status != StatusPending && task.Status != StatusRunning {
		t.Errorf("Unexpected task status: %s", task.Status)
	}
}

func TestManager_GetTask(t *testing.T) {
	mgr := NewManager(nil)
	ctx := context.Background()

	task, _ := mgr.RunTask(ctx, "Test Task")

	retrieved, ok := mgr.GetTask(task.ID)
	if !ok {
		t.Fatalf("Failed to retrieve task %s", task.ID)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrieved.ID)
	}
}

func TestManager_ListTasks(t *testing.T) {
	mgr := NewManager(nil)
	ctx := context.Background()

	mgr.RunTask(ctx, "Task 1")
	mgr.RunTask(ctx, "Task 2")

	tasks := mgr.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestManager_TaskCompletion(t *testing.T) {
	mgr := NewManager(nil)
	ctx := context.Background()

	task, _ := mgr.RunTask(ctx, "Completion Task")

	// Wait for task to complete (mock execution takes ~2.5s)
	// Using a shorter wait for the test if possible, or just checking transition
	timeout := time.After(4 * time.Second)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for task completion")
		case <-tick.C:
			current, _ := mgr.GetTask(task.ID)
			if current.Status == StatusCompleted {
				if current.Progress != 1.0 {
					t.Errorf("Expected progress 1.0 on completion, got %f", current.Progress)
				}
				if current.FinishedAt == nil {
					t.Error("FinishedAt should be set")
				}
				return
			}
		}
	}
}
