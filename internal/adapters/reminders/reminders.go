package reminders

import (
	"context"
	"encoding/json"
	"fmt"
)

// operationResult represents the JSON response from mutating JXA scripts.
type operationResult struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ReadReminders executes the JXA script to read incomplete reminders from
// the specified list and returns the parsed results.
func ReadReminders(ctx context.Context, exec CommandExecutor, listName string) ([]ReminderJSON, error) {
	script := scriptReadReminders(listName)
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("read reminders from %q: %w", listName, err)
	}

	var reminders []ReminderJSON
	if err := json.Unmarshal([]byte(output), &reminders); err != nil {
		return nil, fmt.Errorf("parse reminders JSON: %w", err)
	}

	return reminders, nil
}

// ReadLists executes the JXA script to read all reminder list names.
func ReadLists(ctx context.Context, exec CommandExecutor) ([]string, error) {
	script := scriptReadLists()
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("read reminder lists: %w", err)
	}

	var lists []string
	if err := json.Unmarshal([]byte(output), &lists); err != nil {
		return nil, fmt.Errorf("parse lists JSON: %w", err)
	}

	return lists, nil
}

// CompleteReminder executes the JXA script to mark a reminder as completed.
func CompleteReminder(ctx context.Context, exec CommandExecutor, reminderID string) error {
	script := scriptCompleteReminder(reminderID)
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return fmt.Errorf("complete reminder %q: %w", reminderID, err)
	}

	var result operationResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("parse complete result: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("complete reminder %q: %s", reminderID, result.Error)
	}

	return nil
}

// CreateReminder executes the JXA script to create a new reminder in the
// specified list. Returns the ID of the newly created reminder.
func CreateReminder(ctx context.Context, exec CommandExecutor, name, body string, priority int, listName string) (string, error) {
	script := scriptCreateReminder(name, body, priority, listName)
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return "", fmt.Errorf("create reminder in %q: %w", listName, err)
	}

	var result operationResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return "", fmt.Errorf("parse create result: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("create reminder in %q: %s", listName, result.Error)
	}

	return result.ID, nil
}

// UpdateReminder executes the JXA script to update an existing reminder's
// name, body, and priority.
func UpdateReminder(ctx context.Context, exec CommandExecutor, reminderID, name, body string, priority int) error {
	script := scriptUpdateReminder(reminderID, name, body, priority)
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return fmt.Errorf("update reminder %q: %w", reminderID, err)
	}

	var result operationResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("parse update result: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("update reminder %q: %s", reminderID, result.Error)
	}

	return nil
}

// DeleteReminder executes the JXA script to delete a reminder by ID.
func DeleteReminder(ctx context.Context, exec CommandExecutor, reminderID string) error {
	script := scriptDeleteReminder(reminderID)
	output, err := exec.Execute(ctx, script)
	if err != nil {
		return fmt.Errorf("delete reminder %q: %w", reminderID, err)
	}

	var result operationResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("parse delete result: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("delete reminder %q: %s", reminderID, result.Error)
	}

	return nil
}
