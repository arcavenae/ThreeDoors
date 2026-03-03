package calendar

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MultiSourceReader aggregates events from multiple CalendarReader sources.
// Individual source failures are logged but do not prevent other sources from loading.
type MultiSourceReader struct {
	readers []namedReader
}

type namedReader struct {
	name   string
	reader CalendarReader
}

// NewMultiSourceReader creates a reader that aggregates from the given sources.
func NewMultiSourceReader(readers map[string]CalendarReader) *MultiSourceReader {
	var named []namedReader
	for name, r := range readers {
		named = append(named, namedReader{name: name, reader: r})
	}
	sort.Slice(named, func(i, j int) bool {
		return named[i].name < named[j].name
	})
	return &MultiSourceReader{readers: named}
}

// GetEvents retrieves events from all sources, collecting errors without failing.
// Returns all successfully loaded events and a combined error for any failures.
func (m *MultiSourceReader) GetEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
	var allEvents []CalendarEvent
	var errs []string

	for _, nr := range m.readers {
		if ctx.Err() != nil {
			return allEvents, ctx.Err()
		}

		events, err := nr.reader.GetEvents(ctx, start, end)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", nr.name, err))
			continue
		}
		allEvents = append(allEvents, events...)
	}

	if len(errs) > 0 && len(allEvents) == 0 {
		return nil, fmt.Errorf("all calendar sources failed: %s", strings.Join(errs, "; "))
	}

	// Sort by start time for consistent ordering.
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Start.Before(allEvents[j].Start)
	})

	return allEvents, nil
}

// NewReadersFromConfig creates CalendarReader instances from configuration.
// Returns a MultiSourceReader combining all enabled sources, or nil if calendar is disabled.
func NewReadersFromConfig(cfg *Config) *MultiSourceReader {
	if cfg == nil || !cfg.Enabled || len(cfg.Sources) == 0 {
		return nil
	}

	readers := make(map[string]CalendarReader)
	for i, src := range cfg.Sources {
		name := fmt.Sprintf("%s-%d", src.Type, i)
		switch src.Type {
		case SourceAppleScript:
			readers[name] = NewAppleScriptReader()
		case SourceICS:
			readers[name] = NewICSReader([]string{src.Path})
		case SourceCalDAVCache:
			readers[name] = NewCalDAVCacheReader(src.Path)
		}
	}

	if len(readers) == 0 {
		return nil
	}

	return NewMultiSourceReader(readers)
}
