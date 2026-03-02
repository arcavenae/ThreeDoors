package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeCompletedFile creates a completed.txt file in the given directory with the specified entries.
// entries maps date strings ("2006-01-02") to slices of task texts.
func writeCompletedFile(t *testing.T, dir string, entries map[string][]string) string {
	t.Helper()
	path := filepath.Join(dir, "completed.txt")
	var lines []string
	for date, texts := range entries {
		for i, text := range texts {
			line := fmt.Sprintf("[%s %02d:00:00] fake-uuid-%s-%d | %s", date, i+9, date, i, text)
			lines = append(lines, line)
		}
	}
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test completed.txt: %v", err)
	}
	return path
}

// frozenTime returns a nowFunc that always returns the given time.
func frozenTime(year int, month time.Month, day, hour int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	}
}

func TestNewCompletionCounter(t *testing.T) {
	cc := NewCompletionCounter()
	if cc == nil {
		t.Fatal("NewCompletionCounter() returned nil")
	}
	if cc.GetTodayCount() != 0 {
		t.Errorf("new counter GetTodayCount() = %d, want 0", cc.GetTodayCount())
	}
	if cc.GetYesterdayCount() != 0 {
		t.Errorf("new counter GetYesterdayCount() = %d, want 0", cc.GetYesterdayCount())
	}
	if cc.GetStreak() != 0 {
		t.Errorf("new counter GetStreak() = %d, want 0", cc.GetStreak())
	}
}

func TestNewCompletionCounterWithNow(t *testing.T) {
	frozen := frozenTime(2026, 3, 2, 14)
	cc := NewCompletionCounterWithNow(frozen)
	if cc == nil {
		t.Fatal("NewCompletionCounterWithNow() returned nil")
	}
	if cc.GetTodayCount() != 0 {
		t.Errorf("new counter GetTodayCount() = %d, want 0", cc.GetTodayCount())
	}
}

func TestLoadFromFile_BasicCounts(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	today := "2026-03-02"
	yesterday := "2026-03-01"

	path := writeCompletedFile(t, dir, map[string][]string{
		today:     {"task A", "task B", "task C"},
		yesterday: {"task D", "task E"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetTodayCount(); got != 3 {
		t.Errorf("GetTodayCount() = %d, want 3", got)
	}
	if got := cc.GetYesterdayCount(); got != 2 {
		t.Errorf("GetYesterdayCount() = %d, want 2", got)
	}
}

func TestLoadFromFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "completed.txt")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	cc := NewCompletionCounter()
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error on empty file: %v", err)
	}

	if got := cc.GetTodayCount(); got != 0 {
		t.Errorf("GetTodayCount() = %d, want 0", got)
	}
	if got := cc.GetYesterdayCount(); got != 0 {
		t.Errorf("GetYesterdayCount() = %d, want 0", got)
	}
	if got := cc.GetStreak(); got != 0 {
		t.Errorf("GetStreak() = %d, want 0", got)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	cc := NewCompletionCounter()
	err := cc.LoadFromFile(path)
	if err != nil {
		t.Errorf("LoadFromFile() should return nil for non-existent file, got: %v", err)
	}
	if got := cc.GetTodayCount(); got != 0 {
		t.Errorf("GetTodayCount() = %d, want 0", got)
	}
	if got := cc.GetStreak(); got != 0 {
		t.Errorf("GetStreak() = %d, want 0", got)
	}
}

func TestLoadFromFile_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "completed.txt")
	content := `[2026-03-02 10:00:00] uuid-1 | Valid task
this is garbage
[bad-date 10:00:00] uuid-2 | Bad date
[2026-03-02 11:00:00] uuid-3 | Another valid task
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	frozen := frozenTime(2026, 3, 2, 14)
	cc := NewCompletionCounterWithNow(frozen)
	err := cc.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() should not error on malformed lines, got: %v", err)
	}

	// Only the 2 valid lines should be counted
	if got := cc.GetTodayCount(); got != 2 {
		t.Errorf("GetTodayCount() = %d, want 2 (only valid lines)", got)
	}
}

func TestLoadFromFile_PermissionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "completed.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	// Remove read permission
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("cannot change permissions on this OS: %v", err)
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	cc := NewCompletionCounter()
	err := cc.LoadFromFile(path)
	if err == nil {
		t.Error("LoadFromFile() should return error for permission-denied file")
	}
}

func TestGetStreak_ConsecutiveDays(t *testing.T) {
	dir := t.TempDir()
	// 5 consecutive days ending today (2026-03-02)
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-02-26": {"task"},
		"2026-02-27": {"task"},
		"2026-02-28": {"task"},
		"2026-03-01": {"task"},
		"2026-03-02": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetStreak(); got != 5 {
		t.Errorf("GetStreak() = %d, want 5", got)
	}
}

func TestGetStreak_GapBreaksStreak(t *testing.T) {
	dir := t.TempDir()
	// 3 days, gap, then 2 more days ending today
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-02-25": {"task"},
		"2026-02-26": {"task"},
		"2026-02-27": {"task"},
		// gap on 2026-02-28
		"2026-03-01": {"task"},
		"2026-03-02": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetStreak(); got != 2 {
		t.Errorf("GetStreak() = %d, want 2 (only current streak)", got)
	}
}

func TestGetStreak_NoCompletionsToday_StreakFromYesterday(t *testing.T) {
	dir := t.TempDir()
	// No completions today, 3 consecutive days ending yesterday
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-02-28": {"task"},
		"2026-03-01": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	// Streak of 2 (Feb 28 + Mar 1), walking back from yesterday since today has 0
	if got := cc.GetStreak(); got != 2 {
		t.Errorf("GetStreak() = %d, want 2 (streak up to yesterday)", got)
	}
}

func TestGetStreak_NoData(t *testing.T) {
	cc := NewCompletionCounter()
	if got := cc.GetStreak(); got != 0 {
		t.Errorf("GetStreak() = %d, want 0", got)
	}
}

func TestGetStreak_OnlyToday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetStreak(); got != 1 {
		t.Errorf("GetStreak() = %d, want 1", got)
	}
}

func TestGetStreak_OldCompletionsNoRecent(t *testing.T) {
	dir := t.TempDir()
	// Completions 5 days ago, nothing since
	frozen := frozenTime(2026, 3, 5, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-01": {"task"},
		"2026-03-02": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetStreak(); got != 0 {
		t.Errorf("GetStreak() = %d, want 0 (too old)", got)
	}
}

func TestIncrementToday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"task A", "task B"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetTodayCount(); got != 2 {
		t.Errorf("GetTodayCount() before increment = %d, want 2", got)
	}

	cc.IncrementToday()

	if got := cc.GetTodayCount(); got != 3 {
		t.Errorf("GetTodayCount() after increment = %d, want 3", got)
	}
}

func TestIncrementToday_FromZero(t *testing.T) {
	frozen := frozenTime(2026, 3, 2, 14)
	cc := NewCompletionCounterWithNow(frozen)

	cc.IncrementToday()

	if got := cc.GetTodayCount(); got != 1 {
		t.Errorf("GetTodayCount() after first increment = %d, want 1", got)
	}

	// Streak should now be 1 since we incremented today
	if got := cc.GetStreak(); got != 1 {
		t.Errorf("GetStreak() after first increment = %d, want 1", got)
	}
}

func TestFormatCompletionMessage_TodayBeatsYesterday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c"},
		"2026-03-01": {"d", "e"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	msg := cc.FormatCompletionMessage()

	// Should contain the counts
	if !strings.Contains(msg, "3") {
		t.Errorf("message should contain today count 3, got: %q", msg)
	}
	if !strings.Contains(msg, "2") {
		t.Errorf("message should contain yesterday count 2, got: %q", msg)
	}
	// Should contain "yesterday" text
	if !strings.Contains(msg, "yesterday") {
		t.Errorf("message should contain 'yesterday', got: %q", msg)
	}
}

func TestFormatCompletionMessage_TodayBehindYesterday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
		"2026-03-01": {"c", "d", "e", "f", "g"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	msg := cc.FormatCompletionMessage()

	// Should contain counts but NO positive reinforcement message
	if !strings.Contains(msg, "2") {
		t.Errorf("message should contain today count 2, got: %q", msg)
	}
	if !strings.Contains(msg, "5") {
		t.Errorf("message should contain yesterday count 5, got: %q", msg)
	}
	if !strings.Contains(msg, "yesterday") {
		t.Errorf("message should contain 'yesterday', got: %q", msg)
	}
}

func TestFormatCompletionMessage_NoYesterdayData(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	msg := cc.FormatCompletionMessage()

	// Should have today count but NOT mention yesterday
	if !strings.Contains(msg, "1") {
		t.Errorf("message should contain today count 1, got: %q", msg)
	}
	if strings.Contains(msg, "yesterday") {
		t.Errorf("message should NOT contain 'yesterday' when no yesterday data, got: %q", msg)
	}
}

func TestFormatCompletionMessage_ZeroToday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-01": {"a", "b"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	msg := cc.FormatCompletionMessage()

	// Should return empty string when today = 0
	if msg != "" {
		t.Errorf("FormatCompletionMessage() should return empty string when today=0, got: %q", msg)
	}
}

func TestFormatCompletionMessage_TodayEqualsYesterday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
		"2026-03-01": {"c", "d"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	msg := cc.FormatCompletionMessage()

	// Should contain counts and mention yesterday, but NO positive reinforcement
	// (today is not strictly greater than yesterday)
	if !strings.Contains(msg, "2") {
		t.Errorf("message should contain count 2, got: %q", msg)
	}
	if !strings.Contains(msg, "yesterday") {
		t.Errorf("message should contain 'yesterday', got: %q", msg)
	}
}

func TestGetStreak_MidnightBoundary(t *testing.T) {
	dir := t.TempDir()
	// Frozen at just after midnight on March 2 — today is March 2
	frozen := frozenTime(2026, 3, 2, 0)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-01": {"task"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	// Yesterday (Mar 1) had completions, today (Mar 2) does not yet
	// Streak should count from yesterday backward = 1
	if got := cc.GetStreak(); got != 1 {
		t.Errorf("GetStreak() at midnight boundary = %d, want 1", got)
	}
}

func TestMultipleCompletionsPerDay(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)
	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c", "d", "e"},
		"2026-03-01": {"f"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetTodayCount(); got != 5 {
		t.Errorf("GetTodayCount() = %d, want 5", got)
	}
	if got := cc.GetYesterdayCount(); got != 1 {
		t.Errorf("GetYesterdayCount() = %d, want 1", got)
	}
}

func TestGetStreak_LongStreak(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 10, 14)

	entries := make(map[string][]string)
	for i := 0; i < 10; i++ {
		date := time.Date(2026, 3, 1+i, 0, 0, 0, 0, time.UTC)
		entries[date.Format("2006-01-02")] = []string{"task"}
	}
	path := writeCompletedFile(t, dir, entries)

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if got := cc.GetStreak(); got != 10 {
		t.Errorf("GetStreak() = %d, want 10", got)
	}
}

func TestLoadFromFile_MultipleLoads(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTime(2026, 3, 2, 14)

	path := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"task A"},
	})

	cc := NewCompletionCounterWithNow(frozen)
	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("first LoadFromFile() error: %v", err)
	}

	if got := cc.GetTodayCount(); got != 1 {
		t.Errorf("after first load GetTodayCount() = %d, want 1", got)
	}

	// Write more data and reload
	path = writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"task A", "task B", "task C"},
	})

	if err := cc.LoadFromFile(path); err != nil {
		t.Fatalf("second LoadFromFile() error: %v", err)
	}

	// Should reflect the new data, not accumulate
	if got := cc.GetTodayCount(); got != 3 {
		t.Errorf("after second load GetTodayCount() = %d, want 3", got)
	}
}
