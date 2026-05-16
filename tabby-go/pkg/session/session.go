// Package session provides persistent session state for Tabby Go.
// It saves/restores which tabs were open so they can be re-created on app restart.
package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/robertpelloni/tabby/tabby-go/pkg/settings"
)

// TabState represents a single tab that was open when the app was closed.
type TabState struct {
	Shell  string `json:"shell"`
	Title  string `json:"title"`
	Active bool   `json:"active"`
}

// SessionState represents the full application session state.
type SessionState struct {
	Tabs    []TabState `json:"tabs"`
	Version int        `json:"version"`
}

// sessionFile returns the path to the session state file.
func sessionFile() (string, error) {
	dir, err := settings.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// SaveSession persists the current session state to disk.
func SaveSession(tabs []TabState) error {
	dir, err := settings.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := sessionFile()
	if err != nil {
		return err
	}

	state := SessionState{
		Tabs:    tabs,
		Version: 1,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0o644)
}

// LoadSession reads the last saved session state from disk.
// Returns nil if no session file exists.
func LoadSession() (*SessionState, error) {
	file, err := sessionFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ClearSession removes the session file (e.g. on explicit close-all).
func ClearSession() error {
	file, err := sessionFile()
	if err != nil {
		return err
	}
	return os.Remove(file)
}

// GetPlatformInfo returns basic platform info for session metadata.
func GetPlatformInfo() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}
