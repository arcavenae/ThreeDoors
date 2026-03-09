# Architecture Review: Keybinding Display System

**Date:** 2026-03-08
**Reviewer:** Winston (Architect)
**Input:** Party mode consensus + UX designer review

---

## Architectural Fit Assessment

**Verdict: Clean fit.** The keybinding display system integrates naturally into the existing Bubbletea MVU architecture without requiring structural changes. All modifications are additive — no existing interfaces change, no existing packages are reorganized.

---

## Component Design

### 1. Keybinding Registry (`internal/tui/keybindings.go`)

A new file in the existing `internal/tui/` package. No new package needed — this is a TUI concern.

```go
// KeyBinding represents a single key-action mapping for display purposes.
type KeyBinding struct {
    Key         string // Display string: "a/w/d", "enter", "esc"
    Description string // Human-readable: "Select door", "Confirm"
    Priority    int    // 1=bar, 2=bar if space, 3=overlay only
}

// KeyBindingGroup organizes bindings by category for overlay display.
type KeyBindingGroup struct {
    Name     string       // "Navigation", "Actions", "Views", etc.
    Bindings []KeyBinding
}

// viewKeyBindings returns the keybinding groups for a given view mode.
// This is a pure function — all data is compile-time constant.
func viewKeyBindings(mode ViewMode) []KeyBindingGroup { ... }

// barBindings returns the priority-1 bindings for the bar, filtered by mode.
func barBindings(mode ViewMode) []KeyBinding { ... }

// allKeyBindingGroups returns all bindings across all views for the overlay.
func allKeyBindingGroups() []KeyBindingGroup { ... }
```

**Design decisions:**
- Pure functions, not a struct with methods. There's no state to manage.
- No interface needed. Only the TUI package consumes this.
- Data is hardcoded, not loaded from config. All bindings are known at compile time (D-KB-3 from party mode).
- Priority field enables the bar to select which keys to show without separate bar/overlay data structures.

### 2. Bar Component (`internal/tui/keybinding_bar.go`)

A stateless rendering function, not a Bubbletea model. The bar has no internal state — it receives the current ViewMode, terminal width, and toggle state, and returns a rendered string.

```go
// RenderKeybindingBar renders the concise keybinding bar for the given view.
// Returns empty string if disabled or terminal too small.
func RenderKeybindingBar(mode ViewMode, width int, height int, enabled bool) string { ... }
```

**Rendering logic:**
1. If `!enabled` or `height < 10`: return `""`
2. Get `barBindings(mode)`
3. Format as `key Description   key Description   ...`
4. If `width < 40`: show only `? Help`
5. If total formatted width > terminal width: truncate from right, always keep `?` last
6. Prepend separator line (`─` repeated to width)
7. Apply dim Lipgloss styling

**Height accounting:**
The bar consumes 2 lines (separator + bindings). When the bar is visible, MainModel must subtract 2 from the height passed to view components. This is critical — views that use height for door rendering (DoorsView) must receive the correct available height.

### 3. Overlay Component (`internal/tui/keybinding_overlay.go`)

A stateless rendering function with optional scroll state.

```go
// OverlayState holds the scroll position for the keybinding overlay.
type OverlayState struct {
    ScrollOffset int
    ViewMode     ViewMode // Current view when overlay was opened
}

// RenderKeybindingOverlay renders the full-screen keybinding reference.
func RenderKeybindingOverlay(state OverlayState, width int, height int) string { ... }
```

**Rendering logic:**
1. Get `allKeyBindingGroups()`
2. Sort so the current ViewMode's group appears first (context highlighting)
3. Render bordered box using Lipgloss
4. Apply scroll offset if content exceeds height
5. Fixed footer: "Press ? or esc to close   ↑/↓ to scroll"

### 4. MainModel Integration

Changes to `internal/tui/main_model.go`:

```go
type MainModel struct {
    // ... existing fields ...
    showKeybindingBar     bool          // Toggle state
    showKeybindingOverlay bool          // Overlay visible
    overlayState          OverlayState  // Scroll position
}
```

**Update() changes:**
- Global `?` handler (at MainModel level, before view dispatch):
  - If overlay shown: dismiss overlay
  - If overlay not shown: show overlay
- Global `h` handler (at MainModel level):
  - Toggle `showKeybindingBar`
  - Persist to config.yaml (via tea.Cmd to avoid blocking)
- When overlay is shown: intercept all keys except `?`, `esc`, `↑/↓/j/k` (scroll)
- Guard: skip `?` and `h` handlers when `isTextInputActive()` returns true

**View() changes:**
```go
func (m *MainModel) View() string {
    if m.showKeybindingOverlay {
        return RenderKeybindingOverlay(m.overlayState, m.width, m.height)
    }

    // Calculate available height for content
    contentHeight := m.height
    barOutput := ""
    if m.showKeybindingBar {
        barOutput = RenderKeybindingBar(m.viewMode, m.width, m.height, true)
        if barOutput != "" {
            contentHeight -= 2 // separator + bar
        }
    }

    // Render current view with adjusted height
    viewOutput := m.renderCurrentView(contentHeight)

    if barOutput != "" {
        return viewOutput + "\n" + barOutput
    }
    return viewOutput
}
```

### 5. Config Persistence

Extend the existing `config.yaml` schema:

```yaml
show_keybinding_bar: true  # default true for new installs
```

The config write happens asynchronously via `tea.Cmd` when the user presses `h`. This follows the existing pattern used by theme persistence (Epic 17).

---

## Impact on Existing Code

### Files Modified

| File | Change | Risk |
|------|--------|------|
| `internal/tui/main_model.go` | Add fields, Update() handlers, View() wrapping | Medium — core file, needs careful integration |
| `internal/tui/styles.go` | Add bar styling constants | Low — additive only |

### Files Created

| File | Purpose |
|------|---------|
| `internal/tui/keybindings.go` | Registry data + pure functions |
| `internal/tui/keybindings_test.go` | Registry tests |
| `internal/tui/keybinding_bar.go` | Bar rendering |
| `internal/tui/keybinding_bar_test.go` | Bar rendering tests |
| `internal/tui/keybinding_overlay.go` | Overlay rendering |
| `internal/tui/keybinding_overlay_test.go` | Overlay rendering tests |

### Files NOT Modified

- No changes to individual view files (DoorsView, DetailView, etc.)
- No changes to `internal/core/` or any domain logic
- No changes to `internal/tui/themes/` — bar is theme-independent
- No changes to CLI commands

---

## Performance Analysis

### Bar Rendering Overhead

The bar renders on every `View()` call. Assessment:
- `barBindings()` iterates a small slice (~6 items per view). Cost: negligible.
- String formatting with Lipgloss. Cost: microseconds.
- This is the same order of magnitude as the existing footer message in DoorsView.

**Verdict: No performance concern.**

### Overlay Rendering

The overlay only renders when `showKeybindingOverlay == true`. When not shown, cost is zero (early return in View()). When shown, it renders a static document. Cost: trivially small.

**Verdict: No performance concern.**

### Height Recalculation

Subtracting 2 from height when bar is visible changes the height passed to views. DoorsView uses height for door proportions (Epic 35). The 2-line reduction is negligible relative to typical terminal heights (24-50+ lines).

**Verdict: No layout concern for terminals >= 15 lines. Auto-hide handles smaller terminals.**

---

## Testing Strategy

### Unit Tests

1. **Registry tests** (`keybindings_test.go`):
   - Every ViewMode has at least one binding
   - Every ViewMode has `?` (help) in its bindings
   - Priority-1 bindings don't exceed 8 per view
   - No duplicate keys within a view's bindings
   - `allKeyBindingGroups()` covers all views

2. **Bar rendering tests** (`keybinding_bar_test.go`):
   - Correct output for each view mode
   - Width truncation at 40, 60, 80 columns
   - Height < 10 returns empty string
   - Disabled returns empty string
   - Compact mode at height 10-15

3. **Overlay rendering tests** (`keybinding_overlay_test.go`):
   - Full content rendering
   - Scroll offset works correctly
   - Current view section appears first
   - Footer always visible

### Golden File Tests

New golden files for:
- Bar at each major view mode (DoorsView, DetailView) at 80-col width
- Overlay at 80x24 terminal size
- Compact bar at 60-col width

### Integration Tests

- `?` key opens/closes overlay
- `h` key toggles bar
- Bar content changes when switching views
- Text input views suppress `?` and `h` handlers

---

## Dependency Analysis

### Prerequisites

None. This epic has no dependencies on unmerged work. All required infrastructure exists:
- Lipgloss for styling
- Config.yaml persistence pattern (from Epic 17 theme picker)
- MainModel View() composition pattern
- `isTextInputActive()` guard (from D-059)

### Downstream Impact

None. This is an additive feature with no API changes.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Bar height accounting breaks door rendering | Low | Medium | Comprehensive golden file tests with bar on/off |
| `?` conflicts with future key binding | Low | Low | `?` is universally understood as help; unlikely to be reassigned |
| `h` conflicts with future key binding | Medium | Low | If conflict arises, remap bar toggle; `h` for help/hide is intuitive |
| Overlay blocks important messages | Low | Medium | Overlay dismissible with both `?` and `esc`; flash messages queue until overlay dismissed |

---

## Decisions Summary

| ID | Decision | Rationale |
|----|----------|-----------|
| KB-ARCH-1 | Bar and overlay are stateless rendering functions, not Bubbletea models | No internal state to manage; pure functions are simpler and more testable |
| KB-ARCH-2 | Registry uses pure functions, not a struct | No state, no registration; all data is compile-time constant |
| KB-ARCH-3 | All new code in `internal/tui/` package, no new packages | This is exclusively a TUI concern; new package would be over-abstraction |
| KB-ARCH-4 | MainModel orchestrates bar/overlay, views are unaware | Clean separation; no per-view changes needed |
| KB-ARCH-5 | Overlay intercepts keystrokes when visible | Prevents accidental actions while reading help |
| KB-ARCH-6 | Height adjustment for bar is MainModel's responsibility | Views receive correct available height; no view-level awareness of bar |
| KB-ARCH-7 | Config persistence via async tea.Cmd | Non-blocking; follows existing theme persistence pattern |
