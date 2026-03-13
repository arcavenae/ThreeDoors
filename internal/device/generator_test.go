package device

import (
	"errors"
	"testing"
)

// mockMachineIDReader returns a fixed machine ID for testing.
type mockMachineIDReader struct {
	id  string
	err error
}

func (m *mockMachineIDReader) ReadMachineID() (string, error) {
	return m.id, m.err
}

func TestGenerateDeviceID_Determinism(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{id: "test-machine-id-alpha"}
	configDir := "/tmp/test/.threedoors"

	id1, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() unexpected error: %v", err)
	}

	id2, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() unexpected error: %v", err)
	}

	if id1 != id2 {
		t.Errorf("same inputs produced different IDs: %s vs %s", id1, id2)
	}
}

func TestGenerateDeviceID_DifferentMachines(t *testing.T) {
	t.Parallel()

	configDir := "/tmp/test/.threedoors"

	readerA := &mockMachineIDReader{id: "test-machine-id-alpha"}
	readerB := &mockMachineIDReader{id: "test-machine-id-beta"}

	idA, err := GenerateDeviceID(readerA, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID(alpha) unexpected error: %v", err)
	}

	idB, err := GenerateDeviceID(readerB, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID(beta) unexpected error: %v", err)
	}

	if idA == idB {
		t.Errorf("different machine IDs produced same device ID: %s", idA)
	}
}

func TestGenerateDeviceID_DifferentConfigDirs(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{id: "test-machine-id-alpha"}

	id1, err := GenerateDeviceID(reader, "/home/user1/.threedoors")
	if err != nil {
		t.Fatalf("GenerateDeviceID(dir1) unexpected error: %v", err)
	}

	id2, err := GenerateDeviceID(reader, "/home/user2/.threedoors")
	if err != nil {
		t.Fatalf("GenerateDeviceID(dir2) unexpected error: %v", err)
	}

	if id1 == id2 {
		t.Errorf("different config dirs produced same device ID: %s", id1)
	}
}

func TestGenerateDeviceID_FallbackToUUIDv4(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{err: ErrMachineIDUnavailable}
	configDir := "/tmp/test/.threedoors"

	id, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() should succeed with fallback, got: %v", err)
	}

	if !id.IsValid() {
		t.Errorf("fallback ID should be valid, got: %s", id)
	}
}

func TestGenerateDeviceID_FallbackNotDeterministic(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{err: ErrMachineIDUnavailable}
	configDir := "/tmp/test/.threedoors"

	id1, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() unexpected error: %v", err)
	}

	id2, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() unexpected error: %v", err)
	}

	if id1 == id2 {
		t.Errorf("UUID v4 fallback should produce different IDs each call, but both were: %s", id1)
	}
}

func TestGenerateDeviceID_NonMachineIDError(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{err: errors.New("unexpected error")}
	configDir := "/tmp/test/.threedoors"

	_, err := GenerateDeviceID(reader, configDir)
	if err == nil {
		t.Error("GenerateDeviceID() should return error for non-MachineIDUnavailable errors")
	}
}

func TestGenerateDeviceID_ProducesValidUUID(t *testing.T) {
	t.Parallel()

	reader := &mockMachineIDReader{id: "test-machine-id-alpha"}
	configDir := "/tmp/test/.threedoors"

	id, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		t.Fatalf("GenerateDeviceID() unexpected error: %v", err)
	}

	if !id.IsValid() {
		t.Errorf("generated ID should be valid, got: %s", id)
	}
}
