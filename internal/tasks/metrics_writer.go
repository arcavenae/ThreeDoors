package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// MetricsWriter handles persistence of session metrics to JSON Lines format
// Each session is appended as a single JSON line to sessions.jsonl
type MetricsWriter struct {
	sessionsPath string
}

// NewMetricsWriter creates a new metrics writer
// baseDir should be the ~/.threedoors directory
func NewMetricsWriter(baseDir string) *MetricsWriter {
	return &MetricsWriter{
		sessionsPath: filepath.Join(baseDir, "sessions.jsonl"),
	}
}

// AppendSession writes session metrics as a JSON line to sessions.jsonl
// Creates the file if it doesn't exist
// Returns error if file operations fail (caller should log warning, not crash)
func (mw *MetricsWriter) AppendSession(metrics *SessionMetrics) error {
	// Open file in append mode, create if not exists
	f, err := os.OpenFile(mw.sessionsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Marshal to JSON (compact, single line)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	// Append JSON line with newline
	_, err = f.Write(append(data, '\n'))
	return err
}
