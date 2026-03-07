package reminders

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// RemindersProvider implements core.TaskProvider for Apple Reminders,
// supporting both read and write operations via JXA scripts.
type RemindersProvider struct {
	executor    CommandExecutor
	lists       []string
	defaultList string
	retry       RetryConfig
}

// NewRemindersProvider creates a RemindersProvider that reads from the given
// lists. An empty lists slice means all lists will be read. New reminders
// are created in the first configured list, or "Reminders" if none specified.
func NewRemindersProvider(executor CommandExecutor, lists []string) *RemindersProvider {
	defaultList := "Reminders"
	if len(lists) > 0 {
		defaultList = lists[0]
	}
	return &RemindersProvider{
		executor:    executor,
		lists:       lists,
		defaultList: defaultList,
		retry:       DefaultRetryConfig(),
	}
}

// Name returns the provider identifier.
func (p *RemindersProvider) Name() string {
	return "reminders"
}

// LoadTasks reads incomplete reminders via JXA and maps them to core.Task values.
func (p *RemindersProvider) LoadTasks() ([]*core.Task, error) {
	ctx := context.Background()

	listNames, err := p.resolveListNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve reminder lists: %w", err)
	}

	var tasks []*core.Task
	for _, listName := range listNames {
		reminders, err := ReadReminders(ctx, p.executor, listName)
		if err != nil {
			return nil, fmt.Errorf("load reminders from %q: %w", listName, err)
		}
		for _, r := range reminders {
			tasks = append(tasks, mapReminderToTask(r, listName))
		}
	}

	return tasks, nil
}

// SaveTask persists a single task to Apple Reminders. If the task ID is empty
// or does not match an existing reminder, a new reminder is created and the
// task's ID is updated. Otherwise the existing reminder is updated in place.
func (p *RemindersProvider) SaveTask(task *core.Task) error {
	ctx := context.Background()
	name, body, priority := mapTaskToReminderFields(task)

	if task.ID == "" {
		return p.createReminder(ctx, task, name, body, priority)
	}

	// Try updating first; if the ID isn't found, create instead.
	err := withRetry(ctx, p.retry, func() error {
		return categorizeError(UpdateReminder(ctx, p.executor, task.ID, name, body, priority))
	})
	if errors.Is(err, ErrReminderNotFound) {
		return p.createReminder(ctx, task, name, body, priority)
	}
	return err
}

// SaveTasks persists a batch of tasks by saving each individually.
func (p *RemindersProvider) SaveTasks(tasks []*core.Task) error {
	for _, task := range tasks {
		if err := p.SaveTask(task); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTask removes a reminder by its ID.
func (p *RemindersProvider) DeleteTask(taskID string) error {
	ctx := context.Background()
	return withRetry(ctx, p.retry, func() error {
		return categorizeError(DeleteReminder(ctx, p.executor, taskID))
	})
}

// MarkComplete marks a reminder as completed by its ID.
func (p *RemindersProvider) MarkComplete(taskID string) error {
	ctx := context.Background()
	return withRetry(ctx, p.retry, func() error {
		return categorizeError(CompleteReminder(ctx, p.executor, taskID))
	})
}

// createReminder creates a new reminder and updates task.ID with the new ID.
func (p *RemindersProvider) createReminder(ctx context.Context, task *core.Task, name, body string, priority int) error {
	return withRetry(ctx, p.retry, func() error {
		id, err := CreateReminder(ctx, p.executor, name, body, priority, p.defaultList)
		if err != nil {
			return categorizeError(err)
		}
		task.ID = id
		return nil
	})
}

// Watch returns nil — the reminders provider does not support watching.
func (p *RemindersProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck attempts a lightweight read and reports TCC permission status.
func (p *RemindersProvider) HealthCheck() core.HealthCheckResult {
	start := time.Now().UTC()

	item := core.HealthCheckItem{Name: "Apple Reminders"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ReadLists(ctx, p.executor)
	if err != nil {
		item.Status = core.HealthFail
		errMsg := err.Error()
		if strings.Contains(errMsg, "not allowed") || strings.Contains(errMsg, "denied") || strings.Contains(errMsg, "1002") {
			item.Message = "Reminders access denied by macOS privacy settings"
			item.Suggestion = "Grant Reminders access in System Settings > Privacy & Security > Reminders"
		} else {
			item.Message = fmt.Sprintf("Cannot access Reminders: %s", errMsg)
			item.Suggestion = "Verify osascript is available and Reminders app is installed"
		}
	} else {
		item.Status = core.HealthOK
		item.Message = "Apple Reminders accessible"
	}

	return core.HealthCheckResult{
		Items:    []core.HealthCheckItem{item},
		Overall:  item.Status,
		Duration: time.Since(start),
	}
}

// resolveListNames returns the configured list names, or discovers all lists
// if none were configured.
func (p *RemindersProvider) resolveListNames(ctx context.Context) ([]string, error) {
	if len(p.lists) > 0 {
		return p.lists, nil
	}
	return ReadLists(ctx, p.executor)
}

// mapPriorityToEffort converts Apple Reminders priority (0-9) to core.TaskEffort.
// Reminders priority: 1 = highest, 5 = medium, 9 = lowest, 0 = none.
func mapPriorityToEffort(priority int) core.TaskEffort {
	switch {
	case priority >= 1 && priority <= 4:
		return core.EffortDeepWork
	case priority == 5:
		return core.EffortMedium
	case priority >= 6 && priority <= 9:
		return core.EffortQuickWin
	default:
		return ""
	}
}

// mapReminderToTask converts a ReminderJSON to a core.Task.
func mapReminderToTask(r ReminderJSON, listName string) *core.Task {
	status := core.StatusTodo
	if r.Completed {
		status = core.StatusComplete
	}

	createdAt := parseISOTime(r.CreationDate)
	updatedAt := parseISOTime(r.ModificationDate)

	var contextParts []string
	if r.DueDate != nil && *r.DueDate != "" {
		dueTime := parseISOTime(*r.DueDate)
		if !dueTime.IsZero() {
			contextParts = append(contextParts, fmt.Sprintf("due:%s", dueTime.Format("2006-01-02")))
		}
	}

	var notes []core.TaskNote
	if r.Body != "" {
		notes = append(notes, core.TaskNote{
			Timestamp: createdAt,
			Text:      r.Body,
		})
	}

	task := &core.Task{
		ID:             r.ID,
		Text:           r.Name,
		Context:        strings.Join(contextParts, " "),
		Status:         status,
		Effort:         mapPriorityToEffort(r.Priority),
		Notes:          notes,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		SourceProvider: fmt.Sprintf("reminders:%s", listName),
		SourceRefs: []core.SourceRef{{
			Provider: fmt.Sprintf("reminders:%s", listName),
			NativeID: r.ID,
		}},
	}

	if r.Completed {
		task.CompletedAt = &updatedAt
	}

	return task
}

// mapTaskToReminderFields extracts reminder-compatible fields from a core.Task.
// Text maps to name, first note's text maps to body, effort maps to priority.
func mapTaskToReminderFields(task *core.Task) (name, body string, priority int) {
	name = task.Text
	if len(task.Notes) > 0 {
		body = task.Notes[0].Text
	}
	priority = mapEffortToPriority(task.Effort)
	return
}

// mapEffortToPriority converts core.TaskEffort to Apple Reminders priority.
// deep-work = 1 (highest), medium = 5, quick-win = 9 (lowest), none = 0.
func mapEffortToPriority(effort core.TaskEffort) int {
	switch effort {
	case core.EffortDeepWork:
		return 1
	case core.EffortMedium:
		return 5
	case core.EffortQuickWin:
		return 9
	default:
		return 0
	}
}

// parseISOTime parses an ISO 8601 timestamp, returning zero time on failure.
func parseISOTime(s string) time.Time {
	for _, layout := range []string{
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
