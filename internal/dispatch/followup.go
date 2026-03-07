package dispatch

import (
	"fmt"
	"strings"
)

// FollowUpTask holds the data needed to create a follow-up task in the core task pool.
// This avoids a dependency from dispatch → core.
type FollowUpTask struct {
	Text        string
	Context     string
	DevDispatch *DevDispatch
}

// GenerateFollowUpTasks creates review (and optionally CI-fix) follow-up tasks
// for a completed queue item. existingTexts is the set of current task texts
// used for deduplication.
func GenerateFollowUpTasks(item QueueItem, existingTexts map[string]bool) []FollowUpTask {
	if item.PRNumber == 0 {
		return nil
	}

	var tasks []FollowUpTask

	reviewText := fmt.Sprintf("Review PR #%d: %s", item.PRNumber, item.TaskText)
	if !hasTaskWithPrefix(existingTexts, fmt.Sprintf("Review PR #%d:", item.PRNumber)) {
		tasks = append(tasks, FollowUpTask{
			Text:    reviewText,
			Context: fmt.Sprintf("Auto-generated from task %s", item.TaskID),
			DevDispatch: &DevDispatch{
				PRNumber: item.PRNumber,
			},
		})
	}

	if item.Status == QueueItemFailed && item.Error != "" {
		ciText := fmt.Sprintf("Fix CI on PR #%d: %s", item.PRNumber, item.Error)
		if !hasTaskWithPrefix(existingTexts, fmt.Sprintf("Fix CI on PR #%d:", item.PRNumber)) {
			tasks = append(tasks, FollowUpTask{
				Text:    ciText,
				Context: fmt.Sprintf("Auto-generated from task %s", item.TaskID),
				DevDispatch: &DevDispatch{
					PRNumber: item.PRNumber,
				},
			})
		}
	}

	return tasks
}

// hasTaskWithPrefix checks if any key in the set starts with the given prefix.
func hasTaskWithPrefix(texts map[string]bool, prefix string) bool {
	for text := range texts {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}
