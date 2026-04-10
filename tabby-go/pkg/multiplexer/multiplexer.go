// Package multiplexer provides SSH session sharing/multiplexing.
//
// It allows multiple UI tabs to share a single SSH connection.
// Each tab gets its own shell channel, but the underlying TCP
// connection and authentication are shared. This mirrors the
// TypeScript SSHMultiplexerService behavior.
package multiplexer

import (
	"fmt"
	"sync"
)

// MultiplexerKey uniquely identifies an SSH connection target
type MultiplexerKey struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	ProxyCommand   string `json:"proxyCommand,omitempty"`
	SocksProxyHost string `json:"socksProxyHost,omitempty"`
	SocksProxyPort int    `json:"socksProxyPort,omitempty"`
	HttpProxyHost  string `json:"httpProxyHost,omitempty"`
	HttpProxyPort  int    `json:"httpProxyPort,omitempty"`
	JumpHostKey    string `json:"jumpHostKey,omitempty"`
}

// Entry tracks a shared SSH connection reference
type Entry struct {
	Key           MultiplexerKey
	ConnectionID  string
	RefCount      int
}

// Manager manages shared SSH connections
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Entry // key string -> entry
}

// NewManager creates a new multiplexer manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Entry),
	}
}

// Register adds a new shared connection
func (m *Manager) Register(key MultiplexerKey, connectionID string) {
	keyStr := KeyString(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[keyStr] = &Entry{
		Key:          key,
		ConnectionID: connectionID,
		RefCount:     1,
	}
}

// Get returns an existing shared connection ID, or empty string if none
func (m *Manager) Get(key MultiplexerKey) string {
	keyStr := KeyString(key)

	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.sessions[keyStr]
	if !exists {
		return ""
	}

	entry.RefCount++
	return entry.ConnectionID
}

// Release decrements the reference count and returns true if fully released
func (m *Manager) Release(key MultiplexerKey) bool {
	keyStr := KeyString(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.sessions[keyStr]
	if !exists {
		return true
	}

	entry.RefCount--
	if entry.RefCount <= 0 {
		delete(m.sessions, keyStr)
		return true
	}
	return false
}

// Remove removes a shared connection entry regardless of ref count
func (m *Manager) Remove(connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for keyStr, entry := range m.sessions {
		if entry.ConnectionID == connectionID {
			delete(m.sessions, keyStr)
			return
		}
	}
}

// ListActive returns all active shared connections with their ref counts
func (m *Manager) ListActive() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int)
	for keyStr, entry := range m.sessions {
		result[keyStr] = entry.RefCount
	}
	return result
}

// KeyString creates a unique string key from a MultiplexerKey
func KeyString(key MultiplexerKey) string {
	result := fmt.Sprintf("%s:%d:%s:%s:%s:%d:%s:%d",
		key.Host, key.Port, key.User,
		key.ProxyCommand,
		key.SocksProxyHost, key.SocksProxyPort,
		key.HttpProxyHost, key.HttpProxyPort,
	)
	if key.JumpHostKey != "" {
		result += "$" + key.JumpHostKey
	}
	return result
}
