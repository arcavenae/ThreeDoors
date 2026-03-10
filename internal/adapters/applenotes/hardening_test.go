package applenotes

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Story 3.5.4: Apple Notes Adapter Hardening — TDD Tests
// =============================================================================

// --- AC-1: Configurable Timeout ---

func TestConfig_DefaultTimeout(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Timeout != 10*time.Second {
		t.Errorf("DefaultConfig().Timeout = %v, want 10s", cfg.Timeout)
	}
}

func TestConfig_CustomTimeout(t *testing.T) {
	cfg := Config{Timeout: 30 * time.Second}
	provider := NewAppleNotesProviderWithConfig("TestNote", mockExecutor("", nil), cfg)
	if provider.config.Timeout != 30*time.Second {
		t.Errorf("provider.config.Timeout = %v, want 30s", provider.config.Timeout)
	}
}

func TestConfig_TimeoutAppliedToExecutor(t *testing.T) {
	// Executor that records the context deadline
	var gotDeadline time.Time
	var hasDeadline bool
	executor := func(ctx context.Context, script string) (string, error) {
		gotDeadline, hasDeadline = ctx.Deadline()
		_ = gotDeadline
		if !hasDeadline {
			return "", fmt.Errorf("no deadline set on context")
		}
		return "- [ ] Task", nil
	}

	cfg := Config{
		Timeout: 5 * time.Second,
		Retry:   RetryConfig{MaxRetries: 0},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if !hasDeadline {
		t.Error("executor should receive context with deadline from configured timeout")
	}
}

func TestConfig_AllOperationsUseConfiguredTimeout(t *testing.T) {
	tests := []struct {
		name string
		op   func(p *AppleNotesProvider) error
	}{
		{
			name: "LoadTasks",
			op: func(p *AppleNotesProvider) error {
				_, err := p.LoadTasks()
				return err
			},
		},
		{
			name: "SaveTask",
			op: func(p *AppleNotesProvider) error {
				return p.SaveTask(newTestTask("id", "text", "todo", baseTime))
			},
		},
		{
			name: "DeleteTask",
			op: func(p *AppleNotesProvider) error {
				return p.DeleteTask("some-id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sawDeadline bool
			executor := func(ctx context.Context, script string) (string, error) {
				if _, ok := ctx.Deadline(); ok {
					sawDeadline = true
				}
				return "- [ ] Task", nil
			}
			cfg := Config{
				Timeout: 15 * time.Second,
				Retry:   RetryConfig{MaxRetries: 0},
			}
			p := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)
			_ = tt.op(p)
			if !sawDeadline {
				t.Errorf("%s should use configured timeout via context deadline", tt.name)
			}
		})
	}
}

// --- AC-2: Retry with Exponential Backoff ---

func TestRetry_TransientErrorRetriesUpToMax(t *testing.T) {
	var callCount int32
	transientErr := context.DeadlineExceeded

	executor := func(ctx context.Context, script string) (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "", transientErr
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond, // fast for tests
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error after retries exhausted, got nil")
		return
	}

	// 1 initial + 3 retries = 4 total calls
	got := atomic.LoadInt32(&callCount)
	if got != 4 {
		t.Errorf("expected 4 total calls (1 initial + 3 retries), got %d", got)
	}
}

func TestRetry_PermanentErrorDoesNotRetry(t *testing.T) {
	var callCount int32
	permanentErr := errors.New("execution error: Can't get note \"Missing\"")

	executor := func(ctx context.Context, script string) (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "", permanentErr
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error, got nil")
		return
	}

	got := atomic.LoadInt32(&callCount)
	if got != 1 {
		t.Errorf("permanent errors should not retry: expected 1 call, got %d", got)
	}
}

func TestRetry_ConfigurationErrorDoesNotRetry(t *testing.T) {
	var callCount int32
	configErr := errors.New("Not authorized to send Apple events")

	executor := func(ctx context.Context, script string) (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "", configErr
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error, got nil")
		return
	}

	got := atomic.LoadInt32(&callCount)
	if got != 1 {
		t.Errorf("configuration errors should not retry: expected 1 call, got %d", got)
	}
}

func TestRetry_SucceedsAfterTransientFailure(t *testing.T) {
	var callCount int32

	executor := func(ctx context.Context, script string) (string, error) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 2 {
			return "", context.DeadlineExceeded
		}
		return "- [ ] Task A", nil
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task after retry succeeds, got %d", len(tasks))
	}
}

func TestRetry_ExponentialBackoffTiming(t *testing.T) {
	var timestamps []time.Time

	executor := func(ctx context.Context, script string) (string, error) {
		timestamps = append(timestamps, time.Now())
		return "", context.DeadlineExceeded
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 50 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, _ = provider.LoadTasks()

	if len(timestamps) != 4 {
		t.Fatalf("expected 4 timestamps, got %d", len(timestamps))
	}

	// Verify exponential backoff: delays should roughly be 50ms, 100ms, 200ms
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		expectedMin := time.Duration(1<<(i-1)) * 50 * time.Millisecond / 2 // allow 50% tolerance
		if gap < expectedMin {
			t.Errorf("gap between call %d and %d = %v, expected at least ~%v", i-1, i, gap, expectedMin)
		}
	}
}

func TestRetry_ZeroMaxRetriesDisablesRetry(t *testing.T) {
	var callCount int32

	executor := func(ctx context.Context, script string) (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "", context.DeadlineExceeded
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     0,
			InitialBackoff: 1 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	_, _ = provider.LoadTasks()

	got := atomic.LoadInt32(&callCount)
	if got != 1 {
		t.Errorf("with MaxRetries=0, expected 1 call, got %d", got)
	}
}

// --- AC-3: Error Categorization ---

func TestErrorCategory_Timeout(t *testing.T) {
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", context.DeadlineExceeded),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("expected error")
		return
	}

	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorTransient {
		t.Errorf("timeout should be categorized as transient, got %v", adapterErr.Category)
	}
}

func TestErrorCategory_NoteNotFound(t *testing.T) {
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", errors.New("execution error: Can't get note \"TestNote\"")),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorPermanent {
		t.Errorf("note not found should be permanent, got %v", adapterErr.Category)
	}
}

func TestErrorCategory_OsascriptNotFound(t *testing.T) {
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", exec.ErrNotFound),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorPermanent {
		t.Errorf("osascript not found should be permanent, got %v", adapterErr.Category)
	}
}

func TestErrorCategory_PermissionDenied(t *testing.T) {
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", errors.New("Not authorized to send Apple events to Notes")),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorConfiguration {
		t.Errorf("permission denied should be configuration, got %v", adapterErr.Category)
	}
}

func TestErrorCategory_AutomationNotAllowed(t *testing.T) {
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", errors.New("System Events got an error: osascript is not allowed")),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorConfiguration {
		t.Errorf("automation not allowed should be configuration, got %v", adapterErr.Category)
	}
}

func TestErrorCategory_ExitError(t *testing.T) {
	// exec.ExitError is a permanent error (generic script failure)
	exitErr := &exec.ExitError{}
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", exitErr),
		Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
	)

	_, err := provider.LoadTasks()
	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("error should be *AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Category != ErrorTransient {
		t.Errorf("generic exec exit error should be transient (may be intermittent), got %v", adapterErr.Category)
	}
}

// --- AC-4: User-Friendly Error Messages ---

func TestAdapterError_HasSuggestion(t *testing.T) {
	tests := []struct {
		name       string
		inputErr   error
		wantSubstr string
	}{
		{
			name:       "timeout suggests retry later",
			inputErr:   context.DeadlineExceeded,
			wantSubstr: "retry",
		},
		{
			name:       "note not found suggests check note title",
			inputErr:   errors.New("Can't get note \"Missing\""),
			wantSubstr: "note",
		},
		{
			name:       "permission denied suggests system settings",
			inputErr:   errors.New("Not authorized to send Apple events"),
			wantSubstr: "Privacy",
		},
		{
			name:       "osascript not found suggests macOS",
			inputErr:   exec.ErrNotFound,
			wantSubstr: "macOS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAppleNotesProviderWithConfig("TestNote",
				mockExecutor("", tt.inputErr),
				Config{Timeout: 10 * time.Second, Retry: RetryConfig{MaxRetries: 0}},
			)

			_, err := provider.LoadTasks()
			var adapterErr *AdapterError
			if !errors.As(err, &adapterErr) {
				t.Fatalf("expected *AdapterError, got %T", err)
			}
			if adapterErr.Suggestion == "" {
				t.Error("AdapterError should have a non-empty Suggestion")
			}
			if !strings.Contains(strings.ToLower(adapterErr.Suggestion), strings.ToLower(tt.wantSubstr)) {
				t.Errorf("Suggestion %q should contain %q", adapterErr.Suggestion, tt.wantSubstr)
			}
		})
	}
}

func TestAdapterError_ErrorStringIncludesCategory(t *testing.T) {
	err := &AdapterError{
		Category: ErrorTransient,
		Message:  "osascript timed out",
		Err:      context.DeadlineExceeded,
	}
	if !strings.Contains(err.Error(), "transient") {
		t.Errorf("Error() = %q should contain category 'transient'", err.Error())
	}
}

func TestAdapterError_UnwrapPreservesChain(t *testing.T) {
	inner := context.DeadlineExceeded
	err := &AdapterError{
		Category: ErrorTransient,
		Message:  "timed out",
		Err:      inner,
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("errors.Is should find context.DeadlineExceeded through Unwrap")
	}
}

// --- AC-5: Sync Logging (no sensitive data) ---

func TestSyncLog_RecordsOperations(t *testing.T) {
	var logs []string
	logger := func(msg string) {
		logs = append(logs, msg)
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry:   RetryConfig{MaxRetries: 0},
		Logger:  logger,
	}
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("- [ ] Task A", nil), cfg)

	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logs) == 0 {
		t.Fatal("expected at least one log entry for LoadTasks")
	}

	// Verify log contains operation name
	found := false
	for _, log := range logs {
		if strings.Contains(log, "LoadTasks") || strings.Contains(log, "load") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("logs should contain operation name, got: %v", logs)
	}
}

func TestSyncLog_NoSensitiveData(t *testing.T) {
	var logs []string
	logger := func(msg string) {
		logs = append(logs, msg)
	}

	sensitiveContent := "- [ ] Buy drugs secretly\n- [x] Secret password reminder"
	cfg := Config{
		Timeout: 10 * time.Second,
		Retry:   RetryConfig{MaxRetries: 0},
		Logger:  logger,
	}
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor(sensitiveContent, nil), cfg)

	_, _ = provider.LoadTasks()

	for _, log := range logs {
		if strings.Contains(log, "Buy drugs") || strings.Contains(log, "Secret password") {
			t.Errorf("log should not contain sensitive task content: %q", log)
		}
	}
}

func TestSyncLog_RecordsErrorWithoutContent(t *testing.T) {
	var logs []string
	logger := func(msg string) {
		logs = append(logs, msg)
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry:   RetryConfig{MaxRetries: 0},
		Logger:  logger,
	}
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("", context.DeadlineExceeded), cfg)

	_, _ = provider.LoadTasks()

	if len(logs) == 0 {
		t.Fatal("expected log entry even for failed operations")
	}

	// Should log the error category/type but not raw note content
	found := false
	for _, log := range logs {
		if strings.Contains(log, "error") || strings.Contains(log, "failed") || strings.Contains(log, "timeout") {
			found = true
		}
	}
	if !found {
		t.Errorf("logs should mention the error, got: %v", logs)
	}
}

func TestSyncLog_NilLoggerDoesNotPanic(t *testing.T) {
	cfg := Config{
		Timeout: 10 * time.Second,
		Retry:   RetryConfig{MaxRetries: 0},
		Logger:  nil,
	}
	provider := NewAppleNotesProviderWithConfig("TestNote",
		mockExecutor("- [ ] Task", nil), cfg)

	// Should not panic with nil logger
	_, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Backward Compatibility ---

func TestNewAppleNotesProvider_UsesDefaults(t *testing.T) {
	provider := NewAppleNotesProvider("TestNote")
	if provider.config.Timeout != 10*time.Second {
		t.Errorf("default provider should use 10s timeout, got %v", provider.config.Timeout)
	}
	if provider.config.Retry.MaxRetries != 3 {
		t.Errorf("default provider should have 3 max retries, got %d", provider.config.Retry.MaxRetries)
	}
}

func TestNewAppleNotesProviderWithExecutor_UsesDefaults(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	if provider.config.Timeout != 10*time.Second {
		t.Errorf("default provider should use 10s timeout, got %v", provider.config.Timeout)
	}
}

// --- Retry applies to write operations too ---

func TestRetry_SaveTaskRetriesOnTransient(t *testing.T) {
	var callCount int32

	executor := func(ctx context.Context, script string) (string, error) {
		n := atomic.AddInt32(&callCount, 1)
		// Read succeeds on first call, write fails transiently twice then succeeds
		if strings.Contains(script, "get plaintext") {
			return "- [ ] Task A", nil
		}
		if n <= 3 {
			return "", context.DeadlineExceeded
		}
		return "", nil
	}

	cfg := Config{
		Timeout: 10 * time.Second,
		Retry: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
		},
	}
	provider := NewAppleNotesProviderWithConfig("TestNote", executor, cfg)

	task := newTestTask("id", "Task A", "todo", baseTime)
	err := provider.SaveTask(task)
	if err != nil {
		t.Fatalf("SaveTask should succeed after retries: %v", err)
	}
}
