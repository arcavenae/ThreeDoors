# Epic 39 Governance Sync Investigation

**Date:** 2026-03-09
**Investigator:** fancy-owl (worker agent)
**Trigger:** Supervisor observed repeated governance sync PRs overwriting Epic 39 data

---

## Root Cause Analysis

### What Happened

1. **Two workers independently claimed Epic 39** on 2026-03-08:
   - `work/keybinding-display-planning` created Epic 39 as **Keybinding Display System** (PR #292, merged)
   - `work/beautiful-stats-epic-planning` created Epic 39 as **Beautiful Stats Display** (PR #299, merged after #292)

2. **PR #299 merged with stale Epic 39 numbering.** The Beautiful Stats worker was given Epic 39 before the Keybinding Display worker claimed it. Both were spawned in parallel with no coordination. The collision was detected post-merge.

3. **Manual fix (PR #313)** renumbered Beautiful Stats to Epic 40 and recorded decision D-104 in BOARD.md. The collision residuals were cleaned up.

4. **Project-watchdog then created governance sync PRs referencing the now-stale Epic 39 = Beautiful Stats mapping:**
   - PR #311: "governance sync — Story 39.1 Done (PR #305)" — This was correct (39.1 *is* a Keybinding Display story), but the PR title appeared alongside the collision history, creating confusion
   - PR #311 was closed as superseded by PR #313

5. **Subsequent governance sync PRs (#314, #316, #325) operated correctly** — they updated Epic 39 = Keybinding Display System as intended. The apparent "overwriting" was actually correct updates arriving after the collision was resolved.

### Root Cause: No Epic Number Reservation System

The core problem is that **epic numbers are allocated informally**. The standing order says "PM allocates epic numbers" but:
- No persistent PM agent was running at the time (Epic 37 not yet complete)
- Workers received epic numbers from the supervisor/BMAD pipeline without cross-checking what was already claimed
- Two workers ran in parallel and both used Epic 39

### Why It Looked Like Repeated Overwriting

The confusion arose because:
1. Governance sync PRs touch the same files (ROADMAP.md, epic-list.md, epics-and-stories.md) repeatedly
2. Each sync PR targets a slightly different state of these files
3. When merged out of order or after collision fixes, the diff history looks like "Epic 39 keeps changing"
4. In reality, after PR #313 fixed the collision, all subsequent governance syncs correctly treated Epic 39 = Keybinding Display

### Current State (Verified)

| Doc | Epic 39 Identity | Status |
|-----|-----------------|--------|
| ROADMAP.md | Keybinding Display System | 3/6 (correct) |
| epic-list.md | Keybinding Display System | In Progress (2/6) — **stale, should be 3/6** |
| epics-and-stories.md | Keybinding Display System | In Progress (2/6) — **stale, should be 3/6** |
| Story 39.2 in ROADMAP.md | "Not Started" | **stale — PR #318 merged** |

PR #325 (open) would fix the 39.2 status but hasn't merged yet.

## Contributing Factors

1. **No epic number registry** — Epic numbers exist only in ROADMAP.md, which is a race condition when multiple workers write to it simultaneously
2. **Project-watchdog creates many small PRs** — Each merged story gets its own governance sync PR, creating PR fatigue and merge ordering issues
3. **No locking mechanism** — Nothing prevents two agents from editing planning docs concurrently
4. **Workers self-update planning docs** — Some workers update ROADMAP.md in their implementation PRs, while project-watchdog also updates ROADMAP.md in separate PRs, creating conflicting edits

## Recommendations

### R1: Epic Number Reservation in BOARD.md (Implement Now)

Add a "Reserved Epic Numbers" section to BOARD.md that serves as the single source of truth for what epic number maps to what feature. This is lightweight and doesn't require a PM agent.

### R2: Batch Governance Syncs (Process Change)

Project-watchdog should batch multiple story status updates into a single PR rather than creating one PR per story. This reduces PR fatigue and merge ordering issues.

### R3: Workers Should NOT Update Planning Docs (Process Change)

Story implementation workers should only update their story file. Planning doc updates (ROADMAP.md, epic-list.md, epics-and-stories.md) should be project-watchdog's exclusive responsibility. This eliminates concurrent edit conflicts.

### R4: Housekeeping Epic Not Needed

A dedicated "housekeeping" epic for governance work is unnecessary because:
- Governance syncs are doc-only (no stories/ACs needed)
- Project-watchdog PRs already use a clear "governance sync" prefix
- Adding an epic would create overhead for routine maintenance

Instead, governance sync PRs should reference the story they're updating (e.g., "Story 39.2") rather than claiming their own story.

### R5: Project-Watchdog Definition Update (DONE)

Added to project-watchdog.md:
- Verify epic identity step before updating any epic-level data
- Batching instruction for governance syncs when multiple PRs merge in quick succession

### R6: Resolve CLAUDE.md / Worker.md Contradiction (Needs PM Decision — Q-004)

**The contradiction:**
- `CLAUDE.md` line 46: "Every PR that changes story status MUST also update planning docs in the same PR"
- `agents/worker.md` line 101: "CANNOT: Modify ROADMAP.md, SOUL.md, or CLAUDE.md"

Workers that follow CLAUDE.md update ROADMAP.md, creating concurrent edits with project-watchdog. Workers that follow worker.md skip ROADMAP.md, leaving project-watchdog as the sole updater. This inconsistency is a contributing factor to the governance sync churn.

**Recommendation:** PM should decide one of:
- **Option A:** Workers update their own story file + all three planning docs (CLAUDE.md wins, remove ROADMAP.md from worker.md CANNOT list, retire project-watchdog planning updates)
- **Option B:** Workers update only their story file, project-watchdog handles all planning docs (worker.md wins, update CLAUDE.md to match)
- **Option C:** Keep current mixed approach but add locking/coordination (high complexity, not recommended)

## Changes Made in This PR

1. **BOARD.md:** Added D-112 (epic number registry), D-113 (no housekeeping epic), Q-003 (batch governance syncs), Q-004 (planning doc ownership)
2. **BOARD.md:** Added "Epic Number Registry" section with reservation table and rules
3. **project-watchdog.md:** Added epic identity verification step and batching instruction
4. **Investigation artifact:** This file
