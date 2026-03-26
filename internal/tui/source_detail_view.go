package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SourceHealthCheckResultMsg delivers health check results to the detail view.
type SourceHealthCheckResultMsg struct {
	ConnectionID string
	Result       connection.HealthCheckResult
	Err          string
}

// SourceDetailView displays detailed information about a single connection.
type SourceDetailView struct {
	conn         *connection.Connection
	connMgr      *connection.ConnectionManager
	width        int
	height       int
	healthResult *connection.HealthCheckResult
	healthError  string
}

// NewSourceDetailView creates a SourceDetailView for the given connection.
func NewSourceDetailView(conn *connection.Connection, mgr *connection.ConnectionManager) *SourceDetailView {
	return &SourceDetailView{
		conn:    conn,
		connMgr: mgr,
	}
}

// SetWidth sets the terminal width for rendering.
func (dv *SourceDetailView) SetWidth(w int) {
	dv.width = w
}

// SetHeight sets the terminal height for rendering.
func (dv *SourceDetailView) SetHeight(h int) {
	dv.height = h
}

// Update handles key events and messages for the source detail view.
func (dv *SourceDetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case SourceHealthCheckResultMsg:
		if msg.ConnectionID == dv.conn.ID {
			if msg.Err != "" {
				dv.healthError = msg.Err
				dv.healthResult = nil
			} else {
				dv.healthResult = &msg.Result
				dv.healthError = ""
			}
		}
		return nil

	case tea.KeyMsg:
		return dv.handleKey(msg)
	}

	return nil
}

func (dv *SourceDetailView) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		return func() tea.Msg { return ShowSourcesMsg{} }
	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return nil
		}
		connID := dv.conn.ID
		switch msg.Runes[0] {
		case 'q':
			return func() tea.Msg { return ShowSourcesMsg{} }
		case 'e':
			return func() tea.Msg { return SourceActionMsg{ConnectionID: connID, Action: "edit"} }
		case 'r':
			return func() tea.Msg { return SourceActionMsg{ConnectionID: connID, Action: "reauth"} }
		case 'p':
			return func() tea.Msg { return SourceActionMsg{ConnectionID: connID, Action: "toggle_pause"} }
		case 'd':
			return func() tea.Msg { return SourceActionMsg{ConnectionID: connID, Action: "disconnect"} }
		case 'l':
			return func() tea.Msg { return SourceActionMsg{ConnectionID: connID, Action: "sync_log"} }
		}
	}
	return nil
}

// View renders the source detail view with panels for metadata and health checks.
func (dv *SourceDetailView) View() string {
	width := dv.width
	if width < 20 {
		width = 20
	}

	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render(dv.conn.Label))

	// Determine layout
	if width >= 80 {
		// Two-column layout: metadata left, health checks right
		leftWidth := (width - 3) / 2 // -3 for gap
		rightWidth := width - leftWidth - 3

		leftPanel := dv.renderMetadataPanel(leftWidth)
		rightPanel := dv.renderHealthPanel(rightWidth)

		joined := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)
		fmt.Fprintf(&s, "%s\n", joined)
	} else {
		// Single column: stack panels
		fmt.Fprintf(&s, "%s\n", dv.renderMetadataPanel(width-2))
		fmt.Fprintf(&s, "%s\n", dv.renderHealthPanel(width-2))
	}

	// Footer with keybinding hints
	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" e:edit  r:reauth  p:pause  d:disconnect  l:sync log  esc:back"))

	return s.String()
}

func (dv *SourceDetailView) renderMetadataPanel(width int) string {
	var content strings.Builder

	conn := dv.conn
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Provider
	fmt.Fprintf(&content, "%s %s\n", labelStyle.Render("Provider:"), conn.ProviderName)

	// State
	indicator, statusText := statusIndicatorAndText(conn.State, conn.LastSync, conn.LastError)
	indicatorStyled := coloredIndicator(indicator, conn.State)
	fmt.Fprintf(&content, "%s %s %s\n", labelStyle.Render("Status:"), indicatorStyled, statusText)

	// Last error (if in error state)
	if conn.LastError != "" && conn.State == connection.StateError {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		fmt.Fprintf(&content, "%s %s\n", labelStyle.Render("Error:"), errStyle.Render(conn.LastError))
	}

	// Last sync
	if !conn.LastSync.IsZero() {
		ago := time.Since(conn.LastSync).Truncate(time.Second)
		fmt.Fprintf(&content, "%s %s ago\n", labelStyle.Render("Last Sync:"), ago)
	} else {
		fmt.Fprintf(&content, "%s never\n", labelStyle.Render("Last Sync:"))
	}

	// Task count
	fmt.Fprintf(&content, "%s %d\n", labelStyle.Render("Tasks:"), conn.TaskCount)

	// Sync mode
	fmt.Fprintf(&content, "%s %s\n", labelStyle.Render("Sync Mode:"), conn.SyncMode)

	// Poll interval
	fmt.Fprintf(&content, "%s %s\n", labelStyle.Render("Poll Rate:"), conn.PollInterval)

	// Settings
	if len(conn.Settings) > 0 {
		fmt.Fprintf(&content, "\n%s\n", statsSectionHeaderStyle.Render("Settings"))
		keys := make([]string, 0, len(conn.Settings))
		for k := range conn.Settings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&content, "  %s %s\n", labelStyle.Render(k+":"), conn.Settings[k])
		}
	}

	return makePanel("Connection Info", content.String(), width,
		lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"})
}

func (dv *SourceDetailView) renderHealthPanel(width int) string {
	var content strings.Builder

	if dv.healthError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		fmt.Fprintf(&content, "%s\n", errStyle.Render("Health check failed: "+dv.healthError))
		return makePanel("Health Checks", content.String(), width,
			lipgloss.AdaptiveColor{Light: "#AA0000", Dark: "#FF5555"})
	}

	if dv.healthResult == nil {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		fmt.Fprintf(&content, "%s\n", dimStyle.Render("No health check results yet."))
		fmt.Fprintf(&content, "%s\n", dimStyle.Render("Press 't' in sources list to test."))
		return makePanel("Health Checks", content.String(), width,
			lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"})
	}

	checks := []struct {
		name   string
		passed bool
	}{
		{"API Reachable", dv.healthResult.APIReachable},
		{"Token Valid", dv.healthResult.TokenValid},
		{"Rate Limit OK", dv.healthResult.RateLimitOK},
	}

	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red

	for _, c := range checks {
		if c.passed {
			fmt.Fprintf(&content, " %s %s\n", passStyle.Render("✓"), c.name)
		} else {
			fmt.Fprintf(&content, " %s %s\n", failStyle.Render("✗"), c.name)
		}
	}

	borderColor := lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#55FF55"}
	if !dv.healthResult.Healthy() {
		borderColor = lipgloss.AdaptiveColor{Light: "#AA0000", Dark: "#FF5555"}
	}

	return makePanel("Health Checks", content.String(), width, borderColor)
}
