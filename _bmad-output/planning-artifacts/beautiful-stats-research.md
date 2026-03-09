# Research: Beautiful Stats Display for ThreeDoors TUI

**Date:** 2026-03-08
**Type:** Research spike
**Scope:** How to make ThreeDoors' statistics display "super duper awesome and more interesting and fun and pretty"

---

## 1. Current State Analysis

### What ThreeDoors Tracks (Data Available)

ThreeDoors captures extensive behavioral data in `~/.threedoors/sessions.jsonl` via `SessionTracker` (`internal/core/session_tracker.go`). Each JSONL line contains:

| Data Point | Field | Currently Displayed? |
|-----------|-------|---------------------|
| Session ID & timestamps | `session_id`, `start_time`, `end_time` | No |
| Session duration | `duration_seconds` | No (computed) |
| Tasks completed | `tasks_completed` | Yes (session count in doors view, daily/weekly in CLI) |
| Doors viewed | `doors_viewed` | Partially (day-over-day in greeting) |
| Refreshes used | `refreshes_used` | No (bypass rate computed but not shown in TUI) |
| Detail views opened | `detail_views` | No |
| Notes added | `notes_added` | No |
| Status changes | `status_changes` | No |
| Mood entries | `mood_entries_detail[]` (timestamp, mood, custom_text) | Yes (mood correlation in insights view) |
| Door selections | `door_selections[]` (timestamp, position, task_text) | Yes (position bias in insights view) |
| Task bypasses | `task_bypasses[][]` | Partially (avoidance list in CLI `--patterns`) |
| Door feedback | `door_feedback[]` (task_id, feedback_type, comment) | No |
| Time to first door | `time_to_first_door_seconds` | No |

### How Stats Are Currently Displayed

**Three display surfaces exist today:**

#### 1. TUI Doors View Greeting (`internal/tui/doors_view.go` + `internal/core/greeting_insights.go`)
- One-line multi-dimensional greeting at the top of the doors view
- Shows: tasks today vs yesterday, doors opened comparison, streak days, mood + trend, week-over-week
- Plain text with emoji prefix (`clipboard: `) — no color, no visual flourish
- Encouraging messaging regardless of direction ("every one counts!")

#### 2. TUI Insights View (`internal/tui/insights_view.go`)
- Accessible via a key shortcut from the doors view
- Four sections: Completion Trends (7-day sparkline), Streaks, Mood & Productivity, Door Position Preferences
- Uses Unicode sparkline characters (`sparkChars = []rune{'lower_one_eighth_block', 'lower_quarter_block', ..., 'full_block'}`)
- Plain text with minimal formatting — no Lipgloss styles applied to data
- Tabular text with `fmt.Fprintf` for alignment

#### 3. CLI Stats Command (`internal/cli/stats.go`)
- `threedoors stats` — summary dashboard (today completed, streak, completion rate, total sessions)
- `threedoors stats --daily` — 7-day table (DATE | COMPLETED)
- `threedoors stats --weekly` — week-over-week comparison
- `threedoors stats --patterns` — full pattern analysis (door bias, time-of-day, mood, avoidance)
- Plain text output, tab-aligned tables, no color

### Gap Analysis: Tracked but Never Shown

The following data is captured but never surfaces to the user in any display:
- **Time to first door** — how quickly the user starts engaging
- **Detail view count** — curiosity/engagement metric
- **Notes added count** — engagement metric
- **Door feedback entries** — why doors were declined
- **Session duration** — how long sessions last
- **Bypass patterns over time** — are they avoiding fewer tasks?
- **Hourly productivity data** — peak/slump hours (computed in MCP analytics but not in TUI)
- **Burnout indicators** — computed in MCP analytics, never shown
- **Streak history** — longest streak, average streak (computed in MCP analytics)
- **Weekly summary** — best/worst day, velocity, patterns (computed in MCP analytics)

---

## 2. Available Go TUI Visualization Libraries & Techniques

### Already in the Project

| Library | Version | Capabilities |
|---------|---------|-------------|
| `charmbracelet/lipgloss` v1.1.0 | Styling | Borders, colors, padding, alignment, `JoinHorizontal`/`JoinVertical`, adaptive colors, gradients via `Blend1D()`/`Blend2D()`, color manipulation (`Darken()`, `Lighten()`, `Complementary()`), border gradient via `BorderForegroundBlend()` |
| `charmbracelet/bubbles` v1.0.0 | Components | Progress bar, spinner, viewport, text input, list, paginator |
| `charmbracelet/bubbletea` v1.3.10 | Framework | `tea.Cmd` for async, `tea.Tick` for timers/animation |
| `muesli/termenv` v0.16.0 | Terminal | Color profiles, background detection, hyperlinks |
| Custom sparkline | In-tree | `sparkline()` function in `insights_view.go` using block characters |

### Key Libraries to Evaluate

#### ntcharts (`NimbleMarkets/ntcharts`) — RECOMMENDED
- **Built for Bubbletea** — native integration, implements `tea.Model`
- Chart types: sparkline, bar chart, line chart, streamline chart, heatmap, scatter, waveline, time series, OHLC/candlestick
- Canvas-based rendering with Lipgloss styling
- BubbleZone mouse support
- Last published: January 2026 — actively maintained
- **Effort:** Low. Drop-in Bubbletea components.
- **Risk:** New dependency, but small and purpose-built for the stack.

#### asciigraph (`guptarohit/asciigraph`) — GOOD COMPLEMENT
- Lightweight ASCII line graphs with no dependencies
- `Plot()` for single series, `PlotMany()` for multiple
- Height, width, caption, color options
- Returns a string — easy to embed in any View()
- **Effort:** Very low. `asciigraph.Plot(data)` returns a renderable string.
- **Risk:** Minimal. Zero dependencies.

#### termdash (`mum4k/termdash`) — NOT RECOMMENDED
- Full terminal dashboard framework — would conflict with Bubbletea
- Has its own event loop, incompatible architecture
- Useful for inspiration only.

#### Unicode Block Characters (In-tree)
Already using `sparkChars` (`lower_one_eighth_block` through `full_block`). Additional useful characters:
- Progress bars: `bar_fill` `bar_empty` or `progress_full` / `progress_light` / `progress_medium` / `progress_dark`
- Heatmap shading: `light_shade` `medium_shade` `dark_shade` `full_block`
- Trend arrows: up_arrow down_arrow right_arrow up_right_arrow down_right_arrow
- Box drawing: `box_horizontal` `box_vertical` `box_top_left` `box_top_right` etc.
- Braille patterns: `0x2800-0x28FF` (256 characters for high-resolution dot plots)

### Lipgloss Advanced Techniques Available Today

These features are in the project's current Lipgloss version:

1. **Color gradients** — `lipgloss.Blend1D()` for smooth color transitions across sparklines
2. **Adaptive colors** — `lipgloss.AdaptiveColor{Light: "#333", Dark: "#EEE"}` for light/dark terminal support
3. **Border styles** — `lipgloss.RoundedBorder()`, `lipgloss.ThickBorder()`, `lipgloss.DoubleBorder()`
4. **Border gradients** — `BorderForegroundBlend()` for gradient borders
5. **Layout** — `lipgloss.JoinHorizontal()`, `lipgloss.JoinVertical()`, `lipgloss.Place()` for positioning
6. **Tables** — `lipgloss.Table()` with configurable borders, headers, and cell styles

### Animation Capabilities

Bubbletea supports animation through `tea.Tick`:

```go
// Example: counting animation
func tickCmd() tea.Cmd {
    return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

This enables:
- Counters that tick up from 0 to final value
- Progress bars that fill smoothly
- Sparkline that draws left-to-right
- Fade-in effects (changing text opacity over frames)

---

## 3. Design Ideas — Ranked by Impact vs Effort

### Tier 1: Quick Wins (1-2 stories each, high visual impact)

#### 1A. Colorful Sparkline with Gradient
**Current:** `sparkline()` returns plain Unicode blocks, monochrome
**Proposed:** Apply Lipgloss color gradient from cool (low) to warm (high) across the sparkline characters. Zero completions = dim gray, high completions = bright green/gold.

```
  COMPLETION TRENDS (Last 7 Days)
  Mon  Tue  Wed  Thu  Fri  Sat  Sun
  [dim]lower_one_eighth_block[/dim]    [green]lower_half_block[/green]    [green]upper_three_quarter_block[/green]    [bright_green]full_block[/bright_green]    [gold]full_block[/gold]    [dim]lower_quarter_block[/dim]    [green]lower_half_block[/green]
  1    3    5    7    8    2    4
```

**Impact:** High — transforms the most-viewed stats widget from gray text to vivid visualization.
**Effort:** Low — modify `sparkline()` + `renderCompletionTrends()` in `insights_view.go`.
**SOUL alignment:** Pure visual delight, no guilt/pressure.

#### 1B. Styled Stats Dashboard Layout
**Current:** Plain text with manual spacing
**Proposed:** Use Lipgloss bordered panels for each stats section, with headers styled like the door theme colors.

```
+------------------------------+  +------------------------------+
| COMPLETION TRENDS            |  | STREAKS                      |
| Mon Tue Wed Thu Fri Sat Sun  |  | Current: 5 days              |
| [colorful sparkline]         |  | Longest: 12 days             |
| 1   3   5   7   8   2   4   |  | [5 fire emoji]               |
| This week: 30 | Last: 22 up_arrow |  +------------------------------+
+------------------------------+
+------------------------------+  +------------------------------+
| MOOD & PRODUCTIVITY          |  | DOOR PREFERENCES             |
| Focused:   avg 4.2 tasks     |  | [===========|=====|======]   |
| Energized: avg 3.8 tasks     |  |  Left 40%  Center 25% Right 35%|
| Calm:      avg 2.1 tasks     |  |                              |
+------------------------------+  +------------------------------+
```

**Impact:** High — gives structure and visual weight to the dashboard.
**Effort:** Low — Lipgloss `NewStyle().Border()`, `JoinHorizontal()`.
**SOUL alignment:** "Button feel" — visible, satisfying panels.

#### 1C. Fun Facts Line
**Current:** No fun/celebratory content
**Proposed:** A rotating "fun fact" line at the bottom of the insights view, pulled from real session data.

Examples:
- "You've opened 247 doors since you started!"
- "Your fastest session start: 0.3 seconds to pick a door"
- "You complete the most tasks on Wednesdays"
- "You've been on a 5-day streak — your longest yet was 12 days"
- "You tend to pick the left door 42% of the time"
- "Focused mood = 4.2 tasks/session. That's your superpower!"

**Impact:** High — delightful, surprising, makes stats feel personal.
**Effort:** Low — aggregate from existing PatternAnalyzer data.
**SOUL alignment:** "Hey, look at all the things you did!" Celebrates, doesn't measure.

### Tier 2: Medium Effort (2-3 stories each, significant visual upgrade)

#### 2A. GitHub-Style Activity Heatmap
**Current:** Daily completions shown as a table
**Proposed:** A 7-row (days of week) x N-column (weeks) grid using colored Unicode blocks. Each cell's color intensity represents completion count for that day.

```
  ACTIVITY (Last 8 Weeks)
     W1  W2  W3  W4  W5  W6  W7  W8
  Mo  medium_shade   light_shade   dark_shade   full_block   dark_shade   full_block   dark_shade   light_shade
  Tu  dark_shade   dark_shade   medium_shade   dark_shade   full_block   dark_shade   light_shade   dark_shade
  We  full_block   dark_shade   full_block   dark_shade   dark_shade   medium_shade   dark_shade   full_block
  Th  light_shade   medium_shade   dark_shade   light_shade   dark_shade   dark_shade   dark_shade   medium_shade
  Fr  dark_shade   full_block   dark_shade   dark_shade   medium_shade   dark_shade   full_block   dark_shade
  Sa  [space]   light_shade   [space]   light_shade   [space]   light_shade   [space]   [space]
  Su  [space]   [space]   light_shade   [space]   [space]   [space]   [space]   [space]
```

Color scale: `[space]` = 0, light green = 1-2, medium green = 3-4, bright green = 5+

**Impact:** Very high — instantly recognizable, visually striking, shows patterns at a glance.
**Effort:** Medium — need to aggregate daily data across weeks, render grid, apply colors. ntcharts has a heatmap component that could simplify this.
**SOUL alignment:** Celebrates consistency without pressure. Empty squares are neutral, not guilt-inducing.

#### 2B. Horizontal Bar Charts for Mood Correlation
**Current:** `Focused: avg 4.2 tasks (12 sessions)` — plain text
**Proposed:** Horizontal bars with proportional fill and color.

```
  MOOD & PRODUCTIVITY
  Focused    [=============================]  4.2 avg  (12 sessions)
  Energized  [========================     ]  3.8 avg  (8 sessions)
  Calm       [===============              ]  2.1 avg  (5 sessions)
  Tired      [==========                   ]  1.5 avg  (3 sessions)
```

Bars colored by mood (Focused = blue, Energized = yellow, Calm = green, Tired = gray).

**Impact:** High — visual comparison is instant vs reading numbers.
**Effort:** Low-medium — proportional bar rendering with Lipgloss styling.
**SOUL alignment:** Neutral observation, not judgment.

#### 2C. Time-of-Day Radial/Clock Visualization
**Current:** `morning: 12 sessions, avg 3.5 tasks` — text table
**Proposed:** A simplified clock face or hour-bar chart showing when the user is most active.

```
  WHEN YOU'RE MOST ACTIVE
  12am                  6am                  12pm                 6pm
   |                     |                    |                    |
   [.....][.....][.....][===][=====][========][=======][====][===][.....]
                         ^peak                ^peak
```

Or a compact 24-hour bar:

```
  Activity by Hour
  00  03  06  09  12  15  18  21
  [space][space][space]light_shade dark_shade full_block dark_shade full_block dark_shade light_shade [space][space]
  Peak: 9am-11am, 2pm-4pm
```

**Impact:** Medium-high — reveals patterns the user might not know about themselves.
**Effort:** Medium — aggregate hourly data (already computed in MCP PatternMiner).
**SOUL alignment:** Interesting self-knowledge, not prescriptive.

#### 2D. Animated Counter on Entry
**Current:** Stats appear instantly as text
**Proposed:** When opening the insights view, numbers count up from 0 to their final value over ~500ms. Uses `tea.Tick` for frame updates.

```
  Frame 1: Tasks today: 0     Streak: 0 days    Total: 0
  Frame 5: Tasks today: 3     Streak: 2 days    Total: 47
  Frame 10: Tasks today: 7    Streak: 5 days    Total: 143
  Final:   Tasks today: 7     Streak: 5 days    Total: 247
```

**Impact:** Medium — adds "button feel" polish.
**Effort:** Medium — add animation state to InsightsView, `tea.Tick` scheduling.
**SOUL alignment:** "Every interaction should feel deliberate." Satisfying reveal.

### Tier 3: Full Feature (3-5 stories, architectural work)

#### 3A. "Behind Door #4" Stats Room
**Current:** Stats are a separate view with no thematic connection to doors
**Proposed:** Introduce the stats view as a "secret fourth door" or "trophy room." When the user presses the stats key, the three doors animate aside to reveal a stats room behind them.

```
  +------+  +------+  +------+
  | Door |  | Door |  | Door |
  |  1   |  |  2   |  |  3   |
  +------+  +------+  +------+
        \       |       /
         \      |      /
     +---------------------------+
     |    YOUR TROPHY ROOM       |
     |                           |
     |  [sparkline] [heatmap]    |
     |  [bars]      [fun facts]  |
     +---------------------------+
```

**Impact:** Very high — integrates stats into the door metaphor, makes them feel like a reward.
**Effort:** High — animation, new view state, thematic integration.
**SOUL alignment:** Stats as celebration, not obligation.

#### 3B. Theme-Matched Stats Colors
**Current:** Door themes (modern, classic, sci-fi, shoji, golden, shadow) affect door rendering only
**Proposed:** Each door theme also defines a stats color palette. Sci-fi theme gets neon sparklines, golden theme gets warm amber charts, etc.

**Impact:** Medium — visual cohesion.
**Effort:** Medium — extend `ThemeColors` struct, apply in stats rendering.
**SOUL alignment:** Consistent aesthetic.

#### 3C. Milestone Celebrations
**Current:** No recognition of milestones
**Proposed:** When the user hits certain milestones, display a brief celebration in the stats view:

- 100 doors opened: "Century Club! You've opened 100 doors."
- 50 tasks completed: "Half-century! 50 tasks done."
- 10-day streak: "Double digits! 10 days in a row."
- First session: "Welcome! Your journey starts here."

Display as a styled banner at the top of the insights view, only on the session that crosses the threshold.

**Impact:** High — emotional connection, surprise and delight.
**Effort:** Medium — threshold detection, banner rendering, persistence (don't show twice).
**SOUL alignment:** "Hey, look at all the things you did!" Must NOT feel like gamification pressure.

**CAUTION:** This idea is closest to the SOUL.md "no gamification" boundary. Key distinction: milestones celebrate what happened organically, not what the user should do. No "only 3 more tasks to reach gold level!" language. No badges, no levels, no unlockables.

---

## 4. SOUL.md Alignment Analysis

### Green Light (Fully Aligned)

| Idea | Why It Aligns |
|------|--------------|
| Colorful sparklines | Pure visual delight, shows what happened |
| Styled panels | "Button feel" — satisfying UI response |
| Fun facts | "Hey, look at what you did!" tone |
| Heatmap | Shows patterns neutrally, no guilt for empty days |
| Bar charts | Visual comparison, not judgment |
| Animated counters | "Every interaction should feel deliberate" |
| Theme-matched colors | Consistent aesthetic, user chose their theme |

### Yellow Light (Needs Careful Framing)

| Idea | Risk | Mitigation |
|------|------|-----------|
| Streaks | Could feel like pressure to maintain | Frame as observation: "You've been at it 5 days" not "Don't break your streak!" |
| Milestones | Could feel like gamification | One-time celebration only, no badges/levels/unlocks, no "next milestone" messaging |
| Time-of-day patterns | Could feel prescriptive | Present as "here's when you tend to work" not "you should work at 9am" |
| Burnout indicators | Could feel like guilt/judgment | If shown, frame as caring: "You seem to be taking it easy lately — that's okay!" NOT "Warning: productivity declining" |

### Red Light (Do NOT Implement)

| Idea | Why It Violates SOUL.md |
|------|------------------------|
| Leaderboards | "Not trying to be everything to everyone" — no competitive elements |
| Productivity scores | "Not a productivity report" |
| Overdue task counts | "Not: 'You have 47 overdue tasks'" |
| Achievement badges/levels | "No gamification" |
| "You should..." recommendations | Prescriptive, not the "friend saying let's go" tone |
| Completion targets/goals | Creates pressure, not encouragement |

---

## 5. Recommended Approach (Phased)

### Phase 1: Visual Polish (2-3 stories, ~1 epic)
Quick wins that transform existing data into beautiful displays.

1. **Story: Colorful sparklines with gradient** (Idea 1A)
   - Apply Lipgloss color gradient to existing sparkline
   - Modify `sparkline()` and `renderCompletionTrends()` in `insights_view.go`
   - Estimated effort: 60-90 minutes

2. **Story: Styled insights dashboard layout** (Idea 1B)
   - Wrap each section in Lipgloss bordered panels
   - Apply theme-aware header styles
   - 2-column layout using `JoinHorizontal()`
   - Estimated effort: 90-120 minutes

3. **Story: Fun facts generator** (Idea 1C)
   - New function `GenerateFunFact(analyzer, counter)` in `internal/core/`
   - Aggregate all-time stats: total doors, total tasks, fastest start, best day, etc.
   - Display one random fact per insights view render
   - Estimated effort: 60-90 minutes

### Phase 2: New Visualizations (3-4 stories, ~1 epic)
Add new chart types and richer data displays.

4. **Story: Horizontal bar charts for mood and time-of-day** (Ideas 2B + 2C)
   - Proportional fill bars with mood-specific colors
   - 24-hour activity bar for time-of-day
   - Estimated effort: 90-120 minutes

5. **Story: Activity heatmap** (Idea 2A)
   - GitHub-style 7x8 grid (8 weeks of daily data)
   - Color intensity = completion count
   - Option: use ntcharts heatmap component or build custom
   - Estimated effort: 120-180 minutes (evaluate ntcharts dependency vs in-tree)

6. **Story: Animated stats reveal** (Idea 2D)
   - `tea.Tick`-based counting animation on insights view entry
   - Counters tick up over ~500ms, sparkline draws left-to-right
   - Estimated effort: 120-150 minutes

7. **Story: Surface hidden metrics** (Gap Analysis)
   - Add "Session Highlights" section: avg session duration, fastest start time, detail view engagement
   - Show streak history: current + longest + average
   - Estimated effort: 90-120 minutes

### Phase 3: Thematic Integration (2-3 stories, optional)
Deep integration with the door metaphor and theme system.

8. **Story: Theme-matched stats colors** (Idea 3B)
   - Extend `ThemeColors` with stats palette
   - Apply in all stats rendering functions
   - Estimated effort: 90-120 minutes

9. **Story: Milestone celebrations** (Idea 3C, careful implementation)
   - Detect threshold crossings, show one-time banner
   - Persist "seen milestones" to avoid repeat display
   - Estimated effort: 120-150 minutes

10. **Story: "Trophy Room" door metaphor** (Idea 3A, stretch goal)
    - Stats view as thematic "fourth door" or behind-the-doors space
    - Animation on entry
    - Estimated effort: 180-240 minutes

### Dependency: ntcharts Evaluation

Before Phase 2, evaluate whether to add `NimbleMarkets/ntcharts` as a dependency:

**Pro:** Native Bubbletea integration, heatmap + bar chart + sparkline out of the box, actively maintained.
**Con:** New dependency (SOUL.md: "a little copying is better than a little dependency"), API surface to learn.
**Recommendation:** Evaluate for the heatmap only. The existing in-tree sparkline works well. Bar charts are simple enough to build with Lipgloss. Heatmap is the one chart type that's complex enough to justify a dependency.

---

## 6. ASCII Mockups of Proposed Layouts

### Mockup A: Phase 1 Completed Insights View

```
+===================================================================+
|                    YOUR INSIGHTS DASHBOARD                         |
+===================================================================+

+---------- COMPLETION TRENDS (7 Days) ---------+  +---- STREAKS ----+
|                                                |  |                 |
|  Mon  Tue  Wed  Thu  Fri  Sat  Sun             |  |  Current: 5 day |
|   |    |    |    |    |    |    |               |  |  Longest: 12    |
|   3    5    2    7    8    1    4               |  |  Average: 4.2   |
|                                                |  |                 |
|  This week: 30  |  Last: 22  |  ^ 36%          |  +-----------------+
+------------------------------------------------+

+---------- MOOD & PRODUCTIVITY ----------------+  +- DOOR PICKS ----+
|                                                |  |                 |
|  Focused    [====================]  4.2 avg    |  |  L: 40%         |
|  Energized  [================   ]  3.8 avg    |  |  C: 25%         |
|  Calm       [==========         ]  2.1 avg    |  |  R: 35%         |
|                                                |  |                 |
|  Best mood: Focused (4.2 tasks/session)        |  |  Slight left    |
+------------------------------------------------+  |  preference     |
                                                    +-----------------+

  * You've opened 247 doors since your first session!

  Press Esc to return
```

### Mockup B: Phase 2 with Heatmap

```
+===================================================================+
|                    YOUR INSIGHTS DASHBOARD                         |
+===================================================================+

+---------- ACTIVITY (Last 8 Weeks) ----------------------------------------+
|     W1   W2   W3   W4   W5   W6   W7   W8                                 |
|  Mo  .    #    ##   ##   #    ##   #    .                                   |
|  Tu  #    #    .    #    ##   #    .    #                                   |
|  We  ##   #    ##   #    #    .    #    ##                                  |
|  Th  .    .    #    .    #    #    #    .                                   |
|  Fr  #    ##   #    #    .    #    ##   #                                   |
|  Sa  .    .    .    .    .    .    .    .                                    |
|  Su  .    .    .    .    .    .    .    .                                    |
|                                                                            |
|  Legend:  [space]=0   .=1-2   #=3-4   ##=5+                                |
+----------------------------------------------------------------------------+

+---------- COMPLETION SPARKLINE -----------+  +---- SESSION HIGHLIGHTS -----+
|                                           |  |                             |
|  Mon  Tue  Wed  Thu  Fri  Sat  Sun        |  |  Total doors: 247           |
|   _    |    _    |    |    _    |          |  |  Total tasks: 143           |
|   3    5    2    7    8    1    4          |  |  Avg session: 4.2 min       |
|                                           |  |  Fastest start: 0.3s        |
|  ^ 36% vs last week                      |  |  Peak hour: 10am            |
+-------------------------------------------+  +-----------------------------+

  * Wednesday is your power day — avg 6.2 tasks!

  Press Esc to return
```

### Mockup C: Phase 3 "Trophy Room" Concept

```
  +------+  +------+  +------+
  | Door |  | Door |  | Door |  <-- normal door view
  |  1   |  |  2   |  |  3   |
  +------+  +------+  +------+

  Press I for insights...

  +----- The doors slide apart to reveal: -----+
  |                                             |
  |         ~ YOUR TROPHY ROOM ~                |
  |                                             |
  |   [sparkline]        [heatmap grid]         |
  |   [mood bars]        [fun fact]             |
  |                                             |
  |   ** 100 DOORS OPENED! **                   |
  |   Century Club unlocked on March 5          |
  |                                             |
  +---------------------------------------------+
```

---

## 7. Suggested Epic/Story Structure

### Epic NN: Beautiful Stats Display

**Goal:** Transform the insights dashboard from plain text into a visually delightful, SOUL-aligned celebration of user activity.

**Prerequisites:** None (all data infrastructure exists from Epics 1 and 4).

| Story | Title | Phase | Effort Est. | Priority |
|-------|-------|-------|-------------|----------|
| NN.1 | Colorful sparkline with gradient colors | 1 | 60-90 min | P1 |
| NN.2 | Styled insights dashboard with bordered panels | 1 | 90-120 min | P1 |
| NN.3 | Fun facts generator from session history | 1 | 60-90 min | P1 |
| NN.4 | Horizontal bar charts for mood & time-of-day | 2 | 90-120 min | P1 |
| NN.5 | GitHub-style activity heatmap | 2 | 120-180 min | P2 |
| NN.6 | Animated counter reveal on insights entry | 2 | 120-150 min | P2 |
| NN.7 | Surface hidden metrics (session highlights) | 2 | 90-120 min | P1 |
| NN.8 | Theme-matched stats color palettes | 3 | 90-120 min | P2 |
| NN.9 | Milestone celebrations (careful SOUL alignment) | 3 | 120-150 min | P2 |
| NN.10 | "Trophy Room" thematic integration | 3 | 180-240 min | P2 |

**Dependency decision:** Before NN.5, evaluate `ntcharts` as a dependency for the heatmap. Record decision in `docs/decisions/BOARD.md`.

---

## 8. Rejected Approaches (with Rationale)

### termdash as visualization framework
**Why rejected:** Incompatible architecture. termdash has its own event loop that conflicts with Bubbletea. Would require abandoning the entire TUI framework. ntcharts is the correct choice for Bubbletea-native visualization.

### Full charting library (go-echarts, etc.)
**Why rejected:** These are web-oriented libraries that render to HTML/SVG. Not suitable for terminal rendering. Terminal visualization must use Unicode characters and ANSI colors.

### Achievement/badge system
**Why rejected:** SOUL.md explicitly says "no gamification, no guilt." Badges create extrinsic motivation pressure. Milestones (Idea 3C) are acceptable only because they're one-time observations, not collectibles.

### Productivity scores/grades
**Why rejected:** "Not a productivity report." Assigning scores implies judgment. The ThreeDoors tone is "hey, look at what you did!" not "here's how you rate."

### Daily/weekly goal targets
**Why rejected:** Creates pressure to hit targets. SOUL.md: "Progress over perfection." Every completed task is a win; there's no such thing as falling short.

### Comparative analytics (this week you were less productive than...)
**Why rejected:** Self-comparison over time is acceptable ("3 tasks today vs 5 yesterday — every one counts!"), but framing as "less productive" violates the encouraging tone.

---

## Sources

Research conducted on available libraries and prior art:

- [ntcharts (NimbleMarkets) — Bubbletea-native charts](https://github.com/NimbleMarkets/ntcharts)
- [asciigraph — Lightweight ASCII line graphs for Go](https://github.com/guptarohit/asciigraph)
- [Lipgloss — Terminal styling with gradients and color manipulation](https://github.com/charmbracelet/lipgloss)
- [Bubbles — TUI components for Bubbletea](https://github.com/charmbracelet/bubbles)
- [Bubbletea — TUI framework](https://github.com/charmbracelet/bubbletea)
- [spotify-tui — Terminal Spotify client (theming reference)](https://github.com/Rigellute/spotify-tui)
- [awesome-tuis — Curated list of TUI projects](https://github.com/rothgar/awesome-tuis/)
- [Lazygit — TUI Git client (design patterns)](https://jesseduffield.com/Lazygit-5-Years-On/)
- [GOTUI — Go terminal dashboard with multiple chart types](https://github.com/metaspartan/gotui)
- [chartli — CLI charts including heatmap with Unicode blocks](https://github.com/ahmadawais/chartli)
