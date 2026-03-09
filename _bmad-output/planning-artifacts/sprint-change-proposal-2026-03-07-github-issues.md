# Sprint Change Proposal: GitHub Issues Integration

**Date:** 2026-03-07
**Change Type:** New Epic Proposal (Additive)
**Scope Classification:** Minor — Direct implementation by dev team

---

## Section 1: Issue Summary

**Problem Statement:** ThreeDoors currently lacks integration with GitHub Issues, the primary task tracker for its target audience (developers). The task source expansion research (task-source-expansion-research.md) identifies GitHub Issues as Tier 1 priority due to maximum developer audience overlap (100M+ developers), an excellent official Go SDK (google/go-github), and 2-3 day estimated implementation effort.

**Discovery Context:** Identified during systematic evaluation of task source expansion candidates. The Adapter SDK (Epic 7), Multi-Source Aggregation (Epic 13), and Sync Protocol Hardening (Epic 21) are all complete, meaning all infrastructure prerequisites are in place.

**Evidence:**
- GitHub Issues is the #2 Tier 1 integration (after Todoist, which is Epic 25)
- Official `google/go-github` SDK — actively maintained, excellent quality
- PAT authentication — simple, low friction for developers
- 5,000 requests/hour rate limit — generous
- `assignee` filter enables user-scoped issue queries across repositories

---

## Section 2: Impact Analysis

### Epic Impact
- **No existing epics affected** — this is a purely additive new epic (Epic 26)
- **Dependencies satisfied:** Epic 7 (Adapter SDK), Epic 13 (Multi-Source Aggregation), Epic 21 (Sync Protocol Hardening) are all COMPLETE
- **Pattern established:** Follows identical adapter pattern as Epic 19 (Jira), Epic 20 (Apple Reminders), and Epic 25 (Todoist)

### Story Impact
- No current or future stories require modification
- New stories follow the proven 4-story integration pattern:
  1. HTTP Client / SDK wrapper & auth config
  2. Read-only provider with field mapping
  3. Bidirectional sync & WAL integration
  4. Contract tests & integration testing

### Artifact Conflicts
- **PRD:** Needs new functional requirements (FR93-FR96) for GitHub Issues integration
- **Architecture:** No structural changes needed — adapter layer is designed for this
- **UI/UX:** No changes — existing source badge pattern applies (`[GH]` badge)
- **ROADMAP.md:** Needs new Epic 26 entry

### Technical Impact
- New package: `internal/adapters/github/`
- New dependency: `github.com/google/go-github/v68` (official SDK)
- Config additions: `github` provider settings in `~/.threedoors/config.yaml`
- No changes to existing code — pure addition

---

## Section 3: Recommended Approach

**Selected Path:** Direct Adjustment — Add new Epic 26 within existing plan

**Rationale:**
- All infrastructure is in place (adapter registry, contract tests, WAL, circuit breaker)
- Proven 4-story pattern from Jira/Reminders/Todoist integrations
- No impact on active work (Epics 23, 24, 25)
- 2-3 days estimated effort — low risk
- High value: developer audience overlap is maximum

**Effort Estimate:** Low (2-3 days)
**Risk Level:** Low — official SDK, proven pattern, no architectural changes
**Timeline Impact:** None — parallel track, no blocking dependencies

---

## Section 4: Detailed Change Proposals

### 4.1 PRD Requirements (New)

```
ADD to docs/prd/requirements.md — Phase 4+ Task Source Integration section:

FR93: The system shall integrate with GitHub Issues as a task source using the
official go-github SDK, reading issues with structured field mapping (title to
Text, body to Context, labels to Tags, state to Status), filtering by assignee
and repository scope.

FR94: The system shall support GitHub authentication via Personal Access Token
(PAT) configured in ~/.threedoors/config.yaml, with configurable repository
list (repos) and assignee filter (default: @me) for scoping which issues to
import.

FR95: The system shall map GitHub Issues fields to ThreeDoors task model:
open state maps to todo, closed state maps to complete, labels matching
"priority:*" convention map to Effort, milestone.due_on maps to due date,
and "in progress" label maps to in-progress status.

FR96: The system shall support bidirectional GitHub sync by closing issues via
the GitHub API when tasks are marked complete in ThreeDoors, with offline
queuing via WALProvider.
```

### 4.2 Epics & Stories (New)

```
ADD Epic 26: GitHub Issues Integration

4 stories following established adapter pattern:
- 26.1: GitHub SDK Client & Auth Configuration
- 26.2: Read-Only GitHub Provider with Field Mapping
- 26.3: Bidirectional Sync & WAL Integration
- 26.4: Contract Tests & Integration Testing
```

### 4.3 ROADMAP.md (New Entry)

```
ADD to Active Epics section:

### Epic 26: GitHub Issues Integration (P1) -- 0/4 stories done

GitHub Issues as task source for developer workflows. Official go-github SDK.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 26.1 | GitHub SDK Client & Auth Configuration | Not Started | P1 | Epic 7 (done) |
| 26.2 | Read-Only GitHub Provider | Not Started | P1 | 26.1 |
| 26.3 | Bidirectional Sync & WAL Integration | Not Started | P1 | 26.2 |
| 26.4 | Contract Tests & Integration Testing | Not Started | P1 | 26.2 |
```

---

## Section 5: Implementation Handoff

**Change Scope:** Minor — Direct implementation by development team

**Handoff Recipients:**
- **PM agent:** Update PRD with FR93-FR96
- **Architect agent:** No architecture changes needed (existing adapter pattern)
- **Dev workers:** Implement Epic 26 stories (4 stories, 2-3 days)

**Success Criteria:**
- All 4 stories implemented and merged
- Contract tests pass
- GitHub Issues appear as doors in TUI with `[GH]` badge
- Completing a GitHub issue in ThreeDoors closes it on GitHub
- Offline queuing works via WALProvider

**Recommended Implementation Order:**
1. Epic 25 (Todoist) first — already planned, validates the pattern for non-Atlassian/Apple APIs
2. Epic 26 (GitHub Issues) second — leverages learnings from Todoist implementation

---

## Approval

**Approved by:** BMAD workflow (autonomous pipeline)
**Scope:** Minor — no existing work affected, purely additive
**Next Steps:** Proceed with PRD updates, architecture documentation, and epic/story creation
