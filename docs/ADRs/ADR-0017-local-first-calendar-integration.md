# ADR-0017: Local-First Calendar Integration

- **Status:** Accepted
- **Date:** 2026-02-01
- **Decision Makers:** Architecture review
- **Related PRs:** #65 (Story 12.1), #81 (Story 12.2)

## Context

Calendar awareness (Epic 12) enables time-contextual door selection — showing tasks relevant to the current time of day. Calendar integration typically requires OAuth flows for Google Calendar or Microsoft Graph API.

## Decision

Use **local-only calendar sources**: AppleScript for macOS Calendar.app, `.ics` file parsing, and local CalDAV cache. No OAuth, no cloud API calls.

## Rationale

- Consistent with local-first architecture principle
- No OAuth complexity or token management
- No third-party service dependency at runtime
- macOS Calendar.app is the primary target platform's native calendar
- `.ics` files cover exported/synced calendars from any provider
- CalDAV cache files are locally available when Calendar.app syncs

## Implementation

- AppleScript bridge reads today's events from Calendar.app
- `go-ical` parses `.ics` files for portable calendar support
- Door selection uses event data for time-of-day context weighting
- Calendar data is read-only — ThreeDoors never writes calendar events

## Consequences

### Positive
- No OAuth setup required — works immediately on macOS
- No network calls for calendar data — fast and reliable
- Privacy-preserving — calendar data never leaves the machine
- Works offline (Calendar.app maintains local cache)

### Negative
- macOS-only for AppleScript path (`.ics` parsing is cross-platform)
- Calendar data may be stale if Calendar.app hasn't synced recently
- Cannot create calendar entries for task time-blocking
