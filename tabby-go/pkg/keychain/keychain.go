// Package keychain provides OS-native secure credential storage.
//
// On Windows it uses the Credential Manager via zalin/go-winkeyring.
// On macOS it uses the Keychain via zalin/go-winkeyring.
// On Linux it uses the Secret Service (libsecret) via zalin/go-winkeyring.
// Falls back to the encrypted vault if no keyring is available.
package keychain

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/robertpelloni/tabby/tabby-go/pkg/vault"
)

const (
	serviceName = "TabbyGo"
)

// Keychain provides OS-native credential storage with vault fallback.
type Keychain struct {
	mu     sync.RWMutex
	vault  *vault.Manager
	useOS  bool // true if OS keyring is available
}

// NewKeychain creates a new keychain instance.
func NewKeychain(v *vault.Manager) *Keychain {
	k := &Keychain{
		vault: v,
	}

	// Try to detect OS keyring availability
	// The zalin/go-winkeyring package handles this internally,
	// but we check at init time to set the fallback strategy
	k.useOS = true // will be set false on first error

	return k
}

// Store saves a credential to the OS keychain with vault fallback.
func (k *Keychain) Store(key, value string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.useOS {
		err := k.storeOS(key, value)
		if err != nil {
			log.Printf("OS keychain store failed for %s: %v, falling back to vault", key, err)
			k.useOS = false
		} else {
			return nil
		}
	}

	// Fallback to encrypted vault
	return k.vault.SetSecret("keychain", key, value)
}

// Get retrieves a credential from the OS keychain with vault fallback.
func (k *Keychain) Get(key string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.useOS {
		value, err := k.getOS(key)
		if err != nil {
			log.Printf("OS keychain get failed for %s: %v, trying vault", key, err)
			// Try vault before giving up on OS keyring
			v, verr := k.vault.GetSecret("keychain", key)
			if verr == nil {
				return v, nil
			}
			k.useOS = false
			return "", fmt.Errorf("credential not found: %s", key)
		}
		return value, nil
	}

	return k.vault.GetSecret("keychain", key)
}

// Delete removes a credential from the OS keychain with vault fallback.
func (k *Keychain) Delete(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.useOS {
		err := k.deleteOS(key)
		if err != nil {
			log.Printf("OS keychain delete failed for %s: %v, trying vault", key, err)
			k.useOS = false
		} else {
			// Also try to delete from vault in case it was stored there previously
			_ = k.vault.DeleteSecret("keychain", key)
			return nil
		}
	}

	return k.vault.DeleteSecret("keychain", key)
}

// IsOSKeyringAvailable returns whether the OS keyring is currently available.
func (k *Keychain) IsOSKeyringAvailable() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.useOS
}

// SetOSKeyringEnabled enables or disables OS keyring usage.
func (k *Keychain) SetOSKeyringEnabled(enabled bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.useOS = enabled
}

// Platform-specific implementations

func (k *Keychain) storeOS(key, value string) error {
	switch runtime.GOOS {
	case "windows":
		return k.storeWindows(key, value)
	case "darwin":
		return k.storeDarwin(key, value)
	case "linux":
		return k.storeLinux(key, value)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (k *Keychain) getOS(key string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return k.getWindows(key)
	case "darwin":
		return k.getDarwin(key)
	case "linux":
		return k.getLinux(key)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (k *Keychain) deleteOS(key string) error {
	switch runtime.GOOS {
	case "windows":
		return k.deleteWindows(key)
	case "darwin":
		return k.deleteDarwin(key)
	case "linux":
		return k.deleteLinux(key)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Windows implementation using Credential Manager
// These are stub implementations that will be connected to
// the go-winkeyring library when the dependency is added.

func (k *Keychain) storeWindows(key, value string) error {
	// TODO: Use go-winkeyring when available
	// return winkeyring.SetCredential(serviceName, key, value)
	return fmt.Errorf("OS keyring not yet integrated on Windows")
}

func (k *Keychain) getWindows(key string) (string, error) {
	// TODO: Use go-winkeyring when available
	// return winkeyring.GetCredential(serviceName, key)
	return "", fmt.Errorf("OS keyring not yet integrated on Windows")
}

func (k *Keychain) deleteWindows(key string) error {
	// TODO: Use go-winkeyring when available
	// return winkeyring.DeleteCredential(serviceName, key)
	return fmt.Errorf("OS keyring not yet integrated on Windows")
}

// macOS implementation using Keychain

func (k *Keychain) storeDarwin(key, value string) error {
	// TODO: Use go-winkeyring when available
	return fmt.Errorf("OS keyring not yet integrated on macOS")
}

func (k *Keychain) getDarwin(key string) (string, error) {
	return "", fmt.Errorf("OS keyring not yet integrated on macOS")
}

func (k *Keychain) deleteDarwin(key string) error {
	return fmt.Errorf("OS keyring not yet integrated on macOS")
}

// Linux implementation using Secret Service

func (k *Keychain) storeLinux(key, value string) error {
	// TODO: Use go-winkeyring when available
	return fmt.Errorf("OS keyring not yet integrated on Linux")
}

func (k *Keychain) getLinux(key string) (string, error) {
	return "", fmt.Errorf("OS keyring not yet integrated on Linux")
}

func (k *Keychain) deleteLinux(key string) error {
	return fmt.Errorf("OS keyring not yet integrated on Linux")
}
