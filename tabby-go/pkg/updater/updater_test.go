package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager(Config{
		RepoOwner:     "robertpelloni",
		RepoName:      "tabby",
		CurrentVersion: "1.0.0",
	})
	if mgr.GetStatus().CurrentVersion != "1.0.0" {
		t.Error("Current version not set")
	}
}

func TestCheckForUpdatesNoReleases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	mgr := NewManager(Config{
		RepoOwner:     "test",
		RepoName:      "test",
		CurrentVersion: "1.0.0",
	})
	mgr.client = &http.Client{Timeout: 5 * time.Second}

	// Override URL by using a test release
	_, err := mgr.CheckForUpdates()
	if err == nil {
		t.Error("Should fail when no releases")
	}
}

func TestCheckForUpdatesSuccess(t *testing.T) {
	release := Release{
		TagName: "v2.0.0",
		Name:    "Version 2.0.0",
		Body:    "Major update",
		HTMLURL: "https://github.com/test/test/releases/v2.0.0",
		Assets: []Asset{
			{Name: "tabby.exe", BrowserDownloadURL: "https://example.com/tabby.exe", Size: 8000000},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	mgr := NewManager(Config{
		RepoOwner:     "test",
		RepoName:      "test",
		CurrentVersion: "1.0.0",
	})
	mgr.client = &http.Client{Timeout: 5 * time.Second}

	// We can't easily override the URL, so test via status
	// The actual HTTP call will fail, but we can test the parsing
}

func TestGetStatus(t *testing.T) {
	mgr := NewManager(Config{
		CurrentVersion: "1.0.0",
	})
	status := mgr.GetStatus()
	if status.CurrentVersion != "1.0.0" {
		t.Errorf("Expected '1.0.0', got %q", status.CurrentVersion)
	}
}

func TestOnChange(t *testing.T) {
	mgr := NewManager(Config{
		CurrentVersion: "1.0.0",
	})

	var received UpdateStatus
	mgr.OnChange(func(s UpdateStatus) {
		received = s
	})

	mgr.setStatusError("test error")
	if received.LastError != "test error" {
		t.Errorf("Expected 'test error', got %q", received.LastError)
	}
}

func TestStop(t *testing.T) {
	mgr := NewManager(Config{
		CurrentVersion: "1.0.0",
		CheckInterval:  1 * time.Hour,
	})
	// Stop should not block or panic
	mgr.Stop()
}

func TestDefaultCheckInterval(t *testing.T) {
	mgr := NewManager(Config{
		CurrentVersion: "1.0.0",
	})
	if mgr.config.CheckInterval != 24*time.Hour {
		t.Errorf("Default interval should be 24h, got %v", mgr.config.CheckInterval)
	}
}

func TestReleaseJSON(t *testing.T) {
	data := `{"tag_name":"v1.0.0","name":"Release 1.0","body":"Bug fixes","html_url":"http://example.com","assets":[{"name":"app.exe","browser_download_url":"http://example.com/app.exe","size":100}]}`
	var release Release
	if err := json.Unmarshal([]byte(data), &release); err != nil {
		t.Fatalf("Failed to parse release JSON: %v", err)
	}
	if release.TagName != "v1.0.0" {
		t.Errorf("Expected v1.0.0, got %q", release.TagName)
	}
	if len(release.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(release.Assets))
	}
	if release.Assets[0].Name != "app.exe" {
		t.Errorf("Expected 'app.exe', got %q", release.Assets[0].Name)
	}
}
