# Party Mode Triage: Issue #330 — Dashboard 'q' Key Exits App Instead of Going Back

**Date:** 2026-03-09
**Issue:** #330
**Participants:** Sally (UX Designer), Winston (Architect), Amelia (Developer)
**Rounds:** 2

## Problem Statement

Story 36.3 (PR #276) added a universal quit handler that intercepts 'q' from all non-text-input views. This causes sub-views (insights/dashboard, health, synclog, next steps, avoidance prompt) to exit the app when the user presses 'q', instead of going back to the previous view.

## Options Evaluated

### Option A: Add dashboard to an exemption list in the universal handler

Add sub-views to an `isQuitAllowed()` or similar check so `q` only triggers quit from non-exempted views.

**Pros:** Targeted, easy to implement initially.
**Cons:** Creates a growing maintenance list. Every new view must remember to add itself. The default behavior (quit) is the dangerous one, violating fail-safe defaults. Fragile.

**Verdict:** REJECTED. Maintenance burden grows with every new view. Wrong default.

### Option B: Have dashboard view consume the 'q' key before it reaches the universal handler

Modify sub-views to handle 'q' in their own Update() methods before the universal handler fires.

**Pros:** Per-view control.
**Cons:** Does not work with Bubbletea's message flow. The universal handler at `main_model.go:910-913` fires BEFORE view delegation at line 921. Sub-views never see the 'q' key. Would require restructuring the entire Update pipeline.

**Verdict:** REJECTED. Architecturally impossible without major refactor.

### Option C: Change universal quit to only work from the doors view (ADOPTED)

Remove the universal `q` quit handler (lines 910-913) entirely. The doors view already handles `q` as quit at line 977. In each sub-view's Update method, add `case "q":` that triggers a go-back message (equivalent to Esc).

**Pros:**
- Matches TUI conventions (vim, lazygit, htop, ranger): `q` = "close current thing"
- At root (doors), closing = quit. In sub-view, closing = go back.
- No exemption list. No new abstractions. Clean separation.
- Redundant handler removed (doors already has `q` at line 977).
- Aligns with SOUL.md "work with human nature" principle.

**Cons:**
- Overrides Story 36.3 AC ("q quits from any view").
- Sub-views that forget to add `q` handler will have `q` as no-op (safe default, Esc still works).

**Verdict:** ADOPTED. Best UX, cleanest architecture, safest defaults.

### Option D: Use a different key for universal quit

Change universal quit to a different key (e.g., `Q`, `ctrl+q`).

**Pros:** Avoids the conflict entirely.
**Cons:** `q` is THE standard TUI quit key. Changing it violates user expectations and muscle memory. The problem is not the key, it is the scope.

**Verdict:** REJECTED. Solves wrong problem. Users expect `q` to work.

## Decision

**Option C adopted.** Three-tier `q` key behavior:
1. **Doors view (root):** `q` = quit app
2. **Sub-views (non-input):** `q` = go back to previous view (same as Esc)
3. **Text input views:** `q` = type the character 'q'

This creates a consistent mental model: `q` always means "close what I'm looking at."

## Affected Files (for implementation)

- `internal/tui/main_model.go` — Remove lines 910-913 (universal quit handler)
- `internal/tui/insights_view.go` — Add `q` as go-back in Update()
- `internal/tui/health_view.go` — Add `q` as go-back in Update()
- `internal/tui/synclog_view.go` — Add `q` as go-back in Update()
- `internal/tui/next_steps_view.go` — Add `q` as go-back in Update()
- `internal/tui/avoidance_prompt_view.go` — Add `q` as go-back in Update()
- `internal/tui/main_model_test.go` — Update `TestUniversalQuit_InsightsView_QKeyQuits` to expect go-back; add table-driven test for q-as-back across all sub-views
- `internal/tui/keybindings.go` — Add `q: back` to sub-view bindings

## Story Reference

Story 0.33 created for implementation (Decision D-125). This overrides Story 36.3's "q quits from any view" AC.
