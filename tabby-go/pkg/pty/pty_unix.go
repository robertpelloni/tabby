//go:build !windows

package pty

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Manager manages PTY processes on Unix using creack/pty
type Manager struct {
	ptyInstances map[string]*PTYInstance
	mu           sync.RWMutex
	notify       NotifyFunc
	idCounter    int
}

// NotifyFunc is the callback type for sending notifications
type NotifyFunc func(method string, params interface{})

// PTYInstance represents a running PTY process
type PTYInstance struct {
	ID      string
	PID     int
	Pty     *os.File
	Cmd     *exec.Cmd
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

// Spawn creates a new PTY process
func (m *Manager) Spawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	cmd := exec.Command(params.Command, params.Args...)
	if params.Cwd != "" {
		cmd.Dir = params.Cwd
	}
	if params.Env != nil {
		cmd.Env = os.Environ()
		for k, v := range params.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cols := params.Columns
	if cols == 0 {
		cols = 120
	}
	rows := params.Rows
	if rows == 0 {
		rows = 30
	}

	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("pty")
	}

	instance := &PTYInstance{
		ID:      id,
		PID:     cmd.Process.Pid,
		Pty:     ptyFile,
		Cmd:     cmd,
		Columns: cols,
		Rows:    rows,
		done:    make(chan struct{}),
	}

	m.mu.Lock()
	m.ptyInstances[id] = instance
	m.mu.Unlock()

	go m.forwardOutput(id, ptyFile)
	go m.monitorExit(id, cmd)

	return &api.PTYSpawnResult{
		ID:  id,
		PID: cmd.Process.Pid,
	}, nil
}

// Resize resizes the PTY
func (m *Manager) Resize(id string, columns, rows int) error {
	m.mu.RLock()
	instance, ok := m.ptyInstances[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	instance.Columns = columns
	instance.Rows = rows
	return pty.Setsize(instance.Pty, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(columns),
	})
}

// Write sends data to the PTY
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

	_, err = instance.Pty.Write(decoded)
	return err
}

// Kill terminates a PTY process
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
	instance.Pty.Close()
	return instance.Cmd.Process.Kill()
}

// forwardOutput reads PTY output and sends data notifications
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

// monitorExit waits for a process to exit
func (m *Manager) monitorExit(ptyID string, cmd *exec.Cmd) {
	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	m.notify("pty.exit", map[string]interface{}{
		"ptyId":    ptyID,
		"exitCode": exitCode,
	})

	m.mu.Lock()
	delete(m.ptyInstances, ptyID)
	m.mu.Unlock()
}
