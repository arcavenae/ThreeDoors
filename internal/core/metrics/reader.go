// Package metrics provides a reusable library for reading session metrics
// from the ThreeDoors JSONL session log. It is the primary I/O layer for
// Epic 4 (Learning & Intelligent Door Selection).
package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// Reader reads session metrics from a JSONL file.
// It gracefully handles missing files and corrupted lines.
type Reader struct {
	path string
}

// NewReader creates a Reader for the given sessions.jsonl file path.
func NewReader(path string) *Reader {
	return &Reader{path: path}
}

// ReadAll returns all valid sessions from the JSONL file.
// Returns nil slice and nil error for missing or empty files.
// Corrupted lines are silently skipped.
func (r *Reader) ReadAll() ([]core.SessionMetrics, error) {
	return r.readFiltered(func(core.SessionMetrics) bool { return true })
}

// ReadSince returns sessions with StartTime at or after the given time.
// Returns nil slice and nil error for missing or empty files.
// Corrupted lines are silently skipped.
func (r *Reader) ReadSince(since time.Time) ([]core.SessionMetrics, error) {
	return r.readFiltered(func(sm core.SessionMetrics) bool {
		return !sm.StartTime.Before(since)
	})
}

// ReadLast returns the last n sessions from the file.
// If n <= 0, returns nil. If n exceeds total sessions, returns all.
// Returns nil slice and nil error for missing or empty files.
// Corrupted lines are silently skipped.
func (r *Reader) ReadLast(n int) ([]core.SessionMetrics, error) {
	if n <= 0 {
		return nil, nil
	}

	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// readFiltered reads the JSONL file and returns sessions matching the predicate.
func (r *Reader) readFiltered(pred func(core.SessionMetrics) bool) ([]core.SessionMetrics, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open sessions file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var sessions []core.SessionMetrics
	scanner := core.NewLimitedScanner(f, core.MaxJSONLLineSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var sm core.SessionMetrics
		if err := json.Unmarshal([]byte(line), &sm); err != nil {
			continue // skip corrupted lines
		}
		if pred(sm) {
			sessions = append(sessions, sm)
		}
	}
	if err := scanner.Err(); err != nil {
		return sessions, fmt.Errorf("read sessions file: %w", err)
	}
	return sessions, nil
}
