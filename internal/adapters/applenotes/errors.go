package applenotes

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrorCategory classifies adapter errors for retry and user guidance.
type ErrorCategory int

const (
	// ErrorTransient indicates a temporary failure that may succeed on retry.
	// Examples: timeout, temporary exec failures.
	ErrorTransient ErrorCategory = iota

	// ErrorPermanent indicates a failure that will not resolve by retrying.
	// Examples: note not found, osascript binary missing.
	ErrorPermanent

	// ErrorConfiguration indicates the user must take action to resolve.
	// Examples: permission denied, automation not authorized.
	ErrorConfiguration
)

// String returns the human-readable category name.
func (c ErrorCategory) String() string {
	switch c {
	case ErrorTransient:
		return "transient"
	case ErrorPermanent:
		return "permanent"
	case ErrorConfiguration:
		return "configuration"
	default:
		return "unknown"
	}
}

// AdapterError wraps an underlying error with category, user-friendly message,
// and actionable suggestion.
type AdapterError struct {
	Category   ErrorCategory
	Message    string
	Suggestion string
	Err        error
}

// Error implements the error interface with a user-friendly format.
func (e *AdapterError) Error() string {
	return fmt.Sprintf("apple notes [%s]: %s", e.Category, e.Message)
}

// Unwrap returns the underlying error for errors.Is/errors.As chain traversal.
func (e *AdapterError) Unwrap() error {
	return e.Err
}

// IsTransient returns true if the error is a transient AdapterError.
func IsTransient(err error) bool {
	var adapterErr *AdapterError
	if errors.As(err, &adapterErr) {
		return adapterErr.Category == ErrorTransient
	}
	return false
}

// categorizeError maps raw errors to categorized AdapterError with suggestions.
func (p *AppleNotesProvider) categorizeError(err error) *AdapterError {
	msg := err.Error()

	if errors.Is(err, context.DeadlineExceeded) {
		return &AdapterError{
			Category:   ErrorTransient,
			Message:    "osascript timed out",
			Suggestion: "Apple Notes may be slow to respond. The operation will retry automatically.",
			Err:        err,
		}
	}

	if strings.Contains(msg, "Can't get note") || strings.Contains(msg, "can't get note") {
		return &AdapterError{
			Category:   ErrorPermanent,
			Message:    fmt.Sprintf("note %q not found", p.noteTitle),
			Suggestion: fmt.Sprintf("Verify a note titled %q exists in Apple Notes.", p.noteTitle),
			Err:        err,
		}
	}

	if strings.Contains(msg, "not allowed") || strings.Contains(msg, "Not authorized") {
		return &AdapterError{
			Category:   ErrorConfiguration,
			Message:    "automation permission denied",
			Suggestion: "Grant automation access in System Settings > Privacy & Security > Automation.",
			Err:        err,
		}
	}

	if errors.Is(err, exec.ErrNotFound) {
		return &AdapterError{
			Category:   ErrorPermanent,
			Message:    "osascript not found",
			Suggestion: "Apple Notes integration requires macOS with osascript installed.",
			Err:        err,
		}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &AdapterError{
			Category:   ErrorTransient,
			Message:    "osascript failed",
			Suggestion: "The AppleScript command failed unexpectedly. The operation will retry automatically.",
			Err:        err,
		}
	}

	return &AdapterError{
		Category:   ErrorTransient,
		Message:    fmt.Sprintf("unexpected error: %v", err),
		Suggestion: "An unexpected error occurred. The operation will retry automatically.",
		Err:        err,
	}
}
