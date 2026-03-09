# Party Mode Artifact: Beautiful Stats Feature Design

**Date:** 2026-03-08
**Participants:** John (PM), Winston (Architect), Sally (UX Designer), Amelia (Dev), Quinn (QA)
**Topic:** Finalize beautiful stats feature design for ThreeDoors
**Input:** Research report (14+ tracked dimensions, ntcharts evaluation, 3-phase plan), UX review

---

## Discussion Summary

### Round 1: Library Selection

**John (PM):** The research recommends ntcharts as a Bubbletea-native charting library. But SOUL.md says "a little copying is better than a little dependency." What's the real value-add vs building custom?

**Winston (Architect):** The answer is hybrid, but leaning heavily toward custom Lipgloss. Here's why: sparklines are already in-tree (8 lines of code). Bar charts are trivially built with `strings.Repeat()` + Lipgloss styling. The only chart type that justifies a dependency is the heatmap — and even that's achievable with Unicode shade characters and a nested loop. I'd defer the ntcharts evaluation to when we actually build the heatmap (Story 40.5), and make the dependency decision at that point. Don't add the dependency speculatively.

**Amelia (Dev):** Agree. `lipgloss.Blend1D()` handles gradient coloring. Bar fill is `strings.Repeat("█", fillWidth)` + `strings.Repeat("░", emptyWidth)`. No library needed for Phase 1. For the heatmap, I'd prototype both approaches (custom vs ntcharts) in a spike subtask within the story.

**Sally (UX):** From a design consistency standpoint, in-tree rendering gives us full control over spacing, padding, and alignment. Library components often have their own opinions about layout that fight with the host app. Custom is the right call for Phase 1.

**DECISION D-086:** Phase 1 uses custom Lipgloss rendering (no new dependencies). ntcharts evaluated during Phase 2 heatmap story only if custom approach proves insufficient. Default bias: build custom.

### Round 2: Dashboard Layout

**Sally (UX):** I recommended a single view for Phase 1, with Tab navigation to a detail view in Phase 2. The key principle: the main view must fit in 24 rows at 80 columns without scrolling.

**John (PM):** Single view for Phase 1 is right. But I want to push back slightly on adding Tab navigation in Phase 2 — that adds interaction complexity. Could the detail view be a separate command instead? Like `:insights detail` or pressing `d` within insights?

**Winston (Architect):** Tab is simpler to implement than a new command. It's one KeyMsg check in Update(). A new command means routing changes in MainModel. I'd keep Tab.

**Amelia (Dev):** `tea.KeyTab` is already handled nowhere in InsightsView. Clean addition. The view state is just a boolean: `showDetail bool`. Simple.

**DECISION D-087:** Phase 1 is a single non-scrollable view (no tabs). Phase 2 adds Tab/Shift-Tab navigation between Overview and Detail tabs. No new commands needed. View state tracked as `activeTab int` in InsightsView.

### Round 3: Phase 1 Scope

**John (PM):** What exactly ships in Phase 1? The research has three items: gradient sparklines, styled dashboard layout, fun facts. The UX review adds a "hero number" at the top. That's four things — is that three stories or four?

**Sally (UX):** The hero number is part of the dashboard layout, not a separate feature. It's a styled `fmt.Fprintf(&s, "  %s %d tasks completed since you started %s\n", star, total, star)` with Lipgloss color. One line of code within the layout story.

**Winston (Architect):** I'd structure Phase 1 as three stories:
1. **Dashboard shell** — Lipgloss bordered panels, 2-column layout, hero number, responsive breakpoints, navigation boilerplate (Esc to return, footer)
2. **Gradient sparklines** — modify existing `sparkline()` to return individually-styled runes using `Blend1D()`, update `renderCompletionTrends()`
3. **Fun facts engine** — new `FunFactGenerator` in `internal/core/`, random fact selection, display at bottom of dashboard

**Amelia (Dev):** Story 1 is the biggest — it restructures the entire View() method. Stories 2 and 3 are incremental on top. Dependency chain: 2 depends on 1, 3 depends on 1. Stories 2 and 3 can be parallel.

**DECISION D-088:** Phase 1 scope is exactly 3 stories: dashboard shell (39.1), gradient sparklines (39.2), fun facts engine (39.3). Stories 40.2 and 40.3 depend on 40.1 but not on each other.

### Round 4: Fun Facts — Tone and Content

**John (PM):** SOUL.md is crystal clear: "hey look what you did!" not "you have 47 overdue tasks." Fun facts must celebrate, never guilt. What are the content guidelines?

**Sally (UX):** Every fun fact should pass the "friend test" — would a supportive friend say this? Three rules:
1. **Observe, don't prescribe.** "You complete the most tasks on Wednesdays" (good). "You should work more on Mondays" (bad).
2. **Celebrate totals, not rates.** "You've opened 247 doors!" (good). "Your completion rate dropped 15%" (bad).
3. **Frame gaps as potential.** "You haven't tried logging a mood yet — it unlocks new insights!" (good). "You're missing mood data" (bad).

**Winston (Architect):** I'd add a fourth rule: **No comparisons to past self that imply decline.** "This week you completed 30 tasks" is fine. "This week you completed 30 vs 45 last week" is borderline — the sparkline already shows the trend visually.

**Quinn (QA):** How do we test tone? I'd propose a test that verifies every fun fact string contains no negative words from a banned list: "declined", "dropped", "worse", "less", "failed", "missing", "behind", "overdue".

**DECISION D-089:** Fun facts follow four content rules: (1) observe not prescribe, (2) celebrate totals not rates, (3) frame gaps as potential, (4) no decline comparisons. QA validates with a banned-words test. Fun facts rotate on each view entry using a deterministic-random selection (seed from current date so same fact shown throughout a day, changes next day).

### Round 5: Heatmap

**Winston (Architect):** The heatmap is the most complex visualization. It aggregates daily completion data across 8 weeks into a 7x8 grid. The rendering itself isn't hard — it's a nested loop with shade characters. The data aggregation is the real work: you need to iterate all sessions, bucket by day-of-week and week-number, and find max for normalization.

**Amelia (Dev):** PatternAnalyzer already has `GetDailyCompletions(n)` which returns a `map[string]int` of date->count. Extending that to 56 days (8 weeks) is trivial. The rendering is straightforward: 7 rows, 8 columns, each cell styled with `Blend1D()` based on count/max.

**Sally (UX):** 8 weeks is the right range — long enough to show patterns, short enough to render in 80 columns. The UX review confirmed: use Unicode shade characters (space, ░, ▒, ▓, █) as shape encoding, with color as redundant reinforcement. This ensures accessibility.

**John (PM):** Worth the complexity? Yes — it's an instantly recognizable visualization. But it belongs in Phase 2, not Phase 1. The heatmap goes on the Detail tab.

**DECISION D-090:** Activity heatmap is Phase 2 (Story 40.5), 8-week range, 7x8 grid, Unicode shade characters with color reinforcement. Lives on the Detail tab. Custom implementation using `GetDailyCompletions(56)` + `Blend1D()`. ntcharts evaluated only if custom approach has rendering issues.

### Round 6: Animated Counters

**Sally (UX):** Subtle, not flashy. 300-500ms total animation duration. Numbers tick up smoothly from 0 to final value. The sparkline does NOT animate (it appears fully formed) — only numeric values animate. If the user has seen the view before in this session, skip the animation (only on first entry).

**Winston (Architect):** This requires animation state in InsightsView: `animating bool`, `animationProgress float64`, `animationStarted time.Time`. On first entry, start a `tea.Tick(30ms)` that increments progress until 1.0. Each render multiplies displayed values by progress. After animation completes, set `animating = false` and stop ticking.

**Amelia (Dev):** 30ms tick at 500ms duration = ~16 frames. That's smooth enough. The tick stops when animation completes — no ongoing CPU usage.

**John (PM):** This is Phase 2 (Story 40.7). Nice polish but not essential for the visual upgrade in Phase 1.

**DECISION D-091:** Animated counters are Phase 2 (Story 40.7). Subtle: 300-500ms, ~16 frames at 30ms tick. Only numeric hero values animate. Sparkline and bars render immediately. Animation plays once per view entry (skipped on re-entry within same session). State tracked via `animating bool` and `animationProgress float64`.

### Round 7: Theme Integration

**Sally (UX):** Phase 1 uses an independent color palette (blue-teal-yellow gradient for sparklines, mood-specific bar colors). Theme integration is Phase 3 because it requires extending `ThemeColors` struct, which is a cross-cutting change across all theme files.

**Winston (Architect):** Correct. Adding a `StatsAccent lipgloss.Color` field to `ThemeColors` is the minimal extension. Each theme maps its personality: Classic=warm amber, Modern=cool teal, Sci-Fi=neon green, Shoji=earth tones. The `InsightsView` receives the active theme and uses `StatsAccent` for panel borders and hero number color.

**DECISION D-092:** Phase 1 uses independent palette (no theme coupling). Phase 3 (Story 40.8) extends `ThemeColors` with `StatsAccent` and `StatsGradientStart`/`StatsGradientEnd` fields. Each theme provides its own stats color personality. InsightsView receives theme via constructor injection.

### Round 8: Trophy Room and Milestones

**John (PM):** The UX review recommends deferring Trophy Room and reducing milestones to 3-4 observations. I agree. The Trophy Room adds architectural complexity (new view state, door-sliding animation) for uncertain payoff. And the "trophy" naming has gamification connotations that conflict with SOUL.md.

**Sally (UX):** Trophy Room is deferred. Milestones are in, but with strict framing: observations, not achievements. No "unlocked" language. No "next milestone" messaging. Four milestones max:
1. First session: "Welcome! Your journey starts here."
2. 50 tasks completed: "50 tasks done — half a century of getting things done!"
3. 100 tasks completed: "Triple digits! 100 tasks completed."
4. 10-day streak: "10 days in a row — double digits!"

**Winston (Architect):** Milestones need persistence — a `~/.threedoors/milestones.json` file tracking which milestones have been shown. Otherwise, the user sees "100 tasks completed!" every time they open insights after crossing 100.

**Quinn (QA):** Test cases: milestone shows exactly once, milestone file persists across sessions, no milestone shown if threshold not crossed, milestone banner dismisses automatically after 5 seconds or on keypress.

**DECISION D-093:** Trophy Room deferred indefinitely (too complex, gamification risk). Milestones are Phase 3 (Story 40.9), limited to 4 thresholds, observation language only. Persistence via `milestones.json`. Banner auto-dismisses after 5 seconds.

### Round 9: Epic Number Assignment

**John (PM):** ROADMAP.md shows Epics 0-38 allocated. "Epic 40+: Advanced Features" is a placeholder. This feature gets **Epic 40: Beautiful Stats Display**.

**DECISION D-094:** Epic number is 39. 10 stories total across 3 phases.

---

## Decisions Summary

| ID | Decision | Adopted Approach | Rejected Alternative | Rationale |
|----|----------|-----------------|---------------------|-----------|
| D-086 | Visualization library | Custom Lipgloss (Phase 1), ntcharts evaluated for heatmap only (Phase 2) | ntcharts for all charts; termdash; go-echarts | SOUL.md: "a little copying is better than a little dependency"; sparklines/bars are trivial to build; full control over layout |
| D-087 | Dashboard layout | Single view (Phase 1), Tab navigation (Phase 2) | Scrollable single view; drill-down commands; separate views | Fits 24 rows at 80 cols; Tab is simplest interaction; no routing changes needed |
| D-088 | Phase 1 scope | 3 stories: dashboard shell, gradient sparklines, fun facts | 4+ stories; combined single story | Clean dependencies; parallelizable; right-sized |
| D-089 | Fun facts tone | 4 content rules + banned-words QA test + daily-seeded rotation | Fully random; session-seeded; no guidelines | SOUL.md alignment; testable quality gate; consistent within a day |
| D-090 | Heatmap design | Phase 2, 8 weeks, custom Unicode+color, Detail tab | ntcharts heatmap; 4 weeks; 12 weeks; main view | 8 weeks fits 80 cols; custom for consistency; Detail tab keeps main view clean |
| D-091 | Animated counters | Phase 2, subtle 300-500ms, numbers only, once per entry | Flashy full-view animation; continuous animation; no animation | "Button feel" polish; not distracting; low CPU cost |
| D-092 | Theme integration | Independent palette (Phase 1), theme coupling (Phase 3) | Theme-coupled from start; always independent | Reduces Phase 1 scope; theme extension is cross-cutting |
| D-093 | Trophy room / milestones | Trophy room deferred; milestones in Phase 3 with 4 thresholds, observation language | Trophy room in Phase 3; no milestones; gamified milestones | SOUL.md boundary; trophy naming is gamification; 4 thresholds is sufficient |
| D-094 | Epic number | Epic 40 | N/A | Next available per ROADMAP.md |

## Rejected Approaches (with Rationale)

| Option | Why Rejected |
|--------|-------------|
| ntcharts as primary charting library | Adds dependency when Lipgloss can do the job; layout control concerns; evaluate for heatmap only |
| Trophy Room with door-sliding animation | High complexity, uncertain payoff, "trophy" has gamification connotations, stats should be easily accessible not hidden |
| Gamified milestones ("Century Club unlocked!") | SOUL.md: "no gamification, no guilt"; achievement language creates extrinsic motivation |
| Burnout indicators in TUI | Too judgmental for user-facing display; keep in MCP for AI agent consumption only |
| Productivity scores or grades | SOUL.md: "not a productivity report" |
| "You should work more on..." recommendations | Prescriptive, violates "friend saying let's go" tone |
| Braille pattern characters for high-res charts | Inconsistent rendering across terminals and fonts; Unicode blocks are universally safe |
| Scrollable single view instead of tabs | Users lose spatial context when scrolling; tabs provide clear information hierarchy |
