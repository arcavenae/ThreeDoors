# Party Mode: Space/Enter Toggle for Door Open/Close

> **Status: Planned and Done — see Story 36.4 (Done, PR #405)**
> Implemented as Story 36.4: Space/Enter Toggle — Close Door by Pressing Same Key. Decision D-137.

**Date:** 2026-03-09
**Participants:** Sally (UX Designer), Winston (Architect), Amelia (Dev), Quinn (QA)
**Topic:** Should space/enter act as a toggle to open AND close doors?

## Proposal

Make space/enter act as a toggle in the doors view:
- **Current:** space/enter opens door (→ DetailView), escape closes (→ DoorsView)
- **Proposed:** space/enter toggles (open if closed, close if opened), escape still works

## Discussion Summary

### Sally (UX Designer) — STRONG SUPPORT

- Current behavior is like a door that opens when pushed but requires a special lever to close — nobody builds doors like that
- Toggle pattern already proven in codebase: a/w/d keys toggle selection on/off (Story 36.2)
- Extending to space/enter is pure consistency
- Directly serves SOUL.md's "physical objects" philosophy: doors open AND close with the same gesture
- **Critical scoping point:** Toggle should only apply to DetailView (the "door" itself), NOT to sub-views like feedback forms, snooze date pickers, or search — those are "rooms inside the door"
- Door selection must be preserved on close — returning with no selection breaks the "peeked and closed" gesture

### Winston (Architect) — FEASIBLE, NO CONFLICTS

- In DetailView, space/enter are currently unbound in the passive/reading state
- Active sub-interactions (feedback form via `N`, snooze via `Z`) capture their own input
- The `isTextInputActive()` guard from Story 36.3 is the exact mechanism needed — already exists, battle-tested
- `selectedDoorIndex` persists across view transitions — no additional work for selection preservation
- No keybinding conflicts detected in the Epic 39 keybinding registry

### Amelia (Dev) — MINIMAL CHANGE

- Change site: `main_model.go` Update() for ViewDetail case
- Guard: `!m.isTextInputActive()`
- Action: `m.currentView = ViewDoors`
- ~3 lines of code
- Keybinding registry update: register space/enter in ViewDetail context
- Help text update in doors_view.go

### Quinn (QA) — 5 EDGE CASES

1. **Rapid toggle:** Double-tap space opens then immediately closes. Acceptable (same as double-clicking folder)
2. **Toggle during animation:** Mid-animation target reversal. Harmonica spring handles this naturally
3. **Text input guard:** Space in feedback form must NOT toggle. Covered by `isTextInputActive()`
4. **Snooze date picker:** Enter confirms date, must NOT toggle. Same guard
5. **Stale state:** Task changes while viewing. Toggle is view-level, not task-level — no issue

## Adopted Decision

**D-137: Space/Enter as toggle in DetailView (Story 36.4)**

Toggle space/enter to close door (return to DoorsView) when in DetailView's passive/reading state. Guard with `isTextInputActive()` to prevent conflicts with text input sub-views.

**Rationale:**
- Consistent with existing toggle pattern (a/w/d in Story 36.2)
- Aligns with SOUL.md "physical objects" philosophy
- Minimal code change (~3 lines)
- No keybinding conflicts
- Preserves escape as alternative back-out key

## Rejected Alternatives

**X-050: Toggle applies to all sub-views (not just DetailView)**
- Sub-views (dashboard, health, synclog) have their own enter/space semantics
- Feedback form, snooze picker use enter for confirmation
- Would create ambiguity about what "close" means in nested views

**X-051: Remove escape as back-out when toggle is added**
- Users who've learned escape shouldn't be penalized
- Multiple paths to the same action reduces friction (SOUL.md: "work with human nature")

## Consensus

Unanimous approval from all 4 agents. No dissent, no concerns unresolved.
