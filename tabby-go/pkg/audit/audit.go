// Package audit provides connection audit logging for Tabby Go.
//
// It logs connection events (connect, disconnect, auth, error)
// with timestamps and metadata for security and debugging purposes.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType represents the type of audit event.
type EventType string

const (
	EventConnect      EventType = "connect"
	EventDisconnect   EventType = "disconnect"
	EventAuthSuccess  EventType = "auth_success"
	EventAuthFailure  EventType = "auth_failure"
	EventError        EventType = "error"
	EventForwardStart EventType = "forward_start"
	EventForwardStop  EventType = "forward_stop"
	EventFileTransfer EventType = "file_transfer"
)

// Event represents a single audit log event.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Protocol  string    `json:"protocol"`  // ssh, telnet, serial, local
	Host      string    `json:"host,omitempty"`
	Port      int       `json:"port,omitempty"`
	User      string    `json:"user,omitempty"`
	Message   string    `json:"message,omitempty"`
	SessionID string    `json:"sessionId,omitempty"`
	Duration  string    `json:"duration,omitempty"`
}

// Logger handles audit logging to file.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	enabled  bool
	maxSize  int64 // max file size in bytes before rotation
}

// NewLogger creates a new audit logger.
func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("creating audit log directory: %w", err)
	}

	path := filepath.Join(logDir, "audit.log")
	l := &Logger{
		path:    path,
		enabled: true,
		maxSize: 10 * 1024 * 1024, // 10 MB
	}

	if err := l.open(); err != nil {
		return nil, err
	}

	return l, nil
}

// Log writes an audit event to the log file.
func (l *Logger) Log(event Event) error {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check for rotation
	if l.file != nil {
		if info, err := l.file.Stat(); err == nil && info.Size() > l.maxSize {
			l.rotate()
		}
	}

	event.Timestamp = time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	if l.file != nil {
		_, err = l.file.Write(append(data, '\n'))
		if err != nil {
			return fmt.Errorf("writing event: %w", err)
		}
	}

	return nil
}

// LogConnect logs a connection event.
func (l *Logger) LogConnect(protocol, host string, port int, user, sessionID string) error {
	return l.Log(Event{
		Type:      EventConnect,
		Protocol:  protocol,
		Host:      host,
		Port:      port,
		User:      user,
		SessionID: sessionID,
		Message:   fmt.Sprintf("Connected to %s:%d via %s", host, port, protocol),
	})
}

// LogDisconnect logs a disconnection event.
func (l *Logger) LogDisconnect(protocol, host string, port int, sessionID, duration string) error {
	return l.Log(Event{
		Type:      EventDisconnect,
		Protocol:  protocol,
		Host:      host,
		Port:      port,
		SessionID: sessionID,
		Duration:  duration,
		Message:   fmt.Sprintf("Disconnected from %s:%d (duration: %s)", host, port, duration),
	})
}

// LogAuthSuccess logs a successful authentication event.
func (l *Logger) LogAuthSuccess(protocol, host string, port int, user, method string) error {
	return l.Log(Event{
		Type:     EventAuthSuccess,
		Protocol: protocol,
		Host:     host,
		Port:     port,
		User:     user,
		Message:  fmt.Sprintf("Auth succeeded via %s", method),
	})
}

// LogAuthFailure logs a failed authentication event.
func (l *Logger) LogAuthFailure(protocol, host string, port int, user, reason string) error {
	return l.Log(Event{
		Type:     EventAuthFailure,
		Protocol: protocol,
		Host:     host,
		Port:     port,
		User:     user,
		Message:  fmt.Sprintf("Auth failed: %s", reason),
	})
}

// LogError logs an error event.
func (l *Logger) LogError(protocol, message string) error {
	return l.Log(Event{
		Type:     EventError,
		Protocol: protocol,
		Message:  message,
	})
}

// Close closes the audit log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetEnabled enables or disables audit logging.
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// IsEnabled returns whether audit logging is enabled.
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// GetPath returns the log file path.
func (l *Logger) GetPath() string {
	return l.path
}

func (l *Logger) open() error {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening audit log: %w", err)
	}
	l.file = f
	return nil
}

func (l *Logger) rotate() {
	if l.file != nil {
		l.file.Close()
	}
	// Rename current log to .old
	oldPath := l.path + ".old"
	os.Rename(l.path, oldPath)
	l.open()
}
