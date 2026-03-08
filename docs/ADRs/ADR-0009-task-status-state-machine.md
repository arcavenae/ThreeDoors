# ADR-0009: Task Status State Machine

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #8 (Story 1.2), #18 (Story 1.6), #91 (Story 3.5.2)

## Context

Tasks need lifecycle tracking beyond simple "done/not done". The system must support blocking, progress tracking, and review stages while keeping the model simple enough for a personal task manager.

## Decision

Implement a **five-state status machine** with validated transitions:

```
States: todo, blocked, in-progress, in-review, complete, deferred

Valid transitions:
  todo → in-progress → in-review → complete
  todo → blocked → in-progress
  in-progress → blocked → in-progress
  blocked → todo (unblock)
  Any state → complete (force complete)
  Any state → deferred (snooze, planned Epic 28)
```

## Rationale

- Five states cover the complete task lifecycle without over-complicating
- `blocked` state captures impediments with associated notes
- `in-review` supports workflows where tasks need validation
- Force-complete from any state handles edge cases pragmatically
- `deferred` (added later, Epic 28 planning) supports snooze/postpone workflows

## Consequences

### Positive
- Clear task lifecycle visible in TUI status indicators
- Invalid transitions rejected at the domain layer (not just UI)
- Session metrics can track time in each state
- Transition validation prevents accidental state corruption

### Negative
- Five states may be more than some users need (mitigated by smart defaults)
- Adding new states (like `deferred`) requires updating validation logic
- Status display in constrained spaces (door badges) must be concise
