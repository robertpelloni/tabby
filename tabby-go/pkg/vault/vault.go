// Package vault provides encrypted credential storage for Tabby's Go backend.
//
// It implements PBKDF2 key derivation and AES-256-CBC encryption, mirroring
// the TypeScript VaultService for compatibility. Vault files can be shared
// between the Go backend and the TypeScript frontend.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 100000
	pbkdf2SaltLength = 8  // 64 bits
	aesKeyLength     = 32 // 256 bits
	aesIVLength      = 16 // 128 bits
	hmacLength       = 32
)

// StoredVault represents the on-disk vault format
type StoredVault struct {
	Version  int    `json:"version"`
	Contents string `json:"contents"`
	KeySalt  string `json:"keySalt"`
	IV       string `json:"iv"`
	HMAC     string `json:"hmac,omitempty"`
}

// VaultSecret represents a stored secret
type VaultSecret struct {
	Type  string      `json:"type"`
	Key   interface{} `json:"key"`
	Value string      `json:"value"`
}

// FileSecretKey is the key type for file-type secrets
type FileSecretKey struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Vault represents the decrypted vault contents
type Vault struct {
	Config  interface{}   `json:"config"`
	Secrets []VaultSecret `json:"secrets"`
}

// Manager manages the encrypted credential vault
type Manager struct {
	mu         sync.RWMutex
	vault      *Vault
	filePath   string
	key        []byte
	passphrase string
	open       bool
}

// NewManager creates a new vault manager
func NewManager() *Manager {
	return &Manager{}
}

// Create creates a new encrypted vault file
func (m *Manager) Create(path, passphrase string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	salt := make([]byte, pbkdf2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, aesIVLength)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	key := pbkdf2.Key([]byte(passphrase), salt, pbkdf2Iterations, aesKeyLength, sha512.New)

	vault := &Vault{
		Config:  map[string]interface{}{},
		Secrets: []VaultSecret{},
	}

	m.vault = vault
	m.key = key
	m.passphrase = passphrase
	m.filePath = path
	m.open = true

	return m.save(salt, iv)
}

// Open opens an existing vault with a passphrase
func (m *Manager) Open(path, passphrase string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read vault: %w", err)
	}

	var stored StoredVault
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("failed to parse vault: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(stored.KeySalt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(stored.IV)
	if err != nil {
		return fmt.Errorf("failed to decode IV: %w", err)
	}

	key := pbkdf2.Key([]byte(passphrase), salt, pbkdf2Iterations, aesKeyLength, sha512.New)

	// Verify HMAC if present
	if stored.HMAC != "" {
		if !m.verifyHMAC(key, stored.Contents, stored.HMAC) {
			return fmt.Errorf("invalid passphrase or corrupted vault")
		}
	}

	contents, err := base64.StdEncoding.DecodeString(stored.Contents)
	if err != nil {
		return fmt.Errorf("failed to decode contents: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(contents))
	stream.CryptBlocks(decrypted, contents)

	// Remove PKCS7 padding
	if len(decrypted) > 0 {
		padLen := int(decrypted[len(decrypted)-1])
		if padLen > 0 && padLen <= aes.BlockSize {
			decrypted = decrypted[:len(decrypted)-padLen]
		}
	}

	var vault Vault
	if err := json.Unmarshal(decrypted, &vault); err != nil {
		return fmt.Errorf("failed to unmarshal vault (wrong passphrase?): %w", err)
	}

	m.vault = &vault
	m.key = key
	m.passphrase = passphrase
	m.filePath = path
	m.open = true

	return nil
}

// Close closes the vault, clearing the key from memory
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear the key from memory
	for i := range m.key {
		m.key[i] = 0
	}
	m.key = nil
	m.passphrase = ""
	m.vault = nil
	m.open = false
}

// IsOpen returns whether the vault is open
func (m *Manager) IsOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.open
}

// GetSecret retrieves a secret by type and key
func (m *Manager) GetSecret(secretType string, key interface{}) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.open {
		return "", fmt.Errorf("vault is not open")
	}

	for _, s := range m.vault.Secrets {
		if s.Type == secretType {
			// Compare keys
			if keysMatch(s.Key, key) {
				return s.Value, nil
			}
		}
	}
	return "", fmt.Errorf("secret not found")
}

// SetSecret stores a secret
func (m *Manager) SetSecret(secretType string, key interface{}, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.open {
		return fmt.Errorf("vault is not open")
	}

	// Update existing or add new
	for i, s := range m.vault.Secrets {
		if s.Type == secretType && keysMatch(s.Key, key) {
			m.vault.Secrets[i].Value = value
			return m.persist()
		}
	}

	m.vault.Secrets = append(m.vault.Secrets, VaultSecret{
		Type:  secretType,
		Key:   key,
		Value: value,
	})

	return m.persist()
}

// DeleteSecret removes a secret
func (m *Manager) DeleteSecret(secretType string, key interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.open {
		return fmt.Errorf("vault is not open")
	}

	for i, s := range m.vault.Secrets {
		if s.Type == secretType && keysMatch(s.Key, key) {
			m.vault.Secrets = append(m.vault.Secrets[:i], m.vault.Secrets[i+1:]...)
			return m.persist()
		}
	}

	return fmt.Errorf("secret not found")
}

// ListSecrets returns all secrets of a given type
func (m *Manager) ListSecrets(secretType string) ([]VaultSecret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.open {
		return nil, fmt.Errorf("vault is not open")
	}

	var result []VaultSecret
	for _, s := range m.vault.Secrets {
		if s.Type == secretType {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetPassword is a convenience method to get an SSH password
func (m *Manager) GetPassword(host string, port int, user string) (string, error) {
	return m.GetSecret("ssh-password", map[string]interface{}{
		"host": host,
		"port": port,
		"user": user,
	})
}

// SetPassword is a convenience method to store an SSH password
func (m *Manager) SetPassword(host string, port int, user, password string) error {
	return m.SetSecret("ssh-password", map[string]interface{}{
		"host": host,
		"port": port,
		"user": user,
	}, password)
}

// GetPrivateKeyPassphrase retrieves a private key passphrase
func (m *Manager) GetPrivateKeyPassphrase(keyHash string) (string, error) {
	return m.GetSecret("private-key-passphrase", map[string]interface{}{
		"keyHash": keyHash,
	})
}

// SetPrivateKeyPassphrase stores a private key passphrase
func (m *Manager) SetPrivateKeyPassphrase(keyHash, passphrase string) error {
	return m.SetSecret("private-key-passphrase", map[string]interface{}{
		"keyHash": keyHash,
	}, passphrase)
}

// ---- Internal methods ----

// persist saves the vault to disk
func (m *Manager) persist() error {
	salt := make([]byte, pbkdf2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	iv := make([]byte, aesIVLength)
	if _, err := rand.Read(iv); err != nil {
		return err
	}

	// Re-derive key from passphrase with new salt
	m.key = pbkdf2.Key([]byte(m.passphrase), salt, pbkdf2Iterations, aesKeyLength, sha512.New)

	return m.saveWithKey(salt, iv)
}

// save encrypts and saves the vault
func (m *Manager) save(salt, iv []byte) error {
	return m.saveWithKey(salt, iv)
}

// saveWithKey encrypts and saves the vault using the current key
func (m *Manager) saveWithKey(salt, iv []byte) error {
	plaintext, err := json.Marshal(m.vault)
	if err != nil {
		return fmt.Errorf("failed to marshal vault: %w", err)
	}

	// PKCS7 padding
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext := make([]byte, len(padded))
	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(ciphertext, padded)

	contents := base64.StdEncoding.EncodeToString(ciphertext)
	hmacStr := m.computeHMAC(m.key, contents)

	stored := StoredVault{
		Version:  1,
		Contents: contents,
		KeySalt:  base64.StdEncoding.EncodeToString(salt),
		IV:       base64.StdEncoding.EncodeToString(iv),
		HMAC:     hmacStr,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stored vault: %w", err)
	}

	if m.filePath == "" {
		return fmt.Errorf("no file path set")
	}

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0600)
}

// computeHMAC computes an HMAC-SHA256 for integrity verification
func (m *Manager) computeHMAC(key []byte, data string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// verifyHMAC verifies the HMAC of the vault contents
func (m *Manager) verifyHMAC(key []byte, data, expectedHMAC string) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	computed := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(expectedHMAC))
}

// keysMatch compares two secret keys
func keysMatch(a, b interface{}) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aj) == string(bj)
}

// GenerateRandomKey generates a random key for use with the vault
func GenerateRandomKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
