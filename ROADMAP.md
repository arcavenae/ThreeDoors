# ROADMAP — ThreeDoors

> Source of truth for merge-queue scope checks and worker prioritization.
> Synced periodically by BMAD PM agent from `docs/prd/epics-and-stories.md`.
> Last updated: 2026-03-07

## Priority Legend

- **P0** — Must ship. Blocks other work or users.
- **P1** — Should ship. High value, no blockers.
- **P2** — Nice to have. Lower urgency.

## Active Epics

### Epic 23: CLI Interface (P1) — 8/10 stories done

Non-TUI CLI for power users and LLM agents. Cobra-based, `--json` output.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 23.6 | Task Block, Unblock, and Status Commands | Not Started | P1 | 23.2 (done) |
| 23.7 | Task Edit, Delete, Note, and Search Commands | Not Started | P1 | 23.2 (done) |

### Epic 24: MCP/LLM Integration Server (P1) — 6/8 stories done

Expose task management to LLMs via Model Context Protocol.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 24.5 | TUI Proposal Review View | Not Started | P1 | 24.4 (done) |
| 24.8 | MCP Prompt Templates & Advanced Interaction Patterns | Not Started | P1 | 24.6 (done), 24.7 (done) |

## Completed Epics

| Epic | Title | Stories |
|------|-------|---------|
| 0 | Infrastructure & Process (Backfill) | 19/19 |
| 1 | Three Doors Technical Demo | 7/7 |
| 2 | Apple Notes Integration | 6/6 |
| 3 | Enhanced Interaction | 7/7 |
| 3.5 | Platform Readiness & Tech Debt | 8/8 |
| 4 | Learning & Intelligent Door Selection | 6/6 |
| 5 | macOS Distribution & Packaging | 1/1 |
| 6 | Data Layer & Enrichment | 2/2 |
| 7 | Plugin/Adapter SDK & Registry | 3/3 |
| 8 | Obsidian Integration | 4/4 |
| 9 | Testing Strategy & Quality Gates | 5/5 |
| 10 | First-Run Onboarding | 2/2 |
| 11 | Sync Observability & Offline-First | 3/3 |
| 12 | Calendar Awareness | 2/2 |
| 13 | Multi-Source Aggregation | 2/2 |
| 14 | LLM Task Decomposition | 3/3 |
| 15 | Psychology Research & Validation | 1/1 |
| 17 | Door Theme System | 6/6 |
| 19 | Jira Integration | 4/4 |
| 20 | Apple Reminders Integration | 4/4 |
| 21 | Sync Protocol Hardening | 4/4 |
| 22 | Self-Driving Development Pipeline | 8/8 |

## Icebox (Deferred Indefinitely)

| Epic | Title | Stories | Decision Date | Rationale |
|------|-------|---------|---------------|-----------|
| 16 | iPhone Mobile App (SwiftUI) | 0/7 | 2026-03-07 | No validated user demand; core user is CLI/TUI power user; MCP (Epic 24) may serve mobile-adjacent use cases via LLM agents; adds significant platform/build/distribution complexity |

**Re-entry gate for Epic 16:** Revisit if 5+ distinct user requests for mobile access, OR if MCP proves insufficient for on-the-go task management.

## Out of Scope

Work not listed above is out of scope. Merge-queue should reject PRs that introduce features or epics not on this roadmap without human approval.
