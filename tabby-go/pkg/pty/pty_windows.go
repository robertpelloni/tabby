//go:build windows

package pty

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/UserExistsError/conpty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Manager manages PTY processes on Windows using ConPTY
type Manager struct {
	ptyInstances map[string]*PTYInstance
	mu           sync.RWMutex
	notify       NotifyFunc
	idCounter    int
}

// NotifyFunc is the callback type for sending notifications
type NotifyFunc func(method string, params interface{})

// PTYInstance represents a running ConPTY process
type PTYInstance struct {
	ID      string
	PID     int
	Cpty    *conpty.ConPty
	Columns int
	Rows    int
	done    chan struct{}
}

// NewManager creates a new PTY manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		ptyInstances: make(map[string]*PTYInstance),
		notify:       notify,
	}
}

func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// Spawn creates a new ConPTY process
func (m *Manager) Spawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	cmdLine := params.Command
	for _, arg := range params.Args {
		cmdLine += " " + arg
	}

	cpty, err := conpty.Start(cmdLine)
	if err != nil {
		return nil, fmt.Errorf("failed to start ConPTY: %w", err)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("pty")
	}

	cols := params.Columns
	if cols == 0 {
		cols = 120
	}
	rows := params.Rows
	if rows == 0 {
		rows = 30
	}

	instance := &PTYInstance{
		ID:      id,
		Cpty:    cpty,
		Columns: cols,
		Rows:    rows,
		done:    make(chan struct{}),
	}

	m.mu.Lock()
	m.ptyInstances[id] = instance
	m.mu.Unlock()

	// Resize to requested dimensions
	if err := cpty.Resize(cols, rows); err != nil {
		// Non-fatal: the PTY still works, just at default size
		fmt.Printf("Warning: ConPTY resize failed: %v\n", err)
	}

	// Forward output
	go m.forwardOutput(id, cpty)
	// Monitor exit
	go m.monitorExit(id, cpty)

	return &api.PTYSpawnResult{
		ID:  id,
		PID: 0,
	}, nil
}

// Resize resizes the ConPTY
func (m *Manager) Resize(id string, columns, rows int) error {
	m.mu.RLock()
	instance, ok := m.ptyInstances[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	instance.Columns = columns
	instance.Rows = rows
	return instance.Cpty.Resize(columns, rows)
}

// Write sends data to the ConPTY
func (m *Manager) Write(id string, data string) error {
	m.mu.RLock()
	instance, ok := m.ptyInstances[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("invalid base64 data: %w", err)
	}

	_, err = instance.Cpty.Write(decoded)
	return err
}

// Kill terminates a ConPTY process
func (m *Manager) Kill(id string, signal string) error {
	m.mu.Lock()
	instance, ok := m.ptyInstances[id]
	if ok {
		delete(m.ptyInstances, id)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	close(instance.done)
	return instance.Cpty.Close()
}

// forwardOutput reads ConPTY output and sends data notifications
func (m *Manager) forwardOutput(ptyID string, reader io.Reader) {
	buf := make([]byte, 64*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			m.notify("pty.data", map[string]interface{}{
				"ptyId": ptyID,
				"data":  encoded,
			})
		}
		if err != nil {
			break
		}
	}
}

// monitorExit waits for ConPTY to exit
func (m *Manager) monitorExit(ptyID string, cpty *conpty.ConPty) {
	cpty.Wait(context.Background())
	m.notify("pty.exit", map[string]interface{}{
		"ptyId":    ptyID,
		"exitCode": 0,
	})

	m.mu.Lock()
	delete(m.ptyInstances, ptyID)
	m.mu.Unlock()
}
