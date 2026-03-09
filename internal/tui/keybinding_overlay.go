package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// OverlayState holds the scroll position and context for the keybinding overlay.
type OverlayState struct {
	ScrollOffset int
	ViewMode     ViewMode
}

// Overlay styling — theme-neutral to work with any door theme.
var (
	overlayBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("255"))

	overlayTitleStyle = lipgloss.NewStyle().
				Bold(true)

	overlayCategoryStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true)

	overlayKeyStyle = lipgloss.NewStyle().
			Bold(true)

	overlayFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))
)

const (
	overlayTitle  = "KEYBINDING REFERENCE"
	overlayFooter = "Press ? or esc to close   ↑/↓ to scroll"
)

// RenderKeybindingOverlay renders a full-screen keybinding reference panel.
// It shows all keybindings organized by category, with the current view's
// group highlighted at the top. Supports scrolling when content exceeds height.
func RenderKeybindingOverlay(state OverlayState, width, height int) string {
	if width < 20 || height < 5 {
		return ""
	}

	groups := allKeyBindingGroups()
	groups = reorderGroupsForView(groups, state.ViewMode)

	// Build content lines (without border/title/footer).
	var contentLines []string
	for i, g := range groups {
		if i > 0 {
			contentLines = append(contentLines, "")
		}
		label := g.Name
		if i == 0 {
			label += " (current)"
		}
		contentLines = append(contentLines, overlayCategoryStyle.Render(label))

		for _, b := range g.Bindings {
			contentLines = append(contentLines, formatBinding(b))
		}
	}

	// Available height inside the border: total height minus border (2) minus title (1)
	// minus blank-after-title (1) minus footer (1).
	innerHeight := height - 5
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Inner width: total width minus border sides (2) minus padding (2).
	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Clamp scroll offset.
	maxScroll := len(contentLines) - innerHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scrollOffset := state.ScrollOffset
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}

	// Slice visible content.
	endIdx := scrollOffset + innerHeight
	if endIdx > len(contentLines) {
		endIdx = len(contentLines)
	}
	visible := contentLines[scrollOffset:endIdx]

	// Build the inner body: title + blank line + visible content.
	var body strings.Builder
	title := overlayTitleStyle.Render(overlayTitle)
	// Center the title within inner width.
	titleLen := lipgloss.Width(title)
	titlePad := (innerWidth - titleLen) / 2
	if titlePad < 0 {
		titlePad = 0
	}
	fmt.Fprintf(&body, "%s%s\n", strings.Repeat(" ", titlePad), title)
	body.WriteString("\n")
	for i, line := range visible {
		body.WriteString(line)
		if i < len(visible)-1 {
			body.WriteString("\n")
		}
	}

	// Pad remaining lines to fill the box.
	linesWritten := len(visible)
	for linesWritten < innerHeight {
		body.WriteString("\n")
		linesWritten++
	}

	// Add scroll indicator if there's more content below.
	hasMoreBelow := scrollOffset+innerHeight < len(contentLines)
	hasMoreAbove := scrollOffset > 0

	footer := overlayFooterStyle.Render(overlayFooter)
	if hasMoreBelow || hasMoreAbove {
		var indicator string
		if hasMoreAbove && hasMoreBelow {
			indicator = "  ▲▼ more"
		} else if hasMoreBelow {
			indicator = "  ▼ more"
		} else {
			indicator = "  ▲ more"
		}
		footer = overlayFooterStyle.Render(overlayFooter + indicator)
	}
	fmt.Fprintf(&body, "\n%s", footer)

	// Apply the border.
	bordered := overlayBorderStyle.
		Width(innerWidth).
		Render(body.String())

	return bordered
}

// ClampScrollOffset clamps a scroll offset for the given overlay dimensions.
func ClampScrollOffset(offset, height int) int {
	groups := allKeyBindingGroups()
	contentLines := countContentLines(groups)

	innerHeight := height - 5
	if innerHeight < 1 {
		innerHeight = 1
	}

	maxScroll := contentLines - innerHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if offset < 0 {
		return 0
	}
	if offset > maxScroll {
		return maxScroll
	}
	return offset
}

// countContentLines counts how many lines the overlay content occupies.
func countContentLines(groups []KeyBindingGroup) int {
	total := 0
	for i, g := range groups {
		if i > 0 {
			total++ // blank separator line
		}
		total++ // category header
		total += len(g.Bindings)
	}
	return total
}

// reorderGroupsForView returns a copy of groups with the current view's
// primary group moved to the front. The matching is based on which group
// contains the most bindings from the current view.
func reorderGroupsForView(groups []KeyBindingGroup, mode ViewMode) []KeyBindingGroup {
	viewGroups := viewKeyBindings(mode, false)
	if len(viewGroups) == 0 {
		return groups
	}

	// Find the name of the first group from the current view.
	primaryName := viewGroups[0].Name

	// Find and move the matching group to front.
	result := make([]KeyBindingGroup, 0, len(groups))
	var primary *KeyBindingGroup
	for i := range groups {
		if groups[i].Name == primaryName && primary == nil {
			primary = &groups[i]
		} else {
			result = append(result, groups[i])
		}
	}
	if primary != nil {
		return append([]KeyBindingGroup{*primary}, result...)
	}
	return groups
}

// formatBinding formats a single keybinding as "  key     description"
// with consistent column alignment.
func formatBinding(b KeyBinding) string {
	key := overlayKeyStyle.Render(b.Key)
	keyWidth := lipgloss.Width(key)
	// Pad to 12-char column for key field.
	pad := 12 - keyWidth
	if pad < 1 {
		pad = 1
	}
	return fmt.Sprintf("  %s%s%s", key, strings.Repeat(" ", pad), b.Description)
}
