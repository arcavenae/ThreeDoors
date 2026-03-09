# ROADMAP — ThreeDoors

> Source of truth for merge-queue scope checks and worker prioritization.
> Synced periodically by BMAD PM agent from `docs/prd/epics-and-stories.md`.
> Last updated: 2026-03-09

## Priority Legend

- **P0** — Must ship. Blocks other work or users.
- **P1** — Should ship. High value, no blockers.
- **P2** — Nice to have. Lower urgency.

## Open Issues

### Issue #219: Door Selection UX Improvements

**Status:** Resolved. Epic 36 fully implemented (PRs #272, #276, #277 merged). Issue #219 closed.

Door selection lacks tactile feedback and intuitive interaction patterns.

## Infrastructure Backlog

### Story 0.24: Renovate + Dependabot Automated Dependency Management (P1)

**Status:** Done (PR #402).

Renovate manages Go module dependencies (grouping, auto-merge, OSV vulnerability scanning). Dependabot manages GitHub Actions version pinning. Merge-queue integration via `dependencies` label.

### Story 0.20: CI Churn Reduction (P1)

**Status:** Story created (PR #231). Research complete (PR #233). Implementation not started.

Branch protection & merge queue optimization to reduce cascading CI reruns.

### Story 0.31: CI/Security Hardening — Secrets, Supply Chain & Reproducibility (P1)

**Status:** Done (PR #270).

Pin golangci-lint version, replace third-party release action with gh CLI, add protected environment for release secrets.

### Story 0.28: Issue Tracker & Authority Configuration (P1)

**Status:** Done (PR #255).

Local issue tracker file (`docs/issue-tracker.md`) with authority tier configuration for the envoy agent. Operationalizes party mode research from PRs #227, #232.

### Story 0.29: Envoy Operations Guide (P1)

**Status:** Story created. Implementation not started. Depends on Story 0.28.

Operations documentation for the envoy agent: patrol workflows, cross-agent protocols, staleness thresholds, SOUL.md alignment patterns.

### Story 0.32: Help Display UX — Dedicated Help View (P1)

**Status:** Done (PR #309).

Replace broken `:help` flash message with dedicated scrollable help view. Content runs off-screen and disappears after 3 seconds. Fix: new `ViewHelp` mode, categorized two-column layout, `?` global keybinding.

### Story 0.34: Fix 'q' Key in Sub-Views — Go Back Instead of Quit (P1)

**Status:** Done (PR #361).

Story 36.3 (PR #276) universal quit handler causes sub-views (dashboard, health, synclog, etc.) to exit on 'q' instead of going back. Fix: scope 'q' quit to doors view only; sub-views treat 'q' as go-back (D-128).

## Active Epics

### Epic 25: Todoist Integration (P1) — 3/4 stories done

Todoist as task source via REST API v1. Thin HTTP client, read-only then bidirectional sync.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 25.1 | Todoist HTTP Client & Auth Configuration | Done (PR #308) | P1 | Epic 7 (done) |
| 25.2 | Read-Only Todoist Adapter with Field Mapping | Done (PR #321) | P1 | 25.1 |
| 25.3 | Bidirectional Sync & WAL Integration | In Review | P1 | 25.2 |
| 25.4 | Contract Tests & Integration Testing | Not Started | P1 | 25.2 |

### Epic 27: Daily Planning Mode (P1) — COMPLETE

Guided daily planning ritual for task review and focus selection. Transforms ThreeDoors from reactive task picker into proactive morning engagement tool.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 27.1 | Planning Data Model & Focus Tag | Done (PR #323) | P1 | Epic 1 (done) |
| 27.2 | Review Incomplete Tasks Flow | Done (PR #339) | P1 | 27.1 |
| 27.3 | Focus Selection Flow | Done (PR #352) | P1 | 27.1 |
| 27.4 | Energy Level Matching & Time-of-Day Inference | Done (PR #354) | P1 | 27.1 |
| 27.5 | Planning Session Metrics & CLI/TUI Commands | Done (PR #360) | P1 | 27.1-27.4 |

### Epic 28: Snooze/Defer as First-Class Action (P1) — 4/4 stories done — COMPLETE

Surfaces existing `StatusDeferred` as a first-class user action with date-based snooze and auto-return.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 28.1 | DeferUntil Field, Status Transitions, and Auto-Return Logic | Done (PR #310) | P1 | None |
| 28.2 | Snooze TUI View and Z-Key Binding | Done (PR #338) | P1 | 28.1 |
| 28.3 | Deferred List View and :deferred Command | Done (PR #358) | P1 | 28.1 |
| 28.4 | Session Metrics Logging for Snooze Events | Done (PR #355) | P1 | 28.1 |

### Epic 29: Task Dependencies & Blocked-Task Filtering (P1) — 4/4 stories done — COMPLETE

Native dependency graph support. Blocks tasks with unmet dependencies from door selection.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 29.1 | DependsOn Field, DependencyResolver, and YAML Persistence | Done (PR #307) | P1 | None |
| 29.2 | Door Selection Filter and Auto-Unblock on Completion | Done (PR #319) | P1 | 29.1 |
| 29.3 | TUI Blocked-By Indicator and Dependency Management | Done (PR #340) | P1 | 29.1 |
| 29.4 | Session Metrics Logging for Dependency Events | Done (PR #356) | P1 | 29.1 |

### Epic 32: Undo Task Completion (P1) — COMPLETE

Allow reversing accidental task completion via `complete → todo` transition. Validated pain point from Phase 1 gate.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 32.1 | Status Model — Complete-to-Todo Transition | Done (PR #306) | P1 | None |
| 32.2 | Session Metrics — Undo Complete Event Logging | Done (PR #322) | P1 | 32.1 |
| 32.3 | TUI & CLI Undo Experience | Done (PR #337) | P1 | 32.1, 32.2 |

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

### Epic 33: Seasonal Door Theme Variants (P2) — 3/4 stories done

Time-based seasonal theme variants that auto-switch based on current date. Extends Epic 17's theme infrastructure.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 33.1 | Seasonal Theme Metadata Model and Date-Range Resolver | Done (PR #403) | P2 | Epic 17 (done) |
| 33.2 | Four Seasonal Theme Implementations | Done (PR #409) | P2 | 33.1 |
| 33.3 | Auto-Switch Integration in DoorsView and Config | Done (PR #410) | P2 | 33.1 |
| 33.4 | Seasonal Theme Picker and `:seasonal` Command | Not Started | P2 | 33.2, 33.3 |

### Epic 38: Dual Homebrew Distribution (P1) — 6/6 stories done — COMPLETE

Parallel Homebrew distribution channels (stable + alpha) with signing parity, publishing controls, verification, and retention.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 38.1 | Alpha Homebrew Formula (`threedoors-a`) | Done (PR #273) | P1 | None |
| 38.2 | Alpha Publishing Toggle | Done (PR #287) | P1 | 38.1 |
| 38.3 | Stable Release Signing & Notarization | Done (PR #288) | P1 | None |
| 38.4 | Alpha Release Verification | Done (PR #295) | P2 | 38.1, 38.2 |
| 38.5 | Alpha Release Retention Cleanup | Done (PR #294) | P2 | None |
| 38.6 | Fix Alpha Homebrew Formula Template DSL | Done (PR #312) | P1 | 38.1 |

### Epic 39: Keybinding Display System (P1) — 12/13 stories done (1 cancelled) — COMPLETE

Toggleable keybinding bar and full overlay for TUI discoverability. Context-sensitive bottom bar shows key actions per view; `?` opens comprehensive reference overlay. Global command mode accessibility, command autocomplete, and inline key hints on interactive elements. Door key indicators unified under `h` toggle (D-137).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 39.1 | Keybinding Registry Model | Done (PR #305) | P1 | None |
| 39.2 | Concise Keybinding Bar Component | Done (PR #318) | P1 | 39.1 |
| 39.3 | Full Keybinding Overlay | Done (PR #320) | P1 | 39.1 |
| 39.4 | Toggle Behavior, Config Persistence, and MainModel Integration | Done (PR #346) | P1 | 39.2, 39.3 |
| 39.5 | View-Specific Keybinding Completeness and Polish | Done (PR #384) | P1 | 39.4 |
| 39.6 | Spacebar as Enter Alias in Doors View | Done (PR #303) | P1 | None |
| 39.7 | Global `:` Command Mode | Done (PR #365) | P1 | None |
| 39.8 | Command Autocomplete/Completion | Done (PR #381) | P1 | None |
| 39.9 | Inline Hint Rendering Infrastructure | Done (PR #374) | P1 | 39.1 |
| 39.10 | Door View Inline Hints | Done (PR #387) | P1 | 39.9 |
| 39.11 | Non-Door View Inline Hints | Done (PR #388) | P1 | 39.9 |
| 39.12 | Auto-Fade After N Sessions | Cancelled (D-140) | ~~P2~~ | 39.9, 39.10 |
| 39.13 | Unified Door Key Indicator Toggle | Done (PR #407) | P1 | 39.4, 39.9, 39.10 |

### Epic 40: Beautiful Stats Display (P1) — 10/10 stories done — COMPLETE

Transform the insights dashboard from plain text into a visually delightful, SOUL-aligned celebration of user activity using Lipgloss styled panels, gradient sparklines, bar charts, and fun facts.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 40.1 | Stats Dashboard Shell with Lipgloss Panels | Done (PR #343) | P1 | None |
| 40.2 | Gradient Sparkline with Color-Blind Safe Palette | Done (PR #366) | P1 | 40.1 |
| 40.3 | Fun Facts Engine | Done (PR #371) | P1 | 40.1 |
| 40.4 | Horizontal Bar Charts for Mood Correlation | Done (PR #362) | P1 | 40.1 |
| 40.5 | GitHub-Style Activity Heatmap | Done (PR #391) | P2 | 40.1, 40.8 |
| 40.6 | Surface Hidden Session Metrics | Done (PR #368) | P1 | 40.1 |
| 40.7 | Animated Counter Reveals | Done (PR #392) | P2 | 40.1 |
| 40.8 | Tab Navigation for Detail View | Done (PR #367) | P1 | 40.1 |
| 40.9 | Theme-Matched Stats Color Palettes | Done (PR #380) | P2 | 40.1, 40.2, Epic 17 |
| 40.10 | Milestone Celebrations | Done (PR #393) | P2 | 40.1 |

### Epic 41: Charm Ecosystem Adoption & TUI Polish (P2) — 6/6 stories done — COMPLETE

Systematically adopt underutilized charmbracelet ecosystem components to reduce custom code, improve UX consistency, and deliver on SOUL.md's "physical objects" promise.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 41.1 | Spinner Component for Async Provider Operations | Done (PR #372) | P2 | None |
| 41.2 | Lipgloss Layout Utilities Adoption | Done (PR #370) | P2 | None |
| 41.3 | Viewport Adoption for Help View | Done (PR #364) | P2 | None (sequence after Epic 39 overlay work) |
| 41.4 | Viewport Adoption for Synclog and Keybinding Overlay | Done (PR #379) | P2 | 41.3 |
| 41.5 | Harmonica Door Transition Spike | Done (PR #369) | P2 | None |
| 41.6 | Adaptive Color Profile Support | Done (PR #373) | P2 | None |

### Epic 42: Door-Like Doors — Visual Door Metaphor Enhancement (P2) — 0/4 stories done

Transform rectangular card/panel doors into visually convincing doors using side-mounted handles, hinge marks, threshold lines, crack-of-light selection feedback, and handle turn micro-animations. Based on 5-round party mode research with 7 agents.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 42.1 | Side-Mounted Handle + Hinge Marks | Not Started | P2 | Epic 35 (done), Epic 17 (done) |
| 42.2 | Continuous Threshold / Floor Line | Not Started | P2 | None |
| 42.3 | Crack of Light Effect on Selection | Not Started | P2 | 42.1 |
| 42.4 | Handle Turn Micro-Animation | Not Started | P2 | 42.1 |

**Dependency graph:** Stories 42.1 & 42.2 can parallelize. Stories 42.3 & 42.4 can parallelize after 42.1 completes.

### Epic 43: Connection Manager Infrastructure (P1) — 0/6 stories done

Connection lifecycle layer for data source integrations. State machine, credential storage (system keychain), config schema v3 (named connections), CRUD operations, sync event logging, and migration of existing adapters to the new pattern.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 43.1 | Connection State Machine and ConnectionManager Type | Not Started | P1 | None |
| 43.2 | Keyring Integration with Environment Variable Fallback | Not Started | P1 | None |
| 43.3 | Config Schema v3 Migration with Connections Support | Not Started | P1 | None |
| 43.4 | Connection CRUD Operations | Not Started | P1 | 43.1, 43.2, 43.3 |
| 43.5 | Sync Event Logging Infrastructure | Not Started | P1 | None |
| 43.6 | Migrate Existing Adapters to ConnectionManager Pattern | Not Started | P1 | 43.1-43.5 |

### Epic 44: Sources TUI (P1) — 0/7 stories done

TUI interfaces for data source management: setup wizard (`:connect`), sources dashboard (`:sources`), source detail view, sync log view, status bar health alerts, disconnection flow, and re-authentication flow. Uses `charmbracelet/huh` for wizard forms.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 44.1 | Setup Wizard with huh Forms | Not Started | P1 | Epic 43 |
| 44.2 | Sources Dashboard View | Not Started | P1 | Epic 43 |
| 44.3 | Source Detail View | Not Started | P1 | 44.2 |
| 44.4 | Sync Log View | Not Started | P1 | 43.5 |
| 44.5 | Status Bar Integration for Connection Health Alerts | Not Started | P1 | Epic 43 |
| 44.6 | Disconnection Flow with Task Preservation Options | Not Started | P1 | 44.2 |
| 44.7 | Re-Authentication Flow | Not Started | P1 | 44.3, Epic 46 |

### Epic 45: Sources CLI (P1) — 0/5 stories done

Non-interactive CLI commands for data source management: `threedoors connect`, `threedoors sources` (list/status/test/manage/log), and JSON output support for scripting and CI/automation.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 45.1 | `threedoors connect` Command (Non-Interactive) | Not Started | P1 | Epic 43 |
| 45.2 | `threedoors sources` List/Status/Test Commands | Not Started | P1 | Epic 43 |
| 45.3 | `threedoors sources` Management Commands | Not Started | P1 | Epic 43 |
| 45.4 | `threedoors sources log` Command | Not Started | P1 | 43.5 |
| 45.5 | JSON Output Support for All Sources Commands | Not Started | P1 | 45.1-45.4 |

### Epic 46: OAuth Device Code Flow (P2) — 0/4 stories done

Generic OAuth device code flow client for browser-based authentication. Provider-specific integrations for GitHub and Linear. Silent token refresh with explicit re-auth on expiry.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 46.1 | Generic Device Code Flow Client | Not Started | P2 | None |
| 46.2 | GitHub OAuth Integration | Not Started | P2 | 46.1 |
| 46.3 | Linear OAuth Integration | Not Started | P2 | 46.1 |
| 46.4 | Token Refresh Lifecycle | Not Started | P2 | 46.1 |

### Epic 47: Sync Lifecycle & Advanced Features (P2) — 0/4 stories done

Conflict resolution (last-writer-wins with field-level strategy), orphaned task handling, auto-detection of installed tools in setup wizard, and proactive connection health notifications.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 47.1 | Conflict Resolution Strategy with Logging | Not Started | P2 | Epic 43 |
| 47.2 | Orphaned Task Handling | Not Started | P2 | 47.1 |
| 47.3 | Auto-Detection of Existing Tools in Setup Wizard | Not Started | P2 | 44.1 |
| 47.4 | Proactive Connection Health Notifications | Not Started | P2 | 44.5 |

**Epic 43-47 dependency graph:** Epic 43 is the critical path — all other epics depend on it. Epics 44 (TUI) and 45 (CLI) can parallelize after Epic 43. Epic 46 (OAuth) is independent. Epic 47 (Advanced) depends on 43+44.

### Epic 36: Door Selection Interaction Feedback (P1) — 4/4 stories done — COMPLETE

Make door selection feel responsive and satisfying. Reopened for Story 36.4 (space/enter toggle).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 36.1 | Enhanced Door Selection Visual Feedback | Done (PR #277) | P1 | None |
| 36.2 | Deselect Toggle | Done (PR #272) | P1 | None |
| 36.3 | Universal Quit | Done (PR #276) | P1 | None |
| 36.4 | Space/Enter Toggle — Close Door by Pressing Same Key | Done (PR #405) | P1 | 36.2, 39.6 |

## Completed Epics

| Epic | Title | Stories |
|------|-------|---------|
| 0 | Infrastructure & Process (Backfill) | 11/13 |
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
| 14 | LLM Task Decomposition | 2/2 |
| 15 | Psychology Research & Validation | 2/2 |
| 17 | Door Theme System | 6/6 |
| 18 | Docker E2E & Headless TUI Testing | 5/5 |
| 19 | Jira Integration | 4/4 |
| 20 | Apple Reminders Integration | 4/4 |
| 21 | Sync Protocol Hardening | 4/4 |
| 22 | Self-Driving Development Pipeline | 8/8 |
| 23 | CLI Interface | 11/11 |
| 24 | MCP/LLM Integration Server | 8/8 |
| 25 | Todoist Integration | 4/4 |
| 26 | GitHub Issues Integration | 4/4 |
| 34 | SOUL.md + Custom Development Skills | 4/4 |
| 35 | Door Visual Appearance — Door-Like Proportions | 7/7 |
| 36 | Door Selection Interaction Feedback | 4/4 |
| 37 | Persistent BMAD Agent Infrastructure | 4/4 |
| 29 | Task Dependencies & Blocked-Task Filtering | 4/4 |
| 32 | Undo Task Completion | 3/3 |
| 38 | Dual Homebrew Distribution | 6/6 |
| 27 | Daily Planning Mode | 5/5 |
| 28 | Snooze/Defer as First-Class Action | 4/4 |
| 40 | Beautiful Stats Display | 10/10 |
| 41 | Charm Ecosystem Adoption & TUI Polish | 6/6 |

## Icebox (Deferred Indefinitely)

| Epic | Title | Stories | Decision Date | Rationale |
|------|-------|---------|---------------|-----------|
| 16 | iPhone Mobile App (SwiftUI) | 0/7 | 2026-03-07 | No validated user demand; core user is CLI/TUI power user; MCP (Epic 24) may serve mobile-adjacent use cases via LLM agents; adds significant platform/build/distribution complexity |

**Re-entry gate for Epic 16:** Revisit if 5+ distinct user requests for mobile access, OR if MCP proves insufficient for on-the-go task management.

## Out of Scope

Work not listed above is out of scope. Merge-queue should reject PRs that introduce features or epics not on this roadmap without human approval.
