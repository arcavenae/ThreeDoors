package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the root Bubbletea model for the application.
type Model struct {
	quitting bool
}

// NewModel creates and returns a new Model instance.
func NewModel() Model {
	return Model{}
}

// Init returns nil (no initial command needed).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages: 'q' and ctrl+c trigger tea.Quit.
// All other keys are ignored.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyRunes:
			if string(msg.Runes) == "q" {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the header and quit hint as plain text.
func (m Model) View() string {
	return "ThreeDoors - Technical Demo\n\nPress q to quit\n"
}
