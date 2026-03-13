# Sprint Change Proposal: PRD Coverage Gap Stories

**Date:** 2026-03-13
**Author:** bright-eagle (worker)
**Trigger:** PRD coverage gap analysis (PR #697, clever-tiger)

## Problem Statement

The PRD coverage gap analysis identified 3 features specified in the PRD that lack corresponding epics or stories:

1. **ClickUp Integration** — product-scope.md Phase 5 lists it, but no epic exists (GitHub Issues, Linear, Jira, Todoist all have epics)
2. **Cross-Computer Sync** — product-scope.md Phase 5 lists it, but no epic exists and no architecture research has been done
3. **DMG/pkg Installer (FR25)** — epic-details.md defines Story 5.3 with acceptance criteria, but the story file was never created

## Impact Analysis

**Severity:** Low — all three gaps are Phase 5/P2 items. No users are blocked. No active development depends on these.

**Why act now:** Formalizing these as stories ensures they appear in planning docs and aren't forgotten. The gap analysis (PR #697) created visibility; this proposal turns visibility into trackable work items.

## Proposed Approach

### Gap 1: ClickUp Integration (New Epic, ~4 stories)

Follow the **established integration adapter pattern** used by:
- Epic 19 (Jira) — 4 stories: HTTP Client → Read-Only Provider → Bidirectional Sync → Contract Tests
- Epic 25 (Todoist) — 4 stories: HTTP Client → Read-Only Provider → Bidirectional Sync → Contract Tests
- Epic 26 (GitHub Issues) — 4 stories: SDK Client → Read-Only Provider → Bidirectional Sync → Contract Tests
- Epic 30 (Linear) — 4 stories: GraphQL Client → Read-Only Provider → Bidirectional Sync → Contract Tests

ClickUp uses a REST API (v2). Stories:
1. ClickUp REST API Client & Auth Configuration
2. Read-Only ClickUp Provider with Field Mapping
3. Bidirectional Sync & WAL Integration
4. Contract Tests & Integration Testing

**Dependencies:** Epic 7 (Plugin/Adapter SDK — COMPLETE), Epic 43 (Connection Manager — COMPLETE)

### Gap 2: Cross-Computer Sync (New Epic, ~5-6 stories)

This is architecturally distinct from cross-provider sync (Epics 21, 43, 47). Needs a **research spike** before implementation stories can be fully specified.

Stories:
1. Research Spike — Sync Protocol Design (CRDTs vs LWW, transport, conflict resolution)
2. Device Identity & Registration
3. Sync Transport Layer (git-based, cloud storage, or peer-to-peer — per spike findings)
4. Conflict Resolution for Cross-Machine Edits
5. Offline Queue & Reconciliation
6. Cross-Computer Sync E2E Tests

Story 1 must complete before stories 2-6 can be fully specified. Stories 2-6 are provisional and will be refined after the research spike.

### Gap 3: DMG/pkg Installer (Story 5.3 under existing Epic 5)

Single story. Acceptance criteria already defined in `docs/prd/epic-details.md`:
1. CI generates a signed .pkg installer containing the signed+notarized binary
2. The .pkg installs `threedoors` to `/usr/local/bin/`
3. The .pkg is also notarized with Apple
4. The installer is uploaded to GitHub Releases alongside the raw binaries
5. Double-clicking the .pkg on macOS launches the standard macOS installer UI

**Note:** Epic 5 is currently marked COMPLETE (1/1 stories). Adding Story 5.3 will reopen it as 1/2.

## Rejected Alternatives

1. **Defer all three indefinitely** — Rejected because having PRD requirements without corresponding stories creates invisible gaps that erode planning doc reliability.
2. **Combine ClickUp with Linear epic** — Rejected because each integration is self-contained and the per-provider epic pattern is well-established.
3. **Create Cross-Computer Sync stories without research spike** — Rejected because the architecture is genuinely uncertain (transport, conflict resolution, identity management all need investigation).

## Priority & Timeline

All items are **P2 — Phase 5** (12+ months out). No urgency. This is planning formalization only.
