//go:build windows

package serial

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
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
	ID          string
	PortName    string
	BaudRate    int
	DataBits    int
	StopBits    int
	Parity      string
	FlowControl string
	SerialPort  serial.Port
	Open        bool
	LogFile     *os.File
	LogPath     string
	logMu       sync.Mutex
	done        chan struct{}
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

// BuildMode constructs the serial.Mode given API parameters
func BuildMode(params api.SerialOpenParams) *serial.Mode {
	mode := &serial.Mode{
		BaudRate: params.BaudRate,
	}
	if params.DataBits > 0 {
		mode.DataBits = params.DataBits
	} else {
		mode.DataBits = 8
	}
	switch params.StopBits {
	case 2:
		mode.StopBits = serial.TwoStopBits
	default:
		mode.StopBits = serial.OneStopBit
	}
	switch params.Parity {
	case "even":
		mode.Parity = serial.EvenParity
	case "odd":
		mode.Parity = serial.OddParity
	default:
		mode.Parity = serial.NoParity
	}

	// Initial RTS/DTR status bits based on flow control
	if params.FlowControl == "hardware" {
		mode.InitialStatusBits = &serial.ModemOutputBits{
			RTS: true,
			DTR: true,
		}
	}

	return mode
}

// Open opens a serial port with the specified parameters
func (m *Manager) Open(params api.SerialOpenParams) (*api.SerialOpenResult, error) {
	mode := BuildMode(params)

	port, err := serial.Open(params.Port, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", params.Port, err)
	}

	// Apply hardware handshake lines if specified
	if params.FlowControl == "hardware" {
		_ = port.SetRTS(true)
		_ = port.SetDTR(true)
	}

	id := params.ID
	if id == "" {
		id = m.nextID("serial")
	}

	var logFile *os.File
	var logPath string
	if params.Logging {
		logDir := params.LogDir
		if logDir == "" {
			if home, err := os.UserHomeDir(); err == nil {
				logDir = filepath.Join(home, ".tabby", "serial-logs")
			} else {
				logDir = filepath.Join(os.TempDir(), "tabby-serial-logs")
			}
		}
		_ = os.MkdirAll(logDir, 0700)
		timestamp := time.Now().Format("20060102-150405")
		safePort := filepath.Base(params.Port)
		logPath = filepath.Join(logDir, fmt.Sprintf("serial_%s_%s.log", safePort, timestamp))
		if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			logFile = f
			header := fmt.Sprintf("=== Serial Session Log Started: %s | Port: %s | Baud: %d | FlowControl: %s ===\n",
				time.Now().Format(time.RFC3339), params.Port, params.BaudRate, params.FlowControl)
			_, _ = logFile.WriteString(header)
		}
	}

	conn := &SerialConnection{
		ID:          id,
		PortName:    params.Port,
		BaudRate:    params.BaudRate,
		DataBits:    mode.DataBits,
		StopBits:    int(mode.StopBits),
		Parity:      params.Parity,
		FlowControl: params.FlowControl,
		SerialPort:  port,
		Open:        true,
		LogFile:     logFile,
		LogPath:     logPath,
		done:        make(chan struct{}),
	}

	m.mu.Lock()
	m.connections[id] = conn
	m.mu.Unlock()

	go m.forwardOutput(conn)

	return &api.SerialOpenResult{
		ID:      id,
		LogPath: logPath,
	}, nil
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

	if conn.LogFile != nil {
		conn.logMu.Lock()
		_, _ = conn.LogFile.WriteString(fmt.Sprintf("[%s OUT] %s\n", time.Now().Format("15:04:05.000"), string(decoded)))
		conn.logMu.Unlock()
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
		if conn.LogFile != nil {
			conn.logMu.Lock()
			_, _ = conn.LogFile.WriteString(fmt.Sprintf("=== Serial Session Log Closed: %s ===\n", time.Now().Format(time.RFC3339)))
			_ = conn.LogFile.Close()
			conn.LogFile = nil
			conn.logMu.Unlock()
		}
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
func (m *Manager) forwardOutput(conn *SerialConnection) {
	serialID := conn.ID
	reader := conn.SerialPort
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if conn.LogFile != nil {
				conn.logMu.Lock()
				_, _ = conn.LogFile.WriteString(fmt.Sprintf("[%s IN ] %s\n", time.Now().Format("15:04:05.000"), string(chunk)))
				conn.logMu.Unlock()
			}
			encoded := base64.StdEncoding.EncodeToString(chunk)
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
			if c, exists := m.connections[serialID]; exists {
				if c.LogFile != nil {
					c.logMu.Lock()
					_, _ = c.LogFile.WriteString(fmt.Sprintf("=== Serial Connection Ended: %v ===\n", err))
					_ = c.LogFile.Close()
					c.LogFile = nil
					c.logMu.Unlock()
				}
				delete(m.connections, serialID)
			}
			m.mu.Unlock()
			return
		}
	}
}
