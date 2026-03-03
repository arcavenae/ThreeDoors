package enrichment

import (
	"fmt"
	"time"
)

// CrossReference links two tasks across systems.
type CrossReference struct {
	ID           int64
	SourceTaskID string
	TargetTaskID string
	SourceSystem string
	Relationship string
	CreatedAt    time.Time
}

// AddCrossReference creates a link between two tasks.
func (edb *DB) AddCrossReference(ref *CrossReference) error {
	now := time.Now().UTC()
	result, err := edb.db.Exec(`
		INSERT INTO cross_references (source_task_id, target_task_id, source_system, relationship, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, ref.SourceTaskID, ref.TargetTaskID, ref.SourceSystem, ref.Relationship, now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("add cross reference: %w", err)
	}
	ref.ID, _ = result.LastInsertId()
	ref.CreatedAt = now
	return nil
}

// GetCrossReferences returns all references for a task (as source or target).
func (edb *DB) GetCrossReferences(taskID string) ([]CrossReference, error) {
	rows, err := edb.db.Query(`
		SELECT id, source_task_id, target_task_id, source_system, relationship, created_at
		FROM cross_references
		WHERE source_task_id = ? OR target_task_id = ?
	`, taskID, taskID)
	if err != nil {
		return nil, fmt.Errorf("get cross references for %s: %w", taskID, err)
	}
	defer func() { _ = rows.Close() }()

	var refs []CrossReference
	for rows.Next() {
		var r CrossReference
		var createdAt string
		if err := rows.Scan(&r.ID, &r.SourceTaskID, &r.TargetTaskID, &r.SourceSystem, &r.Relationship, &createdAt); err != nil {
			return nil, fmt.Errorf("scan cross reference: %w", err)
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		refs = append(refs, r)
	}
	return refs, rows.Err()
}

// DeleteCrossReference removes a cross-reference by ID.
func (edb *DB) DeleteCrossReference(id int64) error {
	_, err := edb.db.Exec("DELETE FROM cross_references WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete cross reference %d: %w", id, err)
	}
	return nil
}
