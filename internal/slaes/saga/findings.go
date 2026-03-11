package saga

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// FindingType identifies the kind of JSONL finding entry.
type FindingType string

const (
	FindingSagaDetected FindingType = "saga_detected"
	FindingSagaRecur    FindingType = "saga_recurrence"
)

// SagaFinding is a JSONL entry for a saga detection event. It extends the
// base retrospector findings schema with saga-specific fields.
type SagaFinding struct {
	Type            FindingType      `json:"type"`
	Branch          string           `json:"branch"`
	PR              int              `json:"pr,omitempty"`
	SagaType        SagaType         `json:"saga_type"`
	WorkerCount     int              `json:"worker_count"`
	WorkerNames     []string         `json:"worker_names"`
	FailureRelation FailureRelation  `json:"failure_relation"`
	Recommendation  []Recommendation `json:"recommendations"`
	Timestamp       time.Time        `json:"timestamp"`
	Repo            string           `json:"repo"`
}

// FindingsLog manages the JSONL findings log for saga events.
type FindingsLog struct {
	path       string
	maxEntries int
}

// NewFindingsLog creates a findings log writer. maxEntries controls rolling
// retention — when exceeded, oldest entries are removed.
func NewFindingsLog(path string, maxEntries int) *FindingsLog {
	return &FindingsLog{
		path:       path,
		maxEntries: maxEntries,
	}
}

// Append writes a saga finding to the JSONL log.
func (fl *FindingsLog) Append(finding SagaFinding) error {
	entries, err := fl.ReadAll()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing findings: %w", err)
	}

	entries = append(entries, finding)

	// Enforce rolling retention.
	if fl.maxEntries > 0 && len(entries) > fl.maxEntries {
		entries = entries[len(entries)-fl.maxEntries:]
	}

	return fl.writeAll(entries)
}

// ReadAll reads all saga findings from the JSONL log.
func (fl *FindingsLog) ReadAll() ([]SagaFinding, error) {
	f, err := os.Open(fl.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return parseFindingsFromReader(f)
}

// parseFindingsFromReader parses saga findings from a reader.
func parseFindingsFromReader(r io.Reader) ([]SagaFinding, error) {
	var findings []SagaFinding
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Only parse saga-specific entries; skip other retrospector entries.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		typeField, ok := raw["type"]
		if !ok {
			continue
		}
		var ft FindingType
		if err := json.Unmarshal(typeField, &ft); err != nil {
			continue
		}
		if ft != FindingSagaDetected && ft != FindingSagaRecur {
			continue
		}

		var finding SagaFinding
		if err := json.Unmarshal([]byte(line), &finding); err != nil {
			return nil, fmt.Errorf("line %d: unmarshal finding: %w", lineNum, err)
		}
		findings = append(findings, finding)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan findings: %w", err)
	}
	return findings, nil
}

// writeAll atomically writes all entries to the JSONL file.
func (fl *FindingsLog) writeAll(entries []SagaFinding) error {
	tmpPath := fl.path + ".tmp"
	cleanup := func() { _ = os.Remove(tmpPath) }

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	for _, entry := range entries {
		if err := enc.Encode(entry); err != nil {
			_ = f.Close()
			cleanup()
			return fmt.Errorf("encode finding: %w", err)
		}
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, fl.path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp to findings: %w", err)
	}

	return nil
}

// FindingFromAlert converts a SagaAlert into a SagaFinding for JSONL logging.
func FindingFromAlert(alert SagaAlert, pr int, repo string) SagaFinding {
	names := make([]string, len(alert.Workers))
	for i, w := range alert.Workers {
		names[i] = w.Name
	}
	return SagaFinding{
		Type:            FindingSagaDetected,
		Branch:          alert.Branch,
		PR:              pr,
		SagaType:        alert.Type,
		WorkerCount:     len(alert.Workers),
		WorkerNames:     names,
		FailureRelation: alert.FailureRelation,
		Recommendation:  alert.Recommendations,
		Timestamp:       alert.Timestamp,
		Repo:            repo,
	}
}
