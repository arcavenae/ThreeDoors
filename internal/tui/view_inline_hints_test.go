package tui

import (
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestDetailViewInlineHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		enabled    bool
		wantHints  []string // substrings that must appear when hints enabled
		wantAbsent []string // substrings that must NOT appear when hints disabled
	}{
		{
			name:    "hints enabled shows key labels",
			enabled: true,
			wantHints: []string{
				"[esc]", "[c]", "[b]", "[e]", "[f]",
			},
		},
		{
			name:    "hints disabled shows no bracketed keys",
			enabled: false,
			wantAbsent: []string{
				"[esc]", "[c]", "[b]", "[e]", "[f]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := core.NewTask("Test task for detail view")
			dv := NewDetailView(task, nil, nil, nil)
			dv.SetWidth(80)
			dv.SetInlineHints(tt.enabled)

			output := dv.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected output to contain %q when hints enabled", hint)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(output, absent) {
					t.Errorf("expected output NOT to contain %q when hints disabled, but found it", absent)
				}
			}
		})
	}
}

func TestSearchViewInlineHints(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()

	tests := []struct {
		name       string
		enabled    bool
		wantHints  []string
		wantAbsent []string
	}{
		{
			name:    "hints enabled shows key labels",
			enabled: true,
			wantHints: []string{
				"[esc]", "[enter]",
			},
		},
		{
			name:    "hints disabled shows no bracketed keys",
			enabled: false,
			wantAbsent: []string{
				"[esc]", "[enter]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sv := NewSearchView(pool, nil, nil, nil, nil)
			sv.SetWidth(80)
			sv.SetInlineHints(tt.enabled)

			output := sv.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected output to contain %q when hints enabled", hint)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(output, absent) {
					t.Errorf("expected output NOT to contain %q when hints disabled, but found it", absent)
				}
			}
		})
	}
}

func TestMoodViewInlineHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		enabled    bool
		wantHints  []string
		wantAbsent []string
	}{
		{
			name:    "hints enabled shows numbered labels",
			enabled: true,
			wantHints: []string{
				"[1]", "[2]", "[3]", "[4]", "[5]", "[6]", "[7]",
			},
		},
		{
			name:    "hints disabled shows no bracketed numbers",
			enabled: false,
			wantAbsent: []string{
				"[1]", "[2]", "[3]", "[4]", "[5]", "[6]", "[7]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mv := NewMoodView()
			mv.SetWidth(80)
			mv.SetInlineHints(tt.enabled)

			output := mv.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected output to contain %q when hints enabled", hint)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(output, absent) {
					t.Errorf("expected output NOT to contain %q when hints disabled, but found it", absent)
				}
			}
		})
	}
}

func TestAddTaskViewInlineHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		enabled    bool
		wantHints  []string
		wantAbsent []string
	}{
		{
			name:    "hints enabled shows key labels",
			enabled: true,
			wantHints: []string{
				"[enter]", "[esc]",
			},
		},
		{
			name:    "hints disabled shows no bracketed keys",
			enabled: false,
			wantAbsent: []string{
				"[enter]", "[esc]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			av := NewAddTaskView()
			av.SetWidth(80)
			av.SetInlineHints(tt.enabled)

			output := av.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected output to contain %q when hints enabled", hint)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(output, absent) {
					t.Errorf("expected output NOT to contain %q when hints disabled, but found it", absent)
				}
			}
		})
	}
}

func TestHealthViewInlineHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		enabled    bool
		wantHints  []string
		wantAbsent []string
	}{
		{
			name:    "hints enabled shows esc label",
			enabled: true,
			wantHints: []string{
				"[esc]",
			},
		},
		{
			name:    "hints disabled shows no bracketed keys",
			enabled: false,
			wantAbsent: []string{
				"[esc]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := core.HealthCheckResult{
				Overall: core.HealthOK,
			}
			hv := NewHealthView(result)
			hv.SetWidth(80)
			hv.SetInlineHints(tt.enabled)

			output := hv.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected output to contain %q when hints enabled", hint)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(output, absent) {
					t.Errorf("expected output NOT to contain %q when hints disabled, but found it", absent)
				}
			}
		})
	}
}

func TestHelpViewNoInlineHints(t *testing.T) {
	t.Parallel()

	// Help view should NOT show inline hints since it IS the help.
	hv := NewHelpView()
	hv.SetWidth(80)
	hv.SetHeight(24)

	output := hv.View()

	// Help view should still show its normal navigation footer
	if !strings.Contains(output, "scroll") {
		t.Error("expected help view to contain scroll navigation text")
	}
}

func TestDetailViewHintsFromRegistry(t *testing.T) {
	t.Parallel()

	// Verify that detail bindings include the keys we render as hints.
	bindings := detailBindings()

	expectedKeys := map[string]bool{
		"q/esc/space/enter": false,
		"c":                 false,
		"b":                 false,
		"e":                 false,
		"f":                 false,
	}

	for _, group := range bindings {
		for _, b := range group.Bindings {
			if _, ok := expectedKeys[b.Key]; ok {
				expectedKeys[b.Key] = true
			}
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected detail bindings to include key %q", key)
		}
	}
}

func TestMoodViewHintsFromRegistry(t *testing.T) {
	t.Parallel()

	bindings := moodBindings()

	foundMoodSelect := false
	foundCustom := false
	for _, group := range bindings {
		for _, b := range group.Bindings {
			if b.Key == "1-6" {
				foundMoodSelect = true
			}
			if b.Key == "7" {
				foundCustom = true
			}
		}
	}

	if !foundMoodSelect {
		t.Error("expected mood bindings to include '1-6' for mood selection")
	}
	if !foundCustom {
		t.Error("expected mood bindings to include '7' for custom mood")
	}
}

func TestSearchViewHintsFromRegistry(t *testing.T) {
	t.Parallel()

	bindings := searchBindings()

	expectedKeys := map[string]bool{
		"esc":   false,
		"↑/↓":   false,
		"enter": false,
	}

	for _, group := range bindings {
		for _, b := range group.Bindings {
			if _, ok := expectedKeys[b.Key]; ok {
				expectedKeys[b.Key] = true
			}
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected search bindings to include key %q", key)
		}
	}
}

func TestAddTaskViewHintsFromRegistry(t *testing.T) {
	t.Parallel()

	bindings := addTaskBindings()

	expectedKeys := map[string]bool{
		"enter": false,
		"esc":   false,
	}

	for _, group := range bindings {
		for _, b := range group.Bindings {
			if _, ok := expectedKeys[b.Key]; ok {
				expectedKeys[b.Key] = true
			}
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected add task bindings to include key %q", key)
		}
	}
}

func TestHealthViewHintsFromRegistry(t *testing.T) {
	t.Parallel()

	bindings := healthBindings()

	foundEsc := false
	for _, group := range bindings {
		for _, b := range group.Bindings {
			if b.Key == "q/esc" {
				foundEsc = true
			}
		}
	}

	if !foundEsc {
		t.Error("expected health bindings to include 'q/esc' for back")
	}
}

func TestDetailViewSubModeHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      DetailViewMode
		wantHints []string
	}{
		{
			name:      "blocker input mode shows enter and esc",
			mode:      DetailModeBlockerInput,
			wantHints: []string{"Enter", "Esc"},
		},
		{
			name:      "expand input mode shows enter and esc",
			mode:      DetailModeExpandInput,
			wantHints: []string{"Enter", "Esc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := core.NewTask("Test task")
			dv := NewDetailView(task, nil, nil, nil)
			dv.SetWidth(80)
			dv.SetInlineHints(true)
			dv.mode = tt.mode

			output := dv.View()

			for _, hint := range tt.wantHints {
				if !strings.Contains(output, hint) {
					t.Errorf("expected sub-mode output to contain %q", hint)
				}
			}
		})
	}
}

func TestMoodViewCustomModeHints(t *testing.T) {
	t.Parallel()

	mv := NewMoodView()
	mv.SetWidth(80)
	mv.SetInlineHints(true)
	mv.isCustom = true

	output := mv.View()

	// Custom mode should show enter/esc hints
	if !strings.Contains(output, "Enter") && !strings.Contains(output, "enter") {
		t.Error("expected custom mood mode to contain Enter hint")
	}
}
