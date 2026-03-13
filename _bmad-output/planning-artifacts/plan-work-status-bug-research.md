# /plan-work Story Status Bug — Research & Recommendations

**Date:** 2026-03-12
**Researcher:** gentle-otter (worker)
**Trigger:** Story 0.55 (Cron-Based Agent Heartbeat System) was marked `Done (PR #681)` by kind-badger running `/plan-work`, despite zero implementation tasks being completed. All subsequent `/implement-story` workers saw "Done" and bailed.

---

## 1. Root Cause Analysis

### The Bug

Story 0.55 has 6 implementation tasks (add HEARTBEAT handlers to agent definitions, create polling loops, update MEMORY.md, update ops docs). PR #681 contains **only the story file** — a pure planning artifact. Yet the story status reads `Done (PR #681)`.

### Why It Happened — Competing Instructions

The worker running `/plan-work` receives instructions from **three sources** that conflict:

| Source | Instruction | Effect |
|--------|-------------|--------|
| **Worker definition** (`agents/worker.md`, step 5) | "Update the story file status to `Done (PR #NNN)`" | Worker marks story Done after creating PR |
| **CLAUDE.md** (line 43) | "After implementation, update the story file status to `Done (PR #NNN)`" | Says "after implementation" — but worker reads this as "after my task" |
| **`/plan-work` skill** (`.claude/commands/plan-work.md`) | **No status instruction at all** | Silent — defers to worker's default behavior |

The worker definition's step 5 is a blanket instruction: "after creating your PR, mark the story Done." It doesn't distinguish between *planning* PRs and *implementation* PRs. The worker completed its assigned task (planning), created a PR, and followed step 5 — marking the story Done.

### The Deeper Problem

The system conflates **two different concepts**:

1. **Worker task completion** — "I finished the task I was assigned" (planning)
2. **Story completion** — "All acceptance criteria are met and code is implemented"

A `/plan-work` worker's task is to *create* a story, not to *implement* it. But the worker definition's step 5 doesn't know the difference.

---

## 2. Flow Trace: `/plan-work` → Story Status

```
Supervisor dispatches worker with /plan-work task
  → Worker reads agents/worker.md (its system prompt)
  → Worker executes /plan-work skill (8 phases)
    → Phase 6: Story Creation — creates docs/stories/X.Y.story.md
    → Phase 8: Create PR — commits and creates PR #NNN
  → Worker follows agents/worker.md step 5:
    "Update the story file status to Done (PR #NNN)"  ← BUG
  → Worker runs multiclaude agent complete
```

### What `/plan-work` Phase 8 says (relevant excerpt):

> 1. Stage all changed/created files: [...] Story files [...]
> 2. Create a descriptive commit
> 3. Push the branch and create a PR

No mention of story status. The skill is silent, so the worker falls through to its default behavior (step 5 of its workflow).

### What `/implement-story` Phase 8 says:

> 1. Update the story file status to `In Review (PR #NNN)` (will be updated to `Done` after merge).

This is more nuanced — it uses "In Review" not "Done", and notes Done happens after merge. But note: workers running `/implement-story` also get the worker definition's step 5 which says "Done" unconditionally. The `/implement-story` instruction and the worker definition contradict each other.

---

## 3. Flow Trace: `/implement-story` → Story Status

```
Supervisor dispatches worker with /implement-story task
  → Worker reads agents/worker.md (its system prompt)
  → Worker executes /implement-story skill (8 phases)
    → Phase 1-7: Story prep, enrichment, TDD, implementation, review
    → Phase 8: "Update story status to In Review (PR #NNN)"
  → Worker ALSO has agents/worker.md step 5:
    "Update the story file status to Done (PR #NNN)"
  → Contradiction: skill says "In Review", worker def says "Done"
  → In practice: workers have been setting "Done" (worker def wins)
```

The `/implement-story` skill's Phase 8 is actually correct in spirit — "In Review" until merge. But it's overridden by the worker definition. In practice this hasn't caused problems because the story IS implemented at that point, but it's still a contradiction.

---

## 4. Audit: Other Stories Incorrectly Marked Done

### Methodology

Searched for stories marked "Done" where the associated PR was docs-only (story file creation only, no implementation code).

### Confirmed False Positives

**Story 0.55** (PR #681) — Heartbeat system. 6 implementation tasks, zero done. **CONFIRMED BUG.**

### Likely False Positives (0.x Infrastructure Stories)

These 0.x stories need manual review — they describe implementation tasks but their PRs may have been docs-only:

- **Story 0.50** (PR #556) — needs investigation
- **Story 0.51** (PR #623) — needs investigation
- **Story 0.52** (PR #619) — Multi-Adapter Integration Tests — needs investigation
- **Story 0.53** (PR #621) — Docker E2E — needs investigation

### Not False Positives

Many stories in the 17.x-62.x range appear to have correct Done status — their PRs contain both story file updates AND implementation code. The audit tool's grep matches story file changes in implementation PRs, which is expected.

---

## 5. Recommendations

### R1: Add Explicit Status Instructions to `/plan-work` [HIGH PRIORITY]

**File:** `.claude/commands/plan-work.md`

In Phase 6 (Story Creation), add:

```markdown
4. Set the status of each newly created story to `Not Started`.
   - Planning workers NEVER set story status to `Done` — that is reserved
     for implementation workers who have completed all acceptance criteria.
```

In Phase 8 (Create PR), add:

```markdown
0. **IMPORTANT:** Do NOT update story status to `Done`. Story files created
   by /plan-work should have status `Not Started` or `Draft`. Only
   /implement-story sets `Done` after all ACs are met.
```

### R2: Differentiate Worker Step 5 by Task Type [HIGH PRIORITY]

**File:** `agents/worker.md`

Change step 5 from:

> 5. Update the story file status to `Done (PR #NNN)`

To:

> 5. Update story file status:
>    - **Implementation tasks** (`/implement-story`, feature work, bug fixes): Set status to `Done (PR #NNN)` after all acceptance criteria are met
>    - **Planning tasks** (`/plan-work`, story creation, docs-only): Set newly created story status to `Not Started` — NEVER `Done`
>    - **Rule:** A story is only `Done` when its acceptance criteria are implemented in code, not when the story file is created

### R3: Add Guard to `/implement-story` Phase 1 [MEDIUM PRIORITY]

**File:** `.claude/commands/implement-story.md`

In Phase 1 (Story Preparation), after finding the story file, add a check:

```markdown
4. If the story file has status `Done`:
   - HALT. A story marked Done should not be re-implemented.
   - Check if the Done status is correct by examining the referenced PR.
   - If the PR is docs-only (no implementation code), the Done status is
     incorrect — reset to `Not Started` and proceed.
   - If the PR contains implementation, skip — the story is already done.
```

This prevents the cascading failure where workers see "Done" and bail.

### R4: Add Separate Fields for Planning vs Implementation PRs [LOW PRIORITY]

**Current format:**
```markdown
## Status: Done (PR #681)
```

**Proposed format:**
```markdown
## Status: Not Started
**Planning PR:** #681
**Implementation PR:** —
```

This makes the distinction explicit and machine-readable. However, this requires changes to the story template (`/new-story`), the status parser (`internal/docaudit/parse_story_files.go`), and all downstream tools. The cost is high relative to the benefit — R1+R2+R3 solve the immediate problem without schema changes.

**Recommendation:** Defer R4 unless the bug recurs after R1-R3.

### R5: Update CLAUDE.md Status Language [MEDIUM PRIORITY]

**File:** `CLAUDE.md`

The current language says:

> After implementation, update the story file status to `Done (PR #NNN)`

This is correct but insufficient. Add a clarification:

> - **`Done (PR #NNN)` means implementation is complete** — all acceptance criteria met in code
> - Planning-only PRs (story creation, docs updates, research) do NOT qualify for `Done` status
> - `/plan-work` creates stories with status `Not Started`; only `/implement-story` sets `Done`

### R6: Validate ACs Before Setting Done [LOW PRIORITY]

Should `/implement-story` verify acceptance criteria are actually met before setting Done?

It already does — Phase 7 (Acceptance Review) checks every AC. The issue isn't with `/implement-story` missing validation; it's with `/plan-work` setting Done when no validation was needed (because no implementation happened).

R1-R3 fix the actual bug. AC validation in `/implement-story` is already adequate.

---

## 6. Immediate Fix for Story 0.55

Reset `docs/stories/0.55.story.md` status from `Done (PR #681)` to `Not Started`. This unblocks `/implement-story` workers.

---

## 7. Summary of Recommended Changes

| ID | File | Change | Priority |
|----|------|--------|----------|
| R1 | `.claude/commands/plan-work.md` | Add explicit "set status to Not Started" in Phase 6 & Phase 8 | HIGH |
| R2 | `agents/worker.md` | Differentiate step 5 by task type (planning vs implementation) | HIGH |
| R3 | `.claude/commands/implement-story.md` | Add Done-status guard in Phase 1 | MEDIUM |
| R5 | `CLAUDE.md` | Clarify Done means implementation, not planning | MEDIUM |
| R4 | Story template + parser | Separate Planning PR / Implementation PR fields | LOW (defer) |
| R6 | `/implement-story` | Add AC validation before Done | LOW (already exists) |
| Fix | `docs/stories/0.55.story.md` | Reset status to `Not Started` | IMMEDIATE |

---

## 8. Files Examined

- `.claude/commands/plan-work.md` — /plan-work skill definition (8 phases, no status instruction)
- `.claude/commands/implement-story.md` — /implement-story skill definition (says "In Review", contradicts worker def)
- `.claude/commands/new-story.md` — Story template (sets initial status to "Draft")
- `agents/worker.md` — Worker agent definition (step 5: unconditional "Done")
- `CLAUDE.md` — Project instructions (says "after implementation" but ambiguous)
- `docs/stories/0.55.story.md` — The affected story (6 implementation tasks, zero completed)
- `internal/docaudit/parse_story_files.go` — Status parser (recognizes Done, Not Started, Draft, etc.)
