// Package configsync provides configuration synchronization for Tabby's Go backend.
//
// It watches configuration files for changes and can sync settings across
// multiple Tabby instances or between the Electron and native apps.
package configsync

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// SyncStatus represents the current sync state
type SyncStatus struct {
	LastSyncTime time.Time `json:"lastSyncTime"`
	ConfigHash   string    `json:"configHash"`
	Dirty        bool      `json:"dirty"`
	Error        string    `json:"error,omitempty"`
}

// Service provides configuration synchronization
type Service struct {
	mu       sync.RWMutex
	path     string
	status   SyncStatus
	onChange []func(SyncStatus)
	stop     chan struct{}
	interval time.Duration
}

// NewService creates a new config sync service
func NewService(path string) *Service {
	return &Service{
		path:     path,
		interval: 5 * time.Second,
		stop:     make(chan struct{}),
	}
}

// Start begins watching the config file for changes
func (s *Service) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CheckForChanges()
			case <-s.stop:
				return
			}
		}
	}()
}

// Stop stops watching
func (s *Service) Stop() {
	close(s.stop)
}

// CheckForChanges reads the config file and checks if it has changed
func (s *Service) CheckForChanges() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		s.setStatusError(err.Error())
		return
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	s.mu.Lock()
	if hash != s.status.ConfigHash {
		s.status.ConfigHash = hash
		s.status.Dirty = false
		s.status.LastSyncTime = time.Now()
		s.status.Error = ""
	}
	status := s.status
	s.mu.Unlock()

	s.notifyChange(status)
}

// MarkDirty marks the config as locally modified
func (s *Service) MarkDirty() {
	s.mu.Lock()
	s.status.Dirty = true
	status := s.status
	s.mu.Unlock()
	s.notifyChange(status)
}

// GetStatus returns the current sync status
func (s *Service) GetStatus() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// LoadConfig reads and unmarshals the config file
func (s *Service) LoadConfig(v interface{}) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	return json.Unmarshal(data, v)
}

// SaveConfig marshals and writes the config file atomically
func (s *Service) SaveConfig(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename config: %w", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	s.mu.Lock()
	s.status.ConfigHash = hash
	s.status.Dirty = false
	s.status.LastSyncTime = time.Now()
	status := s.status
	s.mu.Unlock()

	s.notifyChange(status)
	return nil
}

// OnChange registers a callback for status changes
func (s *Service) OnChange(cb func(SyncStatus)) {
	s.mu.Lock()
	s.onChange = append(s.onChange, cb)
	s.mu.Unlock()
}

// setStatusError sets an error in the status
func (s *Service) setStatusError(msg string) {
	s.mu.Lock()
	s.status.Error = msg
	status := s.status
	s.mu.Unlock()
	s.notifyChange(status)
}

// notifyChange calls all registered callbacks
func (s *Service) notifyChange(status SyncStatus) {
	s.mu.RLock()
	callbacks := make([]func(SyncStatus), len(s.onChange))
	copy(callbacks, s.onChange)
	s.mu.RUnlock()

	for _, cb := range callbacks {
		cb(status)
	}
}
