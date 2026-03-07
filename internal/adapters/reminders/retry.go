package reminders

import (
	"context"
	"time"
)

// RetryConfig controls retry behavior for transient failures.
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
}

// DefaultRetryConfig returns sensible defaults for Apple Reminders operations.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
	}
}

// withRetry executes op, retrying on transient errors with exponential backoff.
// Non-transient errors are returned immediately without retry.
func withRetry(ctx context.Context, cfg RetryConfig, op func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			backoff := cfg.InitialBackoff * (1 << (attempt - 1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		lastErr = op()
		if lastErr == nil {
			return nil
		}

		if !isTransient(lastErr) {
			return lastErr
		}
	}

	return lastErr
}
