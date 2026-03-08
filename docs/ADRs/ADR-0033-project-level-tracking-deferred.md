# ADR-0033: Project-Level Work Tracking — Deferred

- **Status:** Deferred
- **Date:** 2026-03-08
- **Decision Makers:** Project founder
- **Related ADRs:** ADR-0032 (BMAD Files as Primary Tracker)

## Context

Cross-project analysis identified that work tracking at the repo level creates silos: overlapping numbering, no cross-project visibility, and no unified status view. The namespace-scoped issuing authority pattern (each repo/agent is an "issuing authority" with locally meaningful IDs, registered at project level for visibility) was proposed as a solution.

## Decision

**Defer this decision.** The problem is real but the solution is more complex than it appears:

1. **ThreeDoors is the only project without an `-orc` orchestrator** — the coordination model differs from other projects
2. **Team members need a live mutex** (like GH Issues or Jira) to coordinate — markdown files can't lock, notify, or prevent concurrent modification
3. **No live mutex exists yet** — and standing one up requires human decisions about tool selection, access control, and workflow
4. **Each human becomes their own issuing authority** — but then there's a project-level adoption step for formal inclusion across repos
5. **The authority model is underdefined** — who admits or rejects changes to the canonical tracker? This must be a human role, not an agent role

## What We Know

- Repo-scoped tracking creates silos (proven across ThreeDoors and peer projects)
- External systems (Jira, GH Issues) provide the locking/notification that markdown lacks
- The namespace-scoped issuing authority pattern is sound in theory (analogous to ThreeDoors' SourceRef canonical ID mapping)
- The coordination complexity scales with the number of repos and agents

## Re-entry Gate

Revisit this decision when:
- A second project under orc becomes actively developed with agents
- The live mutex question (Decision 18's "per-project tracking location") has a concrete answer
- The human authority model for admitting changes is defined
