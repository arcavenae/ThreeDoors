package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMilestoneChecker_CheckMilestones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		totalTasks    int
		currentStreak int
		sessionCount  int
		shown         []string
		wantID        string
		wantNil       bool
	}{
		{
			name:         "first session triggers welcome",
			sessionCount: 1,
			wantID:       "first-session",
		},
		{
			name:         "no sessions no milestone",
			sessionCount: 0,
			wantNil:      true,
		},
		{
			name:         "first session already shown returns nil",
			sessionCount: 1,
			shown:        []string{"first-session"},
			wantNil:      true,
		},
		{
			name:       "50 tasks triggers milestone",
			totalTasks: 50,
			shown:      []string{"first-session"},
			wantID:     "50-tasks",
		},
		{
			name:       "100 tasks triggers milestone over 50",
			totalTasks: 100,
			shown:      []string{"first-session"},
			wantID:     "100-tasks",
		},
		{
			name:       "100 tasks with 100-tasks shown falls to 50-tasks",
			totalTasks: 100,
			shown:      []string{"first-session", "100-tasks"},
			wantID:     "50-tasks",
		},
		{
			name:          "10 day streak triggers milestone",
			currentStreak: 10,
			shown:         []string{"first-session"},
			wantID:        "10-day-streak",
		},
		{
			name:          "9 day streak does not trigger",
			currentStreak: 9,
			shown:         []string{"first-session"},
			wantNil:       true,
		},
		{
			name:       "49 tasks does not trigger 50",
			totalTasks: 49,
			shown:      []string{"first-session"},
			wantNil:    true,
		},
		{
			name:          "all milestones shown returns nil",
			totalTasks:    200,
			currentStreak: 20,
			sessionCount:  50,
			shown:         []string{"first-session", "50-tasks", "100-tasks", "10-day-streak"},
			wantNil:       true,
		},
		{
			name:          "highest priority unshown wins",
			totalTasks:    100,
			currentStreak: 10,
			sessionCount:  5,
			shown:         []string{},
			wantID:        "first-session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			mc := NewMilestoneChecker(dir)
			mc.data.Shown = tt.shown

			got := mc.CheckMilestones(tt.totalTasks, tt.currentStreak, tt.sessionCount)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got milestone %q", got.ID)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected milestone %q, got nil", tt.wantID)
				return
			}
			if got.ID != tt.wantID {
				t.Errorf("expected milestone %q, got %q", tt.wantID, got.ID)
			}
			if got.Message == "" {
				t.Error("milestone message should not be empty")
			}
		})
	}
}

func TestMilestoneChecker_MarkShown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mc := NewMilestoneChecker(dir)

	// Mark first milestone
	if err := mc.MarkShown("first-session"); err != nil {
		t.Fatalf("MarkShown: %v", err)
	}

	if !mc.IsShown("first-session") {
		t.Error("expected first-session to be shown")
	}

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(dir, "milestones.json"))
	if err != nil {
		t.Fatalf("read milestones.json: %v", err)
		return
	}

	var md MilestoneData
	if err := json.Unmarshal(data, &md); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(md.Shown) != 1 || md.Shown[0] != "first-session" {
		t.Errorf("expected [first-session], got %v", md.Shown)
	}
}

func TestMilestoneChecker_MarkShownIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mc := NewMilestoneChecker(dir)

	if err := mc.MarkShown("first-session"); err != nil {
		t.Fatalf("first MarkShown: %v", err)
	}
	if err := mc.MarkShown("first-session"); err != nil {
		t.Fatalf("second MarkShown: %v", err)
	}

	if len(mc.data.Shown) != 1 {
		t.Errorf("expected 1 entry, got %d", len(mc.data.Shown))
	}
}

func TestMilestoneChecker_PersistsAcrossInstantiations(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// First instance marks a milestone
	mc1 := NewMilestoneChecker(dir)
	if err := mc1.MarkShown("50-tasks"); err != nil {
		t.Fatalf("MarkShown: %v", err)
	}

	// Second instance loads and sees it
	mc2 := NewMilestoneChecker(dir)
	if err := mc2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !mc2.IsShown("50-tasks") {
		t.Error("expected 50-tasks to persist across instances")
	}

	// CheckMilestones with 50 tasks should not return 50-tasks (already shown)
	// It may return first-session since that's unshown and session count qualifies.
	got := mc2.CheckMilestones(50, 0, 5)
	if got != nil && got.ID == "50-tasks" {
		t.Error("50-tasks was already shown, should not be returned")
	}
}

func TestMilestoneChecker_LoadNonExistentFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mc := NewMilestoneChecker(dir)

	if err := mc.Load(); err != nil {
		t.Fatalf("Load should succeed for non-existent file: %v", err)
	}

	if len(mc.data.Shown) != 0 {
		t.Errorf("expected empty shown list, got %v", mc.data.Shown)
	}
}

func TestMilestoneChecker_AtomicWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mc := NewMilestoneChecker(dir)

	if err := mc.MarkShown("first-session"); err != nil {
		t.Fatalf("MarkShown: %v", err)
	}

	// Verify no .tmp file remains
	tmpPath := filepath.Join(dir, "milestones.json.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temporary file should not exist after successful write")
	}

	// Verify final file exists and is valid JSON
	data, err := os.ReadFile(filepath.Join(dir, "milestones.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
		return
	}

	var md MilestoneData
	if err := json.Unmarshal(data, &md); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestMilestoneChecker_OnlyHighestPriorityShown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mc := NewMilestoneChecker(dir)

	// User qualifies for multiple milestones
	got := mc.CheckMilestones(100, 10, 5)
	if got == nil {
		t.Fatal("expected a milestone")
		return
	}

	// First-session is highest priority
	if got.ID != "first-session" {
		t.Errorf("expected first-session (highest priority), got %q", got.ID)
	}

	// Mark it shown, next should be 100-tasks (higher priority than 50-tasks and 10-day-streak)
	mc.data.Shown = []string{"first-session"}
	got = mc.CheckMilestones(100, 10, 5)
	if got == nil {
		t.Fatal("expected a milestone after marking first-session")
		return
	}
	if got.ID != "100-tasks" {
		t.Errorf("expected 100-tasks, got %q", got.ID)
	}
}
