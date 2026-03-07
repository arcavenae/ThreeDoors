package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/mcp"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// proposalGroup groups proposals by their target task.
type proposalGroup struct {
	TaskID    string
	Task      *core.Task
	Proposals []*mcp.Proposal
}

// ProposalsView renders the proposal review interface with split pane layout.
type ProposalsView struct {
	store         *mcp.ProposalStore
	pool          *core.TaskPool
	provider      core.TaskProvider
	groups        []proposalGroup
	flatIndex     []flatEntry
	selectedIndex int
	previewMode   bool
	width         int
	height        int
}

// flatEntry maps a flat index to a group and proposal within that group.
type flatEntry struct {
	groupIdx    int
	proposalIdx int
}

// NewProposalsView creates a new ProposalsView.
func NewProposalsView(store *mcp.ProposalStore, pool *core.TaskPool, provider core.TaskProvider) *ProposalsView {
	pv := &ProposalsView{
		store:    store,
		pool:     pool,
		provider: provider,
	}
	pv.refreshProposals()
	return pv
}

// SetWidth sets the terminal width for rendering.
func (pv *ProposalsView) SetWidth(w int) {
	pv.width = w
}

// SetHeight sets the terminal height for rendering.
func (pv *ProposalsView) SetHeight(h int) {
	pv.height = h
}

// refreshProposals reloads pending proposals from the store and groups them by task.
func (pv *ProposalsView) refreshProposals() {
	pending := pv.store.List(mcp.ProposalFilter{Status: mcp.ProposalPending})

	// Group by task ID
	groupMap := make(map[string]*proposalGroup)
	for _, p := range pending {
		g, ok := groupMap[p.TaskID]
		if !ok {
			task := pv.pool.GetTask(p.TaskID)
			g = &proposalGroup{
				TaskID: p.TaskID,
				Task:   task,
			}
			groupMap[p.TaskID] = g
		}
		g.Proposals = append(g.Proposals, p)
	}

	// Sort groups by task ID for stable ordering
	pv.groups = make([]proposalGroup, 0, len(groupMap))
	for _, g := range groupMap {
		// Sort proposals within group by creation time
		sort.Slice(g.Proposals, func(i, j int) bool {
			return g.Proposals[i].CreatedAt.Before(g.Proposals[j].CreatedAt)
		})
		pv.groups = append(pv.groups, *g)
	}
	sort.Slice(pv.groups, func(i, j int) bool {
		return pv.groups[i].TaskID < pv.groups[j].TaskID
	})

	// Build flat index for navigation
	pv.flatIndex = nil
	for gi, g := range pv.groups {
		for pi := range g.Proposals {
			pv.flatIndex = append(pv.flatIndex, flatEntry{groupIdx: gi, proposalIdx: pi})
		}
	}

	// Clamp selection
	if pv.selectedIndex >= len(pv.flatIndex) {
		pv.selectedIndex = len(pv.flatIndex) - 1
	}
	if pv.selectedIndex < 0 {
		pv.selectedIndex = 0
	}
}

// currentProposal returns the currently selected proposal, or nil if none.
func (pv *ProposalsView) currentProposal() *mcp.Proposal {
	if len(pv.flatIndex) == 0 {
		return nil
	}
	entry := pv.flatIndex[pv.selectedIndex]
	return pv.groups[entry.groupIdx].Proposals[entry.proposalIdx]
}

// currentTask returns the task associated with the currently selected proposal.
func (pv *ProposalsView) currentTask() *core.Task {
	if len(pv.flatIndex) == 0 {
		return nil
	}
	entry := pv.flatIndex[pv.selectedIndex]
	return pv.groups[entry.groupIdx].Task
}

// isStale checks whether a proposal's base version differs from the task's current UpdatedAt.
func isStale(p *mcp.Proposal, task *core.Task) bool {
	if task == nil {
		return true
	}
	return !p.BaseVersion.Equal(task.UpdatedAt)
}

// Update handles key presses for the proposal review view.
func (pv *ProposalsView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if pv.previewMode {
			switch msg.String() {
			case "esc", "p":
				pv.previewMode = false
			case "enter":
				pv.previewMode = false
				return pv.approveSelected()
			}
			return nil
		}

		switch msg.String() {
		case "j", "down":
			if pv.selectedIndex < len(pv.flatIndex)-1 {
				pv.selectedIndex++
			}
		case "k", "up":
			if pv.selectedIndex > 0 {
				pv.selectedIndex--
			}
		case "enter":
			return pv.approveSelected()
		case "backspace", "delete":
			return pv.rejectSelected()
		case "tab":
			// Skip — move to next
			if pv.selectedIndex < len(pv.flatIndex)-1 {
				pv.selectedIndex++
			}
		case "ctrl+a":
			return pv.approveAllVisible()
		case "p":
			pv.previewMode = true
		case "esc", "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

// approveSelected approves the currently selected proposal and applies it.
func (pv *ProposalsView) approveSelected() tea.Cmd {
	proposal := pv.currentProposal()
	if proposal == nil {
		return nil
	}
	task := pv.currentTask()
	if task == nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Task %s not found", proposal.TaskID)}
		}
	}

	if isStale(proposal, task) {
		return func() tea.Msg {
			return FlashMsg{Text: "Proposal is stale — task has changed since this suggestion was made"}
		}
	}

	if err := applyProposal(proposal, task); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Apply failed: %v", err)}
		}
	}

	now := time.Now().UTC()
	if err := pv.store.UpdateStatus(proposal.ID, mcp.ProposalApproved, now); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Status update failed: %v", err)}
		}
	}

	if err := pv.provider.SaveTask(task); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Save task failed: %v", err)}
		}
	}

	pid := proposal.ID
	tid := proposal.TaskID
	pv.refreshProposals()
	return func() tea.Msg {
		return ProposalApprovedMsg{ProposalID: pid, TaskID: tid}
	}
}

// rejectSelected rejects the currently selected proposal.
func (pv *ProposalsView) rejectSelected() tea.Cmd {
	proposal := pv.currentProposal()
	if proposal == nil {
		return nil
	}

	now := time.Now().UTC()
	if err := pv.store.UpdateStatus(proposal.ID, mcp.ProposalRejected, now); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Reject failed: %v", err)}
		}
	}

	pid := proposal.ID
	pv.refreshProposals()
	return func() tea.Msg {
		return ProposalRejectedMsg{ProposalID: pid}
	}
}

// approveAllVisible approves all non-stale pending proposals.
func (pv *ProposalsView) approveAllVisible() tea.Cmd {
	approved := 0
	now := time.Now().UTC()

	// Snapshot each task's UpdatedAt before applying any proposals,
	// so applying one proposal doesn't make others for the same task stale.
	taskVersions := make(map[string]time.Time)
	for _, g := range pv.groups {
		if g.Task != nil {
			taskVersions[g.TaskID] = g.Task.UpdatedAt
		}
	}

	for _, g := range pv.groups {
		if g.Task == nil {
			continue
		}
		origVersion, ok := taskVersions[g.TaskID]
		if !ok {
			continue
		}
		for _, p := range g.Proposals {
			if !p.BaseVersion.Equal(origVersion) {
				continue
			}
			if err := applyProposal(p, g.Task); err != nil {
				continue
			}
			if err := pv.store.UpdateStatus(p.ID, mcp.ProposalApproved, now); err != nil {
				continue
			}
			approved++
		}
		// Save the task once after all proposals for it are applied
		if err := pv.provider.SaveTask(g.Task); err != nil {
			continue
		}
	}

	pv.refreshProposals()
	count := approved
	return func() tea.Msg {
		return ProposalBatchApprovedMsg{Count: count}
	}
}

// applyProposal applies the proposal's payload to the task.
func applyProposal(proposal *mcp.Proposal, task *core.Task) error {
	var payload map[string]any
	if err := json.Unmarshal(proposal.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch proposal.Type {
	case mcp.ProposalEnrichMetadata:
		if v, ok := payload["type"].(string); ok {
			task.Type = core.TaskType(v)
		}
		if v, ok := payload["effort"].(string); ok {
			task.Effort = core.TaskEffort(v)
		}
		if v, ok := payload["location"].(string); ok {
			task.Location = core.TaskLocation(v)
		}

	case mcp.ProposalAddContext:
		if v, ok := payload["context"].(string); ok {
			task.Context = v
		}

	case mcp.ProposalAddNote:
		if v, ok := payload["note"].(string); ok {
			task.Notes = append(task.Notes, core.TaskNote{
				Text:      v,
				Timestamp: time.Now().UTC(),
			})
		}

	case mcp.ProposalSuggestBlocker:
		if v, ok := payload["blocker"].(string); ok {
			task.Blocker = v
		}

	case mcp.ProposalSuggestCategory:
		if v, ok := payload["type"].(string); ok {
			task.Type = core.TaskType(v)
		}

	case mcp.ProposalUpdateEffort:
		if v, ok := payload["effort"].(string); ok {
			task.Effort = core.TaskEffort(v)
		}

	case mcp.ProposalAddSubtasks, mcp.ProposalSuggestRelation:
		// These types are noted but don't directly modify the task struct
		if v, ok := payload["note"].(string); ok {
			task.Notes = append(task.Notes, core.TaskNote{
				Text:      v,
				Timestamp: time.Now().UTC(),
			})
		}
	}

	task.UpdatedAt = time.Now().UTC()
	return nil
}

// View renders the proposal review view.
func (pv *ProposalsView) View() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", proposalHeaderStyle.Render("Suggestions"))

	if len(pv.flatIndex) == 0 {
		s.WriteString(helpStyle.Render("No suggestions — LLM clients can propose enrichments via MCP"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Esc to return"))
		return s.String()
	}

	if pv.previewMode {
		return pv.renderPreview()
	}

	// Split pane layout: proposals left, task detail right
	leftWidth := 40
	rightWidth := 40
	if pv.width > 20 {
		leftWidth = (pv.width - 6) / 2
		rightWidth = leftWidth
		if leftWidth < 25 {
			leftWidth = 25
			rightWidth = 25
		}
	}

	leftContent := pv.renderProposalList(leftWidth)
	rightContent := pv.renderTaskDetail(rightWidth)

	leftPane := proposalPaneStyle.Width(leftWidth).Render(leftContent)
	rightPane := proposalPaneStyle.Width(rightWidth).Render(rightContent)

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane))

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Enter approve | Backspace reject | Tab skip | Ctrl+A approve all | p preview | j/k navigate | Esc return"))

	return s.String()
}

// renderProposalList renders the left pane with grouped proposals.
func (pv *ProposalsView) renderProposalList(maxWidth int) string {
	var s strings.Builder

	flatIdx := 0
	for _, g := range pv.groups {
		taskLabel := g.TaskID
		if g.Task != nil {
			taskLabel = truncate(g.Task.Text, maxWidth-4)
		}
		fmt.Fprintf(&s, "%s\n", headerStyle.Render(taskLabel))

		for _, p := range g.Proposals {
			prefix := "  "
			if flatIdx == pv.selectedIndex {
				prefix = "> "
			}

			typeLabel := proposalTypeStyle.Render(string(p.Type))
			rationale := truncate(p.Rationale, maxWidth-10)

			line := fmt.Sprintf("%s%s %s", prefix, typeLabel, rationale)

			if isStale(p, g.Task) {
				line = proposalStaleStyle.Render(line)
			} else if flatIdx == pv.selectedIndex {
				line = proposalSelectedStyle.Render(line)
			}

			fmt.Fprintf(&s, "%s\n", line)
			flatIdx++
		}
		s.WriteString("\n")
	}

	return s.String()
}

// renderTaskDetail renders the right pane with the selected task's details.
func (pv *ProposalsView) renderTaskDetail(maxWidth int) string {
	var s strings.Builder

	task := pv.currentTask()
	proposal := pv.currentProposal()
	if task == nil || proposal == nil {
		s.WriteString(helpStyle.Render("No task selected"))
		return s.String()
	}

	fmt.Fprintf(&s, "%s\n", headerStyle.Render("Task Detail"))
	fmt.Fprintf(&s, "Text: %s\n", truncate(task.Text, maxWidth-6))
	fmt.Fprintf(&s, "Status: %s\n", task.Status)
	if task.Context != "" {
		fmt.Fprintf(&s, "Context: %s\n", truncate(task.Context, maxWidth-9))
	}
	if task.Type != "" {
		fmt.Fprintf(&s, "Type: %s\n", task.Type)
	}
	if task.Effort != "" {
		fmt.Fprintf(&s, "Effort: %s\n", task.Effort)
	}
	if len(task.Notes) > 0 {
		fmt.Fprintf(&s, "Notes: %d\n", len(task.Notes))
	}
	s.WriteString("\n")

	fmt.Fprintf(&s, "%s\n", headerStyle.Render("Proposal"))
	fmt.Fprintf(&s, "Type: %s\n", proposalTypeStyle.Render(string(proposal.Type)))
	fmt.Fprintf(&s, "Source: %s\n", proposal.Source)
	fmt.Fprintf(&s, "Rationale: %s\n", proposal.Rationale)
	fmt.Fprintf(&s, "Created: %s\n", proposal.CreatedAt.Format("2006-01-02 15:04"))

	if isStale(proposal, task) {
		s.WriteString("\n")
		s.WriteString(proposalBadgeStyle.Render("⚠ Task changed since this suggestion"))
	}

	return s.String()
}

// renderPreview renders a diff preview of what the task will look like after applying.
func (pv *ProposalsView) renderPreview() string {
	var s strings.Builder

	proposal := pv.currentProposal()
	task := pv.currentTask()
	if proposal == nil || task == nil {
		s.WriteString(helpStyle.Render("No proposal selected"))
		return s.String()
	}

	fmt.Fprintf(&s, "%s\n\n", proposalHeaderStyle.Render("Preview — Changes"))

	// Build a copy of the task with the proposal applied
	previewTask := *task
	previewTask.Notes = make([]core.TaskNote, len(task.Notes))
	copy(previewTask.Notes, task.Notes)

	if err := applyProposal(proposal, &previewTask); err != nil {
		fmt.Fprintf(&s, "Preview error: %v\n", err)
		return s.String()
	}

	colWidth := 40
	if pv.width > 20 {
		colWidth = (pv.width - 8) / 2
		if colWidth < 20 {
			colWidth = 20
		}
	}

	beforeLines := formatTaskPreview(task, "BEFORE")
	afterLines := formatTaskPreview(&previewTask, "AFTER")

	// Diff highlighting: compare line by line
	var leftRendered, rightRendered strings.Builder
	beforeSplit := strings.Split(beforeLines, "\n")
	afterSplit := strings.Split(afterLines, "\n")

	maxLines := len(beforeSplit)
	if len(afterSplit) > maxLines {
		maxLines = len(afterSplit)
	}

	for i := 0; i < maxLines; i++ {
		bLine := ""
		aLine := ""
		if i < len(beforeSplit) {
			bLine = beforeSplit[i]
		}
		if i < len(afterSplit) {
			aLine = afterSplit[i]
		}

		if bLine != aLine {
			if bLine != "" {
				fmt.Fprintf(&leftRendered, "%s\n", proposalDiffRemoveStyle.Render(bLine))
			}
			if aLine != "" {
				fmt.Fprintf(&rightRendered, "%s\n", proposalDiffAddStyle.Render(aLine))
			}
		} else {
			fmt.Fprintf(&leftRendered, "%s\n", bLine)
			fmt.Fprintf(&rightRendered, "%s\n", aLine)
		}
	}

	leftBox := proposalPaneStyle.Width(colWidth).Render(leftRendered.String())
	rightBox := proposalPaneStyle.Width(colWidth).Render(rightRendered.String())

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Enter approve | p/Esc close preview"))

	return s.String()
}

func formatTaskPreview(t *core.Task, label string) string {
	var s strings.Builder
	fmt.Fprintf(&s, "%s\n", label)
	fmt.Fprintf(&s, "─────────\n")
	fmt.Fprintf(&s, "Text: %s\n", t.Text)
	fmt.Fprintf(&s, "Status: %s\n", t.Status)
	if t.Type != "" {
		fmt.Fprintf(&s, "Type: %s\n", t.Type)
	}
	if t.Effort != "" {
		fmt.Fprintf(&s, "Effort: %s\n", t.Effort)
	}
	if t.Location != "" {
		fmt.Fprintf(&s, "Location: %s\n", t.Location)
	}
	if t.Context != "" {
		fmt.Fprintf(&s, "Context: %s\n", t.Context)
	}
	if t.Blocker != "" {
		fmt.Fprintf(&s, "Blocker: %s\n", t.Blocker)
	}
	if len(t.Notes) > 0 {
		fmt.Fprintf(&s, "Notes: %d\n", len(t.Notes))
		for _, n := range t.Notes {
			fmt.Fprintf(&s, "  - %s\n", n.Text)
		}
	}
	return s.String()
}

// PendingProposalCount returns the number of pending proposals in the store.
func PendingProposalCount(store *mcp.ProposalStore) int {
	if store == nil {
		return 0
	}
	return len(store.List(mcp.ProposalFilter{Status: mcp.ProposalPending}))
}
