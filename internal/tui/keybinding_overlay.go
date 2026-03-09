package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverlayState holds the context for the keybinding overlay.
type OverlayState struct {
	ViewMode ViewMode
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

// KeybindingOverlay manages a full-screen keybinding reference panel with viewport scrolling.
type KeybindingOverlay struct {
	viewport viewport.Model
	state    OverlayState
	width    int
	height   int
	ready    bool
}

// NewKeybindingOverlay creates a new overlay with viewport-based scrolling.
func NewKeybindingOverlay(state OverlayState, width, height int) *KeybindingOverlay {
	ko := &KeybindingOverlay{
		state:  state,
		width:  width,
		height: height,
	}
	ko.initViewport()
	return ko
}

// initViewport creates and configures the viewport for the overlay content.
func (ko *KeybindingOverlay) initViewport() {
	// Available height inside the border: total height minus border (2) minus title (1)
	// minus blank-after-title (1) minus footer (1).
	innerHeight := ko.height - 5
	if innerHeight < 1 {
		innerHeight = 1
	}

	innerWidth := ko.innerWidth()

	ko.viewport = NewScrollableView(innerWidth, innerHeight)
	ko.viewport.SetContent(ko.renderContent())
	ko.ready = true
}

// innerWidth calculates the usable width inside the border.
func (ko *KeybindingOverlay) innerWidth() int {
	// Inner width: total width minus border sides (2) minus padding (2).
	w := ko.width - 4
	if w < 10 {
		w = 10
	}
	return w
}

// renderContent builds the keybinding content lines as a single string.
func (ko *KeybindingOverlay) renderContent() string {
	groups := allKeyBindingGroups()
	groups = reorderGroupsForView(groups, ko.state.ViewMode)

	var s strings.Builder
	for i, g := range groups {
		if i > 0 {
			s.WriteString("\n")
		}
		label := g.Name
		if i == 0 {
			label += " (current)"
		}
		s.WriteString(overlayCategoryStyle.Render(label))
		s.WriteString("\n")

		for _, b := range g.Bindings {
			s.WriteString(formatBinding(b))
			s.WriteString("\n")
		}
	}
	return s.String()
}

// Update handles key events for the overlay.
func (ko *KeybindingOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	ko.viewport, cmd = ko.viewport.Update(msg)
	return cmd
}

// View renders the full overlay with border, title, viewport content, and footer.
func (ko *KeybindingOverlay) View() string {
	if ko.width < 20 || ko.height < 5 {
		return ""
	}

	innerWidth := ko.innerWidth()

	var body strings.Builder
	title := overlayTitleStyle.Render(overlayTitle)
	body.WriteString(lipgloss.Place(innerWidth, 1, lipgloss.Center, lipgloss.Top, title))
	body.WriteString("\n\n")

	body.WriteString(ko.viewport.View())

	// Footer with scroll indicator.
	scrollPct := ko.viewport.ScrollPercent()
	footer := overlayFooterStyle.Render(overlayFooter)

	// Show scroll indicators based on position.
	if ko.viewport.TotalLineCount() > ko.viewport.Height {
		var indicator string
		atTop := scrollPct <= 0
		atBottom := scrollPct >= 1
		if !atTop && !atBottom {
			indicator = "  ▲▼ more"
		} else if atTop {
			indicator = "  ▼ more"
		} else {
			indicator = "  ▲ more"
		}
		footer = overlayFooterStyle.Render(overlayFooter + indicator)
	}
	fmt.Fprintf(&body, "\n%s", footer)

	bordered := overlayBorderStyle.
		Width(innerWidth).
		Render(body.String())

	return bordered
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
