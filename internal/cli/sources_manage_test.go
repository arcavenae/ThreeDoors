package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
)

// newConnectedManager returns a ConnectionManager with a single connection
// in Connected state, ready for pause/sync operations.
func newConnectedManager(t *testing.T) *connection.ConnectionManager {
	t.Helper()
	m := connection.NewConnectionManager(nil)
	conn, err := m.Add("jira", "Work Jira", map[string]string{
		"server": "https://jira.example.com",
	})
	if err != nil {
		t.Fatalf("add conn: %v", err)
	}
	// Transition to connected: Disconnected → Connecting → Connected
	if err := m.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatalf("transition connecting: %v", err)
	}
	if err := m.Transition(conn.ID, connection.StateConnected); err != nil {
		t.Fatalf("transition connected: %v", err)
	}
	return m
}

// newPausedManager returns a ConnectionManager with a single paused connection.
func newPausedManager(t *testing.T) *connection.ConnectionManager {
	t.Helper()
	m := newConnectedManager(t)
	conn, _ := m.GetByLabel("Work Jira")
	if err := m.Transition(conn.ID, connection.StatePaused); err != nil {
		t.Fatalf("transition paused: %v", err)
	}
	return m
}

// newAuthExpiredManager returns a ConnectionManager with a connection in AuthExpired state.
func newAuthExpiredManager(t *testing.T) *connection.ConnectionManager {
	t.Helper()
	m := connection.NewConnectionManager(nil)
	conn, err := m.Add("jira", "Work Jira", map[string]string{})
	if err != nil {
		t.Fatalf("add conn: %v", err)
	}
	if err := m.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatalf("transition connecting: %v", err)
	}
	if err := m.Transition(conn.ID, connection.StateAuthExpired); err != nil {
		t.Fatalf("transition auth_expired: %v", err)
	}
	return m
}

func TestRunSourcesPause(t *testing.T) {
	t.Parallel()

	t.Run("pause connected", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, svc, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesPause: %v", err)
		}
		if !strings.Contains(buf.String(), "Paused sync for") {
			t.Errorf("output = %q, want contains 'Paused sync for'", buf.String())
		}

		conn, _ := manager.GetByLabel("Work Jira")
		if conn.State != connection.StatePaused {
			t.Errorf("state = %s, want paused", conn.State)
		}
	})

	t.Run("pause already paused", func(t *testing.T) {
		t.Parallel()
		manager := newPausedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, svc, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for already paused connection")
		}
		if !strings.Contains(err.Error(), "cannot pause") {
			t.Errorf("error = %v, want contains 'cannot pause'", err)
		}
	})

	t.Run("pause not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, svc, "Nonexistent", &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error = %v, want contains 'not found'", err)
		}
	})

	t.Run("pause json output", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, svc, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesPause: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources pause" {
			t.Errorf("command = %q, want %q", env.Command, "sources pause")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["action"] != "pause" {
			t.Errorf("action = %v, want pause", data["action"])
		}
		if data["status"] != "paused" {
			t.Errorf("status = %v, want paused", data["status"])
		}
	})

	t.Run("pause json not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, svc, "Nope", &buf, true)
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

	t.Run("pause no service", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)

		var buf bytes.Buffer
		err := runSourcesPauseTo(manager, nil, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for nil service")
		}
		if !strings.Contains(err.Error(), "no connection service") {
			t.Errorf("error = %v, want contains 'no connection service'", err)
		}
	})
}

func TestRunSourcesResume(t *testing.T) {
	t.Parallel()

	t.Run("resume paused", func(t *testing.T) {
		t.Parallel()
		manager := newPausedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesResumeTo(manager, svc, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesResume: %v", err)
		}
		if !strings.Contains(buf.String(), "Resumed sync for") {
			t.Errorf("output = %q, want contains 'Resumed sync for'", buf.String())
		}

		conn, _ := manager.GetByLabel("Work Jira")
		if conn.State != connection.StateConnected {
			t.Errorf("state = %s, want connected", conn.State)
		}
	})

	t.Run("resume not paused", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesResumeTo(manager, svc, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for non-paused connection")
		}
		if !strings.Contains(err.Error(), "cannot resume") {
			t.Errorf("error = %v, want contains 'cannot resume'", err)
		}
	})

	t.Run("resume json output", func(t *testing.T) {
		t.Parallel()
		manager := newPausedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesResumeTo(manager, svc, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesResume: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources resume" {
			t.Errorf("command = %q, want %q", env.Command, "sources resume")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["action"] != "resume" {
			t.Errorf("action = %v, want resume", data["action"])
		}
	})

	t.Run("resume not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesResumeTo(manager, svc, "Missing", &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
	})
}

// stubSyncer implements connection.Syncer for tests.
type stubSyncer struct {
	err error
}

func (s *stubSyncer) Sync(_ *connection.Connection, _ string) error {
	return s.err
}

func newTestServiceWithSyncer(t *testing.T, manager *connection.ConnectionManager, checker connection.HealthChecker, syncer connection.Syncer) *connection.ConnectionService {
	t.Helper()
	dir := t.TempDir()
	svc, err := connection.NewConnectionService(connection.ServiceConfig{
		Manager:    manager,
		Creds:      &stubCredentialStore{},
		ConfigPath: dir + "/config.yaml",
		Checker:    checker,
		Syncer:     syncer,
	})
	if err != nil {
		t.Fatalf("NewConnectionService: %v", err)
	}
	return svc
}

func TestRunSourcesSync(t *testing.T) {
	t.Parallel()

	t.Run("sync connected", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		syncer := &stubSyncer{}
		svc := newTestServiceWithSyncer(t, manager, nil, syncer)

		var buf bytes.Buffer
		err := runSourcesSyncTo(manager, svc, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesSync: %v", err)
		}
		if !strings.Contains(buf.String(), "Sync complete for") {
			t.Errorf("output = %q, want contains 'Sync complete for'", buf.String())
		}
	})

	t.Run("sync not connected", func(t *testing.T) {
		t.Parallel()
		manager := newPausedManager(t)
		syncer := &stubSyncer{}
		svc := newTestServiceWithSyncer(t, manager, nil, syncer)

		var buf bytes.Buffer
		err := runSourcesSyncTo(manager, svc, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for paused connection")
		}
		if !strings.Contains(err.Error(), "sync failed") {
			t.Errorf("error = %v, want contains 'sync failed'", err)
		}
	})

	t.Run("sync json output", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		syncer := &stubSyncer{}
		svc := newTestServiceWithSyncer(t, manager, nil, syncer)

		var buf bytes.Buffer
		err := runSourcesSyncTo(manager, svc, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesSync: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources sync" {
			t.Errorf("command = %q, want %q", env.Command, "sources sync")
		}
	})

	t.Run("sync not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)

		var buf bytes.Buffer
		err := runSourcesSyncTo(manager, nil, "Nope", &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
	})
}

func TestRunSourcesDisconnect(t *testing.T) {
	t.Parallel()

	t.Run("disconnect with keep-tasks flag", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Work Jira", true, true, nil, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesDisconnect: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "Disconnected") {
			t.Errorf("output = %q, want contains 'Disconnected'", out)
		}
		if !strings.Contains(out, "preserved") {
			t.Errorf("output = %q, want contains 'preserved'", out)
		}
	})

	t.Run("disconnect without keep-tasks flag", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Work Jira", false, true, nil, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesDisconnect: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "removed") {
			t.Errorf("output = %q, want contains 'removed'", out)
		}
	})

	t.Run("disconnect interactive yes", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		reader := strings.NewReader("y\n")
		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Work Jira", false, false, reader, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesDisconnect: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "preserved") {
			t.Errorf("output = %q, want contains 'preserved' (user answered y)", out)
		}
	})

	t.Run("disconnect interactive no", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		reader := strings.NewReader("n\n")
		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Work Jira", false, false, reader, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesDisconnect: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "removed") {
			t.Errorf("output = %q, want contains 'removed' (user answered n)", out)
		}
	})

	t.Run("disconnect json with keep-tasks", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Work Jira", true, true, nil, &buf, true)
		if err != nil {
			t.Fatalf("runSourcesDisconnect: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources disconnect" {
			t.Errorf("command = %q, want %q", env.Command, "sources disconnect")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["action"] != "disconnect" {
			t.Errorf("action = %v, want disconnect", data["action"])
		}
		status, _ := data["status"].(string)
		if !strings.Contains(status, "preserved") {
			t.Errorf("status = %q, want contains 'preserved'", status)
		}
	})

	t.Run("disconnect not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesDisconnectTo(manager, svc, "Nope", false, true, nil, &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
	})
}

func TestRunSourcesReauth(t *testing.T) {
	t.Parallel()

	t.Run("reauth auth_expired success", func(t *testing.T) {
		t.Parallel()
		manager := newAuthExpiredManager(t)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
		}}
		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		err := runSourcesReauthTo(manager, svc, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesReauth: %v", err)
		}
		if !strings.Contains(buf.String(), "Re-authenticated") {
			t.Errorf("output = %q, want contains 'Re-authenticated'", buf.String())
		}

		conn, _ := manager.GetByLabel("Work Jira")
		if conn.State != connection.StateConnected {
			t.Errorf("state = %s, want connected", conn.State)
		}
	})

	t.Run("reauth connected state rejected", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesReauthTo(manager, svc, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for connected connection")
		}
		if !strings.Contains(err.Error(), "must be auth_expired or error") {
			t.Errorf("error = %v, want contains 'must be auth_expired or error'", err)
		}
	})

	t.Run("reauth health check fails", func(t *testing.T) {
		t.Parallel()
		manager := newAuthExpiredManager(t)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: false,
			TokenValid:   false,
			RateLimitOK:  true,
		}}
		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		err := runSourcesReauthTo(manager, svc, "Work Jira", &buf, false)
		if err == nil {
			t.Fatal("expected error for failed health check")
		}
		if !strings.Contains(err.Error(), "re-authentication failed") {
			t.Errorf("error = %v, want contains 're-authentication failed'", err)
		}

		conn, _ := manager.GetByLabel("Work Jira")
		if conn.State != connection.StateError {
			t.Errorf("state = %s, want error", conn.State)
		}
	})

	t.Run("reauth json output", func(t *testing.T) {
		t.Parallel()
		manager := newAuthExpiredManager(t)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
		}}
		svc := newTestService(t, manager, checker)

		var buf bytes.Buffer
		err := runSourcesReauthTo(manager, svc, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesReauth: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources reauth" {
			t.Errorf("command = %q, want %q", env.Command, "sources reauth")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["action"] != "reauth" {
			t.Errorf("action = %v, want reauth", data["action"])
		}
	})

	t.Run("reauth not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)

		var buf bytes.Buffer
		err := runSourcesReauthTo(manager, svc, "Nope", &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
	})
}

func TestRunSourcesEdit(t *testing.T) {
	t.Parallel()

	t.Run("edit shows reconnect instructions", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)

		var buf bytes.Buffer
		err := runSourcesEditTo(manager, "Work Jira", &buf, false)
		if err != nil {
			t.Fatalf("runSourcesEdit: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "disconnect") {
			t.Errorf("output = %q, want contains 'disconnect'", out)
		}
		if !strings.Contains(out, "threedoors connect jira") {
			t.Errorf("output = %q, want contains 'threedoors connect jira'", out)
		}
	})

	t.Run("edit json output", func(t *testing.T) {
		t.Parallel()
		manager := newConnectedManager(t)

		var buf bytes.Buffer
		err := runSourcesEditTo(manager, "Work Jira", &buf, true)
		if err != nil {
			t.Fatalf("runSourcesEdit: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources edit" {
			t.Errorf("command = %q, want %q", env.Command, "sources edit")
		}
		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["provider"] != "jira" {
			t.Errorf("provider = %v, want jira", data["provider"])
		}
		if data["command"] == nil || data["command"] == "" {
			t.Error("expected non-empty command field")
		}
	})

	t.Run("edit not found", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)

		var buf bytes.Buffer
		err := runSourcesEditTo(manager, "Nope", &buf, false)
		if err == nil {
			t.Fatal("expected error for missing connection")
		}
	})
}

func TestStateHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		state  connection.ConnectionState
		action string
		want   string
	}{
		{"pause already paused", connection.StatePaused, "pause", "already paused"},
		{"pause disconnected", connection.StateDisconnected, "pause", "must be connected"},
		{"resume already connected", connection.StateConnected, "resume", "already active"},
		{"resume disconnected", connection.StateDisconnected, "resume", "must be paused"},
		{"sync paused", connection.StatePaused, "sync", "must be connected"},
		{"unknown action", connection.StateConnected, "unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stateHint(tt.state, tt.action)
			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Errorf("stateHint(%s, %q) = %q, want contains %q", tt.state, tt.action, got, tt.want)
			}
			if tt.want == "" && got != "" {
				t.Errorf("stateHint(%s, %q) = %q, want empty", tt.state, tt.action, got)
			}
		})
	}
}
