package connection

import (
	"encoding/json"
	"testing"
	"time"
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

func TestConnection_JSON(t *testing.T) {
	t.Parallel()

	conn := &Connection{
		ID:           "01HTEST000000000000000001",
		ProviderName: "jira",
		Label:        "Work Jira",
		State:        StateConnected,
		LastSync:     time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
		SyncMode:     "readonly",
		PollInterval: 5 * time.Minute,
		Settings:     map[string]string{"server": "https://jira.example.com"},
		TaskCount:    42,
		CreatedAt:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(conn)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	checks := []struct {
		key  string
		want interface{}
	}{
		{"id", "01HTEST000000000000000001"},
		{"provider", "jira"},
		{"label", "Work Jira"},
		{"state", "connected"},
		{"sync_mode", "readonly"},
		{"task_count", float64(42)},
	}
	for _, c := range checks {
		if m[c.key] != c.want {
			t.Errorf("json[%q] = %v, want %v", c.key, m[c.key], c.want)
		}
	}

	// Round-trip
	var decoded Connection
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.ID != conn.ID {
		t.Errorf("round-trip ID = %q, want %q", decoded.ID, conn.ID)
	}
	if decoded.State != StateConnected {
		t.Errorf("round-trip State = %v, want connected", decoded.State)
	}
	if decoded.TaskCount != 42 {
		t.Errorf("round-trip TaskCount = %d, want 42", decoded.TaskCount)
	}
}

func TestHealthCheckResult_JSON(t *testing.T) {
	t.Parallel()

	result := HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  false,
		TaskCount:    10,
		Details:      map[string]string{"rate_limit": "exceeded"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	if m["api_reachable"] != true {
		t.Errorf("api_reachable = %v, want true", m["api_reachable"])
	}
	if m["token_valid"] != true {
		t.Errorf("token_valid = %v, want true", m["token_valid"])
	}
	if m["rate_limit_ok"] != false {
		t.Errorf("rate_limit_ok = %v, want false", m["rate_limit_ok"])
	}
	if m["task_count"] != float64(10) {
		t.Errorf("task_count = %v, want 10", m["task_count"])
	}

	details, ok := m["details"].(map[string]interface{})
	if !ok {
		t.Fatalf("details type = %T, want map", m["details"])
	}
	if details["rate_limit"] != "exceeded" {
		t.Errorf("details.rate_limit = %v, want exceeded", details["rate_limit"])
	}

	// Round-trip
	var decoded HealthCheckResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.APIReachable != true {
		t.Errorf("round-trip APIReachable = %v, want true", decoded.APIReachable)
	}
	if decoded.RateLimitOK != false {
		t.Errorf("round-trip RateLimitOK = %v, want false", decoded.RateLimitOK)
	}
}
