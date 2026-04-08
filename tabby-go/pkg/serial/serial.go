// Package serial provides serial port communication for Tabby's Go backend.
//
// This is a stub implementation. Full serial port support would require
// the go.bug.st/serial package or similar platform-specific implementation.
package serial

import (
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Manager manages serial port connections
type Manager struct {
	connections map[string]*SerialConnection
	mu          sync.RWMutex
	notify      NotifyFunc
	idCounter   int
}

// NotifyFunc is the callback type for sending notifications
type NotifyFunc func(method string, params interface{})

// SerialConnection represents an open serial port
type SerialConnection struct {
	ID       string
	Port     string
	BaudRate int
	Open     bool
}

// NewManager creates a new serial port manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		connections: make(map[string]*SerialConnection),
		notify:      notify,
	}
}

// nextID generates a unique identifier
func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// Open opens a serial port
// TODO: Implement with go.bug.st/serial for actual serial communication
func (m *Manager) Open(params api.SerialOpenParams) (*api.SerialOpenResult, error) {
	// Stub: In a real implementation, this would:
	// 1. Open the serial port using go.bug.st/serial
	// 2. Configure baud rate, data bits, stop bits, parity, flow control
	// 3. Start a goroutine to read data and send notifications
	
	id := m.nextID("serial")
	conn := &SerialConnection{
		ID:       id,
		Port:     params.Port,
		BaudRate: params.BaudRate,
		Open:     true,
	}

	m.mu.Lock()
	m.connections[id] = conn
	m.mu.Unlock()

	return &api.SerialOpenResult{ID: id}, fmt.Errorf("serial port support not yet fully implemented (port: %s, baud: %d)", params.Port, params.BaudRate)
}

// Write writes data to a serial port
func (m *Manager) Write(id string, data string) error {
	m.mu.RLock()
	_, ok := m.connections[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("serial connection not found: %s", id)
	}

	// TODO: Implement actual write
	return fmt.Errorf("serial port support not yet fully implemented")
}

// Close closes a serial port
func (m *Manager) Close(id string) error {
	m.mu.Lock()
	conn, ok := m.connections[id]
	if ok {
		conn.Open = false
		delete(m.connections, id)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("serial connection not found: %s", id)
	}

	return nil
}

// ListPorts returns available serial ports
// TODO: Implement with go.bug.st/serial
func (m *Manager) ListPorts() ([]string, error) {
	// Stub: would use serial.GetPortsList()
	return []string{}, nil
}
