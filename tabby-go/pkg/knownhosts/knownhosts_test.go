package knownhosts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerCreation(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
	if len(mgr.List()) != 0 {
		t.Error("New manager should have no entries")
	}
}

func TestStoreAndGet(t *testing.T) {
	mgr := NewManager()
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, "SHA256:abc123", []byte("keybytes"))

	entry := mgr.GetFor(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"})
	if entry == nil {
		t.Fatal("Should find stored entry")
	}
	if entry.Digest != "SHA256:abc123" {
		t.Errorf("Expected 'SHA256:abc123', got %q", entry.Digest)
	}
}

func TestGetNonexistent(t *testing.T) {
	mgr := NewManager()
	entry := mgr.GetFor(Selector{Host: "nonexistent.com", Port: 22, Type: "ssh-ed25519"})
	if entry != nil {
		t.Error("Should not find non-existent entry")
	}
}

func TestStoreUpdate(t *testing.T) {
	mgr := NewManager()
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, "SHA256:old", nil)
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, "SHA256:new", nil)

	entry := mgr.GetFor(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"})
	if entry.Digest != "SHA256:new" {
		t.Errorf("Expected updated digest, got %q", entry.Digest)
	}
}

func TestRemove(t *testing.T) {
	mgr := NewManager()
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, "SHA256:abc", nil)
	mgr.Remove(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"})

	entry := mgr.GetFor(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"})
	if entry != nil {
		t.Error("Entry should be removed")
	}
}

func TestVerify(t *testing.T) {
	mgr := NewManager()
	keyBytes := []byte("test-public-key-bytes")
	digest := FingerprintSHA256(keyBytes)
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, digest, keyBytes)

	// Same key should verify
	ok, err := mgr.Verify(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, keyBytes)
	if !ok || err != nil {
		t.Errorf("Expected verification to succeed, got ok=%v err=%v", ok, err)
	}

	// Different key should fail
	ok, err = mgr.Verify(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, []byte("different-key"))
	if ok {
		t.Error("Different key should not verify")
	}

	// Unknown host should return false with no error
	ok, err = mgr.Verify(Selector{Host: "unknown.com", Port: 22, Type: "ssh-ed25519"}, keyBytes)
	if ok || err != nil {
		t.Errorf("Unknown host: ok=%v err=%v", ok, err)
	}
}

func TestList(t *testing.T) {
	mgr := NewManager()
	mgr.Store(Selector{Host: "a.com", Port: 22, Type: "ssh-ed25519"}, "SHA256:a", nil)
	mgr.Store(Selector{Host: "b.com", Port: 22, Type: "ssh-rsa"}, "SHA256:b", nil)

	list := mgr.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(list))
	}
}

func TestFingerprintSHA256(t *testing.T) {
	fp := FingerprintSHA256([]byte("test"))
	if fp[:7] != "SHA256:" {
		t.Errorf("Fingerprint should start with 'SHA256:', got %q", fp)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "known_hosts")

	key1 := []byte("test-public-key-1")
	key2 := []byte("test-public-key-2")
	digest1 := FingerprintSHA256(key1)
	digest2 := FingerprintSHA256(key2)

	mgr := NewManager()
	mgr.Store(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"}, digest1, key1)
	mgr.Store(Selector{Host: "other.com", Port: 2222, Type: "ssh-rsa"}, digest2, key2)

	if err := mgr.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Known hosts file should exist")
	}

	// Load into a new manager
	mgr2 := NewManager()
	if err := mgr2.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	list := mgr2.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 loaded entries, got %d", len(list))
	}

	entry := mgr2.GetFor(Selector{Host: "example.com", Port: 22, Type: "ssh-ed25519"})
	if entry == nil {
		t.Fatal("Should find entry for example.com")
	}
	if entry.Digest != digest1 {
		t.Errorf("Loaded digest mismatch: expected %q, got %q", digest1, entry.Digest)
	}
}

func TestLoadNonexistent(t *testing.T) {
	mgr := NewManager()
	err := mgr.LoadFromFile("/nonexistent/path/known_hosts")
	if err != nil {
		t.Errorf("Loading nonexistent file should not error, got: %v", err)
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		port     int
	}{
		{"example.com", "example.com", 22},
		{"[example.com]:2222", "example.com", 2222},
		{"example.com:2222", "example.com", 2222},
	}

	for _, tt := range tests {
		host, port := parseHostPort(tt.input)
		if host != tt.expected || port != tt.port {
			t.Errorf("parseHostPort(%q) = (%q, %d), want (%q, %d)",
				tt.input, host, port, tt.expected, tt.port)
		}
	}
}

func TestFormatHostPort(t *testing.T) {
	tests := []struct {
		host     string
		port     int
		expected string
	}{
		{"example.com", 22, "example.com"},
		{"example.com", 2222, "example.com:2222"},
		{"::1", 2222, "[::1]:2222"},
	}

	for _, tt := range tests {
		result := formatHostPort(tt.host, tt.port)
		if result != tt.expected {
			t.Errorf("formatHostPort(%q, %d) = %q, want %q",
				tt.host, tt.port, result, tt.expected)
		}
	}
}
