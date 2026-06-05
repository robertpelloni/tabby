package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robertpelloni/tabby/tabby-go/pkg/vdom"
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

// Widget represents a rich UI component managed by the agent
type Widget struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // term, preview, web, sysinfo, vdom
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	VDOM      *vdom.Node `json:"vdom,omitempty"`
}

// NotifyFunc is the callback for sending task updates to the client
type NotifyFunc func(method string, params interface{})

// Manager orchestrates agent tasks and widgets
type Manager struct {
	tasks   map[string]*Task
	widgets map[string]*Widget
	mu      sync.RWMutex
	notify  NotifyFunc
}

// NewManager creates a new agent manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		tasks:   make(map[string]*Task),
		widgets: make(map[string]*Widget),
		notify:  notify,
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
	if description == "Environment Diagnostic" {
		go m.runDiagnostic(task.ID)
	} else {
		go m.executeTask(task.ID)
	}

	m.emitUpdate(task)

	return task, nil
}

func (m *Manager) emitUpdate(task *Task) {
	if m.notify != nil {
		m.notify("agent.taskUpdated", task)
	}
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

	m.emitUpdate(task)

	// Simulate work
	for i := 1; i <= 5; i++ {
		time.Sleep(500 * time.Millisecond)
		m.mu.Lock()
		task.Progress = float64(i) / 5.0
		m.mu.Unlock()
		m.emitUpdate(task)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	finish := time.Now()
	task.FinishedAt = &finish
	task.Status = StatusCompleted
	task.Result = fmt.Sprintf("Successfully completed: %s", task.Description)

	m.emitUpdate(task)
}

func (m *Manager) runDiagnostic(taskID string) {
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
	m.emitUpdate(task)

	binaries := []string{"go", "node", "yarn", "git"}
	results := make([]string, 0)

	for i, bin := range binaries {
		time.Sleep(500 * time.Millisecond)
		path, err := exec.LookPath(bin)
		if err == nil {
			results = append(results, fmt.Sprintf("✅ %s: %s", bin, path))
		} else {
			results = append(results, fmt.Sprintf("❌ %s: Not found", bin))
		}

		m.mu.Lock()
		task.Progress = float64(i+1) / float64(len(binaries))
		m.mu.Unlock()
		m.emitUpdate(task)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	finish := time.Now()
	task.FinishedAt = &finish
	task.Status = StatusCompleted
	task.Result = strings.Join(results, "\n")
	m.emitUpdate(task)
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

// CreateWidget creates a new rich UI widget
func (m *Manager) CreateWidget(widgetType, title string) *Widget {
	widget := &Widget{
		ID:        uuid.New().String(),
		Type:      widgetType,
		Title:     title,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.widgets[widget.ID] = widget
	m.mu.Unlock()

	if m.notify != nil {
		m.notify("agent.widgetCreated", widget)
	}

	return widget
}

// UpdateWidgetVDOM updates the VDOM of a widget
func (m *Manager) UpdateWidgetVDOM(id string, node *vdom.Node) error {
	m.mu.Lock()
	widget, ok := m.widgets[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("widget not found: %s", id)
	}
	widget.VDOM = node
	m.mu.Unlock()

	if m.notify != nil {
		m.notify("agent.widgetUpdated", widget)
	}

	return nil
}
