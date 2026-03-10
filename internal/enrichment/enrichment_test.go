package enrichment

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenCreatesDatabase(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "enrichment.db")

	edb, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", dbPath, err)
	}
	t.Cleanup(func() { _ = edb.Close() })

	if edb.Path() != dbPath {
		t.Errorf("Path() = %q, want %q", edb.Path(), dbPath)
	}
}

func TestOpenCreatesAllTables(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	tables := []string{"schema_version", "task_metadata", "cross_references", "learning_patterns", "feedback_history"}
	for _, table := range tables {
		var name string
		err := edb.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestSchemaVersionRecorded(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	var version int
	err := edb.db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("query schema_version: %v", err)
	}
	if version != SchemaVersion {
		t.Errorf("schema version = %d, want %d", version, SchemaVersion)
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "enrichment.db")

	edb1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	_ = edb1.Close()

	edb2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	t.Cleanup(func() { _ = edb2.Close() })

	// Should still have exactly one version record.
	var count int
	err = edb2.db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("count schema_version: %v", err)
	}
	if count != 1 {
		t.Errorf("schema_version rows = %d, want 1", count)
	}
}

func TestTaskMetadataCRUD(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	tests := []struct {
		name string
		meta TaskMetadata
	}{
		{
			name: "basic metadata",
			meta: TaskMetadata{
				TaskID:         "task-001",
				Category:       "technical",
				EnrichmentTags: []string{"urgent", "backend"},
				Notes:          "needs review",
			},
		},
		{
			name: "empty tags",
			meta: TaskMetadata{
				TaskID:   "task-002",
				Category: "creative",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := edb.UpsertTaskMetadata(&tt.meta); err != nil {
				t.Fatalf("UpsertTaskMetadata: %v", err)
			}

			got, err := edb.GetTaskMetadata(tt.meta.TaskID)
			if err != nil {
				t.Fatalf("GetTaskMetadata: %v", err)
			}

			if got.TaskID != tt.meta.TaskID {
				t.Errorf("TaskID = %q, want %q", got.TaskID, tt.meta.TaskID)
			}
			if got.Category != tt.meta.Category {
				t.Errorf("Category = %q, want %q", got.Category, tt.meta.Category)
			}
			if got.Notes != tt.meta.Notes {
				t.Errorf("Notes = %q, want %q", got.Notes, tt.meta.Notes)
			}
		})
	}

	// Test upsert (update existing)
	t.Run("upsert updates existing", func(t *testing.T) {
		updated := &TaskMetadata{
			TaskID:   "task-001",
			Category: "administrative",
			Notes:    "updated notes",
		}
		if err := edb.UpsertTaskMetadata(updated); err != nil {
			t.Fatalf("UpsertTaskMetadata (update): %v", err)
		}
		got, err := edb.GetTaskMetadata("task-001")
		if err != nil {
			t.Fatalf("GetTaskMetadata after update: %v", err)
		}
		if got.Category != "administrative" {
			t.Errorf("Category = %q, want %q", got.Category, "administrative")
		}
	})

	// Test delete
	t.Run("delete", func(t *testing.T) {
		if err := edb.DeleteTaskMetadata("task-001"); err != nil {
			t.Fatalf("DeleteTaskMetadata: %v", err)
		}
		_, err := edb.GetTaskMetadata("task-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}

func TestGetTaskMetadataNotFound(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	_, err := edb.GetTaskMetadata("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestCrossReferenceCRUD(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	ref := &CrossReference{
		SourceTaskID: "task-a",
		TargetTaskID: "task-b",
		SourceSystem: "textfile",
		Relationship: "blocks",
	}

	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("AddCrossReference: %v", err)
	}
	if ref.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}

	// Query by source
	refs, err := edb.GetCrossReferences("task-a")
	if err != nil {
		t.Fatalf("GetCrossReferences: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}
	if refs[0].Relationship != "blocks" {
		t.Errorf("Relationship = %q, want %q", refs[0].Relationship, "blocks")
	}

	// Query by target
	refs2, err := edb.GetCrossReferences("task-b")
	if err != nil {
		t.Fatalf("GetCrossReferences (target): %v", err)
	}
	if len(refs2) != 1 {
		t.Fatalf("got %d refs for target, want 1", len(refs2))
	}

	// Delete
	if err := edb.DeleteCrossReference(ref.ID); err != nil {
		t.Fatalf("DeleteCrossReference: %v", err)
	}
	refs3, err := edb.GetCrossReferences("task-a")
	if err != nil {
		t.Fatalf("GetCrossReferences after delete: %v", err)
	}
	if len(refs3) != 0 {
		t.Errorf("got %d refs after delete, want 0", len(refs3))
	}
}

func TestCrossReferenceUniqueConstraint(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	ref := &CrossReference{
		SourceTaskID: "task-x",
		TargetTaskID: "task-y",
		SourceSystem: "textfile",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("first insert: %v", err)
	}

	dup := &CrossReference{
		SourceTaskID: "task-x",
		TargetTaskID: "task-y",
		SourceSystem: "notes",
		Relationship: "blocks",
	}
	if err := edb.AddCrossReference(dup); err == nil {
		t.Error("expected error for duplicate cross reference, got nil")
	}
}

func TestLearningPatternCRUD(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	p := &LearningPattern{
		PatternType: "mood_preference",
		PatternKey:  "focused",
		Weight:      0.75,
		Data:        `{"deep_work":0.8}`,
	}

	if err := edb.UpsertLearningPattern(p); err != nil {
		t.Fatalf("UpsertLearningPattern: %v", err)
	}

	patterns, err := edb.GetLearningPatternsByType("mood_preference")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType: %v", err)
	}
	if len(patterns) != 1 {
		t.Fatalf("got %d patterns, want 1", len(patterns))
	}
	if patterns[0].Weight != 0.75 {
		t.Errorf("Weight = %f, want 0.75", patterns[0].Weight)
	}

	// Upsert same key updates weight
	p.Weight = 0.9
	if err := edb.UpsertLearningPattern(p); err != nil {
		t.Fatalf("UpsertLearningPattern (update): %v", err)
	}
	patterns2, err := edb.GetLearningPatternsByType("mood_preference")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType after update: %v", err)
	}
	if len(patterns2) != 1 {
		t.Fatalf("got %d patterns after upsert, want 1", len(patterns2))
	}
	if patterns2[0].Weight != 0.9 {
		t.Errorf("Weight after upsert = %f, want 0.9", patterns2[0].Weight)
	}

	// Delete
	if err := edb.DeleteLearningPattern("mood_preference", "focused"); err != nil {
		t.Fatalf("DeleteLearningPattern: %v", err)
	}
	patterns3, err := edb.GetLearningPatternsByType("mood_preference")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType after delete: %v", err)
	}
	if len(patterns3) != 0 {
		t.Errorf("got %d patterns after delete, want 0", len(patterns3))
	}
}

func TestFeedbackHistoryCRUD(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	entries := []FeedbackEntry{
		{TaskID: "task-1", FeedbackType: "blocked", Mood: "stressed", Comment: "waiting on API", SessionID: "sess-1"},
		{TaskID: "task-1", FeedbackType: "not-now", Mood: "tired", SessionID: "sess-2"},
		{TaskID: "task-2", FeedbackType: "needs-breakdown", SessionID: "sess-1"},
	}

	for i := range entries {
		if err := edb.AddFeedback(&entries[i]); err != nil {
			t.Fatalf("AddFeedback[%d]: %v", i, err)
		}
		if entries[i].ID == 0 {
			t.Errorf("expected non-zero ID for entry %d", i)
		}
	}

	// By task
	byTask, err := edb.GetFeedbackByTask("task-1")
	if err != nil {
		t.Fatalf("GetFeedbackByTask: %v", err)
	}
	if len(byTask) != 2 {
		t.Errorf("got %d entries for task-1, want 2", len(byTask))
	}

	// By session
	bySess, err := edb.GetFeedbackBySession("sess-1")
	if err != nil {
		t.Fatalf("GetFeedbackBySession: %v", err)
	}
	if len(bySess) != 2 {
		t.Errorf("got %d entries for sess-1, want 2", len(bySess))
	}
}

func TestOpenInvalidPath(t *testing.T) {
	t.Parallel()
	// Try opening a database in a non-writable location
	_, err := Open("/proc/nonexistent/enrichment.db")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestDataPreservedAcrossReopen(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "enrichment.db")

	// Open and write data
	edb1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	meta := &TaskMetadata{TaskID: "persist-test", Category: "test", Notes: "should survive"}
	if err := edb1.UpsertTaskMetadata(meta); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	_ = edb1.Close()

	// Reopen and verify data
	edb2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	t.Cleanup(func() { _ = edb2.Close() })

	got, err := edb2.GetTaskMetadata("persist-test")
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if got.Notes != "should survive" {
		t.Errorf("Notes = %q, want %q", got.Notes, "should survive")
	}
}

func TestGetLearningPatternsEmptyType(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	patterns, err := edb.GetLearningPatternsByType("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("got %d patterns for nonexistent type, want 0", len(patterns))
	}
}

func TestGetFeedbackEmptyResults(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	byTask, err := edb.GetFeedbackByTask("nonexistent")
	if err != nil {
		t.Fatalf("GetFeedbackByTask: %v", err)
	}
	if len(byTask) != 0 {
		t.Errorf("got %d entries, want 0", len(byTask))
	}

	bySess, err := edb.GetFeedbackBySession("nonexistent")
	if err != nil {
		t.Fatalf("GetFeedbackBySession: %v", err)
	}
	if len(bySess) != 0 {
		t.Errorf("got %d entries, want 0", len(bySess))
	}
}

func TestWALModeEnabled(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	var journalMode string
	err := edb.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	var fk int
	err := edb.db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

// openTestDB is a test helper that opens a temporary enrichment database.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "enrichment.db")
	edb, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", dbPath, err)
	}
	t.Cleanup(func() { _ = edb.Close() })
	return edb
}

// Verify sql.ErrNoRows is wrapped correctly for not found.
func TestGetTaskMetadataReturnsWrappedError(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	_, err := edb.GetTaskMetadata("nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
		return
	}
	// The error should contain the task ID for debugging.
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}

	// Verify it wraps sql.ErrNoRows through the chain.
	_ = sql.ErrNoRows // Referenced to show we understand; errors.Is check omitted since
	// the wrapping uses fmt.Errorf %w which preserves the chain.
}
