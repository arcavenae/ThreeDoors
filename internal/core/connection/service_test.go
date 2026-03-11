package connection

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// stubCredentialStore is an in-memory CredentialStore for testing.
type stubCredentialStore struct {
	store    map[string]string
	setErr   error
	delErr   error
	setCalls int
	delCalls int
}

func newStubCredentialStore() *stubCredentialStore {
	return &stubCredentialStore{store: make(map[string]string)}
}

func (s *stubCredentialStore) Get(connID string) (string, error) {
	v, ok := s.store[connID]
	if !ok {
		return "", ErrCredentialNotFound
	}
	return v, nil
}

func (s *stubCredentialStore) Set(connID, value string) error {
	s.setCalls++
	if s.setErr != nil {
		return s.setErr
	}
	s.store[connID] = value
	return nil
}

func (s *stubCredentialStore) Delete(connID string) error {
	s.delCalls++
	if s.delErr != nil {
		return s.delErr
	}
	delete(s.store, connID)
	return nil
}

// stubHealthChecker is a HealthChecker for testing.
type stubHealthChecker struct {
	result HealthCheckResult
	err    error
	calls  int
}

func (h *stubHealthChecker) CheckHealth(_ *Connection, _ string) (HealthCheckResult, error) {
	h.calls++
	return h.result, h.err
}

// stubSyncer is a Syncer for testing.
type stubSyncer struct {
	err   error
	calls int
}

func (s *stubSyncer) Sync(_ *Connection, _ string) error {
	s.calls++
	return s.err
}

// newTestService creates a ConnectionService backed by temp files for testing.
func newTestService(t *testing.T) (*ConnectionService, *stubCredentialStore) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	eventLogDir := filepath.Join(tmpDir, "events")

	creds := newStubCredentialStore()
	mgr := NewConnectionManager(nil)
	eventLog := NewSyncEventLog(eventLogDir)

	svc, err := NewConnectionService(ServiceConfig{
		Manager:    mgr,
		Creds:      creds,
		ConfigPath: configPath,
		EventLog:   eventLog,
	})
	if err != nil {
		t.Fatalf("NewConnectionService: %v", err)
	}
	return svc, creds
}

func TestNewConnectionService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     ServiceConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: ServiceConfig{
				Manager:    NewConnectionManager(nil),
				Creds:      newStubCredentialStore(),
				ConfigPath: "/tmp/test.yaml",
			},
			wantErr: false,
		},
		{
			name: "nil manager",
			cfg: ServiceConfig{
				Creds:      newStubCredentialStore(),
				ConfigPath: "/tmp/test.yaml",
			},
			wantErr: true,
		},
		{
			name: "nil creds",
			cfg: ServiceConfig{
				Manager:    NewConnectionManager(nil),
				ConfigPath: "/tmp/test.yaml",
			},
			wantErr: true,
		},
		{
			name: "empty config path",
			cfg: ServiceConfig{
				Manager: NewConnectionManager(nil),
				Creds:   newStubCredentialStore(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewConnectionService(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceAdd(t *testing.T) {
	t.Parallel()

	t.Run("success with credential", func(t *testing.T) {
		t.Parallel()
		svc, creds := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", map[string]string{"project": "inbox"}, "tok_123")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		if conn.ProviderName != "todoist" {
			t.Errorf("ProviderName = %q, want %q", conn.ProviderName, "todoist")
		}
		if conn.Label != "Personal" {
			t.Errorf("Label = %q, want %q", conn.Label, "Personal")
		}
		if conn.State != StateDisconnected {
			t.Errorf("State = %s, want %s", conn.State, StateDisconnected)
		}

		// Credential stored.
		credKey := ConnCredentialKey(conn)
		stored, err := creds.Get(credKey)
		if err != nil {
			t.Fatalf("credential Get: %v", err)
		}
		if stored != "tok_123" {
			t.Errorf("credential = %q, want %q", stored, "tok_123")
		}

		// Manager has the connection.
		got, err := svc.manager.Get(conn.ID)
		if err != nil {
			t.Fatalf("manager Get: %v", err)
		}
		if got.ID != conn.ID {
			t.Errorf("manager connection ID = %q, want %q", got.ID, conn.ID)
		}
	})

	t.Run("success without credential", func(t *testing.T) {
		t.Parallel()
		svc, creds := newTestService(t)

		conn, err := svc.Add("github", "OSS", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		if creds.setCalls != 0 {
			t.Errorf("credential Set called %d times, want 0", creds.setCalls)
		}

		if conn.ProviderName != "github" {
			t.Errorf("ProviderName = %q, want %q", conn.ProviderName, "github")
		}
	})

	t.Run("config persisted", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("jira", "Work", map[string]string{"url": "https://co.atlassian.net"}, "secret")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		// Verify config file was written.
		data, err := os.ReadFile(svc.configPath)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("config file is empty")
		}

		// Config should contain the connection ID.
		if got := string(data); !containsString(got, conn.ID) {
			t.Errorf("config does not contain connection ID %q", conn.ID)
		}
	})

	t.Run("credential store error rolls back", func(t *testing.T) {
		t.Parallel()
		svc, creds := newTestService(t)
		creds.setErr = fmt.Errorf("keyring locked")

		_, err := svc.Add("todoist", "Broken", nil, "tok_bad")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if svc.manager.Count() != 0 {
			t.Errorf("manager count = %d after rollback, want 0", svc.manager.Count())
		}
	})

	t.Run("empty provider name rejected", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		_, err := svc.Add("", "Label", nil, "")
		if err == nil {
			t.Fatal("expected error for empty provider")
		}
	})
}

func TestServiceRemove(t *testing.T) {
	t.Parallel()

	t.Run("remove with keepTasks true", func(t *testing.T) {
		t.Parallel()
		svc, creds := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "tok_rm")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		if err := svc.Remove(conn.ID, true); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		// Connection removed from manager.
		if svc.manager.Count() != 0 {
			t.Errorf("manager count = %d, want 0", svc.manager.Count())
		}

		// Credential deleted.
		credKey := ConnCredentialKey(conn)
		_, err = creds.Get(credKey)
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Errorf("credential should be deleted, got err=%v", err)
		}
	})

	t.Run("remove with keepTasks false", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("github", "OSS", nil, "tok_gh")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		if err := svc.Remove(conn.ID, false); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		if svc.manager.Count() != 0 {
			t.Errorf("manager count = %d, want 0", svc.manager.Count())
		}
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		err := svc.Remove("nonexistent-id", false)
		if err == nil {
			t.Fatal("expected error for nonexistent ID")
		}
	})

	t.Run("config updated after remove", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("jira", "Work", nil, "tok_j")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		if err := svc.Remove(conn.ID, true); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		data, err := os.ReadFile(svc.configPath)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if containsString(string(data), conn.ID) {
			t.Error("config still contains removed connection ID")
		}
	})
}

func TestServicePauseResume(t *testing.T) {
	t.Parallel()

	t.Run("pause connected connection", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		// Transition to Connected: Disconnected → Connecting → Connected.
		if err := svc.manager.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition to Connecting: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition to Connected: %v", err)
		}

		if err := svc.Pause(conn.ID); err != nil {
			t.Fatalf("Pause: %v", err)
		}

		got, _ := svc.manager.Get(conn.ID)
		if got.State != StatePaused {
			t.Errorf("State = %s, want %s", got.State, StatePaused)
		}
	})

	t.Run("resume paused connection", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		// Transition to Paused: Disconnected → Connecting → Connected → Paused.
		if err := svc.manager.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StatePaused); err != nil {
			t.Fatalf("Transition: %v", err)
		}

		if err := svc.Resume(conn.ID); err != nil {
			t.Fatalf("Resume: %v", err)
		}

		got, _ := svc.manager.Get(conn.ID)
		if got.State != StateConnected {
			t.Errorf("State = %s, want %s", got.State, StateConnected)
		}
	})

	t.Run("pause non-connected fails", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		// Connection is Disconnected — cannot pause.
		err = svc.Pause(conn.ID)
		if err == nil {
			t.Fatal("expected error pausing disconnected connection")
		}
	})

	t.Run("resume non-paused fails", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		err = svc.Resume(conn.ID)
		if err == nil {
			t.Fatal("expected error resuming non-paused connection")
		}
	})
}

func TestServiceTestConnection(t *testing.T) {
	t.Parallel()

	t.Run("healthy check", func(t *testing.T) {
		t.Parallel()
		svc, creds := newTestService(t)

		checker := &stubHealthChecker{
			result: HealthCheckResult{
				APIReachable: true,
				TokenValid:   true,
				RateLimitOK:  true,
				TaskCount:    42,
				Details:      map[string]string{"version": "v2"},
			},
		}
		svc.checker = checker

		conn, err := svc.Add("todoist", "Personal", nil, "tok_health")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		result, err := svc.TestConnection(conn.ID)
		if err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		if !result.Healthy() {
			t.Error("expected Healthy() = true")
		}
		if result.TaskCount != 42 {
			t.Errorf("TaskCount = %d, want 42", result.TaskCount)
		}
		if checker.calls != 1 {
			t.Errorf("checker called %d times, want 1", checker.calls)
		}

		// Credential was available to the checker via the store.
		credKey := ConnCredentialKey(conn)
		got, _ := creds.Get(credKey)
		if got != "tok_health" {
			t.Errorf("credential = %q, want %q", got, "tok_health")
		}
	})

	t.Run("unhealthy check", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		checker := &stubHealthChecker{
			result: HealthCheckResult{
				APIReachable: true,
				TokenValid:   false,
				RateLimitOK:  true,
			},
		}
		svc.checker = checker

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		result, err := svc.TestConnection(conn.ID)
		if err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		if result.Healthy() {
			t.Error("expected Healthy() = false with invalid token")
		}
	})

	t.Run("checker error", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		checker := &stubHealthChecker{err: fmt.Errorf("network timeout")}
		svc.checker = checker

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		_, err = svc.TestConnection(conn.ID)
		if err == nil {
			t.Fatal("expected error from checker")
		}
	})

	t.Run("no checker configured", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		_, err = svc.TestConnection(conn.ID)
		if err == nil {
			t.Fatal("expected error with no checker")
		}
	})

	t.Run("nonexistent connection", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)
		svc.checker = &stubHealthChecker{}

		_, err := svc.TestConnection("nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent connection")
		}
	})
}

func TestServiceForceSync(t *testing.T) {
	t.Parallel()

	t.Run("successful sync", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		syncer := &stubSyncer{}
		svc.syncer = syncer

		conn, err := svc.Add("todoist", "Personal", nil, "tok_sync")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		// Transition to Connected.
		if err := svc.manager.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition: %v", err)
		}

		if err := svc.ForceSync(conn.ID); err != nil {
			t.Fatalf("ForceSync: %v", err)
		}

		if syncer.calls != 1 {
			t.Errorf("syncer called %d times, want 1", syncer.calls)
		}

		// Connection should be back in Connected after sync.
		got, _ := svc.manager.Get(conn.ID)
		if got.State != StateConnected {
			t.Errorf("State = %s, want %s", got.State, StateConnected)
		}
	})

	t.Run("sync error transitions to Error state", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		syncer := &stubSyncer{err: fmt.Errorf("API unavailable")}
		svc.syncer = syncer

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StateConnecting); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if err := svc.manager.Transition(conn.ID, StateConnected); err != nil {
			t.Fatalf("Transition: %v", err)
		}

		err = svc.ForceSync(conn.ID)
		if err == nil {
			t.Fatal("expected error from syncer")
		}

		got, _ := svc.manager.Get(conn.ID)
		if got.State != StateError {
			t.Errorf("State = %s, want %s after sync error", got.State, StateError)
		}
		if got.LastError != "API unavailable" {
			t.Errorf("LastError = %q, want %q", got.LastError, "API unavailable")
		}
	})

	t.Run("not connected rejects sync", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)
		svc.syncer = &stubSyncer{}

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		// Connection is Disconnected — cannot sync.
		err = svc.ForceSync(conn.ID)
		if err == nil {
			t.Fatal("expected error syncing disconnected connection")
		}
	})

	t.Run("no syncer configured", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)

		conn, err := svc.Add("todoist", "Personal", nil, "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}

		err = svc.ForceSync(conn.ID)
		if err == nil {
			t.Fatal("expected error with no syncer")
		}
	})

	t.Run("nonexistent connection", func(t *testing.T) {
		t.Parallel()
		svc, _ := newTestService(t)
		svc.syncer = &stubSyncer{}

		err := svc.ForceSync("nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent connection")
		}
	})
}

func TestHealthCheckResultHealthy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result HealthCheckResult
		want   bool
	}{
		{
			name:   "all healthy",
			result: HealthCheckResult{APIReachable: true, TokenValid: true, RateLimitOK: true},
			want:   true,
		},
		{
			name:   "API unreachable",
			result: HealthCheckResult{APIReachable: false, TokenValid: true, RateLimitOK: true},
			want:   false,
		},
		{
			name:   "token invalid",
			result: HealthCheckResult{APIReachable: true, TokenValid: false, RateLimitOK: true},
			want:   false,
		},
		{
			name:   "rate limited",
			result: HealthCheckResult{APIReachable: true, TokenValid: true, RateLimitOK: false},
			want:   false,
		},
		{
			name:   "zero value",
			result: HealthCheckResult{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.Healthy(); got != tt.want {
				t.Errorf("Healthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
