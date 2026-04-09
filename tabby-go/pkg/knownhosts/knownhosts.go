// Package knownhosts provides SSH known host key management for Tabby's Go backend.
//
// It manages a list of known host keys, supporting lookup, storage, and
// verification. Host keys can be stored in memory and optionally persisted
// to a file in the OpenSSH known_hosts format.
package knownhosts

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Selector identifies a host key entry
type Selector struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Type string `json:"type"` // e.g., "ssh-ed25519", "ssh-rsa"
}

// KnownHost represents a known host key entry
type KnownHost struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Digest   string `json:"digest"`   // SHA-256 fingerprint (base64)
	KeyBytes string `json:"keyBytes"` // Public key bytes (base64)
}

// Manager manages known SSH host keys
type Manager struct {
	mu      sync.RWMutex
	entries []KnownHost
}

// NewManager creates a new known hosts manager
func NewManager() *Manager {
	return &Manager{
		entries: make([]KnownHost, 0),
	}
}

// GetFor looks up a known host by selector
func (m *Manager) GetFor(selector Selector) *KnownHost {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.Host == selector.Host && entry.Port == selector.Port && entry.Type == selector.Type {
			return &entry
		}
	}
	return nil
}

// Store saves a known host key
func (m *Manager) Store(selector Selector, digest string, keyBytes []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update existing entry
	for i, entry := range m.entries {
		if entry.Host == selector.Host && entry.Port == selector.Port && entry.Type == selector.Type {
			m.entries[i].Digest = digest
			if keyBytes != nil {
				m.entries[i].KeyBytes = base64.StdEncoding.EncodeToString(keyBytes)
			}
			return
		}
	}

	// Add new entry
	entry := KnownHost{
		Host:   selector.Host,
		Port:   selector.Port,
		Type:   selector.Type,
		Digest: digest,
	}
	if keyBytes != nil {
		entry.KeyBytes = base64.StdEncoding.EncodeToString(keyBytes)
	}
	m.entries = append(m.entries, entry)
}

// Remove removes a known host entry
func (m *Manager) Remove(selector Selector) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.entries {
		if entry.Host == selector.Host && entry.Port == selector.Port && entry.Type == selector.Type {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			return
		}
	}
}

// List returns all known host entries
func (m *Manager) List() []KnownHost {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]KnownHost, len(m.entries))
	copy(result, m.entries)
	return result
}

// Verify checks if a host key matches a known entry.
// Returns true if the key matches, false if it doesn't match or is unknown.
func (m *Manager) Verify(selector Selector, keyBytes []byte) (bool, error) {
	digest := FingerprintSHA256(keyBytes)

	entry := m.GetFor(selector)
	if entry == nil {
		return false, nil // Unknown host
	}

	if entry.Digest == digest {
		return true, nil
	}

	return false, fmt.Errorf("host key mismatch for %s:%d (expected %s, got %s)",
		selector.Host, selector.Port, entry.Digest, digest)
}

// LoadFromFile loads known hosts from an OpenSSH known_hosts file
func (m *Manager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read known_hosts: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse: [host]:port type keybase64
		// or: host type keybase64
		// or: @marker [host]:port type keybase64
		parts := strings.Fields(line)
		offset := 0
		if len(parts) > 0 && strings.HasPrefix(parts[0], "@") {
			offset = 1
		}

		if len(parts) < offset+3 {
			continue
		}

		hostPart := parts[offset]
		keyType := parts[offset+1]
		keyB64 := parts[offset+2]

		host, port := parseHostPort(hostPart)

		keyBytes, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			continue
		}

		digest := FingerprintSHA256(keyBytes)

		m.entries = append(m.entries, KnownHost{
			Host:     host,
			Port:     port,
			Type:     keyType,
			Digest:   digest,
			KeyBytes: keyB64,
		})
	}

	return nil
}

// SaveToFile saves known hosts to an OpenSSH known_hosts file
func (m *Manager) SaveToFile(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("# Tabby known hosts\n")

	for _, entry := range m.entries {
		host := formatHostPort(entry.Host, entry.Port)
		sb.WriteString(fmt.Sprintf("%s %s %s\n", host, entry.Type, entry.KeyBytes))
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

// FingerprintSHA256 computes the SHA-256 fingerprint of a public key
func FingerprintSHA256(keyBytes []byte) string {
	h := sha256.Sum256(keyBytes)
	return "SHA256:" + base64.StdEncoding.EncodeToString(h[:])
}

// parseHostPort parses "[host]:port" or "host" from a known_hosts entry
func parseHostPort(s string) (string, int) {
	// Remove surrounding brackets for IPv6
	s = strings.TrimPrefix(s, "[")
	portIdx := strings.LastIndex(s, "]:")
	if portIdx >= 0 {
		host := s[:portIdx]
		port := 22
		fmt.Sscanf(s[portIdx+2:], "%d", &port)
		return host, port
	}
	// Check for non-standard port without brackets
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			port := 22
			fmt.Sscanf(parts[1], "%d", &port)
			return parts[0], port
		}
	}
	return strings.TrimSuffix(s, ":"), 22
}

// formatHostPort formats host:port for known_hosts
func formatHostPort(host string, port int) string {
	if port == 22 {
		return host
	}
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
