package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ThemePicker is a Bubbletea component for selecting a door theme.
type ThemePicker struct {
	registry     *themes.Registry
	themeNames   []string
	cursor       int
	currentTheme string
	width        int
	height       int
	title        string
	seasonal     bool
}

// NewThemePicker creates a ThemePicker showing non-seasonal themes,
// with cursor positioned on the current theme.
func NewThemePicker(registry *themes.Registry, currentTheme string) *ThemePicker {
	names := registry.NonSeasonalNames()
	cursor := 0
	for i, name := range names {
		if name == currentTheme {
			cursor = i
			break
		}
	}
	return &ThemePicker{
		registry:     registry,
		themeNames:   names,
		cursor:       cursor,
		currentTheme: currentTheme,
		title:        "Select Door Theme",
	}
}

// NewSeasonalThemePicker creates a ThemePicker showing only seasonal themes
// (themes where Season != ""), with cursor positioned on the current theme.
func NewSeasonalThemePicker(registry *themes.Registry, currentTheme string) *ThemePicker {
	names := registry.SeasonalNames()
	cursor := 0
	for i, name := range names {
		if name == currentTheme {
			cursor = i
			break
		}
	}
	return &ThemePicker{
		registry:     registry,
		themeNames:   names,
		cursor:       cursor,
		currentTheme: currentTheme,
		title:        "Select Seasonal Theme",
		seasonal:     true,
	}
}

// IsSeasonal reports whether this picker shows seasonal themes.
func (tp *ThemePicker) IsSeasonal() bool {
	return tp.seasonal
}

// SetWidth sets the available width for rendering.
func (tp *ThemePicker) SetWidth(w int) {
	tp.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (tp *ThemePicker) SetHeight(h int) {
	tp.height = h
}

// Update handles key input for the theme picker.
func (tp *ThemePicker) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return func() tea.Msg { return ThemeCancelledMsg{} }
		}
		switch msg.Type {
		case tea.KeyEscape:
			return func() tea.Msg { return ThemeCancelledMsg{} }
		case tea.KeyEnter:
			name := tp.themeNames[tp.cursor]
			return func() tea.Msg { return ThemeSelectedMsg{Name: name} }
		case tea.KeyLeft, tea.KeyUp:
			if tp.cursor > 0 {
				tp.cursor--
			}
		case tea.KeyRight, tea.KeyDown:
			if tp.cursor < len(tp.themeNames)-1 {
				tp.cursor++
			}
		}
	}
	return nil
}

// View renders the theme picker list with descriptions and current indicator.
func (tp *ThemePicker) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	fmt.Fprintf(&s, "%s\n\n", titleStyle.Render(tp.title))

	for i, name := range tp.themeNames {
		theme, _ := tp.registry.Get(name)

		cursor := "  "
		if i == tp.cursor {
			cursor = "▸ "
		}

		label := name
		if name == tp.currentTheme {
			label += " [current]"
		}

		nameStyle := lipgloss.NewStyle()
		if i == tp.cursor {
			nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("255"))
		} else {
			nameStyle = nameStyle.Foreground(lipgloss.Color("245"))
		}

		desc := ""
		if theme != nil {
			desc = theme.Description
		}
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

		fmt.Fprintf(&s, "%s%s", cursor, nameStyle.Render(label))
		if desc != "" {
			fmt.Fprintf(&s, "  %s", descStyle.Render(desc))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fmt.Fprintf(&s, "%s", helpStyle.Render("↑/↓ navigate | Enter select | Esc cancel"))

	return s.String()
}
