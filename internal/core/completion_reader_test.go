package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCompletionReader(t *testing.T) {
	t.Parallel()
	cr := NewCompletionReader("/tmp/test")
	if cr == nil {
		t.Fatal("NewCompletionReader returned nil")
	}
	if cr.configDir != "/tmp/test" {
		t.Errorf("configDir = %q, want %q", cr.configDir, "/tmp/test")
	}
}

func TestCompletionReader_Read(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fixture   string
		wantCount int
		wantFirst string
		wantLast  string
		wantErr   bool
	}{
		{
			name:      "valid file",
			fixture:   "testdata/completed-valid.txt",
			wantCount: 7,
			wantFirst: "Buy groceries",
			wantLast:  "January task",
		},
		{
			name:      "malformed lines skipped",
			fixture:   "testdata/completed-malformed.txt",
			wantCount: 3,
			wantFirst: "Valid line",
			wantLast:  "Yesterday valid",
		},
		{
			name:      "empty file",
			fixture:   "testdata/completed-empty.txt",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Copy fixture to temp dir so configDir + "/completed.txt" works
			dir := t.TempDir()
			src, err := os.ReadFile(tt.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if err := os.WriteFile(filepath.Join(dir, "completed.txt"), src, 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			cr := NewCompletionReader(dir)
			records, err := cr.Read(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(records) != tt.wantCount {
				t.Fatalf("Read() returned %d records, want %d", len(records), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if records[0].Title != tt.wantFirst {
					t.Errorf("first record title = %q, want %q", records[0].Title, tt.wantFirst)
				}
				if records[len(records)-1].Title != tt.wantLast {
					t.Errorf("last record title = %q, want %q", records[len(records)-1].Title, tt.wantLast)
				}
			}
		})
	}
}

func TestCompletionReader_Read_MissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cr := NewCompletionReader(dir)
	records, err := cr.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error = %v, want nil for missing file", err)
	}
	if len(records) != 0 {
		t.Errorf("Read() returned %d records, want 0 for missing file", len(records))
	}
}

func TestCompletionReader_Read_SortedNewestFirst(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Write records out of order
	content := "[2026-01-01 10:00:00] a | First\n[2026-03-15 10:00:00] b | Third\n[2026-02-01 10:00:00] c | Second\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := NewCompletionReader(dir)
	records, err := cr.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
	if records[0].Title != "Third" {
		t.Errorf("records[0].Title = %q, want %q", records[0].Title, "Third")
	}
	if records[1].Title != "Second" {
		t.Errorf("records[1].Title = %q, want %q", records[1].Title, "Second")
	}
	if records[2].Title != "First" {
		t.Errorf("records[2].Title = %q, want %q", records[2].Title, "First")
	}
}

func TestCompletionReader_Read_FieldParsing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "[2026-03-15 14:30:00] abc-123 | Buy groceries\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := NewCompletionReader(dir)
	records, err := cr.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}

	rec := records[0]
	if rec.Title != "Buy groceries" {
		t.Errorf("Title = %q, want %q", rec.Title, "Buy groceries")
	}
	if rec.TaskID != "abc-123" {
		t.Errorf("TaskID = %q, want %q", rec.TaskID, "abc-123")
	}
	wantTime := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
	if !rec.CompletedAt.Equal(wantTime) {
		t.Errorf("CompletedAt = %v, want %v", rec.CompletedAt, wantTime)
	}
	if rec.Source != "" {
		t.Errorf("Source = %q, want empty", rec.Source)
	}
}

func TestCompletionReader_Today(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// "now" is 2026-03-15 15:00 local time
	now := time.Date(2026, 3, 15, 15, 0, 0, 0, time.Local)

	content := "[2026-03-15 14:30:00] a | Today task 1\n" +
		"[2026-03-15 09:00:00] b | Today task 2\n" +
		"[2026-03-14 16:45:00] c | Yesterday task\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := newCompletionReaderWithNow(dir, func() time.Time { return now })
	records, err := cr.Today(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("Today() returned %d records, want 2", len(records))
	}
	if records[0].Title != "Today task 1" {
		t.Errorf("records[0].Title = %q, want %q", records[0].Title, "Today task 1")
	}
}

func TestCompletionReader_ThisWeek(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// 2026-03-15 is a Sunday. Week started Monday 2026-03-09.
	now := time.Date(2026, 3, 15, 15, 0, 0, 0, time.Local)

	content := "[2026-03-15 10:00:00] a | Sunday task\n" +
		"[2026-03-10 10:00:00] b | Tuesday task\n" +
		"[2026-03-09 08:00:00] c | Monday task\n" +
		"[2026-03-08 10:00:00] d | Last week task\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := newCompletionReaderWithNow(dir, func() time.Time { return now })
	records, err := cr.ThisWeek(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("ThisWeek() returned %d records, want 3", len(records))
	}
}

func TestCompletionReader_ThisMonth(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	now := time.Date(2026, 3, 15, 15, 0, 0, 0, time.Local)

	content := "[2026-03-15 10:00:00] a | This month\n" +
		"[2026-03-01 08:00:00] b | First of month\n" +
		"[2026-02-28 10:00:00] c | Last month\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := newCompletionReaderWithNow(dir, func() time.Time { return now })
	records, err := cr.ThisMonth(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("ThisMonth() returned %d records, want 2", len(records))
	}
}

func TestCompletionReader_Since(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	content := "[2026-03-15 14:30:00] a | Recent\n" +
		"[2026-03-10 10:00:00] b | Mid month\n" +
		"[2026-02-01 10:00:00] c | February\n" +
		"[2026-01-01 10:00:00] d | January\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := NewCompletionReader(dir)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	records, err := cr.Since(context.Background(), since)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("Since() returned %d records, want 2", len(records))
	}
	if records[0].Title != "Recent" {
		t.Errorf("records[0].Title = %q, want %q", records[0].Title, "Recent")
	}
}

func TestCompletionReader_Since_EmptyResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	content := "[2026-01-01 10:00:00] a | Old task\n"
	if err := os.WriteFile(filepath.Join(dir, "completed.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cr := NewCompletionReader(dir)
	since := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	records, err := cr.Since(context.Background(), since)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("Since() returned %d records, want 0", len(records))
	}
}

func TestParseCompletionLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantRec CompletionRecord
	}{
		{
			name:   "valid line",
			line:   "[2026-03-15 14:30:00] abc-123 | Buy groceries",
			wantOK: true,
			wantRec: CompletionRecord{
				Title:       "Buy groceries",
				CompletedAt: time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC),
				TaskID:      "abc-123",
			},
		},
		{
			name:   "empty line",
			line:   "",
			wantOK: false,
		},
		{
			name:   "no bracket",
			line:   "2026-03-15 14:30:00 abc | task",
			wantOK: false,
		},
		{
			name:   "invalid date",
			line:   "[not-a-date 14:30:00] abc | task",
			wantOK: false,
		},
		{
			name:   "no pipe separator",
			line:   "[2026-03-15 14:30:00] abc task",
			wantOK: false,
		},
		{
			name:   "empty title",
			line:   "[2026-03-15 14:30:00] abc |  ",
			wantOK: false,
		},
		{
			name:   "too short",
			line:   "[2026] x | y",
			wantOK: false,
		},
		{
			name:   "title with pipe",
			line:   "[2026-03-15 14:30:00] abc | task | with pipe",
			wantOK: true,
			wantRec: CompletionRecord{
				Title:       "task | with pipe",
				CompletedAt: time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC),
				TaskID:      "abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec, ok := parseCompletionLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("parseCompletionLine(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if rec.Title != tt.wantRec.Title {
				t.Errorf("Title = %q, want %q", rec.Title, tt.wantRec.Title)
			}
			if !rec.CompletedAt.Equal(tt.wantRec.CompletedAt) {
				t.Errorf("CompletedAt = %v, want %v", rec.CompletedAt, tt.wantRec.CompletedAt)
			}
			if rec.TaskID != tt.wantRec.TaskID {
				t.Errorf("TaskID = %q, want %q", rec.TaskID, tt.wantRec.TaskID)
			}
		})
	}
}

func TestCompletionReader_Today_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cr := NewCompletionReader(dir)
	records, err := cr.Today(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("Today() returned %d records for missing file, want 0", len(records))
	}
}

func TestCompletionReader_ThisWeek_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cr := NewCompletionReader(dir)
	records, err := cr.ThisWeek(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("ThisWeek() returned %d records for missing file, want 0", len(records))
	}
}

func TestCompletionReader_ThisMonth_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cr := NewCompletionReader(dir)
	records, err := cr.ThisMonth(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("ThisMonth() returned %d records for missing file, want 0", len(records))
	}
}
