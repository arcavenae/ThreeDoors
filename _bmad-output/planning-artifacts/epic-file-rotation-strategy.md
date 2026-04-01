# Epic File Rotation Strategy — Party Mode Artifact

**Date:** 2026-03-31
**Participants:** Winston (Architect), John (PM), Amelia (Dev), Bob (SM), Murat (Test Architect), BMad Master
**Trigger:** docs/prd/epics-and-stories.md reached 7,198 lines with 77 epics (26 COMPLETE). Agents consuming massive context reading stale completed epics. Planning doc updates slow and error-prone.

---

## Root Cause Analysis

The file `docs/prd/epics-and-stories.md` is an **unbounded append-only data structure** serving two conflicting purposes:

1. **Historical record** — what was planned, what was delivered
2. **Active work reference** — what agents need to know about current scope

**Key metrics:**
- Total lines: 7,198
- Preamble (requirements inventory): 233 lines
- COMPLETE epic lines: 2,336 (32% of file)
- Active/other epic lines: 4,629 (64% of file)
- COMPLETE epics: 26 of 77 total
- Largest single epic: Epic 0 (600 lines — infrastructure backfill)

**Accumulation pattern:** New epics are added as `/plan-work` creates them. Status transitions mark epics COMPLETE but never remove them. No pruning mechanism exists. The file grows monotonically.

**Impact on agents:**
- project-watchdog reads entire file for number allocation (needs ~80 lines)
- merge-queue reads entire file for scope checks (needs epic names only)
- research-supervisor already truncates to ~10KB (acknowledging the problem)
- Workers never need completed epics (they read story files)
- `/plan-work` only needs active section for new epic placement

---

## Adopted Approach: Two-File Split with Epic Index Table

### Architecture

```
docs/prd/
  epics-and-stories.md          # Active epics only (~4,800 lines → shrinks over time)
  epics-and-stories-archive.md  # Completed + icebox epics (~2,400 lines → grows slowly)
  epic-list.md                  # Unchanged (compact summary, already exists)
```

### Active File Structure (epics-and-stories.md)

```markdown
# ThreeDoors - Epic Breakdown

## Quick Reference
- **Total Epics:** 77 (26 complete, 50 active, 1 icebox)
- **Active File:** This file — contains all non-COMPLETE epics
- **Archive File:** epics-and-stories-archive.md (completed + icebox epics)
- **Last Audit:** 2026-03-31

## Epic Index (All Epics — Active + Archived)
| # | Name | Status | Location |
|---|------|--------|----------|
| 0 | Infrastructure & Process | COMPLETE | archive |
| 1 | Three Doors Technical Demo | COMPLETE | archive |
| ... | ... | ... | ... |
| 42 | Application Security Hardening | In Progress | active |
| ... | ... | ... | ... |

## Requirements Inventory
[Existing preamble — maps FRs to epics, stays in active file]

## Epic 33: Seasonal Door Theme Variants
[Full story breakdowns for active epics only]
...
```

### Archive File Structure (epics-and-stories-archive.md)

```markdown
# ThreeDoors - Completed Epic Archive

This file contains full story breakdowns for completed and icebox epics.
For active epics, see `epics-and-stories.md`.
For the complete epic index, see the Epic Index table in `epics-and-stories.md`.

## Epic 0: Infrastructure & Process (Backfill) - COMPLETE
[Full story breakdown preserved]

## Epic 1: Three Doors Technical Demo - COMPLETE
[Full story breakdown preserved]
...

## Epic 16: iPhone Mobile App (SwiftUI) — ICEBOX
[Full details preserved]
```

### Archival Rules

1. **What moves to archive:** Epics with status COMPLETE or ICEBOX, including all their story breakdowns, ACs, and implementation notes
2. **What stays in active:** All other epics (In Progress, Not Started, Blocked, etc.) with full detail
3. **What spans both:** The Epic Index table in the active file references ALL epics with their location (active/archive)
4. **Trigger:** project-watchdog moves epics to archive as part of its existing `SYNC_OPERATIONAL_DATA` handler when an epic's status transitions to COMPLETE
5. **No back-migration:** Reopened epics get a NEW section in the active file with a cross-reference to the archive (e.g., "Reopened from archive — see archive for historical context")

---

## Agent Definition Updates Required

### merge-queue.md
- **Current:** Checks `docs/prd/epics-and-stories.md` in CODEOWNERS-protected path list
- **Change:** Add `docs/prd/epics-and-stories-archive.md` to protected paths
- **Scope check:** Read Epic Index table from active file (not full file) for scope validation

### pr-shepherd.md
- **Current:** Lists `docs/prd/epics-and-stories.md` in CODEOWNERS description
- **Change:** Add archive file to CODEOWNERS description

### project-watchdog.md
- **Current:** Reads full file for number allocation and status sync
- **Change:** 
  - Number allocation reads Epic Index table only
  - Status sync updates active file or archive depending on epic status
  - New responsibility: move COMPLETE epics from active to archive during sync
  - Add SYNC_OPERATIONAL_DATA step: "Check for newly COMPLETE epics → move to archive"

### research-supervisor.md
- **Current:** Loads `epic-list.md` headers (~10KB truncated)
- **Change:** Already truncates — minimal change. Add archive file as optional context source.

### retrospector.md
- **Current:** Cross-checks full planning doc chain
- **Change:** Cross-check both active and archive files for completeness

### worker.md
- **Current:** Forbidden from editing `epics-and-stories.md`
- **Change:** Also forbidden from editing `epics-and-stories-archive.md`

---

## BMAD Framework Changes Required

### Workflow Steps (create-epics-and-stories)
- `step-01-validate-prerequisites.md` — Read Epic Index from active file for prerequisite check
- `step-02-design-epics.md` — Append new epics to active file only
- `step-03-create-stories.md` — Write stories to active file only
- `step-04-final-validation.md` — Validate against active file; verify Epic Index updated

### Agent Definition (bmm/agents/pm.md)
- Add knowledge of two-file structure
- PM writes only to active file; project-watchdog handles archival

### Skill Updates
- `/plan-work` — reads active file for epic placement (no archive needed)
- `/implement-story` — reads story files (no change needed)
- `/reconcile-docs` — must check BOTH active and archive for completeness
- `/bmad-bmm-sprint-status` — reads active file only (completed epics irrelevant to sprint)

---

## CODEOWNERS Update

```
# Add to .github/CODEOWNERS
docs/prd/epics-and-stories-archive.md @skippy
```

---

## Migration Plan

### Phase 1: Create Archive and Split (Single PR)
1. Create `docs/prd/epics-and-stories-archive.md` with header
2. Move all 26 COMPLETE epics + Epic 16 (ICEBOX) to archive
3. Add Epic Index table to top of active file
4. Add Quick Reference section
5. Update requirements inventory preamble if needed
6. Verify: no epic appears in both files, all 77 in index

### Phase 2: Update Agent Definitions (Same PR or Follow-up)
1. Update all 6 agent definition files
2. Update CODEOWNERS
3. Update CLAUDE.md references

### Phase 3: Update BMAD Framework (Same PR or Follow-up)
1. Update 4 workflow step files
2. Update PM agent definition
3. Update `/reconcile-docs` skill

### Phase 4: Verification
1. Run `/reconcile-docs` — should find no drift
2. Verify agent scope checks work (merge-queue test)
3. Verify project-watchdog can allocate numbers from index
4. Verify line counts: active ~4,800, archive ~2,400

**Note:** Phases 1-3 touch CODEOWNERS-protected files and require @skippy approval. Recommend bundling into a single PR for atomic review.

---

## Rejected Alternatives

### Per-Epic Files (One File Per Epic + Index)
**Proposed:** `docs/prd/epics/00-infrastructure.md`, `docs/prd/epics/01-three-doors-demo.md`, etc. with `docs/prd/epics/index.md`.

**Rejected because:**
- Creates 77+ files — massive directory sprawl
- Breaks every `grep` pattern agents use for scope checks
- Atomic updates across multiple stories require multi-file commits
- Index file becomes its own maintenance burden
- The coordination cost is enormous for minimal benefit over two-file split
- Over-engineered for a problem solved by a simple split

### Time-Based Rotation (Quarterly Archives)
**Proposed:** `epics-q1-2026.md`, `epics-q2-2026.md`, etc.

**Rejected because:**
- Epics don't correlate to calendar quarters (Epic 0 created at project start, Epic 77 last week)
- Mixed active/complete epics in each quarterly file
- Agents need temporal knowledge (which quarter?) to find an epic
- Adds arbitrary coupling between time and document structure
- No natural "rotation event" — quarters are meaningless to epic lifecycle

### Size-Based Triggers (Archive When File Exceeds N Lines)
**Proposed:** When `epics-and-stories.md` exceeds 5,000 lines, archive oldest completed epics.

**Rejected because:**
- Threshold is arbitrary — why 5,000 and not 4,000 or 6,000?
- Doesn't solve structural problem, just delays it
- Requires monitoring infrastructure for a document
- Reactive rather than structural — you'd repeatedly hit the trigger
- "Oldest completed" may not be the right selection criteria

---

## Next Steps

1. **Request epic number from supervisor** for this work (do NOT self-assign)
2. Create story file for the implementation
3. Implement the two-file split with Epic Index table
4. Update all agent definitions and BMAD workflow files
5. Create PR (will require @skippy review due to CODEOWNERS protection)

---

## Provenance
- **Autonomy Level:** L2 (AI-paired — human provided direction via task assignment)
- **Method:** Party mode discussion with full BMAD agent panel
- **Review:** Human review required (touches CODEOWNERS-protected files)
