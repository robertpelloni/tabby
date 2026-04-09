package recovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerCreation(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestRegisterTab(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterTab(TabState{
		ID:    "tab-1",
		Type:  "ssh",
		Title: "My Server",
		Options: map[string]interface{}{
			"host": "example.com",
		},
	})

	tabs := mgr.GetRecoverableTabs()
	if len(tabs) != 1 {
		t.Fatalf("Expected 1 tab, got %d", len(tabs))
	}
	if tabs[0].ID != "tab-1" || tabs[0].Type != "ssh" {
		t.Errorf("Tab mismatch: %+v", tabs[0])
	}
	if tabs[0].CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestUnregisterTab(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterTab(TabState{ID: "tab-1", Type: "ssh"})
	mgr.UnregisterTab("tab-1")

	tabs := mgr.GetRecoverableTabs()
	if len(tabs) != 0 {
		t.Errorf("Expected 0 tabs after unregister, got %d", len(tabs))
	}
}

func TestUpdateTab(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterTab(TabState{ID: "tab-1", Type: "ssh", Title: "Old"})
	mgr.UpdateTab("tab-1", "New Title")

	tabs := mgr.GetRecoverableTabs()
	if tabs[0].Title != "New Title" {
		t.Errorf("Expected 'New Title', got %q", tabs[0].Title)
	}
}

func TestRegisterSession(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterSession(SessionState{
		TabID:        "tab-1",
		ConnectionID: "conn-1",
	})

	mgr.mu.RLock()
	sess := mgr.sessions["tab-1"]
	mgr.mu.RUnlock()

	if sess == nil || !sess.Connected {
		t.Error("Session should be registered and connected")
	}
}

func TestUnregisterSession(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterSession(SessionState{TabID: "tab-1", ConnectionID: "conn-1"})
	mgr.UnregisterSession("tab-1")

	mgr.mu.RLock()
	sess := mgr.sessions["tab-1"]
	mgr.mu.RUnlock()

	if sess == nil || sess.Connected {
		t.Error("Session should be disconnected after unregister")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "recovery.json")

	mgr := NewManager()
	mgr.RegisterTab(TabState{
		ID:    "tab-1",
		Type:  "ssh",
		Title: "My Server",
		Options: map[string]interface{}{
			"host": "example.com",
			"port": float64(22),
		},
	})
	mgr.RegisterSession(SessionState{
		TabID:        "tab-1",
		ConnectionID: "conn-1",
	})

	if err := mgr.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Recovery file should not be empty")
	}

	// Load into new manager
	mgr2 := NewManager()
	file, err := mgr2.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if file == nil {
		t.Fatal("File should not be nil")
	}
	if len(file.Tabs) != 1 {
		t.Errorf("Expected 1 tab, got %d", len(file.Tabs))
	}
	if file.Tabs[0].ID != "tab-1" {
		t.Errorf("Tab ID mismatch: %s", file.Tabs[0].ID)
	}
}

func TestLoadNonexistent(t *testing.T) {
	mgr := NewManager()
	file, err := mgr.Load("/nonexistent/path/recovery.json")
	if err != nil {
		t.Errorf("Loading nonexistent file should not error: %v", err)
	}
	if file != nil {
		t.Error("File should be nil for nonexistent path")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "recovery.json")

	mgr := NewManager()
	mgr.RegisterTab(TabState{ID: "tab-1", Type: "ssh"})
	mgr.Save(path)
	mgr.Clear()

	tabs := mgr.GetRecoverableTabs()
	if len(tabs) != 0 {
		t.Errorf("Expected 0 tabs after clear, got %d", len(tabs))
	}

	// File should be removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Recovery file should be deleted after clear")
	}
}

func TestMultipleTabs(t *testing.T) {
	mgr := NewManager()
	for i := 0; i < 5; i++ {
		mgr.RegisterTab(TabState{
			ID:    string(rune('a' + i)),
			Type:  "ssh",
			Title: "Tab " + string(rune('0'+i)),
		})
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	tabs := mgr.GetRecoverableTabs()
	if len(tabs) != 5 {
		t.Errorf("Expected 5 tabs, got %d", len(tabs))
	}
}

func TestGetRecoveryPath(t *testing.T) {
	path := GetRecoveryPath()
	if path == "" {
		t.Error("Recovery path should not be empty")
	}
	if filepath.Base(path) != "recovery.json" {
		t.Errorf("Expected 'recovery.json' basename, got %q", filepath.Base(path))
	}
}
