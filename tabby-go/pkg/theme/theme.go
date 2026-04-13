// Package theme provides theme/color scheme management for Tabby's Go backend.
//
// It handles terminal color schemes (ANSI 16 colors + foreground/background),
// native UI theming (light/dark mode), and CSS variable computation for
// contrast-aware color adjustments.
package theme

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"sync"
)

// ColorScheme represents a terminal color scheme with 16 ANSI colors
type ColorScheme struct {
	Foreground       string `json:"foreground"`
	Background       string `json:"background"`
	Cursor           string `json:"cursor"`
	CursorAccent     string `json:"cursorAccent"`
	SelectionBackground string `json:"selectionBackground"`
	Colors           [16]string `json:"colors"` // ANSI 0-15
}

// Theme represents a complete UI theme
type Theme struct {
	Name               string      `json:"name"`
	DisplayName        string      `json:"displayName,omitempty"`
	CSS                string      `json:"css,omitempty"`
	FollowsColorScheme bool        `json:"followsColorScheme"`
	ColorSchemeDark    ColorScheme `json:"colorSchemeDark"`
	ColorSchemeLight   ColorScheme `json:"colorSchemeLight"`
}

// ThemeMode represents light/dark/auto mode
type ThemeMode string

const (
	ThemeModeAuto  ThemeMode = "auto"
	ThemeModeLight ThemeMode = "light"
	ThemeModeDark  ThemeMode = "dark"
)

// Manager manages themes and color schemes
type Manager struct {
	mu            sync.RWMutex
	themes        []Theme
	activeTheme   string
	mode          ThemeMode
	onChange      []func(Theme)
	customCSS     string
}

// NewManager creates a new theme manager
func NewManager() *Manager {
	m := &Manager{
		mode: ThemeModeAuto,
	}
	m.registerDefaults()
	return m
}

// registerDefaults registers built-in themes
func (m *Manager) registerDefaults() {
	m.themes = []Theme{
		{
			Name:               "default",
			DisplayName:        "Default",
			FollowsColorScheme: true,
			ColorSchemeDark: ColorScheme{
				Foreground: "#cccccc", Background: "#1a1a1a",
				Cursor: "#ffffff", SelectionBackground: "#444444",
				Colors: [16]string{
					"#000000", "#cc0000", "#4e9a06", "#c4a000",
					"#3465a4", "#75507b", "#06989a", "#d3d7cf",
					"#555753", "#ef2929", "#8ae234", "#fce94f",
					"#729fcf", "#ad7fa8", "#34e2e2", "#eeeeec",
				},
			},
			ColorSchemeLight: ColorScheme{
				Foreground: "#333333", Background: "#ffffff",
				Cursor: "#000000", SelectionBackground: "#cccccc",
				Colors: [16]string{
					"#000000", "#cc0000", "#4e9a06", "#c4a000",
					"#3465a4", "#75507b", "#06989a", "#d3d7cf",
					"#555753", "#ef2929", "#8ae234", "#fce94f",
					"#729fcf", "#ad7fa8", "#34e2e2", "#eeeeec",
				},
			},
		},
		{
			Name:               "compact",
			DisplayName:        "Compact",
			FollowsColorScheme: false,
			ColorSchemeDark: ColorScheme{
				Foreground: "#d4d4d4", Background: "#1e1e1e",
				Cursor: "#d4d4d4", SelectionBackground: "#264f78",
				Colors: [16]string{
					"#000000", "#cd3131", "#0dbc79", "#e5e510",
					"#2472c8", "#bc3fbc", "#11a8cd", "#e5e5e5",
					"#666666", "#f14c4c", "#23d18b", "#f5f543",
					"#3b8eea", "#d670d6", "#29b8db", "#ffffff",
				},
			},
			ColorSchemeLight: ColorScheme{
				Foreground: "#1e1e1e", Background: "#ffffff",
				Cursor: "#1e1e1e", SelectionBackground: "#add6ff",
				Colors: [16]string{
					"#000000", "#cd3131", "#00bc00", "#949800",
					"#0451a5", "#bc05bc", "#0598bc", "#555555",
					"#666666", "#cd3131", "#14ce14", "#b5ba00",
					"#0451a5", "#bc05bc", "#0598bc", "#a5a5a5",
				},
			},
		},
		{
			Name:               "seoul256",
			DisplayName:        "Seoul256",
			FollowsColorScheme: false,
			ColorSchemeDark: ColorScheme{
				Foreground: "#d0d0d0", Background: "#3a3a3a",
				Cursor: "#d0d0d0", SelectionBackground: "#4e4e4e",
				Colors: [16]string{
					"#4e4e4e", "#d68787", "#5f865f", "#d8af5f",
					"#85add4", "#d7afaf", "#87afaf", "#d0d0d0",
					"#626262", "#d75f87", "#5faf5f", "#d7d787",
					"#87afd7", "#d7afd7", "#5fd7d7", "#e4e4e4",
				},
			},
			ColorSchemeLight: ColorScheme{
				Foreground: "#3a3a3a", Background: "#e4e4e4",
				Cursor: "#3a3a3a", SelectionBackground: "#c0c0c0",
				Colors: [16]string{
					"#4e4e4e", "#d68787", "#5f865f", "#d8af5f",
					"#85add4", "#d7afaf", "#87afaf", "#d0d0d0",
					"#626262", "#d75f87", "#5faf5f", "#d7d787",
					"#87afd7", "#d7afd7", "#5fd7d7", "#e4e4e4",
				},
			},
		},
	}
}

// SetTheme sets the active theme by name
func (m *Manager) SetTheme(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, theme := range m.themes {
		if theme.Name == name {
			m.activeTheme = name
			m.notifyChange(theme)
			return nil
		}
	}
	return fmt.Errorf("theme not found: %s", name)
}

// GetTheme returns the current active theme
func (m *Manager) GetTheme() Theme {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, theme := range m.themes {
		if theme.Name == m.activeTheme {
			return theme
		}
	}
	return m.themes[0]
}

// GetActiveColorScheme returns the color scheme based on the current mode
func (m *Manager) GetActiveColorScheme() ColorScheme {
	theme := m.GetTheme()
	mode := m.GetMode()

	if mode == ThemeModeLight {
		return theme.ColorSchemeLight
	}
	return theme.ColorSchemeDark
}

// SetMode sets the light/dark/auto mode
func (m *Manager) SetMode(mode ThemeMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	m.notifyChange(m.GetTheme())
}

// GetMode returns the current theme mode
func (m *Manager) GetMode() ThemeMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetCustomCSS sets custom CSS overrides
func (m *Manager) SetCustomCSS(css string) {
	m.mu.Lock()
	m.customCSS = css
	m.mu.Unlock()
}

// GetCustomCSS returns custom CSS overrides
func (m *Manager) GetCustomCSS() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.customCSS
}

// ListThemes returns all available themes
func (m *Manager) ListThemes() []Theme {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Theme, len(m.themes))
	copy(result, m.themes)
	return result
}

// RegisterTheme adds or replaces a theme
func (m *Manager) RegisterTheme(theme Theme) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, t := range m.themes {
		if t.Name == theme.Name {
			m.themes[i] = theme
			return
		}
	}
	m.themes = append(m.themes, theme)
}

// OnChange registers a callback for theme changes
func (m *Manager) OnChange(cb func(Theme)) {
	m.mu.Lock()
	m.onChange = append(m.onChange, cb)
	m.mu.Unlock()
}

// notifyChange calls all registered change callbacks
func (m *Manager) notifyChange(theme Theme) {
	for _, cb := range m.onChange {
		cb(theme)
	}
}

// LoadThemesFromFile loads themes from a JSON file
func (m *Manager) LoadThemesFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read themes file: %w", err)
	}

	var themes []Theme
	if err := json.Unmarshal(data, &themes); err != nil {
		return fmt.Errorf("failed to parse themes: %w", err)
	}

	for _, theme := range themes {
		m.RegisterTheme(theme)
	}
	return nil
}

// SaveThemesToFile saves all themes to a JSON file
func (m *Manager) SaveThemesToFile(path string) error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.themes, "", "  ")
	m.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ParseHexColor parses a hex color string (#RRGGBB) into a color.RGBA
func ParseHexColor(hex string) (color.RGBA, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{}, fmt.Errorf("invalid hex color: %s", hex)
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// Luminance calculates the relative luminance of a color
func Luminance(c color.RGBA) float64 {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// IsDark returns true if a color scheme has a dark background
func IsDark(scheme ColorScheme) bool {
	bg, err := ParseHexColor(scheme.Background)
	if err != nil {
		return true
	}
	fg, err := ParseHexColor(scheme.Foreground)
	if err != nil {
		return true
	}
	return Luminance(bg) < Luminance(fg)
}
