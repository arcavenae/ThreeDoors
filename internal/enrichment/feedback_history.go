package enrichment

import (
	"fmt"
	"time"
)

// FeedbackEntry records door feedback and mood correlation data.
type FeedbackEntry struct {
	ID           int64
	TaskID       string
	FeedbackType string
	Mood         string
	Comment      string
	SessionID    string
	CreatedAt    time.Time
}

// AddFeedback records a feedback entry.
func (edb *DB) AddFeedback(f *FeedbackEntry) error {
	now := time.Now().UTC()
	result, err := edb.db.Exec(`
		INSERT INTO feedback_history (task_id, feedback_type, mood, comment, session_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, f.TaskID, f.FeedbackType, f.Mood, f.Comment, f.SessionID, now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("add feedback: %w", err)
	}
	f.ID, _ = result.LastInsertId()
	f.CreatedAt = now
	return nil
}

// GetFeedbackByTask returns all feedback entries for a task.
func (edb *DB) GetFeedbackByTask(taskID string) ([]FeedbackEntry, error) {
	rows, err := edb.db.Query(`
		SELECT id, task_id, feedback_type, mood, comment, session_id, created_at
		FROM feedback_history
		WHERE task_id = ?
		ORDER BY created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get feedback for task %s: %w", taskID, err)
	}
	defer func() { _ = rows.Close() }()

	var entries []FeedbackEntry
	for rows.Next() {
		var f FeedbackEntry
		var createdAt string
		if err := rows.Scan(&f.ID, &f.TaskID, &f.FeedbackType, &f.Mood, &f.Comment, &f.SessionID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, f)
	}
	return entries, rows.Err()
}

// GetFeedbackBySession returns all feedback entries for a session.
func (edb *DB) GetFeedbackBySession(sessionID string) ([]FeedbackEntry, error) {
	rows, err := edb.db.Query(`
		SELECT id, task_id, feedback_type, mood, comment, session_id, created_at
		FROM feedback_history
		WHERE session_id = ?
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get feedback for session %s: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	var entries []FeedbackEntry
	for rows.Next() {
		var f FeedbackEntry
		var createdAt string
		if err := rows.Scan(&f.ID, &f.TaskID, &f.FeedbackType, &f.Mood, &f.Comment, &f.SessionID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, f)
	}
	return entries, rows.Err()
}
