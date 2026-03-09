# Architecture: Beautiful Stats Display (Epic 40)

**Date:** 2026-03-08
**Input:** Research report, UX review, party mode decisions (D-086 through D-094)

---

## 1. Bubbletea MVU Integration

### View Architecture

The stats dashboard integrates as an enhancement to the existing `InsightsView`, not as a new top-level view. The existing view routing in `MainModel` already handles `viewInsights` state — no routing changes needed.

```
MainModel
  ├── DoorsView (viewDoors)
  ├── DetailView (viewDetail)
  ├── InsightsView (viewInsights)  <-- Enhanced in this epic
  ├── SettingsView (viewSettings)
  └── ...other views
```

### InsightsView State Extension

```go
type InsightsView struct {
    analyzer    *core.PatternAnalyzer
    counter     *core.CompletionCounter
    width       int
    height      int                    // NEW: terminal height for layout
    activeTab   int                    // NEW: 0=Overview, 1=Detail (Phase 2)
    funFact     string                 // NEW: cached fun fact for current entry

    // Phase 2 additions:
    animating         bool             // counter animation active
    animationProgress float64          // 0.0 to 1.0
    viewport          viewport.Model   // bubbles viewport for Detail tab scrolling

    // Phase 3 additions:
    theme             *themes.DoorTheme // active theme for stats colors
    milestoneShown    string            // currently displayed milestone (or "")
    milestoneDismiss  time.Time         // auto-dismiss time
}
```

### Message Types

```go
// Phase 1
type ReturnToDoorsMsg struct{}     // existing

// Phase 2
type StatsAnimationTickMsg struct{} // tea.Tick for counter animation
type TabSwitchMsg struct{ Tab int } // tab navigation

// Phase 3
type MilestoneDismissMsg struct{}   // auto-dismiss after 5s
```

## 2. Component Design

### Phase 1 Components

```
InsightsView.View()
  ├── renderHeader()          // "YOUR INSIGHTS DASHBOARD" with Lipgloss border
  ├── renderHeroNumber()      // "★ 247 tasks completed ★" styled gold
  ├── renderCompletionTrends() // gradient sparkline + week-over-week (ENHANCED)
  ├── renderStreaks()          // bordered panel (ENHANCED with Lipgloss)
  ├── renderMoodCorrelations() // bar charts instead of text (ENHANCED)
  ├── renderDoorPreferences()  // bordered panel (ENHANCED)
  ├── renderFunFact()          // NEW: rotating celebration fact
  └── renderFooter()           // "Press Esc to return"
```

Layout uses `lipgloss.JoinHorizontal()` to place panels side-by-side and `lipgloss.JoinVertical()` to stack rows.

### Phase 2 Components (Tab: Detail)

```
InsightsView.renderDetailTab()
  ├── renderHeatmap()          // NEW: 7x8 activity grid
  ├── renderSessionHighlights() // NEW: hidden metrics surfaced
  ├── renderTimeOfDay()        // NEW: 24-hour activity bar
  └── renderAnimatedCounters() // NEW: tick-up on entry
```

### Phase 3 Components

```
InsightsView (extended)
  ├── renderMilestoneBanner()  // NEW: one-time celebration
  └── All panels use theme.StatsAccent for colors
```

## 3. Data Flow

### Current data path (unchanged):

```
sessions.jsonl → SessionTracker.LoadSessions() → PatternAnalyzer
                                                → CompletionCounter
```

### Stats view data flow:

```
InsightsView.View()
  ├── iv.analyzer.GetDailyCompletions(7)     → sparkline data
  ├── iv.analyzer.GetWeekOverWeek()          → trend data
  ├── iv.counter.GetStreak()                 → streak count
  ├── iv.analyzer.GetMoodCorrelations()      → mood bars data
  ├── iv.analyzer.GetDoorPositionPreferences() → door prefs data
  ├── iv.funFactGen.Generate()               → fun fact string (NEW)
  │
  │ Phase 2:
  ├── iv.analyzer.GetDailyCompletions(56)    → heatmap data (8 weeks)
  ├── iv.analyzer.GetSessionHighlights()     → hidden metrics (NEW method)
  └── iv.counter.GetTotalCompleted()         → hero number
```

### New core methods needed:

```go
// internal/core/pattern_analyzer.go
func (pa *PatternAnalyzer) GetSessionHighlights() SessionHighlights
type SessionHighlights struct {
    TotalDoors         int
    TotalTasks         int
    AvgSessionDuration time.Duration
    FastestFirstDoor   time.Duration
    TotalDetailViews   int
    TotalNotesAdded    int
    LongestStreak      int
    AverageStreak      float64
    PeakHour           int // 0-23
}

// internal/core/fun_facts.go (NEW FILE)
type FunFactGenerator struct {
    analyzer *PatternAnalyzer
    counter  *CompletionCounter
}
func NewFunFactGenerator(analyzer *PatternAnalyzer, counter *CompletionCounter) *FunFactGenerator
func (g *FunFactGenerator) Generate() string
```

## 4. Dependency Management

### Phase 1: No new dependencies
- All rendering uses existing `lipgloss` (v1.1.0)
- Gradient via `lipgloss.Blend1D(colorA, colorB, position)` — already available
- Layout via `lipgloss.JoinHorizontal()` / `JoinVertical()` — already available
- Bordered panels via `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` — already available

### Phase 2: Potential ntcharts evaluation
- Only for heatmap (Story 40.5)
- Decision made within that story: prototype custom first, evaluate ntcharts only if custom has issues
- If adopted: `go get github.com/NimbleMarkets/ntcharts@latest`, pin version in `go.mod`
- ntcharts implements `tea.Model` — embeddable within InsightsView

### No other new dependencies across all phases

## 5. Performance Considerations

### Rendering strategy: Cache, don't recompute

Current `View()` recomputes everything on every render. For plain text this was cheap, but styled panels with Lipgloss are more expensive.

**Approach:**
- Compute data once on view entry (in `Init()` or on first `View()` call)
- Store rendered panel strings in InsightsView fields
- Re-render only on width/height change or tab switch
- `Update()` returns no commands unless animating (Phase 2)

```go
type InsightsView struct {
    // ... fields above ...
    cachedView    string // cached rendered output
    cacheValid    bool   // invalidated on resize or tab switch
    lastWidth     int    // detect resize
}

func (iv *InsightsView) View() string {
    if iv.cacheValid && iv.width == iv.lastWidth {
        return iv.cachedView
    }
    iv.cachedView = iv.buildView()
    iv.cacheValid = true
    iv.lastWidth = iv.width
    return iv.cachedView
}
```

### Animation performance (Phase 2)
- `tea.Tick(30ms)` runs only during animation (~500ms = 16 ticks)
- Tick stops after animation completes — no ongoing CPU
- Only numeric values interpolate; panels don't re-render during animation

## 6. Testing Strategy

### Unit tests (all phases)

| Test Area | Approach |
|-----------|----------|
| `sparkline()` with gradient | Table-driven: verify output length, verify non-empty for non-zero data |
| `FunFactGenerator.Generate()` | Verify non-empty string, verify banned words absent, verify deterministic for same day |
| Layout at different widths | Call `SetWidth()` with 50, 80, 120; verify output doesn't exceed width |
| Bar chart rendering | Verify proportional fill widths, verify labels present |
| Heatmap data aggregation | Table-driven with known session data, verify 7x8 grid |
| Milestone persistence | Verify file write/read, verify show-once behavior |

### Golden file tests

Add golden files for InsightsView at standard widths (80, 120) to detect visual regressions. Place in `internal/tui/testdata/`.

### No snapshot tests for color
Color escape codes vary by terminal. Test structure and content, not ANSI codes. Use `lipgloss.NewRenderer(io.Discard)` or strip ANSI for assertions.

### Race detector
InsightsView has no goroutines or shared state — race detector should pass trivially. Required anyway per CLAUDE.md since we're modifying `internal/tui/`.

## 7. Impact on Existing Code

### Files modified:

| File | Change | Phase |
|------|--------|-------|
| `internal/tui/insights_view.go` | Major rewrite: Lipgloss panels, layout, hero number | 1 |
| `internal/tui/insights_view_test.go` | New tests for layout, sparkline, bars | 1 |
| `internal/core/pattern_analyzer.go` | Add `GetSessionHighlights()` method | 2 |
| `internal/tui/styles.go` | Add stats-specific styles (panel borders, gradient colors) | 1 |

### New files:

| File | Purpose | Phase |
|------|---------|-------|
| `internal/core/fun_facts.go` | FunFactGenerator type | 1 |
| `internal/core/fun_facts_test.go` | Fun fact tests | 1 |
| `internal/tui/testdata/insights_80col.golden` | Golden file | 1 |
| `internal/tui/testdata/insights_120col.golden` | Golden file | 1 |
| `~/.threedoors/milestones.json` | Milestone persistence (runtime) | 3 |

### Files NOT modified:
- `internal/tui/main_model.go` — existing view routing sufficient
- `internal/core/session_tracker.go` — already captures all needed data
- `cmd/threedoors/main.go` — no new initialization needed
- `internal/tui/doors_view.go` — completely separate view

## 8. Fun Facts Engine

### Location: `internal/core/fun_facts.go`

### Design:
- Pure function: takes analyzer + counter data, returns a string
- No side effects, no persistence
- Deterministic: seeded by current UTC date (same fact shown all day, changes tomorrow)
- Pool of ~15-20 fact templates, parameterized with real data

### Fact categories:
1. **Totals** — "You've opened {n} doors since your first session!"
2. **Patterns** — "You complete the most tasks on {day}s"
3. **Speed** — "Your fastest door pick: {n} seconds"
4. **Streaks** — "Your longest streak was {n} days"
5. **Preferences** — "You pick the {position} door {pct}% of the time"
6. **Engagement** — "You've written {n} notes on your tasks"
7. **Mood** — "{mood} is your power mood — avg {n} tasks!"

### Content rules (D-089):
1. Observe, don't prescribe
2. Celebrate totals, not rates
3. Frame gaps as potential
4. No decline comparisons

### Selection algorithm:
```go
func (g *FunFactGenerator) Generate() string {
    facts := g.buildFactPool()
    if len(facts) == 0 {
        return ""
    }
    // Deterministic daily rotation
    day := time.Now().UTC().YearDay() + time.Now().UTC().Year()*366
    idx := day % len(facts)
    return facts[idx]
}
```

## 9. Responsive Layout

### Breakpoints (from UX review):

| Width | Mode | Layout Description |
|-------|------|--------------------|
| < 60 | compact | Single column, all panels stacked, sparkline 5 days |
| 60-79 | narrow | Two columns for small panels (Streaks + Mood) |
| 80-119 | standard | Full two-column layout per mockup A |
| 120+ | wide | Three-column top row (Sparkline + Streaks + Door Picks) |

### Implementation:
```go
func (iv *InsightsView) layoutMode() string {
    switch {
    case iv.width < 60:
        return "compact"
    case iv.width < 80:
        return "narrow"
    case iv.width < 120:
        return "standard"
    default:
        return "wide"
    }
}
```

Each render function checks `iv.layoutMode()` and adjusts panel widths and arrangement accordingly.

## 10. Color Palette

### Phase 1 (independent, color-blind safe):

| Element | Color | Lipgloss |
|---------|-------|----------|
| Hero number | Gold | `lipgloss.Color("#FCD34D")` |
| Sparkline low | Blue | `lipgloss.Color("#3B82F6")` |
| Sparkline high | Yellow | `lipgloss.Color("#EAB308")` |
| Panel border | Gray | `lipgloss.Color("#555555")` |
| Fun fact star | Gold | `lipgloss.Color("#FCD34D")` |
| Mood: Focused | Blue | `lipgloss.Color("#60A5FA")` |
| Mood: Energized | Yellow | `lipgloss.Color("#FBBF24")` |
| Mood: Calm | Green | `lipgloss.Color("#34D399")` |
| Mood: Tired | Gray | `lipgloss.Color("#9CA3AF")` |

All colors use `lipgloss.AdaptiveColor{Light: "...", Dark: "..."}` for light/dark terminal support.

---

## Component Diagram

```
┌─────────────────────────────────────────────┐
│                 MainModel                    │
│  viewState == viewInsights                   │
│  ┌─────────────────────────────────────────┐ │
│  │           InsightsView                  │ │
│  │                                         │ │
│  │  ┌─── Tab 0: Overview ──────────────┐   │ │
│  │  │ renderHeader()                   │   │ │
│  │  │ renderHeroNumber()               │   │ │
│  │  │ ┌────────────┐ ┌──────────────┐  │   │ │
│  │  │ │ Sparkline   │ │ Streaks      │  │   │ │
│  │  │ │ (gradient)  │ │ (panel)      │  │   │ │
│  │  │ └────────────┘ └──────────────┘  │   │ │
│  │  │ ┌────────────┐ ┌──────────────┐  │   │ │
│  │  │ │ Mood Bars   │ │ Door Picks   │  │   │ │
│  │  │ │ (colored)   │ │ (panel)      │  │   │ │
│  │  │ └────────────┘ └──────────────┘  │   │ │
│  │  │ renderFunFact()                  │   │ │
│  │  └──────────────────────────────────┘   │ │
│  │                                         │ │
│  │  ┌─── Tab 1: Detail (Phase 2) ─────┐   │ │
│  │  │ renderHeatmap()                  │   │ │
│  │  │ renderSessionHighlights()        │   │ │
│  │  │ renderTimeOfDay()                │   │ │
│  │  └──────────────────────────────────┘   │ │
│  │                                         │ │
│  │  Data Sources:                          │ │
│  │  ├── PatternAnalyzer                    │ │
│  │  ├── CompletionCounter                  │ │
│  │  └── FunFactGenerator (NEW)             │ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```
