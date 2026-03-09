package core

import (
	"strings"
	"time"
)

const (
	// FocusTag is the inline tag that marks a task as a focus task.
	FocusTag = "+focus"

	// FocusBoost is the scoring bonus applied to focus-tagged tasks in door selection.
	FocusBoost = 5.0

	// focusExpiryHours is the number of hours after which focus boost expires.
	focusExpiryHours = 16
)

// HasFocusTag returns true if the task's text contains the +focus tag.
func HasFocusTag(task *Task) bool {
	for _, word := range strings.Fields(task.Text) {
		if strings.EqualFold(word, FocusTag) {
			return true
		}
	}
	return false
}

// GetFocusTasks returns all tasks in the pool that have the +focus tag.
func GetFocusTasks(pool *TaskPool) []*Task {
	var result []*Task
	for _, t := range pool.GetAllTasks() {
		if HasFocusTag(t) {
			result = append(result, t)
		}
	}
	return result
}

// ClearFocusTags removes the +focus tag from all tasks in the pool.
func ClearFocusTags(pool *TaskPool) {
	for _, t := range pool.GetAllTasks() {
		if HasFocusTag(t) {
			t.Text = RemoveFocusTagFromText(t.Text)
			t.UpdatedAt = time.Now().UTC()
		}
	}
}

// IsFocusExpired returns true if the focus boost has expired.
// Focus expires after 16 hours from the planning session timestamp.
func IsFocusExpired(planningTimestamp time.Time) bool {
	return time.Now().UTC().Sub(planningTimestamp) >= time.Duration(focusExpiryHours)*time.Hour
}

// FocusScoreBoost returns the focus boost for a candidate set of tasks.
// Each focus-tagged task adds FocusBoost to the total score,
// but only if the focus has not expired.
func FocusScoreBoost(tasks []*Task, planningTimestamp time.Time) float64 {
	if IsFocusExpired(planningTimestamp) {
		return 0
	}
	boost := 0.0
	for _, t := range tasks {
		if HasFocusTag(t) {
			boost += FocusBoost
		}
	}
	return boost
}

// RemoveFocusTagFromText removes all occurrences of +focus (case-insensitive) from text.
func RemoveFocusTagFromText(text string) string {
	words := strings.Fields(text)
	var remaining []string
	for _, w := range words {
		if !strings.EqualFold(w, FocusTag) {
			remaining = append(remaining, w)
		}
	}
	return strings.Join(remaining, " ")
}
