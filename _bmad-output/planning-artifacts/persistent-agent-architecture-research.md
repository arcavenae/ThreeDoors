# Persistent BMAD Agent Architecture for Autonomous Project Governance

## 1. Executive Summary

This research evaluates which BMAD roles should become persistent multiclaude agents to make the ThreeDoors project more self-governing. Currently, 3 persistent agents handle operational concerns (merge-queue, pr-shepherd, envoy), but core governance functions — PRD alignment, architecture compliance, story sequencing, coverage monitoring — only happen when the supervisor manually dispatches them.

**Recommendation:** Add **two** persistent agents:
1. **project-watchdog** (PM role) — monitors planning-side drift
2. **arch-watchdog** (Architect role) — monitors implementation-side drift

Plus **two cron jobs** for lower-frequency monitoring:
- Sprint health (SM role) — every 4 hours
- Coverage audit (QA/TEA role) — weekly

Everything else stays ephemeral.

---

## 2. Current Agent Landscape

### Persistent Agents (Always-On)

| Agent | Role | Key Functions |
|-------|------|---------------|
| merge-queue | Merge PRs | CI validation, roadmap scope check, squash merge, emergency mode |
| pr-shepherd | Branch maintenance | Rebase branches, resolve conflicts, keep PRs up-to-date |
| envoy | Community liaison | Issue triage, reporter communication, cross-check merges vs issues |

### Ephemeral Agents (On-Demand)

| Agent | Role | Invocation Pattern |
|-------|------|--------------------|
| worker | Story implementation | Spawned per-story via `/implement-story` or `multiclaude work` |
| reviewer | Code review | Spawned per-PR via `multiclaude review` |
| release-manager | Release recovery | Spawned when release issues detected |

### BMAD Roles (Currently Interactive Only)

| Agent | Name | Key Capabilities |
|-------|------|-------------------|
| PM | John | PRD creation, requirements, stakeholder alignment |
| SM | Bob | Sprint planning, story preparation, agile ceremonies |
| Architect | Winston | System design, API design, scalable patterns |
| QA | Quinn | Test automation, coverage analysis |
| TEA | Murat | Risk-based testing, quality gates, CI governance |
| Dev | Amelia | Story execution, TDD |
| Analyst | Mary | Market research, competitive analysis, requirements |
| Tech Writer | Paige | Documentation, diagrams, standards |
| UX Designer | Sally | User research, interaction design |

---

## 3. Governance Gaps Analysis

| Gap | Impact | Frequency | Current Mitigation |
|-----|--------|-----------|-------------------|
| PRD drift after implementation | High — planning docs diverge from reality | Every merged PR | Manual supervisor audit |
| Architecture doc divergence | High — new patterns undocumented | Every code PR | None |
| Story status not updated | Medium — ROADMAP.md stale | Every story completion | Manual supervisor update |
| Story sequencing violations | Medium — dependencies broken | When multiple stories active | Supervisor monitors |
| Test coverage regression | Medium — quality degrades silently | Per-PR (caught by CI), trends (not caught) | CI per-PR only |
| Documentation staleness | Low — docs drift over weeks | Weekly/monthly | None |
| Research findings unactioned | Low — recommendations ignored | Monthly | None |
| Sprint health blindness | Medium — blocked work unnoticed | Continuous | Supervisor monitors manually |

---

## 4. Role-by-Role Persistence Evaluation

### PM Agent → **PERSISTENT** (project-watchdog)

**Monitoring Surface:**
- Recently merged PRs (`gh pr list --state merged --limit 10`)
- Story files in `docs/stories/` — status alignment with merged PRs
- ROADMAP.md — epic progress accuracy
- PRD — drift detection against implemented features
- Story sequencing — dependency violations

**Trigger Model:**
- Primary: Polling every 10-15 minutes for merged PRs
- Secondary: Messages from arch-watchdog about architecture changes
- Tertiary: Monthly sweep of `docs/research/` for unactioned recommendations

**Authority:**
- CAN: Update story file status, update ROADMAP.md progress, flag PRD drift
- CANNOT: Create new stories, modify code, make scope decisions
- ESCALATES: Scope changes, priority changes → supervisor

**Cost Justification:** Fixes the #1 governance gap (PRD/story drift). Every merged PR should trigger doc updates. Without persistent PM, this only happens when supervisor remembers. With 8 active epics and regular merges, this gap costs hours of manual reconciliation.

### Architect Agent → **PERSISTENT** (arch-watchdog)

**Monitoring Surface:**
- Code changes in `internal/` and `cmd/` against `docs/architecture/`
- New interfaces, packages, or patterns introduced in recent merges
- Architectural decision records — are they being followed?

**Trigger Model:**
- Primary: Polling every 20-30 minutes for merged code PRs
- Secondary: Messages from project-watchdog about PRD changes affecting architecture
- Tertiary: Periodic scan for undocumented patterns

**Authority:**
- CAN: Update architecture docs, open GitHub issues for divergence
- CANNOT: Refactor code, override design decisions
- ESCALATES: Design decision overrides, major architectural changes → supervisor

**Cost Justification:** Fixes the #2 governance gap (code-to-doc divergence). With 210+ PRs merged, some patterns have inevitably drifted from docs. Without persistent monitoring, this only gets caught during story planning when someone reads stale architecture docs.

### SM Agent → **CRON** (every 4 hours)

**Why not persistent:** Merge-queue and pr-shepherd already handle the mechanical aspects of process health (PR merging, branch rebasing). The SM's value is in periodic summarization and risk surfacing, not continuous monitoring. A 4-hourly cron achieves ~70% of persistent value at ~10% of the cost.

**Cron Tasks:**
- Query open PRs for staleness (>24h without activity)
- Check for blocked stories (dependencies unmet)
- Summarize worker activity (active, idle, stuck)
- Report to supervisor if risks detected

### QA/TEA Agent → **CRON** (weekly)

**Why not persistent:** CI runs tests on every PR. Per-PR quality is already gated. What's missing is trend analysis — is overall coverage declining? Are test files growing bloated? This is a weekly concern, not minute-by-minute.

**Cron Tasks:**
- Run `go test -cover ./...` and compare to baseline
- Check test-to-code ratio trends
- Flag packages with declining coverage
- Report to PM if regression detected

### Tech Writer → **EPHEMERAL**

**Why not persistent or cron:** Doc changes are story-driven and infrequent. A weekly staleness check could be folded into the PM's monitoring loop or run as an occasional cron job. Doesn't justify even periodic automation at current project scale.

### Analyst → **EPHEMERAL**

**Why not persistent:** Research docs accumulate slowly. The PM's monthly sweep of `docs/research/` covers the unactioned-recommendations concern. Analyst remains available on-demand for deep research tasks.

### UX Designer → **EPHEMERAL**

**Why not persistent:** Zero continuous monitoring surface for a CLI/TUI project. UX decisions are made during story planning, not discovered through monitoring.

### Dev → **EPHEMERAL**

**Why not persistent:** Dev agents are inherently task-scoped. They implement a story, create a PR, and complete. No monitoring function.

---

## 5. Agent Interaction Architecture

### Communication Model

Agents communicate via `multiclaude message send` — the established multiclaude primitive. No shared state files, no webhooks, no custom protocols.

### Interaction Diagram

```
                    ┌─────────────────┐
                    │   supervisor    │
                    │  (human/agent)  │
                    └────────┬────────┘
                             │ escalations
                    ┌────────┴────────┐
                    │                 │
         ┌─────────▼──────┐  ┌──────▼─────────┐
         │ project-watchdog│  │  arch-watchdog  │
         │  (PM - persist) │  │ (Arch - persist)│
         │                 │  │                 │
         │ Monitors:       │  │ Monitors:       │
         │ - Merged PRs    │◄─┤ - Code changes  │
         │ - Story status  │──►- Arch docs      │
         │ - ROADMAP.md    │  │ - New patterns   │
         │ - PRD alignment │  │ - Design records │
         └────────┬────────┘  └──────┬─────────┘
                  │                   │
         ┌────────┴───────────────────┴────────┐
         │           Message Bus               │
         │    (multiclaude message send)        │
         └────────┬───────────────────┬────────┘
                  │                   │
    ┌─────────────▼──┐          ┌─────▼──────────┐
    │  merge-queue    │          │  pr-shepherd   │
    │  (persist)      │          │  (persist)     │
    │  Merges PRs     │          │  Rebases PRs   │
    └─────────────────┘          └────────────────┘
                  │
    ┌─────────────▼──┐
    │    envoy        │
    │  (persist)      │
    │  Issue triage   │
    └─────────────────┘

    ┌─────────────────┐          ┌────────────────┐
    │  SM cron (4h)   │          │ QA cron (weekly)│
    │  Sprint health  │          │ Coverage audit  │
    └─────────────────┘          └────────────────┘
```

### Message Flow: PR Merge Cascade

```
1. merge-queue merges PR #NNN
   │
2. project-watchdog detects merge (polling)
   ├── Updates story X.Y status → "Done (PR #NNN)"
   ├── Updates ROADMAP.md epic progress
   ├── Checks PRD alignment
   │   └── If drift: messages arch-watchdog
   │         └── arch-watchdog reviews architecture docs
   │             └── If update needed: updates docs, messages project-watchdog
   │                 └── project-watchdog flags affected stories
   │
3. envoy cross-checks open issues
   └── If PR fixes open issue: comments and closes
```

### Anti-Patterns and Safeguards

1. **Circular notification prevention:** Each message includes the triggering PR number as correlation ID. Agents skip PRs they've already processed.

2. **Authority boundaries enforced:**
   - project-watchdog: edits docs/stories/, ROADMAP.md only
   - arch-watchdog: edits docs/architecture/ only
   - Neither modifies code — that requires worker spawning

3. **Rate limiting:** Polling intervals prevent API spam:
   - project-watchdog: 10-15 min
   - arch-watchdog: 20-30 min
   - Neither polls more frequently than once per 10 minutes

4. **Idempotency:** All updates are idempotent. Re-processing a PR produces the same result.

---

## 6. Practical Considerations

### Resource Budget (5 Persistent Agents)

| Agent | Poll Interval | Estimated API Calls/Hour | Notes |
|-------|---------------|--------------------------|-------|
| merge-queue | 5-10 min | 6-12 | Existing |
| pr-shepherd | 10-15 min | 4-6 | Existing |
| envoy | 15-20 min | 3-4 | Existing |
| project-watchdog | 10-15 min | 4-6 | New |
| arch-watchdog | 20-30 min | 2-3 | New |
| **Total** | | **19-31/hour** | |

Adding the cron jobs:
- SM (every 4h): ~6/day
- QA (weekly): ~1/week

### tmux Session Management

5 persistent agents = 5 tmux windows. multiclaude manages these natively. Each agent has its own worktree, preventing file conflicts.

### Scaling Limits

- **Recommended max:** 6-7 persistent agents before coordination overhead dominates
- **Current proposal:** 5 persistent agents — well within limits
- **If adding more:** Evaluate whether consolidation (merging roles into fewer agents) is better than adding agents

### Cron Implementation Options

1. **`/loop` skill:** `multiclaude` has a `/loop` command for recurring tasks
2. **System cron:** `crontab -e` for scheduled `multiclaude work` commands
3. **Supervisor-dispatched:** Supervisor manually runs periodic audits
   - Recommended: `/loop` for SM, system cron for QA (weekly is too long for `/loop`)

---

## 7. Implementation Roadmap

### Phase 1: MVP (Week 1)

1. Create `agents/project-watchdog.md` agent definition
2. Create `agents/arch-watchdog.md` agent definition
3. Spawn both via `multiclaude agents spawn`
4. Set up SM cron: `/loop 4h /bmad-bmm-sprint-status`
5. Monitor for 1 week, adjust polling intervals

### Phase 2: Tuning (Week 2-3)

1. Review agent logs for excessive polling or missed events
2. Adjust polling intervals based on actual merge frequency
3. Add QA weekly cron
4. Evaluate whether a third persistent agent is needed

### Phase 3: Evaluation (Week 4)

1. Measure governance gap closure (story status accuracy, ROADMAP freshness, architecture doc currency)
2. Decide whether to promote SM or QA to persistent
3. Document lessons learned

---

## 8. Draft Agent Definitions

See companion files:
- `agents/project-watchdog.md` — PM-based planning governance agent
- `agents/arch-watchdog.md` — Architect-based implementation governance agent

---

## 9. Party Mode Artifacts

Three rounds of multi-agent deliberation are documented at:
- `_bmad-output/planning-artifacts/persistent-agent-architecture-round1-role-evaluation.md`
- `_bmad-output/planning-artifacts/persistent-agent-architecture-round2-collaboration.md`
- `_bmad-output/planning-artifacts/persistent-agent-architecture-round3-mvp.md`

---

## 10. Key Decisions Summary

| Decision | Adopted | Rejected Alternatives |
|----------|---------|----------------------|
| Number of new persistent agents | 2 (PM + Architect) | 0, 1, 3, or all roles |
| Communication model | Message-driven (`multiclaude message send`) | Shared files, webhooks, dense mesh |
| PM persistence | Yes (project-watchdog) | Cron-only, ephemeral |
| Architect persistence | Yes (arch-watchdog) | Cron-only, ephemeral |
| SM persistence | No → 4-hourly cron | Persistent agent |
| QA persistence | No → weekly cron | Persistent agent |
| Tech Writer persistence | No → ephemeral | Persistent, cron |
| Analyst persistence | No → absorbed by PM | Persistent, cron |
| Hub topology | PM as primary hub, Architect as independent loop | Dense mesh, single hub |
| Circular notification prevention | Correlation ID per PR | Cooldown timers, blacklists |
