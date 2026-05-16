//go:build !windows

package serial

import (
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	serial "go.bug.st/serial"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Manager manages serial port connections using go.bug.st/serial
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
	Port     serial.Port
	Open     bool
	done     chan struct{}
}

// NewManager creates a new serial port manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		connections: make(map[string]*SerialConnection),
		notify:      notify,
	}
}

func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// Open opens a serial port with the specified parameters
func (m *Manager) Open(params api.SerialOpenParams) (*api.SerialOpenResult, error) {
	mode := &serial.Mode{
		BaudRate: params.BaudRate,
	}
	if params.DataBits > 0 {
		mode.DataBits = params.DataBits
	} else {
		mode.DataBits = 8
	}
	if params.StopBits > 0 {
		mode.StopBits = serial.StopBits(params.StopBits)
	} else {
		mode.StopBits = serial.StopBits(1)
	}
	switch params.Parity {
	case "even":
		mode.Parity = serial.EvenParity
	case "odd":
		mode.Parity = serial.OddParity
	default:
		mode.Parity = serial.NoParity
	}

	port, err := serial.Open(params.Port, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", params.Port, err)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("serial")
	}

	conn := &SerialConnection{
		ID:       id,
		PortName: params.Port,
		BaudRate: params.BaudRate,
		SerialPort: port,
		Open:     true,
		done:     make(chan struct{}),
	}

	m.mu.Lock()
	m.connections[id] = conn
	m.mu.Unlock()

	// Start reading data from the serial port
	go m.forwardOutput(id, port)

	return &api.SerialOpenResult{ID: id}, nil
}

// Write writes data to a serial port
func (m *Manager) Write(id string, data string) error {
	m.mu.RLock()
	conn, ok := m.connections[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("serial connection not found: %s", id)
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("invalid base64 data: %w", err)
	}

	_, err = conn.SerialPort.Write(decoded)
	return err
}

// Close closes a serial port
func (m *Manager) Close(id string) error {
	m.mu.Lock()
	conn, ok := m.connections[id]
	if ok {
		conn.Open = false
		close(conn.done)
		delete(m.connections, id)
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("serial connection not found: %s", id)
	}
	return conn.SerialPort.Close()
}

// ListPorts returns available serial ports
func (m *Manager) ListPorts() ([]api.SerialPortInfo, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	var result []api.SerialPortInfo
	for _, p := range ports {
		result = append(result, api.SerialPortInfo{Name: p})
	}
	return result, nil
}

// forwardOutput reads serial port output and sends data notifications
func (m *Manager) forwardOutput(serialID string, reader io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			m.notify("serial.data", map[string]interface{}{
				"serialId": serialID,
				"data":     encoded,
			})
		}
		if err != nil {
			m.notify("serial.exit", map[string]interface{}{
				"serialId": serialID,
				"error":    err.Error(),
			})
			m.mu.Lock()
			delete(m.connections, serialID)
			m.mu.Unlock()
			return
		}
	}
}
