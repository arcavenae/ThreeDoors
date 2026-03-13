# Party Mode: Retrospector Agent Reliability Fixes

**Date:** 2026-03-12
**Participants:** PM (analysis), Architect (infrastructure review)
**Rounds:** 2 (lightweight — infrastructure changes, no UX or code architecture)

## Round 1: Problem Validation & Approach

### PM Assessment

The three problems are validated and interconnected:
- **Messaging bug** prevents retrospector from receiving work → it can't know what to analyze
- **BOARD.md access** prevents output → even if it analyzes correctly, findings are lost
- **Context exhaustion** limits operational window → patterns that span sessions are undetectable

Priority order: Messaging (unblocks operation) > BOARD.md access (unblocks output) > Context resilience (improves quality).

All three should ship together as they collectively enable retrospector's MVP functionality. Without all three, the agent remains non-functional despite Story 51.11's autonomy fix.

### Architect Assessment

**Messaging workaround is appropriate.** The multiclaude daemon's messaging system is external infrastructure. A file-based fallback inbox is a proven pattern (similar to how the JSONL findings log works — append-only, conflict-free). The identity verification probe on startup is good defensive programming.

**Recommendation queue is architecturally sound.** Separating detection (retrospector → queue file) from persistence (project-watchdog → BOARD.md) follows the existing authority model. Retrospector is a read-mostly agent; giving it write access to a critical governance file (BOARD.md) would violate the principle of least authority.

**Checkpointing is the right abstraction.** The checkpoint file should be a simple JSON document, not a complex database. Rolling window state (e.g., "last 10 CI failure rates") is the key thing to preserve — raw data points are already in the JSONL log.

## Round 2: Design Refinements

### Messaging Fallback — File vs. Environment Variable

**Adopted:** File-based inbox (`docs/operations/retrospector-inbox.jsonl`)
- Durable across restarts
- Supervisor can write to it via simple append
- Retrospector polls it alongside `multiclaude message list`
- Same JSONL pattern used throughout the project

**Rejected:** Environment variable passing at spawn time
- Not durable
- Can't receive messages after spawn
- Doesn't solve the ongoing communication problem

### Queue File Location

**Adopted:** `docs/operations/retrospector-recommendations.jsonl`
- Consistent with `docs/operations/retrospector-findings.jsonl` location
- JSONL format matches project patterns
- Clear naming distinguishes findings (observations) from recommendations (actionable proposals)

**Rejected:** `_bmad-output/` directory
- That's for planning artifacts, not operational data
- Would need git tracking which conflicts with `.gitignore` patterns

### Checkpoint Schema

**Adopted:** Minimal JSON with these fields:
```json
{
  "last_pr": 608,
  "last_timestamp": "2026-03-12T14:30:00Z",
  "mode_rotation_index": 2,
  "hours_since_restart": 3.5,
  "prs_since_restart": 12,
  "rolling_windows": {
    "ci_failure_rate_10pr": 0.3,
    "conflict_rate_10pr": 0.1,
    "rebase_avg_10pr": 1.5
  }
}
```

**Rejected:** SQLite or structured database
- Over-engineering for ~10 fields
- File-based JSON is sufficient and matches project patterns

## Decisions Summary

| ID | Decision | Rationale |
|---|---|---|
| Adopted | File-based fallback inbox for messaging | Durable, conflict-free, matches JSONL project patterns |
| Adopted | Recommendation queue file (not direct BOARD.md writes) | Separates detection from persistence; respects authority model |
| Adopted | JSON checkpoint file for state persistence | Minimal, sufficient, file-based |
| Rejected | Fix multiclaude daemon directly | External infrastructure, out of scope for this repo |
| Rejected | Give retrospector PR authority | Violates Phase 1 scope; risk of ungoverned changes |
| Rejected | Environment variables for messaging | Not durable, doesn't solve ongoing communication |
| Rejected | SQLite for checkpointing | Over-engineering |
