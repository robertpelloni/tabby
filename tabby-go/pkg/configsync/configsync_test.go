package configsync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService("/tmp/test-config.json")
	if svc == nil {
		t.Fatal("Service should not be nil")
	}
}

func TestGetStatus(t *testing.T) {
	svc := NewService("/tmp/test-config.json")
	status := svc.GetStatus()
	if status.ConfigHash != "" {
		t.Error("Initial config hash should be empty")
	}
}

func TestMarkDirty(t *testing.T) {
	svc := NewService("/tmp/test-config.json")
	svc.MarkDirty()
	if !svc.GetStatus().Dirty {
		t.Error("Should be dirty after MarkDirty")
	}
}

func TestLoadSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	svc := NewService(path)

	// Save a config
	type TestConfig struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	}
	original := TestConfig{Name: "test", Version: 42}
	if err := svc.SaveConfig(&original); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Config file should exist")
	}

	// Load it back
	var loaded TestConfig
	if err := svc.LoadConfig(&loaded); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.Name != "test" || loaded.Version != 42 {
		t.Errorf("Config mismatch: %+v", loaded)
	}

	// Status should show clean
	if svc.GetStatus().Dirty {
		t.Error("Config should be clean after save")
	}
}

func TestCheckForChanges(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	svc := NewService(path)

	// Write initial content
	os.WriteFile(path, []byte(`{"key": "value"}`), 0644)

	svc.CheckForChanges()
	status := svc.GetStatus()
	if status.ConfigHash == "" {
		t.Error("Should have computed config hash")
	}

	// Modify the file
	os.WriteFile(path, []byte(`{"key": "modified"}`), 0644)

	svc.CheckForChanges()
	newStatus := svc.GetStatus()
	if newStatus.ConfigHash == status.ConfigHash {
		t.Error("Hash should change when file changes")
	}
}

func TestCheckForChangesNonexistent(t *testing.T) {
	svc := NewService("/nonexistent/config.json")
	svc.CheckForChanges()
	// Should not panic
}

func TestOnChange(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	svc := NewService(path)

	var received SyncStatus
	svc.OnChange(func(s SyncStatus) {
		received = s
	})

	os.WriteFile(path, []byte(`{"test": true}`), 0644)
	svc.CheckForChanges()

	if received.ConfigHash == "" {
		t.Error("Should have received status with hash")
	}
}

func TestStop(t *testing.T) {
	svc := NewService("/tmp/test.json")
	// Stop should not panic
	svc.Stop()
}

func TestAtomicSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	svc := NewService(path)

	// Save config - should create file atomically
	data := map[string]string{"key": "value"}
	if err := svc.SaveConfig(&data); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Temp file should not exist
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("Temp file should be cleaned up")
	}
}
