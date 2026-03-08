# ROADMAP — ThreeDoors

> Source of truth for merge-queue scope checks and worker prioritization.
> Synced periodically by BMAD PM agent from `docs/prd/epics-and-stories.md`.
> Last updated: 2026-03-08

## Priority Legend

- **P0** — Must ship. Blocks other work or users.
- **P1** — Should ship. High value, no blockers.
- **P2** — Nice to have. Lower urgency.

## Active Epics

### Epic 25: Todoist Integration (P1) — 0/4 stories done

Todoist as task source via REST API v1. Thin HTTP client, read-only then bidirectional sync.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 25.1 | Todoist HTTP Client & Auth Configuration | Not Started | P1 | Epic 7 (done) |
| 25.2 | Read-Only Todoist Adapter with Field Mapping | Not Started | P1 | 25.1 |
| 25.3 | Bidirectional Sync & WAL Integration | Not Started | P1 | 25.2 |
| 25.4 | Contract Tests & Integration Testing | Not Started | P1 | 25.2 |

### Epic 27: Daily Planning Mode (P1) — 0/5 stories done

Guided daily planning ritual for task review and focus selection. Transforms ThreeDoors from reactive task picker into proactive morning engagement tool.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 27.1 | Planning Data Model & Focus Tag | Not Started | P1 | Epic 1 (done) |
| 27.2 | Review Incomplete Tasks Flow | Not Started | P1 | 27.1 |
| 27.3 | Focus Selection Flow | Not Started | P1 | 27.1 |
| 27.4 | Energy Level Matching & Time-of-Day Inference | Not Started | P1 | 27.1 |
| 27.5 | Planning Session Metrics & CLI/TUI Commands | Not Started | P1 | 27.1-27.4 |

### Epic 28: Snooze/Defer as First-Class Action (P1) — 0/4 stories done

Surfaces existing `StatusDeferred` as a first-class user action with date-based snooze and auto-return.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 28.1 | DeferUntil Field, Status Transitions, and Auto-Return Logic | Not Started | P1 | None |
| 28.2 | Snooze TUI View and Z-Key Binding | Not Started | P1 | 28.1 |
| 28.3 | Deferred List View and :deferred Command | Not Started | P1 | 28.1 |
| 28.4 | Session Metrics Logging for Snooze Events | Not Started | P1 | 28.1 |

### Epic 29: Task Dependencies & Blocked-Task Filtering (P1) — 0/4 stories done

Native dependency graph support. Blocks tasks with unmet dependencies from door selection.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 29.1 | DependsOn Field, DependencyResolver, and YAML Persistence | Not Started | P1 | None |
| 29.2 | Door Selection Filter and Auto-Unblock on Completion | Not Started | P1 | 29.1 |
| 29.3 | TUI Blocked-By Indicator and Dependency Management | Not Started | P1 | 29.1 |
| 29.4 | Session Metrics Logging for Dependency Events | Not Started | P1 | 29.1 |

### Epic 32: Undo Task Completion (P1) — 0/3 stories done

Allow reversing accidental task completion via `complete → todo` transition. Validated pain point from Phase 1 gate.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 32.1 | Status Model — Complete-to-Todo Transition | Not Started | P1 | None |
| 32.2 | Session Metrics — Undo Complete Event Logging | Not Started | P1 | 32.1 |
| 32.3 | TUI & CLI Undo Experience | Not Started | P1 | 32.1, 32.2 |

### Epic 34: SOUL.md + Custom Development Skills (P1) — 0/3 stories done

Project philosophy document and custom Claude Code slash commands for AI agent alignment and developer workflow.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 34.1 | Create SOUL.md Project Philosophy Document | Not Started | P1 | None |
| 34.2 | Create Custom Claude Code Slash Commands | Not Started | P1 | 34.1 |
| 34.3 | Story Template Update and Integration Notes | Not Started | P1 | 34.2 |

### Epic 30: Linear Integration (P2) — 0/4 stories done

Linear as task source via GraphQL API. Best task model alignment of all evaluated services.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 30.1 | Linear GraphQL Client & Auth Configuration | Not Started | P2 | Epic 7 (done) |
| 30.2 | Read-Only Linear Provider with Field Mapping | Not Started | P2 | 30.1 |
| 30.3 | Bidirectional Sync & WAL Integration | Not Started | P2 | 30.2 |
| 30.4 | Contract Tests & Integration Testing | Not Started | P2 | 30.2 |

### Epic 31: Expand/Fork Key Implementations (P2) — 0/5 stories done

Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 31.1 | Task Model ParentID Extension | Not Started | P2 | None |
| 31.2 | Enhanced Expand — Sequential Subtask Creation | Not Started | P2 | 31.1 |
| 31.3 | Subtask List Rendering in Detail View | Not Started | P2 | 31.1, 31.2 |
| 31.4 | Enhanced Fork — Variant Creation with ForkTask Factory | Not Started | P2 | None |
| 31.5 | Design Decision H9 Status Update | Not Started | P2 | 31.1-31.4 |

### Epic 35: Door Visual Appearance — Door-Like Proportions (P1) — 0/7 stories done

Redesign all door themes to visually read as actual doors using portrait orientation, panel dividers, asymmetric handles, and threshold/floor lines. Addresses user feedback that themes look like cards, not doors.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 35.1 | Door Anatomy Model and Height-Aware Render Signature | Not Started | P1 | Epic 17 (done) |
| 35.2 | Classic Theme — Door-Like Proportions | Not Started | P1 | 35.1 |
| 35.3 | Modern Theme — Door-Like Proportions | Not Started | P1 | 35.1 |
| 35.4 | Sci-Fi Theme — Door-Like Proportions | Not Started | P1 | 35.1 |
| 35.5 | Shoji Theme — Door-Like Proportions | Not Started | P1 | 35.1 |
| 35.6 | Golden File Test Regeneration and Accessibility Validation | Not Started | P1 | 35.2-35.5 |
| 35.7 | Shadow/Depth Effect for 3D Door Appearance | Not Started | P2 | 35.2-35.5 |

### Epic 33: Seasonal Door Theme Variants (P2) — 0/4 stories done

Time-based seasonal theme variants that auto-switch based on current date. Extends Epic 17's theme infrastructure.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 33.1 | Seasonal Theme Metadata Model and Date-Range Resolver | Not Started | P2 | Epic 17 (done) |
| 33.2 | Four Seasonal Theme Implementations | Not Started | P2 | 33.1 |
| 33.3 | Auto-Switch Integration in DoorsView and Config | Not Started | P2 | 33.1 |
| 33.4 | Seasonal Theme Picker and `:seasonal` Command | Not Started | P2 | 33.2, 33.3 |

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
| 18 | Docker E2E & Headless TUI Testing | 5/5 |
| 19 | Jira Integration | 4/4 |
| 20 | Apple Reminders Integration | 4/4 |
| 21 | Sync Protocol Hardening | 4/4 |
| 22 | Self-Driving Development Pipeline | 8/8 |
| 23 | CLI Interface | 10/10 |
| 24 | MCP/LLM Integration Server | 8/8 |
| 26 | GitHub Issues Integration | 4/4 |

## Icebox (Deferred Indefinitely)

| Epic | Title | Stories | Decision Date | Rationale |
|------|-------|---------|---------------|-----------|
| 16 | iPhone Mobile App (SwiftUI) | 0/7 | 2026-03-07 | No validated user demand; core user is CLI/TUI power user; MCP (Epic 24) may serve mobile-adjacent use cases via LLM agents; adds significant platform/build/distribution complexity |

**Re-entry gate for Epic 16:** Revisit if 5+ distinct user requests for mobile access, OR if MCP proves insufficient for on-the-go task management.

## Out of Scope

Work not listed above is out of scope. Merge-queue should reject PRs that introduce features or epics not on this roadmap without human approval.
