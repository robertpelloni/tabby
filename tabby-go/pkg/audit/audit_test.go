package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	if !l.IsEnabled() {
		t.Error("Expected logging to be enabled")
	}

	expectedPath := filepath.Join(dir, "audit.log")
	if l.GetPath() != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, l.GetPath())
	}
}

func TestLogConnect(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	err = l.LogConnect("ssh", "example.com", 22, "admin", "sess-123")
	if err != nil {
		t.Fatalf("Failed to log connect: %v", err)
	}

	// Read the log file
	data, err := os.ReadFile(l.GetPath())
	if err != nil {
		t.Fatalf("Failed to read log: %v", err)
	}

	if !strings.Contains(string(data), `"type":"connect"`) {
		t.Error("Expected connect event in log")
	}
	if !strings.Contains(string(data), `"host":"example.com"`) {
		t.Error("Expected host in log")
	}
}

func TestLogDisconnect(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	err = l.LogDisconnect("ssh", "example.com", 22, "sess-123", "5m")
	if err != nil {
		t.Fatalf("Failed to log disconnect: %v", err)
	}
}

func TestLogAuthSuccess(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	err = l.LogAuthSuccess("ssh", "example.com", 22, "admin", "publickey")
	if err != nil {
		t.Fatalf("Failed to log auth: %v", err)
	}
}

func TestLogAuthFailure(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	err = l.LogAuthFailure("ssh", "example.com", 22, "admin", "wrong password")
	if err != nil {
		t.Fatalf("Failed to log auth failure: %v", err)
	}
}

func TestSetEnabled(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	l.SetEnabled(false)
	if l.IsEnabled() {
		t.Error("Expected logging to be disabled")
	}

	// Logging should be a no-op when disabled
	err = l.LogConnect("ssh", "test.com", 22, "user", "sess")
	if err != nil {
		t.Fatalf("Log should not fail when disabled: %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
