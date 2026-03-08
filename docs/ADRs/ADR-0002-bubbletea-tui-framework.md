# ADR-0002: Bubbletea TUI Framework

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #6, #7

## Context

ThreeDoors requires a terminal UI framework to render the "three doors" selection interface, task detail views, status menus, and various interactive flows. The framework must support:
- Multi-view navigation (doors view, detail view, onboarding, etc.)
- Keyboard-driven interaction
- ANSI color rendering for door themes
- Responsive layout based on terminal size
- Testability of UI components

## Considered Options

1. **Bubbletea** (charmbracelet) — Elm Architecture (MVU) for Go terminals
2. **tview** — Rich widget-based TUI library
3. **termbox-go** — Low-level terminal library
4. **tcell** — Another low-level terminal library

## Decision

Use **Bubbletea** with **Lipgloss** for styling and **Bubbles** for pre-built components.

## Rationale

- Elm Architecture (Model-View-Update) enforces clean state management
- Lipgloss provides declarative styling without raw ANSI escape codes
- Bubbles provides text input, list selection, and other components
- `tea.Cmd` mechanism supports async operations (file I/O, timers, sync) cleanly
- `teatest` package enables headless TUI testing (used in Epic 18)
- Active ecosystem with frequent updates

## Consequences

### Positive
- All UI output flows through `View()` — no accidental `fmt.Println`
- `Update()` must remain non-blocking — forces good async patterns via `tea.Cmd`
- Testable via `teatest` headless harness and golden file snapshots
- Clean separation of concerns (model state vs. rendering)

### Negative
- MVU pattern has a learning curve
- Complex multi-view navigation requires custom model composition
- No built-in layout system — Lipgloss layout is manual
- `tea.Cmd` batching for concurrent async operations requires care

### Rules Established
- All user-visible output goes through Bubbletea `View()` — never `fmt.Println`
- Use Lipgloss for styling — never ANSI escape codes directly
- Keep `Update()` fast — no blocking I/O in the update loop
- Use `tea.Cmd` for async operations (file I/O, timers)
