package enrichment

import (
	"fmt"
	"strings"
	"time"
)

// TaskMetadata stores enrichment data for a task that may not be supported
// by the primary storage backend.
type TaskMetadata struct {
	TaskID         string
	Category       string
	EnrichmentTags []string
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UpsertTaskMetadata inserts or updates metadata for a task.
func (edb *DB) UpsertTaskMetadata(m *TaskMetadata) error {
	now := time.Now().UTC()
	tags := strings.Join(m.EnrichmentTags, ",")
	_, err := edb.db.Exec(`
		INSERT INTO task_metadata (task_id, category, enrichment_tags, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET
			category = excluded.category,
			enrichment_tags = excluded.enrichment_tags,
			notes = excluded.notes,
			updated_at = excluded.updated_at
	`, m.TaskID, m.Category, tags, m.Notes, now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("upsert task metadata %s: %w", m.TaskID, err)
	}
	return nil
}

// GetTaskMetadata retrieves metadata for a task by ID.
func (edb *DB) GetTaskMetadata(taskID string) (*TaskMetadata, error) {
	var m TaskMetadata
	var tags, createdAt, updatedAt string
	err := edb.db.QueryRow(
		"SELECT task_id, category, enrichment_tags, notes, created_at, updated_at FROM task_metadata WHERE task_id = ?",
		taskID,
	).Scan(&m.TaskID, &m.Category, &tags, &m.Notes, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("get task metadata %s: %w", taskID, err)
	}

	if tags != "" {
		m.EnrichmentTags = strings.Split(tags, ",")
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &m, nil
}

// DeleteTaskMetadata removes metadata for a task.
func (edb *DB) DeleteTaskMetadata(taskID string) error {
	_, err := edb.db.Exec("DELETE FROM task_metadata WHERE task_id = ?", taskID)
	if err != nil {
		return fmt.Errorf("delete task metadata %s: %w", taskID, err)
	}
	return nil
}
