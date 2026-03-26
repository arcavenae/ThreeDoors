package reminders

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestParseRemindersConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		settings         map[string]string
		wantLists        []string
		wantIncCompleted bool
	}{
		{
			name:             "empty settings",
			settings:         nil,
			wantLists:        nil,
			wantIncCompleted: false,
		},
		{
			name:             "single list",
			settings:         map[string]string{"lists": "Work"},
			wantLists:        []string{"Work"},
			wantIncCompleted: false,
		},
		{
			name:             "multiple lists",
			settings:         map[string]string{"lists": "Work,ThreeDoors,Personal"},
			wantLists:        []string{"Work", "ThreeDoors", "Personal"},
			wantIncCompleted: false,
		},
		{
			name:             "lists with spaces",
			settings:         map[string]string{"lists": " Work , ThreeDoors "},
			wantLists:        []string{"Work", "ThreeDoors"},
			wantIncCompleted: false,
		},
		{
			name:             "include completed true",
			settings:         map[string]string{"include_completed": "true"},
			wantLists:        nil,
			wantIncCompleted: true,
		},
		{
			name:             "include completed TRUE",
			settings:         map[string]string{"include_completed": "TRUE"},
			wantLists:        nil,
			wantIncCompleted: true,
		},
		{
			name:             "include completed false",
			settings:         map[string]string{"include_completed": "false"},
			wantLists:        nil,
			wantIncCompleted: false,
		},
		{
			name:             "include completed invalid",
			settings:         map[string]string{"include_completed": "yes"},
			wantLists:        nil,
			wantIncCompleted: false,
		},
		{
			name:             "all settings",
			settings:         map[string]string{"lists": "Work,ThreeDoors", "include_completed": "true"},
			wantLists:        []string{"Work", "ThreeDoors"},
			wantIncCompleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &core.ProviderConfig{
				Providers: []core.ProviderEntry{
					{Name: "reminders", Settings: tt.settings},
				},
			}

			rc := ParseRemindersConfig(cfg)

			if len(rc.Lists) != len(tt.wantLists) {
				t.Fatalf("Lists: got %v, want %v", rc.Lists, tt.wantLists)
			}
			for i, got := range rc.Lists {
				if got != tt.wantLists[i] {
					t.Errorf("Lists[%d]: got %q, want %q", i, got, tt.wantLists[i])
				}
			}

			if rc.IncludeCompleted != tt.wantIncCompleted {
				t.Errorf("IncludeCompleted: got %v, want %v", rc.IncludeCompleted, tt.wantIncCompleted)
			}
		})
	}
}

func TestParseRemindersConfig_NoRemindersEntry(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		Providers: []core.ProviderEntry{
			{Name: "textfile", Settings: map[string]string{"task_file": "tasks.yaml"}},
		},
	}

	rc := ParseRemindersConfig(cfg)
	if len(rc.Lists) != 0 {
		t.Errorf("expected no lists, got %v", rc.Lists)
	}
	if rc.IncludeCompleted {
		t.Error("expected IncludeCompleted false")
	}
}

func TestParseRemindersConfig_EmptyProviders(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{}

	rc := ParseRemindersConfig(cfg)
	if len(rc.Lists) != 0 {
		t.Errorf("expected no lists, got %v", rc.Lists)
	}
}

func TestParseLists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single", "Work", []string{"Work"}},
		{"multiple", "A,B,C", []string{"A", "B", "C"}},
		{"trailing comma", "A,B,", []string{"A", "B"}},
		{"leading comma", ",A,B", []string{"A", "B"}},
		{"spaces only", " , , ", nil},
		{"mixed spacing", "Work , Personal, ThreeDoors ", []string{"Work", "Personal", "ThreeDoors"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseLists(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseLists(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseLists(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"false", false},
		{"False", false},
		{"yes", false},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := parseBool(tt.input); got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
