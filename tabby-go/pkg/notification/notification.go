// Package notification provides a notification system for Tabby's Go backend.
//
// It supports desktop notifications, service messages, and event logging
// through a unified interface.
package notification

import (
	"fmt"
	"sync"
	"time"
)

// Level represents the severity of a notification
type Level int

const (
	LevelInfo Level = iota
	LevelWarning
	LevelError
)

// Notification represents a single notification
type Notification struct {
	ID        string    `json:"id"`
	Level     Level     `json:"level"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// Manager manages notifications
type Manager struct {
	mu            sync.RWMutex
	notifications []Notification
	onChange      []func([]Notification)
	idCounter     int
}

// NewManager creates a new notification manager
func NewManager() *Manager {
	return &Manager{
		notifications: make([]Notification, 0),
	}
}

// Info creates an info-level notification
func (m *Manager) Info(title, message string) {
	m.add(LevelInfo, title, message)
}

// Warning creates a warning-level notification
func (m *Manager) Warning(title, message string) {
	m.add(LevelWarning, title, message)
}

// Error creates an error-level notification
func (m *Manager) Error(title, message string) {
	m.add(LevelError, title, message)
}

// ServiceMessage logs a service message as an info notification
func (m *Manager) ServiceMessage(connectionID, message string) {
	m.Info(connectionID, message)
}

// GetUnread returns all unread notifications
func (m *Manager) GetUnread() []Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Notification
	for _, n := range m.notifications {
		if !n.Read {
			result = append(result, n)
		}
	}
	return result
}

// GetAll returns all notifications
func (m *Manager) GetAll() []Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Notification, len(m.notifications))
	copy(result, m.notifications)
	return result
}

// MarkRead marks a notification as read
func (m *Manager) MarkRead(id string) {
	m.mu.Lock()
	for i := range m.notifications {
		if m.notifications[i].ID == id {
			m.notifications[i].Read = true
			break
		}
	}
	m.mu.Unlock()
	m.notifyChange()
}

// Clear removes all notifications
func (m *Manager) Clear() {
	m.mu.Lock()
	m.notifications = make([]Notification, 0)
	m.mu.Unlock()
	m.notifyChange()
}

// OnChange registers a callback for notification changes
func (m *Manager) OnChange(cb func([]Notification)) {
	m.mu.Lock()
	m.onChange = append(m.onChange, cb)
	m.mu.Unlock()
}

// add adds a notification
func (m *Manager) add(level Level, title, message string) {
	m.mu.Lock()
	m.idCounter++
	n := Notification{
		ID:        fmt.Sprintf("notif-%d", m.idCounter),
		Level:     level,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}
	m.notifications = append(m.notifications, n)
	// Keep last 100 notifications
	if len(m.notifications) > 100 {
		m.notifications = m.notifications[len(m.notifications)-100:]
	}
	m.mu.Unlock()
	m.notifyChange()
}

// notifyChange calls all registered change callbacks
func (m *Manager) notifyChange() {
	m.mu.RLock()
	callbacks := make([]func([]Notification), len(m.onChange))
	copy(callbacks, m.onChange)
	allNotifs := make([]Notification, len(m.notifications))
	copy(allNotifs, m.notifications)
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb(allNotifs)
	}
}
