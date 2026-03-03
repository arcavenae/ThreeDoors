package calendar

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ICSReader parses .ics (iCalendar) files from configured paths.
type ICSReader struct {
	paths []string
}

// NewICSReader creates an ICSReader for the given file or directory paths.
// Directories are scanned for .ics files. Individual .ics files are read directly.
func NewICSReader(paths []string) *ICSReader {
	return &ICSReader{paths: paths}
}

// GetEvents reads calendar events from all configured .ics paths within the time range.
func (r *ICSReader) GetEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
	var allEvents []CalendarEvent

	for _, p := range r.paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		events, err := r.readPath(ctx, p, start, end)
		if err != nil {
			return nil, fmt.Errorf("ics reader path %s: %w", p, err)
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

func (r *ICSReader) readPath(ctx context.Context, path string, start, end time.Time) ([]CalendarEvent, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	if info.IsDir() {
		return r.readDirectory(ctx, path, start, end)
	}
	return r.readFile(ctx, path, start, end)
}

func (r *ICSReader) readDirectory(ctx context.Context, dir string, start, end time.Time) ([]CalendarEvent, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var allEvents []CalendarEvent
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".ics") {
			continue
		}
		events, err := r.readFile(ctx, filepath.Join(dir, entry.Name()), start, end)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)
	}
	return allEvents, nil
}

func (r *ICSReader) readFile(_ context.Context, path string, start, end time.Time) ([]CalendarEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return parseICS(bufio.NewScanner(f), start, end)
}

// parseICS extracts VEVENT components from iCalendar data, filtering by time range.
func parseICS(scanner *bufio.Scanner, start, end time.Time) ([]CalendarEvent, error) {
	lines, err := unfoldLines(scanner)
	if err != nil {
		return nil, fmt.Errorf("scan ics: %w", err)
	}

	var events []CalendarEvent
	var current *icsEvent
	inEvent := false

	for _, line := range lines {
		if line == "BEGIN:VEVENT" {
			current = &icsEvent{}
			inEvent = true
			continue
		}

		if line == "END:VEVENT" && inEvent {
			if current != nil {
				ev, err := current.toCalendarEvent()
				if err == nil && eventInRange(ev, start, end) {
					events = append(events, ev)
				}
			}
			current = nil
			inEvent = false
			continue
		}

		if inEvent && current != nil {
			current.parseLine(line)
		}
	}

	return events, nil
}

// unfoldLines reads all lines from the scanner and handles RFC 5545 line unfolding.
// Continuation lines (starting with space or tab) are appended to the previous line.
func unfoldLines(scanner *bufio.Scanner) ([]string, error) {
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && len(lines) > 0 {
			lines[len(lines)-1] += line[1:]
		} else {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

type icsEvent struct {
	summary  string
	dtStart  string
	dtEnd    string
	calendar string
}

func (e *icsEvent) parseLine(line string) {
	key, value := splitICSProperty(line)
	switch {
	case key == "SUMMARY":
		e.summary = unescapeICSText(value)
	case strings.HasPrefix(key, "DTSTART"):
		e.dtStart = value
	case strings.HasPrefix(key, "DTEND"):
		e.dtEnd = value
	case key == "X-WR-CALNAME":
		e.calendar = unescapeICSText(value)
	}
}

func (e *icsEvent) toCalendarEvent() (CalendarEvent, error) {
	startTime, allDay, err := parseICSDateTime(e.dtStart)
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("parse dtstart: %w", err)
	}

	var endTime time.Time
	if e.dtEnd != "" {
		endTime, _, err = parseICSDateTime(e.dtEnd)
		if err != nil {
			return CalendarEvent{}, fmt.Errorf("parse dtend: %w", err)
		}
	} else if allDay {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endTime = startTime.Add(time.Hour)
	}

	return CalendarEvent{
		Title:    e.summary,
		Start:    startTime,
		End:      endTime,
		AllDay:   allDay,
		Calendar: e.calendar,
	}, nil
}

// splitICSProperty splits "KEY;PARAMS:VALUE" into key (with params) and value.
func splitICSProperty(line string) (string, string) {
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return line, ""
	}
	key := line[:colonIdx]
	value := line[colonIdx+1:]

	// Strip parameters from key for matching (e.g., "DTSTART;VALUE=DATE" -> "DTSTART").
	semiIdx := strings.Index(key, ";")
	baseKey := key
	if semiIdx >= 0 {
		baseKey = key[:semiIdx]
	}
	_ = baseKey // We return the full key with params for datetime parsing.

	return key, value
}

// parseICSDateTime parses iCalendar date/datetime values.
// Returns the parsed time, whether it's an all-day event, and any error.
func parseICSDateTime(raw string) (time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false, fmt.Errorf("empty datetime")
	}

	// All-day: DATE format (8 digits).
	if len(raw) == 8 {
		t, err := time.Parse("20060102", raw)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("parse date %q: %w", raw, err)
		}
		return t.UTC(), true, nil
	}

	// UTC datetime: ends with Z.
	if strings.HasSuffix(raw, "Z") {
		t, err := time.Parse("20060102T150405Z", raw)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("parse utc datetime %q: %w", raw, err)
		}
		return t.UTC(), false, nil
	}

	// Local datetime (no Z, no TZID handling — treat as local).
	t, err := time.ParseInLocation("20060102T150405", raw, time.Local)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parse local datetime %q: %w", raw, err)
	}
	return t.UTC(), false, nil
}

// unescapeICSText handles iCalendar text escaping per RFC 5545.
func unescapeICSText(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\N`, "\n")
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\,`, ",")
	s = strings.ReplaceAll(s, `\;`, ";")
	return s
}

func eventInRange(ev CalendarEvent, start, end time.Time) bool {
	return ev.End.After(start) && ev.Start.Before(end)
}
