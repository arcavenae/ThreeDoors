package calendar

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetFreeBlocks(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		events     []CalendarEvent
		rangeStart time.Time
		rangeEnd   time.Time
		minBlock   time.Duration
		wantCount  int
		wantFirst  time.Time
	}{
		{
			name:       "no events gives one big block",
			events:     nil,
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   0,
			wantCount:  1,
			wantFirst:  base,
		},
		{
			name: "single event creates two blocks",
			events: []CalendarEvent{
				{Title: "Meeting", Start: base.Add(2 * time.Hour), End: base.Add(3 * time.Hour)},
			},
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   0,
			wantCount:  2,
			wantFirst:  base,
		},
		{
			name: "all-day events are ignored",
			events: []CalendarEvent{
				{Title: "Holiday", Start: base, End: end, AllDay: true},
			},
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   0,
			wantCount:  1,
			wantFirst:  base,
		},
		{
			name: "overlapping events are merged",
			events: []CalendarEvent{
				{Title: "A", Start: base.Add(1 * time.Hour), End: base.Add(3 * time.Hour)},
				{Title: "B", Start: base.Add(2 * time.Hour), End: base.Add(4 * time.Hour)},
			},
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   0,
			wantCount:  2,
			wantFirst:  base,
		},
		{
			name: "min block duration filters short gaps",
			events: []CalendarEvent{
				{Title: "A", Start: base.Add(10 * time.Minute), End: base.Add(20 * time.Minute)},
				{Title: "B", Start: base.Add(25 * time.Minute), End: base.Add(9 * time.Hour)},
			},
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   30 * time.Minute,
			wantCount:  1,
			wantFirst:  base.Add(9 * time.Hour), // trailing block after B
		},
		{
			name: "events outside range are excluded",
			events: []CalendarEvent{
				{Title: "Yesterday", Start: base.Add(-24 * time.Hour), End: base.Add(-23 * time.Hour)},
				{Title: "Tomorrow", Start: end.Add(time.Hour), End: end.Add(2 * time.Hour)},
			},
			rangeStart: base,
			rangeEnd:   end,
			minBlock:   0,
			wantCount:  1,
			wantFirst:  base,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blocks := GetFreeBlocks(tt.events, tt.rangeStart, tt.rangeEnd, tt.minBlock)
			if len(blocks) != tt.wantCount {
				t.Errorf("got %d blocks, want %d", len(blocks), tt.wantCount)
				for i, b := range blocks {
					t.Logf("  block[%d]: %s - %s (%s)", i, b.Start, b.End, b.Duration)
				}
			}
			if tt.wantCount > 0 && len(blocks) > 0 && !blocks[0].Start.Equal(tt.wantFirst) {
				t.Errorf("first block starts at %s, want %s", blocks[0].Start, tt.wantFirst)
			}
			// Verify duration consistency.
			for i, b := range blocks {
				expected := b.End.Sub(b.Start)
				if b.Duration != expected {
					t.Errorf("block[%d] duration %s != end-start %s", i, b.Duration, expected)
				}
			}
		})
	}
}

func TestParseICS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		start     time.Time
		end       time.Time
		wantCount int
		wantTitle string
	}{
		{
			name: "basic event",
			input: `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Test Meeting
DTSTART:20260302T090000Z
DTEND:20260302T100000Z
END:VEVENT
END:VCALENDAR`,
			start:     time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
			wantTitle: "Test Meeting",
		},
		{
			name: "all day event",
			input: `BEGIN:VEVENT
SUMMARY:Conference
DTSTART;VALUE=DATE:20260302
DTEND;VALUE=DATE:20260303
END:VEVENT`,
			start:     time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
			wantTitle: "Conference",
		},
		{
			name: "event outside range is excluded",
			input: `BEGIN:VEVENT
SUMMARY:Past Event
DTSTART:20260301T090000Z
DTEND:20260301T100000Z
END:VEVENT`,
			start:     time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			wantCount: 0,
		},
		{
			name:      "empty calendar",
			input:     `BEGIN:VCALENDAR\nEND:VCALENDAR`,
			start:     time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			wantCount: 0,
		},
		{
			name: "escaped text in summary",
			input: `BEGIN:VEVENT
SUMMARY:Meeting with "Bob" & Alice\, plus others
DTSTART:20260302T140000Z
DTEND:20260302T150000Z
END:VEVENT`,
			start:     time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
			wantTitle: "Meeting with \"Bob\" & Alice, plus others",
		},
		{
			name: "no DTEND defaults to 1 hour",
			input: `BEGIN:VEVENT
SUMMARY:Quick Chat
DTSTART:20260302T090000Z
END:VEVENT`,
			start:     time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
			wantTitle: "Quick Chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			events, err := parseICS(scanner, tt.start, tt.end)
			if err != nil {
				t.Fatalf("parseICS error: %v", err)
			}
			if len(events) != tt.wantCount {
				t.Fatalf("got %d events, want %d", len(events), tt.wantCount)
			}
			if tt.wantTitle != "" && len(events) > 0 && events[0].Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", events[0].Title, tt.wantTitle)
			}
		})
	}
}

func TestParseICSDateTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantDay bool
		wantErr bool
	}{
		{"utc datetime", "20260302T090000Z", false, false},
		{"date only (all day)", "20260302", true, false},
		{"local datetime", "20260302T090000", false, false},
		{"empty string", "", false, true},
		{"invalid format", "not-a-date", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, allDay, err := parseICSDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && allDay != tt.wantDay {
				t.Errorf("allDay = %v, want %v", allDay, tt.wantDay)
			}
		})
	}
}

func TestUnescapeICSText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"escaped comma", `hello\, world`, "hello, world"},
		{"escaped semicolon", `a\;b`, "a;b"},
		{"escaped newline lowercase", `line1\nline2`, "line1\nline2"},
		{"escaped newline uppercase", `line1\Nline2`, "line1\nline2"},
		{"escaped backslash", `path\\file`, `path\file`},
		{"no escaping needed", "plain text", "plain text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := unescapeICSText(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestICSReaderWithFile(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		fixture   string
		wantCount int
	}{
		{"basic.ics", "testdata/basic.ics", 3},
		{"special_chars.ics", "testdata/special_chars.ics", 2},
		{"empty.ics", "testdata/empty.ics", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := NewICSReader([]string{tt.fixture})
			events, err := reader.GetEvents(context.Background(), start, end)
			if err != nil {
				t.Fatalf("GetEvents error: %v", err)
			}
			if len(events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(events), tt.wantCount)
				for i, e := range events {
					t.Logf("  event[%d]: %s (%s - %s, allDay=%v)", i, e.Title, e.Start, e.End, e.AllDay)
				}
			}
		})
	}
}

func TestICSReaderDirectory(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	reader := NewICSReader([]string{"testdata"})
	events, err := reader.GetEvents(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetEvents error: %v", err)
	}
	// basic.ics (3) + special_chars.ics (2) + empty.ics (0) = 5
	if len(events) != 5 {
		t.Errorf("got %d events from directory, want 5", len(events))
	}
}

func TestICSReaderNonexistentPath(t *testing.T) {
	t.Parallel()

	reader := NewICSReader([]string{"/nonexistent/path.ics"})
	_, err := reader.GetEvents(context.Background(), time.Now().UTC(), time.Now().UTC().Add(time.Hour))
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestICSReaderContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reader := NewICSReader([]string{"testdata/basic.ics", "testdata/special_chars.ics"})
	_, err := reader.GetEvents(ctx, time.Now().UTC(), time.Now().UTC().Add(time.Hour))
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestAppleScriptReaderParseOutput(t *testing.T) {
	t.Parallel()

	reader := NewAppleScriptReaderWithExecutor(func(_ context.Context, _ string) (string, error) {
		return "", nil
	})

	tests := []struct {
		name      string
		output    string
		wantCount int
		wantTitle string
	}{
		{
			name:      "empty output",
			output:    "",
			wantCount: 0,
		},
		{
			name:      "single event",
			output:    "Team Standup\tJanuary 2, 2026 at 9:00:00 AM\tJanuary 2, 2026 at 9:15:00 AM\tfalse\tWork",
			wantCount: 1,
			wantTitle: "Team Standup",
		},
		{
			name:      "multiple events",
			output:    "Meeting A\tJanuary 2, 2026 at 9:00:00 AM\tJanuary 2, 2026 at 10:00:00 AM\tfalse\tWork\nMeeting B\tJanuary 2, 2026 at 11:00:00 AM\tJanuary 2, 2026 at 12:00:00 PM\tfalse\tPersonal",
			wantCount: 2,
		},
		{
			name:      "malformed line is skipped",
			output:    "incomplete\tdata",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			events, err := reader.parseOutput(tt.output)
			if err != nil {
				t.Fatalf("parseOutput error: %v", err)
			}
			if len(events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(events), tt.wantCount)
			}
			if tt.wantTitle != "" && len(events) > 0 && events[0].Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", events[0].Title, tt.wantTitle)
			}
		})
	}
}

func TestAppleScriptReaderGetEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		executor  ScriptExecutor
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful query",
			executor: func(_ context.Context, _ string) (string, error) {
				return "Standup\tJanuary 2, 2026 at 9:00:00 AM\tJanuary 2, 2026 at 9:15:00 AM\tfalse\tWork", nil
			},
			wantCount: 1,
		},
		{
			name: "empty calendar",
			executor: func(_ context.Context, _ string) (string, error) {
				return "", nil
			},
			wantCount: 0,
		},
		{
			name: "executor error",
			executor: func(_ context.Context, _ string) (string, error) {
				return "", errors.New("osascript failed")
			},
			wantErr: true,
		},
		{
			name: "permission denied",
			executor: func(_ context.Context, _ string) (string, error) {
				return "", errors.New("not allowed to send Apple events")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := NewAppleScriptReaderWithExecutor(tt.executor)
			events, err := reader.GetEvents(context.Background(), time.Now().UTC(), time.Now().UTC().Add(24*time.Hour))
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func TestAppleScriptEscaping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello", "hello"},
		{"quotes", `say "hello"`, `say \"hello\"`},
		{"backslash", `path\to\file`, `path\\to\\file`},
		{"both", `"test\path"`, `\"test\\path\"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := escapeAppleScript(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalDAVCacheReader(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	t.Run("nonexistent path returns nil", func(t *testing.T) {
		t.Parallel()
		reader := NewCalDAVCacheReader("/nonexistent/path")
		events, err := reader.GetEvents(context.Background(), start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("got %d events, want 0", len(events))
		}
	})

	t.Run("reads ics files from directory tree", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		calDir := filepath.Join(tmpDir, "cal1.caldav", "events")
		if err := os.MkdirAll(calDir, 0o755); err != nil {
			t.Fatal(err)
		}

		icsContent := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Cached Event
DTSTART:20260302T100000Z
DTEND:20260302T110000Z
END:VEVENT
END:VCALENDAR`
		if err := os.WriteFile(filepath.Join(calDir, "event1.ics"), []byte(icsContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Non-ics file should be ignored.
		if err := os.WriteFile(filepath.Join(calDir, "info.plist"), []byte("not ics"), 0o644); err != nil {
			t.Fatal(err)
		}

		reader := NewCalDAVCacheReader(tmpDir)
		events, err := reader.GetEvents(context.Background(), start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("got %d events, want 1", len(events))
		}
		if len(events) > 0 && events[0].Title != "Cached Event" {
			t.Errorf("title = %q, want %q", events[0].Title, "Cached Event")
		}
	})
}

func TestMultiSourceReader(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	t.Run("aggregates from multiple sources", func(t *testing.T) {
		t.Parallel()
		readers := map[string]CalendarReader{
			"source1": &mockReader{events: []CalendarEvent{
				{Title: "Event A", Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)},
			}},
			"source2": &mockReader{events: []CalendarEvent{
				{Title: "Event B", Start: start.Add(3 * time.Hour), End: start.Add(4 * time.Hour)},
			}},
		}
		multi := NewMultiSourceReader(readers)
		events, err := multi.GetEvents(context.Background(), start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("got %d events, want 2", len(events))
		}
	})

	t.Run("graceful fallback on partial failure", func(t *testing.T) {
		t.Parallel()
		readers := map[string]CalendarReader{
			"good":   &mockReader{events: []CalendarEvent{{Title: "OK"}}},
			"broken": &mockReader{err: errors.New("source unavailable")},
		}
		multi := NewMultiSourceReader(readers)
		events, err := multi.GetEvents(context.Background(), start, end)
		if err != nil {
			t.Fatalf("should not fail when some sources work: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("got %d events, want 1", len(events))
		}
	})

	t.Run("error when all sources fail", func(t *testing.T) {
		t.Parallel()
		readers := map[string]CalendarReader{
			"broken1": &mockReader{err: errors.New("fail1")},
			"broken2": &mockReader{err: errors.New("fail2")},
		}
		multi := NewMultiSourceReader(readers)
		_, err := multi.GetEvents(context.Background(), start, end)
		if err == nil {
			t.Error("expected error when all sources fail")
		}
	})

	t.Run("events sorted by start time", func(t *testing.T) {
		t.Parallel()
		readers := map[string]CalendarReader{
			"late":  &mockReader{events: []CalendarEvent{{Title: "Late", Start: start.Add(5 * time.Hour), End: start.Add(6 * time.Hour)}}},
			"early": &mockReader{events: []CalendarEvent{{Title: "Early", Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)}}},
		}
		multi := NewMultiSourceReader(readers)
		events, err := multi.GetEvents(context.Background(), start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}
		if events[0].Title != "Early" {
			t.Errorf("first event = %q, want %q", events[0].Title, "Early")
		}
	})
}

func TestNewReadersFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *Config
		wantNi bool
	}{
		{"nil config", nil, true},
		{"disabled", &Config{Enabled: false}, true},
		{"no sources", &Config{Enabled: true}, true},
		{"with ics source", &Config{
			Enabled: true,
			Sources: []SourceConfig{{Type: SourceICS, Path: "/tmp/test.ics"}},
		}, false},
		{"with applescript source", &Config{
			Enabled: true,
			Sources: []SourceConfig{{Type: SourceAppleScript}},
		}, false},
		{"with caldav source", &Config{
			Enabled: true,
			Sources: []SourceConfig{{Type: SourceCalDAVCache}},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := NewReadersFromConfig(tt.config)
			if (reader == nil) != tt.wantNi {
				t.Errorf("reader nil = %v, want %v", reader == nil, tt.wantNi)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"empty config", Config{}, false},
		{"valid applescript", Config{Sources: []SourceConfig{{Type: SourceAppleScript}}}, false},
		{"valid ics with path", Config{Sources: []SourceConfig{{Type: SourceICS, Path: "/tmp/cal.ics"}}}, false},
		{"ics without path", Config{Sources: []SourceConfig{{Type: SourceICS}}}, true},
		{"valid caldav", Config{Sources: []SourceConfig{{Type: SourceCalDAVCache}}}, false},
		{"unknown type", Config{Sources: []SourceConfig{{Type: "unknown"}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent file returns disabled config", func(t *testing.T) {
		t.Parallel()
		cfg, err := LoadConfig("/nonexistent/config.yaml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Enabled {
			t.Error("expected disabled config")
		}
	})

	t.Run("empty file returns disabled config", func(t *testing.T) {
		t.Parallel()
		tmpFile := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(tmpFile, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(tmpFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Enabled {
			t.Error("expected disabled config")
		}
	})

	t.Run("valid config with calendar section", func(t *testing.T) {
		t.Parallel()
		content := `calendar:
  enabled: true
  sources:
    - type: applescript
    - type: ics
      path: /tmp/calendar.ics
`
		tmpFile := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(tmpFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.Enabled {
			t.Error("expected enabled config")
		}
		if len(cfg.Sources) != 2 {
			t.Errorf("got %d sources, want 2", len(cfg.Sources))
		}
	})

	t.Run("config without calendar section", func(t *testing.T) {
		t.Parallel()
		content := `provider: textfile
note_title: My Tasks
`
		tmpFile := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(tmpFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Enabled {
			t.Error("expected disabled config")
		}
	})

	t.Run("invalid config fails validation", func(t *testing.T) {
		t.Parallel()
		content := `calendar:
  enabled: true
  sources:
    - type: ics
`
		tmpFile := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(tmpFile)
		if err == nil {
			t.Error("expected validation error for ics without path")
		}
	})
}

// mockReader is a test helper implementing CalendarReader.
type mockReader struct {
	events []CalendarEvent
	err    error
}

func (m *mockReader) GetEvents(_ context.Context, _, _ time.Time) ([]CalendarEvent, error) {
	return m.events, m.err
}
