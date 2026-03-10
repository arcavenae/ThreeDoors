package connection

import (
	"testing"
)

func TestNewConnection(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		t.Parallel()
		settings := map[string]string{"url": "https://jira.example.com"}
		conn, err := NewConnection("jira", "Work Jira", settings)
		if err != nil {
			t.Fatalf("NewConnection() error = %v", err)
		}
		if conn.ID == "" {
			t.Error("NewConnection() ID is empty")
		}
		if len(conn.ID) != 26 {
			t.Errorf("NewConnection() ID length = %d, want 26 (ULID)", len(conn.ID))
		}
		if conn.ProviderName != "jira" {
			t.Errorf("ProviderName = %q, want %q", conn.ProviderName, "jira")
		}
		if conn.Label != "Work Jira" {
			t.Errorf("Label = %q, want %q", conn.Label, "Work Jira")
		}
		if conn.State != StateDisconnected {
			t.Errorf("State = %s, want %s", conn.State, StateDisconnected)
		}
		if conn.SyncMode != "readonly" {
			t.Errorf("SyncMode = %q, want %q", conn.SyncMode, "readonly")
		}
		if conn.CreatedAt.IsZero() {
			t.Error("CreatedAt is zero")
		}
	})

	t.Run("settings are copied", func(t *testing.T) {
		t.Parallel()
		settings := map[string]string{"key": "value"}
		conn, err := NewConnection("todoist", "Personal", settings)
		if err != nil {
			t.Fatalf("NewConnection() error = %v", err)
		}
		settings["key"] = "modified"
		if conn.Settings["key"] != "value" {
			t.Error("NewConnection() did not copy settings; mutation propagated")
		}
	})

	t.Run("nil settings", func(t *testing.T) {
		t.Parallel()
		conn, err := NewConnection("github", "OSS", nil)
		if err != nil {
			t.Fatalf("NewConnection() error = %v", err)
		}
		if conn.Settings == nil {
			t.Error("NewConnection() Settings is nil, want empty map")
		}
	})

	t.Run("empty provider name", func(t *testing.T) {
		t.Parallel()
		_, err := NewConnection("", "label", nil)
		if err == nil {
			t.Error("NewConnection() with empty provider name should return error")
		}
	})

	t.Run("empty label", func(t *testing.T) {
		t.Parallel()
		_, err := NewConnection("jira", "", nil)
		if err == nil {
			t.Error("NewConnection() with empty label should return error")
		}
	})

	t.Run("unique IDs", func(t *testing.T) {
		t.Parallel()
		ids := make(map[string]bool)
		for range 100 {
			conn, err := NewConnection("test", "test", nil)
			if err != nil {
				t.Fatalf("NewConnection() error = %v", err)
			}
			if ids[conn.ID] {
				t.Fatalf("duplicate ID generated: %s", conn.ID)
			}
			ids[conn.ID] = true
		}
	})
}
