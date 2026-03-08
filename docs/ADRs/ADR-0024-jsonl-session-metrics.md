# ADR-0024: JSONL Session Metrics Format

- **Status:** Accepted
- **Date:** 2025-11-08
- **Decision Makers:** Project founder
- **Related PRs:** #16 (Story 1.5), #43-#45, #82, #95
- **Related ADRs:** ADR-0003 (YAML Persistence)

## Context

ThreeDoors tracks session-level metrics: which tasks were shown, selected, completed, time spent in sessions, mood correlations, and avoidance patterns. This data powers the learning engine and user insights dashboard.

## Decision

Use **JSONL (JSON Lines)** format for session metrics in `~/.threedoors/metrics.jsonl`:
- One JSON object per line
- Append-only — never modify or delete previous entries
- Each entry has a type field, timestamp, and event-specific data

## Rationale

- Append-only writes are fast and safe (no read-modify-write cycle)
- JSONL is streamable — can process large files line by line
- Each line is independently parseable — partial corruption doesn't lose all data
- Easy to analyze with standard tools (`jq`, `grep`, `wc -l`)
- No schema migration needed — new event types simply add new line formats

## Event Types

- `session_start` / `session_end` — Session boundaries
- `door_shown` / `door_selected` — Door interaction tracking
- `task_completed` / `task_status_changed` — Task lifecycle events
- `mood_recorded` — User mood self-report
- `avoidance_detected` — Pattern analysis results

## Consequences

### Positive
- Append-only is the safest write pattern — no data loss risk
- Session metrics reader library (Story 3.5.6) enables analysis
- Mood correlation analysis (Story 4.3) and avoidance detection (Story 4.4) built on this data
- "Better Than Yesterday" tracking (Story 4.6) uses metrics history

### Negative
- File grows indefinitely (mitigated by rotation/archival in future)
- No query indexes — analysis requires full scan
- Schema evolution handled by adding new event types (old readers ignore unknown types)
