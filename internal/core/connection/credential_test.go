package connection

import (
	"errors"
	"fmt"
	"testing"
)

func TestSlugify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Todoist", "TODOIST"},
		{"two words", "Work Jira", "WORK_JIRA"},
		{"three words", "Personal Todoist Pro", "PERSONAL_TODOIST_PRO"},
		{"already upper", "GITHUB", "GITHUB"},
		{"lowercase", "github", "GITHUB"},
		{"mixed case", "myTodoist", "MYTODOIST"},
		{"special chars", "work-jira!", "WORK_JIRA"},
		{"multiple spaces", "Work   Jira", "WORK_JIRA"},
		{"leading space", " Jira", "JIRA"},
		{"trailing space", "Jira ", "JIRA"},
		{"numbers", "Jira2", "JIRA2"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseConnID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		connID       string
		wantProvider string
		wantLabel    string
	}{
		{"standard", "todoist:Personal", "todoist", "Personal"},
		{"with spaces", "jira:Work Jira", "jira", "Work Jira"},
		{"no colon", "abc123", "", ""},
		{"empty provider", ":label", "", "label"},
		{"empty label", "provider:", "provider", ""},
		{"multiple colons", "github:My:Project", "github", "My:Project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			provider, label := parseConnID(tt.connID)
			if provider != tt.wantProvider {
				t.Errorf("parseConnID(%q) provider = %q, want %q", tt.connID, provider, tt.wantProvider)
			}
			if label != tt.wantLabel {
				t.Errorf("parseConnID(%q) label = %q, want %q", tt.connID, label, tt.wantLabel)
			}
		})
	}
}

func TestMaskCredential(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"non-empty", "secret-token-123", "••••"},
		{"short", "a", "••••"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MaskCredential(tt.input)
			if got != tt.want {
				t.Errorf("MaskCredential(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConnCredentialKey(t *testing.T) {
	t.Parallel()
	conn := &Connection{ProviderName: "todoist", Label: "Personal"}
	got := ConnCredentialKey(conn)
	if got != "todoist:Personal" {
		t.Errorf("ConnCredentialKey() = %q, want %q", got, "todoist:Personal")
	}
}

func TestEnvCredentialStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("connection-specific env var takes priority", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"THREEDOORS_CONN_WORK_JIRA_TOKEN": "conn-token",
			"THREEDOORS_JIRA_TOKEN":           "provider-token",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("jira:Work Jira")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "conn-token" {
			t.Errorf("Get() = %q, want %q", val, "conn-token")
		}
	})

	t.Run("provider-level env var", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"THREEDOORS_TODOIST_TOKEN": "todoist-token",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("todoist:Personal")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "todoist-token" {
			t.Errorf("Get() = %q, want %q", val, "todoist-token")
		}
	})

	t.Run("github GH_TOKEN alias", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"GH_TOKEN": "gh-token",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("github:OSS")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "gh-token" {
			t.Errorf("Get() = %q, want %q", val, "gh-token")
		}
	})

	t.Run("github GITHUB_TOKEN alias", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"GITHUB_TOKEN": "github-token",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("github:OSS")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "github-token" {
			t.Errorf("Get() = %q, want %q", val, "github-token")
		}
	})

	t.Run("GH_TOKEN preferred over GITHUB_TOKEN", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"GH_TOKEN":     "gh-first",
			"GITHUB_TOKEN": "github-second",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("github:OSS")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "gh-first" {
			t.Errorf("Get() = %q, want %q (GH_TOKEN should take priority)", val, "gh-first")
		}
	})

	t.Run("provider-level preferred over alias", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"THREEDOORS_GITHUB_TOKEN": "threedoors-token",
			"GH_TOKEN":                "gh-token",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		val, err := store.Get("github:OSS")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "threedoors-token" {
			t.Errorf("Get() = %q, want %q (provider-level should beat alias)", val, "threedoors-token")
		}
	})

	t.Run("not found returns ErrCredentialNotFound", func(t *testing.T) {
		t.Parallel()
		store := newEnvCredentialStoreWithLookup(makeLookup(nil))

		_, err := store.Get("todoist:Personal")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Get() error = %v, want ErrCredentialNotFound", err)
		}
	})

	t.Run("empty env var value is treated as not found", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"THREEDOORS_TODOIST_TOKEN": "",
		}
		store := newEnvCredentialStoreWithLookup(makeLookup(env))

		_, err := store.Get("todoist:Personal")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Get() error = %v, want ErrCredentialNotFound for empty value", err)
		}
	})

	t.Run("no colon in connID returns not found", func(t *testing.T) {
		t.Parallel()
		store := newEnvCredentialStoreWithLookup(makeLookup(nil))

		_, err := store.Get("abc123")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Get() error = %v, want ErrCredentialNotFound", err)
		}
	})
}

func TestEnvCredentialStore_SetDelete(t *testing.T) {
	t.Parallel()

	store := NewEnvCredentialStore()

	if err := store.Set("any", "value"); err != nil {
		t.Errorf("Set() should be no-op, got error = %v", err)
	}
	if err := store.Delete("any"); err != nil {
		t.Errorf("Delete() should be no-op, got error = %v", err)
	}
}

// memoryStore is an in-memory CredentialStore for testing.
type memoryStore struct {
	data   map[string]string
	setErr error
	getErr error
	delErr error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{data: make(map[string]string)}
}

func (m *memoryStore) Get(connID string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	val, ok := m.data[connID]
	if !ok {
		return "", fmt.Errorf("memory store %s: %w", connID, ErrCredentialNotFound)
	}
	return val, nil
}

func (m *memoryStore) Set(connID, value string) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[connID] = value
	return nil
}

func (m *memoryStore) Delete(connID string) error {
	if m.delErr != nil {
		return m.delErr
	}
	if _, ok := m.data[connID]; !ok {
		return fmt.Errorf("memory store %s: %w", connID, ErrCredentialNotFound)
	}
	delete(m.data, connID)
	return nil
}

func TestChainCredentialStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("first store wins", func(t *testing.T) {
		t.Parallel()
		s1 := newMemoryStore()
		s1.data["todoist:Personal"] = "first"
		s2 := newMemoryStore()
		s2.data["todoist:Personal"] = "second"

		chain := NewChainCredentialStore(s1, s1, s2)
		val, err := chain.Get("todoist:Personal")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "first" {
			t.Errorf("Get() = %q, want %q", val, "first")
		}
	})

	t.Run("falls through to second store", func(t *testing.T) {
		t.Parallel()
		s1 := newMemoryStore()
		s2 := newMemoryStore()
		s2.data["todoist:Personal"] = "fallback"

		var warnings []string
		chain := NewChainCredentialStore(s1, s1, s2)
		chain.SetWarnFunc(func(msg string) { warnings = append(warnings, msg) })

		val, err := chain.Get("todoist:Personal")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "fallback" {
			t.Errorf("Get() = %q, want %q", val, "fallback")
		}
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning, got %d", len(warnings))
		}
	})

	t.Run("all stores miss returns ErrCredentialNotFound", func(t *testing.T) {
		t.Parallel()
		chain := NewChainCredentialStore(newMemoryStore(), newMemoryStore(), newMemoryStore())

		_, err := chain.Get("missing:key")
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("Get() error = %v, want ErrCredentialNotFound", err)
		}
	})

	t.Run("env var overrides keyring in chain", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"THREEDOORS_TODOIST_TOKEN": "env-token",
		}
		envStore := newEnvCredentialStoreWithLookup(makeLookup(env))
		keyringStore := newMemoryStore()
		keyringStore.data["todoist:Personal"] = "keyring-token"

		chain := NewChainCredentialStore(keyringStore, envStore, keyringStore)
		val, err := chain.Get("todoist:Personal")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "env-token" {
			t.Errorf("Get() = %q, want %q (env should override keyring)", val, "env-token")
		}
	})
}

func TestChainCredentialStore_Set(t *testing.T) {
	t.Parallel()

	t.Run("writes to write store", func(t *testing.T) {
		t.Parallel()
		writeStore := newMemoryStore()
		chain := NewChainCredentialStore(writeStore, newMemoryStore())

		err := chain.Set("todoist:Personal", "my-token")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if writeStore.data["todoist:Personal"] != "my-token" {
			t.Error("Set() did not write to the write store")
		}
	})

	t.Run("no write store returns error", func(t *testing.T) {
		t.Parallel()
		chain := NewChainCredentialStore(nil, newMemoryStore())

		err := chain.Set("id", "val")
		if err == nil {
			t.Error("Set() with nil writeStore should return error")
		}
	})
}

func TestChainCredentialStore_Delete(t *testing.T) {
	t.Parallel()

	t.Run("deletes from write store", func(t *testing.T) {
		t.Parallel()
		writeStore := newMemoryStore()
		writeStore.data["todoist:Personal"] = "token"
		chain := NewChainCredentialStore(writeStore, newMemoryStore())

		err := chain.Delete("todoist:Personal")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, ok := writeStore.data["todoist:Personal"]; ok {
			t.Error("Delete() did not remove from write store")
		}
	})

	t.Run("no write store returns error", func(t *testing.T) {
		t.Parallel()
		chain := NewChainCredentialStore(nil, newMemoryStore())

		err := chain.Delete("id")
		if err == nil {
			t.Error("Delete() with nil writeStore should return error")
		}
	})
}

func TestProviderAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		want     int
	}{
		{"github", 2},
		{"GitHub", 2},
		{"GITHUB", 2},
		{"linear", 1},
		{"Linear", 1},
		{"LINEAR", 1},
		{"todoist", 0},
		{"jira", 0},
		{"", 0},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			t.Parallel()
			got := providerAliases(tt.provider)
			if len(got) != tt.want {
				t.Errorf("providerAliases(%q) returned %d aliases, want %d", tt.provider, len(got), tt.want)
			}
		})
	}
}

// makeLookup creates an env lookup function from a map.
func makeLookup(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		val, ok := env[key]
		return val, ok
	}
}
