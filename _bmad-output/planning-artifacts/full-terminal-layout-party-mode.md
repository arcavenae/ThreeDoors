# Full-Terminal Vertical Layout — Party Mode Consensus

> **Status: Not Yet Planned**
> Research and party mode merged (PR #329) but no epic or stories created yet. See also `full-terminal-layout-research.md`.

> Party Mode: 2026-03-09
> Participants: Sally (UX Designer), Winston (Architect), Amelia (Dev), John (PM)
> Topic: Full-terminal vertical layout for ThreeDoors TUI

## Context

ThreeDoors renders content-driven height without filling the terminal. The app doesn't use AltScreen, help uses a hardcoded 20-line page, and most views don't receive height information. Research identified this as the primary reason the app feels like a "demo" rather than a "product."

## Decisions Summary

| # | Decision | Rationale | Alternatives Rejected |
|---|----------|-----------|----------------------|
| 1 | Use `tea.WithAltScreen()` | Standard TUI pattern; clean terminal lifecycle; app "owns" the screen | Stay in normal buffer — scrollback not useful for task picker |
| 2 | Fixed header + Flex middle + Fixed footer layout model | Proven pattern (lazygit, k9s, soft-serve); clean separation of concerns | Proportional zones — over-complicated for the need; Content-driven — current approach, leaves dead space |
| 3 | Door height: `min(max(10, available * 0.5), 25)` with cap | Door metaphor requires proportions that feel like doors (D-031); prevents skyscraper doors on tall terminals | Unlimited growth — absurd proportions; Fixed height — ignores terminal size entirely |
| 4 | Padding split 40% top / 60% bottom | Perceptual centering — content slightly above mathematical center feels natural (used by every major OS for dialogs) | 50/50 — feels too low; Flush top — wastes bottom space |
| 5 | Help/stats/search use full available height | Hardcoded 20 lines is the most glaring current bug; help content is ~50 lines, only showing 20 is broken | Keep fixed page size — wasteful and frustrating on tall terminals |
| 6 | Breakpoint degradation (<10, 10-15, 16-24, 25-40, 40+) | Invisible to users; progressive collapse without error messages; respects existing D-094 bar hiding | "Terminal too small" warnings — hostile UX; No degradation — broken on small terminals |
| 7 | Two-story implementation (MVP + refactor) | MVP delivers 80% of value; refactor is clean architectural separation | Single large story — too much scope risk; Three stories — over-decomposed |
| 8 | Layout work before/concurrent with Story 39.2 | Layout engine provides the footer slot that keybinding bar needs | Bar before layout — would need rework when layout engine arrives |

## Detailed Discussion

### AltScreen (Unanimous YES)

All participants agreed. Sally: "When you open ThreeDoors, you're making a decision: 'I'm going to look at my tasks now.' That deserves a full-screen commitment." Winston: "One-line change. Zero risk." Amelia confirmed: `tea.NewProgram(model, tea.WithAltScreen())` at `main.go:173`.

No participant saw value in preserving scrollback for a task-picker TUI.

### Layout Architecture

Winston proposed the standard three-zone model:

```
┌─ FIXED: Header region (2-3 lines) ─────────────────┐
│ Status bar, greeting                                 │
├─ FLEX: Main content region (fills remaining) ────────┤
│ Doors (centered vertically within this region)       │
│ Status indicators below doors                        │
│                                                      │
│ (breathing room / padding)                           │
│                                                      │
├─ FIXED: Footer region (2-3 lines) ──────────────────┤
│ Keybinding bar (Epic 39), footer message             │
└──────────────────────────────────────────────────────┘
```

Implementation requires a `layoutFull(header, content, footer string) string` function in MainModel that pads output to exactly `m.height` lines.

### Door Sizing Philosophy

Sally's museum analogy drove consensus: "A painting doesn't get bigger because the room is bigger. The whitespace around it becomes more generous, and the art feels more important."

Door height formula: `min(max(10, availableHeight * 0.5), 25)`
- 24-line terminal → compact 10-line doors
- 40-line terminal → comfortable 16-line doors
- 100-line terminal → capped 25-line doors with generous padding

The current formula (`0.6 * full terminal height`) has two problems:
1. Uses full terminal height instead of available height after header/footer
2. No upper cap — produces absurd 60+ line doors on large terminals

### Space Below Doors

Consensus: Don't add new content below the doors. The space below doors should contain:
- Existing status indicators (completion count, conflicts, proposals, sync bar)
- Padding/breathing room
- Footer anchored at terminal bottom

Sally: "Filling space ≠ filling it with stuff. The constraint IS the feature." This aligns directly with SOUL.md: "Show less. Resist the urge to add just one more option."

### Graceful Degradation

| Height | Behavior |
|--------|----------|
| < 10 | Minimal: doors only, no header/footer, no keybinding bar |
| 10-15 | Compact: 1-line header, doors at min 10, 1-line footer |
| 16-24 | Standard: full header, proportional doors, footer with bar |
| 25-40 | Comfortable: breathing room appears |
| 40+ | Spacious: doors capped at 25, generous padding |

Key principle: degradation should be invisible. No error messages or warnings.

### Epic 39 Interaction

D-088 states "Bar rendered by MainModel, views unaware." The layout engine's footer zone is exactly where the bar lives. Building the layout engine first means Story 39.2 (keybinding bar) drops into the footer slot trivially. If done in reverse, Story 39.2 must build its own footer anchoring that gets replaced.

**Recommendation: This layout work is a natural prerequisite for Story 39.2.**

### Architectural Refactoring Need

Currently, DoorsView.View() renders everything — header, doors, and footer — in one method. The layout engine needs these separated:
- `DoorsView.RenderDoors()` — just the three doors
- `DoorsView.RenderStatusSection()` — completion count, conflicts, proposals
- Header (greeting, time context) and footer (help text, footer message) move to MainModel

This refactor is in Story B (follow-up), not MVP. MVP can work by having MainModel pad the DoorsView output to fill height without extracting sub-components.

### Test Impact

John flagged: AltScreen may affect `teatest` test infrastructure. However, golden snapshot tests in `testdata/` test `View()` output, not terminal behavior, so they should still pass. Verification needed.

## MVP Scope (Story A)

1. Add `tea.WithAltScreen()` to program initialization
2. Build layout engine in MainModel (`layoutFull` function)
3. Cap door height at 25 lines, change proportion to 0.5 of available height
4. Vertical centering of doors in available space (40/60 padding split)
5. Help view uses terminal height instead of hardcoded 20
6. Propagate `SetHeight()` to all views that don't have it

## Follow-up Scope (Story B)

1. Extract header/footer from DoorsView into MainModel
2. Graceful degradation breakpoints
3. All secondary views (stats, search, sync log, etc.) fill terminal height
4. View-specific height optimization

## Key Files to Modify

| File | Change | Story |
|------|--------|-------|
| `cmd/threedoors/main.go:173` | Add `tea.WithAltScreen()` | A |
| `internal/tui/main_model.go` | Layout engine, height propagation | A |
| `internal/tui/doors_view.go` | Door height cap, proportion change | A |
| `internal/tui/help_view.go` | Dynamic page size from height | A |
| `internal/tui/doors_view.go` | Extract header/footer rendering | B |
| All view files | Fill available height | B |
