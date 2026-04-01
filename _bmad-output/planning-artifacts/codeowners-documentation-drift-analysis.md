# CODEOWNERS Documentation Drift Analysis

**Date:** 2026-03-31
**Analyst:** worker/lively-deer
**Provenance:** L2 (human-directed planning)

## Problem Statement

CLAUDE.md, merge-queue.md, and pr-shepherd.md all list four files as CODEOWNERS-protected that are NOT actually protected in `.github/CODEOWNERS`. This causes merge-queue to incorrectly apply `status.needs-human` labels to governance sync PRs that touch these files.

## Actual State of `.github/CODEOWNERS`

**Protected (uncommented):**
- `SOUL.md` → @arcaven
- `CLAUDE.md` → @arcaven
- `/.claude/` → @arcaven
- `/.env` → @arcaven
- `/.gitignore` → @arcaven
- `/.github/` → @arcaven
- `/agents/` → @arcaven
- `/_bmad/` → @arcaven

**NOT Protected (commented out):**
- `ROADMAP.md` — commented out (line 17)
- `/docs/prd/epic-list.md` — commented out (line 18)
- `/docs/prd/epics-and-stories.md` — commented out (line 19)

**NOT in CODEOWNERS at all:**
- `docs/decisions/BOARD.md` — intentionally removed per Story 74.1 post-merge amendment. R-015 researched CI-based protection as replacement.

## Why These Files Are Unprotected

**ROADMAP.md, epic-list.md, epics-and-stories.md:** These are the planning doc chain maintained by project-watchdog and PM agents. Protecting them via CODEOWNERS would block the automated planning pipeline (D-162). They were likely commented out during or after Epic 74 implementation when the team discovered this conflict.

**BOARD.md:** Explicitly unprotected per Story 74.1 amendment — nearly every agent PR touches it (research, planning, implementation all write decision entries). R-015 (gentle-tiger) researched CI-based section protection as a long-term alternative (D-190).

## Documentation That Needs Updating

### 1. CLAUDE.md (lines 46-55)
**Current (WRONG):** Lists ROADMAP.md, epic-list.md, epics-and-stories.md, BOARD.md as protected
**Fix:** Move all four to the "Unprotected" section with explanatory notes

### 2. agents/merge-queue.md (lines 17, 21)
**Current (WRONG):** Protected paths list and grep pattern include all four files
**Fix:** Remove from protected paths, remove from grep pattern. This is the ROOT CAUSE of incorrect `status.needs-human` labeling.

### 3. agents/pr-shepherd.md (line 32)
**Current (WRONG):** Lists all four as CODEOWNERS-protected
**Fix:** Remove from protected files list

### 4. docs/branch-protection.md (line 21)
**Current:** Says "No CODEOWNERS file configured" — stale from pre-Epic 74
**Fix:** Update to reflect current CODEOWNERS configuration (low priority, informational doc)

## Impact Assessment

**Severity: Medium-High**
- merge-queue's grep pattern on line 21 of merge-queue.md is the primary damage vector
- Every governance sync PR touching ROADMAP.md gets incorrectly labeled `status.needs-human`
- This blocks automated merges and creates unnecessary human review burden
- The fix is straightforward: update 3 documentation files

## Adopted Approach

Update CLAUDE.md, merge-queue.md, and pr-shepherd.md to match actual .github/CODEOWNERS state. No changes to .github/CODEOWNERS itself — the commented-out state is correct.

## Rejected Alternatives

1. **Re-protect these files in CODEOWNERS** — Rejected because it would block the automated planning pipeline (project-watchdog, PM agents). The files were intentionally unprotected.
2. **Add CI-based protection for planning docs** — Out of scope for this fix. Could be a future enhancement (similar to R-015's approach for BOARD.md).
3. **Make merge-queue read .github/CODEOWNERS directly** — Overly complex. The hardcoded list approach works fine when kept in sync. A drift-detection CI check would be a better long-term fix.
