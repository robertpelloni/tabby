package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultStore(t *testing.T) {
	store := DefaultStore()
	if store == nil {
		t.Fatal("DefaultStore should not be nil")
	}
	if store.Terminal.Font != "Consolas" {
		t.Errorf("Expected default font 'Consolas', got %q", store.Terminal.Font)
	}
	if store.Terminal.FontSize != 14 {
		t.Errorf("Expected default font size 14, got %d", store.Terminal.FontSize)
	}
	if store.SSH.AgentType != "auto" {
		t.Errorf("Expected default SSH agent type 'auto', got %q", store.SSH.AgentType)
	}
	if store.Serial.DefaultBaudRate != 9600 {
		t.Errorf("Expected default baud rate 9600, got %d", store.Serial.DefaultBaudRate)
	}
}

func TestManagerCreation(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
	store := mgr.Get()
	if store == nil {
		t.Fatal("Get() should return default store")
	}
}

func TestLoadNonexistent(t *testing.T) {
	mgr := NewManager()
	err := mgr.Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("Should not error on nonexistent file: %v", err)
	}
	// Should return defaults
	store := mgr.Get()
	if store.Terminal.Font != "Consolas" {
		t.Error("Should have defaults after loading nonexistent file")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	mgr := NewManager()
	mgr.Load(path) // sets the path
	store := mgr.Get()
	store.Terminal.Font = "JetBrains Mono"
	store.Terminal.FontSize = 12
	store.SSH.VerifyHostKeys = false
	mgr.Set(store)

	err := mgr.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Config file should exist after save")
	}

	// Load into new manager
	mgr2 := NewManager()
	err = mgr2.Load(path)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	loaded := mgr2.Get()
	if loaded.Terminal.Font != "JetBrains Mono" {
		t.Errorf("Expected font 'JetBrains Mono', got %q", loaded.Terminal.Font)
	}
	if loaded.Terminal.FontSize != 12 {
		t.Errorf("Expected font size 12, got %d", loaded.Terminal.FontSize)
	}
}

func TestOnChange(t *testing.T) {
	mgr := NewManager()
	changed := false
	mgr.OnChange(func(s *Store) {
		changed = true
	})

	store := mgr.Get()
	store.Terminal.Font = "Test"
	mgr.Set(store)

	if !changed {
		t.Error("OnChange callback should have been called")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("Config path should not be empty")
	}
}

func TestGetConfigPathEnv(t *testing.T) {
	os.Setenv("TABBY_CONFIG", "/custom/path/config.yaml")
	defer os.Unsetenv("TABBY_CONFIG")

	path := GetConfigPath()
	if path != "/custom/path/config.yaml" {
		t.Errorf("Expected env path, got %q", path)
	}
}
