package settings

import (
	"os"
	"runtime"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()

	if s.ColorScheme != "Tabby Default" {
		t.Errorf("expected default ColorScheme 'Tabby Default', got %q", s.ColorScheme)
	}
	if s.FontSize != 14 {
		t.Errorf("expected default FontSize 14, got %d", s.FontSize)
	}
	if s.Theme != "dark" {
		t.Errorf("expected default Theme 'dark', got %q", s.Theme)
	}
	if s.Frontend != "xterm-webgl" {
		t.Errorf("expected default Frontend 'xterm-webgl', got %q", s.Frontend)
	}
	if s.Scrollback != 25000 {
		t.Errorf("expected default Scrollback 25000, got %d", s.Scrollback)
	}
	if s.Opacity != 1.0 {
		t.Errorf("expected default Opacity 1.0, got %f", s.Opacity)
	}
	if !s.Animations {
		t.Errorf("expected Animations to default to true")
	}

	// Verify platform-specific overrides applied
	switch runtime.GOOS {
	case "windows":
		if s.FontFamily != "Consolas" {
			t.Errorf("expected Windows FontFamily 'Consolas', got %q", s.FontFamily)
		}
		if s.RightClick != "clipboard" {
			t.Errorf("expected Windows RightClick 'clipboard', got %q", s.RightClick)
		}
		if !s.UseConPTY {
			t.Errorf("expected Windows UseConPTY to default to true")
		}
	case "darwin":
		if s.FontFamily != "Menlo" {
			t.Errorf("expected macOS FontFamily 'Menlo', got %q", s.FontFamily)
		}
		if !s.PasteOnMiddleClick {
			t.Errorf("expected macOS PasteOnMiddleClick to default to true")
		}
	case "linux":
		if s.FontFamily != "Liberation Mono" {
			t.Errorf("expected Linux FontFamily 'Liberation Mono', got %q", s.FontFamily)
		}
		if s.PasteOnMiddleClick {
			t.Errorf("expected Linux PasteOnMiddleClick to default to false")
		}
	}
}

func TestListColorSchemes(t *testing.T) {
	schemes := ListColorSchemes()
	if len(schemes) == 0 {
		t.Fatalf("expected non-empty list of color schemes")
	}

	// Check for standard built-ins
	foundDefault := false
	foundDracula := false
	for _, name := range schemes {
		if name == "Tabby Default" {
			foundDefault = true
		}
		if name == "Dracula" {
			foundDracula = true
		}
	}
	if !foundDefault {
		t.Errorf("ListColorSchemes missing 'Tabby Default'")
	}
	if !foundDracula {
		t.Errorf("ListColorSchemes missing 'Dracula'")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	// Create a temporary directory to isolate test settings
	tmpDir, err := os.MkdirTemp("", "tabby-settings-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original env vars
	origAppdata := os.Getenv("APPDATA")
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("APPDATA", origAppdata)
		os.Setenv("HOME", origHome)
	}()

	// Force ConfigDir() to point to tmpDir
	os.Unsetenv("APPDATA")
	os.Setenv("HOME", tmpDir)

	cfgDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if cfgDir == "" {
		t.Fatalf("ConfigDir() returned empty string")
	}

	// 1. Loading with no file present should return DefaultSettings without error
	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() with missing file should succeed, got: %v", err)
	}
	if loaded.FontSize != 14 {
		t.Errorf("expected default FontSize 14, got %d", loaded.FontSize)
	}

	// 2. Modify settings and save
	loaded.FontSize = 20
	loaded.ColorScheme = "Nord"
	loaded.CursorBlink = false
	loaded.LastTabClosesWindow = true

	if err := SaveSettings(loaded); err != nil {
		t.Fatalf("SaveSettings() failed: %v", err)
	}

	// Check file was created
	settingsFile, err := SettingsFile()
	if err != nil {
		t.Fatalf("SettingsFile() error: %v", err)
	}
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		t.Fatalf("settings file was not created at %s", settingsFile)
	}

	// 3. Load saved settings and verify round-trip
	reloaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() failed after save: %v", err)
	}
	if reloaded.FontSize != 20 {
		t.Errorf("expected FontSize 20, got %d", reloaded.FontSize)
	}
	if reloaded.ColorScheme != "Nord" {
		t.Errorf("expected ColorScheme 'Nord', got %q", reloaded.ColorScheme)
	}
	if reloaded.CursorBlink {
		t.Errorf("expected CursorBlink false, got true")
	}
	if !reloaded.LastTabClosesWindow {
		t.Errorf("expected LastTabClosesWindow true, got false")
	}

	// 4. ResetSettings should restore defaults
	if err := ResetSettings(); err != nil {
		t.Fatalf("ResetSettings() failed: %v", err)
	}
	resetLoaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() after reset failed: %v", err)
	}
	if resetLoaded.FontSize != 14 {
		t.Errorf("expected FontSize reset to 14, got %d", resetLoaded.FontSize)
	}
	if resetLoaded.LastTabClosesWindow {
		t.Errorf("expected LastTabClosesWindow reset to false, got true")
	}
}
