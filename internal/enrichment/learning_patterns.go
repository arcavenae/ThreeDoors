package enrichment

import (
	"fmt"
	"time"
)

// LearningPattern stores algorithm weights and pattern data for adaptive
// door selection.
type LearningPattern struct {
	ID          int64
	PatternType string
	PatternKey  string
	Weight      float64
	Data        string
	UpdatedAt   time.Time
}

// UpsertLearningPattern inserts or updates a learning pattern entry.
func (edb *DB) UpsertLearningPattern(p *LearningPattern) error {
	now := time.Now().UTC()
	_, err := edb.db.Exec(`
		INSERT INTO learning_patterns (pattern_type, pattern_key, weight, data, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(pattern_type, pattern_key) DO UPDATE SET
			weight = excluded.weight,
			data = excluded.data,
			updated_at = excluded.updated_at
	`, p.PatternType, p.PatternKey, p.Weight, p.Data, now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("upsert learning pattern %s/%s: %w", p.PatternType, p.PatternKey, err)
	}
	return nil
}

// GetLearningPatternsByType returns all patterns of a given type.
func (edb *DB) GetLearningPatternsByType(patternType string) ([]LearningPattern, error) {
	rows, err := edb.db.Query(
		"SELECT id, pattern_type, pattern_key, weight, data, updated_at FROM learning_patterns WHERE pattern_type = ?",
		patternType,
	)
	if err != nil {
		return nil, fmt.Errorf("get learning patterns by type %s: %w", patternType, err)
	}
	defer func() { _ = rows.Close() }()

	var patterns []LearningPattern
	for rows.Next() {
		var p LearningPattern
		var updatedAt string
		if err := rows.Scan(&p.ID, &p.PatternType, &p.PatternKey, &p.Weight, &p.Data, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan learning pattern: %w", err)
		}
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// DeleteLearningPattern removes a pattern by type and key.
func (edb *DB) DeleteLearningPattern(patternType, patternKey string) error {
	_, err := edb.db.Exec(
		"DELETE FROM learning_patterns WHERE pattern_type = ? AND pattern_key = ?",
		patternType, patternKey,
	)
	if err != nil {
		return fmt.Errorf("delete learning pattern %s/%s: %w", patternType, patternKey, err)
	}
	return nil
}
