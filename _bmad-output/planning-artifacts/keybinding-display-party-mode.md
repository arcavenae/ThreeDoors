# Party Mode: Keybinding Display System for ThreeDoors TUI

**Date:** 2026-03-08
**Participants:** John (PM), Sally (UX Designer), Winston (Architect), Quinn (QA), Amelia (Dev)
**Topic:** Toggleable keybinding display bar and full keybinding overlay

---

## Round 1: Initial Feature Analysis

### John (PM): SOUL.md Alignment Check

The first question here is: does this feature pass the SOUL.md smell test? Let me run through the principles:

1. **"Three Doors, Not Three Hundred"** — This is the tension point. A keybinding bar adds visual information. However, it's meta-information (how to use the app) not task information (more things to do). It doesn't add more doors or more choices. It helps the user interact with the three doors they already have. That's discoverability, not feature creep.

2. **"Every Interaction Should Feel Deliberate"** — The bar actively supports this. If users don't know which keys are available, they can't interact deliberately. They fumble. Fumbling is the opposite of deliberate. The bar transforms uncertainty into confidence.

3. **"Progress Over Perfection"** — Users who don't know about `c` for complete or `e` for expand are stuck. The bar directly reduces friction to starting.

4. **"Meet Users Where They Are"** — Power users want a clean screen. New users need guidance. Toggleable solves both. This is exactly the pattern nano, vim, and htop use.

**Verdict: ALIGNED.** This feature reduces friction, improves discoverability, and the toggle mechanism prevents it from becoming visual clutter. It's infrastructure that makes the existing experience better, not a new thing to manage.

### Sally (UX Designer): User Journey Analysis

Looking at this from the user's perspective, there's a clear discoverability problem today. ThreeDoors has 18 view modes and dozens of key bindings. The onboarding flow (Epic 10) teaches some keys, but users forget. The current experience:

1. User opens ThreeDoors → sees three doors
2. User thinks "how do I select one?" → must remember or guess
3. User remembers `a/w/d` → great, but what about `:` commands?
4. User never discovers `e` (expand), `f` (fork), `x` (dispatch) → features are invisible

**The bar solves the "what can I do right now?" question.** This is a fundamental UX primitive — progressive disclosure of available actions.

Key UX insight: The bar should show only the 4-6 most important keys for the current view. Not all keys. The overlay (`?`) shows everything. The bar is a teaser, the overlay is the reference.

### Winston (Architect): Technical Feasibility

From an architecture perspective, this maps cleanly to the existing Bubbletea MVU pattern:

1. **Bar component**: A new component that receives the current ViewMode and returns a rendered string. The MainModel's `View()` appends it to the bottom.
2. **Overlay**: A new ViewMode (`ViewKeybindingOverlay`) or a modal layer that overlays the current view.
3. **Registry**: A data structure mapping ViewMode → []KeyBinding, where each KeyBinding has a key, description, and priority (for bar vs overlay-only).
4. **Toggle state**: A bool on MainModel (`showKeybindingBar`), persisted in config.yaml.

No new interfaces needed. No new packages needed. This lives entirely in `internal/tui/`.

---

## Round 2: Design Deep Dive

### Sally (UX Designer): Bar Layout Proposal

Drawing from nano, vim, and htop patterns:

**Nano-style approach (ADOPTED):**
```
 a←  w↑  d→ Select   s↓ Re-roll   enter Confirm   ? Help   : Cmd
```

Key principles:
- Keys are left-aligned, descriptions follow
- Separator character between groups (double space)
- Muted/dim styling so it doesn't compete with doors
- One line maximum — if terminal is too narrow, truncate from right
- The `?` key is always shown (it's the escape hatch to full help)

**Bar key selection per view (top 6-8 keys only):**

| View | Keys Shown |
|------|-----------|
| DoorsView (no selection) | `a/w/d` Select, `s` Re-roll, `n` Add, `/` Search, `:` Cmd, `?` Help |
| DoorsView (selected) | `enter` Confirm, `a/w/d` Change, `s` Re-roll, `esc` Deselect, `?` Help |
| DetailView | `esc` Back, `c` Complete, `b` Blocked, `e` Expand, `f` Fork, `?` Help |
| SearchView | `esc` Close, `enter` Select, `↑/↓` Navigate, `?` Help |
| AddTaskView | `enter` Save, `esc` Cancel, `?` Help |
| MoodView | `1-5` Select, `esc` Cancel, `?` Help |
| Other views | `esc` Back, `?` Help (minimal) |

### Winston (Architect): Keybinding Registry Design

The registry pattern is the right approach. Each view registers its bindings, and the bar/overlay components query it:

```go
type KeyBinding struct {
    Key         string   // Display string: "a/w/d", "enter", "ctrl+c"
    Description string   // "Select door", "Confirm"
    Priority    int      // 1=always show in bar, 2=bar if space, 3=overlay only
}

type KeyBindingSet struct {
    ViewMode ViewMode
    Bindings []KeyBinding
}
```

The registry is a `map[ViewMode][]KeyBinding` initialized at startup. No runtime registration needed — all bindings are known at compile time.

For the overlay, it should be a modal that renders ON TOP of the current view (not a ViewMode transition). This preserves the user's context — they can see what they're working with while reading the help.

### Quinn (QA): Testing Strategy

Testing plan:
1. **Registry correctness**: Every ViewMode has at least `?` (help) and `esc`/`q` (exit) registered
2. **Bar rendering**: Golden file tests for bar at various widths (80, 120, 40 cols)
3. **Overlay rendering**: Golden file tests for full overlay
4. **Toggle persistence**: Config read/write for `show_keybinding_bar` setting
5. **Auto-hide**: Bar disappears when terminal height < threshold
6. **Context sensitivity**: Bar content changes when ViewMode changes

---

## Round 3: Refinement and Edge Cases

### Sally (UX Designer): Terminal Size Handling

**Small terminal strategy:**
- Terminal height < 10 lines: Bar auto-hides regardless of toggle state (doors need the space)
- Terminal height 10-15 lines: Bar shows in compact mode (keys only, no descriptions)
- Terminal height > 15 lines: Full bar with descriptions

**Small terminal width strategy:**
- Width < 40: Show only `?` Help
- Width 40-80: Show top 4 keys + `?`
- Width > 80: Show full bar

### Winston (Architect): Overlay Design

The overlay should NOT be a new ViewMode. Reasons:
1. Adding ViewMode means MainModel Update() needs a new case
2. ViewMode transition clears the previous view (bad for context)
3. The overlay is ephemeral — press `?` to show, `?` or `esc` to dismiss

Instead, use a `showOverlay bool` on MainModel. In View():
```
if m.showOverlay {
    return renderOverlay(m.viewMode, m.width, m.height)
}
return normalView()
```

The overlay renders a semi-transparent (dimmed background) full-screen panel listing ALL keybindings organized by category.

### John (PM): Scope Decision

We need to resist scope creep here. What's NOT in this epic:
- Custom keybinding remapping (different epic, if ever)
- Animated transitions when bar appears/disappears
- Keybinding tutorials or interactive walkthroughs
- Per-key tooltips on hover (this is a TUI, no hover)

What IS in this epic:
- Static keybinding registry (data model)
- Concise bar component (bottom bar)
- Full overlay component (? key)
- Toggle + persistence (config.yaml)
- Integration with all existing views
- Terminal size adaptation

### Amelia (Dev): Implementation Concerns

Looking at the codebase, the MainModel.View() currently returns the view's output directly. To add the bar, we need to:

1. Calculate available height for the view (total height - bar height if shown)
2. Pass reduced height to view components
3. Append bar to the bottom of every view's output

This means the bar is rendered by MainModel, not by individual views. Views don't know about the bar. This is clean separation.

For the overlay, the `Update()` handler for `?` needs to be in MainModel (global level), similar to how `q` is handled for universal quit (D-059). The overlay captures all keystrokes except `?` (toggle off) and `esc` (dismiss).

---

## Round 4: Final Consensus

### Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| KB-1 | Bar rendering location | MainModel.View() appends bar | Per-view bar rendering | Clean separation; views don't know about bar |
| KB-2 | Overlay implementation | Boolean flag on MainModel, not new ViewMode | New ViewMode for overlay | Preserves user context; overlay is ephemeral |
| KB-3 | Keybinding data source | Compile-time registry (map[ViewMode][]KeyBinding) | Config file / runtime registration | All bindings are known at compile time; YAGNI |
| KB-4 | Bar key selection | Priority-based (1=always, 2=space permitting, 3=overlay only) | Show all keys | SOUL.md: show less, not more |
| KB-5 | Toggle persistence | config.yaml `show_keybinding_bar: true/false` | In-memory only | User preference should survive restarts |
| KB-6 | Toggle key | `?` toggles bar visibility; also opens overlay on first press, second press toggles bar | Separate keys for bar toggle and overlay | `?` is universally understood as "help" |
| KB-7 | Small terminal handling | Auto-hide bar below 10 lines height; compact mode 10-15 lines | Always show bar regardless of size | Doors must have priority for screen space |
| KB-8 | Bar styling | Dim/muted text, single line, Lipgloss styled | Bright/prominent bar | Bar must not compete with doors for visual attention |
| KB-9 | Overlay styling | Full-screen dimmed background with bright keybinding list | Floating popup / partial overlay | Full screen provides best readability; dimmed bg preserves context sense |
| KB-10 | `?` key behavior | Single press: show overlay. `h` key: toggle bar. | `?` for both | Separating concerns: `?` = "help me now", `h` = "I want persistent help bar" |

### Revised Decision KB-6/KB-10

After further discussion, the team revised the `?` behavior:

- **`?`** opens/closes the full keybinding overlay (modal help screen)
- **`h`** toggles the persistent bottom bar on/off
- This separates "I need help right now" (`?`) from "I want ongoing reference" (`h`)
- Both are documented in the bar itself and in the overlay

---

## Story Breakdown (PM Recommendation)

1. **39.1: Keybinding Registry Model** — Data types and per-view keybinding definitions
2. **39.2: Concise Keybinding Bar Component** — Bottom bar rendering with Lipgloss styling
3. **39.3: Full Keybinding Overlay** — Modal overlay showing all bindings
4. **39.4: Toggle Behavior and Config Persistence** — `h` toggle, config.yaml persistence, terminal size adaptation
5. **39.5: View Integration and Polish** — Wire bar/overlay into all 18 view modes, ensure context-sensitivity

---

## Appendix: ASCII Mockups

### Doors View with Bar

```
  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
  │              │   │              │   │              │
  │   Door 1     │   │   Door 2     │   │   Door 3     │
  │   Fix bug    │   │   Write docs │   │   Review PR  │
  │              │   │              │   │              │
  │              │   │              │   │              │
  └──────────────┘   └──────────────┘   └──────────────┘

  Pick a door. Any door. Let's go.

  a←  w↑  d→ Select   s↓ Re-roll   n Add   / Search   : Cmd   ? Help
```

### Detail View with Bar

```
  ── Fix authentication bug ──────────────────────────

  Status: active
  Source: tasks.txt
  Created: 2026-03-08

  This needs to be fixed before the release...

  esc Back   c Complete   b Blocked   e Expand   f Fork   ? Help
```

### Full Overlay (? key)

```
  ╔══════════════════════════════════════════════════╗
  ║              KEYBINDING REFERENCE                ║
  ╠══════════════════════════════════════════════════╣
  ║                                                  ║
  ║  NAVIGATION                                      ║
  ║  a / ←     Select Door 1                         ║
  ║  w / ↑     Select Door 2                         ║
  ║  d / →     Select Door 3                         ║
  ║  s / ↓     Re-roll doors                         ║
  ║  enter     Confirm selection                     ║
  ║  esc       Back / Deselect                       ║
  ║                                                  ║
  ║  ACTIONS                                         ║
  ║  c          Complete task                        ║
  ║  b          Mark blocked                         ║
  ║  i          Mark in-progress                     ║
  ║  e          Expand (add subtasks)                ║
  ║  f          Fork (create variant)                ║
  ║  n          Add new task                         ║
  ║  u          Undo completion                      ║
  ║                                                  ║
  ║  VIEWS                                           ║
  ║  /          Search                               ║
  ║  :          Command mode                         ║
  ║  m          Mood check                           ║
  ║  S          Sync status                          ║
  ║                                                  ║
  ║  DISPLAY                                         ║
  ║  h          Toggle keybinding bar                ║
  ║  ?          This help overlay                    ║
  ║  q          Quit                                 ║
  ║                                                  ║
  ║           Press ? or esc to close                ║
  ╚══════════════════════════════════════════════════╝
```
