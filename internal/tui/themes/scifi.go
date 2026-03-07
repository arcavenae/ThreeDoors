package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewSciFiTheme creates the Sci-Fi/Spaceship theme: double-line outer frame,
// shade-filled side rails, single content panel with inline ACCESS label.
// When selected, uses bright shade (▓) instead of light (░).
func NewSciFiTheme() *DoorTheme {
	frameColor := lipgloss.Color("39")
	selectedColor := lipgloss.Color("51")

	return &DoorTheme{
		Name:        "scifi",
		Description: "Sci-fi spaceship — double-line frame, shade panels, ACCESS label",
		Render:      scifiRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("236"),
			Accent:   lipgloss.Color("45"),
			Selected: selectedColor,
		},
		MinWidth: 16,
	}
}

func scifiRender(frameColor, selectedColor lipgloss.Color) func(string, int, bool) string {
	return func(content string, width int, selected bool) string {
		color := frameColor
		shadeChar := "░"
		if selected {
			color = selectedColor
			shadeChar = "▓"
		}
		style := lipgloss.NewStyle().Foreground(color)

		// Layout: ║░│ content │░║
		// Rail width: 1 shade char on each side
		// Total border overhead: 2 (║) + 2 (░ x2) + 2 (│) = 6
		railW := 1
		innerBorder := 6
		contentW := width - innerBorder
		if contentW < 1 {
			contentW = 1
		}

		// Word-wrap content with 2-char padding on each side
		textW := contentW - 4
		if textW < 1 {
			textW = 1
		}
		wrapped := ansi.Wordwrap(content, textW, "")
		contentLines := strings.Split(wrapped, "\n")

		rail := strings.Repeat(shadeChar, railW)

		var b strings.Builder

		// Top border: ╔═╤══════════════════════╤═╗
		fmt.Fprintf(&b, "%s\n", style.Render(
			"╔"+strings.Repeat("═", railW)+"╤"+strings.Repeat("═", contentW)+"╤"+strings.Repeat("═", railW)+"╗"))

		blankContent := strings.Repeat(" ", contentW)

		// Blank line
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// Content lines with 2-char padding
		for _, line := range contentLines {
			lineWidth := ansi.StringWidth(line)
			pad := contentW - 2 - lineWidth
			if pad < 0 {
				pad = 0
			}
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
				style.Render("║"), rail, style.Render("│"),
				"  "+line+strings.Repeat(" ", pad),
				style.Render("│"), rail, style.Render("║"))
		}

		// Blank lines after content
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// ACCESS label right-aligned with 2-char padding
		label := "[ACCESS]"
		leftPad := contentW - len(label) - 2
		if leftPad < 0 {
			leftPad = 0
		}
		labelRight := contentW - leftPad - len(label)
		if labelRight < 0 {
			labelRight = 0
		}
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"),
			strings.Repeat(" ", leftPad)+label+strings.Repeat(" ", labelRight),
			style.Render("│"), rail, style.Render("║"))

		// Blank line after ACCESS
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// Bottom border: ╚═╧══════════════════════╧═╝
		fmt.Fprintf(&b, "%s", style.Render(
			"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+"╝"))

		return b.String()
	}
}
