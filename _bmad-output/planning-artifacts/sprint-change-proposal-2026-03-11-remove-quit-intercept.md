# Sprint Change Proposal: Remove Quit Intercept (Session Reflection Prompt)

**Date:** 2026-03-11
**Triggered by:** User report — quit action intercepted by "Session Reflection" improvement prompt, requiring a second keypress to actually exit. Violates SOUL.md core philosophy.
**Severity:** Medium

## Problem Statement

When a user presses `q` to quit ThreeDoors after a productive session (1+ task completions or 5+ minutes), the app intercepts the quit and displays a "Session Reflection" prompt (`ImprovementView`) asking "What's one thing you could improve about this list/task/goal right now?" The user must then press Enter or Esc to actually exit.

This is hostile UX. The user asked to quit. The app should quit. Intercepting quit to solicit feedback adds friction where the entire project philosophy demands reducing it.

## SOUL.md Violations

Three core principles are violated:

1. **"Every design decision should reduce friction to starting, not add friction for 'correctness.'"** (Progress Over Perfection) — The improvement prompt adds a friction step on every exit from a productive session, optimizing for data collection over user experience.

2. **"Would this reduce friction? If yes, do it. If it adds a step for the user, don't."** (Design Principles for AI Agents, #1) — The improvement prompt adds a mandatory step. By the project's own decision framework, it should not exist.

3. **"Opening ThreeDoors should feel like a friend saying: 'Hey, here are three things you could do right now.'"** (The Feeling We're Going For) — A friend doesn't block you when you're leaving the room to ask "wait, what would you improve?" That's a survey, not a friendship.

## Impact Analysis

- **User-facing:** Every productive session (1+ completions OR 5+ minutes) triggers the intercept
- **Code scope:** `internal/tui/improvement_view.go`, `internal/tui/main_model.go` (RequestQuitMsg handler), related messages and styles
- **Data loss:** None. Improvements.txt data collection stops. Existing data is unaffected.
- **Risk:** Zero — removing code, not adding it

## Proposed Approach

Remove the `ImprovementView` entirely. When `RequestQuitMsg` is received, always return `tea.Quit` immediately — no conditional check, no intercept view, no second keypress.

**Files to modify:**
- `internal/tui/main_model.go` — Simplify `RequestQuitMsg` handler to just `return m, tea.Quit`
- `internal/tui/main_model.go` — Remove `ViewImprovement` from ViewMode enum and all references
- **Files to delete:**
  - `internal/tui/improvement_view.go`
  - `internal/tui/improvement_view_test.go`
- **Files to clean up:**
  - `internal/tui/messages.go` — Remove `ImprovementSubmittedMsg`, `ImprovementSkippedMsg`
  - `internal/tui/styles.go` — Remove `improvementHeaderStyle` if unused elsewhere
  - `internal/core/improvement_writer.go` — Delete (dead code after removal)
  - `internal/core/improvement_writer_test.go` — Delete

## Rejected Alternatives

1. **Make the prompt optional/configurable** — Rejected. Adding a config toggle for a feature that violates the project's soul doesn't fix the philosophy violation; it just makes the violation opt-out instead of opt-in. If it shouldn't exist, remove it.

2. **Move the prompt to a non-blocking position** — Rejected. The user pressed quit. They want to quit. Any additional UI before quitting is friction, regardless of positioning.

3. **Reduce the frequency (only show after N sessions)** — Rejected. Less hostile is still hostile. The right frequency for an exit intercept in a friction-reduction app is zero.

## Stories Required

1. **Story 0.50: Remove Session Reflection Quit Intercept** — Remove `ImprovementView`, simplify `RequestQuitMsg` handler, delete dead code and tests. Update Story 3.6 status to note its reversal.

## Effort Estimate

**XS** — Straightforward code deletion with minimal test updates.
