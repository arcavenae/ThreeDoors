package retrospector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	// MaxActiveEntries is the threshold at which older entries are archived.
	MaxActiveEntries = 1000
	// RetainEntries is the number of entries kept in the active log after archival.
	RetainEntries = 500
)

// FindingsLog manages the JSONL findings log with rolling archival.
type FindingsLog struct {
	logPath    string
	archiveDir string
}

// NewFindingsLog creates a FindingsLog for the given base directory.
// The active log is stored at baseDir/retrospector-findings.jsonl.
// Archives are stored at baseDir/archive/retrospector-findings-YYYY-MM.jsonl.
func NewFindingsLog(baseDir string) *FindingsLog {
	return &FindingsLog{
		logPath:    filepath.Join(baseDir, "retrospector-findings.jsonl"),
		archiveDir: filepath.Join(baseDir, "archive"),
	}
}

// Append adds a finding to the active log and triggers archival if needed.
func (fl *FindingsLog) Append(f Finding) error {
	if err := fl.appendEntry(f); err != nil {
		return fmt.Errorf("append finding PR #%d: %w", f.PR, err)
	}

	count, err := fl.countEntries()
	if err != nil {
		return fmt.Errorf("count entries: %w", err)
	}

	if count > MaxActiveEntries {
		if err := fl.archive(); err != nil {
			return fmt.Errorf("archive findings: %w", err)
		}
	}

	return nil
}

// ReadAll reads all findings from the active log.
func (fl *FindingsLog) ReadAll() ([]Finding, error) {
	f, err := os.Open(fl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open findings log: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only

	return scanFindings(f)
}

// Path returns the path to the active findings log.
func (fl *FindingsLog) Path() string {
	return fl.logPath
}

func (fl *FindingsLog) appendEntry(f Finding) error {
	file, err := os.OpenFile(fl.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open log for append: %w", err)
	}
	defer file.Close() //nolint:errcheck // best-effort close on append

	data, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal finding: %w", err)
	}

	_, err = file.Write(append(data, '\n'))
	return err
}

func (fl *FindingsLog) countEntries() (int, error) {
	f, err := os.Open(fl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close() //nolint:errcheck // read-only

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

// archive moves older entries to a dated archive file and retains only the
// most recent RetainEntries in the active log. Uses atomic write (tmp+rename).
func (fl *FindingsLog) archive() error {
	findings, err := fl.ReadAll()
	if err != nil {
		return err
	}

	if len(findings) <= RetainEntries {
		return nil
	}

	// Sort by timestamp ascending to ensure chronological order
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Timestamp.Before(findings[j].Timestamp)
	})

	// Split: archive the older entries, retain the newer ones
	archiveEntries := findings[:len(findings)-RetainEntries]
	retainEntries := findings[len(findings)-RetainEntries:]

	// Write archive entries
	if err := os.MkdirAll(fl.archiveDir, 0o700); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	archiveMonth := time.Now().UTC().Format("2006-01")
	archivePath := filepath.Join(fl.archiveDir, fmt.Sprintf("retrospector-findings-%s.jsonl", archiveMonth))

	archiveFile, err := os.OpenFile(archivePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open archive file: %w", err)
	}
	defer archiveFile.Close() //nolint:errcheck // best-effort

	for _, entry := range archiveEntries {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("marshal archive entry: %w", err)
		}
		if _, err := archiveFile.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write archive entry: %w", err)
		}
	}

	// Atomic write: retained entries to tmp, then rename
	tmpPath := fl.logPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create tmp log: %w", err)
	}

	for _, entry := range retainEntries {
		data, err := json.Marshal(entry)
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("marshal retained entry: %w", err)
		}
		if _, err := tmpFile.Write(append(data, '\n')); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("write retained entry: %w", err)
		}
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync tmp log: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close tmp log: %w", err)
	}

	if err := os.Rename(tmpPath, fl.logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename tmp to active log: %w", err)
	}

	return nil
}

func scanFindings(f *os.File) ([]Finding, error) {
	var findings []Finding
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var finding Finding
		if err := json.Unmarshal(line, &finding); err != nil {
			return nil, fmt.Errorf("unmarshal finding: %w", err)
		}
		findings = append(findings, finding)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan findings: %w", err)
	}

	return findings, nil
}
