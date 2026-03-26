package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ReauthDialog presents a masked token input for re-authenticating
// a connection whose token has expired.
type ReauthDialog struct {
	conn      *connection.Connection
	tokenHelp string
	width     int
	height    int
	form      *huh.Form
	newToken  string
}

// NewReauthDialog creates a re-authentication dialog for the given connection.
// tokenHelp provides guidance text (e.g., "Settings → Integrations → API token").
func NewReauthDialog(conn *connection.Connection, tokenHelp string) *ReauthDialog {
	d := &ReauthDialog{
		conn:      conn,
		tokenHelp: tokenHelp,
	}
	d.buildForm()
	return d
}

// SetWidth sets the terminal width for rendering.
func (d *ReauthDialog) SetWidth(w int) {
	d.width = w
}

// SetHeight sets the terminal height for rendering.
func (d *ReauthDialog) SetHeight(h int) {
	d.height = h
}

func (d *ReauthDialog) buildForm() {
	input := huh.NewInput().
		Title("New API Token").
		EchoMode(huh.EchoModePassword).
		Value(&d.newToken).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("token must not be empty")
			}
			return nil
		})

	if d.tokenHelp != "" {
		input = input.Description(d.tokenHelp)
	}

	d.form = huh.NewForm(
		huh.NewGroup(input),
	).WithShowHelp(false).WithShowErrors(true)

	d.form.Init()
}

// Init satisfies the Bubbletea model interface.
func (d *ReauthDialog) Init() tea.Cmd {
	return d.form.Init()
}

// Update handles messages for the re-auth dialog.
func (d *ReauthDialog) Update(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			return func() tea.Msg { return ReauthCancelledMsg{} }
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
		return func() tea.Msg { return ReauthCancelledMsg{} }
	}

	return cmd
}

func (d *ReauthDialog) completeDialog() tea.Cmd {
	connID := d.conn.ID
	token := strings.TrimSpace(d.newToken)
	return func() tea.Msg {
		return ReauthCompleteMsg{
			ConnectionID: connID,
			NewToken:     token,
		}
	}
}

// View renders the re-auth dialog.
func (d *ReauthDialog) View() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Padding(0, 2)

	fmt.Fprintf(&s, "%s\n", headerStyle.Render(
		fmt.Sprintf("Re-authenticate %q", d.conn.Label),
	))
	fmt.Fprintf(&s, "%s\n\n", descStyle.Render(
		fmt.Sprintf("Connection %q (%s) requires a new API token.", d.conn.Label, d.conn.ProviderName),
	))

	fmt.Fprintf(&s, "%s", d.form.View())

	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" esc:cancel"))

	return s.String()
}
