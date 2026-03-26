package tui

import (
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/dispatch"
)

// --- SourceBadgeLabel ---

func TestSourceBadgeLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{"textfile", "textfile", "TXT"},
		{"obsidian", "obsidian", "OBS"},
		{"applenotes", "applenotes", "NOTES"},
		{"unknown short", "jira", "JIRA"},
		{"todoist", "todoist", "TDT"},
		{"unknown long", "longname", "LONG"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SourceBadgeLabel(tt.provider)
			if got != tt.want {
				t.Errorf("SourceBadgeLabel(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

// --- SourceBadge (rendered) ---

func TestSourceBadge_ContainsLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		contains string
	}{
		{"textfile", "textfile", "TXT"},
		{"obsidian", "obsidian", "OBS"},
		{"applenotes", "applenotes", "NOTES"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SourceBadge(tt.provider)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("SourceBadge(%q) = %q, want to contain %q", tt.provider, got, tt.contains)
			}
		})
	}
}

func TestSourceBadge_EmptyProviderReturnsEmpty(t *testing.T) {
	t.Parallel()
	got := SourceBadge("")
	if got != "" {
		t.Errorf("SourceBadge(\"\") = %q, want empty", got)
	}
}

// --- Integration: DoorsView shows source badge ---

func TestDoorsView_ShowsSourceBadge(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("test task from obsidian")
	task.SourceProvider = "obsidian"
	pool.AddTask(task)

	// Add more tasks to fill doors
	for i := 0; i < 3; i++ {
		t2 := core.NewTask("filler task")
		t2.SourceProvider = "textfile"
		pool.AddTask(t2)
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	view := dv.View()

	// The rendered view should contain source badge text
	if !strings.Contains(view, "OBS") && !strings.Contains(view, "TXT") {
		t.Error("expected door view to contain source badge (OBS or TXT)")
	}
}

// --- Integration: DetailView shows source provider ---

func TestDetailView_ShowsSourceProvider(t *testing.T) {
	t.Parallel()
	task := core.NewTask("test task from obsidian")
	task.SourceProvider = "obsidian"
	pool := core.NewTaskPool()
	pool.AddTask(task)
	tracker := core.NewSessionTracker()

	dv := NewDetailView(task, tracker, nil, pool)
	view := dv.View()

	if !strings.Contains(view, "obsidian") && !strings.Contains(view, "OBS") {
		t.Error("expected detail view to contain source provider info")
	}
}

// --- Integration: SearchView shows source badge ---

func TestSearchView_ShowsSourceBadge(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("searchable task from obsidian")
	task.SourceProvider = "obsidian"
	pool.AddTask(task)

	tracker := core.NewSessionTracker()
	sv := NewSearchView(pool, tracker, nil, nil, nil)

	// Simulate search
	sv.results = []*core.Task{task}
	view := sv.View()

	if !strings.Contains(view, "OBS") {
		t.Error("expected search view to contain source badge OBS")
	}
}

// --- DevDispatchBadge ---

func TestDevDispatchBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		task     *core.Task
		contains string
		empty    bool
	}{
		{
			name:  "nil DevDispatch returns empty",
			task:  core.NewTask("no dispatch"),
			empty: true,
		},
		{
			name: "queued shows QUEUED",
			task: func() *core.Task {
				task := core.NewTask("queued task")
				task.DevDispatch = &dispatch.DevDispatch{Queued: true}
				return task
			}(),
			contains: "QUEUED",
		},
		{
			name: "open PR shows PR number",
			task: func() *core.Task {
				task := core.NewTask("pr task")
				task.DevDispatch = &dispatch.DevDispatch{PRNumber: 42, PRStatus: "open"}
				return task
			}(),
			contains: "PR #42",
		},
		{
			name: "merged PR shows MERGED",
			task: func() *core.Task {
				task := core.NewTask("merged task")
				task.DevDispatch = &dispatch.DevDispatch{PRNumber: 42, PRStatus: "merged"}
				return task
			}(),
			contains: "MERGED",
		},
		{
			name: "no PR no queued returns empty",
			task: func() *core.Task {
				task := core.NewTask("empty dispatch")
				task.DevDispatch = &dispatch.DevDispatch{}
				return task
			}(),
			empty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DevDispatchBadge(tt.task)
			if tt.empty {
				if got != "" {
					t.Errorf("DevDispatchBadge() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Errorf("DevDispatchBadge() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// --- DuplicateIndicator ---

func TestDuplicateIndicator(t *testing.T) {
	t.Parallel()
	indicator := DuplicateIndicator()
	if !strings.Contains(indicator, "Possible duplicate") {
		t.Errorf("DuplicateIndicator() = %q, want to contain 'Possible duplicate'", indicator)
	}
}

// --- Integration: DoorsView shows duplicate indicator ---

func TestDoorsView_ShowsDuplicateIndicator(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task1 := core.NewTask("buy groceries from the store")
	task1.SourceProvider = "textfile"
	pool.AddTask(task1)

	task2 := core.NewTask("buy groceries from the store")
	task2.SourceProvider = "obsidian"
	pool.AddTask(task2)

	task3 := core.NewTask("filler task")
	task3.SourceProvider = "textfile"
	pool.AddTask(task3)

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.SetDuplicateTaskIDs(map[string]bool{task1.ID: true, task2.ID: true})
	view := dv.View()

	if !strings.Contains(view, "Possible duplicate") {
		t.Error("expected door view to show duplicate indicator for flagged tasks")
	}
}

func TestDoorsView_NoDuplicateIndicatorWhenNotFlagged(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	for i := 0; i < 3; i++ {
		task := core.NewTask("unique task")
		task.SourceProvider = "textfile"
		pool.AddTask(task)
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	view := dv.View()

	if strings.Contains(view, "Possible duplicate") {
		t.Error("expected no duplicate indicator when no tasks are flagged")
	}
}

// --- Integration: DetailView shows duplicate indicator ---

func TestDetailView_ShowsDuplicateIndicator(t *testing.T) {
	t.Parallel()
	task := core.NewTask("buy groceries from the store")
	task.SourceProvider = "textfile"
	pool := core.NewTaskPool()
	pool.AddTask(task)
	tracker := core.NewSessionTracker()

	dv := NewDetailView(task, tracker, nil, pool)
	dv.SetDuplicateInfo(true, nil, nil)
	view := dv.View()

	if !strings.Contains(view, "Possible duplicate") {
		t.Error("expected detail view to show duplicate indicator")
	}
}

func TestDetailView_ShowsDupHintWhenFlagged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := core.NewDedupStore(dir + "/dedup.yaml")
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	taskA := core.NewTask("buy groceries")
	taskA.SourceProvider = "textfile"
	taskB := core.NewTask("buy groceries")
	taskB.SourceProvider = "obsidian"
	pair := &core.DuplicatePair{TaskA: taskA, TaskB: taskB, Similarity: 0.95}

	pool := core.NewTaskPool()
	pool.AddTask(taskA)
	pool.AddTask(taskB)
	tracker := core.NewSessionTracker()

	dv := NewDetailView(taskA, tracker, nil, pool)
	dv.SetDuplicateInfo(true, store, pair)
	view := dv.View()

	if !strings.Contains(view, "ismiss") {
		t.Errorf("expected detail view to show dismiss hint, got: %s", view)
	}
	if !strings.Contains(view, "merge") {
		t.Errorf("expected detail view to show merge hint, got: %s", view)
	}
}

// --- Integration: SearchView shows duplicate indicator ---

func TestSearchView_ShowsDuplicateIndicator(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("buy groceries from the store")
	task.SourceProvider = "textfile"
	pool.AddTask(task)

	tracker := core.NewSessionTracker()
	sv := NewSearchView(pool, tracker, nil, nil, nil)
	sv.SetDuplicateTaskIDs(map[string]bool{task.ID: true})
	sv.results = []*core.Task{task}
	view := sv.View()

	if !strings.Contains(view, "Possible duplicate") {
		t.Error("expected search view to show duplicate indicator for flagged task")
	}
}

// --- DetailView duplicate dismiss action ---

func TestDetailView_DismissDuplicate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := core.NewDedupStore(dir + "/dedup.yaml")
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	taskA := core.NewTask("buy groceries")
	taskA.SourceProvider = "textfile"
	taskB := core.NewTask("buy groceries")
	taskB.SourceProvider = "obsidian"
	pair := &core.DuplicatePair{TaskA: taskA, TaskB: taskB, Similarity: 0.95}

	pool := core.NewTaskPool()
	pool.AddTask(taskA)
	pool.AddTask(taskB)
	tracker := core.NewSessionTracker()

	dv := NewDetailView(taskA, tracker, nil, pool)
	dv.SetDuplicateInfo(true, store, pair)

	// Press 'd' to dismiss
	cmd := dv.Update(keyMsg("d"))
	if cmd == nil {
		t.Fatal("expected command from dismiss action")
		return
	}

	// Verify decision was persisted
	if !store.HasDecision(taskA.ID, taskB.ID) {
		t.Error("expected decision to be recorded after dismiss")
	}
	decision, _ := store.GetDecision(taskA.ID, taskB.ID)
	if decision != core.DecisionDistinct {
		t.Errorf("expected decision %q, got %q", core.DecisionDistinct, decision)
	}
}

// --- DetailView duplicate merge action ---

func TestDetailView_MergeDuplicate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := core.NewDedupStore(dir + "/dedup.yaml")
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	taskA := core.NewTask("buy groceries")
	taskA.SourceProvider = "textfile"
	taskB := core.NewTask("buy groceries")
	taskB.SourceProvider = "obsidian"
	pair := &core.DuplicatePair{TaskA: taskA, TaskB: taskB, Similarity: 0.95}

	pool := core.NewTaskPool()
	pool.AddTask(taskA)
	pool.AddTask(taskB)
	tracker := core.NewSessionTracker()

	dv := NewDetailView(taskA, tracker, nil, pool)
	dv.SetDuplicateInfo(true, store, pair)

	// Press 'y' to merge
	cmd := dv.Update(keyMsg("y"))
	if cmd == nil {
		t.Fatal("expected command from merge action")
		return
	}

	// Execute the command and check the message
	msg := cmd()
	mergeMsg, ok := msg.(DuplicateMergedMsg)
	if !ok {
		t.Fatalf("expected DuplicateMergedMsg, got %T", msg)
	}
	// The removed task should be the other one (taskB since we're viewing taskA)
	if mergeMsg.RemovedTask.ID != taskB.ID {
		t.Errorf("expected removed task %s, got %s", taskB.ID, mergeMsg.RemovedTask.ID)
	}

	// Verify decision was persisted
	if !store.HasDecision(taskA.ID, taskB.ID) {
		t.Error("expected decision to be recorded after merge")
	}
	decision, _ := store.GetDecision(taskA.ID, taskB.ID)
	if decision != core.DecisionDuplicate {
		t.Errorf("expected decision %q, got %q", core.DecisionDuplicate, decision)
	}
}
