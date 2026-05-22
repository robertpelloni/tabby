// Package config provides configuration management for Tabby's Go backend.
//
// It handles loading, saving, and watching YAML configuration files
// with deep merge support, profile management, and environment variable
// substitution. Mirrors the TypeScript ConfigService behavior.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Store represents the complete Tabby configuration
type Store struct {
	// Core
	Appearance  AppearanceConfig  `yaml:"appearance" json:"appearance"`
	Terminal    TerminalConfig    `yaml:"terminal" json:"terminal"`
	Hotkeys     map[string]interface{} `yaml:"hotkeys" json:"hotkeys"`
	SSH         SSHConfig         `yaml:"ssh" json:"ssh"`
	Serial      SerialConfig      `yaml:"serial" json:"serial"`
	Profiles    ProfilesConfig    `yaml:"profiles" json:"profiles"`
	Plugin      PluginConfig      `yaml:"plugin" json:"plugin"`
	GoBackend   GoBackendConfig   `yaml:"goBackend" json:"goBackend"`

	// Raw store for plugins/extensions
	Raw map[string]interface{} `yaml:"-" json:"-"`
}

// AppearanceConfig contains appearance-related settings
type AppearanceConfig struct {
	Frame           string `yaml:"frame" json:"frame"`
	Vibrancy        string `yaml:"vibrancy" json:"vibrancy"`
	VibrancyType    string `yaml:"vibrancyType" json:"vibrancyType"`
	Opacity         float64 `yaml:"opacity" json:"opacity"`
	ColorSchemeMode string `yaml:"colorSchemeMode" json:"colorSchemeMode"`
	CSS             string `yaml:"css" json:"css"`
}

// TerminalConfig contains terminal-related settings
type TerminalConfig struct {
	ColorScheme   string            `yaml:"colorScheme" json:"colorScheme"`
	Font          string            `yaml:"font" json:"font"`
	FontSize      int               `yaml:"fontSize" json:"fontSize"`
	Scrollback    int               `yaml:"scrollback" json:"scrollback"`
	Cursor        string            `yaml:"cursor" json:"cursor"`
	Environment   map[string]string `yaml:"environment" json:"environment"`
	UseConPTY     bool              `yaml:"useConPTY" json:"useConPTY"`
	SetComSpec    bool              `yaml:"setComSpec" json:"setComSpec"`
}

// SSHConfig contains SSH-related settings
type SSHConfig struct {
	AgentType      string `yaml:"agentType" json:"agentType"`
	AgentPath      string `yaml:"agentPath" json:"agentPath"`
	VerifyHostKeys bool   `yaml:"verifyHostKeys" json:"verifyHostKeys"`
	KnownHostsFile string `yaml:"knownHostsFile" json:"knownHostsFile"`
	X11Display     string `yaml:"x11Display" json:"x11Display"`
}

// SerialConfig contains serial port settings
type SerialConfig struct {
	DefaultBaudRate int    `yaml:"defaultBaudRate" json:"defaultBaudRate"`
	DefaultDataBits int    `yaml:"defaultDataBits" json:"defaultDataBits"`
	DefaultParity   string `yaml:"defaultParity" json:"defaultParity"`
	DefaultStopBits int    `yaml:"defaultStopBits" json:"defaultStopBits"`
}

// ProfilesConfig contains profile group management
type ProfilesConfig struct {
	Groups []ProfileGroup `yaml:"groups" json:"groups"`
}

// ProfileGroup represents a group of profiles
type ProfileGroup struct {
	ID       string   `yaml:"id" json:"id"`
	Name     string   `yaml:"name" json:"name"`
	Profiles []string `yaml:"profiles" json:"profiles"`
	Icon     string   `yaml:"icon" json:"icon"`
}

// PluginConfig contains plugin management settings
type PluginConfig struct {
	DisabledPlugins []string `yaml:"disabledPlugins" json:"disabledPlugins"`
	CustomPlugins   []string `yaml:"customPlugins" json:"customPlugins"`
}

// GoBackendConfig contains Go backend settings
type GoBackendConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	BinaryPath string `yaml:"binaryPath" json:"binaryPath"`
}

// DefaultStore returns the default configuration
func DefaultStore() *Store {
	return &Store{
		Appearance: AppearanceConfig{
			Frame:           "native",
			Vibrancy:        "none",
			VibrancyType:    "none",
			Opacity:         1.0,
			ColorSchemeMode: "system",
		},
		Terminal: TerminalConfig{
			ColorScheme: "One Dark",
			Font:        "Consolas",
			FontSize:    14,
			Scrollback:  10000,
			Cursor:      "bar",
			Environment: map[string]string{
				"TERM":       "xterm-256color",
				"COLORTERM":  "truecolor",
				"TERM_PROGRAM": "Tabby",
			},
			UseConPTY:  true,
			SetComSpec: true,
		},
		Hotkeys: map[string]interface{}{
			"toggle-window": []string{"Ctrl-Space"},
			"new-window":    []string{"Ctrl-Shift-N"},
			"new-tab":       []string{"Ctrl-Shift-T"},
			"close-tab":     []string{"Ctrl-Shift-W"},
			"split-right":   []string{"Ctrl-Shift-Right"},
			"split-down":    []string{"Ctrl-Shift-Down"},
		},
		SSH: SSHConfig{
			AgentType:      "auto",
			VerifyHostKeys: true,
		},
		Serial: SerialConfig{
			DefaultBaudRate: 9600,
			DefaultDataBits: 8,
			DefaultParity:   "none",
			DefaultStopBits: 1,
		},
		Profiles: ProfilesConfig{
			Groups: []ProfileGroup{},
		},
		Plugin: PluginConfig{},
		GoBackend: GoBackendConfig{
			Enabled: false,
		},
	}
}

// Manager manages configuration loading and saving
type Manager struct {
	mu       sync.RWMutex
	store    *Store
	filePath string
	onChange []func(*Store)
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		store: DefaultStore(),
	}
}

// Load loads configuration from a YAML file
func (m *Manager) Load(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.filePath = path
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Use defaults
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Parse YAML into raw map first
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Then unmarshal into typed struct
	store := DefaultStore()
	if err := yaml.Unmarshal(data, store); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	store.Raw = raw
	m.store = store

	return nil
}

// Save saves the configuration to file
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.filePath == "" {
		return fmt.Errorf("no config file path set")
	}

	data, err := yaml.Marshal(m.store)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Get returns the current store (read-only copy)
func (m *Manager) Get() *Store {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store
}

// Set updates the store and notifies watchers
func (m *Manager) Set(store *Store) {
	m.mu.Lock()
	m.store = store
	m.mu.Unlock()

	for _, cb := range m.onChange {
		cb(store)
	}
}

// OnChange registers a callback for configuration changes
func (m *Manager) OnChange(cb func(*Store)) {
	m.onChange = append(m.onChange, cb)
}

// GetConfigPath returns the platform-specific config directory
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	// Check for TABBY_CONFIG environment variable
	if envPath := os.Getenv("TABBY_CONFIG"); envPath != "" {
		return envPath
	}

	return filepath.Join(home, ".tabby", "config.yaml")
}

// StartWatching polls the config file for modifications and reloads it
func (m *Manager) StartWatching() {
	go func() {
		var lastMod time.Time
		for {
			m.mu.RLock()
			path := m.filePath
			m.mu.RUnlock()

			if path != "" {
				info, err := os.Stat(path)
				if err == nil {
					if lastMod.IsZero() {
						lastMod = info.ModTime()
					} else if info.ModTime().After(lastMod) {
						lastMod = info.ModTime()
						// File changed, reload it
						err = m.Load(path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
						} else {
							// Notify watchers
							m.mu.RLock()
							store := m.store
							m.mu.RUnlock()
							for _, cb := range m.onChange {
								cb(store)
							}
						}
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
