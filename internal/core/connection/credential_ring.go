package connection

import (
	"fmt"

	"github.com/99designs/keyring"
)

// KeyringCredentialStore stores credentials in the system keychain
// via 99designs/keyring. Service name is "threedoors", keys are
// "connection:<connID>:token".
type KeyringCredentialStore struct {
	ring keyring.Keyring
}

// NewKeyringCredentialStore opens or creates a keyring with the given backend config.
// If the system keychain is unavailable, returns an error — callers should fall back
// to the encrypted file store.
func NewKeyringCredentialStore(backends []keyring.BackendType) (*KeyringCredentialStore, error) {
	if len(backends) == 0 {
		backends = keyring.AvailableBackends()
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:                    "threedoors",
		AllowedBackends:                backends,
		KeychainTrustApplication:       true,
		FileDir:                        "~/.threedoors/keyring",
		FilePasswordFunc:               keyring.TerminalPrompt,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}
	return &KeyringCredentialStore{ring: ring}, nil
}

// newKeyringCredentialStoreFromRing creates a KeyringCredentialStore with an
// injected keyring for testing.
func newKeyringCredentialStoreFromRing(ring keyring.Keyring) *KeyringCredentialStore {
	return &KeyringCredentialStore{ring: ring}
}

func keyringKey(connID string) string {
	return "connection:" + connID + ":token"
}

// Get retrieves a credential from the system keychain.
func (k *KeyringCredentialStore) Get(connID string) (string, error) {
	item, err := k.ring.Get(keyringKey(connID))
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("keyring credential %s: %w", connID, ErrCredentialNotFound)
		}
		return "", fmt.Errorf("keyring get %s: %w", connID, err)
	}
	return string(item.Data), nil
}

// Set stores a credential in the system keychain.
func (k *KeyringCredentialStore) Set(connID, value string) error {
	err := k.ring.Set(keyring.Item{
		Key:  keyringKey(connID),
		Data: []byte(value),
	})
	if err != nil {
		return fmt.Errorf("keyring set %s: %w", connID, err)
	}
	return nil
}

// Delete removes a credential from the system keychain.
func (k *KeyringCredentialStore) Delete(connID string) error {
	err := k.ring.Remove(keyringKey(connID))
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return fmt.Errorf("keyring credential %s: %w", connID, ErrCredentialNotFound)
		}
		return fmt.Errorf("keyring delete %s: %w", connID, err)
	}
	return nil
}
