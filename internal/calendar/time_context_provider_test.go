package calendar

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCalendarTimeContextProvider_ReaderError(t *testing.T) {
	t.Parallel()
	reader := &mockReader{err: errors.New("calendar unavailable")}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.HasCalendar {
		t.Error("expected HasCalendar=false on reader error")
	}
}

func TestCalendarTimeContextProvider_NoEvents(t *testing.T) {
	t.Parallel()
	reader := &mockReader{events: nil}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tc.HasCalendar {
		t.Error("expected HasCalendar=true when reader succeeds")
	}
	if tc.NextEventIn != 0 {
		t.Errorf("NextEventIn = %v, want 0 (no events)", tc.NextEventIn)
	}
	if tc.AvailableTime != 4*time.Hour {
		t.Errorf("AvailableTime = %v, want 4h (full lookahead)", tc.AvailableTime)
	}
}

func TestCalendarTimeContextProvider_UpcomingEvent(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reader := &mockReader{
		events: []CalendarEvent{
			{Title: "Team Standup", Start: now.Add(45 * time.Minute), End: now.Add(75 * time.Minute)},
		},
	}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tc.HasCalendar {
		t.Error("expected HasCalendar=true")
	}
	if tc.NextEventName != "Team Standup" {
		t.Errorf("NextEventName = %q, want %q", tc.NextEventName, "Team Standup")
	}
	// Allow some tolerance for test execution time
	if tc.NextEventIn < 44*time.Minute || tc.NextEventIn > 46*time.Minute {
		t.Errorf("NextEventIn = %v, want ~45min", tc.NextEventIn)
	}
}

func TestCalendarTimeContextProvider_SkipsAllDayEvents(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reader := &mockReader{
		events: []CalendarEvent{
			{Title: "Holiday", Start: now, End: now.Add(24 * time.Hour), AllDay: true},
			{Title: "Meeting", Start: now.Add(30 * time.Minute), End: now.Add(60 * time.Minute)},
		},
	}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.NextEventName != "Meeting" {
		t.Errorf("NextEventName = %q, want %q (should skip all-day)", tc.NextEventName, "Meeting")
	}
}

func TestCalendarTimeContextProvider_PicksEarliestEvent(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reader := &mockReader{
		events: []CalendarEvent{
			{Title: "Later Meeting", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
			{Title: "Soon Meeting", Start: now.Add(20 * time.Minute), End: now.Add(40 * time.Minute)},
		},
	}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.NextEventName != "Soon Meeting" {
		t.Errorf("NextEventName = %q, want %q", tc.NextEventName, "Soon Meeting")
	}
}

func TestCalendarTimeContextProvider_SkipsPastEvents(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reader := &mockReader{
		events: []CalendarEvent{
			{Title: "Past", Start: now.Add(-30 * time.Minute), End: now.Add(-10 * time.Minute)},
			{Title: "Current", Start: now.Add(-5 * time.Minute), End: now.Add(25 * time.Minute)},
			{Title: "Future", Start: now.Add(60 * time.Minute), End: now.Add(90 * time.Minute)},
		},
	}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "Current" started in the past so Start is not After(now); "Future" is the next upcoming
	if tc.NextEventName != "Future" {
		t.Errorf("NextEventName = %q, want %q (should skip past/current)", tc.NextEventName, "Future")
	}
}

func TestCalendarTimeContextProvider_DefaultLookAhead(t *testing.T) {
	t.Parallel()
	reader := &mockReader{events: nil}
	provider := NewCalendarTimeContextProvider(reader, 0)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.AvailableTime != 4*time.Hour {
		t.Errorf("AvailableTime = %v, want 4h (default lookahead)", tc.AvailableTime)
	}
}

func TestCalendarTimeContextProvider_OnlyAllDayEvents(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	reader := &mockReader{
		events: []CalendarEvent{
			{Title: "Holiday", Start: now, End: now.Add(24 * time.Hour), AllDay: true},
		},
	}
	provider := NewCalendarTimeContextProvider(reader, 4*time.Hour)

	tc, err := provider.GetTimeContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tc.HasCalendar {
		t.Error("expected HasCalendar=true")
	}
	if tc.NextEventIn != 0 {
		t.Errorf("NextEventIn = %v, want 0 (no timed events)", tc.NextEventIn)
	}
}
