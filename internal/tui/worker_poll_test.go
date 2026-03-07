package tui

import (
	"errors"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/dispatch"
)

func TestMapHistoryStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
		want   dispatch.QueueItemStatus
	}{
		{"completed maps to completed", "completed", dispatch.QueueItemCompleted},
		{"open maps to completed", "open", dispatch.QueueItemCompleted},
		{"merged maps to completed", "merged", dispatch.QueueItemCompleted},
		{"failed maps to failed", "failed", dispatch.QueueItemFailed},
		{"no-pr maps to failed", "no-pr", dispatch.QueueItemFailed},
		{"running stays dispatched", "running", dispatch.QueueItemDispatched},
		{"unknown stays dispatched", "unknown", dispatch.QueueItemDispatched},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapHistoryStatus(tt.status)
			if got != tt.want {
				t.Errorf("mapHistoryStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestMapPRStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
		want   string
	}{
		{"open", "open", "open"},
		{"merged", "merged", "merged"},
		{"completed maps to open", "completed", "open"},
		{"running passes through", "running", "running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapPRStatus(tt.status)
			if got != tt.want {
				t.Errorf("mapPRStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestHasDispatchedItems_NoQueue(t *testing.T) {
	t.Parallel()
	m := &MainModel{}
	if m.hasDispatchedItems() {
		t.Error("expected false with nil devQueue")
	}
}

func TestHasDispatchedItems_WithQueue(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)

	_ = q.Add(dispatch.QueueItem{TaskID: "t1", TaskText: "task 1", Status: dispatch.QueueItemPending})
	m := &MainModel{devQueue: q}
	if m.hasDispatchedItems() {
		t.Error("expected false with only pending items")
	}

	_ = q.Add(dispatch.QueueItem{TaskID: "t2", TaskText: "task 2", Status: dispatch.QueueItemDispatched, WorkerName: "fox"})
	if !m.hasDispatchedItems() {
		t.Error("expected true with dispatched item")
	}
}

func TestStartPollingIfNeeded(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	_ = q.Add(dispatch.QueueItem{TaskID: "t1", TaskText: "task", Status: dispatch.QueueItemDispatched, WorkerName: "fox"})

	m := &MainModel{devQueue: q}

	// Should start polling
	cmd := m.startPollingIfNeeded()
	if cmd == nil {
		t.Fatal("expected non-nil cmd when dispatched items exist")
	}
	if !m.pollingActive {
		t.Error("expected pollingActive to be true")
	}

	// Should not double-start
	cmd = m.startPollingIfNeeded()
	if cmd != nil {
		t.Error("expected nil cmd when already polling")
	}
}

func TestStartPollingIfNeeded_NoDispatchedItems(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	_ = q.Add(dispatch.QueueItem{TaskID: "t1", TaskText: "task", Status: dispatch.QueueItemPending})

	m := &MainModel{devQueue: q}
	cmd := m.startPollingIfNeeded()
	if cmd != nil {
		t.Error("expected nil cmd when no dispatched items")
	}
	if m.pollingActive {
		t.Error("expected pollingActive to be false")
	}
}

func TestHandleWorkerStatus_MatchesAndUpdatesQueue(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	pool := core.NewTaskPool()
	task := core.NewTask("implement feature")
	pool.AddTask(task)

	_ = q.Add(dispatch.QueueItem{
		ID:         "dq-test01",
		TaskID:     task.ID,
		TaskText:   "implement feature",
		Status:     dispatch.QueueItemDispatched,
		WorkerName: "happy-fox",
	})

	provider := &noopProvider{}
	m := &MainModel{
		devQueue:      q,
		pool:          pool,
		provider:      provider,
		pollingActive: true,
	}

	msg := WorkerStatusMsg{
		History: []dispatch.HistoryEntry{
			{
				WorkerName: "happy-fox",
				Status:     "open",
				PRNumber:   42,
				PRURL:      "https://github.com/example/repo/pull/42",
			},
		},
	}

	cmd := m.handleWorkerStatus(msg)

	// Queue item should be updated
	item, err := q.Get("dq-test01")
	if err != nil {
		t.Fatalf("Get queue item: %v", err)
	}
	if item.Status != dispatch.QueueItemCompleted {
		t.Errorf("queue item status = %q, want %q", item.Status, dispatch.QueueItemCompleted)
	}
	if item.PRNumber != 42 {
		t.Errorf("queue item PRNumber = %d, want 42", item.PRNumber)
	}
	if item.PRURL != "https://github.com/example/repo/pull/42" {
		t.Errorf("queue item PRURL = %q, want URL", item.PRURL)
	}

	// Task DevDispatch should be updated
	updated := pool.GetTask(task.ID)
	if updated.DevDispatch == nil {
		t.Fatal("task DevDispatch is nil")
	}
	if updated.DevDispatch.PRNumber != 42 {
		t.Errorf("task DevDispatch.PRNumber = %d, want 42", updated.DevDispatch.PRNumber)
	}
	if updated.DevDispatch.PRStatus != "open" {
		t.Errorf("task DevDispatch.PRStatus = %q, want %q", updated.DevDispatch.PRStatus, "open")
	}

	// No more dispatched items, polling should stop
	if m.pollingActive {
		t.Error("pollingActive should be false after all items completed")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no more dispatched items")
	}
}

func TestHandleWorkerStatus_Error_ContinuesPolling(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	_ = q.Add(dispatch.QueueItem{
		TaskID:     "t1",
		TaskText:   "task",
		Status:     dispatch.QueueItemDispatched,
		WorkerName: "fox",
	})

	m := &MainModel{devQueue: q, pollingActive: true}
	msg := WorkerStatusMsg{Err: errors.New("connection refused")}

	cmd := m.handleWorkerStatus(msg)
	if cmd == nil {
		t.Error("expected tick cmd to continue polling after error")
	}
	if !m.pollingActive {
		t.Error("pollingActive should remain true when dispatched items exist")
	}
}

func TestHandleWorkerStatus_Error_StopsWhenNoneDispatched(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	_ = q.Add(dispatch.QueueItem{
		TaskID:   "t1",
		TaskText: "task",
		Status:   dispatch.QueueItemCompleted,
	})

	m := &MainModel{devQueue: q, pollingActive: true}
	msg := WorkerStatusMsg{Err: errors.New("timeout")}

	cmd := m.handleWorkerStatus(msg)
	if cmd != nil {
		t.Error("expected nil cmd when no dispatched items after error")
	}
	if m.pollingActive {
		t.Error("pollingActive should be false when no dispatched items")
	}
}

func TestHandleWorkerStatus_NoMatch(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	_ = q.Add(dispatch.QueueItem{
		ID:         "dq-nomatch",
		TaskID:     "t1",
		TaskText:   "task",
		Status:     dispatch.QueueItemDispatched,
		WorkerName: "happy-fox",
	})

	m := &MainModel{devQueue: q, pollingActive: true}
	msg := WorkerStatusMsg{
		History: []dispatch.HistoryEntry{
			{WorkerName: "brave-lion", Status: "completed", PRNumber: 99},
		},
	}

	cmd := m.handleWorkerStatus(msg)

	// Item should remain dispatched
	item, _ := q.Get("dq-nomatch")
	if item.Status != dispatch.QueueItemDispatched {
		t.Errorf("queue item status = %q, want dispatched (no match)", item.Status)
	}

	// Should continue polling
	if cmd == nil {
		t.Error("expected tick cmd since dispatched item remains")
	}
}

func TestHandleWorkerStatus_MergedStatus(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	pool := core.NewTaskPool()
	task := core.NewTask("fix bug")
	pool.AddTask(task)

	_ = q.Add(dispatch.QueueItem{
		ID:         "dq-merged",
		TaskID:     task.ID,
		TaskText:   "fix bug",
		Status:     dispatch.QueueItemDispatched,
		WorkerName: "brave-lion",
	})

	provider := &noopProvider{}
	m := &MainModel{
		devQueue:      q,
		pool:          pool,
		provider:      provider,
		pollingActive: true,
	}

	msg := WorkerStatusMsg{
		History: []dispatch.HistoryEntry{
			{WorkerName: "brave-lion", Status: "merged", PRNumber: 55},
		},
	}

	m.handleWorkerStatus(msg)

	updated := pool.GetTask(task.ID)
	if updated.DevDispatch == nil {
		t.Fatal("DevDispatch is nil")
	}
	if updated.DevDispatch.PRStatus != "merged" {
		t.Errorf("PRStatus = %q, want %q", updated.DevDispatch.PRStatus, "merged")
	}
}

func TestHandleWorkerStatus_FailedStatus(t *testing.T) {
	t.Parallel()
	q := setupTestQueue(t)
	pool := core.NewTaskPool()
	task := core.NewTask("task")
	pool.AddTask(task)

	_ = q.Add(dispatch.QueueItem{
		ID:         "dq-failed",
		TaskID:     task.ID,
		TaskText:   "task",
		Status:     dispatch.QueueItemDispatched,
		WorkerName: "sad-panda",
	})

	provider := &noopProvider{}
	m := &MainModel{devQueue: q, pool: pool, provider: provider, pollingActive: true}

	msg := WorkerStatusMsg{
		History: []dispatch.HistoryEntry{
			{WorkerName: "sad-panda", Status: "no-pr"},
		},
	}

	m.handleWorkerStatus(msg)

	item, _ := q.Get("dq-failed")
	if item.Status != dispatch.QueueItemFailed {
		t.Errorf("queue item status = %q, want %q", item.Status, dispatch.QueueItemFailed)
	}
	if item.CompletedAt == nil {
		t.Error("CompletedAt should be set for failed items")
	}
}

// noopProvider implements core.TaskProvider for testing.
type noopProvider struct{}

func (p *noopProvider) Name() string                        { return "noop" }
func (p *noopProvider) LoadTasks() ([]*core.Task, error)    { return nil, nil }
func (p *noopProvider) SaveTask(_ *core.Task) error         { return nil }
func (p *noopProvider) SaveTasks(_ []*core.Task) error      { return nil }
func (p *noopProvider) DeleteTask(_ string) error           { return nil }
func (p *noopProvider) MarkComplete(_ string) error         { return nil }
func (p *noopProvider) Watch() <-chan core.ChangeEvent      { return nil }
func (p *noopProvider) HealthCheck() core.HealthCheckResult { return core.HealthCheckResult{} }

func setupTestQueue(t *testing.T) *dispatch.DevQueue {
	t.Helper()
	path := t.TempDir() + "/test-queue.yaml"
	q, err := dispatch.NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue: %v", err)
	}
	t.Cleanup(func() {
		// queue auto-saves, nothing extra needed
	})
	return q
}
