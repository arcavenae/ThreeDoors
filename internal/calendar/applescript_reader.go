package calendar

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ScriptExecutor abstracts osascript execution for testability.
type ScriptExecutor func(ctx context.Context, script string) (string, error)

// DefaultExecutor runs osascript on macOS.
func DefaultExecutor(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// AppleScriptReader reads macOS Calendar.app events via AppleScript.
// No OAuth required — reads directly from the local Calendar database.
type AppleScriptReader struct {
	executor ScriptExecutor
	timeout  time.Duration
}

// NewAppleScriptReader creates an AppleScriptReader with the default osascript executor.
func NewAppleScriptReader() *AppleScriptReader {
	return &AppleScriptReader{
		executor: DefaultExecutor,
		timeout:  5 * time.Second,
	}
}

// NewAppleScriptReaderWithExecutor creates an AppleScriptReader with a custom executor for testing.
func NewAppleScriptReaderWithExecutor(executor ScriptExecutor) *AppleScriptReader {
	return &AppleScriptReader{
		executor: executor,
		timeout:  5 * time.Second,
	}
}

// GetEvents retrieves calendar events from Calendar.app within the specified time range.
func (r *AppleScriptReader) GetEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	script := r.buildScript(start, end)
	output, err := r.executor(ctx, script)
	if err != nil {
		return nil, r.wrapError(err)
	}

	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	return r.parseOutput(output)
}

// buildScript generates the AppleScript to query Calendar.app events.
// Uses tab-delimited output: title\tstart\tend\tallDay\tcalendarName
func (r *AppleScriptReader) buildScript(start, end time.Time) string {
	startStr := escapeAppleScript(start.Local().Format("January 2, 2006 3:04:05 PM"))
	endStr := escapeAppleScript(end.Local().Format("January 2, 2006 3:04:05 PM"))

	return fmt.Sprintf(`
set startDate to date "%s"
set endDate to date "%s"
set output to ""
tell application "Calendar"
    repeat with cal in calendars
        set calName to name of cal
        set evts to (every event of cal whose start date >= startDate and start date < endDate)
        repeat with evt in evts
            set evtTitle to summary of evt
            set evtStart to start date of evt
            set evtEnd to end date of evt
            set evtAllDay to allday event of evt
            set output to output & evtTitle & tab & (evtStart as text) & tab & (evtEnd as text) & tab & evtAllDay & tab & calName & linefeed
        end repeat
    end repeat
end tell
return output`, startStr, endStr)
}

// parseOutput parses the tab-delimited AppleScript output into CalendarEvents.
func (r *AppleScriptReader) parseOutput(output string) ([]CalendarEvent, error) {
	var events []CalendarEvent

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			continue
		}

		title := parts[0]
		startTime, err := parseAppleScriptDate(parts[1])
		if err != nil {
			continue
		}
		endTime, err := parseAppleScriptDate(parts[2])
		if err != nil {
			continue
		}
		allDay := strings.ToLower(strings.TrimSpace(parts[3])) == "true"
		calName := parts[4]

		events = append(events, CalendarEvent{
			Title:    title,
			Start:    startTime.UTC(),
			End:      endTime.UTC(),
			AllDay:   allDay,
			Calendar: calName,
		})
	}

	return events, nil
}

// parseAppleScriptDate parses date strings returned by AppleScript.
// AppleScript returns dates in the system's locale format. We try several common formats.
func parseAppleScriptDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	formats := []string{
		"Monday, January 2, 2006 at 3:04:05 PM",
		"January 2, 2006 at 3:04:05 PM",
		"1/2/2006 3:04:05 PM",
		"2006-01-02 15:04:05",
		"Monday, January 2, 2006 at 15:04:05",
		"January 2, 2006 at 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.ParseInLocation(format, s, time.Local)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("parse applescript date %q: unrecognized format", s)
}

// escapeAppleScript escapes a string for embedding inside AppleScript double-quoted strings.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

func (r *AppleScriptReader) wrapError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("calendar applescript: timed out: %w", err)
	}

	msg := err.Error()

	if strings.Contains(msg, "not allowed") || strings.Contains(msg, "Not authorized") {
		return fmt.Errorf("calendar applescript: automation permission denied — grant Calendar access in System Settings > Privacy > Automation: %w", err)
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("calendar applescript: osascript failed: %w", err)
	}

	return fmt.Errorf("calendar applescript: %w", err)
}
