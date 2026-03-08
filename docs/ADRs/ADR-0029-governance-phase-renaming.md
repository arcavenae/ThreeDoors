# ADR-0029: Rename Phase 5+ to Phase 4.5 (Active Governance)

- **Status:** Accepted
- **Date:** 2026-03-08
- **Decision Makers:** Project founder
- **Related PRs:** #265
- **Related ADRs:** ADR-0026 (Self-Driving Development Pipeline)

## Context

`product-scope.md` defines "Phase 5+: Autonomous Project Governance" as future work (12+ months out). In reality, the project is already building governance infrastructure: watchdog agents (project-watchdog, arch-watchdog), community envoy, decisions board, and sprint health monitoring. This work is currently invisible — scattered across Epic 0 backfill stories (0.20-0.29) rather than having its own recognized phase.

The "12+ months out" framing causes governance work to be treated as optional infrastructure rather than the active development focus it has become.

## Decision

Rename "Phase 5+: Autonomous Project Governance" to "Phase 4.5: Active Governance" in `product-scope.md`. Add formal functional requirements and non-functional requirements to the section. Keep governance stories in Epic 0 for now rather than creating a separate epic.

## Rationale

- The governance work is happening NOW, not in the future — the phase label should reflect reality
- "Phase 4.5" positions it between the current feature work (Phase 4) and future expansion (Phase 5), acknowledging it's an active parallel track
- Adding FRs/NFRs makes the scope explicit and lets agents/PM track progress
- Keeping stories in Epic 0 avoids a disruptive restructuring — can promote to its own epic later if the story count justifies it

## Rejected Alternatives

- **Promote to dedicated Epic 36:** Would make governance work more visible but requires migrating Epic 0 stories and renumbering. Premature given that the scope is still crystallizing.
- **Update Phase 5+ with FRs/NFRs only (keep the name):** Doesn't acknowledge that this work is active, not future. The "12+ months out" framing would remain misleading.
