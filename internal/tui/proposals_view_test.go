package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/mcp"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestProposalStore(t *testing.T, pool *core.TaskPool) *mcp.ProposalStore {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "proposals.jsonl")
	store, err := mcp.NewProposalStore(path, pool)
	if err != nil {
		t.Fatalf("create proposal store: %v", err)
		return nil
	}
	return store
}

func addTestTask(t *testing.T, pool *core.TaskPool, id, text string) *core.Task {
	t.Helper()
	task := &core.Task{
		ID:        id,
		Text:      text,
		Status:    core.StatusTodo,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	pool.AddTask(task)
	return task
}

func addTestProposal(t *testing.T, store *mcp.ProposalStore, taskID string, baseVersion time.Time, pType mcp.ProposalType, payload map[string]any, rationale string) *mcp.Proposal {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
		return nil
	}
	p, err := mcp.NewProposal(pType, taskID, baseVersion, data, "mcp:test", rationale)
	if err != nil {
		t.Fatalf("create proposal: %v", err)
		return nil
	}
	if err := store.Create(p); err != nil {
		t.Fatalf("store proposal: %v", err)
	}
	return p
}

type mockTaskProvider struct {
	tasks map[string]*core.Task
}

func newMockProvider() *mockTaskProvider {
	return &mockTaskProvider{tasks: make(map[string]*core.Task)}
}

func (m *mockTaskProvider) Name() string                     { return "mock" }
func (m *mockTaskProvider) LoadTasks() ([]*core.Task, error) { return nil, nil }
func (m *mockTaskProvider) SaveTask(task *core.Task) error {
	m.tasks[task.ID] = task
	return nil
}
func (m *mockTaskProvider) SaveTasks(_ []*core.Task) error { return nil }
func (m *mockTaskProvider) DeleteTask(_ string) error      { return nil }
func (m *mockTaskProvider) MarkComplete(_ string) error    { return nil }
func (m *mockTaskProvider) Watch() <-chan core.ChangeEvent { return nil }
func (m *mockTaskProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{}
}

func TestProposalsView_EmptyState(t *testing.T) {
	pool := core.NewTaskPool()
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	pv := NewProposalsView(store, pool, provider)
	view := pv.View()

	if !contains(view, "No suggestions") {
		t.Errorf("empty state should show 'No suggestions', got %q", view)
	}
	if !contains(view, "LLM clients can propose enrichments via MCP") {
		t.Errorf("empty state should show MCP guidance, got %q", view)
	}
}

func TestProposalsView_GroupsByTask(t *testing.T) {
	pool := core.NewTaskPool()
	task1 := addTestTask(t, pool, "t1", "Fix bug in login")
	task2 := addTestTask(t, pool, "t2", "Add dark mode")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task1.UpdatedAt, mcp.ProposalAddContext, map[string]any{"context": "auth module"}, "Adds context")
	addTestProposal(t, store, "t1", task1.UpdatedAt, mcp.ProposalUpdateEffort, map[string]any{"effort": "M"}, "Medium effort")
	addTestProposal(t, store, "t2", task2.UpdatedAt, mcp.ProposalAddNote, map[string]any{"note": "Consider CSS vars"}, "Styling approach")

	pv := NewProposalsView(store, pool, provider)

	if len(pv.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(pv.groups))
	}
	if len(pv.flatIndex) != 3 {
		t.Fatalf("expected 3 flat entries, got %d", len(pv.flatIndex))
	}

	// Verify groups have correct task associations
	for _, g := range pv.groups {
		if g.Task == nil {
			t.Errorf("group for task %s has nil Task", g.TaskID)
		}
	}
}

func TestProposalsView_Navigation(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddContext, map[string]any{"context": "ctx1"}, "First")
	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddNote, map[string]any{"note": "note1"}, "Second")

	pv := NewProposalsView(store, pool, provider)

	if pv.selectedIndex != 0 {
		t.Fatalf("initial selectedIndex should be 0, got %d", pv.selectedIndex)
	}

	// Navigate down
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if pv.selectedIndex != 1 {
		t.Errorf("after j, selectedIndex should be 1, got %d", pv.selectedIndex)
	}

	// Don't go past end
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if pv.selectedIndex != 1 {
		t.Errorf("should stay at 1, got %d", pv.selectedIndex)
	}

	// Navigate up
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if pv.selectedIndex != 0 {
		t.Errorf("after k, selectedIndex should be 0, got %d", pv.selectedIndex)
	}

	// Don't go below 0
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if pv.selectedIndex != 0 {
		t.Errorf("should stay at 0, got %d", pv.selectedIndex)
	}
}

func TestProposalsView_ApproveSelected(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddContext, map[string]any{"context": "added context"}, "Add context")

	pv := NewProposalsView(store, pool, provider)

	// Approve
	cmd := pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("approve should return a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ProposalApprovedMsg); !ok {
		t.Fatalf("expected ProposalApprovedMsg, got %T", msg)
	}

	// Task should have been updated
	updatedTask := pool.GetTask("t1")
	if updatedTask.Context != "added context" {
		t.Errorf("expected context 'added context', got %q", updatedTask.Context)
	}

	// Provider should have been called
	if _, ok := provider.tasks["t1"]; !ok {
		t.Error("provider.SaveTask should have been called for t1")
	}

	// No more pending proposals
	pending := store.List(mcp.ProposalFilter{Status: mcp.ProposalPending})
	if len(pending) != 0 {
		t.Errorf("expected 0 pending proposals after approve, got %d", len(pending))
	}
}

func TestProposalsView_RejectSelected(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddNote, map[string]any{"note": "unwanted"}, "A note")

	pv := NewProposalsView(store, pool, provider)

	cmd := pv.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatal("reject should return a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ProposalRejectedMsg); !ok {
		t.Fatalf("expected ProposalRejectedMsg, got %T", msg)
	}

	// Proposal should be rejected
	pending := store.List(mcp.ProposalFilter{Status: mcp.ProposalPending})
	if len(pending) != 0 {
		t.Errorf("expected 0 pending proposals after reject, got %d", len(pending))
	}

	rejected := store.List(mcp.ProposalFilter{Status: mcp.ProposalRejected})
	if len(rejected) != 1 {
		t.Errorf("expected 1 rejected proposal, got %d", len(rejected))
	}
}

func TestProposalsView_StaleDetection(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	oldVersion := task.UpdatedAt
	addTestProposal(t, store, "t1", oldVersion, mcp.ProposalAddContext, map[string]any{"context": "stale"}, "Stale proposal")

	// Modify the task after proposal was created
	task.UpdatedAt = time.Now().UTC().Add(time.Hour)

	pv := NewProposalsView(store, pool, provider)

	// Proposal should be detected as stale
	proposal := pv.currentProposal()
	if !isStale(proposal, task) {
		t.Error("proposal should be stale after task was modified")
	}

	// View should show stale indicator (may be word-wrapped)
	pv.SetWidth(80)
	view := pv.View()
	if !contains(view, "Task changed since") {
		t.Errorf("stale proposal should show warning, got %q", view)
	}

	// Trying to approve should fail with flash
	cmd := pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("approve of stale should return a command")
		return
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Fatalf("expected FlashMsg for stale approval, got %T", msg)
	} else if !contains(fm.Text, "stale") {
		t.Errorf("expected stale warning in flash, got %q", fm.Text)
	}
}

func TestProposalsView_PreviewMode(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalEnrichMetadata, map[string]any{"type": "technical", "effort": "L"}, "Enrich metadata")

	pv := NewProposalsView(store, pool, provider)
	pv.SetWidth(80)

	// Enter preview mode
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if !pv.previewMode {
		t.Error("should be in preview mode after pressing p")
	}

	view := pv.View()
	if !contains(view, "Preview") {
		t.Errorf("preview mode should show 'Preview', got %q", view)
	}
	if !contains(view, "BEFORE") && !contains(view, "AFTER") {
		t.Errorf("preview should show BEFORE/AFTER labels")
	}

	// Exit preview mode
	pv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if pv.previewMode {
		t.Error("should exit preview mode after Esc")
	}
}

func TestProposalsView_EscReturns(t *testing.T) {
	pool := core.NewTaskPool()
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	pv := NewProposalsView(store, pool, provider)

	cmd := pv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc should return a command")
		return
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestProposalsView_TabSkip(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddContext, map[string]any{"context": "c1"}, "First")
	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddNote, map[string]any{"note": "n1"}, "Second")

	pv := NewProposalsView(store, pool, provider)

	if pv.selectedIndex != 0 {
		t.Fatalf("initial index should be 0")
	}

	pv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if pv.selectedIndex != 1 {
		t.Errorf("Tab should move to next, got %d", pv.selectedIndex)
	}
}

func TestProposalsView_ApproveAll(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)
	provider := newMockProvider()

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddContext, map[string]any{"context": "ctx"}, "Context")
	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalUpdateEffort, map[string]any{"effort": "S"}, "Effort")

	pv := NewProposalsView(store, pool, provider)

	cmd := pv.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd == nil {
		t.Fatal("Ctrl+A should return a command")
		return
	}

	msg := cmd()
	batchMsg, ok := msg.(ProposalBatchApprovedMsg)
	if !ok {
		t.Fatalf("expected ProposalBatchApprovedMsg, got %T", msg)
	}
	if batchMsg.Count != 2 {
		t.Errorf("expected 2 approved, got %d", batchMsg.Count)
	}

	pending := store.List(mcp.ProposalFilter{Status: mcp.ProposalPending})
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after approve all, got %d", len(pending))
	}
}

func TestApplyProposal(t *testing.T) {
	tests := []struct {
		name     string
		pType    mcp.ProposalType
		payload  map[string]any
		validate func(t *testing.T, task *core.Task)
	}{
		{
			name:    "enrich metadata sets type and effort",
			pType:   mcp.ProposalEnrichMetadata,
			payload: map[string]any{"type": "technical", "effort": "L"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Type != core.TaskType("technical") {
					t.Errorf("expected type 'technical', got %q", task.Type)
				}
				if task.Effort != core.TaskEffort("L") {
					t.Errorf("expected effort 'L', got %q", task.Effort)
				}
			},
		},
		{
			name:    "add context sets context field",
			pType:   mcp.ProposalAddContext,
			payload: map[string]any{"context": "authentication module"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Context != "authentication module" {
					t.Errorf("expected context 'authentication module', got %q", task.Context)
				}
			},
		},
		{
			name:    "add note appends to notes",
			pType:   mcp.ProposalAddNote,
			payload: map[string]any{"note": "consider caching"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if len(task.Notes) != 1 {
					t.Fatalf("expected 1 note, got %d", len(task.Notes))
				}
				if task.Notes[0].Text != "consider caching" {
					t.Errorf("expected note text 'consider caching', got %q", task.Notes[0].Text)
				}
			},
		},
		{
			name:    "suggest blocker sets blocker",
			pType:   mcp.ProposalSuggestBlocker,
			payload: map[string]any{"blocker": "waiting on API access"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Blocker != "waiting on API access" {
					t.Errorf("expected blocker 'waiting on API access', got %q", task.Blocker)
				}
			},
		},
		{
			name:    "suggest category sets type",
			pType:   mcp.ProposalSuggestCategory,
			payload: map[string]any{"type": "creative"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Type != core.TaskType("creative") {
					t.Errorf("expected type 'creative', got %q", task.Type)
				}
			},
		},
		{
			name:    "update effort sets effort",
			pType:   mcp.ProposalUpdateEffort,
			payload: map[string]any{"effort": "XL"},
			validate: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Effort != core.TaskEffort("XL") {
					t.Errorf("expected effort 'XL', got %q", task.Effort)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &core.Task{
				ID:        "test-id",
				Text:      "Test task",
				Status:    core.StatusTodo,
				UpdatedAt: time.Now().UTC(),
			}

			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("marshal: %v", err)
				return
			}

			proposal := &mcp.Proposal{
				ID:      "p1",
				Type:    tt.pType,
				TaskID:  "test-id",
				Payload: data,
			}

			if err := applyProposal(proposal, task); err != nil {
				t.Fatalf("applyProposal: %v", err)
			}

			tt.validate(t, task)
		})
	}
}

func TestPendingProposalCount(t *testing.T) {
	pool := core.NewTaskPool()
	task := addTestTask(t, pool, "t1", "Test task")
	store := newTestProposalStore(t, pool)

	if count := PendingProposalCount(store); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	addTestProposal(t, store, "t1", task.UpdatedAt, mcp.ProposalAddNote, map[string]any{"note": "test"}, "A note")
	if count := PendingProposalCount(store); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := PendingProposalCount(nil); count != 0 {
		t.Errorf("nil store should return 0, got %d", count)
	}
}

func TestIsStale(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		baseVersion time.Time
		taskUpdated time.Time
		nilTask     bool
		wantStale   bool
	}{
		{"matching versions", now, now, false, false},
		{"task updated after proposal", now, now.Add(time.Hour), false, true},
		{"nil task", now, time.Time{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &mcp.Proposal{BaseVersion: tt.baseVersion}
			var task *core.Task
			if !tt.nilTask {
				task = &core.Task{UpdatedAt: tt.taskUpdated}
			}
			got := isStale(p, task)
			if got != tt.wantStale {
				t.Errorf("isStale() = %v, want %v", got, tt.wantStale)
			}
		})
	}
}

func TestDoorsView_ProposalBadge(t *testing.T) {
	pool := core.NewTaskPool()
	addTestTask(t, pool, "t1", "Task 1")
	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.SetWidth(80)
	dv.SetHeight(40)

	// No proposals — badge should not appear
	view := dv.View()
	if contains(view, "suggestions") {
		t.Error("should not show suggestions badge when count is 0")
	}

	// Set pending proposals
	dv.SetPendingProposals(3)
	view = dv.View()
	if !contains(view, "3 suggestions") {
		t.Errorf("should show '3 suggestions' badge, got %q", view)
	}
}

func TestSearchView_SuggestionsCommand(t *testing.T) {
	pool := core.NewTaskPool()
	tracker := core.NewSessionTracker()
	sv := NewSearchView(pool, tracker, nil, nil, nil)

	sv.textInput.SetValue(":suggestions")
	sv.checkCommandMode()

	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":suggestions should return a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ShowProposalsMsg); !ok {
		t.Fatalf("expected ShowProposalsMsg, got %T", msg)
	}
}

// Helper — also used by other test files in this package.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Ensure testdata directory exists for temp files.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
