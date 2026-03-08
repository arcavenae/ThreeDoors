# Party Mode: Undo Task Completion Discussion

**Date:** 2026-03-08
**Topic:** Epic 32 — Undo Task Completion
**Participants:** PM, Architect, Dev, QA, UX Designer, SM

---

## Discussion Summary

### PM (Product Manager)

The undo feature was explicitly flagged in our Phase 1 Validation Gate as a pain point. Users accidentally marking tasks complete is a real friction source. I support adding this as Epic 32 with P1 priority. The scope is intentionally minimal — `complete → todo` only. We don't need a general undo system or undo history. Keep it simple.

**Recommendation:** Approve as proposed. FR127 captures the requirement cleanly.

### Architect

From an architecture perspective, this is one of the cleanest changes we can make. The status transition model is table-driven — adding one entry to `validTransitions[StatusComplete]` enables the feature across all surfaces (TUI, CLI, MCP). The only subtlety is clearing `CompletedAt` when undoing, which is a 3-line addition to `UpdateStatus()`.

**Key design decisions:**
1. Only `complete → todo` is supported (not `complete → in-progress`). The task returns to the start of its lifecycle.
2. `CompletedAt` must be cleared to maintain data model invariants.
3. The completed.txt log is append-only — we should NOT remove entries from it. Instead, log an `undo_complete` event in session metrics so the behavioral data tells the full story.
4. No time limit on undo — if a user completed a task last week and realizes the mistake, they should still be able to undo it.

**Architecture impact:** Minimal. No new packages, no new interfaces, no new persistence formats.

### Dev

Implementation is straightforward:
1. Story 32.1: Status model change (transition table + CompletedAt clearing + tests) — trivial
2. Story 32.2: TUI integration (status menu already shows valid transitions, need confirmation UX + session metrics) — low effort
3. Story 32.3: Adapter awareness (ensure adapters handle `complete → todo` gracefully for bidirectional sync adapters) — low effort

The existing table-driven test infrastructure in `task_status_test.go` makes adding test coverage for the new transition very clean.

### QA

Test coverage requirements:
- Unit tests for `complete → todo` transition
- Unit tests for `CompletedAt` clearing
- Verify all other transitions remain unchanged (regression)
- TUI test for undo flow
- Contract test update ensuring adapters handle re-opened tasks
- Edge case: undo a task that was completed and then its file was modified externally

### UX Designer

The undo experience should be seamless:
- In the detail view of a completed task, the status menu should show "todo" as a valid option
- Consider showing a brief confirmation: "Task uncompleted. Returned to todo."
- No additional keybinding needed — the existing `S` (status) menu handles this naturally
- The doors view should immediately include the uncompleted task in the eligible pool
- No time limit on undo — accidental completion can happen any time

### SM (Scrum Master)

Sprint impact is minimal. This is 2-3 stories, low complexity, no dependencies on other active epics. Can be parallelized with Epic 23 and 24 remaining work.

---

## Consensus

All agents **unanimously approve** the Sprint Change Proposal for Epic 32: Undo Task Completion.

### Adopted Recommendations

1. **Simple reverse transition only:** `complete → todo` — no general undo history system
2. **No time limit:** Undo should work regardless of when the task was completed
3. **Append-only completed log:** Don't modify completed.txt; log undo events separately in session metrics
4. **Natural TUI integration:** Use existing status menu mechanism; add brief confirmation toast
5. **Session metrics event:** Log `undo_complete` events with task ID and original completion timestamp
6. **Adapter graceful handling:** Ensure adapters don't crash when a task transitions from complete back to todo; full "reopen" API support is deferred to individual adapter epics

### Priority

P1 — Validated user pain point, minimal effort, high quality-of-life impact.
