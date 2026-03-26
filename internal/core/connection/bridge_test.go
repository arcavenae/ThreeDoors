package connection

import (
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// stubTaskProvider is a minimal TaskProvider for bridge tests.
type stubTaskProvider struct {
	name        string
	tasks       []*core.Task
	loadErr     error
	healthItems []core.HealthCheckItem
	loadCalls   int
}

func (s *stubTaskProvider) Name() string { return s.name }
func (s *stubTaskProvider) LoadTasks() ([]*core.Task, error) {
	s.loadCalls++
	return s.tasks, s.loadErr
}
func (s *stubTaskProvider) SaveTask(_ *core.Task) error    { return core.ErrReadOnly }
func (s *stubTaskProvider) SaveTasks(_ []*core.Task) error { return core.ErrReadOnly }
func (s *stubTaskProvider) DeleteTask(_ string) error      { return core.ErrReadOnly }
func (s *stubTaskProvider) MarkComplete(_ string) error    { return core.ErrReadOnly }
func (s *stubTaskProvider) Watch() <-chan core.ChangeEvent { return nil }
func (s *stubTaskProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{
		Items:   s.healthItems,
		Overall: core.HealthOK,
	}
}

func TestProviderBridge_RegisterAndRetrieve(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()
	provider := &stubTaskProvider{name: "textfile"}

	bridge.Register("conn-1", provider)

	got := bridge.Provider("conn-1")
	if got == nil {
		t.Fatal("expected provider, got nil")
	}
	if got.Name() != "textfile" {
		t.Errorf("got name %q, want %q", got.Name(), "textfile")
	}

	// Not found
	if bridge.Provider("no-such") != nil {
		t.Error("expected nil for unknown connection")
	}
}

func TestProviderBridge_Unregister(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()
	bridge.Register("conn-1", &stubTaskProvider{name: "textfile"})
	bridge.Unregister("conn-1")

	if bridge.Provider("conn-1") != nil {
		t.Error("expected nil after unregister")
	}
}

func TestProviderBridge_Providers(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()
	bridge.Register("a", &stubTaskProvider{name: "jira"})
	bridge.Register("b", &stubTaskProvider{name: "github"})

	all := bridge.Providers()
	if len(all) != 2 {
		t.Fatalf("got %d providers, want 2", len(all))
	}
}

func TestProviderBridge_CheckHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		items     []core.HealthCheckItem
		wantReach bool
		wantToken bool
		wantCount int
	}{
		{
			name: "all healthy with task count",
			items: []core.HealthCheckItem{
				{Name: "Database", Status: core.HealthOK, Message: "5 tasks loaded successfully"},
				{Name: "Task File", Status: core.HealthOK, Message: "exists"},
			},
			wantReach: true,
			wantToken: true,
			wantCount: 5,
		},
		{
			name: "database fail",
			items: []core.HealthCheckItem{
				{Name: "Database", Status: core.HealthFail, Message: "broken"},
			},
			wantReach: false,
			wantToken: true,
			wantCount: 0,
		},
		{
			name: "auth fail",
			items: []core.HealthCheckItem{
				{Name: "Authentication", Status: core.HealthFail, Message: "token expired"},
			},
			wantReach: false,
			wantToken: false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bridge := NewProviderBridge()
			provider := &stubTaskProvider{
				name:        "test",
				healthItems: tt.items,
			}
			bridge.Register("conn-1", provider)

			conn := &Connection{ID: "conn-1", ProviderName: "test"}
			result, err := bridge.CheckHealth(conn, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.APIReachable != tt.wantReach {
				t.Errorf("APIReachable = %v, want %v", result.APIReachable, tt.wantReach)
			}
			if result.TokenValid != tt.wantToken {
				t.Errorf("TokenValid = %v, want %v", result.TokenValid, tt.wantToken)
			}
			if result.TaskCount != tt.wantCount {
				t.Errorf("TaskCount = %d, want %d", result.TaskCount, tt.wantCount)
			}
			if !result.RateLimitOK {
				t.Error("RateLimitOK should always be true")
			}
		})
	}
}

func TestProviderBridge_CheckHealth_NoProvider(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()
	conn := &Connection{ID: "missing", ProviderName: "test"}

	_, err := bridge.CheckHealth(conn, "")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestProviderBridge_Sync(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()

	tasks := []*core.Task{
		{ID: "t1", Text: "task 1"},
		{ID: "t2", Text: "task 2"},
	}
	provider := &stubTaskProvider{name: "textfile", tasks: tasks}
	bridge.Register("conn-1", provider)

	conn := &Connection{ID: "conn-1", ProviderName: "textfile"}
	err := bridge.Sync(conn, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.TaskCount != 2 {
		t.Errorf("TaskCount = %d, want 2", conn.TaskCount)
	}
	if conn.LastSync.IsZero() {
		t.Error("LastSync should be set after sync")
	}
	if provider.loadCalls != 1 {
		t.Errorf("LoadTasks called %d times, want 1", provider.loadCalls)
	}
}

func TestProviderBridge_Sync_Error(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()

	provider := &stubTaskProvider{
		name:    "broken",
		loadErr: errTestSync,
	}
	bridge.Register("conn-1", provider)

	conn := &Connection{ID: "conn-1", ProviderName: "broken"}
	err := bridge.Sync(conn, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProviderBridge_Sync_NoProvider(t *testing.T) {
	t.Parallel()
	bridge := NewProviderBridge()
	conn := &Connection{ID: "missing", ProviderName: "test"}

	err := bridge.Sync(conn, "")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

var errTestSync = func() error {
	return &syncError{}
}()

type syncError struct{}

func (e *syncError) Error() string { return "sync failed" }

func TestMapHealthCheckResult_EmptyItems(t *testing.T) {
	t.Parallel()
	result := mapHealthCheckResult(core.HealthCheckResult{})

	if !result.APIReachable {
		t.Error("APIReachable should be true for empty items")
	}
	if !result.TokenValid {
		t.Error("TokenValid should be true for empty items")
	}
	if !result.RateLimitOK {
		t.Error("RateLimitOK should be true")
	}
	if result.TaskCount != 0 {
		t.Errorf("TaskCount = %d, want 0", result.TaskCount)
	}
}

func TestMapHealthCheckResult_Details(t *testing.T) {
	t.Parallel()
	cr := core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{Name: "Database", Status: core.HealthOK, Message: "10 tasks loaded"},
			{Name: "Sync", Status: core.HealthWarn, Message: "stale"},
		},
		Duration: time.Second,
	}

	result := mapHealthCheckResult(cr)
	if len(result.Details) != 2 {
		t.Errorf("got %d details, want 2", len(result.Details))
	}
	if result.TaskCount != 10 {
		t.Errorf("TaskCount = %d, want 10", result.TaskCount)
	}
}
