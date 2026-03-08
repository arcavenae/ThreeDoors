# Party Mode Artifact: Planning Docs Reconciliation (Issue #252)

**Date:** 2026-03-08
**Trigger:** GitHub Issue #252 — analyst audit found significant drift across epic-list.md, epics-and-stories.md, and ROADMAP.md
**Participants:** PM, Architect, SM, Analyst (simulated party mode)

## Problem Statement

Three planning documents have drifted out of sync:
1. `docs/prd/epic-list.md` — Stale statuses (6 epics wrong), missing 10 epics entirely, wrong story counts
2. `docs/prd/epics-and-stories.md` — Missing epics 27-29, 31-32
3. `ROADMAP.md` — Missing Epic 36

Additionally: orphan Story 0.22 has no story file, and Epics 29-31 have no story files at all.

## Root Cause Analysis

**Why do these files drift?**
1. **No automated validation** — Nothing checks that a merged PR updating story status also updates the three planning docs
2. **Multiple writers** — Workers, supervisor, PM agent, and manual edits all touch these files at different times
3. **Different update cadences** — ROADMAP.md is updated by PM sprint audits; epic-list.md was maintained manually during early phases; epics-and-stories.md grows when new epics are planned but isn't backfilled when work completes
4. **Rapid completion velocity** — Epics 19-24, 26, 34, 35 all completed in a compressed timeframe, outpacing doc updates

## Adopted Approach

### Option A: Manual One-Time Fix + Lightweight Prevention (ADOPTED)

**Rationale:** The drift is a documentation debt problem, not an architectural problem. The story files ARE the source of truth and they're correct. The fix is mechanical: update the three docs to match reality, then add a lightweight check to prevent future drift.

**What this entails:**
1. **One-time fix:** Update all three files to reflect actual story file statuses
2. **Prevention:** Add a checklist item to the story completion workflow (in CLAUDE.md or story template) reminding workers to update epic-list.md status when marking a story done
3. **Periodic audit:** PM sprint status audit (`/bmad-bmm-sprint-status`) already runs — ensure it catches planning doc drift

**Advantages:**
- Minimal overhead — no new tooling
- Fixes the immediate problem
- Leverages existing PM audit loop

**Risks:**
- Manual process can still drift, but PM audit is the safety net

### Rejected Options

#### Option B: Consolidate to Single File
**What:** Merge epic-list.md and epics-and-stories.md into one doc, keep ROADMAP.md as the scope-gated subset.
**Why rejected:** epic-list.md is a quick-reference summary; epics-and-stories.md has full story breakdowns with FRs and detailed descriptions. They serve different audiences (quick status check vs. detailed planning). Merging would make the detailed file even larger (already 50K+ tokens). The problem isn't too many files — it's that updates don't propagate.

#### Option C: Auto-Generate from Story Files
**What:** Write a script that scans `docs/stories/*.story.md` and regenerates epic-list.md and updates ROADMAP.md automatically.
**Why rejected:** Story files don't contain all the metadata in epic-list.md (FRs, deliverables, research links, phase groupings). Auto-generation would either lose this rich context or require restructuring all story files. The cost exceeds the benefit for a ~quarterly drift problem. Worth revisiting if drift becomes chronic despite PM audits.

#### Option D: CI Check for Doc Consistency
**What:** Add a CI step that validates story statuses match planning doc statuses.
**Why rejected:** Too complex for the current team size and velocity. Would need to parse YAML frontmatter from story files AND markdown tables from planning docs. Fragile, high maintenance. PM audit loop is sufficient for now.

## Implementation Recommendation

Create Story 0.30 to execute the one-time fix:
1. Update epic-list.md with correct statuses, add missing epics, fix story counts
2. Update epics-and-stories.md — add missing epics 27-29, 31-32
3. Update ROADMAP.md — add Epic 36
4. Create Story 0.22 story file for the orphan PR #242
5. Add a "planning doc update" reminder to the story completion checklist

## Scope Note

This story does NOT create missing story files for Epics 29-31. Those are implementation planning tasks that belong to their respective epic owners/PM. This story only fixes the planning doc inconsistencies.
