package reminders

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"
)

func TestCategorizeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantIs  error
		wantNil bool
	}{
		{
			name:    "nil input returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:   "deadline exceeded maps to ErrTimeout",
			err:    context.DeadlineExceeded,
			wantIs: ErrTimeout,
		},
		{
			name:   "wrapped deadline maps to ErrTimeout",
			err:    fmt.Errorf("op failed: %w", context.DeadlineExceeded),
			wantIs: ErrTimeout,
		},
		{
			name:   "not allowed maps to ErrPermissionDenied",
			err:    errors.New("osascript: not allowed assistant access"),
			wantIs: ErrPermissionDenied,
		},
		{
			name:   "denied maps to ErrPermissionDenied",
			err:    errors.New("access denied"),
			wantIs: ErrPermissionDenied,
		},
		{
			name:   "error code 1002 maps to ErrPermissionDenied",
			err:    errors.New("error 1002: Reminders access"),
			wantIs: ErrPermissionDenied,
		},
		{
			name:   "reminder not found maps to ErrReminderNotFound",
			err:    errors.New("complete reminder: reminder not found"),
			wantIs: ErrReminderNotFound,
		},
		{
			name: "uncategorized error returned as-is",
			err:  errors.New("some random error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := categorizeError(tt.err)

			if tt.wantNil {
				if got != nil {
					t.Errorf("categorizeError() = %v, want nil", got)
				}
				return
			}

			if tt.wantIs != nil {
				if !errors.Is(got, tt.wantIs) {
					t.Errorf("categorizeError() = %v, want errors.Is(%v)", got, tt.wantIs)
				}
				// Verify original error is still in the chain
				if !errors.Is(got, tt.err) {
					t.Errorf("categorizeError() lost original error in chain")
				}
			}
		})
	}
}

func TestCategorizeError_PreservesChain(t *testing.T) {
	t.Parallel()

	original := context.DeadlineExceeded
	result := categorizeError(original)

	if !errors.Is(result, ErrTimeout) {
		t.Error("result should match ErrTimeout")
	}
	if !errors.Is(result, original) {
		t.Error("result should still match original error")
	}
}

func TestIsTransient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrPermissionDenied is not transient",
			err:  fmt.Errorf("wrap: %w", ErrPermissionDenied),
			want: false,
		},
		{
			name: "ErrReminderNotFound is not transient",
			err:  fmt.Errorf("wrap: %w", ErrReminderNotFound),
			want: false,
		},
		{
			name: "ErrTimeout is transient",
			err:  fmt.Errorf("wrap: %w", ErrTimeout),
			want: true,
		},
		{
			name: "exec.ExitError is transient",
			err:  &exec.ExitError{},
			want: true,
		},
		{
			name: "generic error is not transient",
			err:  errors.New("something else"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTransient(tt.err); got != tt.want {
				t.Errorf("isTransient() = %v, want %v", got, tt.want)
			}
		})
	}
}
