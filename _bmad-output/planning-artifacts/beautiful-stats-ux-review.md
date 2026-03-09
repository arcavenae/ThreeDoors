# UX Design Review: Beautiful Stats Display

**Date:** 2026-03-08
**Reviewer:** UX Designer (BMAD pipeline)
**Input:** Beautiful stats research report, SOUL.md, existing `insights_view.go`

---

## 1. Visual Hierarchy Recommendations

The stats dashboard should follow an **inverted pyramid** — the most impactful, personally meaningful data at the top, progressively more granular detail below. The user's eyes should land on something that makes them feel good immediately.

### Recommended hierarchy (top to bottom):

1. **Hero number** — a single prominent stat that anchors the view. "You've completed 247 tasks!" rendered large and colorful. This is the "hey look what you did!" moment.
2. **Completion sparkline** — 7-day visual trend. Immediately shows trajectory without reading numbers.
3. **Two-column panels** — Streaks + Mood side by side. Equal visual weight, bordered panels.
4. **Fun fact** — rotating celebratory insight at the bottom. Casual, surprising, personal.
5. **Navigation hint** — subtle help text for drill-down or return.

### Why this order works for SOUL.md:

- Opens with celebration (hero number), not analysis
- Trend line is observational, not judgmental
- Panels are scannable — user picks what interests them
- Fun fact is a delightful closer, like a fortune cookie
- No "you should" messaging anywhere in the hierarchy

## 2. ASCII Mockups

### Mockup A: Phase 1 — 80-Column Layout (Minimum)

```
 ╭──────────────────────────────────────────────────────────────────────────╮
 │                      YOUR INSIGHTS DASHBOARD                            │
 ╰──────────────────────────────────────────────────────────────────────────╯

    ✦ 247 tasks completed since you started ✦

 ╭─── COMPLETION TRENDS (7 Days) ──────────────────────────────────────────╮
 │                                                                         │
 │  Mon   Tue   Wed   Thu   Fri   Sat   Sun                                │
 │   ▁     ▃     ▅     ▇     █     ▂     ▄      <- gradient colored        │
 │   1     3     5     7     8     2     4                                  │
 │                                                                         │
 │  This week: 30  |  Last: 22  |  ↑ 36%                                   │
 ╰─────────────────────────────────────────────────────────────────────────╯

 ╭─── STREAKS ─────────────╮  ╭─── MOOD & PRODUCTIVITY ───────────────────╮
 │                          │  │                                            │
 │  Current: 5 days         │  │  Focused    ████████████████░░░  4.2 avg   │
 │  Longest: 12 days        │  │  Energized  █████████████░░░░░  3.8 avg   │
 │  Average: 4.2 days       │  │  Calm       ████████░░░░░░░░░  2.1 avg   │
 │                          │  │                                            │
 ╰──────────────────────────╯  ╰────────────────────────────────────────────╯

    ★ Wednesday is your power day — avg 6.2 tasks!

 Press Esc to return · Tab for more stats
```

### Mockup B: Phase 1 — 120-Column+ Layout (Wide Terminal)

```
 ╭────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
 │                                          YOUR INSIGHTS DASHBOARD                                              │
 ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

    ✦ 247 tasks completed since you started ✦

 ╭─── COMPLETION TRENDS (7 Days) ──────────────────────╮  ╭─── STREAKS ────────────────╮  ╭─── DOOR PICKS ────╮
 │                                                      │  │                             │  │                    │
 │  Mon   Tue   Wed   Thu   Fri   Sat   Sun             │  │  Current: 5 days            │  │  Left:   40%       │
 │   ▁     ▃     ▅     ▇     █     ▂     ▄              │  │  Longest: 12 days           │  │  Center: 25%       │
 │   1     3     5     7     8     2     4               │  │  Average: 4.2 days          │  │  Right:  35%       │
 │                                                      │  │                             │  │                    │
 │  This week: 30  |  Last: 22  |  ↑ 36%                │  ╰─────────────────────────────╯  │  Slight left       │
 ╰──────────────────────────────────────────────────────╯                                    │  preference        │
                                                                                             ╰────────────────────╯
 ╭─── MOOD & PRODUCTIVITY ────────────────────────────────────────────────────────────────────────────────────────╮
 │                                                                                                                │
 │  Focused    ████████████████████████████████░░░░░░░░  4.2 avg  (12 sessions)                                   │
 │  Energized  ████████████████████████████░░░░░░░░░░░  3.8 avg  (8 sessions)                                    │
 │  Calm       ████████████████░░░░░░░░░░░░░░░░░░░░░░  2.1 avg  (5 sessions)                                    │
 │                                                                                                                │
 ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

    ★ You've been on a 5-day streak — your longest yet was 12!

 Press Esc to return · Tab for more stats
```

### Mockup C: Phase 2 — Heatmap Addition

```
 ╭─── ACTIVITY (Last 8 Weeks) ─────────────────────────────────────────────╮
 │      W1   W2   W3   W4   W5   W6   W7   W8                             │
 │  Mo   ░    ▒    ▓    █    ▓    █    ▓    ░                              │
 │  Tu   ▓    ▓    ░    ▓    █    ▓    ░    ▓                              │
 │  We   █    ▓    █    ▓    ▓    ░    ▓    █                              │
 │  Th   ░    ░    ▓    ░    ▓    ▓    ▓    ░                              │
 │  Fr   ▓    █    ▓    ▓    ░    ▓    █    ▓                              │
 │  Sa        ░         ░         ░                                        │
 │  Su                  ░                                                  │
 │                                                                         │
 │  Legend:  [space]=0   ░=1-2   ▒=3-4   ▓=5-6   █=7+                     │
 ╰─────────────────────────────────────────────────────────────────────────╯
```

## 3. Information Architecture

### Main View (Default Insights Dashboard)
Show immediately on entry. No scrolling needed at 80 columns.

| Section | Data Source | Rationale |
|---------|-----------|-----------|
| Hero number (total tasks) | CompletionCounter | Instant celebration moment |
| 7-day sparkline | PatternAnalyzer.GetDailyCompletions | Most relevant trend |
| Week-over-week | PatternAnalyzer.GetWeekOverWeek | Context for sparkline |
| Streaks panel | CompletionCounter.GetStreak | Motivational but not pressuring |
| Mood panel (bars) | PatternAnalyzer.GetMoodCorrelations | Self-knowledge |
| Fun fact | New FunFactGenerator | Delight |

### Drill-Down (Tab to access, Phase 2+)
Optional deeper view. User navigates here by choice.

| Section | Data Source | Rationale |
|---------|-----------|-----------|
| Activity heatmap (8 weeks) | Aggregated daily data | Patterns at a glance |
| Door position preferences | PatternAnalyzer.GetDoorPositionPreferences | Interesting but niche |
| Session highlights | SessionTracker (hidden metrics) | Time to first door, avg duration, detail views |
| Time-of-day activity | MCP PatternMiner (or new computation) | Self-knowledge |

### NOT on any view
- Burnout indicators (too judgmental for TUI, leave in MCP for AI agent use)
- Bypass patterns raw data (already shown as avoidance list in CLI `--patterns`)
- Raw JSONL metrics (developer tool, not user-facing)

## 4. Color and Theme Integration

### Phase 1: Independent Color Palette
Use a single stats-specific palette that works regardless of door theme. This keeps Phase 1 simple.

**Recommended palette:**
- Sparkline gradient: dim gray (#666) through teal (#2DD4BF) to gold (#FCD34D) — cool-to-warm
- Panel borders: subtle gray (#555) — don't compete with content
- Hero number: gold/amber (#FCD34D) — celebration color
- Fun fact star: same gold
- Bar chart fills: mood-specific (Focused=blue #60A5FA, Energized=yellow #FBBF24, Calm=green #34D399, Tired=gray #9CA3AF)
- Use `lipgloss.AdaptiveColor` for light/dark terminal support

### Phase 3: Theme Integration
Extend `ThemeColors` struct with a `StatsAccent` color. Each theme maps its personality:
- Classic: warm amber bars, traditional feel
- Modern: cool blues and teals, clean lines
- Sci-Fi: neon green/cyan, terminal aesthetic
- Shoji: muted earth tones, understated elegance

## 5. Terminal Size Adaptation

### Breakpoints

| Width | Layout | What Changes |
|-------|--------|-------------|
| < 60 cols | Single column, compact | All panels stack vertically. No side-by-side. Sparkline truncated to 5 days. |
| 60-79 cols | Narrow two-column | Streaks + Mood side-by-side (panels shrink). Sparkline full 7 days. |
| 80-119 cols | Standard two-column | Full layout per Mockup A. All panels present. |
| 120+ cols | Wide three-column | Sparkline + Streaks + Door picks in a row (Mockup B). Mood gets full width below. |

### Implementation approach
- `InsightsView.SetWidth(w)` already exists
- Add a `layoutMode()` method that returns `compact|narrow|standard|wide` based on width
- Each render function checks layout mode and adjusts its output
- Never horizontal-scroll — truncate or abbreviate instead

### Height considerations
- Phase 1 main view should fit in 24 rows (standard terminal height) at 80 columns
- If content exceeds height, use a viewport (bubbles viewport component) with scroll indicator
- Heatmap (Phase 2) may push past 24 rows — it belongs in the drill-down tab

## 6. Interaction Design

### Navigation Model
Keep it minimal and consistent with existing ThreeDoors patterns.

| Key | Action | Context |
|-----|--------|---------|
| `i` or `:insights` | Enter insights view | From doors view (existing) |
| `Esc` | Return to doors view | From insights view (existing) |
| `Tab` | Switch between main/detail tabs | Phase 2+ only |
| `Shift+Tab` | Switch back | Phase 2+ only |
| `j`/`k` or `Up`/`Down` | Scroll within tab | If content exceeds viewport |

### Navigation design principles
- **No modes within the stats view** in Phase 1. It's a single rendered screen. Press Esc to leave.
- **Tab navigation** in Phase 2 adds one level of depth. Tab label shows which view is active: `[Overview] Details` or `Overview [Details]`
- **No drill-down into individual stats**. This is a dashboard, not a database explorer. If the user wants raw data, they use `threedoors stats --patterns` in the CLI.
- **Fun fact refreshes on each entry.** User gets a new fact every time they open insights. No manual refresh needed.

## 7. Phase 2/3 Feature Evaluation

### Worth Pursuing

| Feature | Verdict | Rationale |
|---------|---------|-----------|
| Activity heatmap | YES (Phase 2) | Instantly recognizable pattern visualization. GitHub made this a cultural icon. Users understand it intuitively. Worth the complexity if ntcharts provides a ready-made component; build custom if ntcharts is rejected. |
| Bar charts for mood | YES (Phase 2) | Simple, high-impact visual upgrade over plain text numbers. Lipgloss can render proportional bars with minimal code. |
| Surface hidden metrics | YES (Phase 2) | Data is already captured — just display it. Low effort, high delight ("you didn't know we tracked that!"). |
| Animated counters | YES (Phase 2, subtle) | Should be brief (300-500ms) and subtle. Numbers tick up smoothly, not flashy. If it feels like a video game, it's too much. |
| Theme-matched colors | YES (Phase 3) | Natural extension of Epic 17. Makes stats feel integrated rather than bolted on. |

### Over-Designed / Risky

| Feature | Verdict | Rationale |
|---------|---------|-----------|
| Trophy Room ("fourth door") | DEFER | Cool concept but high effort for uncertain payoff. The door-sliding animation adds complexity to the Bubbletea model state machine. The stats view already serves the purpose. Revisit if users express desire for thematic stats integration. |
| Milestone celebrations | CAREFUL (Phase 3, reduced scope) | One-time banner is acceptable. But "Century Club" naming and "unlocked" language trend toward gamification. Reframe: "You've completed 100 tasks!" (observation) not "Century Club unlocked!" (achievement). Keep it to 3-4 milestones max (first session, 50 tasks, 100 tasks, 10-day streak). No "next milestone" messaging. |

### Why Trophy Room should be deferred:
1. The door animation adds a new view state transition that must be tested and maintained
2. The metaphor ("behind the doors") implies stats are hidden — but they should be easily accessible
3. The "trophy room" name has gamification connotations (trophies = prizes = competition)
4. A well-designed insights dashboard achieves the same delight without the complexity
5. If we revisit this, rename it: "Your Room" or "Your Story" instead of "Trophy Room"

## 8. Accessibility Considerations

### Color-Blind Safe Palettes

The sparkline gradient and mood bar colors must be distinguishable for all major color vision deficiencies (protanopia, deuteranopia, tritanopia).

**Recommended approach:**
- Use `lipgloss.AdaptiveColor` with separate Light and Dark palettes
- Sparkline gradient: blue (#3B82F6) -> teal (#14B8A6) -> yellow (#EAB308) — avoids red-green axis entirely
- Mood bars: use pattern/texture in addition to color. Focused=solid fill, Energized=dotted fill, Calm=light fill, Tired=very light fill. This ensures distinguishability even in grayscale.
- Heatmap: use Unicode shade characters (space, ░, ▒, ▓, █) in addition to color. The character itself encodes intensity, so color is redundant reinforcement, not the only signal.

### High Contrast Mode

- If terminal background is detected as light (`termenv.HasDarkBackground()` returns false), invert the palette to dark-on-light
- Panel borders should use standard box-drawing characters, not thin Unicode that might render poorly in some terminals
- Fun fact star (`★`) is safe — universal Unicode support
- Avoid emoji for data display (sparkline/bars use Unicode blocks, which have excellent terminal support)

### Screen Reader Considerations

- Bubbletea does not have native screen reader support, but text-based output is inherently readable by screen readers
- Ensure all visual elements have text equivalents: sparkline should include the number below each bar
- Bar charts should include the numeric value after the bar: `Focused ████████░░ 4.2 avg`
- Heatmap should include the legend with explicit number ranges

### Terminal Compatibility

- Test on: iTerm2, Terminal.app, Alacritty, Kitty, tmux, screen
- Unicode block characters (▁▂▃▄▅▆▇█) are safe across all modern terminals
- Box-drawing characters (╭╮╰╯│─) are safe across all terminals with UTF-8
- Braille patterns (proposed in research) should NOT be used — inconsistent rendering across terminals and fonts
- Rounded borders (`lipgloss.RoundedBorder()`) are preferred over sharp corners for warmth, but fall back to `lipgloss.NormalBorder()` if rendering issues detected

---

## Summary of Recommendations

1. **Approve Phase 1** with the visual hierarchy described above (hero number, sparkline, panels, fun fact)
2. **Use a stats-independent color palette** in Phase 1; theme integration deferred to Phase 3
3. **Implement responsive layout** with 4 breakpoints (< 60, 60-79, 80-119, 120+)
4. **Tab navigation** for Phase 2 drill-down (main view + detail view)
5. **Build heatmap and bar charts in Phase 2** — ntcharts evaluation for heatmap only
6. **Defer Trophy Room** — high complexity, uncertain value, gamification risk
7. **Reduce milestone scope** — 3-4 observations, no achievement language
8. **Color-blind safe palette** using blue-teal-yellow gradient, Unicode shape redundancy
