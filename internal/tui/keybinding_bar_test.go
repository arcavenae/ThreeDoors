package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func TestRenderKeybindingBar_Disabled(t *testing.T) {
	t.Parallel()
	got := RenderKeybindingBar(ViewDoors, 80, 24, false, false)
	if got != "" {
		t.Errorf("disabled bar should return empty string, got %q", got)
	}
}

func TestRenderKeybindingBar_HeightThresholds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		height      int
		wantEmpty   bool
		wantCompact bool
	}{
		{"height 5 hidden", 5, true, false},
		{"height 9 hidden", 9, true, false},
		{"height 10 compact", 10, false, true},
		{"height 12 compact", 12, false, true},
		{"height 15 compact", 15, false, true},
		{"height 16 full", 16, false, false},
		{"height 24 full", 24, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderKeybindingBar(ViewDoors, 80, tt.height, true, false)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("height %d: want empty, got %q", tt.height, got)
				}
				return
			}
			if got == "" {
				t.Fatalf("height %d: want non-empty bar", tt.height)
			}
			lines := strings.Split(got, "\n")
			if len(lines) < 2 {
				t.Fatalf("bar should have separator + content, got %d lines", len(lines))
			}
			barLine := stripANSI(lines[1])
			if tt.wantCompact {
				if strings.Contains(barLine, "select door") {
					t.Errorf("compact mode should not contain descriptions")
				}
			} else {
				if !strings.Contains(barLine, "select door") && !strings.Contains(barLine, "re-roll") {
					t.Errorf("full mode should contain descriptions, got %q", barLine)
				}
			}
		})
	}
}

func TestRenderKeybindingBar_WidthBreakpoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		width    int
		wantKeys []string
		maxKeys  int
	}{
		{"width 30 minimal", 30, []string{"?"}, 0},
		{"width 39 minimal", 39, []string{"?"}, 0},
		{"width 45 narrow", 45, []string{"?"}, 3},
		{"width 55 narrow", 55, []string{"?"}, 3},
		{"width 65 medium", 65, []string{"?"}, 5},
		{"width 75 medium", 75, []string{"?"}, 5},
		{"width 100 full", 100, []string{"?", "a/w/d", "s"}, 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderKeybindingBar(ViewDoors, tt.width, 24, true, false)
			if got == "" {
				t.Fatal("expected non-empty bar")
			}
			for _, key := range tt.wantKeys {
				if !strings.Contains(got, key) {
					t.Errorf("width %d: bar should contain key %q", tt.width, key)
				}
			}
		})
	}
}

func TestRenderKeybindingBar_HelpAlwaysLast(t *testing.T) {
	t.Parallel()
	modes := []struct {
		name         string
		mode         ViewMode
		doorSelected bool
	}{
		{"doors", ViewDoors, false},
		{"doors selected", ViewDoors, true},
		{"detail", ViewDetail, false},
		{"mood", ViewMood, false},
		{"search", ViewSearch, false},
		{"add task", ViewAddTask, false},
	}
	for _, tt := range modes {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderKeybindingBar(tt.mode, 100, 24, true, tt.doorSelected)
			if got == "" {
				t.Fatal("expected non-empty bar")
			}
			lines := strings.Split(got, "\n")
			barLine := stripANSI(lines[len(lines)-1])
			if !strings.Contains(barLine, "?") {
				t.Error("help key '?' not found in bar")
			}
			plainTrimmed := strings.TrimSpace(barLine)
			if !strings.HasSuffix(plainTrimmed, "? help") {
				t.Errorf("help should be last item, got %q", plainTrimmed)
			}
		})
	}
}

func TestRenderKeybindingBar_ContextSensitive(t *testing.T) {
	t.Parallel()
	doorsBar := RenderKeybindingBar(ViewDoors, 100, 24, true, false)
	detailBar := RenderKeybindingBar(ViewDetail, 100, 24, true, false)

	if doorsBar == detailBar {
		t.Error("doors and detail bars should have different content")
	}
	if !strings.Contains(doorsBar, "a/w/d") {
		t.Error("doors bar should contain 'a/w/d'")
	}
	if !strings.Contains(detailBar, "esc") {
		t.Error("detail bar should contain 'esc'")
	}
}

func TestRenderKeybindingBar_DoorSelectedVariant(t *testing.T) {
	t.Parallel()
	notSelected := RenderKeybindingBar(ViewDoors, 100, 24, true, false)
	selected := RenderKeybindingBar(ViewDoors, 100, 24, true, true)

	if notSelected == selected {
		t.Error("door selected and unselected bars should differ")
	}
	if !strings.Contains(selected, "enter") {
		t.Error("selected door bar should contain 'enter'")
	}
	if !strings.Contains(selected, "deselect") {
		t.Error("selected door bar should contain 'deselect'")
	}
}

func TestRenderKeybindingBar_SeparatorLine(t *testing.T) {
	t.Parallel()
	got := RenderKeybindingBar(ViewDoors, 80, 24, true, false)
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected separator + bar, got %d lines", len(lines))
	}
	plainSep := stripANSI(lines[0])
	if !strings.Contains(plainSep, "─") {
		t.Errorf("separator should contain ─ characters, got %q", plainSep)
	}
	expectedLen := len(strings.Repeat("─", 80))
	if len(plainSep) != expectedLen {
		t.Errorf("separator should be %d chars wide, got %d", expectedLen, len(plainSep))
	}
}

func TestRenderKeybindingBar_AllModes(t *testing.T) {
	t.Parallel()
	modes := []ViewMode{
		ViewDoors, ViewDetail, ViewMood, ViewSearch, ViewHealth,
		ViewAddTask, ViewValuesGoals, ViewFeedback, ViewImprovement,
		ViewNextSteps, ViewAvoidancePrompt, ViewInsights, ViewOnboarding,
		ViewConflict, ViewSyncLog, ViewThemePicker, ViewDevQueue, ViewProposals,
	}
	for _, mode := range modes {
		got := RenderKeybindingBar(mode, 80, 24, true, false)
		if got == "" {
			t.Errorf("mode %d: expected non-empty bar", mode)
		}
		if !strings.Contains(got, "?") {
			t.Errorf("mode %d: bar should always contain help key '?'", mode)
		}
	}
}

func TestFormatBar_Empty(t *testing.T) {
	t.Parallel()
	got := formatBar(nil, false, 80)
	if got != "" {
		t.Errorf("expected empty string for nil bindings, got %q", got)
	}
}

func TestFormatBar_TruncationKeepsHelp(t *testing.T) {
	t.Parallel()
	bindings := []KeyBinding{
		{Key: "a/w/d", Description: "select door", Priority: PriorityAlways},
		{Key: "s", Description: "re-roll", Priority: PriorityAlways},
		{Key: "n", Description: "add task", Priority: PriorityAlways},
		{Key: ":", Description: "command", Priority: PriorityAlways},
		{Key: "?", Description: "help", Priority: PriorityAlways},
	}

	got := formatBar(bindings, false, 15)
	if !strings.Contains(got, "?") {
		t.Error("truncated bar must always contain help key")
	}
}

// --- Golden file tests ---
// Golden tests use termenv.Ascii for deterministic output across environments.

const updateGoldenFiles = false // Set to true to regenerate golden files.

func testGoldenBar(t *testing.T, name, got string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name+".golden")

	if updateGoldenFiles {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v (run with updateGoldenFiles=true to create)", goldenPath, err)
	}
	if got != string(want) {
		t.Errorf("golden file mismatch %s:\n--- want ---\n%s\n--- got ---\n%s", goldenPath, string(want), got)
	}
}

func TestGolden_DoorsView_Full(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	got := RenderKeybindingBar(ViewDoors, 80, 24, true, false)
	testGoldenBar(t, "keybinding_bar_doors_full", got)
}

func TestGolden_DetailView_Full(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	got := RenderKeybindingBar(ViewDetail, 80, 24, true, false)
	testGoldenBar(t, "keybinding_bar_detail_full", got)
}

func TestGolden_CompactMode(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	got := RenderKeybindingBar(ViewDoors, 60, 12, true, false)
	testGoldenBar(t, "keybinding_bar_compact", got)
}

func TestGolden_TruncatedMode(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	got := RenderKeybindingBar(ViewDoors, 45, 24, true, false)
	testGoldenBar(t, "keybinding_bar_truncated", got)
}

func TestGolden_Disabled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	got := RenderKeybindingBar(ViewDoors, 80, 24, false, false)
	testGoldenBar(t, "keybinding_bar_disabled", got)
}
