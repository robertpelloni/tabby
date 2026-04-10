// Package updater provides update checking and management for Tabby's Go backend.
//
// It can check GitHub releases for new versions, download updates,
// and notify the frontend about available updates.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset (downloadable file)
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateStatus represents the current update status
type UpdateStatus struct {
	CurrentVersion string   `json:"currentVersion"`
	LatestVersion  string   `json:"latestVersion"`
	UpdateAvailable bool    `json:"updateAvailable"`
	Release        *Release `json:"release,omitempty"`
	LastError      string   `json:"lastError,omitempty"`
	LastChecked    time.Time `json:"lastChecked"`
}

// Config configures the updater
type Config struct {
	RepoOwner    string
	RepoName     string
	CurrentVersion string
	CheckInterval time.Duration
}

// Manager manages update checking
type Manager struct {
	mu       sync.RWMutex
	config   Config
	status   UpdateStatus
	stop     chan struct{}
	onChange []func(UpdateStatus)
	client   *http.Client
}

// NewManager creates a new update manager
func NewManager(config Config) *Manager {
	if config.CheckInterval == 0 {
		config.CheckInterval = 24 * time.Hour
	}
	return &Manager{
		config: config,
		status: UpdateStatus{
			CurrentVersion: config.CurrentVersion,
		},
		stop:   make(chan struct{}),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// CheckForUpdates checks GitHub for the latest release
func (m *Manager) CheckForUpdates() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		m.config.RepoOwner, m.config.RepoName)

	resp, err := m.client.Get(url)
	if err != nil {
		m.setStatusError(err.Error())
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		m.setStatusError("no releases found")
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != 200 {
		m.setStatusError(fmt.Sprintf("GitHub API returned %d", resp.StatusCode))
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	m.mu.Lock()
	m.status.LatestVersion = release.TagName
	m.status.UpdateAvailable = release.TagName != m.config.CurrentVersion
	m.status.Release = &release
	m.status.LastChecked = time.Now()
	m.status.LastError = ""
	status := m.status
	m.mu.Unlock()

	m.notifyChange(status)
	return &release, nil
}

// GetStatus returns the current update status
func (m *Manager) GetStatus() UpdateStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// StartPeriodicChecks starts periodic update checking
func (m *Manager) StartPeriodicChecks() {
	go func() {
		// Check immediately on start
		m.CheckForUpdates()

		ticker := time.NewTicker(m.config.CheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CheckForUpdates()
			case <-m.stop:
				return
			}
		}
	}()
}

// Stop stops periodic update checking
func (m *Manager) Stop() {
	close(m.stop)
}

// OnChange registers a callback for status changes
func (m *Manager) OnChange(cb func(UpdateStatus)) {
	m.mu.Lock()
	m.onChange = append(m.onChange, cb)
	m.mu.Unlock()
}

// setStatusError sets an error in the status
func (m *Manager) setStatusError(msg string) {
	m.mu.Lock()
	m.status.LastError = msg
	m.status.LastChecked = time.Now()
	status := m.status
	m.mu.Unlock()
	m.notifyChange(status)
}

// notifyChange calls all registered callbacks
func (m *Manager) notifyChange(status UpdateStatus) {
	m.mu.RLock()
	callbacks := make([]func(UpdateStatus), len(m.onChange))
	copy(callbacks, m.onChange)
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb(status)
	}
}
