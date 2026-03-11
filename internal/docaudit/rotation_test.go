package docaudit

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNextMode_FirstRun(t *testing.T) {
	t.Parallel()
	state := RotationState{}
	mode, ready := NextMode(state, time.Now().UTC())
	if !ready {
		t.Fatal("expected ready on first run")
	}
	if mode != ModeDocConsistency {
		t.Errorf("mode = %q, want %q", mode, ModeDocConsistency)
	}
}

func TestNextMode_Rotation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		lastMode  RotationMode
		lastRun   time.Time
		wantMode  RotationMode
		wantReady bool
	}{
		{ModeDocConsistency, now.Add(-5 * time.Hour), ModeConflictAnalysis, true},
		{ModeConflictAnalysis, now.Add(-5 * time.Hour), ModeCIAnalysis, true},
		{ModeCIAnalysis, now.Add(-5 * time.Hour), ModeProcessWaste, true},
		{ModeProcessWaste, now.Add(-5 * time.Hour), ModeDocConsistency, true},
		// Not enough time passed.
		{ModeDocConsistency, now.Add(-1 * time.Hour), "", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.lastMode), func(t *testing.T) {
			t.Parallel()
			state := RotationState{
				LastMode: tt.lastMode,
				LastRun:  tt.lastRun,
			}
			mode, ready := NextMode(state, now)
			if ready != tt.wantReady {
				t.Errorf("ready = %v, want %v", ready, tt.wantReady)
			}
			if mode != tt.wantMode {
				t.Errorf("mode = %q, want %q", mode, tt.wantMode)
			}
		})
	}
}

func TestRecordRun(t *testing.T) {
	t.Parallel()
	state := RotationState{}
	now := time.Now().UTC()

	RecordRun(&state, ModeDocConsistency, now, true)

	if state.LastMode != ModeDocConsistency {
		t.Errorf("LastMode = %q, want %q", state.LastMode, ModeDocConsistency)
	}
	if !state.LastRun.Equal(now) {
		t.Errorf("LastRun = %v, want %v", state.LastRun, now)
	}
	if state.CycleCount != 1 {
		t.Errorf("CycleCount = %d, want 1", state.CycleCount)
	}
	if len(state.ModeHistory) != 1 {
		t.Fatalf("ModeHistory len = %d, want 1", len(state.ModeHistory))
	}
	if !state.ModeHistory[0].Clean {
		t.Error("expected clean run")
	}
}

func TestRecordRun_HistoryTruncation(t *testing.T) {
	t.Parallel()
	state := RotationState{}
	now := time.Now().UTC()

	// Add 25 runs — should keep only the last 20.
	for i := range 25 {
		RecordRun(&state, ModeDocConsistency, now.Add(time.Duration(i)*time.Hour), true)
	}

	if len(state.ModeHistory) != 20 {
		t.Errorf("ModeHistory len = %d, want 20", len(state.ModeHistory))
	}
}

func TestSaveAndLoadRotationState(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "rotation.json")

	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	original := RotationState{
		LastMode:   ModeDocConsistency,
		LastRun:    now,
		CycleCount: 5,
		ModeHistory: []ModeRun{
			{Mode: ModeDocConsistency, Timestamp: now, Clean: true},
		},
	}

	if err := SaveRotationState(path, original); err != nil {
		t.Fatalf("SaveRotationState() error: %v", err)
	}

	loaded, err := LoadRotationState(path)
	if err != nil {
		t.Fatalf("LoadRotationState() error: %v", err)
	}

	if loaded.LastMode != original.LastMode {
		t.Errorf("LastMode = %q, want %q", loaded.LastMode, original.LastMode)
	}
	if loaded.CycleCount != original.CycleCount {
		t.Errorf("CycleCount = %d, want %d", loaded.CycleCount, original.CycleCount)
	}
}

func TestLoadRotationState_MissingFile(t *testing.T) {
	t.Parallel()
	state, err := LoadRotationState("/nonexistent/rotation.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if state.LastMode != "" {
		t.Errorf("LastMode = %q, want empty", state.LastMode)
	}
}
