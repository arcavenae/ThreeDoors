# Investigation: `:dashboard` Command Exits App Immediately

**Date:** 2026-03-09
**Reporter:** User via supervisor
**Severity:** UX Bug (Medium)

## Summary

The `:dashboard` command sometimes exits the app instead of showing the insights dashboard. The inconsistency is caused by the universal quit handler (Story 36.3, PR #276) capturing the `q` key in ViewInsights, where users expect it to go back.

## Root Cause

**The universal `q` quit handler (line 910-913 of `main_model.go`) fires in ViewInsights because `isTextInputActive()` returns `false` for that view.**

### Flow Analysis

1. User presses `:` in doors view → opens search view with `:` prefilled
2. User types `dashboard` and presses Enter
3. `search_view.executeCommand()` returns `func() tea.Msg { return ShowInsightsMsg{} }`
4. Search input is cleared, `isCommandMode` set to false
5. Bubbletea executes the cmd → `ShowInsightsMsg` arrives at `main_model.Update`
6. Handler at line 352 creates `InsightsView`, sets `viewMode = ViewInsights`
7. **Now in ViewInsights**: pressing `q` triggers the universal quit handler

### Why It's Inconsistent

- **"Works"**: User presses `Esc` to leave insights → `ReturnToDoorsMsg` → returns to doors view
- **"Exits"**: User presses `q` to leave insights (expecting "go back") → `RequestQuitMsg` → app exits

The `q` key is NOT listed in `insightsBindings()` (only `esc: back` and `?: help` are shown), so users have no indication that `q` will quit the entire application.

### Timeline

| Commit | Story | What it did |
|--------|-------|-------------|
| `2e707a9` | Story 4.5 | Added InsightsView and `:dashboard` command |
| `59f0855` | Story 36.3 | Added universal `q` quit from all non-input views |

Story 36.3 retroactively changed the behavior of `q` in InsightsView. Before 36.3, pressing `q` in insights was a no-op (fell through to `updateInsights` which returned nil). After 36.3, `q` exits the app.

## Affected Views

This same issue applies to ALL non-text-input views where the user might expect `q` to go back rather than quit:

- `ViewInsights` — no `q` handler, universal quit fires
- `ViewHealth` — same
- `ViewSyncLog` — same
- `ViewNextSteps` — same
- `ViewAvoidancePrompt` — same

Only `ViewDoors` explicitly handles `q` as quit (line 977), which is the correct behavior since it's the root view.

## Proposed Fix Options

### Option A: Make `q` go back in sub-views (Recommended)

In `isTextInputActive()` or a new `isQuitAllowed()` check, only allow `q` to quit from `ViewDoors`. All other views should treat `q` as "go back" (equivalent to Esc).

**Pros:** Consistent UX — `q` only quits from the main doors view. Sub-views use `q` as back.
**Cons:** Changes behavior from Story 36.3's intent. Need to verify no sub-views already use `q` for something else.

### Option B: Add `q` as explicit keybinding in InsightsView

Modify `InsightsView.Update()` to handle `q` as a "go back" action (same as Esc), and update the universal quit handler to skip views that handle `q` themselves.

**Pros:** Targeted fix, doesn't affect other views.
**Cons:** Doesn't fix the same issue in other sub-views.

### Option C: Change universal quit to only fire in ViewDoors

Move the `q` quit handler from the universal position (line 910) into `updateDoors` only (it's already there at line 977, making the universal handler redundant for doors).

**Pros:** Clean separation — quit from doors, back from everywhere else.
**Cons:** Most dramatic change, needs careful testing.

## Recommendation

**Option A** is recommended. The universal quit handler should be scoped to `ViewDoors` only. Sub-views should treat `q` as "go back" (Esc equivalent). This matches user expectations in TUI applications where `q` typically means "close this panel" not "exit the program."

A new story should be created for the fix since it touches the keybinding architecture from Story 36.3.

## Files Involved

- `internal/tui/main_model.go` — lines 910-913 (universal quit), line 977 (doors quit), line 1184-1210 (`isTextInputActive`)
- `internal/tui/insights_view.go` — `Update()` only handles Esc
- `internal/tui/keybindings.go` — `insightsBindings()` doesn't list `q`
- `internal/tui/main_model_test.go` — line 930 `TestUniversalQuit_InsightsView_QKeyQuits` confirms current behavior

## Related

- Story 4.5 (User Insights Dashboard) — original implementation
- Story 36.3 (Universal Quit) — introduced the regression
- Epic 40 (Beautiful Stats) — planned dashboard improvements (Story 40.1 not started)
