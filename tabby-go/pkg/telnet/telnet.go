// Package telnet implements a Telnet client for Tabby's Go backend.
//
// Features:
// - Full Telnet protocol (RFC 854) with IAC command handling
// - Telnet option negotiation (WILL/WONT/DO/DONT)
// - Terminal type suboption (XTERM-256COLOR)
// - NAWS (Negotiate About Window Size) for terminal resize
// - Echo mode negotiation
// - Suppress Go Ahead support
// - Raw mode (no protocol processing) for non-Telnet connections
package telnet

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// Telnet commands (RFC 854)
const (
	IAC  byte = 255 // Interpret As Command
	DONT byte = 254
	DO   byte = 253
	WONT byte = 252
	WILL byte = 251
	SB   byte = 250 // Suboption Begin
	GA   byte = 249 // Go Ahead
	SE   byte = 240 // Suboption End

	// Suboption
	SUBOPT_SEND byte = 1
)

// Telnet options
const (
	OPT_ECHO      byte = 1
	OPT_SGA       byte = 3  // Suppress Go Ahead
	OPT_STATUS    byte = 5
	OPT_TTYPE     byte = 24 // Terminal Type
	OPT_NAWS      byte = 31 // Negotiate About Window Size
	OPT_TSPEED    byte = 32 // Terminal Speed
	OPT_RFC       byte = 33 // Remote Flow Control
	OPT_XDISPLOC  byte = 35 // X Display Location
	OPT_ENVIRON   byte = 36 // Environment
	OPT_NEWENV    byte = 39 // New Environment
	OPT_AUTH      byte = 37
)

// NotifyFunc sends notifications to the client
type NotifyFunc func(method string, params interface{})

// Manager manages Telnet connections
type Manager struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	notify      NotifyFunc
	idCounter   int
}

// Connection represents an active Telnet connection
type Connection struct {
	ID         string
	Host       string
	Port       int
	Conn       net.Conn
	Open       bool
	telnetMode bool
	lastWidth  int
	lastHeight int
	mu         sync.Mutex
}

// NewManager creates a new Telnet connection manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		notify:      notify,
	}
}

// TelnetConnectParams contains parameters for establishing a Telnet connection
type TelnetConnectParams struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TelnetConnectResult is returned after a successful connection
type TelnetConnectResult struct {
	ConnectionID string `json:"connectionId"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
}

// Connect establishes a new Telnet connection
func (m *Manager) Connect(params TelnetConnectParams) (*TelnetConnectResult, error) {
	port := params.Port
	if port == 0 {
		port = 23
	}
	addr := net.JoinHostPort(params.Host, fmt.Sprintf("%d", port))

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	m.mu.Lock()
	m.idCounter++
	connID := fmt.Sprintf("telnet-%d-%d", time.Now().UnixMilli(), m.idCounter)
	m.mu.Unlock()

	tc := &Connection{
		ID:    connID,
		Host:  params.Host,
		Port:  params.Port,
		Conn:  conn,
		Open:  true,
	}

	m.mu.Lock()
	m.connections[connID] = tc
	m.mu.Unlock()

	// Start reading
	go m.readLoop(tc)

	return &TelnetConnectResult{
		ConnectionID: connID,
		Host:         params.Host,
		Port:         params.Port,
	}, nil
}

// Write sends data to a Telnet connection
func (m *Manager) Write(connectionID string, data string) error {
	m.mu.RLock()
	tc, ok := m.connections[connectionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("invalid base64 data: %w", err)
	}

	_, err = tc.Conn.Write(decoded)
	return err
}

// Resize handles terminal resize via NAWS
func (m *Manager) Resize(connectionID string, width, height int) error {
	m.mu.RLock()
	tc, ok := m.connections[connectionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	tc.mu.Lock()
	tc.lastWidth = width
	tc.lastHeight = height
	tc.mu.Unlock()

	if tc.telnetMode {
		return m.sendNAWS(tc, width, height)
	}
	return nil
}

// Close closes a Telnet connection
func (m *Manager) Close(connectionID string) error {
	m.mu.Lock()
	tc, ok := m.connections[connectionID]
	if ok {
		delete(m.connections, connectionID)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	tc.Open = false
	return tc.Conn.Close()
}

// ListConnections returns all active connection IDs
func (m *Manager) ListConnections() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.connections))
	for id := range m.connections {
		ids = append(ids, id)
	}
	return ids
}

// readLoop reads data from the connection and processes Telnet protocol
func (m *Manager) readLoop(tc *Connection) {
	buf := make([]byte, 32*1024)
	for {
		n, err := tc.Conn.Read(buf)
		if n > 0 {
			data := buf[:n]
			if tc.telnetMode {
				data = m.processTelnet(tc, data)
			} else if data[0] == IAC && len(data) >= 3 {
				// Switch to Telnet protocol mode
				tc.telnetMode = true
				m.initTelnetOptions(tc)
				data = m.processTelnet(tc, data)
			}

			if len(data) > 0 {
				m.notify("telnet.data", api.DataNotification{
					ConnectionID: tc.ID,
					Data:         base64.StdEncoding.EncodeToString(data),
				})
			}
		}
		if err != nil {
			if err != io.EOF {
				m.notify("telnet.serviceMessage", api.ServiceMessageNotification{
					ConnectionID: tc.ID,
					Message:      fmt.Sprintf("Connection error: %v", err),
				})
			}
			tc.Open = false
			m.notify("telnet.exit", api.ExitNotification{
				ConnectionID: tc.ID,
				ExitCode:     0,
			})
			// Clean up
			m.mu.Lock()
			delete(m.connections, tc.ID)
			m.mu.Unlock()
			return
		}
	}
}

// initTelnetOptions sends initial Telnet option requests
func (m *Manager) initTelnetOptions(tc *Connection) {
	m.sendCommand(tc, DO, OPT_SGA)
	m.sendCommand(tc, WILL, OPT_TTYPE)
	m.sendCommand(tc, WILL, OPT_NAWS)
}

// processTelnet handles Telnet protocol bytes and returns the data payload
func (m *Manager) processTelnet(tc *Connection, data []byte) []byte {
	var cleanData []byte
	i := 0
	for i < len(data) {
		if data[i] == IAC {
			if i+1 >= len(data) {
				break
			}
			cmd := data[i+1]

			// Escaped IAC (0xFF 0xFF = literal 0xFF)
			if cmd == IAC {
				cleanData = append(cleanData, 0xFF)
				i += 2
				continue
			}

			switch cmd {
			case WILL, WONT, DO, DONT:
				if i+2 >= len(data) {
					i = len(data)
					continue
				}
				opt := data[i+2]
				m.handleOption(tc, cmd, opt)
				i += 3

			case SB: // Suboption
				end := findSuboptionEnd(data, i)
				if end == -1 {
					// Incomplete suboption, wait for more data
					i = len(data)
					continue
				}
				m.handleSuboption(tc, data[i+3:end-2], data[i+2])
				i = end

			default:
				// Other commands (GA, etc.) — skip
				i += 2
			}
		} else {
			cleanData = append(cleanData, data[i])
			i++
		}
	}
	return cleanData
}

// handleOption handles Telnet option negotiation
func (m *Manager) handleOption(tc *Connection, cmd, opt byte) {
	switch cmd {
	case WILL:
		switch opt {
		case OPT_ECHO, OPT_SGA:
			m.sendCommand(tc, DO, opt)
		default:
			m.sendCommand(tc, DONT, opt)
		}

	case WONT:
		m.sendCommand(tc, DONT, opt)

	case DO:
		switch opt {
		case OPT_NAWS:
			m.sendCommand(tc, WILL, opt)
			tc.mu.Lock()
			w, h := tc.lastWidth, tc.lastHeight
			tc.mu.Unlock()
			if w > 0 && h > 0 {
				m.sendNAWS(tc, w, h)
			}
		case OPT_TTYPE:
			m.sendCommand(tc, WILL, opt)
		case OPT_ECHO:
			m.sendCommand(tc, WILL, opt)
		default:
			m.sendCommand(tc, WONT, opt)
		}

	case DONT:
		m.sendCommand(tc, WONT, opt)
	}
}

// handleSuboption handles Telnet suboption data
func (m *Manager) handleSuboption(tc *Connection, data []byte, opt byte) {
	if opt == OPT_TTYPE && len(data) > 0 && data[0] == SUBOPT_SEND {
		// Respond with terminal type
		ttype := []byte("XTERM-256COLOR")
		msg := []byte{IAC, SB, OPT_TTYPE, 0}
		msg = append(msg, ttype...)
		msg = append(msg, IAC, SE)
		if tc.Conn != nil {
			tc.Conn.Write(msg)
		}
	}
}

// sendCommand sends a Telnet command
func (m *Manager) sendCommand(tc *Connection, cmd, opt byte) {
	if tc.Conn != nil {
		tc.Conn.Write([]byte{IAC, cmd, opt})
	}
}

// sendNAWS sends a window size suboption
func (m *Manager) sendNAWS(tc *Connection, width, height int) error {
	if tc.Conn == nil {
		return nil
	}
	msg := []byte{
		IAC, SB, OPT_NAWS,
		byte(width >> 8), byte(width & 0xFF),
		byte(height >> 8), byte(height & 0xFF),
		IAC, SE,
	}
	_, err := tc.Conn.Write(msg)
	return err
}

// findSuboptionEnd finds the end of a Telnet suboption (IAC SE)
func findSuboptionEnd(data []byte, start int) int {
	for i := start + 3; i < len(data)-1; i++ {
		if data[i] == IAC && data[i+1] == SE {
			return i + 2
		}
	}
	return -1
}
