package connection

import (
	"errors"
	"testing"

	"github.com/99designs/keyring"
)

// mockKeyring implements keyring.Keyring for testing.
type mockKeyring struct {
	items map[string]keyring.Item
}

func newMockKeyring() *mockKeyring {
	return &mockKeyring{items: make(map[string]keyring.Item)}
}

func (m *mockKeyring) Get(key string) (keyring.Item, error) {
	item, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return item, nil
}

func (m *mockKeyring) Set(item keyring.Item) error {
	m.items[item.Key] = item
	return nil
}

func (m *mockKeyring) Remove(key string) error {
	if _, ok := m.items[key]; !ok {
		return keyring.ErrKeyNotFound
	}
	delete(m.items, key)
	return nil
}

func (m *mockKeyring) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockKeyring) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func TestKeyringCredentialStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		ring := newMockKeyring()
		ring.items["connection:todoist:Personal:token"] = keyring.Item{
			Key:  "connection:todoist:Personal:token",
			Data: []byte("my-secret"),
		}
		store := newKeyringCredentialStoreFromRing(ring)

		val, err := store.Get("todoist:Personal")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "my-secret" {
			t.Errorf("Get() = %q, want %q", val, "my-secret")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		store := newKeyringCredentialStoreFromRing(newMockKeyring())

		_, err := store.Get("missing:key")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Get() error = %v, want ErrCredentialNotFound", err)
		}
	})
}

func TestKeyringCredentialStore_Set(t *testing.T) {
	t.Parallel()

	ring := newMockKeyring()
	store := newKeyringCredentialStoreFromRing(ring)

	err := store.Set("todoist:Personal", "new-token")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	item, ok := ring.items["connection:todoist:Personal:token"]
	if !ok {
		t.Fatal("Set() did not store item in keyring")
	}
	if string(item.Data) != "new-token" {
		t.Errorf("stored data = %q, want %q", string(item.Data), "new-token")
	}
}

func TestKeyringCredentialStore_Delete(t *testing.T) {
	t.Parallel()

	t.Run("existing key", func(t *testing.T) {
		t.Parallel()
		ring := newMockKeyring()
		ring.items["connection:jira:Work:token"] = keyring.Item{
			Key:  "connection:jira:Work:token",
			Data: []byte("token"),
		}
		store := newKeyringCredentialStoreFromRing(ring)

		err := store.Delete("jira:Work")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, ok := ring.items["connection:jira:Work:token"]; ok {
			t.Error("Delete() did not remove item from keyring")
		}
	})

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()
		store := newKeyringCredentialStoreFromRing(newMockKeyring())

		err := store.Delete("missing:key")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Delete() error = %v, want ErrCredentialNotFound", err)
		}
	})
}

func TestKeyringKey(t *testing.T) {
	t.Parallel()
	got := keyringKey("todoist:Personal")
	want := "connection:todoist:Personal:token"
	if got != want {
		t.Errorf("keyringKey() = %q, want %q", got, want)
	}
}
