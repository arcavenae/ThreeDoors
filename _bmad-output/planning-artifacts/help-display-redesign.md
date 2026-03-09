# Help Display UX Redesign — Research & Design Decisions

> Party mode artifact produced 2026-03-08. UX designer analysis + architectural review.

## Problem Statement

The `:help` command in ThreeDoors TUI has two critical UX failures:

1. **Content runs off screen:** Help is a single ~300-character line with no wrapping. On an 80-column terminal, most of the content is invisible.
2. **Content disappears:** Help is rendered as a `FlashMsg` — a temporary message that auto-clears after 3 seconds via `ClearFlashCmd()`. Users cannot read it.

### Root Cause Analysis

In `internal/tui/search_view.go` (line 225-228), the `:help` command handler returns:

```go
case "help":
    return func() tea.Msg {
        return FlashMsg{Text: "Commands: :add <text>, :add-ctx, :add --why, :tag, :theme, :dispatch, :suggestions, :goals [edit], :mood [mood], :stats, :dashboard, :insights [mood|avoidance], :health, :synclog, :devqueue, :help, :quit | Keys: / search, a/w/d select, s re-roll, S suggestions, Enter open, m mood, L link, X xrefs, q quit"}
    }
```

`FlashMsg` is designed for transient confirmations ("Task added!", "Theme changed to Sci-Fi"). It is fundamentally wrong for help content because:
- Flash messages auto-clear after `flashDuration = 3 * time.Second` (defined in `styles.go`)
- Flash is rendered as a single appended line at the bottom of the current view (`main_model.go` line 1552-1553)
- No width awareness — the string is rendered raw via `flashStyle.Render(m.flash)`

Meanwhile, every other informational view in ThreeDoors uses a **dedicated ViewMode** with its own `Update()`/`View()` methods and persistent display until user dismissal. Examples: `ViewThemePicker`, `ViewSyncLog`, `ViewHealth`, `ViewInsights`, `ViewDevQueue`.

---

## UX Designer Analysis

### Reference TUI Patterns

| App | Help Pattern | Trigger | Navigation | Dismissal | Context-Sensitive? |
|-----|-------------|---------|------------|-----------|-------------------|
| nano | Full-screen help view | ^G | PgUp/PgDn | ^X (exit) | No |
| vim | Split buffer help | :help | Full buffer navigation | :q | Yes (topic-based) |
| htop | Scrollable overlay | F1 | Arrow keys | Esc/q | No |
| lazygit | Keybinding panel | ? | j/k scroll | Esc/q | Yes (per view) |
| tig | Help screen | h | j/k scroll | q | Partial |

### Key UX Principles for ThreeDoors Help

1. **Help must persist.** A user who asks for help is lost — they need time to read, scan, and find what they need. 3 seconds is hostile.

2. **Help must fit the screen.** 80-column terminal is the baseline. Content must word-wrap or use a columnar layout that adapts to `m.width`.

3. **Help should be categorized.** A flat list of 20+ commands separated by `|` is cognitively overwhelming. Group by function: navigation, task actions, commands, views.

4. **Help should follow existing patterns.** ThreeDoors already has a well-established ViewMode pattern. Help should be a dedicated view, not a special case.

5. **SOUL.md alignment:** "Every interaction should feel deliberate." A flash message is the opposite of deliberate — it's something happening *to* you, not something you control.

### Recommendation: Dedicated Help View (ViewHelp)

The lazygit pattern is the strongest match for ThreeDoors:
- **Dedicated full-screen view** via new `ViewHelp` ViewMode
- **Categorized content** grouped by function
- **Scrollable** via j/k and PgUp/PgDn (matching SyncLogView pattern)
- **Dismissed** via Esc or q (matching every other view)
- **Width-aware** via `SetWidth()` method (matching every other view)

This is the path of least surprise — it works exactly like `:synclog`, `:theme`, `:health`, and every other colon command that shows persistent content.

---

## Party Mode — Design Decisions

### Participants
- UX Designer (lead): Interaction pattern authority
- Architect: Implementation feasibility and pattern consistency
- PM: Scope and priority

### Decision 1: Display Pattern — Dedicated ViewHelp (adopted)

**Adopted:** New `ViewHelp` ViewMode with its own `HelpView` struct implementing `Update()` and `View()`.

**Rationale:**
- Consistent with 100% of existing informational views (SyncLog, Health, Insights, ThemePicker, DevQueue, Proposals)
- Follows Bubbletea MVU pattern — help is just another view
- No new infrastructure needed — ViewMode enum, `main_model.go` routing, and `SetWidth()` are all established patterns
- SyncLogView is the closest analog (scrollable text content, no interactive controls beyond navigation)

**Rejected alternatives:**
- **Flash message (current):** Fundamentally broken — auto-clears, no wrapping, no scroll. Not fixable without becoming a dedicated view anyway.
- **Overlay/modal:** Bubbletea doesn't have a native overlay concept. Would require z-index simulation, partial rendering of underlying view, and complex focus management. Over-engineered for help.
- **External pager (less/more):** Would suspend the TUI, losing Bubbletea's event loop. Re-entering the TUI after pager exit is fragile. Alien to the in-app interaction model.
- **Bottom bar hint (nano-style):** Works for a few keybindings but ThreeDoors has 20+ commands and 15+ keybindings. Cannot fit in a status bar. Could supplement but not replace a help view.
- **Split pane:** Bubbletea has no native split-pane. Would require manual layout calculation and half-height rendering of both views. Unnecessary complexity when full-screen help is simpler and sufficient.

### Decision 2: Width Handling — Adaptive Two-Column Layout

**Adopted:** Two-column key/command + description layout that adapts to terminal width. Left column is fixed-width (key/command), right column fills remaining space with word wrap.

**Rationale:**
- Key-value layout is the standard for help screens across all TUI apps
- Two-column is scannable — users look for the key on the left, description on the right
- Word wrap on the description column prevents off-screen content
- Lipgloss `lipgloss.Width()` and manual string truncation/wrapping are well-established in the codebase

**Rejected alternatives:**
- **Single-column list:** Wastes horizontal space; keys and descriptions on separate lines doubles the height
- **Three-column:** Over-designed; key + description is sufficient
- **Raw word wrap on the full line:** Breaks visual alignment; wrapping mid-key-description makes scanning impossible

### Decision 3: Navigation — j/k + PgUp/PgDn (matching SyncLogView)

**Adopted:** Line-by-line scroll via j/k (and arrow keys), page scroll via PgUp/PgDn and Space.

**Rationale:**
- Identical to SyncLogView's existing navigation model
- j/k is expected by vim-literate TUI users (ThreeDoors' core audience)
- Arrow keys cover non-vim users
- Space for page-down matches `less` convention

**Rejected alternatives:**
- **No scrolling (single page only):** Help content may exceed terminal height, especially in short terminals. Scrolling is required.
- **Mouse scroll:** Not a primary interaction mode for ThreeDoors (keyboard-first). Can be added later if Bubbletea mouse support is enabled.

### Decision 4: Content Organization — Categorized Groups

**Adopted:** Help content organized into named sections:

1. **Navigation** — a/w/d door select, Enter open, s re-roll, q quit
2. **Task Actions** — c complete, b blocked, i in-progress, e expand, f fork, p procrastinate, m mood, L link, X xrefs
3. **Commands** — :add, :add-ctx, :add --why, :tag, :theme, :dispatch, :suggestions, :goals, :mood, :stats, :dashboard, :insights, :health, :synclog, :devqueue, :help, :quit
4. **Search** — / enter search, Esc cancel

Each section has a styled header. Sections are separated by a blank line.

**Rationale:**
- Categorization reduces cognitive load — users scan headers to find the right section
- Matches how lazygit and htop organize help
- Categories align with ThreeDoors' interaction model (navigate doors, act on tasks, run commands)

**Rejected alternatives:**
- **Flat alphabetical list:** No logical grouping; user must read everything to find what they need
- **Context-sensitive (per-view):** More useful in theory but significantly more complex — each view would need to declare its keybindings. Defer to a future enhancement. Global help is the right starting point.
- **Searchable help:** Over-engineered for the current content volume. If help grows past 50 entries, revisit.

### Decision 5: Dismissal — Esc or q (standard)

**Adopted:** Esc and q both return to previous view via `ReturnToDoorsMsg`.

**Rationale:**
- Identical to every other view (SyncLog, Health, ThemePicker, etc.)
- q is the universal quit/back key in ThreeDoors (per Epic 36, Story 36.3 — universal quit)
- Esc is the universal cancel/back key

**Rejected alternatives:**
- **Any key dismisses:** Dangerous — accidental keypress loses context. User might need to re-read.
- **Only Esc:** Inconsistent with other views that accept q.

### Decision 6: Trigger — `:help` command + `?` global key

**Adopted:** Keep `:help` command. Also add `?` as a global keybinding (Shift+/) that opens help from any view.

**Rationale:**
- `:help` is discoverable through the command system (already exists)
- `?` is the de facto standard for "show help" in TUI apps (vim, less, man, lazygit, tig)
- `?` doesn't conflict with any existing keybinding
- Global binding means help is accessible even before the user knows about commands

**Rejected alternatives:**
- **F1:** Not reliably transmitted by all terminal emulators. Conflicts with tmux in some configs.
- **h key:** Already could conflict with future keybindings; `?` is more universally understood as "help"
- **`:help` only:** Requires knowing the command system exists. Chicken-and-egg problem for new users.

---

## Decisions Summary

| ID | Decision | Rationale |
|----|----------|-----------|
| D-086 | Dedicated ViewHelp view mode for `:help` display | Consistent with all existing informational views; FlashMsg is fundamentally wrong for help content |
| D-087 | `?` global keybinding opens help from any view | TUI standard (vim, less, lazygit); solves discoverability; no conflicts |

---

## Implementation Sketch (for story reference — not code)

### New Files
- `internal/tui/help_view.go` — HelpView struct with Update(), View(), SetWidth()
- `internal/tui/help_view_test.go` — table-driven tests, golden file tests

### Modified Files
- `internal/tui/main_model.go` — Add `ViewHelp` to ViewMode enum, add `helpView *HelpView` field, add routing in Update()/View(), add `?` key handler
- `internal/tui/search_view.go` — Change `:help` case to return `ShowHelpMsg{}` instead of `FlashMsg`
- `internal/tui/messages.go` — Add `ShowHelpMsg` type

### Pattern to Follow
`SyncLogView` is the closest existing analog:
- Scrollable content with offset-based pagination
- j/k/PgUp/PgDn navigation
- q/Esc dismissal via `ReturnToDoorsMsg`
- `SetWidth()` for terminal width awareness
- Help bar at bottom showing available keys

### Content Source
Help content should be defined as structured data (slices of section/entry structs) in the help view file, not as a raw string. This makes it maintainable and testable.

### Width Calculation
```
Key column:  max(len(key) for all entries) + 2 padding
Desc column: terminal_width - key_column_width - left_margin - right_margin
```
Description text wraps at the desc column boundary.
