package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewSciFiTheme creates the Sci-Fi/Spaceship theme: double-line outer frame,
// shade-filled side rails, upper content panel, lower control panel with
// ACCESS label. When selected, uses bright shade (▓) instead of light (░).
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
		MinWidth: 18,
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

		// Layout: ║░░│ content │░░║
		// Rail width: 2 shade chars on each side
		// Total border overhead: 2 (║) + 4 (░░ x2) + 2 (│) = 8
		railW := 2
		innerBorder := 8
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

		// Top border: ╔══╤════════════════╤══╗
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

		// Blank line after content
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// Mid-bar: ╠══╪════════════════╪══╣
		fmt.Fprintf(&b, "%s\n", style.Render(
			"╠"+strings.Repeat("═", railW)+"╪"+strings.Repeat("═", contentW)+"╪"+strings.Repeat("═", railW)+"╣"))

		// Lower control panel

		// Shade line
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// ACCESS label centered
		label := "[ACCESS]"
		labelPad := (contentW - len(label)) / 2
		if labelPad < 0 {
			labelPad = 0
		}
		rightPad := contentW - labelPad - len(label)
		if rightPad < 0 {
			rightPad = 0
		}
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"),
			strings.Repeat(" ", labelPad)+label+strings.Repeat(" ", rightPad),
			style.Render("│"), rail, style.Render("║"))

		// Blank line
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

		// Bottom border: ╚══╧════════════════╧══╝
		fmt.Fprintf(&b, "%s", style.Render(
			"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+"╝"))

		return b.String()
	}
}
