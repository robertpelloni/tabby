package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndOpen(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()

	// Create vault
	err := mgr.Create(vaultPath, "testpassphrase")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	if !mgr.IsOpen() {
		t.Error("Vault should be open after creation")
	}

	// Close and reopen
	mgr.Close()
	if mgr.IsOpen() {
		t.Error("Vault should be closed after Close()")
	}

	err = mgr.Open(vaultPath, "testpassphrase")
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}

	if !mgr.IsOpen() {
		t.Error("Vault should be open after Open()")
	}
}

func TestWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	err := mgr.Create(vaultPath, "correct")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	mgr.Close()

	err = mgr.Open(vaultPath, "wrong")
	if err == nil {
		t.Error("Should fail with wrong passphrase")
	}
}

func TestSetAndGetSecret(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	err := mgr.SetSecret("test", map[string]string{"key": "value1"}, "secretdata")
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	value, err := mgr.GetSecret("test", map[string]string{"key": "value1"})
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if value != "secretdata" {
		t.Errorf("Expected 'secretdata', got %q", value)
	}
}

func TestGetNonexistentSecret(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	_, err := mgr.GetSecret("nonexistent", "key")
	if err == nil {
		t.Error("Should fail getting non-existent secret")
	}
}

func TestDeleteSecret(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	mgr.SetSecret("test", "key1", "value1")
	mgr.SetSecret("test", "key2", "value2")

	err := mgr.DeleteSecret("test", "key1")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	_, err = mgr.GetSecret("test", "key1")
	if err == nil {
		t.Error("Should fail getting deleted secret")
	}

	// key2 should still exist
	value, err := mgr.GetSecret("test", "key2")
	if err != nil || value != "value2" {
		t.Errorf("key2 should still exist, got err=%v value=%q", err, value)
	}
}

func TestUpdateSecret(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	mgr.SetSecret("test", "key1", "value1")
	mgr.SetSecret("test", "key1", "updated")

	value, err := mgr.GetSecret("test", "key1")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if value != "updated" {
		t.Errorf("Expected 'updated', got %q", value)
	}
}

func TestListSecrets(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	mgr.SetSecret("ssh-password", "host1", "pass1")
	mgr.SetSecret("ssh-password", "host2", "pass2")
	mgr.SetSecret("other-type", "key", "value")

	secrets, err := mgr.ListSecrets("ssh-password")
	if err != nil {
		t.Fatalf("Failed to list secrets: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(secrets))
	}
}

func TestPasswordHelpers(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	err := mgr.SetPassword("example.com", 22, "admin", "secret123")
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	pass, err := mgr.GetPassword("example.com", 22, "admin")
	if err != nil {
		t.Fatalf("Failed to get password: %v", err)
	}

	if pass != "secret123" {
		t.Errorf("Expected 'secret123', got %q", pass)
	}
}

func TestPrivateKeyPassphrase(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	mgr := NewManager()
	mgr.Create(vaultPath, "pass")

	err := mgr.SetPrivateKeyPassphrase("abc123hash", "mypassphrase")
	if err != nil {
		t.Fatalf("Failed to set passphrase: %v", err)
	}

	pp, err := mgr.GetPrivateKeyPassphrase("abc123hash")
	if err != nil {
		t.Fatalf("Failed to get passphrase: %v", err)
	}

	if pp != "mypassphrase" {
		t.Errorf("Expected 'mypassphrase', got %q", pp)
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	// Create and add secrets
	mgr1 := NewManager()
	mgr1.Create(vaultPath, "passphrase")
	mgr1.SetSecret("test", "key1", "value1")
	mgr1.SetPassword("host.com", 22, "user", "pass123")
	mgr1.Close()

	// Verify file exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file should exist")
	}

	// Reopen and verify
	mgr2 := NewManager()
	err := mgr2.Open(vaultPath, "passphrase")
	if err != nil {
		t.Fatalf("Failed to reopen vault: %v", err)
	}

	value, err := mgr2.GetSecret("test", "key1")
	if err != nil || value != "value1" {
		t.Errorf("Expected 'value1', got %q err=%v", value, err)
	}

	pass, err := mgr2.GetPassword("host.com", 22, "user")
	if err != nil || pass != "pass123" {
		t.Errorf("Expected 'pass123', got %q err=%v", pass, err)
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key1, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	key2, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if key1 == key2 {
		t.Error("Two random keys should not be equal")
	}

	if len(key1) < 32 {
		t.Errorf("Key should be at least 32 chars, got %d", len(key1))
	}
}
