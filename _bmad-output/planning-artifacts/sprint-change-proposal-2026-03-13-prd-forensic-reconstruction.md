# Sprint Change Proposal: PRD Forensic Reconstruction

**Date:** 2026-03-13
**Triggered by:** PRD process bypass investigation (`_bmad-output/planning-artifacts/prd-process-bypass-investigation.md`)
**Severity:** High — systemic process failure affecting 42 completed epics

## Problem Statement

Since 2026-03-08, the three PRD content documents (`requirements.md`, `product-scope.md`, `epic-details.md`) have not been updated despite 42 epics being completed. The index docs (`epic-list.md`, `epics-and-stories.md`) remained current via project-watchdog governance syncs, masking the rot in content docs. The PRD — supposedly the source of truth for what the product does — no longer describes the product that was built.

## Impact Analysis

**What's affected:**
- `epic-details.md`: Contains detailed breakdowns for Epics 1-21 only. 42 completed epics (22-65) have NO detail entry.
- `requirements.md`: Missing functional requirements for ~30 completed epics. Highest covered FR is FR151; many epics that reference FR numbers in `epic-list.md` never had those FRs written.
- `product-scope.md`: Stops at Phase 5+ (Autonomous Governance). Missing scope for ~20 feature areas delivered since.

**What breaks if we don't fix it:**
- PRD is useless as a product specification
- New team members cannot understand what was built by reading the PRD
- BMAD PM agent cannot make informed scope decisions
- The "delete code, rebuild from specs" principle (NFR-DX6) is impossible

## Proposed Approach

**Batch forensic reconstruction** of all three PRD content docs, using `epic-list.md` (current) and story files as source material. Organized into 4 thematic batches:

### Batch 1: Core Feature Epics (highest user-facing impact)
Epics 22 (Self-Driving Dev), 23 (CLI), 24 (MCP), 27 (Daily Planning)*, 28 (Snooze)*, 29 (Dependencies)*, 31 (Expand/Fork)*, 32 (Undo)*
*Already have requirements.md entries — need epic-details.md only

### Batch 2: Integration Epics
Epics 25 (Todoist), 26 (GitHub Issues), 30 (Linear), 43 (Connection Manager), 44 (Sources TUI), 45 (Sources CLI), 46 (OAuth), 47 (Sync Lifecycle), 63 (ClickUp)

### Batch 3: Visual/UX Epics
Epics 33 (Seasonal Themes), 35 (Door Visual), 36 (Door Feedback), 39 (Keybinding Display), 40 (Beautiful Stats), 41 (Charm Ecosystem), 48 (Door-Like Doors), 56 (Door Visual Redesign), 59 (Full-Terminal Layout)

### Batch 4: Dev Infrastructure & Governance
Epics 34 (SOUL.md)*, 37 (Persistent BMAD), 38 (Dual Homebrew), 42 (Security), 49 (Doctor), 50 (Bug Reporting), 51 (SLAES), 52 (Envoy Firewall), 53 (Remote Collab), 55 (CI Optimization), 57 (LLM CLI Services), 58 (Supervisor Handover), 60 (README), 61 (GitHub Pages), 62 (Retrospector Reliability), 65 (CLI Tests)
*Already has requirements.md entries — need epic-details.md only

## Rejected Alternatives

1. **Create stories and dispatch workers** — Rejected. This is a docs-only reconstruction task. Spawning multiple workers for PRD edits risks merge conflicts in the same files. Single-worker execution is safer.

2. **Option B from investigation: Incremental reconstruction** — Rejected for Tier 1 (completed epics). We have all the information needed now. Deferring reconstruction to "when the epic comes up for implementation" makes no sense for already-completed work.

3. **Only update epic-details.md** — Rejected. All three content docs are stale and all three need reconstruction for PRD integrity.

4. **Write detailed per-requirement entries for all 42 epics** — Partially rejected. For epics that already have FR entries referenced in epic-list.md but not written in requirements.md, we'll write them. For infrastructure/governance epics that don't naturally map to user-facing FRs, we'll use NFR entries instead.

## Scope of Work

This is a single PR with amendments to three files:
1. `docs/prd/requirements.md` — Add FR/NFR entries for ~30 epics missing requirements
2. `docs/prd/product-scope.md` — Add Phase 4-6+ scope sections for delivered features
3. `docs/prd/epic-details.md` — Add detailed breakdowns for all 42 epics (22-65)

## Effort Estimate

**Large** — estimated 2-3 hours of focused writing. Single PR, single worker.
