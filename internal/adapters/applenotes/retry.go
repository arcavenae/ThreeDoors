package applenotes

import (
	"context"
	"time"
)

// LogFunc is a function that accepts a log message string.
type LogFunc func(msg string)

// RetryConfig controls retry behavior for transient failures.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
}

// Config holds all configurable settings for the Apple Notes adapter.
type Config struct {
	Timeout time.Duration
	Retry   RetryConfig
	Logger  LogFunc
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
	}
}

// DefaultConfig returns the default adapter configuration.
func DefaultConfig() Config {
	return Config{
		Timeout: 10 * time.Second,
		Retry:   DefaultRetryConfig(),
	}
}

// executeWithRetry runs an operation, retrying on transient errors with exponential backoff.
func (p *AppleNotesProvider) executeWithRetry(ctx context.Context, script string) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= p.config.Retry.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := p.config.Retry.InitialBackoff * (1 << (attempt - 1))
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		output, err := p.executor(ctx, script)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Categorize the error to decide whether to retry
		adapterErr := p.categorizeError(err)
		if adapterErr.Category != ErrorTransient {
			return "", adapterErr
		}
	}

	return "", p.categorizeError(lastErr)
}

// log emits a log message if a logger is configured.
func (p *AppleNotesProvider) log(msg string) {
	if p.config.Logger != nil {
		p.config.Logger(msg)
	}
}
