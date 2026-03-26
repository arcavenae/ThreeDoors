package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusTarget = 3
	focusMax    = 5
)

// EnergyAll is a sentinel used when the user cycles past Low to show all tasks.
const EnergyAll = "all"

// SelectCompleteMsg is sent when the user confirms their focus selection.
type SelectCompleteMsg struct {
	FocusTasks     []*core.Task
	EnergyLevel    core.EnergyLevel
	EnergyOverride bool
}

// SelectCancelMsg is sent when the user cancels (Esc) back to review step.
type SelectCancelMsg struct{}

// SelectView displays a scrollable task list for focus selection filtered by energy level.
type SelectView struct {
	allTasks       []*core.Task
	filtered       []*core.Task
	selected       map[string]bool // task ID -> selected
	cursor         int
	energy         string // core.EnergyLevel or EnergyAll
	energyOverride bool
	width          int
	height         int
	scrollOffset   int
	showHelp       bool
	confirmEmpty   bool // awaiting Y/N for zero-selection
	startTime      time.Time
}

// NewSelectView creates a SelectView for the given task pool and energy level.
func NewSelectView(pool []*core.Task, energy core.EnergyLevel) *SelectView {
	sv := &SelectView{
		allTasks:  pool,
		selected:  make(map[string]bool),
		energy:    string(energy),
		startTime: time.Now().UTC(),
	}
	sv.refilter()
	return sv
}

// SetWidth sets the terminal width for responsive rendering.
func (sv *SelectView) SetWidth(w int) {
	sv.width = w
}

// SetHeight sets the terminal height for scroll calculations.
func (sv *SelectView) SetHeight(h int) {
	sv.height = h
}

// Init returns nil — no initial command needed.
func (sv *SelectView) Init() tea.Cmd {
	return nil
}

// Update handles key input.
func (sv *SelectView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if sv.confirmEmpty {
			return sv.handleConfirmInput(msg)
		}
		if sv.showHelp {
			return sv.handleHelpInput(msg)
		}
		return sv.handleKeyInput(msg)
	}
	return nil
}

func (sv *SelectView) handleConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch strings.ToLower(msg.String()) {
	case "y":
		sv.confirmEmpty = false
		return sv.completeCmd()
	case "n", "esc":
		sv.confirmEmpty = false
	}
	return nil
}

func (sv *SelectView) handleHelpInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "?", "esc":
		sv.showHelp = false
	}
	return nil
}

func (sv *SelectView) handleKeyInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyUp:
		sv.moveCursor(-1)
	case tea.KeyDown:
		sv.moveCursor(1)
	case tea.KeyEscape:
		return func() tea.Msg { return SelectCancelMsg{} }
	case tea.KeyEnter:
		return sv.handleEnter()
	case tea.KeySpace:
		sv.toggleSelection()
	case tea.KeyRunes:
		return sv.handleRune(msg.String())
	}
	return nil
}

func (sv *SelectView) handleRune(key string) tea.Cmd {
	switch strings.ToLower(key) {
	case "k":
		sv.moveCursor(-1)
	case "j":
		sv.moveCursor(1)
	case " ":
		sv.toggleSelection()
	case "e":
		sv.cycleEnergy()
	case "?":
		sv.showHelp = true
	}
	return nil
}

func (sv *SelectView) moveCursor(delta int) {
	if len(sv.filtered) == 0 {
		return
	}
	sv.cursor += delta
	if sv.cursor < 0 {
		sv.cursor = 0
	}
	if sv.cursor >= len(sv.filtered) {
		sv.cursor = len(sv.filtered) - 1
	}
	sv.adjustScroll()
}

func (sv *SelectView) adjustScroll() {
	visibleLines := sv.visibleLines()
	if visibleLines <= 0 {
		return
	}
	if sv.cursor < sv.scrollOffset {
		sv.scrollOffset = sv.cursor
	}
	if sv.cursor >= sv.scrollOffset+visibleLines {
		sv.scrollOffset = sv.cursor - visibleLines + 1
	}
}

func (sv *SelectView) visibleLines() int {
	// Reserve lines for header (4), footer (4), border (4)
	available := sv.height - 12
	if available < 3 {
		return 3
	}
	return available
}

func (sv *SelectView) toggleSelection() {
	if len(sv.filtered) == 0 {
		return
	}
	task := sv.filtered[sv.cursor]
	if sv.selected[task.ID] {
		// Deselect
		delete(sv.selected, task.ID)
		task.Text = core.RemoveFocusTagFromText(task.Text)
		task.UpdatedAt = time.Now().UTC()
	} else {
		// Check max
		if sv.selectedCount() >= focusMax {
			return
		}
		sv.selected[task.ID] = true
		if !core.HasFocusTag(task) {
			task.Text = task.Text + " +focus"
			task.UpdatedAt = time.Now().UTC()
		}
	}
}

func (sv *SelectView) selectedCount() int {
	return len(sv.selected)
}

func (sv *SelectView) cycleEnergy() {
	switch sv.energy {
	case string(core.EnergyHigh):
		sv.energy = string(core.EnergyMedium)
	case string(core.EnergyMedium):
		sv.energy = string(core.EnergyLow)
	case string(core.EnergyLow):
		sv.energy = EnergyAll
	default:
		sv.energy = string(core.EnergyHigh)
	}
	sv.energyOverride = true
	sv.refilter()
}

func (sv *SelectView) refilter() {
	sv.filtered = nil
	for _, t := range sv.allTasks {
		if sv.energy == EnergyAll || core.MatchesEnergy(t, core.EnergyLevel(sv.energy)) {
			sv.filtered = append(sv.filtered, t)
		}
	}
	// Sort: selected first, then by energy match score
	sv.sortFiltered()
	// Reset cursor if out of range
	if sv.cursor >= len(sv.filtered) {
		sv.cursor = 0
	}
	sv.scrollOffset = 0
}

func (sv *SelectView) sortFiltered() {
	// Simple stable sort: selected tasks first, then by effort match quality
	// Using insertion sort for stability
	for i := 1; i < len(sv.filtered); i++ {
		for j := i; j > 0 && sv.taskSortLess(sv.filtered[j], sv.filtered[j-1]); j-- {
			sv.filtered[j], sv.filtered[j-1] = sv.filtered[j-1], sv.filtered[j]
		}
	}
}

func (sv *SelectView) taskSortLess(a, b *core.Task) bool {
	aSelected := sv.selected[a.ID]
	bSelected := sv.selected[b.ID]
	if aSelected != bSelected {
		return aSelected
	}
	return sv.energyMatchScore(a) > sv.energyMatchScore(b)
}

func (sv *SelectView) energyMatchScore(t *core.Task) int {
	if sv.energy == EnergyAll {
		return 0
	}
	if core.MatchesEnergy(t, core.EnergyLevel(sv.energy)) {
		if t.Effort != "" {
			return 2 // explicit match
		}
		return 1 // no effort tag (matches anything)
	}
	return 0
}

func (sv *SelectView) handleEnter() tea.Cmd {
	if sv.selectedCount() == 0 {
		sv.confirmEmpty = true
		return nil
	}
	return sv.completeCmd()
}

func (sv *SelectView) completeCmd() tea.Cmd {
	var focusTasks []*core.Task
	for _, t := range sv.allTasks {
		if sv.selected[t.ID] {
			focusTasks = append(focusTasks, t)
		}
	}
	energy := sv.energy
	override := sv.energyOverride
	return func() tea.Msg {
		return SelectCompleteMsg{
			FocusTasks:     focusTasks,
			EnergyLevel:    core.EnergyLevel(energy),
			EnergyOverride: override,
		}
	}
}

// View renders the select step.
func (sv *SelectView) View() string {
	if sv.showHelp {
		return sv.renderHelp()
	}

	var s strings.Builder

	sv.renderHeader(&s)
	sv.renderEnergyBar(&s)
	sv.renderTaskList(&s)
	sv.renderStatusMessages(&s)
	sv.renderActionKeys(&s)
	sv.renderFooter(&s)

	w := sv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

func (sv *SelectView) renderHeader(s *strings.Builder) {
	s.WriteString(selectHeaderStyle.Render("Select Focus Tasks"))
	fmt.Fprintf(s, "  %s\n", sv.selectionCounter())
	s.WriteString("\n")
}

func (sv *SelectView) selectionCounter() string {
	count := sv.selectedCount()
	return selectCountStyle.Render(fmt.Sprintf("Focus: %d/%d selected", count, focusTarget))
}

func (sv *SelectView) renderEnergyBar(s *strings.Builder) {
	label := sv.energyLabel()
	fmt.Fprintf(s, "Energy: %s\n\n", selectEnergyStyle.Render(label))
}

func (sv *SelectView) energyLabel() string {
	switch sv.energy {
	case string(core.EnergyHigh):
		return "High (morning)"
	case string(core.EnergyMedium):
		return "Medium (afternoon)"
	case string(core.EnergyLow):
		return "Low (evening)"
	case EnergyAll:
		return "All (no filter)"
	default:
		return string(sv.energy)
	}
}

func (sv *SelectView) renderTaskList(s *strings.Builder) {
	if len(sv.filtered) == 0 {
		s.WriteString("No matching tasks\n")
		s.WriteString(helpStyle.Render("Press E to switch energy filter"))
		s.WriteString("\n")
		return
	}

	visible := sv.visibleLines()
	end := sv.scrollOffset + visible
	if end > len(sv.filtered) {
		end = len(sv.filtered)
	}

	for i := sv.scrollOffset; i < end; i++ {
		t := sv.filtered[i]
		line := sv.renderTaskLine(t, i)
		if i == sv.cursor {
			line = searchSelectedStyle.Render(line)
		} else {
			line = searchResultStyle.Render(line)
		}
		s.WriteString(line)
		s.WriteString("\n")
	}
}

func (sv *SelectView) renderTaskLine(t *core.Task, _ int) string {
	var parts []string

	// Checkbox
	if sv.selected[t.ID] {
		parts = append(parts, "[x]")
	} else {
		parts = append(parts, "[ ]")
	}

	// Task text (strip +focus tag for display)
	displayText := core.RemoveFocusTagFromText(t.Text)
	parts = append(parts, displayText)

	// Tags
	var tags []string
	if t.Effort != "" {
		tags = append(tags, string(t.Effort))
	}
	if t.Type != "" {
		tags = append(tags, string(t.Type))
	}
	if t.Status != "" {
		tags = append(tags, string(t.Status))
	}
	if len(tags) > 0 {
		parts = append(parts, "("+strings.Join(tags, ", ")+")")
	}

	return strings.Join(parts, " ")
}

func (sv *SelectView) renderStatusMessages(s *strings.Builder) {
	count := sv.selectedCount()
	if sv.confirmEmpty {
		s.WriteString("\n")
		s.WriteString(selectWarningStyle.Render("No focus tasks selected. Continue anyway? [Y/N]"))
		s.WriteString("\n")
		return
	}
	if count == focusTarget {
		s.WriteString("\n")
		s.WriteString(selectTargetStyle.Render("Target reached!"))
		s.WriteString("\n")
	} else if count >= focusMax {
		s.WriteString("\n")
		s.WriteString(selectWarningStyle.Render("Maximum 5 focus tasks"))
		s.WriteString("\n")
	}
}

func (sv *SelectView) renderActionKeys(s *strings.Builder) {
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("[Space]Toggle  [E]nergy  [Enter]Confirm  [Esc]Back  [?]Help"))
	s.WriteString("\n")
}

func (sv *SelectView) renderFooter(s *strings.Builder) {
	s.WriteString(helpStyle.Render("Step 2/3 — Select Focus"))
}

func (sv *SelectView) renderHelp() string {
	var s strings.Builder
	s.WriteString(selectHeaderStyle.Render("Select Help"))
	s.WriteString("\n\n")
	s.WriteString("  Space — Toggle focus selection on current task\n")
	s.WriteString("  E — Cycle energy filter: High → Medium → Low → All\n")
	s.WriteString("  Up/Down or k/j — Navigate task list\n")
	s.WriteString("  Enter — Confirm selection and advance to Step 3\n")
	s.WriteString("  Esc — Return to Step 1 (review)\n")
	s.WriteString("  ? — Toggle this help\n")
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press ? or Esc to close"))

	w := sv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

// Styles for the select view
var (
	selectHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	selectCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	selectEnergyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Bold(true)

	selectTargetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82"))

	selectWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)
)
