# Full-Terminal Vertical Layout — Research Findings

> Research conducted: 2026-03-09
> Scope: ThreeDoors TUI vertical height and full-terminal layout

## Problem Statement

ThreeDoors currently renders in a content-driven height regardless of terminal size. If the terminal is taller, the app doesn't expand to fill it. This results in:
- Wasted vertical space below the app
- Scrollable views like `:help` crammed into a fixed 20-line window
- The app doesn't feel like it "owns" the terminal

## Current State Analysis

### Terminal Initialization

- **File**: `cmd/threedoors/main.go:173`
- Program created with `tea.NewProgram(model)` — **no `tea.WithAltScreen()`**
- Runs in normal scrollback buffer, not alternate screen buffer
- This is unusual for a TUI app that wants to own the full terminal

### Window Size Handling

- **File**: `internal/tui/main_model.go:253-308`
- `tea.WindowSizeMsg` IS handled — width and height are stored and propagated
- Width propagated to: DoorsView, DetailView, MoodView, SearchView, HealthView, InsightsView, AddTaskView, ValuesView, FeedbackView, ImprovementView, NextStepsView, OnboardingView, ConflictView, SyncLogView, ThemePickerView, DevQueueView, HelpView
- **Height propagated to**: DoorsView and ProposalsView ONLY
- Most views have no concept of available height

### Door Height Calculation

- **File**: `internal/tui/doors_view.go:268-273`
- `doorHeight = int(float64(dv.height) * 0.6)` — 60% of full terminal height
- Minimum: 10 lines
- No maximum cap — doors can grow to 60+ lines on tall terminals
- Uses full terminal height, not available height after header/footer

### MainModel.View() Layout

- **File**: `internal/tui/main_model.go:1546-1640`
- Simple switch statement — returns whatever each view emits
- No padding, no height management, no vertical centering
- Output height = content height (variable, always less than terminal)

### DoorsView.View() Layout (doors_view.go:235-396)

Current top-to-bottom layout:
1. Header: "ThreeDoors - Technical Demo" (1 line)
2. Greeting message (1 line)
3. Multi-dimensional greeting — conditional (0-1 lines)
4. Time context badge — conditional (0-1 lines)
5. Blank line
6. Three doors rendered horizontally (doorHeight lines)
7. Completed count — conditional (0-2 lines)
8. Conflict notification — conditional (0-2 lines)
9. Proposal badges — conditional (0-2 lines)
10. Sync status bar — conditional (0-2 lines)
11. Help text line (1 line)
12. Footer message (1 line)

Total: ~doorHeight + 8-16 lines of non-door content

### Help View

- **File**: `internal/tui/help_view.go:10`
- `helpPageSize = 20` — hardcoded, completely ignores terminal height
- Has `SetWidth()` but no `SetHeight()`
- On a 50-line terminal, help still shows only 20 lines of content

### Keybinding Overlay (Reference Implementation)

- **File**: `internal/tui/keybinding_overlay.go:44`
- DOES use full terminal height: `innerHeight := height - 5`
- Scrollable with proper offset clamping
- This is the pattern to follow for other views

### Keybinding Bar

- **File**: `internal/tui/keybinding_bar.go`
- Height-aware: hidden below 10 lines, compact 10-15, full above 15
- Returns 2-line string (separator + bar)
- Needs stable anchor point at terminal bottom (per D-088)

## Comparison with Other Bubbletea Apps

### lazygit
- Uses AltScreen
- Fixed header + fixed footer + flex middle panels
- Content fills available space

### charm/soft-serve
- Uses AltScreen
- Header/content/footer layout
- Content region scrollable

### k9s
- Uses AltScreen
- Header bar + content panels + footer bar
- All panels fill available height

**Pattern**: Every serious Bubbletea TUI uses AltScreen + fixed-header/flex-content/fixed-footer.

## SOUL.md Alignment

- "Every Interaction Should Feel Deliberate" — dead space below the app is accidental, not deliberate
- "button feel" — the app should feel like it owns its space
- "Opening ThreeDoors should feel like a friend saying: Here are three things..." — a friend who only takes up half your screen doesn't command attention
- "Show less. Resist the urge to add just one more option." — filling space with whitespace, not more content

## Key Files

| File | Lines | Relevance |
|------|-------|-----------|
| `cmd/threedoors/main.go` | 173 | Program initialization (add AltScreen) |
| `internal/tui/main_model.go` | 253-308, 1546-1640 | WindowSizeMsg handler, View() layout |
| `internal/tui/doors_view.go` | 235-396 | Door rendering, header/footer embedded |
| `internal/tui/help_view.go` | 10, 213-244 | Hardcoded page size, View() rendering |
| `internal/tui/keybinding_overlay.go` | 44-151 | Reference implementation for full-height |
| `internal/tui/keybinding_bar.go` | 41 | Footer component, height breakpoints |
