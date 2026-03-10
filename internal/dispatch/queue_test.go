package dispatch

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestQueue(t *testing.T) *DevQueue {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-queue.yaml")
	q, err := NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue: %v", err)
		return nil
	}
	return q
}

func TestDevQueueAdd(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{
		TaskID:   "task-1",
		TaskText: "Fix the bug",
	}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	items := q.List()
	if len(items) != 1 {
		t.Fatalf("List() len = %d, want 1", len(items))
	}
	if !strings.HasPrefix(items[0].ID, "dq-") {
		t.Errorf("ID = %q, want dq- prefix", items[0].ID)
	}
	if items[0].Status != QueueItemPending {
		t.Errorf("Status = %q, want %q", items[0].Status, QueueItemPending)
	}
	if items[0].QueuedAt == nil {
		t.Error("QueuedAt should be set automatically")
	}
	if items[0].TaskText != "Fix the bug" {
		t.Errorf("TaskText = %q, want %q", items[0].TaskText, "Fix the bug")
	}
}

func TestDevQueueAddWithExplicitID(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{
		ID:       "dq-custom01",
		TaskID:   "task-1",
		TaskText: "Custom ID item",
	}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := q.Get("dq-custom01")
	if err != nil {
		t.Fatalf("Get: %v", err)
		return
	}
	if got.ID != "dq-custom01" {
		t.Errorf("ID = %q, want %q", got.ID, "dq-custom01")
	}
}

func TestDevQueueAddDuplicate(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{ID: "dq-dup00001", TaskID: "task-1", TaskText: "Dup test"}
	if err := q.Add(item); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if err := q.Add(item); err == nil {
		t.Error("second Add should fail for duplicate ID")
	}
}

func TestDevQueueGet(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{ID: "dq-get00001", TaskID: "task-1", TaskText: "Get test"}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := q.Get("dq-get00001")
	if err != nil {
		t.Fatalf("Get: %v", err)
		return
	}
	if got.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", got.TaskID, "task-1")
	}
}

func TestDevQueueGetNotFound(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	_, err := q.Get("dq-nonexist")
	if err == nil {
		t.Fatal("Get should return error for non-existent ID")
		return
	}
	if !errors.Is(err, ErrQueueItemNotFound) {
		t.Errorf("err = %v, want ErrQueueItemNotFound", err)
	}
}

func TestDevQueueUpdate(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{ID: "dq-upd00001", TaskID: "task-1", TaskText: "Update test"}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	now := time.Now().UTC()
	err := q.Update("dq-upd00001", func(qi *QueueItem) {
		qi.Status = QueueItemDispatched
		qi.DispatchedAt = &now
		qi.WorkerName = "worker-1"
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
		return
	}

	got, err := q.Get("dq-upd00001")
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
		return
	}
	if got.Status != QueueItemDispatched {
		t.Errorf("Status = %q, want %q", got.Status, QueueItemDispatched)
	}
	if got.WorkerName != "worker-1" {
		t.Errorf("WorkerName = %q, want %q", got.WorkerName, "worker-1")
	}
}

func TestDevQueueUpdateNotFound(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	err := q.Update("dq-nonexist", func(qi *QueueItem) {
		qi.Status = QueueItemCompleted
	})
	if err == nil {
		t.Fatal("Update should return error for non-existent ID")
		return
	}
	if !errors.Is(err, ErrQueueItemNotFound) {
		t.Errorf("err = %v, want ErrQueueItemNotFound", err)
	}
}

func TestDevQueueList(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	for i := range 3 {
		item := QueueItem{
			TaskID:   "task-" + string(rune('1'+i)),
			TaskText: "List test item",
		}
		if err := q.Add(item); err != nil {
			t.Fatalf("Add item %d: %v", i, err)
		}
	}

	items := q.List()
	if len(items) != 3 {
		t.Errorf("List() len = %d, want 3", len(items))
	}
}

func TestDevQueueListReturnsCopy(t *testing.T) {
	t.Parallel()
	q := newTestQueue(t)

	item := QueueItem{TaskID: "task-1", TaskText: "Copy test"}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	items := q.List()
	items[0].TaskText = "mutated"

	original, err := q.Get(items[0].ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
		return
	}
	if original.TaskText == "mutated" {
		t.Error("List() should return a copy, not a reference to internal state")
	}
}

func TestDevQueuePersistence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "persist-queue.yaml")

	q1, err := NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue: %v", err)
		return
	}

	item := QueueItem{ID: "dq-persist1", TaskID: "task-1", TaskText: "Persist test"}
	if err := q1.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	q2, err := NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue reload: %v", err)
		return
	}

	got, err := q2.Get("dq-persist1")
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
		return
	}
	if got.TaskText != "Persist test" {
		t.Errorf("TaskText = %q, want %q", got.TaskText, "Persist test")
	}
}

func TestDevQueueAtomicWriteCleanup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "atomic-queue.yaml")

	q, err := NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue: %v", err)
		return
	}

	item := QueueItem{TaskID: "task-1", TaskText: "Atomic test"}
	if err := q.Add(item); err != nil {
		t.Fatalf("Add: %v", err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file %s should not exist after successful write", tmpPath)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("final file %s should exist: %v", path, err)
	}
}

func TestDevQueueEmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty-queue.yaml")

	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("create empty file: %v", err)
	}

	q, err := NewDevQueue(path)
	if err != nil {
		t.Fatalf("NewDevQueue with empty file: %v", err)
		return
	}

	items := q.List()
	if len(items) != 0 {
		t.Errorf("List() len = %d, want 0", len(items))
	}
}

func TestGenerateQueueID(t *testing.T) {
	t.Parallel()

	id := generateQueueID()
	if !strings.HasPrefix(id, "dq-") {
		t.Errorf("ID = %q, want dq- prefix", id)
	}
	if len(id) != 11 { // "dq-" (3) + 8 chars
		t.Errorf("ID len = %d, want 11", len(id))
	}

	id2 := generateQueueID()
	if id == id2 {
		t.Error("generateQueueID should produce unique IDs")
	}
}
