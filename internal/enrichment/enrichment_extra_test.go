package enrichment

import (
	"path/filepath"
	"testing"
)

func TestDeleteTaskMetadata_Nonexistent(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	// Deleting non-existent should succeed (no-op)
	if err := edb.DeleteTaskMetadata("does-not-exist"); err != nil {
		t.Fatalf("DeleteTaskMetadata for nonexistent: %v", err)
	}
}

func TestDeleteCrossReference_Nonexistent(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	if err := edb.DeleteCrossReference(99999); err != nil {
		t.Fatalf("DeleteCrossReference for nonexistent: %v", err)
	}
}

func TestDeleteLearningPattern_Nonexistent(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	if err := edb.DeleteLearningPattern("fake-type", "fake-key"); err != nil {
		t.Fatalf("DeleteLearningPattern for nonexistent: %v", err)
	}
}

func TestOpen_CreatesDirectory(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "nested", "dirs", "enrichment.db")

	edb, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open with nested dirs: %v", err)
	}
	t.Cleanup(func() { _ = edb.Close() })

	if edb.Path() != dbPath {
		t.Errorf("Path() = %q, want %q", edb.Path(), dbPath)
	}
}

func TestGetCrossReferences_NoMatches(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	refs, err := edb.GetCrossReferences("nonexistent-task")
	if err != nil {
		t.Fatalf("GetCrossReferences: %v", err)
	}
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestUpsertTaskMetadata_MultipleTimes(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	for i := 0; i < 3; i++ {
		meta := &TaskMetadata{
			TaskID:   "upsert-test",
			Category: "test",
			Notes:    "round",
		}
		if err := edb.UpsertTaskMetadata(meta); err != nil {
			t.Fatalf("UpsertTaskMetadata round %d: %v", i, err)
		}
	}

	got, err := edb.GetTaskMetadata("upsert-test")
	if err != nil {
		t.Fatalf("GetTaskMetadata: %v", err)
	}
	if got.TaskID != "upsert-test" {
		t.Errorf("TaskID = %q, want %q", got.TaskID, "upsert-test")
	}
}
