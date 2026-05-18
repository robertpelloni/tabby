package keychain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robertpelloni/tabby/tabby-go/pkg/vault"
)

func TestKeychainStoreAndGet(t *testing.T) {
	v := vault.NewManager()

	// Create and open a temp vault
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	err := v.Create(vaultPath, "test-passphrase")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	err = v.Open(vaultPath, "test-passphrase")
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}
	defer v.Close()

	k := NewKeychain(v)

	// OS keyring will fail in test env, so it should fall back to vault
	err = k.Store("test-key", "test-value")
	if err != nil {
		t.Fatalf("Failed to store: %v", err)
	}

	value, err := k.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}
}

func TestKeychainDelete(t *testing.T) {
	v := vault.NewManager()
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	_ = v.Create(vaultPath, "test-passphrase")
	_ = v.Open(vaultPath, "test-passphrase")
	defer v.Close()

	k := NewKeychain(v)

	_ = k.Store("delete-key", "delete-value")
	err := k.Delete("delete-key")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, err = k.Get("delete-key")
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestKeychainGetNonexistent(t *testing.T) {
	v := vault.NewManager()
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	_ = v.Create(vaultPath, "test-passphrase")
	_ = v.Open(vaultPath, "test-passphrase")
	defer v.Close()

	k := NewKeychain(v)

	_, err := k.Get("nonexistent-key")
	if err == nil {
		t.Error("Expected error for nonexistent key, got nil")
	}
}

func TestKeychainOSAvailability(t *testing.T) {
	v := vault.NewManager()
	k := NewKeychain(v)

	// Initially should try OS keyring
	if !k.IsOSKeyringAvailable() {
		t.Error("Expected OS keyring to be initially available")
	}

	// Can disable it
	k.SetOSKeyringEnabled(false)
	if k.IsOSKeyringAvailable() {
		t.Error("Expected OS keyring to be disabled after SetOSKeyringEnabled(false)")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
