package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LinkViewStep tracks the current step in the link creation flow.
type LinkViewStep int

const (
	linkStepSelectTarget LinkViewStep = iota
	linkStepSelectRelationship
)

var relationshipOptions = []string{"related", "blocks", "depends-on"}

// LinkView handles creating cross-references between tasks.
type LinkView struct {
	sourceTask    *tasks.Task
	pool          *tasks.TaskPool
	enrichDB      *enrichment.DB
	step          LinkViewStep
	searchInput   string
	results       []*tasks.Task
	selectedIndex int
	relIndex      int
	width         int
}

// NewLinkView creates a new link creation view.
func NewLinkView(source *tasks.Task, pool *tasks.TaskPool, db *enrichment.DB) *LinkView {
	return &LinkView{
		sourceTask:    source,
		pool:          pool,
		enrichDB:      db,
		step:          linkStepSelectTarget,
		selectedIndex: -1,
	}
}

// SetWidth sets the terminal width.
func (lv *LinkView) SetWidth(w int) {
	lv.width = w
}

// Update handles key input for the link view.
func (lv *LinkView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch lv.step {
		case linkStepSelectTarget:
			return lv.handleTargetSelection(msg)
		case linkStepSelectRelationship:
			return lv.handleRelationshipSelection(msg)
		}
	}
	return nil
}

func (lv *LinkView) handleTargetSelection(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		return func() tea.Msg { return LinkCancelledMsg{} }
	case tea.KeyEnter:
		if lv.selectedIndex >= 0 && lv.selectedIndex < len(lv.results) {
			lv.step = linkStepSelectRelationship
			lv.relIndex = 0
		}
		return nil
	case tea.KeyUp:
		if lv.selectedIndex > 0 {
			lv.selectedIndex--
		}
		return nil
	case tea.KeyDown:
		if lv.selectedIndex < len(lv.results)-1 {
			lv.selectedIndex++
		}
		return nil
	case tea.KeyBackspace:
		if len(lv.searchInput) > 0 {
			lv.searchInput = lv.searchInput[:len(lv.searchInput)-1]
			lv.filterTasks()
		}
		return nil
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			lv.searchInput += string(msg.Runes)
			lv.filterTasks()
		}
		return nil
	}
}

func (lv *LinkView) handleRelationshipSelection(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		lv.step = linkStepSelectTarget
		return nil
	case tea.KeyEnter:
		return lv.createLink()
	case tea.KeyUp:
		if lv.relIndex > 0 {
			lv.relIndex--
		}
		return nil
	case tea.KeyDown:
		if lv.relIndex < len(relationshipOptions)-1 {
			lv.relIndex++
		}
		return nil
	default:
		return nil
	}
}

func (lv *LinkView) filterTasks() {
	lv.results = nil
	lv.selectedIndex = -1
	if lv.searchInput == "" {
		// Show all tasks except the source
		allTasks := lv.pool.GetAllTasks()
		for _, t := range allTasks {
			if t.ID != lv.sourceTask.ID {
				lv.results = append(lv.results, t)
			}
		}
	} else {
		lowerQuery := strings.ToLower(lv.searchInput)
		allTasks := lv.pool.GetAllTasks()
		for _, t := range allTasks {
			if t.ID != lv.sourceTask.ID && strings.Contains(strings.ToLower(t.Text), lowerQuery) {
				lv.results = append(lv.results, t)
			}
		}
	}
	if len(lv.results) > 0 {
		lv.selectedIndex = 0
	}
}

func (lv *LinkView) createLink() tea.Cmd {
	if lv.selectedIndex < 0 || lv.selectedIndex >= len(lv.results) {
		return nil
	}
	target := lv.results[lv.selectedIndex]
	rel := relationshipOptions[lv.relIndex]

	ref := &enrichment.CrossReference{
		SourceTaskID: lv.sourceTask.ID,
		TargetTaskID: target.ID,
		SourceSystem: "user",
		Relationship: rel,
	}
	if err := lv.enrichDB.AddCrossReference(ref); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Failed to create link: %v", err)}
		}
	}

	sourceID := lv.sourceTask.ID
	targetID := target.ID
	return func() tea.Msg {
		return LinkCreatedMsg{
			SourceTaskID: sourceID,
			TargetTaskID: targetID,
			Relationship: rel,
		}
	}
}

// View renders the link creation view.
func (lv *LinkView) View() string {
	s := strings.Builder{}

	w := lv.width - 6
	if w < 40 {
		w = 40
	}

	s.WriteString(headerStyle.Render("LINK TASK"))
	s.WriteString("\n\n")

	sourceStyle := lipgloss.NewStyle().Bold(true)
	fmt.Fprintf(&s, "Source: %s\n\n", sourceStyle.Render(lv.sourceTask.Text))

	switch lv.step {
	case linkStepSelectTarget:
		s.WriteString("Search for target task:\n")
		fmt.Fprintf(&s, "> %s_\n\n", lv.searchInput)

		if len(lv.results) == 0 && lv.searchInput != "" {
			s.WriteString(helpStyle.Render("No matching tasks found"))
			s.WriteString("\n")
		} else {
			maxShow := 10
			if len(lv.results) < maxShow {
				maxShow = len(lv.results)
			}
			for i := 0; i < maxShow; i++ {
				t := lv.results[i]
				line := fmt.Sprintf("  %s", t.Text)
				if i == lv.selectedIndex {
					line = searchSelectedStyle.Render(line)
				} else {
					line = searchResultStyle.Render(line)
				}
				s.WriteString(line)
				s.WriteString("\n")
			}
			if len(lv.results) > maxShow {
				fmt.Fprintf(&s, "  ... and %d more\n", len(lv.results)-maxShow)
			}
		}
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Type to search | ↑/↓ navigate | Enter select | Esc cancel"))

	case linkStepSelectRelationship:
		target := lv.results[lv.selectedIndex]
		targetStyle := lipgloss.NewStyle().Bold(true)
		fmt.Fprintf(&s, "Target: %s\n\n", targetStyle.Render(target.Text))
		s.WriteString("Select relationship:\n\n")
		for i, rel := range relationshipOptions {
			if i == lv.relIndex {
				fmt.Fprintf(&s, "  > %s\n", rel)
			} else {
				fmt.Fprintf(&s, "    %s\n", rel)
			}
		}
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("↑/↓ select | Enter confirm | Esc back"))
	}

	return detailBorder.Width(w).Render(s.String())
}
