package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ShowSourcesMsg is sent to open the sources dashboard view.
type ShowSourcesMsg struct{}

// ShowConnectWizardMsg is sent when the user presses 'a' to add a new connection.
type ShowConnectWizardMsg struct{}

// ShowSourceDetailMsg is sent when the user presses Enter on a connection.
type ShowSourceDetailMsg struct {
	ConnectionID string
}

// SourceActionMsg is sent for connection actions (pause, resync, test, disconnect).
type SourceActionMsg struct {
	ConnectionID string
	Action       string // "toggle_pause", "resync", "test", "disconnect"
}

// SourcesView displays the sources dashboard with all connections.
type SourcesView struct {
	connMgr       *connection.ConnectionManager
	width         int
	height        int
	selectedIndex int
}

// NewSourcesView creates a new SourcesView.
func NewSourcesView(connMgr *connection.ConnectionManager) *SourcesView {
	return &SourcesView{
		connMgr: connMgr,
	}
}

// SetWidth sets the terminal width for rendering.
func (sv *SourcesView) SetWidth(w int) {
	sv.width = w
}

// SetHeight sets the terminal height for rendering.
func (sv *SourcesView) SetHeight(h int) {
	sv.height = h
}

// Update handles key events for the sources view.
func (sv *SourcesView) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	conns := sv.connMgr.List()

	switch keyMsg.Type {
	case tea.KeyEsc:
		return func() tea.Msg { return ReturnToDoorsMsg{} }

	case tea.KeyEnter:
		if len(conns) == 0 {
			return nil
		}
		selected := conns[sv.selectedIndex]
		return func() tea.Msg { return ShowSourceDetailMsg{ConnectionID: selected.ID} }

	case tea.KeyRunes:
		if len(keyMsg.Runes) == 0 {
			return nil
		}
		switch keyMsg.Runes[0] {
		case 'q':
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case 'a':
			return func() tea.Msg { return ShowConnectWizardMsg{} }
		case 'j':
			if sv.selectedIndex < len(conns)-1 {
				sv.selectedIndex++
			}
		case 'k':
			if sv.selectedIndex > 0 {
				sv.selectedIndex--
			}
		case 'd':
			return sv.connectionAction(conns, "disconnect")
		case 'p':
			return sv.connectionAction(conns, "toggle_pause")
		case 'r':
			return sv.connectionAction(conns, "resync")
		case 't':
			return sv.connectionAction(conns, "test")
		}

	case tea.KeyUp:
		if sv.selectedIndex > 0 {
			sv.selectedIndex--
		}
	case tea.KeyDown:
		if sv.selectedIndex < len(conns)-1 {
			sv.selectedIndex++
		}
	}

	return nil
}

// connectionAction returns a command for the selected connection, or nil if empty.
func (sv *SourcesView) connectionAction(conns []*connection.Connection, action string) tea.Cmd {
	if len(conns) == 0 {
		return nil
	}
	selected := conns[sv.selectedIndex]
	return func() tea.Msg {
		return SourceActionMsg{ConnectionID: selected.ID, Action: action}
	}
}

// View renders the sources dashboard.
func (sv *SourcesView) View() string {
	conns := sv.connMgr.List()

	width := sv.width
	if width < 20 {
		width = 20
	}

	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("Sources"))

	if len(conns) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Padding(1, 2)
		fmt.Fprintf(&s, "%s\n", emptyStyle.Render("No connections configured.\nPress 'a' to add a connection."))
		return s.String()
	}

	// Clamp selectedIndex
	if sv.selectedIndex >= len(conns) {
		sv.selectedIndex = len(conns) - 1
	}

	// Render each connection row
	for i, conn := range conns {
		indicator, statusText := statusIndicatorAndText(conn.State, conn.LastSync, conn.LastError)

		// Color the indicator
		indicatorStyled := coloredIndicator(indicator, conn.State)

		// Task count display
		taskCountStr := fmt.Sprintf("%d tasks", conn.TaskCount)
		if conn.State == connection.StateAuthExpired {
			taskCountStr = "0 tasks"
		}

		// Build the row
		label := conn.Label
		if width < 60 {
			// Compact: indicator label status
			row := fmt.Sprintf(" %s %s  %s  %s", indicatorStyled, label, statusText, taskCountStr)
			if i == sv.selectedIndex {
				row = lipgloss.NewStyle().
					Background(lipgloss.Color("236")).
					Bold(true).
					Width(width - 2).
					Render(row)
			}
			fmt.Fprintf(&s, "%s\n", row)
		} else {
			// Standard: indicator label — status — task count
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
			countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

			row := fmt.Sprintf(" %s %-20s %s  %s",
				indicatorStyled,
				label,
				statusStyle.Render(statusText),
				countStyle.Render(taskCountStr),
			)

			if i == sv.selectedIndex {
				row = lipgloss.NewStyle().
					Background(lipgloss.Color("236")).
					Bold(true).
					Width(width - 2).
					Render(row)
			}
			fmt.Fprintf(&s, "%s\n", row)
		}
	}

	// Footer with keybinding hints
	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" a:add  d:disconnect  p:pause  r:resync  t:test  enter:detail  esc:back"))

	return s.String()
}

// statusIndicatorAndText returns the status indicator character and descriptive text
// for a connection state.
func statusIndicatorAndText(state connection.ConnectionState, lastSync time.Time, lastError string) (string, string) {
	switch state {
	case connection.StateConnected:
		if lastSync.IsZero() {
			return "●", "Synced"
		}
		ago := time.Since(lastSync).Truncate(time.Minute)
		if ago < time.Minute {
			return "●", "Synced just now"
		}
		return "●", fmt.Sprintf("Synced %dm ago", int(ago.Minutes()))

	case connection.StatePaused:
		return "○", "Paused"

	case connection.StateAuthExpired:
		return "⚠", "Auth expired"

	case connection.StateError:
		if lastError != "" {
			return "●", fmt.Sprintf("Error: %s", lastError)
		}
		return "●", "Error"

	case connection.StateDisconnected:
		return "○", "Disconnected"

	case connection.StateConnecting:
		return "…", "Connecting"

	case connection.StateSyncing:
		return "↻", "Syncing"

	default:
		return "?", "Unknown"
	}
}

// coloredIndicator returns the indicator with appropriate color styling.
func coloredIndicator(indicator string, state connection.ConnectionState) string {
	var color lipgloss.Color
	switch state {
	case connection.StateConnected, connection.StateSyncing:
		color = lipgloss.Color("42") // green
	case connection.StatePaused, connection.StateDisconnected:
		color = lipgloss.Color("243") // gray
	case connection.StateAuthExpired:
		color = lipgloss.Color("220") // yellow
	case connection.StateError:
		color = lipgloss.Color("196") // red
	case connection.StateConnecting:
		color = lipgloss.Color("243") // gray
	default:
		color = lipgloss.Color("243")
	}
	return lipgloss.NewStyle().Foreground(color).Render(indicator)
}
