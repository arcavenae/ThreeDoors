# Course Correction: Unified Door Key Indicators

> **Status: Planned and Done — see Story 39.13 (Done, PR #407)**
> Implemented as Story 39.13: Unified Door Key Indicator Toggle. Referenced in Epic 39 (epic-list.md). Decision D-137, D-138.

**Date:** 2026-03-09
**Triggered by:** User feedback on keybinding/tooltip visual approach
**Scope:** Epic 39 — Keybinding Display System
**Affects:** Stories 39.4, 39.9, 39.10, 39.12 (all Done), new Story 39.13

## Problem Statement

The current implementation has **two separate toggle mechanisms** for keybinding discoverability:

1. **Keybinding bar** (`h` key, `show_keybinding_bar` config) — a bottom bar showing context-sensitive key actions per view, rendered by MainModel in the footer slot (D-088, D-091)
2. **Inline hints** (`:hints` command, `show_inline_hints` config) — `[a]` `[w]` `[d]` rendered on door frames as "doorknob" decorations, with auto-fade after N sessions (D-125, D-126)

**The separation creates UX confusion:**
- Two different toggles for conceptually the same feature (key discoverability)
- When both are ON, door selection keys appear in both places (redundant)
- The bar duplicates information already visible on the doors themselves
- Users must learn two commands (`h` and `:hints`) for one concept

## Proposed Change

**Unify under the `h` toggle with hints rendered ON the doors:**

1. `h` key toggles `show_keybinding_bar` — when ON, door key indicators `[a]` `[w]` `[d]` appear at the base center of each door
2. When OFF, **no** key indicators appear anywhere — not on doors, not in a bottom bar
3. Remove the separate `show_inline_hints` / `:hints` mechanism (or alias it to `h` behavior)
4. Keep `?` overlay untouched (comprehensive reference is a different concept)

## What Changes vs. What Stays

### Stays the Same
- `h` key toggle keystroke
- `show_keybinding_bar` config persistence in config.yaml
- `?` overlay (full keybinding reference) — completely unaffected
- Inline hint rendering infrastructure (renderInlineHint, renderDoorHint, theme Render() hint parameter)
- Door hint visual styling (selection awareness, dim/bright states)

### Changes
- **Toggle target:** `h` now controls inline door hints (not a bottom bar)
- **Config unification:** Single config field controls the feature (`show_keybinding_bar` repurposed or renamed)
- **Bottom bar removal for doors view:** The keybinding bar in doors view is removed or significantly reduced (door keys no longer duplicated there)
- **`:hints` command:** Either removed or made an alias for the `h` toggle
- **Auto-fade:** The session-based auto-fade (D-126) may be retained or simplified since the feature is now manual-toggle only
- **Non-door views:** The keybinding bar remains useful for non-door views (detail, search, mood, etc.) where inline hints wouldn't make sense — this needs a decision

## Decisions Required

1. **D-137: Unify bar and inline hints under `h` toggle** — Merge the two separate mechanisms
2. **What happens to the bar in non-door views?** Option A: Keep bar for non-door views, show door hints in doors view. Option B: Remove bar entirely, hints only appear on doors.
3. **Config field naming** — Reuse `show_keybinding_bar` (backward compatible) or rename to `show_key_hints`?
4. **Auto-fade retention** — Keep session-based auto-fade (D-126) or make it purely manual toggle?

## Decisions Summary

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| D-137 | Unify bar toggle and inline hints under `h` key | Keep separate mechanisms (D-125) | Reduces cognitive load; one concept = one toggle; eliminates redundancy |
| D-138 | Keep bar for non-door views; door hints replace bar in doors view | Remove bar entirely across all views | Non-door views lack spatial anchors for inline hints; bar remains useful there |
| D-139 | Rename config to `show_key_hints`; migration reads old field | Keep `show_keybinding_bar` name | New name better describes unified behavior; auto-migration is trivial |
| D-140 | Remove auto-fade; `h` is purely manual toggle | Keep session-based auto-fade (D-126) | Auto-fade was designed for onboarding scaffolding; `h` toggle is a power-user feature; users control their own experience |
| X-079 | — | Remove `:hints` command entirely | Keep `:hints` as alias for `h` toggle for command-mode accessibility; some users prefer typing to key press |

## Impact Assessment

- **Overrides:** D-125 (separate mechanisms), D-126 (auto-fade), partially D-091 (h = bar)
- **Preserves:** D-088 (MainModel renders), D-089 (overlay separate), D-090 (registry), D-092 (defaults ON), D-093 (dim styling), D-094 (auto-hide small terminals)
- **New story:** 39.13 — Unified Door Key Indicator Toggle
- **Story 39.12 impact:** Auto-Fade story becomes N/A if auto-fade is removed; should be updated to Cancelled

## Risk Assessment

- **Low risk:** The inline hint infrastructure already exists and works (39.9, 39.10)
- **Low risk:** The `h` toggle mechanism already exists and works (39.4)
- **Medium risk:** Removing the bottom bar in doors view changes the visual layout — needs testing across terminal sizes
- **Low risk:** Config migration is trivial (rename field, read old field as fallback)
