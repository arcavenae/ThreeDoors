# Sprint Change Proposal: Beautiful Stats Display

**Date:** 2026-03-08
**Type:** Course Correction — Feature Enhancement
**Requestor:** arcaven
**Priority:** P1 (Phase 1), P2 (Phase 2-3)

---

## Problem Statement

ThreeDoors tracks 14+ behavioral dimensions in session data (`sessions.jsonl`) but only surfaces approximately 6 of them to users. The MCP analytics layer computes rich derived data (burnout indicators, hourly productivity, streak history, weekly summaries) that never appears in the TUI. The existing insights view (`insights_view.go`) is plain monochrome text with manual spacing — no Lipgloss styling, no color gradients, no visual hierarchy.

This is a significant UX gap: users generate valuable behavioral data through daily use but cannot see their own patterns in a compelling way. The current experience violates SOUL.md's principle that "every interaction should feel deliberate" — the insights view feels like an afterthought rather than a celebration.

## Impact Analysis

### What's affected
- `internal/tui/insights_view.go` — primary rendering (208 lines, needs major enhancement)
- `internal/core/greeting_insights.go` — greeting data (may feed fun facts)
- `internal/core/session_tracker.go` — data source (no changes needed, already captures everything)
- `internal/tui/themes/` — theme integration for stats colors (Phase 3)
- `internal/tui/main_model.go` — view routing if adding new navigation

### What's NOT affected
- Task management flow (doors view, detail view, completion)
- Data persistence (no schema changes)
- CLI stats commands (separate scope, could benefit later)
- Existing adapters and providers

### Risk assessment
- **Low risk** — this is a display-only enhancement. No data model changes, no persistence changes, no new state transitions.
- **Dependency risk** — ntcharts library evaluation needed before Phase 2 heatmap. Decision: evaluate for heatmap only; sparklines and bars built in-tree with Lipgloss.
- **SOUL.md risk** — milestones/celebrations must be carefully framed (celebrate, never guilt). Research identifies clear red lines.

## Proposed Approach (3-Phase)

### Phase 1: Visual Polish (3 stories, P1)
Transform existing data into beautiful displays. No new data sources needed.
1. Stats dashboard shell — new view structure with Lipgloss bordered panels, 2-column layout, navigation
2. Gradient sparklines — color-coded completion trends using `lipgloss.Blend1D()`
3. Fun facts engine — rotating celebration facts from session history

### Phase 2: New Visualizations (4 stories, P1/P2)
Add new chart types and surface hidden data.
4. Horizontal bar charts — mood correlation and time-of-day visualizations
5. Activity heatmap — GitHub-style weekly grid (evaluates ntcharts dependency)
6. Surface hidden metrics — session duration, fastest start, detail view engagement, streak history
7. Animated counter reveals — numbers tick up on view entry via `tea.Tick`

### Phase 3: Thematic Integration (3 stories, P2)
Deep integration with door metaphor and themes. Conditional on Phase 1/2 validation.
8. Theme-matched stats colors — stats inherit door theme palette
9. Milestone celebrations — one-time banners for achievement thresholds
10. Trophy room concept — stats as "fourth door" behind the three doors

## Rejected Alternatives

| Alternative | Why Rejected |
|------------|-------------|
| termdash framework | Incompatible event loop, conflicts with Bubbletea |
| Full charting library (go-echarts) | Web-oriented (HTML/SVG), not terminal |
| Achievement/badge system | SOUL.md: "no gamification, no guilt" |
| Productivity scores/grades | SOUL.md: "not a productivity report" |
| Daily/weekly goal targets | Creates pressure; every completed task is a win |

## Effort Estimate

| Phase | Stories | Estimated Effort | Priority |
|-------|---------|-----------------|----------|
| Phase 1 | 3 | 3.5-5 hours | P1 |
| Phase 2 | 4 | 7-9.5 hours | P1/P2 |
| Phase 3 | 3 | 6.5-8.5 hours | P2 |
| **Total** | **10** | **17-23 hours** | Mixed |

## Recommendation

**Approve Phase 1 immediately** — high visual impact, low effort, zero risk. Phase 2 should follow once Phase 1 lands and proves the approach. Phase 3 is conditional on user feedback and SOUL.md alignment review of milestone celebrations.

The epic should be assigned the next available number (Epic 40) and added to the Active Epics section of ROADMAP.md.
