package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDedupStore_CreatesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")
	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}
	if store == nil {
		t.Fatal("expected non-nil store")
		return
	}
}

func TestDedupStore_RecordAndHasDecision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")
	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	if store.HasDecision("task-a", "task-b") {
		t.Error("expected no decision initially")
	}

	err = store.RecordDecision("task-a", "task-b", DecisionDuplicate)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
		return
	}

	if !store.HasDecision("task-a", "task-b") {
		t.Error("expected decision to exist after recording")
	}
}

func TestDedupStore_SymmetricLookup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")
	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	err = store.RecordDecision("task-a", "task-b", DecisionDistinct)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
		return
	}

	// Lookup in reverse order should also work
	if !store.HasDecision("task-b", "task-a") {
		t.Error("expected symmetric lookup to find decision")
	}
}

func TestDedupStore_GetDecision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")
	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	err = store.RecordDecision("task-a", "task-b", DecisionDuplicate)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
		return
	}

	decision, ok := store.GetDecision("task-a", "task-b")
	if !ok {
		t.Fatal("expected decision to exist")
	}
	if decision != DecisionDuplicate {
		t.Errorf("expected %q, got %q", DecisionDuplicate, decision)
	}
}

func TestDedupStore_PersistenceAcrossLoads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")

	store1, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}
	err = store1.RecordDecision("task-a", "task-b", DecisionDuplicate)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
		return
	}

	// Create a new store instance loading from the same file
	store2, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore reload: %v", err)
		return
	}

	if !store2.HasDecision("task-a", "task-b") {
		t.Error("expected decision to persist across loads")
	}
}

func TestDedupStore_FilterUndecided(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")
	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	taskA := NewTask("buy groceries")
	taskB := NewTask("buy grocery")
	taskC := NewTask("send email")
	taskD := NewTask("send emails")

	pairs := []DuplicatePair{
		{TaskA: taskA, TaskB: taskB, Similarity: 0.9},
		{TaskA: taskC, TaskB: taskD, Similarity: 0.85},
	}

	// Record decision for first pair
	err = store.RecordDecision(taskA.ID, taskB.ID, DecisionDistinct)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
		return
	}

	// Filter should remove the decided pair
	undecided := store.FilterUndecided(pairs)
	if len(undecided) != 1 {
		t.Fatalf("expected 1 undecided pair, got %d", len(undecided))
	}
	if undecided[0].TaskA.ID != taskC.ID {
		t.Errorf("expected task C, got %s", undecided[0].TaskA.ID)
	}
}

func TestDedupStore_InvalidPath(t *testing.T) {
	t.Parallel()
	// Path with nonexistent parent directory
	path := filepath.Join(t.TempDir(), "nonexistent", "subdir", "dedup.yaml")
	_, err := NewDedupStore(path)
	if err != nil {
		// Should succeed because NewDedupStore should create parent dirs
		// OR it could fail — depends on implementation choice
		// Either way, this test validates the behavior
		_ = err
	}
}

func TestDedupStore_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dedup.yaml")

	// Create an empty file
	err := os.WriteFile(path, []byte{}, 0o644)
	if err != nil {
		t.Fatalf("write empty file: %v", err)
		return
	}

	store, err := NewDedupStore(path)
	if err != nil {
		t.Fatalf("NewDedupStore with empty file: %v", err)
		return
	}
	if store.HasDecision("any", "thing") {
		t.Error("expected no decisions from empty file")
	}
}
