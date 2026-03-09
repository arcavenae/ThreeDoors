# UX Designer Review: Keybinding Display System

**Date:** 2026-03-08
**Reviewer:** Sally (UX Designer)
**Input:** Party mode consensus from keybinding-display-party-mode.md

---

## Executive Summary

The keybinding display feature addresses a real discoverability gap in ThreeDoors. With 18 view modes and 40+ keybindings, users currently rely on memory or onboarding recall. The proposed bar + overlay pattern is well-established (nano, vim, htop) and aligns with SOUL.md's principle of reducing friction.

**Recommendation: APPROVED with refinements below.**

---

## Visual Hierarchy Analysis

### Current State

The TUI has a clear visual hierarchy:
1. **Primary:** Three door panels (the focal point)
2. **Secondary:** Greeting text and footer message
3. **Tertiary:** Status indicators (sync, conflicts, proposals)

### Proposed State with Bar

The bar must sit at the **quaternary** level — below everything else in visual importance:

1. **Primary:** Three door panels (unchanged)
2. **Secondary:** Greeting + footer (unchanged)
3. **Tertiary:** Status indicators (unchanged)
4. **Quaternary:** Keybinding bar (new — subtle, available, not attention-grabbing)

### Styling Recommendations

The bar MUST be visually recessive:
- **Color:** Use `Faint(true)` or a dim gray foreground. Never the theme's accent color.
- **Weight:** No bold text. The bar should feel like a watermark, not a toolbar.
- **Separator:** A thin horizontal rule (single `─` character repeated) above the bar separates it from content. Styled with the same dim treatment.
- **Background:** No background color. Transparent to the terminal default.

Why: The doors are the product. The bar is infrastructure. Infrastructure should be invisible until you need it.

---

## Information Density Analysis

### Bar Content Strategy: "The Five Essential Actions"

For each view, show at most 5-6 key groups. The rule: **if a user can only learn 5 things about this screen, what would they be?**

#### DoorsView (no selection)

```
 a/w/d Select   s Re-roll   n Add   : Cmd   h Bar   ? Help
```

Why these 6:
- `a/w/d` — The primary action. You must select a door.
- `s` — Re-roll is the escape valve. Users need to know they can get new doors.
- `n` — Adding tasks is a power feature people forget about.
- `:` — Command mode unlocks all hidden commands.
- `h` — Self-referential: tells users how to hide the bar.
- `?` — The escape hatch to full reference.

Dropped from bar (overlay-only): `/` (search, redundant with `:search`), `m` (mood, secondary), `S` (sync status, rare), `q` (quit, discoverable).

#### DoorsView (door selected)

```
 enter Confirm   a/w/d Change   s New doors   esc Deselect   ? Help
```

Why the change: Once a door is selected, the primary action changes to confirm. The bar should reflect the user's new decision context.

#### DetailView

```
 esc Back   c Done   b Blocked   e Expand   f Fork   ? Help
```

Why: These are the core task management actions. `i` (in-progress), `p` (procrastinate), `r` (reconsider), `x` (dispatch), `g` (generate), `d` (delete), `u` (undo) are secondary — they go in the overlay.

#### Text Input Views (AddTask, Search)

```
 enter Submit   esc Cancel   ? Help
```

Minimal. The user is typing — don't distract with navigation keys that aren't active.

#### Modal/Selection Views (Mood, ThemePicker, Onboarding)

```
 ↑/↓ Navigate   enter Select   esc Back   ? Help
```

Generic navigation pattern. Keep it simple.

---

## Discoverability vs Clutter Tradeoff

### The Tension

SOUL.md says "show less." But showing zero help means users don't discover features. The resolution:

**Default: bar ON for new users, OFF for power users.**

Implementation:
- First-run (no config.yaml exists): `show_keybinding_bar: true`
- After user presses `h` to hide: `show_keybinding_bar: false` (persisted)
- User can always press `h` to toggle or `?` for full help

This matches the nano model: nano shows `^G Help ^O Write Out` by default. Power users know to look at the menu. New users see the bar and learn.

### Progressive Disclosure Ladder

1. **Bar** (persistent, 5-6 keys) — "What can I do right now?"
2. **Overlay** (on-demand, all keys) — "What are ALL my options?"
3. **Onboarding** (first-run, tutorial) — "Let me teach you the basics"

The bar is the middle rung. It bridges the gap between onboarding (which users forget) and the overlay (which users don't know exists until they see `?` in the bar).

---

## Interaction Patterns

### Bar Toggle (`h`)

- Press `h`: bar disappears immediately. No animation. (SOUL.md: "every interaction should feel deliberate" — instant response)
- Press `h` again: bar reappears immediately.
- State persisted to config.yaml on change.

### Overlay (`?`)

- Press `?`: overlay appears instantly, covering the full screen with a dimmed background.
- The overlay shows ALL keybindings for ALL views, organized by category.
- Press `?` or `esc` to dismiss.
- The overlay is context-highlighted: the current view's section is at the top or visually emphasized.

Why not view-specific overlay content? Because the overlay is a reference document. Users open it when they're confused and might not know which view they're in. Showing everything with context highlighting serves both "where am I?" and "what can I do?" questions.

### Key Conflict Check

- `?` is currently unused across all views. Safe.
- `h` is currently unused across all views. Safe.

---

## Accessibility and Readability

### Contrast Requirements

- Bar text must be readable against common terminal backgrounds (dark and light themes).
- Use Lipgloss `Foreground()` with adaptive color (ANSI 240-245 range for dark themes, 100-105 for light themes).
- Do NOT use `Faint(true)` alone — some terminals render faint text as invisible. Use explicit dim foreground color instead.

### Terminal Size Graceful Degradation

| Terminal Height | Behavior |
|----------------|----------|
| < 10 lines | Bar hidden (doors need all available space) |
| 10-15 lines | Compact bar: keys only, no descriptions (`a w d s n : ? h`) |
| > 15 lines | Full bar with descriptions |

| Terminal Width | Behavior |
|---------------|----------|
| < 40 cols | Bar shows only `? Help` |
| 40-60 cols | Bar shows 3 most important keys + `?` |
| 60-80 cols | Bar shows 5 keys + `?` |
| > 80 cols | Full bar |

### Overlay Size Handling

The overlay should scroll if the terminal is too short to show all keybindings. Use `j/k` or `↑/↓` to scroll the overlay content. Always show "Press ? or esc to close" at the bottom (fixed footer within the overlay).

---

## Consistency with Existing Patterns

### Theme Integration

The bar should NOT be themed (unlike doors). Rationale:
- The bar is chrome/infrastructure, not content
- Themed bars would require updating every theme for a non-core feature
- A neutral dim style works with all themes
- This matches how nano's help bar isn't themed with the file's syntax highlighting

The overlay border can use the current theme's accent color for the frame, providing a subtle connection to the visual identity without requiring theme-specific overlay implementations.

### Separator Line

Above the bar, render a thin separator:
```
 ────────────────────────────────────────────────────
 a/w/d Select   s Re-roll   n Add   : Cmd   h Bar   ? Help
```

The separator uses `lipgloss.NewStyle().Foreground(dimColor)` and is a string of `─` repeated to terminal width. This visually separates content from chrome.

---

## Recommendations Summary

1. **Bar defaults ON for new users, OFF once toggled** — progressive disclosure
2. **Max 5-6 key groups in bar** — resist the urge to show everything
3. **Bar is visually recessive** — dim foreground, no background, no bold
4. **`h` toggles bar, `?` opens overlay** — separate "persistent reference" from "help me now"
5. **Overlay shows ALL keys for ALL views** — reference document, not context-sensitive filter
6. **Current view highlighted in overlay** — helps "where am I?" orientation
7. **Terminal size adaptation is mandatory** — doors always get priority for space
8. **No theme integration for bar** — neutral dim style works universally
9. **Separator line above bar** — clear visual boundary between content and chrome
10. **Overlay scrollable** — supports terminals that can't show all bindings at once
