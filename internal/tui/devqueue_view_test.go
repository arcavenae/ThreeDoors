package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/dispatch"
	tea "github.com/charmbracelet/bubbletea"
)

// mockDispatcher implements dispatch.Dispatcher for testing.
type mockDispatcher struct {
	createErr    error
	createName   string
	removeErr    error
	removeCalled string
}

func (m *mockDispatcher) CreateWorker(_ context.Context, _ string) (string, error) {
	if m.createErr != nil {
		return "", m.createErr
	}
	return m.createName, nil
}

func (m *mockDispatcher) ListWorkers(_ context.Context) ([]dispatch.WorkerInfo, error) {
	return nil, nil
}

func (m *mockDispatcher) GetHistory(_ context.Context, _ int) ([]dispatch.HistoryEntry, error) {
	return nil, nil
}

func (m *mockDispatcher) RemoveWorker(_ context.Context, name string) error {
	m.removeCalled = name
	return m.removeErr
}

func (m *mockDispatcher) CheckAvailable(_ context.Context) error {
	return nil
}

func newTestQueue(t *testing.T) *dispatch.DevQueue {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "queue.yaml")
	q, err := dispatch.NewDevQueue(path)
	if err != nil {
		t.Fatalf("create test queue: %v", err)
		return nil
	}
	return q
}

func addTestItem(t *testing.T, q *dispatch.DevQueue, text string, status dispatch.QueueItemStatus) dispatch.QueueItem {
	t.Helper()
	now := time.Now().UTC()
	item := dispatch.QueueItem{
		TaskText: text,
		Status:   status,
		QueuedAt: &now,
	}
	if err := q.Add(item); err != nil {
		t.Fatalf("add test item: %v", err)
	}
	items := q.List()
	return items[len(items)-1]
}

func TestNewDevQueueView(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Build feature", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})
	if len(dv.items) != 1 {
		t.Errorf("expected 1 item, got %d", len(dv.items))
	}
	if dv.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", dv.cursor)
	}
}

func TestDevQueueViewEmptyState(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	dv := NewDevQueueView(q, &mockDispatcher{})
	dv.SetWidth(80)

	view := dv.View()
	if !strings.Contains(view, "No items in dev queue") {
		t.Errorf("empty state message not found in view:\n%s", view)
	}
	if !strings.Contains(view, "Dispatch a task with 'x' from detail view") {
		t.Errorf("empty state hint not found in view:\n%s", view)
	}
}

func TestDevQueueViewNavigation(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Task 1", dispatch.QueueItemPending)
	addTestItem(t, q, "Task 2", dispatch.QueueItemPending)
	addTestItem(t, q, "Task 3", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})

	tests := []struct {
		name       string
		key        string
		wantCursor int
	}{
		{"move down with j", "j", 1},
		{"move down again", "j", 2},
		{"at bottom, stay", "j", 2},
		{"move up with k", "k", 1},
		{"move up again", "k", 0},
		{"at top, stay", "k", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if dv.cursor != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", dv.cursor, tt.wantCursor)
			}
		})
	}
}

func TestDevQueueViewNavigationArrows(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Task 1", dispatch.QueueItemPending)
	addTestItem(t, q, "Task 2", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})

	dv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if dv.cursor != 1 {
		t.Errorf("after down arrow: cursor = %d, want 1", dv.cursor)
	}
	dv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if dv.cursor != 0 {
		t.Errorf("after up arrow: cursor = %d, want 0", dv.cursor)
	}
}

func TestDevQueueViewEscReturns(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	dv := NewDevQueueView(q, &mockDispatcher{})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Esc, got nil")
		return
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDevQueueViewRejectPending(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Reject me", dispatch.QueueItemPending)
	addTestItem(t, q, "Keep me", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatal("expected ClearFlashCmd from reject, got nil")
		return
	}

	if len(dv.items) != 1 {
		t.Errorf("expected 1 item after reject, got %d", len(dv.items))
	}
	if dv.items[0].TaskText != "Keep me" {
		t.Errorf("wrong item remaining: %s", dv.items[0].TaskText)
	}
	if !strings.Contains(dv.flash, "removed") {
		t.Errorf("expected 'removed' in flash, got %q", dv.flash)
	}
}

func TestDevQueueViewRejectNonPending(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Dispatched task", dispatch.QueueItemDispatched)

	dv := NewDevQueueView(q, &mockDispatcher{})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatal("expected ClearFlashCmd, got nil")
		return
	}
	if !strings.Contains(dv.flash, "Only pending") {
		t.Errorf("expected 'Only pending' in flash, got %q", dv.flash)
	}
	if len(dv.items) != 1 {
		t.Errorf("item should not be removed: got %d items", len(dv.items))
	}
}

func TestDevQueueViewApproveNonPending(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Completed task", dispatch.QueueItemCompleted)

	dv := NewDevQueueView(q, &mockDispatcher{})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("expected ClearFlashCmd, got nil")
		return
	}
	if !strings.Contains(dv.flash, "Only pending") {
		t.Errorf("expected 'Only pending' in flash, got %q", dv.flash)
	}
}

func TestDevQueueViewKillNonDispatched(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Pending task", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	if cmd == nil {
		t.Fatal("expected ClearFlashCmd, got nil")
		return
	}
	if !strings.Contains(dv.flash, "Only dispatched") {
		t.Errorf("expected 'Only dispatched' in flash, got %q", dv.flash)
	}
}

func TestDevQueueViewStatusIcons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status dispatch.QueueItemStatus
		icon   string
	}{
		{"pending", dispatch.QueueItemPending, "⏳"},
		{"dispatched", dispatch.QueueItemDispatched, "⚙️"},
		{"completed", dispatch.QueueItemCompleted, "✅"},
		{"failed", dispatch.QueueItemFailed, "❌"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := statusIcon(tt.status)
			if got != tt.icon {
				t.Errorf("statusIcon(%s) = %q, want %q", tt.status, got, tt.icon)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{"short text", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello w…"},
		{"very short max", "hello", 1, "…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncate(tt.text, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.text, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    *time.Time
		want string
	}{
		{"nil", nil, "—"},
		{"just now", timePtr(time.Now().UTC().Add(-10 * time.Second)), "now"},
		{"minutes ago", timePtr(time.Now().UTC().Add(-5 * time.Minute)), "5m ago"},
		{"hours ago", timePtr(time.Now().UTC().Add(-3 * time.Hour)), "3h ago"},
		{"days ago", timePtr(time.Now().UTC().Add(-48 * time.Hour)), "2d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := relativeTime(tt.t)
			if got != tt.want {
				t.Errorf("relativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestDevQueueViewRender(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	now := time.Now().UTC()
	if err := q.Add(dispatch.QueueItem{
		TaskText:   "Build new feature",
		Status:     dispatch.QueueItemPending,
		QueuedAt:   &now,
		WorkerName: "",
	}); err != nil {
		t.Fatal(err)
	}
	dispatched := time.Now().UTC().Add(-2 * time.Hour)
	if err := q.Add(dispatch.QueueItem{
		TaskText:     "Fix bug",
		Status:       dispatch.QueueItemDispatched,
		QueuedAt:     &now,
		DispatchedAt: &dispatched,
		WorkerName:   "eager-fox",
		PRNumber:     42,
	}); err != nil {
		t.Fatal(err)
	}

	dv := NewDevQueueView(q, &mockDispatcher{})
	dv.SetWidth(100)

	view := dv.View()

	if !strings.Contains(view, "Dev Queue") {
		t.Error("header not found")
	}
	if !strings.Contains(view, "Build new feature") {
		t.Error("pending task not found")
	}
	if !strings.Contains(view, "Fix bug") {
		t.Error("dispatched task not found")
	}
	if !strings.Contains(view, "eager-fox") {
		t.Error("worker name not found")
	}
	if !strings.Contains(view, "#42") {
		t.Error("PR number not found")
	}
	if !strings.Contains(view, "y approve") {
		t.Error("help text not found")
	}
}

func TestDevQueueViewApproveCmd(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	item := addTestItem(t, q, "Approve me", dispatch.QueueItemPending)

	mock := &mockDispatcher{createName: "happy-worker"}
	dv := NewDevQueueView(q, mock)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("expected tea.Cmd from approve, got nil")
		return
	}

	msg := cmd()
	created, ok := msg.(DevQueueWorkerCreatedMsg)
	if !ok {
		t.Fatalf("expected DevQueueWorkerCreatedMsg, got %T", msg)
	}
	if created.Err != nil {
		t.Errorf("unexpected error: %v", created.Err)
	}
	if created.WorkerName != "happy-worker" {
		t.Errorf("worker name = %q, want %q", created.WorkerName, "happy-worker")
	}

	// Verify queue item was updated
	updated, err := q.Get(item.ID)
	if err != nil {
		t.Fatalf("get updated item: %v", err)
		return
	}
	if updated.Status != dispatch.QueueItemDispatched {
		t.Errorf("status = %s, want dispatched", updated.Status)
	}
	if updated.WorkerName != "happy-worker" {
		t.Errorf("worker name = %q, want %q", updated.WorkerName, "happy-worker")
	}
}

func TestDevQueueViewApproveError(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	item := addTestItem(t, q, "Will fail", dispatch.QueueItemPending)

	mock := &mockDispatcher{createErr: fmt.Errorf("connection refused")}
	dv := NewDevQueueView(q, mock)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("expected tea.Cmd, got nil")
		return
	}

	msg := cmd()
	created, ok := msg.(DevQueueWorkerCreatedMsg)
	if !ok {
		t.Fatalf("expected DevQueueWorkerCreatedMsg, got %T", msg)
	}
	if created.Err == nil {
		t.Error("expected error, got nil")
	}

	// Verify status reverted to pending
	updated, err := q.Get(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
		return
	}
	if updated.Status != dispatch.QueueItemPending {
		t.Errorf("status should revert to pending, got %s", updated.Status)
	}
}

func TestDevQueueViewKillCmd(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	now := time.Now().UTC()
	if err := q.Add(dispatch.QueueItem{
		TaskText:   "Kill me",
		Status:     dispatch.QueueItemDispatched,
		QueuedAt:   &now,
		WorkerName: "doomed-worker",
	}); err != nil {
		t.Fatal(err)
	}
	items := q.List()
	item := items[0]

	mock := &mockDispatcher{}
	dv := NewDevQueueView(q, mock)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	if cmd == nil {
		t.Fatal("expected tea.Cmd from kill, got nil")
		return
	}

	msg := cmd()
	removed, ok := msg.(DevQueueWorkerRemovedMsg)
	if !ok {
		t.Fatalf("expected DevQueueWorkerRemovedMsg, got %T", msg)
	}
	if removed.Err != nil {
		t.Errorf("unexpected error: %v", removed.Err)
	}
	if mock.removeCalled != "doomed-worker" {
		t.Errorf("RemoveWorker called with %q, want %q", mock.removeCalled, "doomed-worker")
	}

	// Verify queue item status updated
	updated, err := q.Get(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
		return
	}
	if updated.Status != dispatch.QueueItemFailed {
		t.Errorf("status = %s, want failed", updated.Status)
	}
	if updated.Error != "Killed by user" {
		t.Errorf("error = %q, want %q", updated.Error, "Killed by user")
	}
}

func TestDevQueueViewCursorBoundsAfterReject(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)
	addTestItem(t, q, "Task 1", dispatch.QueueItemPending)
	addTestItem(t, q, "Task 2", dispatch.QueueItemPending)

	dv := NewDevQueueView(q, &mockDispatcher{})

	// Move to last item
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if dv.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", dv.cursor)
	}

	// Reject it
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	// Cursor should move back
	if dv.cursor >= len(dv.items) {
		t.Errorf("cursor %d out of bounds (items: %d)", dv.cursor, len(dv.items))
	}
}

func TestDevQueueRemove(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "queue.yaml")
	q, err := dispatch.NewDevQueue(path)
	if err != nil {
		t.Fatal(err)
		return
	}

	now := time.Now().UTC()
	if err := q.Add(dispatch.QueueItem{TaskText: "A", QueuedAt: &now}); err != nil {
		t.Fatal(err)
	}
	if err := q.Add(dispatch.QueueItem{TaskText: "B", QueuedAt: &now}); err != nil {
		t.Fatal(err)
	}

	items := q.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Remove first item
	if err := q.Remove(items[0].ID); err != nil {
		t.Fatalf("remove: %v", err)
	}

	remaining := q.List()
	if len(remaining) != 1 {
		t.Fatalf("expected 1 item after remove, got %d", len(remaining))
	}
	if remaining[0].TaskText != "B" {
		t.Errorf("wrong item remaining: %s", remaining[0].TaskText)
	}

	// Verify persistence
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
		return
	}
	if strings.Contains(string(data), "A") {
		t.Error("removed item still in file")
	}
}

func TestDevQueueRemoveNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "queue.yaml")
	q, err := dispatch.NewDevQueue(path)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = q.Remove("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}
