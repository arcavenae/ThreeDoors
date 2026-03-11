package saga

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

// WorkerRecord represents a single worker dispatch observation.
type WorkerRecord struct {
	Name      string    `json:"name"`
	Branch    string    `json:"branch"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// DispatchTracker tracks worker-to-branch mappings over time and detects
// same-branch overlaps within a configurable time window.
type DispatchTracker struct {
	window  time.Duration
	records []WorkerRecord
}

// NewDispatchTracker creates a tracker with the given overlap detection window.
func NewDispatchTracker(window time.Duration) *DispatchTracker {
	return &DispatchTracker{
		window:  window,
		records: nil,
	}
}

// AddRecord records a worker dispatch observation.
func (dt *DispatchTracker) AddRecord(rec WorkerRecord) {
	dt.records = append(dt.records, rec)
}

// Records returns a copy of all tracked records.
func (dt *DispatchTracker) Records() []WorkerRecord {
	out := make([]WorkerRecord, len(dt.records))
	copy(out, dt.records)
	return out
}

// BranchOverlap represents a detected same-branch overlap.
type BranchOverlap struct {
	Branch  string
	Workers []WorkerRecord
}

// DetectOverlaps finds branches where 2+ distinct workers were dispatched
// within the configured time window.
func (dt *DispatchTracker) DetectOverlaps() []BranchOverlap {
	// Group records by branch.
	byBranch := make(map[string][]WorkerRecord)
	for _, r := range dt.records {
		byBranch[r.Branch] = append(byBranch[r.Branch], r)
	}

	var overlaps []BranchOverlap
	for branch, records := range byBranch {
		// Deduplicate by worker name — keep only the latest record per worker.
		latest := make(map[string]WorkerRecord)
		for _, r := range records {
			if prev, ok := latest[r.Name]; !ok || r.Timestamp.After(prev.Timestamp) {
				latest[r.Name] = r
			}
		}

		if len(latest) < 2 {
			continue
		}

		// Check if any pair of distinct workers falls within the window.
		workers := make([]WorkerRecord, 0, len(latest))
		for _, r := range latest {
			workers = append(workers, r)
		}

		inWindow := findWorkersInWindow(workers, dt.window)
		if len(inWindow) >= 2 {
			overlaps = append(overlaps, BranchOverlap{
				Branch:  branch,
				Workers: inWindow,
			})
		}
	}

	return overlaps
}

// findWorkersInWindow returns workers whose timestamps span less than the given window.
func findWorkersInWindow(workers []WorkerRecord, window time.Duration) []WorkerRecord {
	if len(workers) < 2 {
		return nil
	}

	var earliest, latest time.Time
	for i, w := range workers {
		if i == 0 || w.Timestamp.Before(earliest) {
			earliest = w.Timestamp
		}
		if i == 0 || w.Timestamp.After(latest) {
			latest = w.Timestamp
		}
	}

	if latest.Sub(earliest) <= window {
		return workers
	}
	return nil
}

// ParseWorkerList parses the output of `multiclaude worker list` into WorkerRecords.
// Expected format: lines with "name", "branch", and "status" fields separated by whitespace.
// The exact format is: NAME BRANCH STATUS (tab or space separated).
// Lines starting with '#' or empty lines are skipped.
func ParseWorkerList(r io.Reader, now time.Time) ([]WorkerRecord, error) {
	var records []WorkerRecord
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "NAME") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("line %d: expected at least 3 fields, got %d: %q", lineNum, len(fields), line)
		}

		records = append(records, WorkerRecord{
			Name:      fields[0],
			Branch:    fields[1],
			Status:    fields[2],
			Timestamp: now,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan worker list: %w", err)
	}
	return records, nil
}
