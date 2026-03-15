package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// CompletionRecord represents a single completed task entry.
type CompletionRecord struct {
	Title       string
	CompletedAt time.Time
	Source      string
	TaskID      string
}

// CompletionReader reads and aggregates completed task data from completed.txt.
type CompletionReader struct {
	configDir string
	nowFunc   func() time.Time
}

// NewCompletionReader creates a CompletionReader that reads from the given config directory.
func NewCompletionReader(configDir string) *CompletionReader {
	return &CompletionReader{
		configDir: configDir,
		nowFunc:   time.Now,
	}
}

// newCompletionReaderWithNow creates a CompletionReader with an injected time function for testing.
func newCompletionReaderWithNow(configDir string, nowFunc func() time.Time) *CompletionReader {
	return &CompletionReader{
		configDir: configDir,
		nowFunc:   nowFunc,
	}
}

// Read returns all completion records from completed.txt, sorted newest-first.
// Returns an empty slice and nil error if the file doesn't exist.
func (cr *CompletionReader) Read(_ context.Context) ([]CompletionRecord, error) {
	path := cr.configDir + "/completed.txt"
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []CompletionRecord{}, nil
		}
		return nil, fmt.Errorf("open completed file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on read-only file

	var records []CompletionRecord
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rec, ok := parseCompletionLine(line)
		if !ok {
			log.Printf("completion_reader: skipping malformed line: %q", line)
			continue
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read completed file: %w", err)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CompletedAt.After(records[j].CompletedAt)
	})

	return records, nil
}

// Today returns completion records from today in the local timezone.
func (cr *CompletionReader) Today(ctx context.Context) ([]CompletionRecord, error) {
	now := cr.nowFunc()
	local := now.Local()
	startOfDay := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	return cr.Since(ctx, startOfDay)
}

// ThisWeek returns completion records from the current calendar week (Monday-based) in local timezone.
func (cr *CompletionReader) ThisWeek(ctx context.Context) ([]CompletionRecord, error) {
	now := cr.nowFunc()
	local := now.Local()
	weekday := local.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysBack := int(weekday) - int(time.Monday)
	startOfWeek := time.Date(local.Year(), local.Month(), local.Day()-daysBack, 0, 0, 0, 0, local.Location())
	return cr.Since(ctx, startOfWeek)
}

// ThisMonth returns completion records from the current calendar month in local timezone.
func (cr *CompletionReader) ThisMonth(ctx context.Context) ([]CompletionRecord, error) {
	now := cr.nowFunc()
	local := now.Local()
	startOfMonth := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, local.Location())
	return cr.Since(ctx, startOfMonth)
}

// Since returns completion records from the given time onward, sorted newest-first.
func (cr *CompletionReader) Since(ctx context.Context, since time.Time) ([]CompletionRecord, error) {
	all, err := cr.Read(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []CompletionRecord
	for _, rec := range all {
		if !rec.CompletedAt.Before(since) {
			filtered = append(filtered, rec)
		}
	}
	return filtered, nil
}

// parseCompletionLine parses a line from completed.txt into a CompletionRecord.
// Expected format: [YYYY-MM-DD HH:MM:SS] uuid | text
// Returns false if the line is malformed.
func parseCompletionLine(line string) (CompletionRecord, bool) {
	// Minimum: [2006-01-02 15:04:05] x | y
	if len(line) < 23 || line[0] != '[' {
		return CompletionRecord{}, false
	}

	closeBracket := strings.IndexByte(line, ']')
	if closeBracket < 20 {
		return CompletionRecord{}, false
	}

	timestampStr := line[1:closeBracket]
	completedAt, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		return CompletionRecord{}, false
	}

	rest := line[closeBracket+1:]
	rest = strings.TrimLeft(rest, " ")

	pipeIdx := strings.Index(rest, " | ")
	if pipeIdx < 0 {
		return CompletionRecord{}, false
	}

	taskID := strings.TrimSpace(rest[:pipeIdx])
	title := strings.TrimSpace(rest[pipeIdx+3:])

	if title == "" {
		return CompletionRecord{}, false
	}

	return CompletionRecord{
		Title:       title,
		CompletedAt: completedAt.UTC(),
		TaskID:      taskID,
	}, true
}
