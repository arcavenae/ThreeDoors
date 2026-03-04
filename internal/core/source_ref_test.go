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
	}

	yamlStr := string(data)
	if contains(yamlStr, "source_refs") {
		t.Error("expected source_refs to be omitted from YAML when empty")
	}
}
