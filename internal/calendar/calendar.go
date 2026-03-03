package calendar

import (
	"context"
	"sort"
	"time"
)

// CalendarEvent represents a calendar event read from local sources.
type CalendarEvent struct {
	Title    string
	Start    time.Time
	End      time.Time
	AllDay   bool
	Calendar string
}

// TimeBlock represents a free time block between calendar events.
type TimeBlock struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// CalendarReader reads local calendar sources to determine upcoming events
// and available time blocks. Strictly local-first — no OAuth, no cloud APIs.
type CalendarReader interface {
	GetEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error)
}

// GetFreeBlocks computes free time blocks between events within a time range.
// Events are sorted by start time and overlapping events are merged before
// computing gaps. The minimum block duration filters out trivially short gaps.
func GetFreeBlocks(events []CalendarEvent, rangeStart, rangeEnd time.Time, minBlockDuration time.Duration) []TimeBlock {
	// Filter to non-all-day events within the range.
	var relevant []CalendarEvent
	for _, e := range events {
		if e.AllDay {
			continue
		}
		if e.End.After(rangeStart) && e.Start.Before(rangeEnd) {
			relevant = append(relevant, e)
		}
	}

	if len(relevant) == 0 {
		d := rangeEnd.Sub(rangeStart)
		if d >= minBlockDuration {
			return []TimeBlock{{Start: rangeStart, End: rangeEnd, Duration: d}}
		}
		return nil
	}

	// Sort by start time.
	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].Start.Before(relevant[j].Start)
	})

	// Merge overlapping events into busy intervals.
	type interval struct{ start, end time.Time }
	merged := []interval{{start: relevant[0].Start, end: relevant[0].End}}
	for _, e := range relevant[1:] {
		last := &merged[len(merged)-1]
		if !e.Start.After(last.end) {
			if e.End.After(last.end) {
				last.end = e.End
			}
		} else {
			merged = append(merged, interval{start: e.Start, end: e.End})
		}
	}

	// Compute gaps.
	var blocks []TimeBlock
	cursor := rangeStart

	for _, busy := range merged {
		gapStart := cursor
		gapEnd := busy.start
		if gapStart.Before(rangeStart) {
			gapStart = rangeStart
		}
		if gapEnd.After(rangeEnd) {
			gapEnd = rangeEnd
		}
		if gapEnd.After(gapStart) {
			d := gapEnd.Sub(gapStart)
			if d >= minBlockDuration {
				blocks = append(blocks, TimeBlock{Start: gapStart, End: gapEnd, Duration: d})
			}
		}
		cursor = busy.end
	}

	// Trailing gap after last event.
	if cursor.Before(rangeEnd) {
		d := rangeEnd.Sub(cursor)
		if d >= minBlockDuration {
			blocks = append(blocks, TimeBlock{Start: cursor, End: rangeEnd, Duration: d})
		}
	}

	return blocks
}
