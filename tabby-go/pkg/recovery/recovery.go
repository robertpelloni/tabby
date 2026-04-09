// Package recovery provides tab/session recovery for Tabby's Go backend.
//
// When the application restarts (e.g., after an update or crash), the recovery
// system can restore previously open tabs and their connection state.
package recovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TabState represents the persistent state of a terminal tab
type TabState struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "ssh", "local", "serial", "telnet"
	Title      string                 `json:"title,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	ProfileID  string                 `json:"profileId,omitempty"`
	Options    map[string]interface{} `json:"options"`
	Recoverable bool                  `json:"recoverable"`
}

// SessionState represents the state of a connected session
type SessionState struct {
	TabID       string    `json:"tabId"`
	ConnectionID string   `json:"connectionId,omitempty"`
	Connected   bool      `json:"connected"`
	LastActive  time.Time `json:"lastActive"`
}

// RecoveryFile is the on-disk format for session recovery
type RecoveryFile struct {
	Version  int          `json:"version"`
	SavedAt  time.Time    `json:"savedAt"`
	Tabs     []TabState   `json:"tabs"`
	Sessions []SessionState `json:"sessions"`
}

// Manager manages session state persistence and recovery
type Manager struct {
	mu       sync.RWMutex
	tabs     map[string]*TabState
	sessions map[string]*SessionState
	filePath string
}

// NewManager creates a new recovery manager
func NewManager() *Manager {
	return &Manager{
		tabs:     make(map[string]*TabState),
		sessions: make(map[string]*SessionState),
	}
}

// RegisterTab registers a tab for recovery
func (m *Manager) RegisterTab(state TabState) {
	m.mu.Lock()
	state.CreatedAt = time.Now()
	state.Recoverable = true
	m.tabs[state.ID] = &state
	m.mu.Unlock()
}

// UnregisterTab removes a tab from recovery
func (m *Manager) UnregisterTab(tabID string) {
	m.mu.Lock()
	delete(m.tabs, tabID)
	delete(m.sessions, tabID)
	m.mu.Unlock()
}

// UpdateTab updates a tab's state
func (m *Manager) UpdateTab(tabID string, title string) {
	m.mu.Lock()
	if tab, ok := m.tabs[tabID]; ok {
		tab.Title = title
	}
	m.mu.Unlock()
}

// RegisterSession registers a session for recovery
func (m *Manager) RegisterSession(state SessionState) {
	m.mu.Lock()
	state.LastActive = time.Now()
	state.Connected = true
	m.sessions[state.TabID] = &state
	m.mu.Unlock()
}

// UnregisterSession removes a session from recovery
func (m *Manager) UnregisterSession(tabID string) {
	m.mu.Lock()
	if sess, ok := m.sessions[tabID]; ok {
		sess.Connected = false
	}
	m.mu.Unlock()
}

// GetRecoverableTabs returns all tabs that can be recovered
func (m *Manager) GetRecoverableTabs() []TabState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []TabState
	for _, tab := range m.tabs {
		if tab.Recoverable {
			result = append(result, *tab)
		}
	}
	return result
}

// Save persists the recovery state to disk
func (m *Manager) Save(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.filePath = path

	file := RecoveryFile{
		Version: 1,
		SavedAt: time.Now(),
	}

	for _, tab := range m.tabs {
		file.Tabs = append(file.Tabs, *tab)
	}
	for _, sess := range m.sessions {
		file.Sessions = append(file.Sessions, *sess)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal recovery state: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Load restores the recovery state from disk
func (m *Manager) Load(path string) (*RecoveryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read recovery file: %w", err)
	}

	var file RecoveryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse recovery file: %w", err)
	}

	m.mu.Lock()
	for i := range file.Tabs {
		tab := file.Tabs[i]
		m.tabs[tab.ID] = &tab
	}
	for i := range file.Sessions {
		sess := file.Sessions[i]
		sess.Connected = false // Sessions can't be truly recovered
		m.sessions[sess.TabID] = &sess
	}
	m.filePath = path
	m.mu.Unlock()

	return &file, nil
}

// Clear removes the recovery file and clears in-memory state
func (m *Manager) Clear() {
	m.mu.Lock()
	m.tabs = make(map[string]*TabState)
	m.sessions = make(map[string]*SessionState)
	path := m.filePath
	m.mu.Unlock()

	if path != "" {
		os.Remove(path)
	}
}

// GetRecoveryPath returns the platform-specific recovery file path
func GetRecoveryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".tabby", "recovery.json")
}
