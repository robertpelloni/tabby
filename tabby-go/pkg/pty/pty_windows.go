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

type Manager struct {
	ptyInstances map[string]*PTYInstance
	mu           sync.RWMutex
	notify       NotifyFunc
	idCounter    int
}

type NotifyFunc func(method string, params interface{})

type PTYInstance struct {
	ID    string
	PID   int
	Pty   *conpty.ConPty
	done  chan struct{}
}

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

func (m *Manager) Spawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	cpty, err := conpty.Start(params.Command + " " + joinArgs(params.Args))
	if err != nil {
		return nil, fmt.Errorf("failed to start ConPTY: %w", err)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("pty")
	}

	instance := &PTYInstance{
		ID:    id,
		PID:   0, // ConPTY doesn't expose the underlying PID easily this way
		Pty:   cpty,
		done:  make(chan struct{}),
	}

	m.mu.Lock()
	m.ptyInstances[id] = instance
	m.mu.Unlock()

	go m.forwardOutput(id, cpty)
	go m.monitorExit(id, cpty)

	return &api.PTYSpawnResult{
		ID:  id,
		PID: 0,
	}, nil
}

func (m *Manager) Resize(id string, columns, rows int) error {
	m.mu.RLock()
	instance, ok := m.ptyInstances[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("PTY not found: %s", id)
	}

	return instance.Pty.Resize(columns, rows)
}

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

	return instance.Pty.Close()
}

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

func (m *Manager) monitorExit(ptyID string, cpty *conpty.ConPty) {
	cpty.Wait(context.Background())
	m.notify("pty.exit", api.ExitNotification{
		PTYID:    ptyID,
		ExitCode: 0,
	})

	m.mu.Lock()
	delete(m.ptyInstances, ptyID)
	m.mu.Unlock()
}

func joinArgs(args []string) string {
	res := ""
	for _, arg := range args {
		res += " " + arg
	}
	return res
}
