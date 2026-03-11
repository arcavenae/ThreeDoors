package retrospector

import (
	"path/filepath"
	"testing"
)

func TestKillSwitchNewEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	ks, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}
	if ks.IsReadOnly() {
		t.Error("new kill switch should not be read-only")
	}
	if ks.ConsecutiveRejects() != 0 {
		t.Errorf("consecutive rejects = %d, want 0", ks.ConsecutiveRejects())
	}
}

func TestKillSwitchTriggersAfterThreeRejects(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	ks, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// First two rejections should not trigger
	for i := 1; i <= 2; i++ {
		if err := ks.RecordOutcome("P-00"+string(rune('0'+i)), OutcomeRejected); err != nil {
			t.Fatalf("RecordOutcome %d: %v", i, err)
		}
		if ks.IsReadOnly() {
			t.Errorf("should not be read-only after %d rejections", i)
		}
	}

	// Third rejection triggers kill switch
	if err := ks.RecordOutcome("P-003", OutcomeRejected); err != nil {
		t.Fatalf("RecordOutcome 3: %v", err)
	}
	if !ks.IsReadOnly() {
		t.Error("should be read-only after 3 consecutive rejections")
	}
	if !ks.NeedsRecalibration() {
		t.Error("should need recalibration after kill switch trigger")
	}
}

func TestKillSwitchResetsOnAcceptance(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	ks, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// Two rejections then an acceptance
	if err := ks.RecordOutcome("P-001", OutcomeRejected); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}
	if err := ks.RecordOutcome("P-002", OutcomeRejected); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}
	if err := ks.RecordOutcome("P-003", OutcomeAccepted); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}

	if ks.ConsecutiveRejects() != 0 {
		t.Errorf("consecutive rejects = %d, want 0 after acceptance", ks.ConsecutiveRejects())
	}
	if ks.IsReadOnly() {
		t.Error("should not be read-only after acceptance resets counter")
	}
}

func TestKillSwitchResetsOnDeferred(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	ks, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	if err := ks.RecordOutcome("P-001", OutcomeRejected); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}
	if err := ks.RecordOutcome("P-002", OutcomeDeferred); err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}

	if ks.ConsecutiveRejects() != 0 {
		t.Errorf("consecutive rejects = %d, want 0 after deferred", ks.ConsecutiveRejects())
	}
}

func TestKillSwitchPersistence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	// Create and trigger
	ks1, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := ks1.RecordOutcome("P-001", OutcomeRejected); err != nil {
			t.Fatalf("RecordOutcome: %v", err)
		}
	}

	// Load from disk
	ks2, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch reload: %v", err)
	}
	if !ks2.IsReadOnly() {
		t.Error("persisted state should be read-only")
	}
	if ks2.ConsecutiveRejects() != 3 {
		t.Errorf("persisted consecutive rejects = %d, want 3", ks2.ConsecutiveRejects())
	}
}

func TestKillSwitchReset(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "killswitch.json")

	ks, err := NewKillSwitch(path)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// Trigger
	for i := 0; i < 3; i++ {
		if err := ks.RecordOutcome("P-001", OutcomeRejected); err != nil {
			t.Fatalf("RecordOutcome: %v", err)
		}
	}
	if !ks.IsReadOnly() {
		t.Fatal("should be read-only before reset")
	}

	// Reset
	if err := ks.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if ks.IsReadOnly() {
		t.Error("should not be read-only after reset")
	}
	if ks.NeedsRecalibration() {
		t.Error("should not need recalibration after reset")
	}
	if ks.ConsecutiveRejects() != 0 {
		t.Errorf("consecutive rejects = %d, want 0 after reset", ks.ConsecutiveRejects())
	}
}
