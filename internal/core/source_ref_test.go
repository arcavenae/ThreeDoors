package core

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSourceRefHasSourceRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		refs     []SourceRef
		provider string
		nativeID string
		want     bool
	}{
		{
			name:     "empty refs",
			refs:     nil,
			provider: "jira",
			nativeID: "PROJ-42",
			want:     false,
		},
		{
			name:     "matching ref",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "jira",
			nativeID: "PROJ-42",
			want:     true,
		},
		{
			name:     "different provider",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "reminders",
			nativeID: "PROJ-42",
			want:     false,
		},
		{
			name:     "different native ID",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "jira",
			nativeID: "PROJ-99",
			want:     false,
		},
		{
			name: "multiple refs with match",
			refs: []SourceRef{
				{Provider: "textfile", NativeID: "abc"},
				{Provider: "jira", NativeID: "PROJ-42"},
			},
			provider: "jira",
			nativeID: "PROJ-42",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{SourceRefs: tt.refs}
			got := task.HasSourceRef(tt.provider, tt.nativeID)
			if got != tt.want {
				t.Errorf("HasSourceRef(%q, %q) = %v, want %v", tt.provider, tt.nativeID, got, tt.want)
			}
		})
	}
}

func TestSourceRefAddSourceRef(t *testing.T) {
	t.Parallel()

	t.Run("adds new ref", func(t *testing.T) {
		t.Parallel()
		task := &Task{}
		task.AddSourceRef("jira", "PROJ-42")

		if len(task.SourceRefs) != 1 {
			t.Fatalf("expected 1 ref, got %d", len(task.SourceRefs))
		}
		if task.SourceRefs[0].Provider != "jira" || task.SourceRefs[0].NativeID != "PROJ-42" {
			t.Errorf("unexpected ref: %+v", task.SourceRefs[0])
		}
	})

	t.Run("does not add duplicate", func(t *testing.T) {
		t.Parallel()
		task := &Task{SourceRefs: []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}}}
		task.AddSourceRef("jira", "PROJ-42")

		if len(task.SourceRefs) != 1 {
			t.Errorf("expected 1 ref (no duplicate), got %d", len(task.SourceRefs))
		}
	})

	t.Run("adds ref for different provider", func(t *testing.T) {
		t.Parallel()
		task := &Task{SourceRefs: []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}}}
		task.AddSourceRef("reminders", "x-apple-reminder://abc")

		if len(task.SourceRefs) != 2 {
			t.Errorf("expected 2 refs, got %d", len(task.SourceRefs))
		}
	})
}

func TestSourceRefEffectiveSourceProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sourceProvider string
		sourceRefs     []SourceRef
		want           string
	}{
		{
			name:           "no refs, no legacy",
			sourceProvider: "",
			sourceRefs:     nil,
			want:           "",
		},
		{
			name:           "legacy only",
			sourceProvider: "textfile",
			sourceRefs:     nil,
			want:           "textfile",
		},
		{
			name:           "refs take precedence",
			sourceProvider: "textfile",
			sourceRefs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			want:           "jira",
		},
		{
			name:       "multiple refs returns first",
			sourceRefs: []SourceRef{{Provider: "reminders", NativeID: "abc"}, {Provider: "jira", NativeID: "PROJ-42"}},
			want:       "reminders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{SourceProvider: tt.sourceProvider, SourceRefs: tt.sourceRefs}
			got := task.EffectiveSourceProvider()
			if got != tt.want {
				t.Errorf("EffectiveSourceProvider() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSourceRefMigrateSourceProvider(t *testing.T) {
	t.Parallel()

	t.Run("migrates legacy field", func(t *testing.T) {
		t.Parallel()
		task := &Task{
			ID:             "test-id-123",
			SourceProvider: "textfile",
		}
		task.MigrateSourceProvider()

		if len(task.SourceRefs) != 1 {
			t.Fatalf("expected 1 ref after migration, got %d", len(task.SourceRefs))
		}
		if task.SourceRefs[0].Provider != "textfile" {
			t.Errorf("expected provider %q, got %q", "textfile", task.SourceRefs[0].Provider)
		}
		if task.SourceRefs[0].NativeID != "test-id-123" {
			t.Errorf("expected native ID %q, got %q", "test-id-123", task.SourceRefs[0].NativeID)
		}
	})

	t.Run("no-op when refs already populated", func(t *testing.T) {
		t.Parallel()
		task := &Task{
			ID:             "test-id",
			SourceProvider: "textfile",
			SourceRefs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-1"}},
		}
		task.MigrateSourceProvider()

		if len(task.SourceRefs) != 1 {
			t.Errorf("expected 1 ref (unchanged), got %d", len(task.SourceRefs))
		}
		if task.SourceRefs[0].Provider != "jira" {
			t.Errorf("expected existing ref preserved, got %q", task.SourceRefs[0].Provider)
		}
	})

	t.Run("no-op when source provider empty", func(t *testing.T) {
		t.Parallel()
		task := &Task{ID: "test-id"}
		task.MigrateSourceProvider()

		if len(task.SourceRefs) != 0 {
			t.Errorf("expected 0 refs, got %d", len(task.SourceRefs))
		}
	})
}

func TestSourceRefYAMLRoundTrip(t *testing.T) {
	t.Parallel()

	task := NewTask("test task")
	task.AddSourceRef("jira", "PROJ-42")
	task.AddSourceRef("reminders", "x-apple-reminder://abc")

	data, err := yaml.Marshal(task)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
		return
	}

	var restored Task
	if err := yaml.Unmarshal(data, &restored); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}

	if len(restored.SourceRefs) != 2 {
		t.Fatalf("expected 2 refs after round-trip, got %d", len(restored.SourceRefs))
	}
	if restored.SourceRefs[0].Provider != "jira" || restored.SourceRefs[0].NativeID != "PROJ-42" {
		t.Errorf("first ref mismatch: %+v", restored.SourceRefs[0])
	}
	if restored.SourceRefs[1].Provider != "reminders" || restored.SourceRefs[1].NativeID != "x-apple-reminder://abc" {
		t.Errorf("second ref mismatch: %+v", restored.SourceRefs[1])
	}
}

func TestSourceRefJSONRoundTrip(t *testing.T) {
	t.Parallel()

	task := NewTask("test task")
	task.AddSourceRef("jira", "PROJ-42")

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
		return
	}

	var restored Task
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(restored.SourceRefs) != 1 {
		t.Fatalf("expected 1 ref after round-trip, got %d", len(restored.SourceRefs))
	}
	if restored.SourceRefs[0].Provider != "jira" || restored.SourceRefs[0].NativeID != "PROJ-42" {
		t.Errorf("ref mismatch: %+v", restored.SourceRefs[0])
	}
}

func TestSourceRefOmittedWhenEmpty(t *testing.T) {
	t.Parallel()

	task := NewTask("test task")
	data, err := yaml.Marshal(task)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
		return
	}

	yamlStr := string(data)
	if contains(yamlStr, "source_refs") {
		t.Error("expected source_refs to be omitted from YAML when empty")
	}
}

func TestMigrateTasks_Batch(t *testing.T) {
	t.Parallel()

	tasks := []*Task{
		{ID: "t1", SourceProvider: "textfile"},
		{ID: "t2", SourceProvider: "obsidian"},
		{ID: "t3", SourceProvider: ""}, // no source provider
		{ID: "t4", SourceProvider: "jira", SourceRefs: []SourceRef{{Provider: "jira", NativeID: "PROJ-1"}}},
	}

	MigrateTasks(tasks)

	// t1: should gain a SourceRef
	if len(tasks[0].SourceRefs) != 1 || tasks[0].SourceRefs[0].Provider != "textfile" {
		t.Errorf("t1: expected migration to textfile ref, got %+v", tasks[0].SourceRefs)
	}

	// t2: should gain a SourceRef
	if len(tasks[1].SourceRefs) != 1 || tasks[1].SourceRefs[0].Provider != "obsidian" {
		t.Errorf("t2: expected migration to obsidian ref, got %+v", tasks[1].SourceRefs)
	}

	// t3: no source provider, should remain empty
	if len(tasks[2].SourceRefs) != 0 {
		t.Errorf("t3: expected no refs for empty source provider, got %+v", tasks[2].SourceRefs)
	}

	// t4: already has refs, should not be modified
	if len(tasks[3].SourceRefs) != 1 || tasks[3].SourceRefs[0].NativeID != "PROJ-1" {
		t.Errorf("t4: existing refs should be preserved, got %+v", tasks[3].SourceRefs)
	}
}

func TestSourceRefBackwardCompatibility_EmptySourceRefs(t *testing.T) {
	t.Parallel()

	// Task with no SourceRefs should behave identically to pre-SourceRef behavior
	task := &Task{
		ID:             "t1",
		Text:           "legacy task",
		SourceProvider: "textfile",
	}

	// EffectiveSourceProvider falls back to legacy field
	if got := task.EffectiveSourceProvider(); got != "textfile" {
		t.Errorf("EffectiveSourceProvider() = %q, want %q", got, "textfile")
	}

	// HasSourceRef returns false
	if task.HasSourceRef("textfile", "t1") {
		t.Error("HasSourceRef() should return false for empty refs")
	}

	// FindBySourceRef on pool returns nil
	pool := NewTaskPool()
	pool.AddTask(task)
	if pool.FindBySourceRef("textfile", "t1") != nil {
		t.Error("FindBySourceRef() should return nil for task with no SourceRefs")
	}
}
