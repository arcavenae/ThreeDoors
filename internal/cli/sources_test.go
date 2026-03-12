package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
)

// newTestManager creates a ConnectionManager with test connections.
func newTestManager(t *testing.T) *connection.ConnectionManager {
	t.Helper()
	m := connection.NewConnectionManager(nil)
	conn1, err := m.Add("jira", "Work Jira", map[string]string{
		"server": "https://jira.example.com",
		"filter": "project=WORK",
	})
	if err != nil {
		t.Fatalf("add conn1: %v", err)
	}
	conn1.TaskCount = 42
	conn1.LastSync = time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC)

	conn2, err := m.Add("todoist", "Personal Todoist", map[string]string{})
	if err != nil {
		t.Fatalf("add conn2: %v", err)
	}
	conn2.TaskCount = 7

	return m
}

func TestRunSourcesList(t *testing.T) {
	t.Parallel()

	t.Run("table output with connections", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{})

		// Override runSourcesList to write to our buffer
		err := runSourcesListTo(cmd, manager, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesList: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "NAME") {
			t.Error("missing table header NAME")
		}
		if !strings.Contains(out, "PROVIDER") {
			t.Error("missing table header PROVIDER")
		}
		if !strings.Contains(out, "STATUS") {
			t.Error("missing table header STATUS")
		}
		if !strings.Contains(out, "LAST SYNC") {
			t.Error("missing table header LAST SYNC")
		}
		if !strings.Contains(out, "TASKS") {
			t.Error("missing table header TASKS")
		}
		if !strings.Contains(out, "Personal Todoist") {
			t.Error("missing connection: Personal Todoist")
		}
		if !strings.Contains(out, "Work Jira") {
			t.Error("missing connection: Work Jira")
		}
		if !strings.Contains(out, "jira") {
			t.Error("missing provider: jira")
		}
		if !strings.Contains(out, "42") {
			t.Error("missing task count: 42")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesListTo(cmd, manager, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesList: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "No connections configured") {
			t.Errorf("expected 'No connections configured', got %q", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesListTo(cmd, manager, &buf, true)
		if err != nil {
			t.Fatalf("runSourcesList: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources" {
			t.Errorf("command = %q, want %q", env.Command, "sources")
		}

		// Data should be a list
		items, ok := env.Data.([]interface{})
		if !ok {
			t.Fatalf("data type = %T, want []interface{}", env.Data)
		}
		if len(items) != 2 {
			t.Errorf("len(data) = %d, want 2", len(items))
		}
	})
}

func TestRunSourcesStatus(t *testing.T) {
	t.Parallel()

	t.Run("existing connection", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesStatusTo(cmd, manager, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesStatus: %v", err)
		}

		out := buf.String()
		for _, want := range []string{
			"Name:          Work Jira",
			"Provider:      jira",
			"Status:        disconnected",
			"Server:        https://jira.example.com",
			"Tasks Active:  42",
			"Filter:        project=WORK",
			"Sync Mode:     readonly",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("missing %q in output:\n%s", want, out)
			}
		}
	})

	t.Run("connection not found", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesStatusTo(cmd, manager, "Nonexistent", &buf, false)
		if err == nil {
			t.Fatal("expected error for nonexistent connection")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error = %v, want contains 'not found'", err)
		}
	})

	t.Run("json output", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesStatusTo(cmd, manager, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesStatus: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources status" {
			t.Errorf("command = %q, want %q", env.Command, "sources status")
		}

		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["name"] != "Work Jira" {
			t.Errorf("name = %v, want Work Jira", data["name"])
		}
		if data["provider"] != "jira" {
			t.Errorf("provider = %v, want jira", data["provider"])
		}
	})

	t.Run("json not found", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		err := runSourcesStatusTo(cmd, manager, "Nope", &buf, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Error == nil {
			t.Fatal("expected JSON error envelope")
		}
		if env.Error.Code != ExitNotFound {
			t.Errorf("error code = %d, want %d", env.Error.Code, ExitNotFound)
		}
	})
}

func TestRunSourcesTest(t *testing.T) {
	t.Parallel()

	t.Run("healthy connection", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)
		conn, _ := manager.GetByLabel("Work Jira")

		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
			TaskCount:    42,
		}}

		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		exitCode := runSourcesTestTo(cmd, manager, svc, conn.Label, &buf, false)

		out := buf.String()
		if !strings.Contains(out, "Health check: Work Jira") {
			t.Errorf("missing health check header in output:\n%s", out)
		}
		if !strings.Contains(out, "✓ DNS resolution") {
			t.Errorf("missing DNS check in output:\n%s", out)
		}
		if !strings.Contains(out, "✓ Authentication") {
			t.Errorf("missing auth check in output:\n%s", out)
		}
		if exitCode != 0 {
			t.Errorf("exit code = %d, want 0", exitCode)
		}
	})

	t.Run("unhealthy connection", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)
		conn, _ := manager.GetByLabel("Work Jira")

		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: false,
			TokenValid:   true,
			RateLimitOK:  true,
		}}

		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		exitCode := runSourcesTestTo(cmd, manager, svc, conn.Label, &buf, false)

		out := buf.String()
		if !strings.Contains(out, "✗ DNS resolution") {
			t.Errorf("missing failed DNS check in output:\n%s", out)
		}
		if exitCode != 1 {
			t.Errorf("exit code = %d, want 1", exitCode)
		}
	})

	t.Run("connection not found", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		exitCode := runSourcesTestTo(cmd, manager, nil, "Nonexistent", &buf, false)
		if exitCode != ExitNotFound {
			t.Errorf("exit code = %d, want %d", exitCode, ExitNotFound)
		}
	})

	t.Run("json healthy", func(t *testing.T) {
		t.Parallel()
		manager := newTestManager(t)
		conn, _ := manager.GetByLabel("Work Jira")

		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
		}}

		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		cmd := newSourcesCmd()
		exitCode := runSourcesTestTo(cmd, manager, svc, conn.Label, &buf, true)

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources test" {
			t.Errorf("command = %q, want %q", env.Command, "sources test")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["healthy"] != true {
			t.Errorf("healthy = %v, want true", data["healthy"])
		}
		if exitCode != 0 {
			t.Errorf("exit code = %d, want 0", exitCode)
		}
	})
}

func TestFormatSyncTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"zero time", time.Time{}, "never"},
		{"specific time", time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC), "2026-03-10 14:30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatSyncTime(tt.time)
			if got != tt.want {
				t.Errorf("formatSyncTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetByLabel(t *testing.T) {
	t.Parallel()

	manager := newTestManager(t)

	t.Run("exact match", func(t *testing.T) {
		t.Parallel()
		conn, err := manager.GetByLabel("Work Jira")
		if err != nil {
			t.Fatalf("GetByLabel: %v", err)
		}
		if conn.Label != "Work Jira" {
			t.Errorf("label = %q, want %q", conn.Label, "Work Jira")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		t.Parallel()
		conn, err := manager.GetByLabel("work jira")
		if err != nil {
			t.Fatalf("GetByLabel: %v", err)
		}
		if conn.Label != "Work Jira" {
			t.Errorf("label = %q, want %q", conn.Label, "Work Jira")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := manager.GetByLabel("Nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent label")
		}
	})
}

// stubHealthChecker implements connection.HealthChecker for tests.
type stubHealthChecker struct {
	result connection.HealthCheckResult
	err    error
}

func (s *stubHealthChecker) CheckHealth(_ *connection.Connection, _ string) (connection.HealthCheckResult, error) {
	return s.result, s.err
}

// stubCredentialStore implements connection.CredentialStore for tests.
type stubCredentialStore struct{}

func (s *stubCredentialStore) Get(_ string) (string, error) { return "", nil }
func (s *stubCredentialStore) Set(_, _ string) error        { return nil }
func (s *stubCredentialStore) Delete(_ string) error        { return nil }

func newTestService(t *testing.T, manager *connection.ConnectionManager, checker connection.HealthChecker) *connection.ConnectionService {
	t.Helper()

	dir := t.TempDir()
	svc, err := connection.NewConnectionService(connection.ServiceConfig{
		Manager:    manager,
		Creds:      &stubCredentialStore{},
		ConfigPath: dir + "/config.yaml",
		Checker:    checker,
	})
	if err != nil {
		t.Fatalf("NewConnectionService: %v", err)
	}
	return svc
}
