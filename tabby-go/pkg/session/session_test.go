package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSession(t *testing.T) {
	// Use a temp dir to avoid polluting real config
	tmpDir := t.TempDir()
	originalConfigDir := os.Getenv("APPDATA")

	// Set APPDATA so ConfigDir uses our temp dir
	if err := os.Setenv("APPDATA", tmpDir); err != nil {
		t.Skip("Cannot set APPDATA")
	}
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("APPDATA", originalConfigDir)
		} else {
			os.Unsetenv("APPDATA")
		}
	}()

	tabs := []TabState{
		{Shell: "powershell.exe", Title: "PowerShell", Active: true},
		{Shell: "cmd.exe", Title: "CMD", Active: false},
	}

	// Save
	if err := SaveSession(tabs); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Verify file exists
	file, err := sessionFile()
	if err != nil {
		t.Fatalf("sessionFile failed: %v", err)
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Fatal("Session file was not created")
	}

	// Load
	state, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if state == nil {
		t.Fatal("LoadSession returned nil")
	}
	if len(state.Tabs) != 2 {
		t.Fatalf("Expected 2 tabs, got %d", len(state.Tabs))
	}
	if state.Tabs[0].Shell != "powershell.exe" {
		t.Errorf("Expected shell 'powershell.exe', got '%s'", state.Tabs[0].Shell)
	}
	if !state.Tabs[0].Active {
		t.Error("First tab should be active")
	}
	if state.Tabs[1].Active {
		t.Error("Second tab should not be active")
	}

	// Clear
	if err := ClearSession(); err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("Session file should have been deleted")
	}
}

func TestLoadSessionNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Setenv("APPDATA", tmpDir); err != nil {
		t.Skip("Cannot set APPDATA")
	}
	defer os.Unsetenv("APPDATA")

	state, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession should not error on missing file: %v", err)
	}
	if state != nil {
		t.Error("LoadSession should return nil when no session file exists")
	}
}

func TestSessionFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Setenv("APPDATA", tmpDir); err != nil {
		t.Skip("Cannot set APPDATA")
	}
	defer os.Unsetenv("APPDATA")

	file, err := sessionFile()
	if err != nil {
		t.Fatalf("sessionFile failed: %v", err)
	}
	expected := filepath.Join(tmpDir, "Tabby", "session.json")
	if file != expected {
		t.Errorf("Expected %s, got %s", expected, file)
	}
}
