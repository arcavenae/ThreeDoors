package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// DisconnectDialog presents a confirmation dialog when disconnecting a source.
// It asks what to do with synced tasks before removing the connection.
type DisconnectDialog struct {
	conn       *connection.Connection
	width      int
	height     int
	form       *huh.Form
	taskAction string // "keep" or "remove"
	confirmed  bool
}

// NewDisconnectDialog creates a disconnect dialog for the given connection.
func NewDisconnectDialog(conn *connection.Connection) *DisconnectDialog {
	d := &DisconnectDialog{
		conn:       conn,
		taskAction: "keep", // safe default
	}
	d.buildForm()
	return d
}

// SetWidth sets the terminal width for rendering.
func (d *DisconnectDialog) SetWidth(w int) {
	d.width = w
}

// SetHeight sets the terminal height for rendering.
func (d *DisconnectDialog) SetHeight(h int) {
	d.height = h
}

func (d *DisconnectDialog) buildForm() {
	d.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What should happen to synced tasks?").
				Options(
					huh.NewOption("Keep tasks locally (remove source attribution)", "keep"),
					huh.NewOption("Remove synced tasks", "remove"),
				).
				Value(&d.taskAction),

			huh.NewConfirm().
				Title("Proceed with disconnection?").
				Affirmative("Disconnect").
				Negative("Cancel").
				Value(&d.confirmed),
		),
	).WithShowHelp(false).WithShowErrors(true)

	d.form.Init()
}

// Init satisfies the Bubbletea model interface.
func (d *DisconnectDialog) Init() tea.Cmd {
	return d.form.Init()
}

// Update handles messages for the disconnect dialog.
func (d *DisconnectDialog) Update(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			return func() tea.Msg { return DisconnectCancelledMsg{} }
		}
	}

	model, cmd := d.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		d.form = f
	}

	if d.form.State == huh.StateCompleted {
		return d.completeDialog()
	}

	if d.form.State == huh.StateAborted {
		return func() tea.Msg { return DisconnectCancelledMsg{} }
	}

	return cmd
}

func (d *DisconnectDialog) completeDialog() tea.Cmd {
	if !d.confirmed {
		return func() tea.Msg { return DisconnectCancelledMsg{} }
	}

	connID := d.conn.ID
	keepTasks := d.taskAction == "keep"
	return func() tea.Msg {
		return DisconnectConfirmedMsg{
			ConnectionID: connID,
			KeepTasks:    keepTasks,
		}
	}
}

// View renders the disconnect dialog.
func (d *DisconnectDialog) View() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Padding(0, 2)

	fmt.Fprintf(&s, "%s\n", headerStyle.Render(
		fmt.Sprintf("Disconnect %q", d.conn.Label),
	))
	fmt.Fprintf(&s, "%s\n\n", descStyle.Render(
		fmt.Sprintf("This will remove the %s connection and delete stored credentials.", d.conn.ProviderName),
	))

	fmt.Fprintf(&s, "%s", d.form.View())

	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" esc:cancel"))

	return s.String()
}
