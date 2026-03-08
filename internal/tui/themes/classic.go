package themes

import "github.com/charmbracelet/lipgloss"

// NewClassicTheme creates the Classic theme that wraps the existing
// Lipgloss doorStyle/selectedDoorStyle rendering from the TUI.
func NewClassicTheme() *DoorTheme {
	frameColor := lipgloss.Color("63")
	selectedColor := lipgloss.Color("255")

	unselectedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(frameColor).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(selectedColor).
		Padding(1, 2)

	return &DoorTheme{
		Name:        "classic",
		Description: "Classic rounded border — the original ThreeDoors look",
		Render: func(content string, width int, _ int, selected bool) string {
			if selected {
				return selectedStyle.Width(width).Render(content)
			}
			return unselectedStyle.Width(width).Render(content)
		},
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("0"),
			Accent:   frameColor,
			Selected: selectedColor,
		},
		MinWidth:  15,
		MinHeight: 10,
	}
}
