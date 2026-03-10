package dispatch

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// guardrailMockDispatcher implements Dispatcher for guardrail tests.
type guardrailMockDispatcher struct {
	workers []WorkerInfo
	listErr error
}

func (m *guardrailMockDispatcher) CreateWorker(_ context.Context, _ string) (string, error) {
	return "test-worker", nil
}

func (m *guardrailMockDispatcher) ListWorkers(_ context.Context) ([]WorkerInfo, error) {
	return m.workers, m.listErr
}

func (m *guardrailMockDispatcher) GetHistory(_ context.Context, _ int) ([]HistoryEntry, error) {
	return nil, nil
}

func (m *guardrailMockDispatcher) RemoveWorker(_ context.Context, _ string) error {
	return nil
}

func (m *guardrailMockDispatcher) CheckAvailable(_ context.Context) error {
	return nil
}

func newTestGuardrailChecker(t *testing.T, cfg DevDispatchConfig, workers []WorkerInfo) (*GuardrailChecker, *AuditLogger) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dev-dispatch.log")
	audit, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
		return nil, nil
	}

	disp := &guardrailMockDispatcher{workers: workers}
	return NewGuardrailChecker(cfg, disp, audit), audit
}

func TestCheckMaxConcurrentUnderLimit(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{MaxConcurrent: 3}
	checker, _ := newTestGuardrailChecker(t, cfg, []WorkerInfo{
		{Name: "w1", Status: "running"},
	})

	if err := checker.CheckMaxConcurrent(context.Background()); err != nil {
		t.Errorf("should pass with 1/3 workers: %v", err)
	}
}

func TestCheckMaxConcurrentAtLimit(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{MaxConcurrent: 2}
	checker, _ := newTestGuardrailChecker(t, cfg, []WorkerInfo{
		{Name: "w1", Status: "running"},
		{Name: "w2", Status: "running"},
	})

	err := checker.CheckMaxConcurrent(context.Background())
	if err == nil {
		t.Fatal("should fail at limit")
		return
	}

	var violation *GuardrailViolation
	if !errors.As(err, &violation) {
		t.Fatalf("error should be GuardrailViolation, got %T", err)
	}
	if violation.Message != "Max concurrent workers reached (2/2). Wait for a worker to complete." {
		t.Errorf("message = %q", violation.Message)
	}
}

func TestCheckMaxConcurrentListError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	audit, _ := NewAuditLogger(filepath.Join(dir, "audit.log"))
	disp := &guardrailMockDispatcher{listErr: errors.New("connection failed")}
	checker := NewGuardrailChecker(DevDispatchConfig{MaxConcurrent: 2}, disp, audit)

	err := checker.CheckMaxConcurrent(context.Background())
	if err == nil {
		t.Fatal("should propagate list error")
		return
	}
	// Should not be a GuardrailViolation — it's an infrastructure error
	var violation *GuardrailViolation
	if errors.As(err, &violation) {
		t.Error("infrastructure error should not be a GuardrailViolation")
	}
}

func TestCheckCooldownNoHistory(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{CooldownMinutes: 5}
	checker, _ := newTestGuardrailChecker(t, cfg, nil)

	if err := checker.CheckCooldown("task-1"); err != nil {
		t.Errorf("should pass with no history: %v", err)
	}
}

func TestCheckCooldownExpired(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{CooldownMinutes: 5}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	// Log a dispatch 10 minutes ago
	old := time.Now().UTC().Add(-10 * time.Minute)
	if err := audit.Log(AuditEntry{Timestamp: old, EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if err := checker.CheckCooldown("task-1"); err != nil {
		t.Errorf("should pass after cooldown expired: %v", err)
	}
}

func TestCheckCooldownActive(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{CooldownMinutes: 5}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	// Log a dispatch 1 minute ago
	recent := time.Now().UTC().Add(-1 * time.Minute)
	if err := audit.Log(AuditEntry{Timestamp: recent, EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
		t.Fatalf("Log: %v", err)
	}

	err := checker.CheckCooldown("task-1")
	if err == nil {
		t.Fatal("should fail during cooldown")
		return
	}

	var violation *GuardrailViolation
	if !errors.As(err, &violation) {
		t.Fatalf("error should be GuardrailViolation, got %T", err)
	}
}

func TestCheckCooldownPerTask(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{CooldownMinutes: 5}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	// Log a recent dispatch for task-1
	recent := time.Now().UTC().Add(-1 * time.Minute)
	if err := audit.Log(AuditEntry{Timestamp: recent, EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// task-2 should pass — cooldown is per-task
	if err := checker.CheckCooldown("task-2"); err != nil {
		t.Errorf("different task should not be affected by cooldown: %v", err)
	}
}

func TestCheckDailyLimitUnderLimit(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{DailyLimit: 10}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	now := time.Now().UTC()
	if err := audit.Log(AuditEntry{Timestamp: now, EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if err := checker.CheckDailyLimit(); err != nil {
		t.Errorf("should pass with 1/10 dispatches: %v", err)
	}
}

func TestCheckDailyLimitAtLimit(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{DailyLimit: 3}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	now := time.Now().UTC()
	for i := range 3 {
		if err := audit.Log(AuditEntry{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			EventType: AuditDispatch,
			TaskID:    "task-1",
		}); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	err := checker.CheckDailyLimit()
	if err == nil {
		t.Fatal("should fail at limit")
		return
	}

	var violation *GuardrailViolation
	if !errors.As(err, &violation) {
		t.Fatalf("error should be GuardrailViolation, got %T", err)
	}
	if violation.Message != "Daily dispatch limit reached (3/3)." {
		t.Errorf("message = %q", violation.Message)
	}
}

func TestCheckDailyLimitIgnoresOldDays(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{DailyLimit: 2}
	checker, audit := newTestGuardrailChecker(t, cfg, nil)

	// Log dispatches from yesterday
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	for range 5 {
		if err := audit.Log(AuditEntry{Timestamp: yesterday, EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	if err := checker.CheckDailyLimit(); err != nil {
		t.Errorf("yesterday's dispatches should not count: %v", err)
	}
}

func TestCheckAllPassesAllChecks(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{MaxConcurrent: 5, CooldownMinutes: 1, DailyLimit: 100}
	checker, _ := newTestGuardrailChecker(t, cfg, nil)

	if err := checker.CheckAll(context.Background(), "task-1"); err != nil {
		t.Errorf("CheckAll should pass: %v", err)
	}
}

func TestCheckAllFailsOnFirstViolation(t *testing.T) {
	t.Parallel()
	cfg := DevDispatchConfig{MaxConcurrent: 1, CooldownMinutes: 5, DailyLimit: 100}
	checker, _ := newTestGuardrailChecker(t, cfg, []WorkerInfo{
		{Name: "w1", Status: "running"},
	})

	err := checker.CheckAll(context.Background(), "task-1")
	if err == nil {
		t.Fatal("CheckAll should fail")
		return
	}

	var violation *GuardrailViolation
	if !errors.As(err, &violation) {
		t.Fatalf("error should be GuardrailViolation, got %T", err)
	}
}

func TestCheckCooldownNilAudit(t *testing.T) {
	t.Parallel()
	disp := &guardrailMockDispatcher{}
	checker := NewGuardrailChecker(DevDispatchConfig{CooldownMinutes: 5}, disp, nil)

	if err := checker.CheckCooldown("task-1"); err != nil {
		t.Errorf("should pass with nil audit: %v", err)
	}
}

func TestCheckDailyLimitNilAudit(t *testing.T) {
	t.Parallel()
	disp := &guardrailMockDispatcher{}
	checker := NewGuardrailChecker(DevDispatchConfig{DailyLimit: 10}, disp, nil)

	if err := checker.CheckDailyLimit(); err != nil {
		t.Errorf("should pass with nil audit: %v", err)
	}
}

func TestBuildDryRunCommand(t *testing.T) {
	t.Parallel()
	item := QueueItem{
		TaskText: "Fix the login bug",
	}
	cmd := BuildDryRunCommand(item)
	if cmd == "" {
		t.Fatal("command should not be empty")
	}
	if len(cmd) < 20 {
		t.Errorf("command seems too short: %q", cmd)
	}
}

func TestGuardrailViolationError(t *testing.T) {
	t.Parallel()
	v := &GuardrailViolation{Message: "test message"}
	if v.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", v.Error(), "test message")
	}
}
