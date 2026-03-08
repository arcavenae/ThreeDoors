# Sprint Change Proposal: Persistent BMAD Agent Infrastructure

**Date:** 2026-03-08
**Triggered By:** Research spike PR #249 — Persistent BMAD Agent Architecture
**Scope Classification:** Moderate — backlog expansion with new epic
**Mode:** Direct Adjustment

---

## 1. Issue Summary

### Problem Statement

ThreeDoors has grown to 210+ merged PRs across 30+ completed epics. The project currently operates with 3 persistent agents (merge-queue, pr-shepherd, envoy) handling operational concerns — but core governance functions are entirely manual:

- **PRD drift:** After every PR merge, planning docs (story status, ROADMAP.md, PRD alignment) only get updated when the supervisor manually dispatches audits
- **Architecture divergence:** New code patterns and interfaces introduced in PRs aren't reflected in architecture docs — nobody monitors for this
- **Story status staleness:** Completed stories often remain marked "In Progress" or "Not Started" until the next manual PM audit
- **Sprint health blindness:** No systematic monitoring for blocked stories, stale PRs, or idle workers

### Discovery Context

PR #249 conducted 3 rounds of party mode with all BMAD agents to evaluate which roles should become persistent. The research produced unanimous consensus on a focused approach: 2 new persistent agents + 2 cron jobs.

### Evidence

- 3 party mode artifacts documenting the evaluation, collaboration patterns, and MVP selection
- Governance gap analysis showing PRD drift as highest-frequency gap (every merged PR)
- Resource budget analysis showing 5 persistent agents (3 existing + 2 new) is well within limits
- Projected coverage: 90% of PRD drift, 80% of architecture divergence

---

## 2. Impact Analysis

### Epic Impact

**No existing epics are affected.** This is purely additive infrastructure work.

| Existing Entity | Impact |
|----------------|--------|
| Epics 25-35 (feature epics) | No impact — persistent agents will *support* these by keeping docs in sync |
| Epic 0 (Infrastructure) | Could house infrastructure stories, but a new epic is cleaner |
| Agents: merge-queue, pr-shepherd, envoy | No changes needed — new agents complement, don't overlap |

### Story Impact

No existing stories require modification.

### Artifact Conflicts

| Artifact | Change Needed | Severity |
|----------|--------------|----------|
| PRD product-scope.md | Add "Autonomous Governance" section under new phase | Medium |
| Architecture docs | Add agent communication architecture section | Medium |
| ROADMAP.md | Add new Epic 37 with stories | Low (after stories created) |
| Agent definitions (agents/) | Create project-watchdog.md, arch-watchdog.md | Core deliverable |
| Cron configuration | Set up SM (4h) and QA (weekly) cron jobs | Core deliverable |

### Technical Impact

- **No code changes** to the ThreeDoors application itself
- **Agent definitions:** New markdown files in `agents/` directory
- **Cron setup:** Uses existing multiclaude `/loop` skill and/or system cron
- **Resource usage:** +2 persistent agents, ~6-9 additional API calls/hour (well within budget)

---

## 3. Recommended Approach

### Selected Path: Direct Adjustment

Add a new **Epic 37: Persistent BMAD Agent Infrastructure** with 4 stories covering the full lifecycle from agent definition to tuning.

### Rationale

- **Low risk:** No code changes, only agent definitions and documentation
- **High value:** Automates the #1 and #2 governance gaps (PRD drift, architecture divergence)
- **Reversible:** Persistent agents can be stopped instantly if they cause issues
- **Well-researched:** 3 rounds of party mode with all BMAD agents produced unanimous consensus
- **Incremental:** Start with 2 persistent agents, add more only if clear gap emerges

### Effort Estimate

- **Overall:** Medium (4 stories, primarily documentation and configuration)
- **Timeline:** 1-2 days for implementation, 2 weeks for tuning/observation

### Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Agents create noise (excessive messages) | Low | Polling intervals are conservative (10-30 min); chatty agent anti-pattern explicitly addressed |
| Agents conflict on file edits | Very Low | Authority boundaries prevent overlap (PM edits planning docs, Architect edits architecture docs) |
| Resource overhead too high | Low | Budget analysis shows 19-31 API calls/hour total (5 agents); well within limits |
| Circular notification loops | Low | Correlation ID per PR prevents reprocessing |

---

## 4. Detailed Change Proposals

### 4.1 PRD: Add Autonomous Governance Phase

**File:** `docs/prd/product-scope.md`

**Section:** Add after Phase 5 (or as sub-section of Phase 5)

**NEW:**
```markdown
## Phase 5+: Autonomous Project Governance

**In Scope:**
- Persistent project-watchdog agent (PM role): merged PR monitoring, story status updates, ROADMAP.md sync, PRD drift detection, monthly research sweep
- Persistent arch-watchdog agent (Architect role): code-to-architecture-doc alignment, undocumented pattern detection, architectural drift flagging
- Sprint health cron (SM role): 4-hourly sprint status summary, blocked story detection, stale PR alerts
- Coverage audit cron (QA/TEA role): weekly test coverage trend analysis, regression flagging
- Agent communication via multiclaude message bus (no shared state files)
- Agent authority boundaries: each agent edits only its designated docs
- Idempotent, rate-limited polling with correlation IDs for cascade prevention

**Out of Scope for this Phase:**
- Tech Writer persistent agent (fold into PM's monitoring loop)
- Analyst persistent agent (fold into PM's monthly research sweep)
- UX Designer persistent agent (no continuous monitoring surface for CLI/TUI)
- Webhook-based event triggers (polling is sufficient at current scale)
```

### 4.2 Architecture: Agent Communication Architecture

**File:** New section in `docs/architecture/` or addendum to existing infrastructure docs

**Content:** Document the agent interaction architecture from the research:
- Message-driven communication via `multiclaude message send`
- Hub topology: PM as primary hub, Architect as independent loop
- PR merge cascade: merge-queue → project-watchdog → arch-watchdog → envoy
- Authority boundaries table
- Anti-patterns: circular notifications, authority creep, chatty agents
- Resource budget and scaling limits

### 4.3 New Epic 37: Persistent BMAD Agent Infrastructure

**Stories:**

| Story | Title | Priority | Depends On |
|-------|-------|----------|------------|
| 37.1 | Agent Definitions — project-watchdog and arch-watchdog | P1 | None |
| 37.2 | Cron Configuration — SM Sprint Health and QA Coverage Audit | P1 | None |
| 37.3 | Agent Communication Protocol and Authority Boundaries | P1 | 37.1 |
| 37.4 | Monitoring, Tuning, and Phase 1 Evaluation | P1 | 37.1, 36.2 |

---

## 5. Implementation Handoff

### Scope Classification: Moderate

This requires backlog expansion (new epic + stories) plus documentation updates to PRD and architecture docs.

### Handoff Recipients

| Role | Responsibility |
|------|---------------|
| PM (John) | Edit PRD to add autonomous governance scope; create story files |
| Architect (Winston) | Review/create agent communication architecture documentation |
| SM (Bob) | Validate story sequencing and add to sprint plan |
| Supervisor | Update ROADMAP.md with Epic 37 after stories are created |
| Workers | Implement stories 36.1-36.4 via `/implement-story` |

### Success Criteria

1. Two new persistent agents (project-watchdog, arch-watchdog) running and polling
2. SM cron running every 4 hours via `/loop`
3. QA cron running weekly
4. Story status updates happening automatically on PR merge
5. ROADMAP.md staying current without manual intervention
6. Architecture docs updated when code patterns change
7. No circular notification loops or authority boundary violations after 2 weeks

### Next Steps (Pipeline)

1. **Party Mode** — Validate this course correction proposal with all BMAD agents
2. **Architect Review** — Create agent communication architecture document
3. **PRD Edit** — Incorporate adopted changes from party mode
4. **Epic/Story Planning** — Create story files for Epic 37
5. **ROADMAP Update** — Add Epic 37 to ROADMAP.md

---

## Checklist Summary

| Section | Status |
|---------|--------|
| 1. Trigger & Context | [x] Complete |
| 2. Epic Impact | [x] Complete — 1 new epic needed |
| 3. Artifact Conflicts | [x] Complete — PRD, Architecture, Agent defs need updates |
| 4. Path Forward | [x] Complete — Direct Adjustment selected |
| 5. Proposal Components | [x] Complete |
| 6. Review & Handoff | [x] Complete |
