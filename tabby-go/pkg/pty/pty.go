// Package pty provides cross-platform PTY (pseudo-terminal) management
// for Tabby's Go backend.
//
// On Unix systems, it uses github.com/creack/pty.
// On Windows, it would use ConPTY (currently a stub).
package pty

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Manager manages PTY processes
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
	ID       string
	PID      int
	Cmd      *exec.Cmd
	Stdin    io.WriteCloser
	done     chan struct{}
}

// NewManager creates a new PTY manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		ptyInstances: make(map[string]*PTYInstance),
		notify:       notify,
	}
}

// nextID generates a unique identifier
func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// Spawn creates a new PTY process
// Note: Full PTY support requires the creack/pty package on Unix.
// This is a simplified implementation that uses exec.Cmd with pipes.
// For proper PTY support (pseudo-terminal), the creack/pty package should be used.
func (m *Manager) Spawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	// Build the command
	cmd := exec.Command(params.Command, params.Args...)

	// Set working directory
	if params.Cwd != "" {
		cmd.Dir = params.Cwd
	}

	// Set environment
	if params.Env != nil {
		// Start with current environment, then overlay
		cmd.Env = os.Environ()
		for k, v := range params.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Get stdin pipe for writing
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Get stdout pipe for reading
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Get stderr pipe for reading
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("pty")
	}
	instance := &PTYInstance{
		ID:    id,
		PID:   cmd.Process.Pid,
		Cmd:   cmd,
		Stdin: stdin,
		done:  make(chan struct{}),
	}

	m.mu.Lock()
	m.ptyInstances[id] = instance
	m.mu.Unlock()

	// Forward stdout
	go m.forwardOutput(id, stdout)
	// Forward stderr
	go m.forwardOutput(id, stderr)
	// Monitor exit
	go m.monitorExit(id, cmd)

	return &api.PTYSpawnResult{
		ID:  id,
		PID: cmd.Process.Pid,
	}, nil
}

// Resize resizes the PTY (requires actual PTY, not just exec.Cmd)
func (m *Manager) Resize(id string, columns, rows int) error {
	m.mu.RLock()
	_, ok := m.ptyInstances[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	// TODO: Implement actual PTY resize using creack/pty
	// For now, this is a no-op since we're using exec.Cmd with pipes
	return nil
}

// Write sends data to the PTY's stdin
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

	_, err = instance.Stdin.Write(decoded)
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

	instance.Stdin.Close()
	if signal == "" {
		return instance.Cmd.Process.Kill()
	}
	return instance.Cmd.Process.Kill()
}

// forwardOutput reads from a reader and sends data notifications
func (m *Manager) forwardOutput(ptyID string, reader io.Reader) {
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			m.notify("pty.data", api.DataNotification{
				PTYID: ptyID,
				Data:  encoded,
			})
		}
		if err != nil {
			break
		}
	}
}

// monitorExit waits for a process to exit and sends a notification
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

	m.notify("pty.exit", api.ExitNotification{
		PTYID:    ptyID,
		ExitCode: exitCode,
	})

	m.mu.Lock()
	delete(m.ptyInstances, ptyID)
	m.mu.Unlock()
}
