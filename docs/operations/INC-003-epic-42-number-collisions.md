# INC-003: Epic 42 Number Collision Saga

**Date:** 2026-03-09 to 2026-03-10 (ongoing)
**Severity:** High
**Duration:** ~21 hours (first collision 2026-03-09T23:33Z through halt order 2026-03-10T20:12Z)
**Status:** Resolved (PR #419 closed, renumber PR #421 merged, Epic 42 now allocated to Security Hardening)

---

## Summary

Four parallel `/plan-work` workers self-assigned Epic 42 to three different features within a 2-minute window on 2026-03-09. The collision went partially detected — merge-queue flagged one pair — but the full scope wasn't understood for nearly 21 hours. The resulting chaos produced 1 wasted PR (closed with 7 merge conflicts after sitting for 20+ hours), 1 renumber PR, decision board inconsistencies, and a stale Epic Number Registry. This was the project's worst process failure, occurring during its busiest period (390+ merged PRs).

---

## The Three Claimants

| Feature | Worker | Branch | PR | Claim Time (UTC) | Outcome |
|---------|--------|--------|-----|-------------------|---------|
| Door-Like Doors | gentle-tiger | work/gentle-tiger | #417 | 2026-03-09T23:33:47Z | Merged first, then renumbered to Epic 48 (PR #421) |
| Application Security Hardening | witty-owl | work/witty-owl | #418 | 2026-03-09T23:33:52Z | Merged second, kept Epic 42 (supervisor decision) |
| ThreeDoors Doctor Command | silly-tiger | work/silly-tiger | #419 | 2026-03-09T23:36:02Z | Closed with 7 merge conflicts after 20+ hours |

A fourth worker (zealous-penguin, PR #420) originally planned Epics 42-46 for data source setup UX, but self-corrected to Epics 43-47 after noticing Epic 42 was already claimed. PR #420's body explicitly notes: "Epic 42 was claimed by another agent for 'Door-Like Doors' — renumbered from 42-46 to 43-47 to avoid collision." This was the one worker that actually checked.

---

## Timeline

### 2026-03-09T22:56–22:59 UTC — Research PRs Merge

- **PR #412** merged (doors-more-doorlike party mode research)
- **PR #413** merged (doctor command research report)
- A security audit research artifact was also available
- These research outputs fed into the `/plan-work` workers dispatched next

### 2026-03-09T23:33 UTC — The 2-Minute Collision Window

All three Epic 42 claims happened within a 2-minute, 15-second span:

| Time | Event |
|------|-------|
| 23:33:23 | gentle-tiger commits "Epic 42 — Door-Like Doors" (on branch) |
| 23:33:28 | witty-owl commits "Epic 42 — Application Security Hardening" (on branch) |
| 23:33:47 | **PR #417** created (Door-Like Doors as Epic 42) |
| 23:33:52 | **PR #418** created (Security Hardening as Epic 42) |
| 23:34:21 | **PR #417 merged** — Door-Like Doors is now Epic 42 on main |
| 23:35:39 | silly-tiger commits "formalize doctor command research into Epic 42" |
| 23:36:02 | **PR #419** created (Doctor Command as Epic 42) — already collides with merged #417 AND pending #418 |

**5 seconds separated creation of PRs #417 and #418.** Both had Epic 42 story files (42.1–42.4 and 42.1–42.5 respectively), both updated ROADMAP.md, epic-list.md, and epics-and-stories.md, and both reserved Epic 42 in the BOARD.md registry. PR #417 merged 35 seconds after creation.

### 2026-03-09T23:44 UTC — Partial Self-Correction

- **PR #420** (zealous-penguin, data source setup UX) detects Epic 42 is taken
- Self-renumbers from Epics 42-46 to Epics 43-47
- **PR #420 merged** at 23:45 — the only worker to check the registry before claiming

### 2026-03-09T23:44 UTC — Merge-Queue Detection

Merge-queue posts a comment on PR #419:

> **Merge-queue flag:** Story number collision detected. Both PR #419 (ThreeDoors Doctor) and PR #418 (Application Security Hardening) use Epic 42 story numbers 42.1–42.5 for different features.

This identified the #418/#419 collision but did not address that #417 (Door-Like Doors) had already merged as Epic 42 and also collided.

### 2026-03-10T00:21–00:39 UTC — First Renumber Attempt

- **PR #421 created and merged** — renumbers Door-Like Doors from Epic 42 → Epic 48
  - Renames story files 42.1–42.4 → 48.1–48.4
  - Updates ROADMAP.md, epic-list.md, epics-and-stories.md
  - Updates BOARD.md decision D-141 and rejected X-080 through X-083
  - This frees Epic 42 for Security Hardening

- **PR #418 merged** at 00:39 — Security Hardening claims Epic 42 with stories 42.1–42.5
  - Creates decision D-153 on BOARD.md

- **PR #423 merged** at 00:39 — agent hardening (INC-001 follow-up), unrelated but concurrent

### 2026-03-10T00:39 UTC → 20:12 UTC — PR #419 Sits with 7 Merge Conflicts

PR #419 (Doctor Command, 14 changed files, 1043 insertions) sat open for nearly 20 hours with 7 merge conflicts. It had Epic 42 story files 42.1–42.10 (overlapping with Security Hardening's 42.1–42.5 that were now on main). Nobody caught the root cause — the PR was a zombie, dead on arrival from the moment PR #417 merged.

### 2026-03-10T20:12 UTC — Final Resolution

PR #419 closed with comment:

> Closing: Epic 42 number collision — Epic 42 is already allocated to Application Security Hardening (D-153). Recreating this content as **Epic 49: ThreeDoors Doctor — Self-Diagnosis Command** in a new PR.

---

## Forensic Statistics

### Duration

- **First collision to resolution:** ~21 hours (23:33 UTC Mar 9 → 20:12 UTC Mar 10)
- **PR #419 sitting dead with merge conflicts:** ~20.5 hours
- **Time between first claim and renumber PR:** ~47 minutes (#417 created → #421 merged)
- **Time between renumber and PR #419 closure:** ~19.5 hours

### PRs Involved

| PR | Purpose | Status | Wasted? |
|----|---------|--------|---------|
| #417 | Door-Like Doors as Epic 42 | Merged, then content renumbered | Partially — content preserved but numbering wrong |
| #418 | Security Hardening as Epic 42 | Merged (final owner) | No |
| #419 | Doctor Command as Epic 42 | **Closed (7 conflicts)** | **Yes — 14 files, 1043 lines, fully wasted** |
| #420 | Data Source UX (self-corrected from 42-46 to 43-47) | Merged | No — self-corrected |
| #421 | Renumber Door-Like Doors 42→48 | Merged | Necessary cleanup |

**Total PRs created due to collision:** 5 (3 colliding + 1 renumber + 1 self-correcting)
**Wasted PRs:** 1 (PR #419, Doctor Command — fully dead)
**Renumber PRs:** 1 merged (#421), at least 1 more needed (Doctor → Epic 49, not yet created)

### Worker Dispatches

- **4 `/plan-work` workers** dispatched concurrently (gentle-tiger, witty-owl, silly-tiger, zealous-penguin)
- **1 renumber worker** dispatched (fix/renumber-doorlike-doors-epic42-to-48)
- **At least 1 additional worker** will be needed to recreate Doctor Command as Epic 49
- **Total worker dispatches touching Epic 42:** 6+

### Decision Board Damage

- D-141 was claimed by both Door-Like Doors and Doctor Command
- Security Hardening created D-153 instead (avoiding the D-141 collision)
- D-141 was updated during renumber to reference Epic 48 instead of 42
- The Epic Number Registry on BOARD.md **is still stale** — shows Epic 42 as "Door-Like Doors" with status "Stories created" even though Door-Like Doors is now Epic 48 and Epic 42 is Security Hardening

### Lines of Code Wasted

- PR #419: **1,043 lines** across 14 files — all wasted (will need to be recreated as Epic 49)
- PR #421: **~300 lines** of renaming overhead
- Total throwaway work: **~1,300+ lines**

---

## Root Cause Analysis

### Primary Cause: No Mutex on Epic Number Allocation

Four parallel `/plan-work` workers were dispatched simultaneously. Each worker independently:

1. Read the Epic Number Registry on BOARD.md
2. Saw Epic 42 was the next available number
3. Assigned Epic 42 to their feature
4. Created story files, updated planning docs, and submitted a PR

There was no locking mechanism. The registry is a markdown table — it can't enforce mutual exclusion. Multiple agents reading it at the same moment all see the same "next available" number.

### Secondary Cause: Race Condition in Merge Timing

PR #417 merged 35 seconds after creation, before PR #418 was even submitted. But both workers had already committed their changes with Epic 42 before either PR existed. The collision was baked in at commit time, not PR time. Even if merge-queue had been faster to detect, the damage was done.

### Tertiary Cause: No Pre-Merge Validation for Epic Numbers

Merge-queue did not check for epic number conflicts before merging PR #417. It merged a PR that created `42.1.story.md` through `42.4.story.md` without verifying that no other pending PR also created files with those names. This allowed the first collision to land on main.

### Contributing Factors

1. **Registry was advisory, not enforced.** The Epic Number Registry (created PR #327, merged 2026-03-09T05:03Z) existed for less than 19 hours before this collision. Its rules said "Reserve the number here FIRST, before creating story files" — but nothing enforced this. Workers read it, didn't write to it first, and all grabbed the same number.

2. **project-watchdog as mutex was declared but not implemented.** MEMORY.md stated "project-watchdog is the MUTEX for epic and story numbers — sole allocator, no exceptions." But this was a memory-only policy — it wasn't in any agent definition file, wasn't enforced by tooling, and the `/plan-work` workers didn't know about it.

3. **Workers were dispatched concurrently.** The supervisor dispatched all four `/plan-work` workers in the same batch. Sequential dispatch with a "claim your epic number first" step would have prevented the collision.

4. **This was a repeat incident.** Epic 39 had the exact same collision (PR #327 documents it): two workers claimed Epic 39 for different features, requiring a renumber to Epic 40. The registry was the mitigation from that incident — but it was necessary-not-sufficient.

---

## Why Mitigations Failed

### Mitigation 1: Epic Number Registry (BOARD.md)

**Created:** PR #327, merged 2026-03-09T05:03Z
**How it was supposed to work:** Workers check the registry, find the next available number, reserve it, then proceed.
**Why it failed:** The registry is a markdown file. It has no locking, no compare-and-swap, no transaction semantics. Four workers reading it simultaneously all see the same state. Writing to it requires a PR, which creates its own race condition (whose PR merges first?).

### Mitigation 2: project-watchdog as Allocator

**Created:** Supervisor MEMORY.md entry, ~2026-03-09
**How it was supposed to work:** All epic number requests go through project-watchdog, which maintains a single serial allocation sequence.
**Why it failed:** This was a soft policy in one agent's memory, not a hard constraint. The `/plan-work` workers had no instruction to "ask project-watchdog for a number." They followed their own logic: read the registry, pick the next number, proceed.

### Mitigation 3: Merge-Queue Collision Detection

**When it worked:** Merge-queue correctly flagged the #418/#419 collision in a PR comment.
**Why it was insufficient:** It detected the collision *after* PR #417 had already merged. Detection without prevention only limits blast radius — it doesn't prevent the collision itself.

---

## Residual Damage (as of 2026-03-10)

1. **Epic Number Registry is stale.** BOARD.md still shows Epic 42 as "Door-Like Doors — Visual Door Metaphor Enhancement" with status "Stories created." The actual state is:
   - Epic 42 = Application Security Hardening (D-153)
   - Epic 48 = Door-Like Doors (renumbered via PR #421)
   - Epic 49 = Doctor Command (not yet created — PR #419 closed, content needs recreation)

2. **Doctor Command content is orphaned.** The 10 story files and planning doc updates from PR #419 exist only on the closed `work/silly-tiger` branch. They need to be recreated as Epic 49 in a new PR.

3. **Decision ID D-141 was overloaded.** Both Door-Like Doors and Doctor Command claimed D-141. The renumber PR updated D-141 to reference Epic 48 (Door-Like Doors). Doctor Command's D-141 through D-145 (on the closed PR #419 branch) will need new IDs when recreated.

4. **Registry row 48 says "next available"** — but 48 is now Door-Like Doors. The registry is completely out of sync with reality.

---

## Comparison with Epic 39 Collision

| Dimension | Epic 39 (PR #327) | Epic 42 (this incident) |
|-----------|-------------------|-------------------------|
| **Workers involved** | 2 | 4 |
| **Features colliding** | 2 (Keybinding Display, Beautiful Stats) | 3 (Door-Like Doors, Security Hardening, Doctor Command) |
| **Time to detect** | Unknown (post-merge discovery) | ~11 minutes (merge-queue comment) |
| **Renumber PRs** | 1 (Epic 39 → 40) | 1 merged (#421), 1+ pending |
| **Wasted PRs** | 0 (both features preserved) | 1 (PR #419 fully wasted) |
| **Mitigation created** | Epic Number Registry | project-watchdog mutex (policy), halt order |
| **Did mitigation prevent recurrence?** | **No** — Epic 42 collision happened 18 hours later | TBD |

The pattern is identical: parallel workers self-assign, no mutex, collision on merge. The Epic 39 fix (a registry) was necessary but not sufficient. The Epic 42 fix (a designated allocator) was declared but not enforced.

---

## Lessons Learned

### 1. Advisory registries don't prevent races

A markdown table that says "check here first" cannot prevent concurrent readers from seeing the same state. Registries work for human-speed coordination (one person checks, allocates, moves on). They fail at agent-speed coordination (four agents check simultaneously).

**Recommendation:** Epic number allocation must go through a serialized chokepoint — either the supervisor dispatches workers with pre-assigned numbers, or a single agent (project-watchdog) responds to allocation requests sequentially.

### 2. Soft policies in MEMORY.md don't reach workers

Declaring "project-watchdog is the mutex" in supervisor memory doesn't help if workers don't know about it. Workers follow their task description and their agent definition — not the supervisor's memory.

**Recommendation:** If a policy must be enforced across all workers, it must be in the worker agent definition (`agents/worker.md`) or in the task description itself. Memory-only policies are invisible to other agents.

### 3. Concurrent dispatch of epic-creating workers is inherently unsafe

The collision window was 2 minutes. No registry, no mutex, no detection system can reliably prevent races in a 2-minute window when 4 agents are writing to the same files.

**Recommendation:** Never dispatch more than one `/plan-work` worker at a time, OR pre-assign epic numbers in the task description before dispatch. Sequential dispatch with pre-assigned numbers eliminates the race entirely.

### 4. Merge-queue needs pre-merge conflict detection for story files

Merge-queue caught the collision after the fact. It should refuse to merge a PR that creates story files `N.X.story.md` if another pending PR also creates story files with the same epic number N.

**Recommendation:** Add a pre-merge check: scan all open PRs for overlapping story file names before merging any planning PR.

### 5. Post-incident mitigations must be validated, not just declared

The Epic 39 collision produced a mitigation (the registry). The registry was never tested under concurrent load. When 4 workers hit it simultaneously 18 hours later, it failed exactly as predicted by computer science (readers-writers problem without locking).

**Recommendation:** After creating a mitigation, mentally simulate the failure mode again. Ask: "If the same thing happened right now, would this mitigation actually prevent it?" If the answer involves humans remembering to follow a process, it will fail under pressure.

### 6. Zombie PRs waste context and attention

PR #419 sat open with 7 merge conflicts for 20+ hours. Nobody closed it because nobody understood the root cause was a number collision — it looked like a routine merge conflict. Every agent that checked PR status saw it and spent cycles considering whether to rebase it.

**Recommendation:** When a collision is detected, immediately close the losing PRs with a clear explanation. Don't leave them open hoping someone will fix the conflicts.

---

## Recommended Process Changes

### Immediate

1. **Fix the Epic Number Registry.** Update BOARD.md to reflect reality:
   - Epic 42 = Application Security Hardening (D-153)
   - Epic 48 = Door-Like Doors (D-141)
   - Epic 49 = Doctor Command (pending creation)
   - Next available = 50

2. **Recreate Doctor Command as Epic 49.** Salvage content from the closed `work/silly-tiger` branch, renumber all references from 42 to 49, assign new decision IDs.

### Structural

3. **Pre-assign epic numbers in worker task descriptions.** Supervisor (or project-watchdog) allocates the number BEFORE dispatching the worker. Worker task includes: "You are creating Epic N. Do not change this number."

4. **Sequential dispatch for planning workers.** Never dispatch more than one `/plan-work` worker simultaneously. If multiple features need planning, dispatch them one at a time, waiting for each to merge before dispatching the next.

5. **Add pre-merge story file check to merge-queue.** Before merging any PR that creates `docs/stories/N.*.story.md` files, check all other open PRs for files with the same epic number N.

6. **Add epic number allocation to worker.md agent definition.** Workers must never self-assign epic numbers. The rule must be in the agent definition, not just supervisor memory.

---

## Related

- [INC-001: pr-shepherd Contamination of Shared Checkout](INC-001-pr-shepherd-contamination.md) — occurred concurrently, contributed to delayed detection
- [INC-002: Destructive Git Sync Override](INC-002-destructive-git-sync-override.md) — another process failure from the same period
- PR #327: Epic Number Registry creation (response to Epic 39 collision)
- PR #417: Door-Like Doors as Epic 42 (first claim, merged)
- PR #418: Security Hardening as Epic 42 (second claim, merged as final owner)
- PR #419: Doctor Command as Epic 42 (third claim, closed — 7 merge conflicts)
- PR #420: Data Source UX (self-corrected from 42-46 to 43-47)
- PR #421: Renumber Door-Like Doors from Epic 42 → Epic 48
- BOARD.md Epic Number Registry: created to prevent this, failed under concurrent load
