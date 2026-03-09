# Sprint Change Proposal: Linear Integration

**Date:** 2026-03-07
**Proposed by:** BMAD Correct Course Workflow
**Change Type:** New Epic Proposal
**Scope Classification:** Moderate

---

## Section 1: Issue Summary

**Problem Statement:** ThreeDoors needs to expand its task source integrations to serve engineering teams that use Linear for project management. Linear was identified as Tier 2 priority in the task source expansion research (`task-source-expansion-research.md`), ranking 3rd after Todoist (Epic 25) and GitHub Issues (Epic 26).

**Context:** Linear has the best task model alignment of all evaluated services — 6 workflow states map cleanly to ThreeDoors statuses, priority (0-4) and estimates map to Effort, and due dates, labels, and Markdown descriptions all have direct mappings. The target audience (engineering teams) overlaps strongly with ThreeDoors users.

**Evidence:**
- Research document rates Linear's task model richness as "Excellent" — the highest rating
- GraphQL-only API with no Go SDK requires manual query construction (4-5 days effort)
- Auth complexity is low (personal API key or OAuth 2.0)
- Rate limits are generous (5,000 requests/hour)
- Growing adoption among engineering teams

---

## Section 2: Impact Analysis

### Epic Impact

- **No existing epics affected** — this is a net-new epic proposal (Epic 30)
- **No dependencies on incomplete work** — Epic 7 (Adapter SDK & Registry) is complete, providing the infrastructure for new adapters
- **No conflict with active epics** — Epics 23, 24, 25, 26, 28, 29 are independent
- **Priority:** P2 (nice to have) — should be implemented after Tier 1 integrations (Todoist Epic 25, GitHub Issues Epic 26)

### Story Impact

No existing stories require modification. New stories will follow the established pattern from Epics 25 and 26:
1. SDK/Client & Auth Configuration
2. Read-Only Provider with Field Mapping
3. Bidirectional Sync & WAL Integration
4. Contract Tests & Integration Testing

### Artifact Conflicts

| Artifact | Impact | Action Needed |
|----------|--------|---------------|
| PRD (requirements.md) | Addition | Add FR116-FR119 for Linear integration |
| Architecture (components.md) | Addition | Add LinearAdapter component |
| Epics & Stories | Addition | Create Epic 30 with 4 stories |
| ROADMAP.md | Addition | Add Epic 30 entry |
| Config schema | Minor | Add Linear provider settings |

### Technical Impact

- **New dependency:** Generic GraphQL client library (e.g., `github.com/hasura/go-graphql-client` or `github.com/shurcooL/graphql`)
- **New package:** `internal/adapters/linear/` — GraphQL client, provider, field mapping
- **Existing infrastructure reused:** TaskProvider interface, AdapterRegistry, WALProvider, contract test suite, MultiSourceAggregator
- **No breaking changes** to existing code

---

## Section 3: Recommended Approach

**Selected Path:** Direct Adjustment (Option 1) — Add new Epic 30 with stories within existing plan structure.

**Rationale:**
- The adapter SDK (Epic 7) and multi-source aggregation (Epic 13) are already complete — Linear is simply another adapter implementation
- The pattern is well-established from Todoist (Epic 25) and GitHub Issues (Epic 26) — follow the same 4-story breakdown
- No rollback or MVP review needed — this is additive scope with no impact on existing features
- P2 priority means it doesn't block or delay any P1 work

**Effort Estimate:** 4-5 days (Medium)
- GraphQL query construction and typed response parsing
- Cursor-based pagination handling
- Team discovery and selection
- Field mapping (status workflow states, priority, estimates)
- Contract tests with mocked GraphQL responses

**Risk Assessment:** Low
- Well-understood adapter pattern reduces implementation risk
- GraphQL-only API is the primary complexity (no Go SDK available)
- Team-scoped queries require configuration but are straightforward

**Timeline Impact:** None on active epics. Linear integration slots into the backlog after Tier 1 integrations.

---

## Section 4: Detailed Change Proposals

### PRD Changes (docs/prd/requirements.md)

**ADD** new section after GitHub Issues Integration:

```
## Linear Integration

FR116: The system shall integrate with Linear as a task source using the Linear GraphQL API,
reading issues with structured field mapping (title to Text, description to Context, labels to Tags,
state.type to Status with full workflow state mapping, priority to Effort, estimate to Effort,
dueDate to due date), filtered by team and assignee

FR117: The system shall support Linear authentication via personal API key configured in
~/.threedoors/config.yaml or LINEAR_API_KEY environment variable, with a configurable team
slug/ID for scoping which issues to import

FR118: The system shall map Linear workflow states to ThreeDoors statuses: triage/backlog -> todo,
unstarted -> todo, started -> in-progress, completed -> complete, cancelled -> archived;
and map Linear priority values (0=no priority, 1=urgent, 2=high, 3=medium, 4=low) to
ThreeDoors Effort with appropriate inversion

FR119: The system shall support bidirectional Linear sync by transitioning issues to "Done"
state via the Linear GraphQL API when tasks are marked complete in ThreeDoors, with offline
queuing via WALProvider
```

### Architecture Changes (docs/architecture/components.md)

**ADD** new LinearAdapter component in Adapter Layer section (after GitHubAdapter).

### Epics & Stories Changes

**ADD** Epic 30: Linear Integration with 4 stories:
- 30.1: Linear GraphQL Client & Auth Configuration
- 30.2: Read-Only Linear Provider with Field Mapping
- 30.3: Bidirectional Sync & WAL Integration
- 30.4: Contract Tests & Integration Testing

### ROADMAP.md Changes

**ADD** Epic 30 to Active Epics section with P2 priority.

---

## Section 5: Implementation Handoff

**Change Scope:** Moderate — Requires backlog addition and planning artifact updates.

**Handoff Plan:**

| Role | Responsibility |
|------|---------------|
| PM Agent | Update PRD with FR116-FR119 |
| Architect Agent | Define LinearAdapter component architecture |
| SM Agent | Create Epic 30 breakdown with stories |
| Development Workers | Implement stories (after Epic 25/26) |

**Success Criteria:**
- PRD updated with Linear integration requirements (FR116-FR119)
- Architecture document includes LinearAdapter component specification
- Epic 30 created with 4 stories, acceptance criteria, and dependency graph
- ROADMAP.md updated with Epic 30
- All planning artifacts committed and PR created

---

## Change Navigation Checklist Status

### Section 1: Understand the Trigger and Context
- [x] 1.1 — Trigger: Task source expansion research identified Linear as Tier 2 priority
- [x] 1.2 — Core problem: Need Linear integration for engineering team users
- [x] 1.3 — Evidence: Research document with API analysis and field mapping evaluation

### Section 2: Epic Impact Assessment
- [x] 2.1 — No current epic affected (net-new)
- [x] 2.2 — Add new Epic 30
- [x] 2.3 — No impact on remaining epics
- [x] 2.4 — No epics invalidated
- [x] 2.5 — P2 priority, after Tier 1 integrations

### Section 3: Artifact Conflict and Impact Analysis
- [x] 3.1 — PRD needs FR116-FR119
- [x] 3.2 — Architecture needs LinearAdapter component
- [N/A] 3.3 — No UI/UX impact (backend adapter)
- [x] 3.4 — ROADMAP.md needs Epic 30 entry

### Section 4: Path Forward Evaluation
- [x] 4.1 — Direct Adjustment: Viable (Low effort, Low risk)
- [N/A] 4.2 — Rollback: Not applicable
- [N/A] 4.3 — MVP Review: No MVP impact
- [x] 4.4 — Selected: Direct Adjustment

### Section 5: Sprint Change Proposal Components
- [x] 5.1 — Issue summary complete
- [x] 5.2 — Epic and artifact impact documented
- [x] 5.3 — Recommended path with rationale
- [x] 5.4 — PRD MVP not affected; action plan defined
- [x] 5.5 — Handoff plan established

### Section 6: Final Review and Handoff
- [x] 6.1 — All sections addressed
- [x] 6.2 — Proposal accuracy verified
- [x] 6.3 — Approved (automated BMAD pipeline)
- [x] 6.4 — Sprint status update deferred to epic creation step
- [x] 6.5 — Next steps: Party Mode -> PRD Edit -> Architecture -> Epics & Stories
