# Party Mode: PRD Coverage Gap Stories

**Date:** 2026-03-13
**Trigger:** Sprint change proposal for PRD coverage gaps
**Participants:** PM (story structure), Architect (technical feasibility), SM (story readiness)

## Context

Three PRD coverage gaps need story formalization. Two follow established patterns, one requires research.

## Discussion

### ClickUp Integration

**PM:** Straightforward pattern replication. Four prior integrations (Jira, Todoist, GitHub Issues, Linear) all use the same 4-story structure. ClickUp REST API v2 is well-documented, token-based auth. No PRD changes needed — product-scope.md already lists ClickUp.

**Architect:** Agree. The adapter pattern (Epic 7), Connection Manager (Epic 43), and sync infrastructure (Epics 21, 47) are all in place. ClickUp maps cleanly: Tasks → ThreeDoors tasks, Lists → source context, Statuses → task states. REST API v2 with API token auth — simpler than Linear's GraphQL. No architectural concerns.

**SM:** Stories should follow the exact template from Linear (Epic 30) since it's the most recent integration. All stories P2, Not Started.

### Cross-Computer Sync

**PM:** This is the one gap that genuinely needs architecture work. The PRD mentions it but provides no detail on implementation approach. Story 1 must be a research spike.

**Architect:** Key architectural questions for the spike:
1. **Transport:** Git-based sync (use existing `.threedoors/` repo), cloud intermediary (S3/GCS), or peer-to-peer?
2. **Conflict resolution:** CRDTs (operational transform) vs last-writer-wins with timestamps vs manual resolution?
3. **Identity:** How do devices identify themselves? UUID per install? User account system?
4. **Scope:** Full task state or just task list membership? What about session logs, analytics, config?

The existing WAL (Epic 21) and sync infrastructure help but are designed for provider-to-local sync, not device-to-device. This is a fundamentally different sync topology.

**SM:** Research spike story must have clear deliverables — an ADR documenting the chosen approach. Implementation stories 2-6 should be marked as provisional, refined after spike completes.

**PM:** Agreed. I'll mark stories 2-6 as "Provisional — pending research spike" in their descriptions.

### DMG/pkg Installer

**PM:** Single story. ACs already defined in epic-details.md Story 5.3. Just needs a story file created.

**Architect:** One concern: Epic 5 is marked COMPLETE. Adding a new story reopens it. This is cosmetically awkward but factually correct — FR25 was never actually satisfied by Story 5.1 (which only did Homebrew). The epics-and-stories.md incorrectly marks FR25 as "✅ (Story 5.1)" — that should be corrected.

**SM:** Agreed. The story file creation also triggers an update to Epic 5's status in planning docs.

## Consensus

1. **ClickUp Integration:** Proceed with 4-story epic following Linear pattern. P2/Phase 5.
2. **Cross-Computer Sync:** Proceed with research spike + 5 provisional stories. P2/Phase 5.
3. **DMG/pkg Installer:** Create Story 5.3 per epic-details.md. P2. Fix FR25 mapping in epics-and-stories.md.
4. **No architectural changes needed** for ClickUp or DMG/pkg. Cross-Computer Sync architecture deferred to research spike.

## Adopted Approach

- Create stories following established patterns
- Cross-Computer Sync stories 2-6 are provisional pending research spike
- All statuses set to Not Started
- Epic 5 reopened from COMPLETE to reflect Story 5.3 addition

## Rejected Options

- **Full party mode with all agents** — Overkill for pattern-following work. ClickUp is the 5th integration using the same template.
- **Architectural review for ClickUp** — No novel architecture. Connection Manager (Epic 43) handles all integration plumbing.
- **Deferring story creation** — Creates invisible planning gaps. Better to formalize now even if implementation is months away.
