package hotkey

import (
	"testing"
)

func TestResolveKeySequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ctrl-shift-t", "Ctrl-Shift-T"},
		{"Ctrl-Shift-T", "Ctrl-Shift-T"},
		{"CTRL-SHIFT-T", "Ctrl-Shift-T"},
		{"cmd-c", "Command-C"},
		{"super-v", "Super-V"},
		{"win-a", "Super-A"},
		{"meta-x", "Super-X"},
		{"alt-enter", "Alt-Enter"},
		{"f11", "F11"},
		{"F11", "F11"},
		{"ctrl-comma", "Ctrl-Comma"},
		{"ctrl-space", "Ctrl-Space"},
	}

	for _, tt := range tests {
		result := ResolveKeySequence(tt.input)
		if result != tt.expected {
			t.Errorf("ResolveKeySequence(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestManagerRegister(t *testing.T) {
	mgr := NewManager()
	h := &Hotkey{ID: "test", Name: "Test", Keys: []string{"Ctrl-T"}}
	called := false
	err := mgr.Register(h, func() { called = true })
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	if !mgr.Trigger("test") {
		t.Error("Trigger should return true for registered hotkey")
	}
	if !called {
		t.Error("Handler should have been called")
	}
}

func TestManagerConflict(t *testing.T) {
	mgr := NewManager()
	h1 := &Hotkey{ID: "test1", Name: "Test 1", Keys: []string{"Ctrl-T"}}
	h2 := &Hotkey{ID: "test2", Name: "Test 2", Keys: []string{"Ctrl-T"}}

	mgr.Register(h1, func() {})
	err := mgr.Register(h2, func() {})
	if err == nil {
		t.Error("Should fail with conflicting hotkey")
	}
}

func TestManagerNoConflict(t *testing.T) {
	mgr := NewManager()
	h1 := &Hotkey{ID: "test1", Name: "Test 1", Keys: []string{"Ctrl-T"}}
	h2 := &Hotkey{ID: "test2", Name: "Test 2", Keys: []string{"Ctrl-Shift-T"}}

	err := mgr.Register(h1, func() {})
	if err != nil {
		t.Fatalf("First register should succeed: %v", err)
	}
	err = mgr.Register(h2, func() {})
	if err != nil {
		t.Fatalf("Second register should succeed: %v", err)
	}
}

func TestManagerUnregister(t *testing.T) {
	mgr := NewManager()
	h := &Hotkey{ID: "test", Name: "Test", Keys: []string{"Ctrl-T"}}
	mgr.Register(h, func() {})

	mgr.Unregister("test")
	if mgr.Trigger("test") {
		t.Error("Trigger should return false after unregister")
	}
}

func TestManagerList(t *testing.T) {
	mgr := NewManager()
	mgr.Register(&Hotkey{ID: "h1", Name: "Hotkey 1", Keys: []string{"Ctrl-A"}}, func() {})
	mgr.Register(&Hotkey{ID: "h2", Name: "Hotkey 2", Keys: []string{"Ctrl-B"}}, func() {})

	list := mgr.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 hotkeys, got %d", len(list))
	}
}

func TestManagerGet(t *testing.T) {
	mgr := NewManager()
	h := &Hotkey{ID: "test", Name: "Test", Keys: []string{"Ctrl-T"}, Category: "general"}
	mgr.Register(h, func() {})

	got, ok := mgr.Get("test")
	if !ok {
		t.Error("Should find registered hotkey")
	}
	if got.Category != "general" {
		t.Errorf("Expected category 'general', got %q", got.Category)
	}

	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Error("Should not find unregistered hotkey")
	}
}

func TestDefaultHotkeys(t *testing.T) {
	defaults := DefaultHotkeys()
	if len(defaults) == 0 {
		t.Error("DefaultHotkeys should not be empty")
	}

	// Verify specific hotkeys exist
	ids := make(map[string]bool)
	for _, h := range defaults {
		ids[h.ID] = true
	}

	for _, id := range []string{"new-tab", "close-tab", "copy", "paste", "fullscreen"} {
		if !ids[id] {
			t.Errorf("Missing default hotkey: %s", id)
		}
	}
}

func TestKeysConflict(t *testing.T) {
	tests := []struct {
		a, b     []string
		conflict bool
	}{
		{[]string{"Ctrl-T"}, []string{"Ctrl-T"}, true},
		{[]string{"Ctrl-T"}, []string{"ctrl-t"}, true},
		{[]string{"Ctrl-T"}, []string{"Ctrl-Shift-T"}, false},
		{[]string{"Ctrl-T"}, []string{"Alt-T"}, false},
		{[]string{"Ctrl-T", "Ctrl-W"}, []string{"Ctrl-W"}, true},
	}

	for _, tt := range tests {
		result := keysConflict(tt.a, tt.b)
		if result != tt.conflict {
			t.Errorf("keysConflict(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.conflict)
		}
	}
}
