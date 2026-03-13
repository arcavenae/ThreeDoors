package device

import (
	"errors"
	"testing"
)

func TestNewDeviceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid v5", "a1b2c3d4-e5f6-5a7b-8c9d-0e1f2a3b4c5d", false},
		{"empty string", "", true},
		{"not a uuid", "not-a-uuid", true},
		{"partial uuid", "550e8400-e29b-41d4", true},
		{"uuid without dashes", "550e8400e29b41d4a716446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := NewDeviceID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewDeviceID(%q) = %v, want error", tt.input, id)
				}
				if !errors.Is(err, ErrInvalidDeviceID) {
					t.Errorf("NewDeviceID(%q) error = %v, want ErrInvalidDeviceID", tt.input, err)
				}
			} else {
				if err != nil {
					t.Errorf("NewDeviceID(%q) unexpected error: %v", tt.input, err)
				}
				if id == "" {
					t.Errorf("NewDeviceID(%q) returned empty DeviceID", tt.input)
				}
			}
		})
	}
}

func TestDeviceID_String(t *testing.T) {
	t.Parallel()

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	s := id.String()
	if s == "" {
		t.Error("DeviceID.String() returned empty string")
	}
}

func TestDeviceID_ZeroValue(t *testing.T) {
	t.Parallel()

	var id DeviceID
	if id.IsValid() {
		t.Error("zero-value DeviceID should not be valid")
	}
}
