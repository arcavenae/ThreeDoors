package themes

import "github.com/charmbracelet/lipgloss"

// ThemeColors holds the color palette for a door theme.
type ThemeColors struct {
	Frame    lipgloss.Color
	Fill     lipgloss.Color
	Accent   lipgloss.Color
	Selected lipgloss.Color
}

// DoorTheme defines the visual frame for a door.
type DoorTheme struct {
	Name        string
	Description string
	Render      func(content string, width int, selected bool) string
	Colors      ThemeColors
	MinWidth    int
}

// DefaultThemeName is the theme used when no theme is specified.
const DefaultThemeName = "modern"
