package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles syncing workflows, profiles, and environments
type Manager struct {
	syncPath string
}

func NewManager() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		syncPath: filepath.Join(home, ".tabby", "cloud_sync.json"),
	}
}

type SyncData struct {
	Workflows []Workflow `json:"workflows"`
	Profiles  []Profile  `json:"profiles"`
	EnvVars   []EnvVar   `json:"envVars"`
}

type Workflow struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Tags    []string `json:"tags"`
}

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PushParams struct {
	Data SyncData `json:"data"`
}

type PushResult struct {
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
}

func (m *Manager) Push(params PushParams) (*PushResult, error) {
	// Mock implementation: saves to local file representing cloud storage
	data, err := json.MarshalIndent(params.Data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync data: %v", err)
	}

	dir := filepath.Dir(m.syncPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	if err := os.WriteFile(m.syncPath, data, 0600); err != nil {
		return nil, err
	}

	return &PushResult{
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

type PullResult struct {
	Data      SyncData `json:"data"`
	Timestamp string   `json:"timestamp"`
}

func (m *Manager) Pull() (*PullResult, error) {
	// Mock implementation: reads from local file representing cloud storage
	data, err := os.ReadFile(m.syncPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &PullResult{
				Data: SyncData{
					Workflows: []Workflow{},
					Profiles:  []Profile{},
					EnvVars:   []EnvVar{},
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		}
		return nil, fmt.Errorf("failed to read sync data: %v", err)
	}

	var syncData SyncData
	if err := json.Unmarshal(data, &syncData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync data: %v", err)
	}

	return &PullResult{
		Data:      syncData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
