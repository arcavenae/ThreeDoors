# Sprint Change Proposal: Todoist Integration

**Date:** 2026-03-07
**Proposed by:** BMAD Pipeline (witty-tiger worker)
**Change Scope:** Minor — Direct implementation by development team

---

## Section 1: Issue Summary

**Problem Statement:** ThreeDoors currently supports five task sources (text files, Apple Notes, Obsidian, Jira, Apple Reminders) but lacks integration with Todoist, the most popular standalone personal task manager with 50M+ users. The task source expansion research (`task-source-expansion-research.md`) identifies Todoist as the #1 recommended integration based on user base overlap, implementation effort, and field mapping quality.

**Context:** The adapter infrastructure is mature — Epic 7 (Plugin/Adapter SDK & Registry) established the `TaskProvider` interface, `AdapterRegistry`, and contract test suite. Epics 19 (Jira) and 20 (Apple Reminders) demonstrated the pattern for API-based and IPC-based adapters respectively. Epic 21 (Sync Protocol Hardening) added per-provider sync scheduling, circuit breakers, and canonical ID mapping. All infrastructure needed for a Todoist adapter is in place.

**Evidence:**
- Research document evaluates 10 integration candidates; Todoist ranks #1 (Tier 1)
- Todoist REST API v1 is clean, well-documented, with simple API key auth
- Field mapping covers Text, Context, Status (binary), Effort (via priority inversion), Due date, and Tags
- Estimated implementation effort: 2-3 days
- No Go SDK needed — thin HTTP client against REST API v1 is preferred over deprecated v2 libraries

---

## Section 2: Impact Analysis

### Epic Impact
- **No existing epics affected.** This is a purely additive new epic (proposed: Epic 25).
- **Dependencies satisfied:** Epic 7 (adapter SDK), Epic 13 (multi-source aggregation), Epic 21 (sync hardening) are all complete.
- **No blocking relationships** with remaining work (Epics 23, 24).

### Story Impact
- No current or future stories require changes.
- New stories to be created within the new epic.

### Artifact Conflicts

| Artifact | Impact | Change Needed |
|----------|--------|---------------|
| PRD (`docs/prd/product-scope.md`) | Phase 4 lists Todoist as "deferred to Phase 5+" | Move Todoist from Phase 5 to active scope |
| PRD (`docs/prd/requirements.md`) | No Todoist-specific requirements exist | Add FR89-FR92 for Todoist integration |
| Architecture | No conflicts — adapter pattern handles this | Add TodoistAdapter to component docs |
| ROADMAP.md | Todoist not listed | Add Epic 25 to Active Epics |
| Epics document | No Todoist epic exists | Add Epic 25 with 4 stories |

### Technical Impact
- New package: `internal/adapters/todoist/`
- New HTTP client for Todoist REST API v1
- Config extension: `todoist` provider in `~/.threedoors/config.yaml`
- Contract test compliance required
- No changes to existing code — purely additive

---

## Section 3: Recommended Approach

**Selected Path:** Direct Adjustment — Add new Epic 25 (Todoist Integration) with 4 stories.

**Rationale:**
- All adapter infrastructure is in place (Epic 7, 13, 21)
- Pattern is well-established from Jira (Epic 19) and Apple Reminders (Epic 20)
- Low implementation effort (2-3 days)
- Zero risk to existing functionality — purely additive
- High user value — Todoist has the largest personal task manager user base

**Effort Estimate:** Low (2-3 days)
**Risk Level:** Low
**Timeline Impact:** None — parallel to existing Epics 23/24

---

## Section 4: Detailed Change Proposals

### 4.1 PRD Product Scope Update

**File:** `docs/prd/product-scope.md`
**Section:** Phase 4 and Phase 5

**OLD (Phase 4 Out of Scope):**
```
- Todoist, Linear, GitHub Issues, ClickUp integrations (deferred to Phase 5+)
```

**NEW (Phase 4 Out of Scope):**
```
- Linear, GitHub Issues, ClickUp integrations (deferred to Phase 5+)
```

**OLD (Phase 5 In Scope):**
```
- Additional integrations (Todoist, Linear, GitHub Issues, ClickUp)
```

**NEW (Phase 5 In Scope):**
```
- Additional integrations (Linear, GitHub Issues, ClickUp)
```

**NEW addition to Phase 4 In Scope:**
```
- Todoist integration: read-only adapter (REST API v1, API token auth, priority-to-effort mapping), then bidirectional sync (complete tasks via API, WAL queuing)
```

**Rationale:** Todoist is the simplest and highest-value integration candidate. Moving it to the current active phase reflects the mature adapter infrastructure and low implementation effort.

### 4.2 PRD Requirements Addition

**File:** `docs/prd/requirements.md`
**Section:** New "Todoist Integration" subsection under Phase 4+

**NEW:**
```
**Todoist Integration:**

**FR89:** The system shall integrate with Todoist as a task source using the REST API v1,
reading tasks with structured field mapping (content to Text, description to Context,
labels to Tags, priority to Effort with scale inversion)

**FR90:** The system shall support Todoist authentication via personal API token configured
in `~/.threedoors/config.yaml`, with optional Todoist filter expressions for scoping
which tasks to import

**FR91:** The system shall map Todoist priority values (1=normal, 2=high, 3=urgent, 4=critical)
to ThreeDoors Effort values with appropriate scale inversion (Todoist 4 -> highest effort)

**FR92:** The system shall support bidirectional Todoist sync by completing tasks via the
REST API when tasks are marked complete in ThreeDoors, with offline queuing via WALProvider
```

### 4.3 New Epic: Epic 25 — Todoist Integration

**Stories:**

| Story | Title | Priority | Depends On |
|-------|-------|----------|------------|
| 25.1 | Todoist HTTP Client & Auth Configuration | P1 | Epic 7 (done) |
| 25.2 | Read-Only Todoist Adapter with Field Mapping | P1 | 25.1 |
| 25.3 | Bidirectional Sync & WAL Integration | P1 | 25.2, Epic 21 (done) |
| 25.4 | Contract Tests & Integration Testing | P1 | 25.2 |

### 4.4 ROADMAP.md Update

**Add to Active Epics:**
```
### Epic 25: Todoist Integration (P1) -- 0/4 stories done

Todoist as a task source via REST API v1. Read-only first, then bidirectional sync.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 25.1 | Todoist HTTP Client & Auth Configuration | Not Started | P1 | Epic 7 (done) |
| 25.2 | Read-Only Todoist Adapter with Field Mapping | Not Started | P1 | 25.1 |
| 25.3 | Bidirectional Sync & WAL Integration | Not Started | P1 | 25.2 |
| 25.4 | Contract Tests & Integration Testing | Not Started | P1 | 25.2 |
```

---

## Section 5: Implementation Handoff

**Change Scope Classification:** Minor — Direct implementation by development team.

**Handoff Plan:**
1. **PM Agent:** Update PRD (product-scope.md, requirements.md) with Todoist integration
2. **Architect Agent:** Update architecture docs with TodoistAdapter component
3. **Dev Team:** Create epic and story files, implement stories 25.1-25.4
4. **QA:** Verify contract test compliance, integration test coverage

**Success Criteria:**
- Todoist adapter passes all contract tests (`adapters.RunContractTests`)
- Tasks from Todoist appear in ThreeDoors TUI with `[TDT]` source badge
- Priority-to-Effort mapping correctly inverts Todoist's 1-4 scale
- Completing a task in ThreeDoors marks it complete in Todoist (when online)
- Offline changes are queued via WALProvider and replayed on reconnection
- Rate limiting (450 req/15min) is respected with appropriate backoff

---

## Approval

**Status:** Approved (autonomous BMAD pipeline — YOLO mode)
**Scope:** Minor — no existing work affected, purely additive
**Next Steps:** Proceed with PRD updates, architecture updates, epic/story creation
