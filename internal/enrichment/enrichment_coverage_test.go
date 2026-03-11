package enrichment

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestClose(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	if err := edb.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	// Second close should error (db already closed).
	err := edb.Close()
	if err == nil {
		t.Log("second Close() returned nil (driver-dependent behavior)")
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	edb, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = edb.Close() })

	if got := edb.Path(); got != dbPath {
		t.Errorf("Path() = %q, want %q", got, dbPath)
	}
}

func TestGetTaskMetadata_WrapsErrNoRows(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	_, err := edb.GetTaskMetadata("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected error chain to contain sql.ErrNoRows, got: %v", err)
	}
}

func TestAddCrossReference_SetsIDAndCreatedAt(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	ref := &CrossReference{
		SourceTaskID: "src-1",
		TargetTaskID: "tgt-1",
		SourceSystem: "test",
		Relationship: "depends-on",
	}

	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("AddCrossReference: %v", err)
	}

	if ref.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}
	if ref.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt after insert")
	}
}

func TestAddFeedback_SetsIDAndCreatedAt(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	f := &FeedbackEntry{
		TaskID:       "task-1",
		FeedbackType: "completed",
		Mood:         "happy",
		Comment:      "went well",
		SessionID:    "sess-1",
	}

	if err := edb.AddFeedback(f); err != nil {
		t.Fatalf("AddFeedback: %v", err)
	}

	if f.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}
	if f.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt after insert")
	}
}

func TestGetFeedbackByTask_ReturnsAll(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	for _, ft := range []string{"first", "second", "third"} {
		f := &FeedbackEntry{
			TaskID:       "order-test",
			FeedbackType: ft,
			SessionID:    "sess-1",
		}
		if err := edb.AddFeedback(f); err != nil {
			t.Fatalf("AddFeedback(%s): %v", ft, err)
		}
	}

	entries, err := edb.GetFeedbackByTask("order-test")
	if err != nil {
		t.Fatalf("GetFeedbackByTask: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	// Verify all feedback types are present.
	types := make(map[string]bool)
	for _, e := range entries {
		types[e.FeedbackType] = true
	}
	for _, want := range []string{"first", "second", "third"} {
		if !types[want] {
			t.Errorf("missing feedback type %q", want)
		}
	}
}

func TestGetFeedbackBySession_ReturnsAll(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	for _, ft := range []string{"alpha", "beta", "gamma"} {
		f := &FeedbackEntry{
			TaskID:       "task-" + ft,
			FeedbackType: ft,
			SessionID:    "order-sess",
		}
		if err := edb.AddFeedback(f); err != nil {
			t.Fatalf("AddFeedback(%s): %v", ft, err)
		}
	}

	entries, err := edb.GetFeedbackBySession("order-sess")
	if err != nil {
		t.Fatalf("GetFeedbackBySession: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	types := make(map[string]bool)
	for _, e := range entries {
		types[e.FeedbackType] = true
	}
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !types[want] {
			t.Errorf("missing feedback type %q", want)
		}
	}
}

func TestUpsertLearningPattern_UpdatesWeight(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	p := &LearningPattern{
		PatternType: "test-type",
		PatternKey:  "test-key",
		Weight:      0.5,
		Data:        `{"v":1}`,
	}
	if err := edb.UpsertLearningPattern(p); err != nil {
		t.Fatalf("initial upsert: %v", err)
	}

	p.Weight = 0.99
	p.Data = `{"v":2}`
	if err := edb.UpsertLearningPattern(p); err != nil {
		t.Fatalf("update upsert: %v", err)
	}

	patterns, err := edb.GetLearningPatternsByType("test-type")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType: %v", err)
	}
	if len(patterns) != 1 {
		t.Fatalf("got %d patterns, want 1 (upsert should not create duplicates)", len(patterns))
	}
	if patterns[0].Weight != 0.99 {
		t.Errorf("Weight = %f, want 0.99", patterns[0].Weight)
	}
	if patterns[0].Data != `{"v":2}` {
		t.Errorf("Data = %q, want %q", patterns[0].Data, `{"v":2}`)
	}
	if patterns[0].UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero after upsert")
	}
}

func TestTaskMetadata_EmptyTags(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	meta := &TaskMetadata{
		TaskID:   "no-tags",
		Category: "test",
	}
	if err := edb.UpsertTaskMetadata(meta); err != nil {
		t.Fatalf("UpsertTaskMetadata: %v", err)
	}

	got, err := edb.GetTaskMetadata("no-tags")
	if err != nil {
		t.Fatalf("GetTaskMetadata: %v", err)
	}

	if len(got.EnrichmentTags) != 0 {
		t.Errorf("EnrichmentTags = %v, want empty", got.EnrichmentTags)
	}
}

func TestTaskMetadata_MultipleTagsRoundTrip(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	tags := []string{"urgent", "backend", "review-needed"}
	meta := &TaskMetadata{
		TaskID:         "multi-tags",
		Category:       "dev",
		EnrichmentTags: tags,
		Notes:          "lots of tags",
	}
	if err := edb.UpsertTaskMetadata(meta); err != nil {
		t.Fatalf("UpsertTaskMetadata: %v", err)
	}

	got, err := edb.GetTaskMetadata("multi-tags")
	if err != nil {
		t.Fatalf("GetTaskMetadata: %v", err)
	}

	if len(got.EnrichmentTags) != len(tags) {
		t.Fatalf("got %d tags, want %d", len(got.EnrichmentTags), len(tags))
	}
	for i, tag := range tags {
		if got.EnrichmentTags[i] != tag {
			t.Errorf("tag[%d] = %q, want %q", i, got.EnrichmentTags[i], tag)
		}
	}
}

func TestTaskMetadata_TimestampsSet(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	meta := &TaskMetadata{
		TaskID:   "ts-test",
		Category: "test",
	}
	if err := edb.UpsertTaskMetadata(meta); err != nil {
		t.Fatalf("UpsertTaskMetadata: %v", err)
	}

	got, err := edb.GetTaskMetadata("ts-test")
	if err != nil {
		t.Fatalf("GetTaskMetadata: %v", err)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestCrossReference_MultiplePairs(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	// Create refs: A->B, A->C, D->A
	refs := []CrossReference{
		{SourceTaskID: "A", TargetTaskID: "B", Relationship: "blocks"},
		{SourceTaskID: "A", TargetTaskID: "C", Relationship: "related"},
		{SourceTaskID: "D", TargetTaskID: "A", Relationship: "depends-on"},
	}
	for i := range refs {
		if err := edb.AddCrossReference(&refs[i]); err != nil {
			t.Fatalf("AddCrossReference[%d]: %v", i, err)
		}
	}

	// Query for A — should return all 3 (A is source in 2, target in 1).
	got, err := edb.GetCrossReferences("A")
	if err != nil {
		t.Fatalf("GetCrossReferences(A): %v", err)
	}
	if len(got) != 3 {
		t.Errorf("got %d refs for A, want 3", len(got))
	}

	// Query for B — should return 1 (A->B).
	got2, err := edb.GetCrossReferences("B")
	if err != nil {
		t.Fatalf("GetCrossReferences(B): %v", err)
	}
	if len(got2) != 1 {
		t.Errorf("got %d refs for B, want 1", len(got2))
	}
}

func TestDeleteCrossReference_VerifiesRemoval(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	ref := &CrossReference{
		SourceTaskID: "del-src",
		TargetTaskID: "del-tgt",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("AddCrossReference: %v", err)
	}

	if err := edb.DeleteCrossReference(ref.ID); err != nil {
		t.Fatalf("DeleteCrossReference: %v", err)
	}

	got, err := edb.GetCrossReferences("del-src")
	if err != nil {
		t.Fatalf("GetCrossReferences after delete: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d refs after delete, want 0", len(got))
	}
}

func TestDeleteLearningPattern_VerifiesRemoval(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	p := &LearningPattern{
		PatternType: "del-type",
		PatternKey:  "del-key",
		Weight:      1.0,
	}
	if err := edb.UpsertLearningPattern(p); err != nil {
		t.Fatalf("UpsertLearningPattern: %v", err)
	}

	if err := edb.DeleteLearningPattern("del-type", "del-key"); err != nil {
		t.Fatalf("DeleteLearningPattern: %v", err)
	}

	patterns, err := edb.GetLearningPatternsByType("del-type")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType after delete: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("got %d patterns after delete, want 0", len(patterns))
	}
}

func TestDeleteTaskMetadata_VerifiesRemoval(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	meta := &TaskMetadata{
		TaskID:   "del-meta",
		Category: "test",
	}
	if err := edb.UpsertTaskMetadata(meta); err != nil {
		t.Fatalf("UpsertTaskMetadata: %v", err)
	}

	if err := edb.DeleteTaskMetadata("del-meta"); err != nil {
		t.Fatalf("DeleteTaskMetadata: %v", err)
	}

	_, err := edb.GetTaskMetadata("del-meta")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestCurrentVersion_ReturnsSchemaVersion(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	version, err := edb.currentVersion()
	if err != nil {
		t.Fatalf("currentVersion: %v", err)
	}
	if version != SchemaVersion {
		t.Errorf("currentVersion = %d, want %d", version, SchemaVersion)
	}
}

func TestApplyMigration_UnknownVersion(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	err := edb.applyMigration(999)
	if err == nil {
		t.Fatal("expected error for unknown migration version, got nil")
	}
}

func TestMultipleFeedbackForSameTask(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	for i := 0; i < 5; i++ {
		f := &FeedbackEntry{
			TaskID:       "multi-fb",
			FeedbackType: "not-now",
			Mood:         "neutral",
			SessionID:    "sess-multi",
		}
		if err := edb.AddFeedback(f); err != nil {
			t.Fatalf("AddFeedback[%d]: %v", i, err)
		}
	}

	entries, err := edb.GetFeedbackByTask("multi-fb")
	if err != nil {
		t.Fatalf("GetFeedbackByTask: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("got %d entries, want 5", len(entries))
	}
}

func TestGetLearningPatternsByType_MultipleKeys(t *testing.T) {
	t.Parallel()
	edb := openTestDB(t)

	for _, key := range []string{"key-a", "key-b", "key-c"} {
		p := &LearningPattern{
			PatternType: "multi-key",
			PatternKey:  key,
			Weight:      0.5,
		}
		if err := edb.UpsertLearningPattern(p); err != nil {
			t.Fatalf("UpsertLearningPattern(%s): %v", key, err)
		}
	}

	patterns, err := edb.GetLearningPatternsByType("multi-key")
	if err != nil {
		t.Fatalf("GetLearningPatternsByType: %v", err)
	}
	if len(patterns) != 3 {
		t.Errorf("got %d patterns, want 3", len(patterns))
	}
}

func TestOpen_ClosedDBOperationsFail(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "closed.db")
	edb, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = edb.Close()

	// Operations on closed DB should error.
	_, err = edb.GetTaskMetadata("any")
	if err == nil {
		t.Error("expected error on closed DB, got nil")
	}

	err = edb.UpsertTaskMetadata(&TaskMetadata{TaskID: "any"})
	if err == nil {
		t.Error("expected error on UpsertTaskMetadata with closed DB")
	}

	err = edb.AddCrossReference(&CrossReference{SourceTaskID: "a", TargetTaskID: "b"})
	if err == nil {
		t.Error("expected error on AddCrossReference with closed DB")
	}

	_, err = edb.GetCrossReferences("any")
	if err == nil {
		t.Error("expected error on GetCrossReferences with closed DB")
	}

	err = edb.AddFeedback(&FeedbackEntry{TaskID: "any", FeedbackType: "test"})
	if err == nil {
		t.Error("expected error on AddFeedback with closed DB")
	}

	_, err = edb.GetFeedbackByTask("any")
	if err == nil {
		t.Error("expected error on GetFeedbackByTask with closed DB")
	}

	_, err = edb.GetFeedbackBySession("any")
	if err == nil {
		t.Error("expected error on GetFeedbackBySession with closed DB")
	}

	err = edb.UpsertLearningPattern(&LearningPattern{PatternType: "t", PatternKey: "k"})
	if err == nil {
		t.Error("expected error on UpsertLearningPattern with closed DB")
	}

	_, err = edb.GetLearningPatternsByType("any")
	if err == nil {
		t.Error("expected error on GetLearningPatternsByType with closed DB")
	}

	err = edb.DeleteLearningPattern("t", "k")
	if err == nil {
		t.Error("expected error on DeleteLearningPattern with closed DB")
	}

	err = edb.DeleteCrossReference(1)
	if err == nil {
		t.Error("expected error on DeleteCrossReference with closed DB")
	}

	err = edb.DeleteTaskMetadata("any")
	if err == nil {
		t.Error("expected error on DeleteTaskMetadata with closed DB")
	}
}
