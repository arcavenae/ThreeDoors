package core

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"testing"
)

// --- Door Selection: Fisher-Yates & Diversity ---

func TestSelectDoorsWithRand_FisherYatesPartialShuffle(t *testing.T) {
	t.Parallel()

	// Verify selection returns exactly count tasks from the available pool,
	// and that the diversity-based selection picks high-quality sets.
	pool := poolFromTasks(
		newCategorizedTestTask("t1", "Task 1", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
		newCategorizedTestTask("t2", "Task 2", StatusTodo, TypeAdministrative, EffortMedium, LocationWork),
		newCategorizedTestTask("t3", "Task 3", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
		newCategorizedTestTask("t4", "Task 4", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
		newCategorizedTestTask("t5", "Task 5", StatusTodo, TypeCreative, EffortMedium, LocationHome),
		newCategorizedTestTask("t6", "Task 6", StatusTodo, TypeAdministrative, EffortDeepWork, LocationWork),
	)

	rng := rand.New(rand.NewPCG(99, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	if len(selected) != 3 {
		t.Fatalf("expected 3 doors, got %d", len(selected))
	}

	// All selected tasks must come from the pool
	allIDs := make(map[string]bool)
	for _, task := range pool.GetAllTasks() {
		allIDs[task.ID] = true
	}
	for _, task := range selected {
		if !allIDs[task.ID] {
			t.Errorf("selected task %s not in original pool", task.ID)
		}
	}

	// With 4 types, 3 efforts, 4 locations available, diversity should be good
	score := DiversityScore(selected)
	if score < 6 {
		t.Errorf("expected diversity score >= 6 from diverse pool, got %d", score)
	}
}

func TestSelectDoorsWithRand_CountVariations(t *testing.T) {
	t.Parallel()

	tasks := make([]*Task, 8)
	for i := range tasks {
		tasks[i] = newTestTask(fmt.Sprintf("t%d", i), fmt.Sprintf("Task %d", i), StatusTodo, baseTime)
	}

	tests := []struct {
		name      string
		count     int
		wantCount int
	}{
		{"count 1", 1, 1},
		{"count 2", 2, 2},
		{"count 3 (default doors)", 3, 3},
		{"count equals available", 8, 8},
		{"count exceeds available", 15, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := poolFromTasks(tasks...)
			rng := rand.New(rand.NewPCG(42, 0))
			selected := selectDoorsWithRand(pool, tt.count, rng)
			if len(selected) != tt.wantCount {
				t.Errorf("selectDoorsWithRand(count=%d) returned %d tasks, want %d",
					tt.count, len(selected), tt.wantCount)
			}
		})
	}
}

func TestSelectDoorsWithRand_NoDuplicates(t *testing.T) {
	t.Parallel()

	pool := poolFromTasks(
		newTestTask("t1", "Task 1", StatusTodo, baseTime),
		newTestTask("t2", "Task 2", StatusTodo, baseTime),
		newTestTask("t3", "Task 3", StatusTodo, baseTime),
		newTestTask("t4", "Task 4", StatusTodo, baseTime),
		newTestTask("t5", "Task 5", StatusTodo, baseTime),
	)

	for seed := uint64(0); seed < 50; seed++ {
		freshPool := poolFromTasks(pool.GetAllTasks()...)
		rng := rand.New(rand.NewPCG(seed, 0))
		selected := selectDoorsWithRand(freshPool, 3, rng)

		seen := make(map[string]bool)
		for _, task := range selected {
			if seen[task.ID] {
				t.Errorf("seed %d: duplicate task %s in selection", seed, task.ID)
			}
			seen[task.ID] = true
		}
	}
}

func TestSelectDoorsWithRand_MarksRecentlyShown(t *testing.T) {
	t.Parallel()

	pool := poolFromTasks(
		newTestTask("t1", "Task 1", StatusTodo, baseTime),
		newTestTask("t2", "Task 2", StatusTodo, baseTime),
		newTestTask("t3", "Task 3", StatusTodo, baseTime),
		newTestTask("t4", "Task 4", StatusTodo, baseTime),
		newTestTask("t5", "Task 5", StatusTodo, baseTime),
	)

	rng := rand.New(rand.NewPCG(42, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	for _, task := range selected {
		if !pool.IsRecentlyShown(task.ID) {
			t.Errorf("selected task %s should be marked as recently shown", task.ID)
		}
	}
}

func TestSelectRandomDoors_Baseline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tasks     []Task
		count     int
		exclude   []Task
		wantCount int
		wantNil   bool
	}{
		{
			name:    "nil tasks",
			tasks:   nil,
			count:   3,
			wantNil: true,
		},
		{
			name:    "empty tasks",
			tasks:   []Task{},
			count:   3,
			wantNil: true,
		},
		{
			name:    "negative count",
			tasks:   []Task{{Text: "A"}},
			count:   -1,
			wantNil: true,
		},
		{
			name:      "single task count 1",
			tasks:     []Task{{Text: "A"}},
			count:     1,
			wantCount: 1,
		},
		{
			name:      "more tasks than count",
			tasks:     []Task{{Text: "A"}, {Text: "B"}, {Text: "C"}, {Text: "D"}},
			count:     2,
			wantCount: 2,
		},
		{
			name:      "fewer tasks than count",
			tasks:     []Task{{Text: "A"}, {Text: "B"}},
			count:     5,
			wantCount: 2,
		},
		{
			name:      "with exclusion — enough non-excluded",
			tasks:     []Task{{Text: "A"}, {Text: "B"}, {Text: "C"}, {Text: "D"}, {Text: "E"}},
			count:     3,
			exclude:   []Task{{Text: "A"}, {Text: "B"}},
			wantCount: 3,
		},
		{
			name:      "exclusion fallback — not enough non-excluded",
			tasks:     []Task{{Text: "A"}, {Text: "B"}, {Text: "C"}},
			count:     3,
			exclude:   []Task{{Text: "A"}, {Text: "B"}},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SelectRandomDoors(tt.tasks, tt.count, tt.exclude)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != tt.wantCount {
				t.Errorf("expected %d tasks, got %d", tt.wantCount, len(result))
			}
		})
	}
}

// --- Ring Buffer ---

func TestRingBuffer_Wraparound(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	// Fill buffer to capacity (10 items)
	for i := range 10 {
		pool.MarkRecentlyShown(fmt.Sprintf("task-%d", i))
	}

	// All 10 should be recently shown
	for i := range 10 {
		if !pool.IsRecentlyShown(fmt.Sprintf("task-%d", i)) {
			t.Errorf("task-%d should be recently shown after filling buffer", i)
		}
	}

	// Add one more — should evict task-0 (oldest)
	pool.MarkRecentlyShown("task-10")

	if pool.IsRecentlyShown("task-0") {
		t.Error("task-0 should have been evicted after wraparound")
	}
	if !pool.IsRecentlyShown("task-10") {
		t.Error("task-10 should be recently shown after adding")
	}

	// Tasks 1-9 should still be present
	for i := 1; i <= 9; i++ {
		if !pool.IsRecentlyShown(fmt.Sprintf("task-%d", i)) {
			t.Errorf("task-%d should still be recently shown", i)
		}
	}
}

func TestRingBuffer_FullCycleEviction(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	// Add 20 items — should cycle through buffer twice
	for i := range 20 {
		pool.MarkRecentlyShown(fmt.Sprintf("item-%d", i))
	}

	// Only the last 10 should remain
	for i := range 10 {
		if pool.IsRecentlyShown(fmt.Sprintf("item-%d", i)) {
			t.Errorf("item-%d should have been evicted", i)
		}
	}
	for i := 10; i < 20; i++ {
		if !pool.IsRecentlyShown(fmt.Sprintf("item-%d", i)) {
			t.Errorf("item-%d should be recently shown", i)
		}
	}
}

func TestRingBuffer_EmptyBuffer(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	if pool.IsRecentlyShown("nonexistent") {
		t.Error("empty buffer should not contain any task")
	}
}

func TestRingBuffer_DuplicateMarks(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	// Mark same task multiple times — consumes buffer slots
	for range 5 {
		pool.MarkRecentlyShown("same-task")
	}

	if !pool.IsRecentlyShown("same-task") {
		t.Error("repeatedly marked task should still be recently shown")
	}
}

// --- Status Management: Full Transition Matrix ---

func TestStatusTransition_FullMatrix(t *testing.T) {
	t.Parallel()

	allStatuses := []TaskStatus{
		StatusTodo, StatusBlocked, StatusInProgress,
		StatusInReview, StatusComplete, StatusDeferred, StatusArchived,
	}

	// Expected valid transitions per the state machine
	expected := map[TaskStatus]map[TaskStatus]bool{
		StatusTodo: {
			StatusInProgress: true, StatusBlocked: true,
			StatusComplete: true, StatusDeferred: true, StatusArchived: true,
		},
		StatusBlocked: {
			StatusTodo: true, StatusInProgress: true, StatusComplete: true, StatusDeferred: true,
		},
		StatusInProgress: {
			StatusBlocked: true, StatusInReview: true, StatusComplete: true, StatusDeferred: true,
		},
		StatusInReview: {
			StatusInProgress: true, StatusComplete: true,
		},
		StatusComplete: {StatusTodo: true},
		StatusDeferred: {StatusTodo: true},
		StatusArchived: {},
	}

	for _, from := range allStatuses {
		for _, to := range allStatuses {
			name := fmt.Sprintf("%s_to_%s", from, to)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				got := IsValidTransition(from, to)
				want := expected[from][to]
				if got != want {
					t.Errorf("IsValidTransition(%q, %q) = %v, want %v", from, to, got, want)
				}
			})
		}
	}
}

func TestStatusTransition_SelfTransitionNoOp(t *testing.T) {
	t.Parallel()

	// UpdateStatus treats same-status as no-op (returns nil, no change)
	statuses := []TaskStatus{
		StatusTodo, StatusBlocked, StatusInProgress,
		StatusInReview, StatusComplete, StatusDeferred, StatusArchived,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()
			task := newTestTask("self", "Self transition", status, baseTime)
			// For blocked/in-progress/etc we need valid initial state
			originalUpdatedAt := task.UpdatedAt

			err := task.UpdateStatus(status)
			if err != nil {
				t.Errorf("self-transition for %q should be no-op, got error: %v", status, err)
			}
			if task.Status != status {
				t.Errorf("status should remain %q, got %q", status, task.Status)
			}
			if task.UpdatedAt != originalUpdatedAt {
				t.Error("UpdatedAt should not change on self-transition")
			}
		})
	}
}

func TestStatusTransition_CompletedAtSetForTerminalStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		from          TaskStatus
		to            TaskStatus
		wantCompleted bool
	}{
		{"todo to complete", StatusTodo, StatusComplete, true},
		{"todo to archived", StatusTodo, StatusArchived, true},
		{"todo to in-progress", StatusTodo, StatusInProgress, false},
		{"todo to blocked", StatusTodo, StatusBlocked, false},
		{"todo to deferred", StatusTodo, StatusDeferred, false},
		{"in-progress to complete", StatusInProgress, StatusComplete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := newTestTask("term", "Terminal test", tt.from, baseTime)
			err := task.UpdateStatus(tt.to)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantCompleted && task.CompletedAt == nil {
				t.Error("CompletedAt should be set for terminal state")
			}
			if !tt.wantCompleted && task.CompletedAt != nil {
				t.Error("CompletedAt should not be set for non-terminal state")
			}
		})
	}
}

func TestStatusTransition_BlockerClearedOnNonBlocked(t *testing.T) {
	t.Parallel()

	task := newTestTask("blk", "Blocked task", StatusTodo, baseTime)

	// Move to blocked and set blocker
	if err := task.UpdateStatus(StatusBlocked); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := task.SetBlocker("waiting on dependency"); err != nil {
		t.Fatalf("unexpected error setting blocker: %v", err)
	}
	if task.Blocker == "" {
		t.Fatal("blocker should be set")
	}

	// Move to in-progress — blocker should clear
	if err := task.UpdateStatus(StatusInProgress); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Blocker != "" {
		t.Errorf("blocker should be cleared after moving to in-progress, got %q", task.Blocker)
	}
}

func TestSetBlocker_OnlyWhenBlocked(t *testing.T) {
	t.Parallel()

	statuses := []TaskStatus{
		StatusTodo, StatusInProgress, StatusInReview,
		StatusComplete, StatusDeferred, StatusArchived,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()
			task := newTestTask("sb", "Test", status, baseTime)
			err := task.SetBlocker("some reason")
			if err == nil {
				t.Errorf("SetBlocker should fail for status %q", status)
			}
		})
	}

	// Should succeed for blocked
	task := newTestTask("sb-ok", "Test", StatusBlocked, baseTime)
	if err := task.SetBlocker("valid reason"); err != nil {
		t.Errorf("SetBlocker should succeed for blocked status: %v", err)
	}
}

func TestValidateStatus_AllValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		valid bool
	}{
		{"todo", true},
		{"blocked", true},
		{"in-progress", true},
		{"in-review", true},
		{"complete", true},
		{"deferred", true},
		{"archived", true},
		{"", false},
		{"invalid", false},
		{"COMPLETE", false},
		{"Todo", false},
		{"done", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			t.Parallel()
			err := ValidateStatus(tt.input)
			if tt.valid && err != nil {
				t.Errorf("ValidateStatus(%q) returned error: %v", tt.input, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateStatus(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestGetValidTransitions_AllStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status    TaskStatus
		wantCount int
	}{
		{StatusTodo, 5},       // in-progress, blocked, complete, deferred, archived
		{StatusBlocked, 4},    // todo, in-progress, complete, deferred
		{StatusInProgress, 4}, // blocked, in-review, complete, deferred
		{StatusInReview, 2},   // in-progress, complete
		{StatusComplete, 1},   // todo (undo)
		{StatusDeferred, 1},   // todo
		{StatusArchived, 0},   // terminal
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			transitions := GetValidTransitions(tt.status)
			if len(transitions) != tt.wantCount {
				t.Errorf("GetValidTransitions(%q) returned %d transitions, want %d",
					tt.status, len(transitions), tt.wantCount)
			}
		})
	}
}

func TestGetValidTransitions_UnknownStatus(t *testing.T) {
	t.Parallel()

	transitions := GetValidTransitions(TaskStatus("nonexistent"))
	if transitions != nil {
		t.Errorf("expected nil for unknown status, got %v", transitions)
	}
}

// --- Task Pool: CRUD & Filtering ---

func TestTaskPool_CRUD_Lifecycle(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	// Add
	task := newTestTask("crud-1", "CRUD test task", StatusTodo, baseTime)
	pool.AddTask(task)
	if pool.Count() != 1 {
		t.Fatalf("expected count 1 after add, got %d", pool.Count())
	}

	// Get
	got := pool.GetTask("crud-1")
	if got == nil {
		t.Fatal("expected to get task back")
	}
	if got.Text != "CRUD test task" {
		t.Errorf("expected text %q, got %q", "CRUD test task", got.Text)
	}

	// Update
	task.Text = "Updated text"
	pool.UpdateTask(task)
	got = pool.GetTask("crud-1")
	if got.Text != "Updated text" {
		t.Errorf("expected updated text, got %q", got.Text)
	}
	if pool.Count() != 1 {
		t.Errorf("count should still be 1 after update, got %d", pool.Count())
	}

	// Remove
	pool.RemoveTask("crud-1")
	if pool.GetTask("crud-1") != nil {
		t.Error("expected task to be removed")
	}
	if pool.Count() != 0 {
		t.Errorf("expected count 0 after remove, got %d", pool.Count())
	}
}

func TestTaskPool_GetNonExistent(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	if pool.GetTask("nonexistent") != nil {
		t.Error("expected nil for nonexistent task")
	}
}

func TestTaskPool_RemoveNonExistent(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	pool.AddTask(newTestTask("exists", "Task", StatusTodo, baseTime))

	// Removing nonexistent should not panic or affect existing tasks
	pool.RemoveTask("nonexistent")
	if pool.Count() != 1 {
		t.Errorf("removing nonexistent should not affect count, got %d", pool.Count())
	}
}

func TestTaskPool_UpdateOverwrites(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	original := newCategorizedTestTask("ow-1", "Original", StatusTodo, TypeCreative, EffortQuickWin, LocationHome)
	pool.AddTask(original)

	replacement := newCategorizedTestTask("ow-1", "Replaced", StatusInProgress, TypeTechnical, EffortDeepWork, LocationWork)
	pool.UpdateTask(replacement)

	got := pool.GetTask("ow-1")
	if got.Text != "Replaced" {
		t.Errorf("expected text %q, got %q", "Replaced", got.Text)
	}
	if got.Status != StatusInProgress {
		t.Errorf("expected status %q, got %q", StatusInProgress, got.Status)
	}
	if got.Type != TypeTechnical {
		t.Errorf("expected type %q, got %q", TypeTechnical, got.Type)
	}
}

func TestTaskPool_GetAllTasks(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	ids := []string{"a", "b", "c", "d", "e"}
	for _, id := range ids {
		pool.AddTask(newTestTask(id, "Task "+id, StatusTodo, baseTime))
	}

	all := pool.GetAllTasks()
	if len(all) != 5 {
		t.Fatalf("expected 5 tasks, got %d", len(all))
	}

	gotIDs := make([]string, len(all))
	for i, task := range all {
		gotIDs[i] = task.ID
	}
	sort.Strings(gotIDs)
	sort.Strings(ids)
	for i := range ids {
		if gotIDs[i] != ids[i] {
			t.Errorf("expected ID %q at position %d, got %q", ids[i], i, gotIDs[i])
		}
	}
}

func TestTaskPool_GetTasksByStatus_AllStatuses(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	pool.AddTask(newTestTask("t1", "Todo", StatusTodo, baseTime))
	pool.AddTask(newTestTask("t2", "Blocked", StatusBlocked, baseTime))
	pool.AddTask(newTestTask("t3", "InProgress", StatusInProgress, baseTime))
	pool.AddTask(newTestTask("t4", "InReview", StatusInReview, baseTime))
	pool.AddTask(newTestTask("t5", "Complete", StatusComplete, baseTime))
	pool.AddTask(newTestTask("t6", "Deferred", StatusDeferred, baseTime))
	pool.AddTask(newTestTask("t7", "Archived", StatusArchived, baseTime))
	pool.AddTask(newTestTask("t8", "Todo2", StatusTodo, baseTime))

	tests := []struct {
		status    TaskStatus
		wantCount int
	}{
		{StatusTodo, 2},
		{StatusBlocked, 1},
		{StatusInProgress, 1},
		{StatusInReview, 1},
		{StatusComplete, 1},
		{StatusDeferred, 1},
		{StatusArchived, 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			result := pool.GetTasksByStatus(tt.status)
			if len(result) != tt.wantCount {
				t.Errorf("GetTasksByStatus(%q) returned %d, want %d",
					tt.status, len(result), tt.wantCount)
			}
		})
	}
}

func TestTaskPool_GetTasksByStatus_Empty(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	result := pool.GetTasksByStatus(StatusTodo)
	if len(result) != 0 {
		t.Errorf("expected 0 tasks from empty pool, got %d", len(result))
	}
}

func TestTaskPool_GetAvailableForDoors_EligibilityRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		statuses  []TaskStatus
		wantCount int
	}{
		{
			"all eligible statuses",
			[]TaskStatus{StatusTodo, StatusBlocked, StatusInProgress},
			3,
		},
		{
			"excludes complete",
			[]TaskStatus{StatusTodo, StatusComplete, StatusTodo},
			2,
		},
		{
			"excludes deferred",
			[]TaskStatus{StatusTodo, StatusDeferred, StatusTodo},
			2,
		},
		{
			"excludes archived",
			[]TaskStatus{StatusTodo, StatusArchived, StatusTodo},
			2,
		},
		{
			"excludes in-review",
			[]TaskStatus{StatusTodo, StatusInReview, StatusTodo, StatusTodo},
			3,
		},
		{
			"all non-eligible",
			[]TaskStatus{StatusComplete, StatusDeferred, StatusArchived, StatusInReview},
			0,
		},
		{
			"mixed eligible and non-eligible",
			[]TaskStatus{StatusTodo, StatusComplete, StatusBlocked, StatusArchived, StatusInProgress},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := NewTaskPool()
			for i, status := range tt.statuses {
				pool.AddTask(newTestTask(fmt.Sprintf("t%d", i), fmt.Sprintf("Task %d", i), status, baseTime))
			}
			available := pool.GetAvailableForDoors()
			if len(available) != tt.wantCount {
				t.Errorf("expected %d available, got %d", tt.wantCount, len(available))
			}
		})
	}
}

func TestTaskPool_GetAvailableForDoors_RecentlyShownFallback(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	pool.AddTask(newTestTask("t1", "Task 1", StatusTodo, baseTime))
	pool.AddTask(newTestTask("t2", "Task 2", StatusTodo, baseTime))

	// Mark both as recently shown
	pool.MarkRecentlyShown("t1")
	pool.MarkRecentlyShown("t2")

	// With < 3 non-recent eligible tasks, should fall back to include recently shown
	available := pool.GetAvailableForDoors()
	if len(available) != 2 {
		t.Errorf("expected 2 available (fallback includes recently shown), got %d", len(available))
	}
}

func TestTaskPool_GetAvailableForDoors_NoFallbackWhenEnough(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	for i := range 6 {
		pool.AddTask(newTestTask(fmt.Sprintf("t%d", i), fmt.Sprintf("Task %d", i), StatusTodo, baseTime))
	}

	// Mark 3 as recently shown — 3 remain non-recent (>= 3 threshold)
	pool.MarkRecentlyShown("t0")
	pool.MarkRecentlyShown("t1")
	pool.MarkRecentlyShown("t2")

	available := pool.GetAvailableForDoors()
	if len(available) != 3 {
		t.Errorf("expected 3 available (no fallback needed), got %d", len(available))
	}

	// Verify recently shown are excluded
	for _, task := range available {
		if task.ID == "t0" || task.ID == "t1" || task.ID == "t2" {
			t.Errorf("recently shown task %s should be excluded", task.ID)
		}
	}
}

func TestTaskPool_GetAvailableForDoors_EmptyPool(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	available := pool.GetAvailableForDoors()
	if len(available) != 0 {
		t.Errorf("expected 0 available from empty pool, got %d", len(available))
	}
}

func TestTaskPool_CountAccuracy(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	if pool.Count() != 0 {
		t.Errorf("expected 0, got %d", pool.Count())
	}

	for i := range 10 {
		pool.AddTask(newTestTask(fmt.Sprintf("c%d", i), "Task", StatusTodo, baseTime))
		if pool.Count() != i+1 {
			t.Errorf("after adding %d tasks, count = %d", i+1, pool.Count())
		}
	}

	for i := range 5 {
		pool.RemoveTask(fmt.Sprintf("c%d", i))
		if pool.Count() != 10-(i+1) {
			t.Errorf("after removing %d tasks, count = %d, want %d", i+1, pool.Count(), 10-(i+1))
		}
	}
}

// --- Task Validation ---

func TestTask_Validate_Baseline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			"valid todo task",
			newTestTask("v1", "Valid task", StatusTodo, baseTime),
			false,
		},
		{
			"valid categorized task",
			newCategorizedTestTask("v2", "Categorized", StatusTodo, TypeCreative, EffortMedium, LocationHome),
			false,
		},
		{
			"missing ID",
			&Task{Text: "No ID", Status: StatusTodo, CreatedAt: baseTime, UpdatedAt: baseTime},
			true,
		},
		{
			"empty text",
			&Task{ID: "v3", Text: "", Status: StatusTodo, CreatedAt: baseTime, UpdatedAt: baseTime},
			true,
		},
		{
			"invalid status",
			&Task{ID: "v4", Text: "Task", Status: TaskStatus("invalid"), CreatedAt: baseTime, UpdatedAt: baseTime},
			true,
		},
		{
			"zero createdAt",
			&Task{ID: "v5", Text: "Task", Status: StatusTodo},
			true,
		},
		{
			"updatedAt before createdAt",
			&Task{ID: "v6", Text: "Task", Status: StatusTodo, CreatedAt: laterTime, UpdatedAt: baseTime},
			true,
		},
		{
			"completedAt on non-terminal status",
			func() *Task {
				t := newTestTask("v7", "Task", StatusTodo, baseTime)
				now := laterTime
				t.CompletedAt = &now
				return t
			}(),
			true,
		},
		{
			"valid complete with completedAt",
			newCompletedTestTask("v8", "Done task", laterTime),
			false,
		},
		{
			"invalid task type",
			func() *Task {
				t := newTestTask("v9", "Task", StatusTodo, baseTime)
				t.Type = TaskType("invalid-type")
				return t
			}(),
			true,
		},
		{
			"invalid effort",
			func() *Task {
				t := newTestTask("v10", "Task", StatusTodo, baseTime)
				t.Effort = TaskEffort("invalid-effort")
				return t
			}(),
			true,
		},
		{
			"invalid location",
			func() *Task {
				t := newTestTask("v11", "Task", StatusTodo, baseTime)
				t.Location = TaskLocation("invalid-loc")
				return t
			}(),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.task.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// --- Diversity Score ---

func TestDiversityScore_Baseline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tasks []*Task
		want  int
	}{
		{"nil input", nil, 0},
		{"empty input", []*Task{}, 0},
		{
			"single uncategorized",
			[]*Task{newTestTask("1", "t", StatusTodo, baseTime)},
			3,
		},
		{
			"three identical categories",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
				newCategorizedTestTask("3", "t3", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
			},
			3,
		},
		{
			"maximum diversity 3 tasks",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeAdministrative, EffortMedium, LocationWork),
				newCategorizedTestTask("3", "t3", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
			},
			9,
		},
		{
			"partial diversity — same effort",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortMedium, LocationHome),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
			},
			5, // 2 types + 1 effort + 2 locations
		},
		{
			"mix of categorized and uncategorized",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
				newTestTask("2", "t2", StatusTodo, baseTime), // uncategorized = ""
			},
			6, // 2 types + 2 efforts + 2 locations (empty counts as unique)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DiversityScore(tt.tasks)
			if got != tt.want {
				t.Errorf("DiversityScore() = %d, want %d", got, tt.want)
			}
		})
	}
}
