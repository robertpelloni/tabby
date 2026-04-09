// Package hotkey provides keyboard shortcut management for Tabby's Go backend.
//
// It handles hotkey registration, parsing of key sequences (e.g., "Ctrl-Shift-T"),
// conflict detection, and platform-specific key normalization.
package hotkey

import (
	"fmt"
	"strings"
	"sync"
)

// Hotkey represents a keyboard shortcut
type Hotkey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keys        []string `json:"keys"`    // e.g., ["Ctrl-Shift-T"]
	Category    string   `json:"category"`
	Global       bool     `json:"global"`
}

// Manager manages hotkey registration and dispatch
type Manager struct {
	mu       sync.RWMutex
	hotkeys  map[string]*Hotkey
	handlers map[string]func()
}

// NewManager creates a new hotkey manager
func NewManager() *Manager {
	return &Manager{
		hotkeys:  make(map[string]*Hotkey),
		handlers: make(map[string]func()),
	}
}

// Register registers a hotkey with a handler
func (m *Manager) Register(hotkey *Hotkey, handler func()) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for conflicts
	for id, existing := range m.hotkeys {
		if id != hotkey.ID && keysConflict(existing.Keys, hotkey.Keys) {
			return fmt.Errorf("hotkey conflict: %q conflicts with %q", hotkey.ID, id)
		}
	}

	m.hotkeys[hotkey.ID] = hotkey
	m.handlers[hotkey.ID] = handler
	return nil
}

// Unregister removes a hotkey
func (m *Manager) Unregister(id string) {
	m.mu.Lock()
	delete(m.hotkeys, id)
	delete(m.handlers, id)
	m.mu.Unlock()
}

// Trigger fires a hotkey by ID
func (m *Manager) Trigger(id string) bool {
	m.mu.RLock()
	handler, ok := m.handlers[id]
	m.mu.RUnlock()

	if ok && handler != nil {
		handler()
		return true
	}
	return false
}

// Get returns a hotkey by ID
func (m *Manager) Get(id string) (*Hotkey, bool) {
	m.mu.RLock()
	h, ok := m.hotkeys[id]
	m.mu.RUnlock()
	return h, ok
}

// List returns all registered hotkeys
func (m *Manager) List() []*Hotkey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Hotkey, 0, len(m.hotkeys))
	for _, h := range m.hotkeys {
		result = append(result, h)
	}
	return result
}

// ResolveKeySequence normalizes a key sequence to a canonical form
func ResolveKeySequence(seq string) string {
	seq = strings.TrimSpace(seq)

	// Normalize modifier names
	replacements := map[string]string{
		"cmd":   "Command",
		"super": "Super",
		"win":   "Super",
		"meta":  "Super",
		"ctrl":  "Ctrl",
		"control": "Ctrl",
		"alt":   "Alt",
		"option": "Alt",
		"shift": "Shift",
	}

	parts := strings.Split(seq, "-")
	for i, part := range parts {
		lower := strings.ToLower(part)
		if replacement, ok := replacements[lower]; ok {
			parts[i] = replacement
		} else {
			// Capitalize the key
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}

	return strings.Join(parts, "-")
}

// keysConflict checks if two key sets conflict
func keysConflict(a, b []string) bool {
	for _, ak := range a {
		for _, bk := range b {
			if ResolveKeySequence(ak) == ResolveKeySequence(bk) {
				return true
			}
		}
	}
	return false
}

// DefaultHotkeys returns the default hotkey set
func DefaultHotkeys() []*Hotkey {
	return []*Hotkey{
		{ID: "toggle-window", Name: "Toggle Window", Keys: []string{"Ctrl-Space"}, Category: "window", Global: true},
		{ID: "new-window", Name: "New Window", Keys: []string{"Ctrl-Shift-N"}, Category: "window"},
		{ID: "close-window", Name: "Close Window", Keys: []string{"Ctrl-Shift-Q"}, Category: "window"},
		{ID: "new-tab", Name: "New Tab", Keys: []string{"Ctrl-Shift-T"}, Category: "tabs"},
		{ID: "close-tab", Name: "Close Tab", Keys: []string{"Ctrl-Shift-W"}, Category: "tabs"},
		{ID: "next-tab", Name: "Next Tab", Keys: []string{"Ctrl-Shift-Right"}, Category: "tabs"},
		{ID: "prev-tab", Name: "Previous Tab", Keys: []string{"Ctrl-Shift-Left"}, Category: "tabs"},
		{ID: "split-right", Name: "Split Right", Keys: []string{"Ctrl-Shift-E"}, Category: "split"},
		{ID: "split-down", Name: "Split Down", Keys: []string{"Ctrl-Shift-O"}, Category: "split"},
		{ID: "focus-up", Name: "Focus Up", Keys: []string{"Ctrl-Alt-Up"}, Category: "split"},
		{ID: "focus-down", Name: "Focus Down", Keys: []string{"Ctrl-Alt-Down"}, Category: "split"},
		{ID: "focus-left", Name: "Focus Left", Keys: []string{"Ctrl-Alt-Left"}, Category: "split"},
		{ID: "focus-right", Name: "Focus Right", Keys: []string{"Ctrl-Alt-Right"}, Category: "split"},
		{ID: "copy", Name: "Copy", Keys: []string{"Ctrl-Shift-C"}, Category: "clipboard"},
		{ID: "paste", Name: "Paste", Keys: []string{"Ctrl-Shift-V"}, Category: "clipboard"},
		{ID: "select-all", Name: "Select All", Keys: []string{"Ctrl-Shift-A"}, Category: "clipboard"},
		{ID: "zoom-in", Name: "Zoom In", Keys: []string{"Ctrl-Plus"}, Category: "view"},
		{ID: "zoom-out", Name: "Zoom Out", Keys: []string{"Ctrl-Minus"}, Category: "view"},
		{ID: "zoom-reset", Name: "Reset Zoom", Keys: []string{"Ctrl-0"}, Category: "view"},
		{ID: "fullscreen", Name: "Toggle Fullscreen", Keys: []string{"F11"}, Category: "view"},
		{ID: "settings", Name: "Open Settings", Keys: []string{"Ctrl-Comma"}, Category: "general"},
		{ID: "command-palette", Name: "Command Palette", Keys: []string{"Ctrl-Shift-P"}, Category: "general"},
		{ID: "profile-selector", Name: "Profile Selector", Keys: []string{"Ctrl-Shift-A"}, Category: "general"},
	}
}
