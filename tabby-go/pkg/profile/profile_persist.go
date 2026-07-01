// Package profile provides connection profile persistence for Tabby Go.
// This file adds disk persistence (JSON load/save) and SSH config import
// on top of the types defined in profile.go.
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/settings"
)

// ConnectionProfile represents a saved connection profile that can be persisted to disk.
// This is a lightweight version of Profile that is JSON-serializable for disk storage.
type ConnectionProfile struct {
	ID                  string      `json:"id"`
	Type                ProfileType `json:"type"`
	Name                string      `json:"name"`
	Group               string      `json:"group,omitempty"`
	Icon                string      `json:"icon,omitempty"`
	Color               string      `json:"color,omitempty"`
	DisableDynamicTitle bool        `json:"disableDynamicTitle,omitempty"`
	Options             interface{} `json:"options"`
	CreatedAt           time.Time   `json:"createdAt"`
	UpdatedAt           time.Time   `json:"updatedAt"`
}

// profilesFile returns the path to the profiles file.
func profilesFile() (string, error) {
	dir, err := settings.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles.json"), nil
}

// LoadProfiles loads saved connection profiles from disk.
func LoadProfiles() ([]ConnectionProfile, error) {
	file, err := profilesFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return []ConnectionProfile{}, nil
		}
		return nil, err
	}
	var profiles []ConnectionProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

// SaveProfiles persists connection profiles to disk.
func SaveProfiles(profiles []ConnectionProfile) error {
	file, err := profilesFile()
	if err != nil {
		return err
	}
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}

// ImportSSHConfig imports profiles from an SSH config file (~/.ssh/config).
// It parses the OpenSSH config format and returns the imported profiles as ConnectionProfiles.
func ImportSSHConfigAsProfiles(path string) ([]ConnectionProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH config: %w", err)
	}

	var imported []ConnectionProfile
	var current *ConnectionProfile
	var currentOpts *SSHProfileOptions

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		keyword := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch keyword {
		case "host":
			if current != nil {
				imported = append(imported, *current)
			}
			currentOpts = &SSHProfileOptions{Host: value, Port: 22, Auth: "agent"}
			current = &ConnectionProfile{
				ID:        fmt.Sprintf("ssh-%d", time.Now().UnixNano()),
				Type:      TypeSSH,
				Name:      value,
				Options:   currentOpts,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		case "hostname":
			if currentOpts != nil {
				currentOpts.Host = value
			}
		case "port":
			if currentOpts != nil {
				fmt.Sscanf(value, "%d", &currentOpts.Port)
			}
		case "user":
			if currentOpts != nil {
				currentOpts.User = value
			}
		case "identityfile":
			if currentOpts != nil {
				if strings.HasPrefix(value, "~/") {
					home, _ := os.UserHomeDir()
					value = filepath.Join(home, value[2:])
				}
				currentOpts.PrivateKeys = append(currentOpts.PrivateKeys, value)
				currentOpts.Auth = "publicKey"
			}
		case "proxycommand":
			if currentOpts != nil {
				currentOpts.ProxyCommand = value
			}
		case "forwardagent":
			if currentOpts != nil {
				currentOpts.AgentForward = value == "yes"
			}
		case "serveraliveinterval":
			if currentOpts != nil {
				fmt.Sscanf(value, "%d", &currentOpts.KeepaliveInterval)
			}
		case "serveralivecountmax":
			if currentOpts != nil {
				fmt.Sscanf(value, "%d", &currentOpts.KeepaliveCountMax)
			}
		case "connecttimeout":
			if currentOpts != nil {
				fmt.Sscanf(value, "%d", &currentOpts.ReadyTimeout)
			}

		case "proxyjump":
			if currentOpts != nil {
				currentOpts.ProxyJump = value
			}
		}
	}
	if current != nil {
		imported = append(imported, *current)
	}
	return imported, nil
}
