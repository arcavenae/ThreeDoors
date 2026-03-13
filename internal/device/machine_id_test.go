package device

import (
	"errors"
	"testing"
)

func TestMockMachineIDReader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reader  MachineIDReader
		wantID  string
		wantErr bool
	}{
		{
			name:   "returns configured id",
			reader: &mockMachineIDReader{id: "test-machine-id"},
			wantID: "test-machine-id",
		},
		{
			name:    "returns configured error",
			reader:  &mockMachineIDReader{err: ErrMachineIDUnavailable},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := tt.reader.ReadMachineID()
			if tt.wantErr {
				if err == nil {
					t.Error("ReadMachineID() should return error")
				}
				if !errors.Is(err, ErrMachineIDUnavailable) {
					t.Errorf("ReadMachineID() error = %v, want ErrMachineIDUnavailable", err)
				}
			} else {
				if err != nil {
					t.Errorf("ReadMachineID() unexpected error: %v", err)
				}
				if id != tt.wantID {
					t.Errorf("ReadMachineID() = %s, want %s", id, tt.wantID)
				}
			}
		})
	}
}

func TestPlatformMachineIDReader(t *testing.T) {
	t.Parallel()

	reader := NewPlatformMachineIDReader()
	id, err := reader.ReadMachineID()
	// On CI or containers, machine-id may not be available.
	// We just verify the function doesn't panic and returns
	// either a valid ID or ErrMachineIDUnavailable.
	if err != nil {
		if !errors.Is(err, ErrMachineIDUnavailable) {
			t.Errorf("ReadMachineID() unexpected error type: %v", err)
		}
		return
	}

	if id == "" {
		t.Error("ReadMachineID() returned empty string without error")
	}
}
