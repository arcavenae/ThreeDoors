package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ThemeSelectedMsg is sent when the user selects or skips the theme picker.
type ThemeSelectedMsg struct {
	ThemeName string
	Skipped   bool
}

// ThemePicker is a reusable Bubbletea component that lets users browse
// and select a door theme with a live preview.
type ThemePicker struct {
	registry   *themes.Registry
	themeNames []string
	cursor     int
	confirmed  bool
	width      int
}

// NewThemePicker creates a theme picker initialized with all themes from the registry.
func NewThemePicker(reg *themes.Registry) *ThemePicker {
	return &ThemePicker{
		registry:   reg,
		themeNames: reg.Names(),
	}
}

// SetWidth sets the terminal width for rendering.
func (tp *ThemePicker) SetWidth(w int) {
	tp.width = w
}

// SelectedThemeName returns the name of the currently highlighted theme.
func (tp *ThemePicker) SelectedThemeName() string {
	if len(tp.themeNames) == 0 {
		return themes.DefaultThemeName
	}
	return tp.themeNames[tp.cursor]
}

// Update handles key input for the theme picker.
func (tp *ThemePicker) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyLeft:
		tp.cursor--
		if tp.cursor < 0 {
			tp.cursor = len(tp.themeNames) - 1
		}
	case tea.KeyRight:
		tp.cursor++
		if tp.cursor >= len(tp.themeNames) {
			tp.cursor = 0
		}
	case tea.KeyEnter:
		tp.confirmed = true
		name := tp.SelectedThemeName()
		return func() tea.Msg {
			return ThemeSelectedMsg{ThemeName: name}
		}
	case tea.KeyEscape:
		tp.confirmed = true
		return func() tea.Msg {
			return ThemeSelectedMsg{ThemeName: themes.DefaultThemeName, Skipped: true}
		}
	default:
		key := keyMsg.String()
		switch key {
		case "left", "h":
			tp.cursor--
			if tp.cursor < 0 {
				tp.cursor = len(tp.themeNames) - 1
			}
		case "right", "l":
			tp.cursor++
			if tp.cursor >= len(tp.themeNames) {
				tp.cursor = 0
			}
		}
	}
	return nil
}

// View renders the theme picker with a live preview.
func (tp *ThemePicker) View() string {
	if len(tp.themeNames) == 0 {
		return "No themes available."
	}

	w := tp.width - 6
	if w < 40 {
		w = 40
	}

	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Choose Your Door Theme"))

	currentName := tp.themeNames[tp.cursor]
	theme, ok := tp.registry.Get(currentName)
	if !ok {
		fmt.Fprintf(&s, "Theme not found: %s\n", currentName)
		return s.String()
	}

	// Theme name and description
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	fmt.Fprintf(&s, "  %s\n", nameStyle.Render(theme.Name))
	fmt.Fprintf(&s, "  %s\n\n", helpStyle.Render(theme.Description))

	// Preview: render sample doors using this theme
	tp.renderPreview(&s, theme, w)

	fmt.Fprintf(&s, "\n")

	// Navigation indicator
	indicator := tp.navigationIndicator()
	fmt.Fprintf(&s, "  %s\n\n", indicator)

	fmt.Fprintf(&s, "%s\n", helpStyle.Render("← → to browse | Enter to confirm | Esc to skip"))

	return s.String()
}

func (tp *ThemePicker) renderPreview(s *strings.Builder, theme *themes.DoorTheme, maxWidth int) {
	sampleText := "Sample Task"

	// Check if we can fit 3 doors side-by-side
	doorWidth := theme.MinWidth + 4
	if doorWidth < 18 {
		doorWidth = 18
	}
	totalWidth := doorWidth*3 + 4 // 3 doors + gaps

	if maxWidth >= totalWidth {
		// Horizontal layout: 3 doors side-by-side
		doors := make([]string, 3)
		labels := []string{"Door 1", "Door 2", "Door 3"}
		for i := 0; i < 3; i++ {
			rendered := theme.Render(labels[i]+"\n"+sampleText, doorWidth, i == 1)
			doors[i] = rendered
		}
		joined := lipgloss.JoinHorizontal(lipgloss.Top, doors[0], "  ", doors[1], "  ", doors[2])
		fmt.Fprintf(s, "%s\n", joined)
	} else {
		// Narrow terminal: single door preview
		singleWidth := maxWidth - 4
		if singleWidth < theme.MinWidth {
			singleWidth = theme.MinWidth
		}
		rendered := theme.Render(sampleText, singleWidth, true)
		fmt.Fprintf(s, "%s\n", rendered)
	}
}

func (tp *ThemePicker) navigationIndicator() string {
	var dots strings.Builder
	for i := range tp.themeNames {
		if i == tp.cursor {
			dots.WriteString("●")
		} else {
			dots.WriteString("○")
		}
		if i < len(tp.themeNames)-1 {
			dots.WriteString(" ")
		}
	}
	return dots.String()
}
