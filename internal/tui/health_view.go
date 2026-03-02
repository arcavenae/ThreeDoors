package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// HealthView displays health check results.
type HealthView struct {
	result tasks.HealthCheckResult
	width  int
}

// NewHealthView creates a new HealthView with the given result.
func NewHealthView(result tasks.HealthCheckResult) *HealthView {
	return &HealthView{result: result}
}

// SetWidth sets the terminal width for rendering.
func (hv *HealthView) SetWidth(w int) {
	hv.width = w
}

// Update handles messages for the health view.
func (hv *HealthView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyEscape {
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

// View renders the health check results.
func (hv *HealthView) View() string {
	s := strings.Builder{}

	s.WriteString(headerStyle.Render("ThreeDoors - Health Check"))
	s.WriteString("\n\n")

	for _, item := range hv.result.Items {
		var statusStr string
		switch item.Status {
		case tasks.HealthOK:
			statusStr = healthOKStyle.Render("[OK]")
		case tasks.HealthFail:
			statusStr = healthFailStyle.Render("[FAIL]")
		case tasks.HealthWarn:
			statusStr = healthWarnStyle.Render("[WARN]")
		}

		fmt.Fprintf(&s, "  %s %s: %s\n", statusStr, item.Name, item.Message)

		if item.Suggestion != "" {
			s.WriteString(healthSuggestionStyle.Render(fmt.Sprintf("  → %s", item.Suggestion)))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")

	// Overall footer
	var overallStyle func(string) string
	switch hv.result.Overall {
	case tasks.HealthOK:
		overallStyle = func(s string) string { return healthOKStyle.Render(s) }
	case tasks.HealthFail:
		overallStyle = func(s string) string { return healthFailStyle.Render(s) }
	case tasks.HealthWarn:
		overallStyle = func(s string) string { return healthWarnStyle.Render(s) }
	default:
		overallStyle = func(s string) string { return s }
	}

	s.WriteString(overallStyle(fmt.Sprintf("Overall: %s", hv.result.Overall)))
	fmt.Fprintf(&s, " | Completed in %s", hv.result.Duration.Round(time.Millisecond))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press Esc to return"))

	return s.String()
}
