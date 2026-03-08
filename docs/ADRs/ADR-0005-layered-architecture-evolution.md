# ADR-0005: Layered Architecture Evolution

- **Status:** Accepted
- **Date:** 2025-11-07 (Phase 1), evolved through 2026-03-08
- **Decision Makers:** Project founder, architecture reviews
- **Related PRs:** #38, #39, #90, #91, #92
- **Related ADRs:** ADR-0004 (Monolithic CLI), ADR-0006 (TaskProvider Interface)

## Context

The initial Tech Demo (Phase 1) used a simple two-layer architecture: TUI and Tasks. As the project grew to support multiple task providers, sync engines, calendar awareness, LLM integration, and multi-source aggregation, the architecture needed to evolve without a full rewrite.

## Decision

Evolve from **two-layer** to **five-layer** architecture incrementally:

1. **TUI Layer** (`internal/tui`) — Bubbletea views, keyboard handling, rendering
2. **Core Domain** (`internal/tasks`) — Task management, door selection, metrics
3. **Adapter Layer** (`internal/tasks/providers`) — Pluggable `TaskProvider` implementations
4. **Sync Engine** (`internal/tasks/sync`) — Offline-first queue, conflict resolution
5. **Intelligence Layer** (`internal/tasks/intelligence`) — LLM decomposition, calendar awareness

Each layer was introduced only when its epic required it, not preemptively.

## Rationale

- YAGNI principle — don't build layers until needed
- Each epic naturally introduced one new layer
- Existing code didn't need wholesale refactoring
- Epic 3.5 (Platform Readiness & Tech Debt) performed the core domain extraction

## Key Milestones

| Phase | Layers | Introduced By |
|-------|--------|---------------|
| Phase 1 (Epic 1) | TUI + Tasks | Initial implementation |
| Phase 2 (Epic 3.5) | Core domain extraction | PR #90 (Story 3.5.1) |
| Phase 2 (Epic 7) | Adapter registry | PR #68 (Story 7.1) |
| Phase 3 (Epic 11) | Sync engine | PR #62 (Story 11.1) |
| Phase 3 (Epic 14) | Intelligence layer | PR #63 (Story 14.1) |

## Consequences

### Positive
- Each layer is independently testable
- New providers added without modifying core domain
- Sync engine operates independently of specific providers
- Clean dependency flow: TUI → Core → Adapters

### Negative
- Package boundaries required periodic refactoring (Epic 3.5)
- Some cross-layer concerns (e.g., error types) needed careful management
- Five layers in a single binary can feel over-structured for simpler operations
