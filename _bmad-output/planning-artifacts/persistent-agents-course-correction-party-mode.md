# Party Mode Artifact: Persistent Agent Course Correction Validation

**Date:** 2026-03-08
**Topic:** Validate PM's sprint change proposal for persistent BMAD agent infrastructure
**Context:** Course correction based on PR #249 research (3 prior party mode rounds)
**Participants:** John (PM), Winston (Architect), Bob (SM), Murat (TEA), Amelia (Dev), Mary (Analyst)

---

## Adopted Approach

### 1. Two Persistent Agents — CONFIRMED

The party unanimously confirmed the research recommendation: add **project-watchdog** (PM role) and **arch-watchdog** (Architect role) as persistent agents.

**Rationale reinforced:**
- These two cover the highest-frequency governance gaps (PRD drift at ~90%, architecture divergence at ~80%)
- 5 total persistent agents (3 existing + 2 new) is within the recommended 6-7 agent limit
- Both have continuous monitoring surfaces that justify always-on status
- Independent monitoring loops with message bridges avoid single-point-of-failure

### 2. SM as 4-Hour Cron — CONFIRMED

Bob (SM) confirmed that the SM role's value is in periodic *pattern detection* ("3 stories blocked for 48h"), not continuous monitoring. 4-hourly cron via `/loop 4h /bmad-bmm-sprint-status` achieves ~70% of persistent value at ~10% of cost.

### 3. QA as Weekly Cron — CONFIRMED

Murat (TEA) confirmed that CI catches per-PR regressions; the weekly cron fills the *trend analysis* gap. Added requirement: define coverage baseline storage location and comparison mechanism in Story 37.2.

### 4. Epic 37 with 4 Stories — CONFIRMED (with refinements)

Bob proposed reducing to 3 stories by folding 37.3 into 37.1. After cross-talk, Winston and John agreed to keep 4 stories but clarified the scope distinction:
- **37.1:** Agent definition files in `agents/` directory (operational artifacts)
- **37.3:** Architecture documentation in `docs/architecture/` (design documentation — interaction diagram, authority boundaries, anti-patterns)

These are different artifacts in different locations serving different purposes.

### 5. PRD Scope Addition — CONFIRMED

Add "Autonomous Governance Infrastructure" to PRD product-scope.md under a new Phase 5+ section. All agents agreed this is correctly scoped as infrastructure, not product feature.

### 6. Agent Authority Boundaries — STRENGTHENED

Winston recommended embedding authority boundaries directly in each agent definition file, not just in documentation. This makes boundaries enforceable at the agent level, not just advisory.

**Adopted authority model:**

| Agent | Can Directly Edit | Must Spawn Worker | Must Escalate |
|-------|-------------------|-------------------|---------------|
| project-watchdog | Story files, ROADMAP.md | New story creation | Scope decisions, priority changes |
| arch-watchdog | Architecture docs | Code refactoring | Design decision overrides |
| SM cron | Sprint status doc | Nothing | Blocked items, risk alerts |
| QA cron | Coverage reports | Test improvements | Coverage policy changes |

### 7. Idempotency as Explicit Test Case — ADOPTED

Murat's recommendation: add to Story 37.1 acceptance criteria: "Verify that re-processing a previously-seen PR produces no duplicate messages or file edits." Makes the correlation ID anti-pattern testable.

### 8. Restart Recovery Behavior — ADOPTED

Murat's recommendation: agent definitions should document restart behavior — "On restart, re-scan last 10 merged PRs to catch any missed during downtime." This makes recovery automatic.

### 9. QA Cron Baseline Specification — ADOPTED

Amelia's recommendation: Story 37.2 must specify where the coverage baseline is stored and how comparison works. Not just "run `go test -cover`" — define the full mechanism.

### 10. Analyst Function Absorption — CONFIRMED

Mary confirmed that the analyst role generates research episodically, not continuously. Correctly absorbed into PM's monthly research sweep. Deferred evaluation: revisit if research backlog grows beyond monthly sweep capacity.

---

## Rejected Options

### Reduce to 3 Stories (fold 37.3 into 37.1)
- **Proposed by:** Bob (SM)
- **Why rejected:** Winston and John clarified that 37.1 (agent definitions in `agents/`) and 37.3 (architecture documentation in `docs/architecture/`) are distinct artifacts. Agent definitions are operational; architecture docs explain the design rationale, interaction patterns, and anti-patterns. Keeping them separate allows independent implementation and review.

### Add SM as Third Persistent Agent
- **Proposed by:** No one (evaluated in prior research rounds)
- **Why rejected:** Bob himself confirmed that SM's value is in periodic summarization, not continuous monitoring. Merge-queue and pr-shepherd already handle mechanical process health. The 4-hourly cron achieves ~70% of persistent value at minimal cost.

### Make Analyst Persistent
- **Proposed by:** No one
- **Why rejected:** Mary confirmed research docs accumulate slowly (days/weeks). Monthly sweep by PM is sufficient. Revisit only if backlog grows beyond monthly capacity.

### Defer All to Cron (No New Persistent Agents)
- **Proposed by:** No one (evaluated in prior research rounds)
- **Why rejected:** Cron jobs lack contextual awareness and message-reactivity. The PR merge cascade (merge → story update → ROADMAP update → PRD check → architect notification) requires persistent state across cascading checks.

### Single Hub Architecture (PM Only)
- **Proposed by:** No one (evaluated in prior research rounds)
- **Why rejected:** PM can't assess code-level architecture compliance. Architect needs an independent monitoring loop for code-to-doc divergence that requires domain expertise PM doesn't have.

---

## Refinements to Sprint Change Proposal

Based on party mode discussion, the following refinements should be incorporated:

1. **Story 37.1 ACs:** Add idempotency verification ("re-processing a seen PR produces no duplicates") and restart recovery behavior documentation
2. **Story 37.2 ACs:** Add coverage baseline storage location and comparison mechanism specification
3. **Story 37.3 scope:** Clarify this creates `docs/architecture/agent-communication.md` (or similar) — distinct from agent definition files
4. **Agent definitions:** Embed authority boundaries directly in each agent's definition file
5. **Architecture doc:** Document startup ordering (independent), failure recovery (idempotent + re-scan), and conflict resolution (separate worktrees, separate file domains)

---

## Consensus Summary

| Item | Decision | Votes |
|------|----------|-------|
| 2 persistent agents (PM + Architect) | Adopt | 6/6 |
| SM as 4h cron | Adopt | 6/6 |
| QA as weekly cron | Adopt | 6/6 |
| Epic 37 with 4 stories | Adopt (with refinements) | 5/6 (Bob initially proposed 3, accepted 4 after discussion) |
| PRD scope addition | Adopt | 6/6 |
| Authority boundaries in agent defs | Adopt (strengthened) | 6/6 |
| Idempotency test case | Adopt | 6/6 |
| Restart recovery spec | Adopt | 6/6 |
| QA baseline specification | Adopt | 6/6 |
