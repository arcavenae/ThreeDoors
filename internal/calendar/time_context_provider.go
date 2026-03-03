package calendar

import (
	"context"
	"sort"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
)

// CalendarTimeContextProvider implements tasks.TimeContextProvider using a CalendarReader.
type CalendarTimeContextProvider struct {
	reader    CalendarReader
	lookAhead time.Duration
}

// NewCalendarTimeContextProvider creates a TimeContextProvider backed by a CalendarReader.
// lookAhead controls how far into the future to scan for events (default: 4 hours).
func NewCalendarTimeContextProvider(reader CalendarReader, lookAhead time.Duration) *CalendarTimeContextProvider {
	if lookAhead <= 0 {
		lookAhead = 4 * time.Hour
	}
	return &CalendarTimeContextProvider{
		reader:    reader,
		lookAhead: lookAhead,
	}
}

// GetTimeContext reads upcoming calendar events and computes the current time context.
func (p *CalendarTimeContextProvider) GetTimeContext(ctx context.Context) (*tasks.TimeContext, error) {
	now := time.Now().UTC()
	end := now.Add(p.lookAhead)

	events, err := p.reader.GetEvents(ctx, now, end)
	if err != nil {
		return &tasks.TimeContext{HasCalendar: false}, nil
	}

	// Filter to non-all-day future events and sort by start time.
	var upcoming []CalendarEvent
	for _, e := range events {
		if e.AllDay {
			continue
		}
		if e.Start.After(now) {
			upcoming = append(upcoming, e)
		}
	}

	if len(upcoming) == 0 {
		// No upcoming events — large available block
		return &tasks.TimeContext{
			HasCalendar:   true,
			AvailableTime: p.lookAhead,
			NextEventIn:   0,
		}, nil
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Start.Before(upcoming[j].Start)
	})

	next := upcoming[0]
	timeUntil := next.Start.Sub(now)

	return &tasks.TimeContext{
		HasCalendar:   true,
		NextEventIn:   timeUntil,
		AvailableTime: timeUntil,
		NextEventName: next.Title,
	}, nil
}
