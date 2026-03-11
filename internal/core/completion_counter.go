package core

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

// CompletionCounter tracks daily task completion counts parsed from completed.txt.
// It provides today/yesterday counts, streak calculation, and formatted display messages.
//
// Performance note: currently reads the entire completed.txt on load. If the file
// grows to 10k+ lines, consider reading only the last N lines or indexing by date.
type CompletionCounter struct {
	dateCounts map[string]int // "2006-01-02" -> count
	nowFunc    func() time.Time
}

// NewCompletionCounter creates a CompletionCounter using the system clock.
func NewCompletionCounter() *CompletionCounter {
	return &CompletionCounter{
		dateCounts: make(map[string]int),
		nowFunc:    time.Now,
	}
}

// NewCompletionCounterWithNow creates a CompletionCounter with an injected time function for testing.
func NewCompletionCounterWithNow(nowFunc func() time.Time) *CompletionCounter {
	return &CompletionCounter{
		dateCounts: make(map[string]int),
		nowFunc:    nowFunc,
	}
}

// LoadFromFile parses completed.txt and groups completions by date.
// Format: [YYYY-MM-DD HH:MM:SS] uuid | text
// Returns nil for non-existent files (new user). Returns error for permission issues.
// Malformed lines are silently skipped.
func (cc *CompletionCounter) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cc.dateCounts = make(map[string]int)
			return nil
		}
		return fmt.Errorf("failed to open completed file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on read-only file

	counts := make(map[string]int)
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		date := parseDateFromLine(line)
		if date != "" {
			counts[date]++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading completed file: %w", err)
	}

	cc.dateCounts = counts
	return nil
}

// parseDateFromLine extracts the YYYY-MM-DD date from a completed.txt line.
// Expected format: [YYYY-MM-DD HH:MM:SS] uuid | text
// Returns empty string for malformed lines.
func parseDateFromLine(line string) string {
	if len(line) < 12 || line[0] != '[' {
		return ""
	}
	dateStr := line[1:11] // extract "YYYY-MM-DD"
	// Validate it's a real date
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ""
	}
	return dateStr
}

// GetTodayCount returns the number of completions for today (UTC).
func (cc *CompletionCounter) GetTodayCount() int {
	today := cc.nowFunc().UTC().Format("2006-01-02")
	return cc.dateCounts[today]
}

// GetYesterdayCount returns the number of completions for yesterday (UTC).
func (cc *CompletionCounter) GetYesterdayCount() int {
	yesterday := cc.nowFunc().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	return cc.dateCounts[yesterday]
}

// IncrementToday adds one to today's completion count in memory.
// Called after a task is completed in the current session to avoid re-reading the file.
func (cc *CompletionCounter) IncrementToday() {
	today := cc.nowFunc().UTC().Format("2006-01-02")
	cc.dateCounts[today]++
}

// GetStreak returns the number of consecutive days with at least one completion,
// walking backward from today (if today has completions) or yesterday (if not).
// Returns 0 if there are no recent completions.
func (cc *CompletionCounter) GetStreak() int {
	if len(cc.dateCounts) == 0 {
		return 0
	}

	now := cc.nowFunc().UTC()
	today := now.Format("2006-01-02")

	// Start from today if it has completions, otherwise from yesterday
	startDate := now
	if cc.dateCounts[today] == 0 {
		startDate = now.AddDate(0, 0, -1)
		yesterday := startDate.Format("2006-01-02")
		if cc.dateCounts[yesterday] == 0 {
			return 0
		}
	}

	streak := 0
	for {
		dateStr := startDate.AddDate(0, 0, -streak).Format("2006-01-02")
		if cc.dateCounts[dateStr] > 0 {
			streak++
		} else {
			break
		}
	}
	return streak
}

// FormatCompletionMessage returns a display string for the daily completion comparison.
// Rules:
//   - If today = 0: returns "" (empty)
//   - If today > 0, yesterday = 0: returns "Completed today: X"
//   - If today > yesterday: returns "Completed today: X (yesterday: Y) - [positive message]"
//   - If today <= yesterday: returns "Completed today: X (yesterday: Y)"
func (cc *CompletionCounter) FormatCompletionMessage() string {
	todayCount := cc.GetTodayCount()
	if todayCount == 0 {
		return ""
	}

	yesterdayCount := cc.GetYesterdayCount()
	if yesterdayCount == 0 {
		return fmt.Sprintf("Completed today: %d", todayCount)
	}

	msg := fmt.Sprintf("Completed today: %d (yesterday: %d)", todayCount, yesterdayCount)
	if todayCount > yesterdayCount {
		msg += " - " + positiveMessage()
	}
	return msg
}

// positiveMessage returns a randomly selected positive reinforcement message.
func positiveMessage() string {
	messages := []string{
		"You're ahead of yesterday!",
		"Beating yesterday's pace!",
		"On a roll!",
		"Momentum building!",
	}
	return messages[rand.IntN(len(messages))]
}
