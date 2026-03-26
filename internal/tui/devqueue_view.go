package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/dispatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DevQueueView displays the dev dispatch queue with approve/reject/kill actions.
type DevQueueView struct {
	queue      *dispatch.DevQueue
	dispatcher dispatch.Dispatcher
	items      []dispatch.QueueItem
	cursor     int
	width      int
	height     int
	flash      string
}

// NewDevQueueView creates a new DevQueueView.
func NewDevQueueView(queue *dispatch.DevQueue, dispatcher dispatch.Dispatcher) *DevQueueView {
	dv := &DevQueueView{
		queue:      queue,
		dispatcher: dispatcher,
	}
	dv.refreshItems()
	return dv
}

// SetWidth sets the terminal width for rendering.
func (dv *DevQueueView) SetWidth(w int) {
	dv.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (dv *DevQueueView) SetHeight(h int) {
	dv.height = h
}

// Update handles key input for the dev queue view.
func (dv *DevQueueView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case DevQueueWorkerCreatedMsg:
		if msg.Err != nil {
			dv.flash = fmt.Sprintf("Dispatch failed: %s", msg.Err.Error())
		} else {
			dv.flash = fmt.Sprintf("Worker %s created", msg.WorkerName)
		}
		dv.refreshItems()
		return ClearFlashCmd()

	case DevQueueWorkerRemovedMsg:
		if msg.Err != nil {
			dv.flash = fmt.Sprintf("Kill failed: %s", msg.Err.Error())
		} else {
			dv.flash = "Worker killed"
		}
		dv.refreshItems()
		return ClearFlashCmd()

	case ClearFlashMsg:
		dv.flash = ""
		return nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }

		case "j", "down":
			if dv.cursor < len(dv.items)-1 {
				dv.cursor++
			}

		case "k", "up":
			if dv.cursor > 0 {
				dv.cursor--
			}

		case "y":
			return dv.approveItem()

		case "n":
			return dv.rejectItem()

		case "K":
			return dv.killItem()
		}
	}
	return nil
}

// View renders the dev queue view.
func (dv *DevQueueView) View() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", devQueueHeaderStyle.Render("Dev Queue"))

	if len(dv.items) == 0 {
		s.WriteString(helpStyle.Render("No items in dev queue. Dispatch a task with 'x' from detail view."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Esc to return"))
		return s.String()
	}

	// Calculate column widths
	maxTaskWidth := dv.width - 45
	if maxTaskWidth < 20 {
		maxTaskWidth = 20
	}

	for i, item := range dv.items {
		icon := statusIcon(item.Status)
		taskText := truncate(item.TaskText, maxTaskWidth)
		worker := item.WorkerName
		if worker == "" {
			worker = "—"
		}
		pr := "—"
		if item.PRNumber > 0 {
			pr = fmt.Sprintf("#%d", item.PRNumber)
		}
		queuedAt := relativeTime(item.QueuedAt)

		line := fmt.Sprintf("  %s  %-*s  %-12s  %-5s  %s",
			icon, maxTaskWidth, taskText, worker, pr, queuedAt)

		if i == dv.cursor {
			fmt.Fprintf(&s, "%s\n", devQueueSelectedStyle.Render(line))
		} else {
			fmt.Fprintf(&s, "%s\n", devQueueItemStyle.Render(line))
		}
	}

	s.WriteString("\n")

	if dv.flash != "" {
		fmt.Fprintf(&s, "%s\n", flashStyle.Render(dv.flash))
	}

	s.WriteString(helpStyle.Render("j/k navigate | y approve | n reject | K kill | Esc return"))

	return s.String()
}

// approveItem dispatches the selected pending item via Dispatcher.CreateWorker.
func (dv *DevQueueView) approveItem() tea.Cmd {
	if dv.cursor >= len(dv.items) {
		return nil
	}
	item := dv.items[dv.cursor]
	if item.Status != dispatch.QueueItemPending {
		dv.flash = "Only pending items can be approved"
		return ClearFlashCmd()
	}

	queue := dv.queue
	dispatcher := dv.dispatcher
	itemID := item.ID
	taskDesc := dispatch.BuildTaskDescription(item)

	return func() tea.Msg {
		// Update status to dispatched
		now := time.Now().UTC()
		if err := queue.Update(itemID, func(qi *dispatch.QueueItem) {
			qi.Status = dispatch.QueueItemDispatched
			qi.DispatchedAt = &now
		}); err != nil {
			return DevQueueWorkerCreatedMsg{Err: fmt.Errorf("update queue item %s: %w", itemID, err)}
		}

		workerName, err := dispatcher.CreateWorker(context.Background(), taskDesc)
		if err != nil {
			// Revert status on failure
			_ = queue.Update(itemID, func(qi *dispatch.QueueItem) {
				qi.Status = dispatch.QueueItemPending
				qi.DispatchedAt = nil
			})
			return DevQueueWorkerCreatedMsg{Err: fmt.Errorf("create worker: %w", err)}
		}

		_ = queue.Update(itemID, func(qi *dispatch.QueueItem) {
			qi.WorkerName = workerName
		})

		return DevQueueWorkerCreatedMsg{WorkerName: workerName}
	}
}

// rejectItem removes the selected pending item from the queue.
func (dv *DevQueueView) rejectItem() tea.Cmd {
	if dv.cursor >= len(dv.items) {
		return nil
	}
	item := dv.items[dv.cursor]
	if item.Status != dispatch.QueueItemPending {
		dv.flash = "Only pending items can be rejected"
		return ClearFlashCmd()
	}

	if err := dv.queue.Remove(item.ID); err != nil {
		dv.flash = fmt.Sprintf("Remove failed: %s", err.Error())
		return ClearFlashCmd()
	}

	dv.refreshItems()
	if dv.cursor >= len(dv.items) && dv.cursor > 0 {
		dv.cursor--
	}

	dv.flash = "Item removed from queue"
	return ClearFlashCmd()
}

// killItem stops a dispatched/running worker.
func (dv *DevQueueView) killItem() tea.Cmd {
	if dv.cursor >= len(dv.items) {
		return nil
	}
	item := dv.items[dv.cursor]
	if item.Status != dispatch.QueueItemDispatched {
		dv.flash = "Only dispatched items can be killed"
		return ClearFlashCmd()
	}

	queue := dv.queue
	dispatcher := dv.dispatcher
	itemID := item.ID
	workerName := item.WorkerName

	return func() tea.Msg {
		err := dispatcher.RemoveWorker(context.Background(), workerName)
		if err != nil {
			return DevQueueWorkerRemovedMsg{Err: fmt.Errorf("remove worker %s: %w", workerName, err)}
		}

		_ = queue.Update(itemID, func(qi *dispatch.QueueItem) {
			qi.Status = dispatch.QueueItemFailed
			qi.Error = "Killed by user"
		})

		return DevQueueWorkerRemovedMsg{}
	}
}

// refreshItems reloads items from the queue.
func (dv *DevQueueView) refreshItems() {
	dv.items = dv.queue.List()
}

// statusIcon returns the icon for a queue item status.
func statusIcon(status dispatch.QueueItemStatus) string {
	switch status {
	case dispatch.QueueItemPending:
		return "⏳"
	case dispatch.QueueItemDispatched:
		return "⚙️"
	case dispatch.QueueItemCompleted:
		return "✅"
	case dispatch.QueueItemFailed:
		return "❌"
	default:
		return "?"
	}
}

// truncate shortens text to maxLen, appending "…" if truncated.
func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 1 {
		return "…"
	}
	return text[:maxLen-1] + "…"
}

// relativeTime formats a time pointer as a relative duration like "2m ago".
func relativeTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// Dev queue styles
var (
	devQueueHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	devQueueItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	devQueueSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Background(lipgloss.Color("236"))
)
