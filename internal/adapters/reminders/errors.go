package reminders

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Sentinel errors for categorizing Apple Reminders failures.
var (
	// ErrPermissionDenied indicates macOS TCC denied Reminders access.
	ErrPermissionDenied = errors.New("reminders: permission denied")

	// ErrReminderNotFound indicates the target reminder ID does not exist.
	ErrReminderNotFound = errors.New("reminders: reminder not found")

	// ErrTimeout indicates the osascript operation exceeded its deadline.
	ErrTimeout = errors.New("reminders: operation timed out")
)

// categorizeError wraps a raw error with the appropriate sentinel error type.
// Returns nil for nil input. Non-categorizable errors are returned unwrapped.
func categorizeError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}

	msg := err.Error()

	if strings.Contains(msg, "not allowed") || strings.Contains(msg, "denied") || strings.Contains(msg, "1002") {
		return fmt.Errorf("%w: %w", ErrPermissionDenied, err)
	}

	if strings.Contains(msg, "reminder not found") {
		return fmt.Errorf("%w: %w", ErrReminderNotFound, err)
	}

	return err
}

// isTransient returns true if the error represents a transient failure
// that may succeed on retry.
func isTransient(err error) bool {
	if errors.Is(err, ErrPermissionDenied) || errors.Is(err, ErrReminderNotFound) {
		return false
	}
	if errors.Is(err, ErrTimeout) {
		return true
	}
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}
