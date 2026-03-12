package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DisconnectOption represents the task preservation choice.
type DisconnectOption int

const (
	DisconnectKeepTasks DisconnectOption = iota
	DisconnectRemoveTasks
)

var disconnectOptionLabels = [2]string{
	"Keep tasks locally",
	"Remove synced tasks",
}

// DisconnectDialog displays a confirmation dialog for disconnecting a data source.
type DisconnectDialog struct {
	connectionID    string
	connectionLabel string
	cursor          int
	width           int
}

// NewDisconnectDialog creates a disconnect dialog for the given connection.
func NewDisconnectDialog(connectionID, connectionLabel string) *DisconnectDialog {
	return &DisconnectDialog{
		connectionID:    connectionID,
		connectionLabel: connectionLabel,
	}
}

// SetWidth sets the terminal width for rendering.
func (d *DisconnectDialog) SetWidth(w int) {
	d.width = w
}

// Update handles key input for the disconnect dialog.
func (d *DisconnectDialog) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "j":
		if d.cursor < 1 {
			d.cursor++
		}
	case "esc":
		return func() tea.Msg { return DisconnectCancelledMsg{} }
	case "enter":
		keepTasks := d.cursor == int(DisconnectKeepTasks)
		connID := d.connectionID
		return func() tea.Msg {
			return DisconnectConfirmedMsg{
				ConnectionID: connID,
				KeepTasks:    keepTasks,
			}
		}
	}
	return nil
}

// View renders the disconnect dialog.
func (d *DisconnectDialog) View() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", headerStyle.Render("DISCONNECT SOURCE"))

	label := d.connectionLabel
	if len(label) > 50 {
		label = label[:47] + "..."
	}
	fmt.Fprintf(&s, "  Disconnect %q?\n\n", label)
	fmt.Fprintf(&s, "  What should happen to synced tasks?\n\n")

	for i, optLabel := range disconnectOptionLabels {
		prefix := "  "
		if i == d.cursor {
			prefix = "> "
		}
		fmt.Fprintf(&s, "  %s%s\n", prefix, optLabel)
	}

	fmt.Fprintf(&s, "\n  [Enter] Disconnect  [Esc] Cancel\n")

	return s.String()
}
