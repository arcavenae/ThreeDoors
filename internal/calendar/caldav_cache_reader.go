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

const defaultCalDAVCachePath = "Library/Calendars"

// CalDAVCacheReader reads calendar events from the local CalDAV cache
// stored on macOS at ~/Library/Calendars/. This directory contains
// .caldav/ subdirectories with .ics files representing cached events.
type CalDAVCacheReader struct {
	basePath string
}

// NewCalDAVCacheReader creates a reader for the CalDAV cache at the given base path.
// If basePath is empty, defaults to ~/Library/Calendars/.
func NewCalDAVCacheReader(basePath string) *CalDAVCacheReader {
	if basePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			basePath = filepath.Join("~", defaultCalDAVCachePath)
		} else {
			basePath = filepath.Join(home, defaultCalDAVCachePath)
		}
	}
	return &CalDAVCacheReader{basePath: basePath}
}

// GetEvents reads cached calendar events from the CalDAV cache directory.
// It walks the directory tree looking for .ics files and parses each one.
func (r *CalDAVCacheReader) GetEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
	if _, err := os.Stat(r.basePath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("caldav cache stat %s: %w", r.basePath, err)
	}

	var allEvents []CalendarEvent

	err := filepath.WalkDir(r.basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible entries.
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".ics") {
			return nil
		}

		events, err := r.readICSFile(ctx, path, start, end)
		if err != nil {
			return nil // Skip unparseable files.
		}
		allEvents = append(allEvents, events...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("caldav cache walk %s: %w", r.basePath, err)
	}

	return allEvents, nil
}

func (r *CalDAVCacheReader) readICSFile(_ context.Context, path string, start, end time.Time) ([]CalendarEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return parseICS(bufio.NewScanner(f), start, end)
}
