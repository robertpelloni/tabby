// Package profile provides connection profile management for Tabby's Go backend.
//
// Profiles define connection settings for SSH, local terminals, serial ports,
// and Telnet connections. They support groups, icons, colors, and can be
// imported from SSH config files.
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// ProfileType defines the type of a connection profile
type ProfileType string

const (
	TypeSSH    ProfileType = "ssh"
	TypeLocal  ProfileType = "local"
	TypeSerial ProfileType = "serial"
	TypeTelnet ProfileType = "telnet"
)

// Profile represents a connection profile
type Profile struct {
	ID         string      `json:"id"`
	Type       ProfileType `json:"type"`
	Name       string      `json:"name"`
	Group      string      `json:"group,omitempty"`
	Icon       string      `json:"icon,omitempty"`
	Color      string      `json:"color,omitempty"`
	DisableDynamicTitle bool `json:"disableDynamicTitle,omitempty"`
	Options    interface{} `json:"options"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

// SSHProfileOptions contains SSH-specific profile options
type SSHProfileOptions struct {
	Host              string            `json:"host"`
	Port              int               `json:"port"`
	User              string            `json:"user"`
	Auth              string            `json:"auth"` // "password", "publicKey", "agent", "keyboardInteractive"
	Password          string            `json:"password,omitempty"`
	PrivateKeys       []string          `json:"privateKeys,omitempty"`
	AgentForward      bool              `json:"agentForward,omitempty"`
	X11               bool              `json:"x11,omitempty"`
	KeepaliveInterval int               `json:"keepaliveInterval,omitempty"`
	KeepaliveCountMax int               `json:"keepaliveCountMax,omitempty"`
	ReadyTimeout      int               `json:"readyTimeout,omitempty"`
	JumpHost          string            `json:"jumpHost,omitempty"`
	ProxyCommand      string            `json:"proxyCommand,omitempty"`
	Algorithms        *api.SSHAlgorithms `json:"algorithms,omitempty"`
	ForwardedPorts    []ForwardedPort   `json:"forwardedPorts,omitempty"`
	Environment       map[string]string `json:"environment,omitempty"`
}

// ForwardedPort represents a port forwarding configuration
type ForwardedPort struct {
	Type          api.PortForwardType `json:"type"`
	Host          string              `json:"host"`
	Port          int                 `json:"port"`
	TargetAddress string              `json:"targetAddress"`
	TargetPort    int                 `json:"targetPort"`
}

// LocalProfileOptions contains local terminal profile options
type LocalProfileOptions struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Shell   string            `json:"shell,omitempty"`
}

// SerialProfileOptions contains serial port profile options
type SerialProfileOptions struct {
	Port        string `json:"port"`
	BaudRate    int    `json:"baudRate"`
	DataBits    int    `json:"dataBits,omitempty"`
	StopBits    int    `json:"stopBits,omitempty"`
	Parity      string `json:"parity,omitempty"`
	FlowControl string `json:"flowControl,omitempty"`
}

// TelnetProfileOptions contains Telnet profile options
type TelnetProfileOptions struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Manager manages connection profiles
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
}

// NewManager creates a new profile manager
func NewManager() *Manager {
	return &Manager{
		profiles: make(map[string]*Profile),
	}
}

// Add adds a profile
func (m *Manager) Add(profile *Profile) error {
	if profile.ID == "" {
		return fmt.Errorf("profile must have an ID")
	}

	profile.UpdatedAt = time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = profile.UpdatedAt
	}

	m.mu.Lock()
	m.profiles[profile.ID] = profile
	m.mu.Unlock()

	return nil
}

// Get returns a profile by ID
func (m *Manager) Get(id string) (*Profile, bool) {
	m.mu.RLock()
	p, ok := m.profiles[id]
	m.mu.RUnlock()
	return p, ok
}

// Remove removes a profile by ID
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	delete(m.profiles, id)
	m.mu.Unlock()
}

// List returns all profiles
func (m *Manager) List() []*Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Profile, 0, len(m.profiles))
	for _, p := range m.profiles {
		result = append(result, p)
	}
	return result
}

// ListByType returns profiles of a specific type
func (m *Manager) ListByType(profileType ProfileType) []*Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Profile
	for _, p := range m.profiles {
		if p.Type == profileType {
			result = append(result, p)
		}
	}
	return result
}

// ListByGroup returns profiles in a specific group
func (m *Manager) ListByGroup(group string) []*Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Profile
	for _, p := range m.profiles {
		if p.Group == group {
			result = append(result, p)
		}
	}
	return result
}

// Update updates a profile
func (m *Manager) Update(id string, profile *Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.profiles[id]; !ok {
		return fmt.Errorf("profile not found: %s", id)
	}

	profile.ID = id
	profile.UpdatedAt = time.Now()
	m.profiles[id] = profile
	return nil
}

// ImportSSHConfig imports profiles from an SSH config file
func (m *Manager) ImportSSHConfig(path string) ([]*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH config: %w", err)
	}

	var imported []*Profile
	var current *Profile

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
				m.Add(current)
				imported = append(imported, current)
			}
			current = &Profile{
				ID:   fmt.Sprintf("ssh-import-%s", value),
				Type: TypeSSH,
				Name: value,
				Options: &SSHProfileOptions{
					Host: value,
					Port: 22,
					Auth: "agent",
				},
			}

		case "hostname":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				opts.Host = value
			}

		case "port":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				fmt.Sscanf(value, "%d", &opts.Port)
			}

		case "user":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				opts.User = value
			}

		case "identityfile":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~/") {
					home, _ := os.UserHomeDir()
					value = filepath.Join(home, value[2:])
				}
				opts.PrivateKeys = append(opts.PrivateKeys, value)
				opts.Auth = "publicKey"
			}

		case "proxycommand":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				opts.ProxyCommand = value
			}

		case "forwardagent":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				opts.AgentForward = value == "yes"
			}

		case "serveraliveinterval":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				fmt.Sscanf(value, "%d", &opts.KeepaliveInterval)
			}

		case "serveralivecountmax":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				fmt.Sscanf(value, "%d", &opts.KeepaliveCountMax)
			}

		case "connecttimeout":
			if current != nil {
				opts := current.Options.(*SSHProfileOptions)
				fmt.Sscanf(value, "%d", &opts.ReadyTimeout)
			}
		}
	}

	if current != nil {
		m.Add(current)
		imported = append(imported, current)
	}

	return imported, nil
}

// GetDefaultSSHConfigPath returns the default SSH config file path
func GetDefaultSSHConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}
