package tui

import (
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/tui/themes"
)

// --- Story 36.1: Enhanced Door Selection Visual Feedback ---

// AC1+AC3: Theme path — selected door content differs from unselected door content
func TestDoorsView_ThemePath_SelectedContentDiffersFromUnselected(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("Alpha", "Beta", "Gamma")
	dv.SetWidth(120)

	var selectedContent string
	var unselectedContents []string
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "capture-theme",
		Description: "captures content for testing",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			if selected {
				selectedContent = content
			} else {
				unselectedContents = append(unselectedContents, content)
			}
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("capture-theme")
	dv.selectedDoorIndex = 1

	_ = dv.View()

	if selectedContent == "" {
		t.Fatal("selected door content should not be empty")
	}
	if len(unselectedContents) == 0 {
		t.Fatal("unselected door content should not be empty")
	}

	// Selected content must differ from unselected content beyond just the task text
	// (both get styling applied before passing to theme)
	if selectedContent == unselectedContents[0] {
		t.Error("selected and unselected door content should differ due to emphasis styling")
	}
}

// AC2: No selection (index -1) means all doors get identical styling treatment
func TestDoorsView_ThemePath_NoSelection_AllContentUniform(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("SameTask", "SameTask", "SameTask")
	dv.SetWidth(120)

	var capturedContents []string
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "capture-theme",
		Description: "captures content for testing",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			capturedContents = append(capturedContents, content)
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("capture-theme")
	// selectedDoorIndex is -1 by default

	_ = dv.View()

	if len(capturedContents) < 3 {
		t.Fatalf("expected 3 captured contents, got %d", len(capturedContents))
	}

	// With identical task text and no selection, all content should be styled the same way
	if capturedContents[0] != capturedContents[1] || capturedContents[1] != capturedContents[2] {
		t.Error("with no selection and identical tasks, all door content styling should be uniform")
	}
}

// AC2: No selection means none of the doors are passed as selected to the theme
func TestDoorsView_ThemePath_NoSelection_NoneMarkedSelected(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("Alpha", "Beta", "Gamma")
	dv.SetWidth(120)

	selectedCount := 0
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "capture-theme",
		Description: "captures content for testing",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			if selected {
				selectedCount++
			}
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("capture-theme")
	// selectedDoorIndex is -1 by default

	_ = dv.View()

	if selectedCount != 0 {
		t.Errorf("with no selection, no doors should be marked selected; got %d", selectedCount)
	}
}

// AC1: Exactly one door is marked selected when selectedDoorIndex >= 0
func TestDoorsView_ThemePath_ExactlyOneSelected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		index int
	}{
		{"door 0", 0},
		{"door 1", 1},
		{"door 2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dv := newTestDoorsView("Alpha", "Beta", "Gamma")
			dv.SetWidth(120)

			selectedCount := 0
			unselectedCount := 0
			registry := themes.NewRegistry()
			registry.Register(&themes.DoorTheme{
				Name:        "count-theme",
				Description: "counts selection states",
				Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
					if selected {
						selectedCount++
					} else {
						unselectedCount++
					}
					return content
				},
				MinWidth: 10,
			})
			dv.SetThemeRegistry(registry)
			dv.SetThemeByName("count-theme")
			dv.selectedDoorIndex = tt.index

			_ = dv.View()

			if selectedCount != 1 {
				t.Errorf("expected exactly 1 selected door, got %d", selectedCount)
			}
			if unselectedCount != 2 {
				t.Errorf("expected exactly 2 unselected doors, got %d", unselectedCount)
			}
		})
	}
}

// AC4: Fallback (non-theme) rendering produces different output for selected vs no-selection
func TestDoorsView_FallbackPath_SelectedVsNoSelection(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("Alpha", "Beta", "Gamma")
	dv.SetWidth(120)
	// No theme — uses fallback path

	dv.selectedDoorIndex = -1
	viewNoSelection := dv.View()

	dv.selectedDoorIndex = 1
	viewWithSelection := dv.View()

	if viewNoSelection == viewWithSelection {
		t.Error("view output should differ between no-selection and selected states in fallback path")
	}
}

// AC4: Fallback path — selected door uses DoubleBorder (structural emphasis visible in plain text)
func TestDoorsView_FallbackPath_SelectedDoorDoubleFrame(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("Alpha", "Beta", "Gamma")
	dv.SetWidth(120)
	dv.selectedDoorIndex = 0
	// No theme — uses fallback path

	view := dv.View()

	// DoubleBorder uses box-drawing characters: ║ (U+2551) and ═ (U+2550)
	hasDoubleVertical := strings.Contains(view, "\u2551")
	hasDoubleHorizontal := strings.Contains(view, "\u2550")
	if !hasDoubleVertical && !hasDoubleHorizontal {
		t.Error("selected door in fallback mode should use double border characters for structural emphasis")
	}
}

// AC5: Style objects have the correct properties

func TestSelectedDoorStyle_HasDoubleBorder(t *testing.T) {
	t.Parallel()
	// Verify the style uses DoubleBorder by checking rendered output
	rendered := selectedDoorStyle.Width(30).Render("test")
	// DoubleBorder uses ║ (U+2551)
	if !strings.Contains(rendered, "\u2551") {
		t.Error("selectedDoorStyle should use DoubleBorder")
	}
}

func TestUnselectedDoorStyle_Exists(t *testing.T) {
	t.Parallel()
	// Verify the unselected style can render without panic
	rendered := unselectedDoorStyle.Width(30).Render("test")
	if rendered == "" {
		t.Error("unselectedDoorStyle should produce non-empty output")
	}
}

func TestSelectedContentStyle_Exists(t *testing.T) {
	t.Parallel()
	rendered := selectedContentStyle.Render("test")
	if rendered == "" {
		t.Error("selectedContentStyle should produce non-empty output")
	}
}

func TestUnselectedContentStyle_Exists(t *testing.T) {
	t.Parallel()
	rendered := unselectedContentStyle.Render("test")
	if rendered == "" {
		t.Error("unselectedContentStyle should produce non-empty output")
	}
}

// Table-driven test for all selection indices
func TestDoorsView_SelectionStates_ThemePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		selectedIndex     int
		wantSelectedCount int
		wantContentDiffer bool
	}{
		{"no selection", -1, 0, false},
		{"door 0 selected", 0, 1, true},
		{"door 1 selected", 1, 1, true},
		{"door 2 selected", 2, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dv := newTestDoorsView("Alpha", "Beta", "Gamma")
			dv.SetWidth(120)

			selectedCount := 0
			var selectedContent, firstUnselectedContent string
			registry := themes.NewRegistry()
			registry.Register(&themes.DoorTheme{
				Name:        "state-theme",
				Description: "tracks selection state",
				Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
					if selected {
						selectedCount++
						selectedContent = content
					} else if firstUnselectedContent == "" {
						firstUnselectedContent = content
					}
					return content
				},
				MinWidth: 10,
			})
			dv.SetThemeRegistry(registry)
			dv.SetThemeByName("state-theme")
			dv.selectedDoorIndex = tt.selectedIndex

			_ = dv.View()

			if selectedCount != tt.wantSelectedCount {
				t.Errorf("selected count: got %d, want %d", selectedCount, tt.wantSelectedCount)
			}

			if tt.wantContentDiffer && selectedContent != "" && firstUnselectedContent != "" {
				if selectedContent == firstUnselectedContent {
					t.Error("selected and unselected content should have different styling applied")
				}
			}
		})
	}
}
