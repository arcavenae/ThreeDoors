package core

import (
	"testing"
	"time"
)

// makeTaskWithClock creates a task with specific vector clock and field versions for testing.
func makeTaskWithClock(id, text, deviceID string, clock VectorClock, fields map[string]FieldVersion) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:            id,
		Text:          text,
		Status:        StatusTodo,
		CreatedAt:     now,
		UpdatedAt:     now,
		FieldVersions: fields,
		VectorClock:   map[string]uint64(clock),
	}
}

// makeBaseTask creates a simple task for use as a merge base.
func makeBaseTask(id, text string) *Task {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &Task{
		ID:        id,
		Text:      text,
		Status:    StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestThreeWayMergeTask_NonOverlapping(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")
	base.Effort = EffortQuickWin

	local := *base
	local.Text = "Local changed text"
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceA", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	local.VectorClock = map[string]uint64{"deviceA": 1}

	remote := *base
	remote.Effort = EffortDeepWork
	remote.FieldVersions = map[string]FieldVersion{
		"effort": {DeviceID: "deviceB", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	remote.VectorClock = map[string]uint64{"deviceB": 1}

	result := ThreeWayMergeTask(base, &local, &remote)

	if result.Task.Text != "Local changed text" {
		t.Errorf("expected local text, got %q", result.Task.Text)
	}
	if result.Task.Effort != EffortDeepWork {
		t.Errorf("expected remote effort %q, got %q", EffortDeepWork, result.Task.Effort)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected no conflicts for non-overlapping merge, got %d", len(result.Conflicts))
	}
}

func TestThreeWayMergeTask_OverlappingCausal(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")

	local := *base
	local.Text = "A's text"
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceA", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 3},
	}
	local.VectorClock = map[string]uint64{"deviceA": 3, "deviceB": 2}

	remote := *base
	remote.Text = "B's text"
	remote.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceB", UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), Version: 2},
	}
	remote.VectorClock = map[string]uint64{"deviceA": 2, "deviceB": 2}

	result := ThreeWayMergeTask(base, &local, &remote)

	// A's clock {A:3,B:2} happened-after B's clock {A:2,B:2}
	if result.Task.Text != "A's text" {
		t.Errorf("expected A's text (causal winner), got %q", result.Task.Text)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict detail, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Reason != "causal" {
		t.Errorf("expected reason 'causal', got %q", result.Conflicts[0].Reason)
	}
	if result.Conflicts[0].Winner != "local" {
		t.Errorf("expected winner 'local', got %q", result.Conflicts[0].Winner)
	}
}

func TestThreeWayMergeTask_OverlappingConcurrentTimestamp(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")

	local := *base
	local.Text = "A's text"
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceA", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	local.VectorClock = map[string]uint64{"deviceA": 3, "deviceB": 2}

	remote := *base
	remote.Text = "B's text"
	remote.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceB", UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	remote.VectorClock = map[string]uint64{"deviceA": 2, "deviceB": 4}

	result := ThreeWayMergeTask(base, &local, &remote)

	// Clocks are concurrent {A:3,B:2} vs {A:2,B:4}
	// Remote timestamp (Jan 3) is later than local (Jan 2)
	if result.Task.Text != "B's text" {
		t.Errorf("expected B's text (later timestamp), got %q", result.Task.Text)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Reason != "timestamp" {
		t.Errorf("expected reason 'timestamp', got %q", result.Conflicts[0].Reason)
	}
}

func TestThreeWayMergeTask_OverlappingTimestampTie(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")
	sameTime := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

	local := *base
	local.Text = "aaa's text"
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "aaa", UpdatedAt: sameTime, Version: 1},
	}
	local.VectorClock = map[string]uint64{"aaa": 1}

	remote := *base
	remote.Text = "bbb's text"
	remote.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "bbb", UpdatedAt: sameTime, Version: 1},
	}
	remote.VectorClock = map[string]uint64{"bbb": 1}

	result := ThreeWayMergeTask(base, &local, &remote)

	// Concurrent clocks, same timestamp — tiebreaker: "bbb" > "aaa" lexicographically
	if result.Task.Text != "bbb's text" {
		t.Errorf("expected bbb's text (lexicographic tiebreaker), got %q", result.Task.Text)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Reason != "tiebreaker" {
		t.Errorf("expected reason 'tiebreaker', got %q", result.Conflicts[0].Reason)
	}
}

func TestThreeWayMergeTask_ThreeWayMerge(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")
	base.Status = StatusTodo

	local := *base
	local.Text = "A modified text"
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "deviceA", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	local.VectorClock = map[string]uint64{"deviceA": 1}

	remote := *base
	remote.Status = StatusInProgress
	remote.FieldVersions = map[string]FieldVersion{
		"status": {DeviceID: "deviceB", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 1},
	}
	remote.VectorClock = map[string]uint64{"deviceB": 1}

	result := ThreeWayMergeTask(base, &local, &remote)

	if result.Task.Text != "A modified text" {
		t.Errorf("expected A's text, got %q", result.Task.Text)
	}
	if result.Task.Status != StatusInProgress {
		t.Errorf("expected B's status %q, got %q", StatusInProgress, result.Task.Status)
	}
}

func TestThreeWayMergeTaskLists_ConcurrentCreate(t *testing.T) {
	t.Parallel()

	base := []*Task{}
	local := []*Task{makeTaskWithClock("task-Y", "Task Y from A", "deviceA", VectorClock{"deviceA": 1}, nil)}
	remote := []*Task{makeTaskWithClock("task-Z", "Task Z from B", "deviceB", VectorClock{"deviceB": 1}, nil)}

	outcome := ThreeWayMergeTaskLists(base, local, remote)

	if len(outcome.MergedTasks) != 2 {
		t.Fatalf("expected 2 merged tasks, got %d", len(outcome.MergedTasks))
	}

	foundY, foundZ := false, false
	for _, task := range outcome.MergedTasks {
		if task.ID == "task-Y" {
			foundY = true
		}
		if task.ID == "task-Z" {
			foundZ = true
		}
	}
	if !foundY {
		t.Error("missing task Y from local create")
	}
	if !foundZ {
		t.Error("missing task Z from remote create")
	}
}

func TestThreeWayMergeTaskLists_ConcurrentDeleteVsModify(t *testing.T) {
	t.Parallel()

	base := []*Task{makeBaseTask("task-X", "Original task X")}

	// Local deletes task X (absent from local list)
	local := []*Task{}

	// Remote modifies task X
	remoteTask := makeBaseTask("task-X", "Modified task X")
	remoteTask.Text = "Modified by B"
	remote := []*Task{remoteTask}

	outcome := ThreeWayMergeTaskLists(base, local, remote)

	// Modify-wins: task X should be preserved with B's changes
	if len(outcome.MergedTasks) != 1 {
		t.Fatalf("expected 1 merged task (modify-wins), got %d", len(outcome.MergedTasks))
	}
	if outcome.MergedTasks[0].Text != "Modified by B" {
		t.Errorf("expected modified text, got %q", outcome.MergedTasks[0].Text)
	}
}

func TestThreeWayMergeTaskLists_BothDelete(t *testing.T) {
	t.Parallel()

	base := []*Task{
		makeBaseTask("task-1", "Task 1"),
		makeBaseTask("task-2", "Task 2"),
		makeBaseTask("task-3", "Task 3"),
	}
	local := []*Task{}
	remote := []*Task{}

	outcome := ThreeWayMergeTaskLists(base, local, remote)

	if len(outcome.MergedTasks) != 0 {
		t.Errorf("expected 0 merged tasks when both sides delete all, got %d", len(outcome.MergedTasks))
	}
}

func TestThreeWayMergeTaskLists_ConcurrentAddsToEmpty(t *testing.T) {
	t.Parallel()

	base := []*Task{}
	local := []*Task{makeBaseTask("task-Y", "Task Y")}
	remote := []*Task{makeBaseTask("task-Z", "Task Z")}

	outcome := ThreeWayMergeTaskLists(base, local, remote)

	if len(outcome.MergedTasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(outcome.MergedTasks))
	}
}

func TestThreeWayMergeTask_LegacyNilFieldVersions(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original")

	local := *base
	local.Text = "Local change"
	// No FieldVersions or VectorClock (legacy format)

	remote := *base
	remote.Text = "Remote change"
	// No FieldVersions or VectorClock (legacy format)

	result := ThreeWayMergeTask(base, &local, &remote)

	// Both have nil clocks → concurrent → both have zero-time timestamps → tiebreaker
	// Should not panic
	if result.Task == nil {
		t.Fatal("expected non-nil merged task")
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
}

func TestThreeWayMergeTask_SameChangeBothSides(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original text")

	local := *base
	local.Text = "Same new text"

	remote := *base
	remote.Text = "Same new text"

	result := ThreeWayMergeTask(base, &local, &remote)

	if result.Task.Text != "Same new text" {
		t.Errorf("expected 'Same new text', got %q", result.Task.Text)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected no conflicts when both sides make same change, got %d", len(result.Conflicts))
	}
}

func TestResolveFieldConflict_AllCascadeLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		localVer    FieldVersion
		remoteVer   FieldVersion
		localClock  VectorClock
		remoteClock VectorClock
		wantWinner  string
		wantReason  string
	}{
		{
			name:        "causal - local after",
			localVer:    FieldVersion{DeviceID: "A", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			remoteVer:   FieldVersion{DeviceID: "B", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
			localClock:  VectorClock{"A": 3, "B": 2},
			remoteClock: VectorClock{"A": 2, "B": 2},
			wantWinner:  "local",
			wantReason:  "causal",
		},
		{
			name:        "causal - remote after",
			localVer:    FieldVersion{DeviceID: "A", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
			remoteVer:   FieldVersion{DeviceID: "B", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			localClock:  VectorClock{"A": 2, "B": 2},
			remoteClock: VectorClock{"A": 3, "B": 2},
			wantWinner:  "remote",
			wantReason:  "causal",
		},
		{
			name:        "concurrent - timestamp wins",
			localVer:    FieldVersion{DeviceID: "A", UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
			remoteVer:   FieldVersion{DeviceID: "B", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			localClock:  VectorClock{"A": 3, "B": 2},
			remoteClock: VectorClock{"A": 2, "B": 4},
			wantWinner:  "local",
			wantReason:  "timestamp",
		},
		{
			name:        "concurrent - device tiebreaker",
			localVer:    FieldVersion{DeviceID: "aaa", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			remoteVer:   FieldVersion{DeviceID: "bbb", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			localClock:  VectorClock{"aaa": 1},
			remoteClock: VectorClock{"bbb": 1},
			wantWinner:  "remote",
			wantReason:  "tiebreaker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			winner, reason := ResolveFieldConflict(tt.localVer, tt.remoteVer, tt.localClock, tt.remoteClock)
			if winner != tt.wantWinner {
				t.Errorf("winner: expected %q, got %q", tt.wantWinner, winner)
			}
			if reason != tt.wantReason {
				t.Errorf("reason: expected %q, got %q", tt.wantReason, reason)
			}
		})
	}
}

func TestThreeWayMergeTask_VectorClockMerged(t *testing.T) {
	t.Parallel()

	base := makeBaseTask("task-1", "Original")

	local := *base
	local.Text = "A's text"
	local.VectorClock = map[string]uint64{"A": 3, "B": 2}
	local.FieldVersions = map[string]FieldVersion{
		"text": {DeviceID: "A", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 3},
	}

	remote := *base
	remote.Context = "B's context"
	remote.VectorClock = map[string]uint64{"A": 1, "B": 5}
	remote.FieldVersions = map[string]FieldVersion{
		"context": {DeviceID: "B", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Version: 5},
	}

	result := ThreeWayMergeTask(base, &local, &remote)

	// Vector clocks should be merged: max(A: 3,1)=3, max(B: 2,5)=5
	vc := VectorClock(result.Task.VectorClock)
	if vc["A"] != 3 {
		t.Errorf("expected merged clock A=3, got A=%d", vc["A"])
	}
	if vc["B"] != 5 {
		t.Errorf("expected merged clock B=5, got B=%d", vc["B"])
	}
}
