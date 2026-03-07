package reminders

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestWithRetry_SucceedsFirstAttempt(t *testing.T) {
	t.Parallel()

	calls := 0
	err := withRetry(context.Background(), DefaultRetryConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("withRetry() error = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("op called %d times, want 1", calls)
	}
}

func TestWithRetry_RetriesTransientErrors(t *testing.T) {
	t.Parallel()

	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond}
	calls := 0
	err := withRetry(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return fmt.Errorf("transient: %w", ErrTimeout)
		}
		return nil
	})
	if err != nil {
		t.Errorf("withRetry() error = %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("op called %d times, want 3", calls)
	}
}

func TestWithRetry_StopsOnPermanentError(t *testing.T) {
	t.Parallel()

	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond}
	calls := 0
	err := withRetry(context.Background(), cfg, func() error {
		calls++
		return fmt.Errorf("permanent: %w", ErrReminderNotFound)
	})

	if !errors.Is(err, ErrReminderNotFound) {
		t.Errorf("withRetry() error = %v, want ErrReminderNotFound", err)
	}
	if calls != 1 {
		t.Errorf("op called %d times, want 1 (should stop on permanent error)", calls)
	}
}

func TestWithRetry_ExhaustsAttempts(t *testing.T) {
	t.Parallel()

	cfg := RetryConfig{MaxAttempts: 2, InitialBackoff: time.Millisecond}
	calls := 0
	err := withRetry(context.Background(), cfg, func() error {
		calls++
		return fmt.Errorf("still failing: %w", ErrTimeout)
	})

	if err == nil {
		t.Error("withRetry() should return error after exhausting attempts")
	}
	if calls != 2 {
		t.Errorf("op called %d times, want 2", calls)
	}
}

func TestWithRetry_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cfg := RetryConfig{MaxAttempts: 5, InitialBackoff: time.Second}

	calls := 0
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := withRetry(ctx, cfg, func() error {
		calls++
		return fmt.Errorf("transient: %w", ErrTimeout)
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("withRetry() error = %v, want context.Canceled", err)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", cfg.MaxAttempts)
	}
	if cfg.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", cfg.InitialBackoff)
	}
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	t.Parallel()

	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: 10 * time.Millisecond}
	var timestamps []time.Time

	_ = withRetry(context.Background(), cfg, func() error {
		timestamps = append(timestamps, time.Now().UTC())
		return fmt.Errorf("fail: %w", ErrTimeout)
	})

	if len(timestamps) != 3 {
		t.Fatalf("got %d timestamps, want 3", len(timestamps))
	}

	// First retry should wait ~10ms, second ~20ms
	gap1 := timestamps[1].Sub(timestamps[0])
	gap2 := timestamps[2].Sub(timestamps[1])

	if gap1 < 5*time.Millisecond {
		t.Errorf("first backoff too short: %v", gap1)
	}
	if gap2 < gap1/2 {
		t.Errorf("second backoff (%v) should be longer than first (%v)", gap2, gap1)
	}
}
