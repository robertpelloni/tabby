package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

// Task represents an autonomous task being executed by the agent
type Task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	Progress    float64    `json:"progress"` // 0.0 to 1.0
}

// Manager orchestrates agent tasks
type Manager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

// NewManager creates a new agent manager
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

// RunTask starts a new task
func (m *Manager) RunTask(ctx context.Context, description string) (*Task, error) {
	task := &Task{
		ID:          uuid.New().String(),
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		Progress:    0,
	}

	m.mu.Lock()
	m.tasks[task.ID] = task
	m.mu.Unlock()

	// Start task execution in a goroutine
	go m.executeTask(task.ID)

	return task, nil
}

// executeTask is a mock execution engine
func (m *Manager) executeTask(taskID string) {
	m.mu.Lock()
	task, ok := m.tasks[taskID]
	if !ok {
		m.mu.Unlock()
		return
	}
	now := time.Now()
	task.StartedAt = &now
	task.Status = StatusRunning
	m.mu.Unlock()

	// Simulate work
	for i := 1; i <= 5; i++ {
		time.Sleep(500 * time.Millisecond)
		m.mu.Lock()
		task.Progress = float64(i) / 5.0
		m.mu.Unlock()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	finish := time.Now()
	task.FinishedAt = &finish
	task.Status = StatusCompleted
	task.Result = fmt.Sprintf("Successfully completed: %s", task.Description)
}

// GetTask returns a task by ID
func (m *Manager) GetTask(id string) (*Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[id]
	return task, ok
}

// ListTasks returns all tasks
func (m *Manager) ListTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}
