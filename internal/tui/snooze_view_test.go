package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSnoozeView(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)

	if v.task != task {
		t.Error("expected task to be set")
	}
	if v.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", v.cursor)
	}
	if v.inputMode {
		t.Error("expected inputMode to be false")
	}
}

func TestSnoozeViewCursorMovement(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)

	tests := []struct {
		name       string
		key        string
		wantCursor int
	}{
		{"down from 0", "down", 1},
		{"down from 1", "down", 2},
		{"down from 2", "down", 3},
		{"down at bottom stays", "down", 3},
		{"up from 3", "up", 2},
		{"up from 2", "up", 1},
		{"up from 1", "up", 0},
		{"up at top stays", "up", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if v.cursor != tt.wantCursor {
				t.Errorf("got cursor=%d, want %d", v.cursor, tt.wantCursor)
			}
		})
	}
}

func TestSnoozeViewCursorMovementVim(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if v.cursor != 1 {
		t.Errorf("j should move down: got cursor=%d, want 1", v.cursor)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if v.cursor != 0 {
		t.Errorf("k should move up: got cursor=%d, want 0", v.cursor)
	}
}

func TestSnoozeViewEscCancels(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(SnoozeCancelledMsg); !ok {
		t.Errorf("expected SnoozeCancelledMsg, got %T", msg)
	}
}

func TestSnoozeViewSelectTomorrow(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	// cursor is at 0 (Tomorrow)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter")
	}
	msg := cmd()
	snoozed, ok := msg.(TaskSnoozedMsg)
	if !ok {
		t.Fatalf("expected TaskSnoozedMsg, got %T", msg)
	}
	if snoozed.Task != task {
		t.Error("wrong task in message")
	}
	if snoozed.Option != "tomorrow" {
		t.Errorf("expected option 'tomorrow', got %q", snoozed.Option)
	}
	if snoozed.DeferDate == nil {
		t.Fatal("expected DeferDate to be non-nil for tomorrow")
	}
}

func TestSnoozeViewSelectNextWeek(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.cursor = 1 // Next Week

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter")
	}
	msg := cmd()
	snoozed, ok := msg.(TaskSnoozedMsg)
	if !ok {
		t.Fatalf("expected TaskSnoozedMsg, got %T", msg)
	}
	if snoozed.Option != "next_week" {
		t.Errorf("expected option 'next_week', got %q", snoozed.Option)
	}
	if snoozed.DeferDate == nil {
		t.Fatal("expected DeferDate to be non-nil for next week")
	}
}

func TestSnoozeViewSelectPickDate(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.cursor = 2 // Pick Date

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("selecting Pick Date should enter input mode, not emit a command")
	}
	if !v.inputMode {
		t.Error("expected inputMode to be true")
	}
}

func TestSnoozeViewSelectSomeday(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.cursor = 3 // Someday

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter")
	}
	msg := cmd()
	snoozed, ok := msg.(TaskSnoozedMsg)
	if !ok {
		t.Fatalf("expected TaskSnoozedMsg, got %T", msg)
	}
	if snoozed.Option != "someday" {
		t.Errorf("expected option 'someday', got %q", snoozed.Option)
	}
	if snoozed.DeferDate != nil {
		t.Error("expected DeferDate to be nil for someday")
	}
}

func TestSnoozeViewDateInput(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.cursor = 2                             // Pick Date
	v.Update(tea.KeyMsg{Type: tea.KeyEnter}) // enter input mode

	// Type a valid date
	for _, ch := range "2026-04-15" {
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	if v.dateInput != "2026-04-15" {
		t.Errorf("expected dateInput='2026-04-15', got %q", v.dateInput)
	}

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter on valid date")
	}
	msg := cmd()
	snoozed, ok := msg.(TaskSnoozedMsg)
	if !ok {
		t.Fatalf("expected TaskSnoozedMsg, got %T", msg)
	}
	if snoozed.Option != "pick_date" {
		t.Errorf("expected option 'pick_date', got %q", snoozed.Option)
	}
	if snoozed.DeferDate == nil {
		t.Fatal("expected DeferDate to be non-nil for pick_date")
	}
}

func TestSnoozeViewDateInputInvalid(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.cursor = 2
	v.Update(tea.KeyMsg{Type: tea.KeyEnter}) // enter input mode

	for _, ch := range "not-a-date" {
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected no command on invalid date")
	}
	if v.errMsg == "" {
		t.Error("expected error message to be set")
	}
	if !v.inputMode {
		t.Error("should remain in input mode on error")
	}
}

func TestSnoozeViewDateInputBackspace(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.inputMode = true
	v.dateInput = "2026"

	v.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if v.dateInput != "202" {
		t.Errorf("expected '202', got %q", v.dateInput)
	}
}

func TestSnoozeViewDateInputEscReturnsToMenu(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	v := NewSnoozeView(task)
	v.inputMode = true
	v.dateInput = "2026-04"

	v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if v.inputMode {
		t.Error("expected inputMode to be false after Esc")
	}
	if v.dateInput != "" {
		t.Errorf("expected dateInput cleared, got %q", v.dateInput)
	}
}

func TestTomorrow9am(t *testing.T) {
	t.Parallel()
	result := tomorrow9am()
	// Convert back to local for verification
	local := result.In(time.Now().Location())

	now := time.Now().Local()
	expectedDay := now.Day() + 1
	if local.Day() != expectedDay && local.Month() != now.Month() {
		// Handle month rollover: just check it's after now
		if !result.After(time.Now().UTC()) {
			t.Error("tomorrow9am should be in the future")
		}
	}
	if local.Hour() != 9 || local.Minute() != 0 || local.Second() != 0 {
		t.Errorf("expected 09:00:00 local, got %02d:%02d:%02d", local.Hour(), local.Minute(), local.Second())
	}
}

func TestNextMonday9am(t *testing.T) {
	t.Parallel()
	result := nextMonday9am()
	local := result.In(time.Now().Location())

	if local.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", local.Weekday())
	}
	if local.Hour() != 9 || local.Minute() != 0 || local.Second() != 0 {
		t.Errorf("expected 09:00:00 local, got %02d:%02d:%02d", local.Hour(), local.Minute(), local.Second())
	}
	if !result.After(time.Now().UTC()) {
		t.Error("nextMonday9am should be in the future")
	}
}

func TestNextMonday9amOnMonday(t *testing.T) {
	t.Parallel()
	// nextMonday9am should always return a future Monday, even if today is Monday
	result := nextMonday9am()
	local := result.In(time.Now().Location())

	if local.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", local.Weekday())
	}
	// Should be at least 1 day in the future
	if !result.After(time.Now().UTC()) {
		t.Error("nextMonday9am should be in the future even on Monday")
	}
}

func TestParsePickDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid date", "2026-04-15", false},
		{"valid date end of month", "2026-12-31", false},
		{"invalid format", "15-04-2026", true},
		{"invalid text", "not-a-date", true},
		{"empty", "", true},
		{"partial", "2026-04", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parsePickDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePickDate(%q) error=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if err == nil {
				local := result.In(time.Now().Location())
				if local.Hour() != 9 || local.Minute() != 0 {
					t.Errorf("expected 09:00 local, got %02d:%02d", local.Hour(), local.Minute())
				}
			}
		})
	}
}

func TestSnoozeViewRender(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Buy groceries")
	v := NewSnoozeView(task)
	v.SetWidth(80)

	output := v.View()
	if !strings.Contains(output, "SNOOZE TASK") {
		t.Error("expected header 'SNOOZE TASK'")
	}
	if !strings.Contains(output, "Buy groceries") {
		t.Error("expected task text in output")
	}
	if !strings.Contains(output, "Tomorrow") {
		t.Error("expected 'Tomorrow' option")
	}
	if !strings.Contains(output, "Next Week") {
		t.Error("expected 'Next Week' option")
	}
	if !strings.Contains(output, "Pick Date") {
		t.Error("expected 'Pick Date' option")
	}
	if !strings.Contains(output, "Someday") {
		t.Error("expected 'Someday' option")
	}
	if !strings.Contains(output, "> Tomorrow") {
		t.Error("expected cursor on Tomorrow (first item)")
	}
}

func TestSnoozeViewRenderInputMode(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Buy groceries")
	v := NewSnoozeView(task)
	v.SetWidth(80)
	v.inputMode = true
	v.dateInput = "2026-04"

	output := v.View()
	if !strings.Contains(output, "YYYY-MM-DD") {
		t.Error("expected date format hint")
	}
	if !strings.Contains(output, "2026-04") {
		t.Error("expected date input to be shown")
	}
}

func TestSnoozeViewRenderInputModeWithError(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Buy groceries")
	v := NewSnoozeView(task)
	v.SetWidth(80)
	v.inputMode = true
	v.errMsg = "invalid date format, use YYYY-MM-DD"

	output := v.View()
	if !strings.Contains(output, "invalid date format") {
		t.Error("expected error message in output")
	}
}

func TestSnoozeViewLongTaskTextTruncated(t *testing.T) {
	t.Parallel()
	longText := strings.Repeat("A", 100)
	task := core.NewTask(longText)
	v := NewSnoozeView(task)
	v.SetWidth(80)

	output := v.View()
	if strings.Contains(output, longText) {
		t.Error("long task text should be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("truncated text should end with ...")
	}
}

func TestDetailViewZKeyOpenSnooze(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	dv := NewDetailView(task, nil, nil, nil)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if cmd == nil {
		t.Fatal("expected a command from Z key")
	}
	msg := cmd()
	showSnooze, ok := msg.(ShowSnoozeMsg)
	if !ok {
		t.Fatalf("expected ShowSnoozeMsg, got %T", msg)
	}
	if showSnooze.Task != task {
		t.Error("wrong task in ShowSnoozeMsg")
	}
}

func TestDetailViewOptionsBarIncludesSnooze(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test task")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(120)

	output := dv.View()
	if !strings.Contains(output, "[Z]Snooze") {
		t.Error("expected [Z]Snooze in options bar")
	}
}
