# Scoped Labels Research — GitHub Labels for Human-LLM Collaborative Projects

**Date:** 2026-03-08
**Researcher:** witty-raccoon (worker agent)
**Purpose:** Design a comprehensive, scoped label taxonomy for ThreeDoors using '.' as field separator

---

## 1. Project Context

ThreeDoors is a public GitHub project run by a solo human directing autonomous AI agent teams:
- **Persistent agents:** supervisor, merge-queue, pr-shepherd, envoy, watchdogs
- **Ephemeral agents:** workers (spawned per story)
- **BMAD agents:** PM, SM, Architect, QA, Dev, TEA (invoked for deliberation)
- **Human:** repo owner/director — reviews PRs, makes scope/direction decisions

The project has 21 existing labels (see Section 3), 275+ merged PRs, 40 epics, and uses story-driven development with BMAD methodology.

## 2. Research Sources

### 2.1 Industry Best Practices

**CPython (Python core)** — One of the largest GitHub projects, uses prefixed labels:
- `type-bug`, `type-feature`, `type-crash`, `type-security`
- `OS-linux`, `OS-windows`, `OS-android`
- Version labels: `3.13`, `3.14`
- Process: `pending`, `stale`, `triaged`, `release-blocker`, `DO-NOT-MERGE`
- Key insight: **Topic labels double as expert notification** — subscribers get pinged

**Sean Trane's Logical Labels** — Prefix-based taxonomy:
- `effort/1` through `effort/13` (Fibonacci), `priority/now`, `priority/soon`
- `state/approved`, `state/blocked`, `state/inactive`, `state/pending`
- `type/bug`, `type/feature`, `type/chore`, `type/docs`
- `work/chaotic`, `work/complex`, `work/complicated`, `work/obvious` (Cynefin)
- Key insight: **Labels provide priority, effort, and decision-making state** — not just classification

**GitLab Scoped Labels** — Native scoped label support using `::` separator:
- Only one label per scope can be applied (e.g., `priority::high` replaces `priority::low`)
- GitHub lacks native scoped label support — must be enforced by convention or automation
- Key insight: **Scoped labels prevent contradictory states** (can't be both `status.blocked` and `status.in-review`)

**abdonrd/github-labels** — Reusable label packages via npm for standardization across repos

### 2.2 Agentic Engineering Patterns

**Andrej Karpathy — Software 3.0 / Agentic Engineering:**
- AI agents are "brilliant interns with perfect recall but no judgment"
- Human remains architect/reviewer/decision-maker
- "Context engineering" = filling the context window with the right information for the next step
- Specialized agents work on isolated portions, coordinate through shared state
- **Implication for labels:** Labels ARE shared state. They are the most lightweight coordination signal available on GitHub.

**Addy Osmani — Agentic Engineering:**
- Agents don't coordinate laterally — they receive scoped tasks from humans and report upward
- Test suites serve as state signaling mechanism (pass/fail = objective status)
- Version control and CI create explicit handoff points
- **Implication for labels:** Labels should signal to the orchestrator (supervisor/human), not enable peer-to-peer agent negotiation

**Anthropic — Building Effective Agents:**
- "Simple, composable patterns" over complex frameworks
- Orchestrator-workers: central LLM delegates and synthesizes
- Agents should "pause for human feedback at checkpoints or when encountering blockers"
- Invest in agent-computer interfaces (ACI) like HCI design
- **Implication for labels:** Labels are an ACI. They should be as carefully designed as API surfaces.

**Anthropic — Effective Harnesses for Long-Running Agents:**
- Progress files document what each session accomplished
- Feature status tracking: explicit pass/fail states per requirement
- Git commits with descriptive messages as state change signals
- Incremental progress: one feature per session
- **Implication for labels:** Labels complement progress files. They provide the glanceable overview; files provide depth.

**Anthropic — Multi-Agent Research System:**
- Orchestrator-worker pattern with lead agent coordinating subagents
- Each subagent needs: objective, output format, tool/source guidance, task boundaries
- Memory system persists context across long conversations
- **Implication for labels:** Label scopes map to agent concerns. Each label should have a clear consumer.

### 2.3 Academic Research

**ACM DAI 2025 — "The Manager Agent as Unifying Research Challenge":**
- Manager agent must modify workflow in real-time: revise task graph, adjust roles, reassign
- Priority and state metadata essential for orchestration
- **Implication for labels:** Labels are the task graph metadata layer for GitHub-native orchestration

### 2.4 ThreeDoors-Specific Context

**Envoy Rules of Behavior (party mode, 5 rounds):**
- Envoy has autonomous labeling authority for: bug, enhancement, question, documentation, priority
- Envoy MUST escalate: closing issues, scope decisions, priority overrides
- Three-tier authority model: owner > contributor > community
- Three-category direction alignment: aligned / misaligned / gray-area
- Staleness thresholds: 14d (no envoy update), 30d (no linked story), 21d (PR open)

**Existing Agent Label Consumers:**
| Agent | Reads Labels | Sets Labels | Uses Labels For |
|-------|-------------|-------------|-----------------|
| Envoy | All | triage.*, priority.*, scope.*, type.* | Issue triage workflow, reporter communication |
| Merge-queue | scope.*, priority.* | — | Scope gating (reject out-of-scope PRs) |
| PR-shepherd | priority.* | — | Rebase priority ordering |
| Supervisor | All | Any (override authority) | Dispatch decisions, status overview |
| Workers | — | — | Not label-aware (story-driven) |
| Human | All | Any | Direction, review, override |

## 3. Current Label Inventory (21 labels)

| Current Label | Separator | Scoped? |
|--------------|-----------|---------|
| bug | — | No |
| documentation | — | No |
| duplicate | — | No |
| enhancement | — | No |
| question | — | No |
| multiclaude | — | No |
| needs-human-input | — | No |
| triage:new | `:` | Yes |
| triage:complete | `:` | Yes |
| triage:in-progress | `:` | Yes |
| triage:needs-info | `:` | Yes |
| priority:P0 | `:` | Yes |
| priority:P1 | `:` | Yes |
| priority:P2 | `:` | Yes |
| priority:P3 | `:` | Yes |
| scope:in-scope | `:` | Yes |
| scope:out-of-scope | `:` | Yes |
| scope:needs-decision | `:` | Yes |
| infrastructure | — | No |
| ux | — | No |
| fast-track | — | No |

### Problems with Current Labels

1. **Inconsistent separators:** Mix of `:` (triage, priority, scope) and no separator (bug, enhancement)
2. **No agent ownership signal:** No way to know which agent is responsible for an issue/PR
3. **No epic/story tracking:** Issues can't be linked to epics via labels
4. **Missing PR workflow labels:** No `status.in-review`, `status.changes-requested`, `status.approved`
5. **No blocked/dependency signal:** Can't mark issues blocked on other work
6. **Redundant with GitHub defaults:** `bug`, `documentation`, `enhancement`, `question`, `duplicate` overlap with GitHub's default labels
7. **`multiclaude` label is vague:** Doesn't distinguish which agent or what kind of management

## 4. Design Principles for Scoped Labels

### 4.1 Separator Choice: `.` (dot)

**Rationale:**
- `:` is already used but creates ambiguity with GitHub's default label rendering
- `::` is GitLab convention, looks unusual on GitHub
- `/` conflicts with path separators in documentation references
- `.` is clean, readable, universally understood as a namespace separator
- Consistent with DNS, Java packages, Python modules

### 4.2 Core Design Rules

1. **Every label must have at least one concrete consumer** — if no agent or human reads it, delete it
2. **Scoped labels are mutually exclusive within scope** — an issue can't be `priority.p0` AND `priority.p2`
3. **Labels signal state, not describe content** — prefer `status.blocked` over `blocked-by-dependency`
4. **Labels should trigger actions** — each label change should prompt a specific agent behavior
5. **Compatible with GitHub search** — `label:priority.p0 label:status.blocked` should be useful queries
6. **No label bloat** — if two labels always co-occur, merge them; if a label is never queried, delete it

### 4.3 Audience Matrix

| Audience | Primary Concern | Label Interaction |
|----------|----------------|-------------------|
| Human director | Priority, status overview, scope decisions | Reads all, sets scope.*, overrides any |
| Envoy | Triage state, reporter communication | Sets triage.*, type.*, priority.* |
| Merge-queue | Scope compliance, merge readiness | Reads scope.*, status.* |
| PR-shepherd | Rebase priority, staleness | Reads priority.*, status.* |
| Supervisor | Everything — dispatch, monitoring | Reads all, sets agent.*, status.* |
| Public contributors | What to work on, what's blocked | Reads type.*, status.*, contrib.* |

## 5. Proposed Scoped Label Taxonomy

### 5.1 Type Labels (issue classification)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `type.bug` | Something isn't working | Envoy, contributors | `bug` |
| `type.feature` | New feature or enhancement | Envoy, contributors | `enhancement` |
| `type.docs` | Documentation improvement | Envoy, contributors | `documentation` |
| `type.question` | Question about usage | Envoy | `question` |
| `type.infra` | CI/CD, tooling, build system | Envoy, merge-queue | `infrastructure` |
| `type.ux` | User experience improvement | Envoy | `ux` |

**Color family:** Blue tones (#0075ca variants)

### 5.2 Priority Labels (urgency signaling)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `priority.p0` | Blocks users — immediate response | All agents | `priority:P0` |
| `priority.p1` | Important — this sprint | Supervisor, PR-shepherd | `priority:P1` |
| `priority.p2` | Backlog — when capacity allows | Supervisor | `priority:P2` |
| `priority.p3` | Someday/maybe | — | `priority:P3` |

**Color family:** Red gradient (P0=#B60205, P1=#D93F0B, P2=#FBCA04, P3=#D4C5F9)

### 5.3 Triage Labels (envoy workflow state)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `triage.new` | Issue detected, acknowledgment posted | Envoy | `triage:new` |
| `triage.in-progress` | Envoy actively triaging | Envoy | `triage:in-progress` |
| `triage.needs-info` | Waiting on reporter for clarification | Envoy | `triage:needs-info` |
| `triage.complete` | Triage finished, story may exist | Envoy, supervisor | `triage:complete` |

**Color family:** Cyan gradient (#C2E0F4 to #1D76B5)

### 5.4 Scope Labels (roadmap alignment)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `scope.in-scope` | Fits current roadmap | Merge-queue, envoy | `scope:in-scope` |
| `scope.out-of-scope` | Outside current project direction | Envoy, merge-queue | `scope:out-of-scope` |
| `scope.needs-decision` | Requires human evaluation | Supervisor, human | `scope:needs-decision` |

**Color family:** Green/Red/Yellow (#0E8A16, #E11D48, #FEF2C0)

### 5.5 Status Labels (workflow state for PRs and issues)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `status.blocked` | Blocked on dependency or decision | PR-shepherd, supervisor | — (new) |
| `status.in-review` | PR awaiting review | Human, merge-queue | — (new) |
| `status.changes-requested` | Review feedback needs addressing | Workers, PR-shepherd | — (new) |
| `status.approved` | Approved, ready to merge | Merge-queue | — (new) |
| `status.stale` | No activity past staleness threshold | Envoy, PR-shepherd | — (new) |
| `status.needs-human` | Blocked on human action | Supervisor, human | `needs-human-input` |

**Color family:** Purple/Orange tones

### 5.6 Agent Labels (ownership/routing)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `agent.envoy` | Envoy is responsible | Envoy | — (new) |
| `agent.merge-queue` | Merge-queue is responsible | Merge-queue | — (new) |
| `agent.pr-shepherd` | PR-shepherd is responsible | PR-shepherd | — (new) |
| `agent.worker` | Assigned to a worker | Supervisor | — (new) |
| `agent.supervisor` | Supervisor is handling | Supervisor | `multiclaude` |

**Color family:** Green tones (#0E8A16 variants)

**Key insight from agentic engineering research:** Agent labels serve the same purpose as team assignment in project management — they answer "who is responsible?" at a glance. This is the primary coordination signal in our orchestrator-workers architecture.

### 5.7 Contributor Labels (community engagement)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `contrib.good-first-issue` | Good for newcomers | Contributors | — (new) |
| `contrib.help-wanted` | Community help welcome | Contributors | — (new) |

**Color family:** Friendly green (#7057ff)

### 5.8 Resolution Labels (outcome tracking)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `resolution.duplicate` | Duplicate of another issue | Envoy | `duplicate` |
| `resolution.wontfix` | Not aligned with project direction | Envoy, supervisor | — (new) |
| `resolution.fixed` | Fixed by merged PR | Envoy | — (new) |

**Color family:** Gray tones (#cfd3d7)

### 5.9 Process Labels (workflow shortcuts)

| Label | Description | Consumer | Replaces |
|-------|------------|----------|----------|
| `process.fast-track` | Trivial fix, skip full triage | Envoy | `fast-track` |
| `process.party-mode` | Requires party mode deliberation | Supervisor | — (new) |

**Color family:** Light blue (#C5DEF5)

## 6. Labels Considered and Rejected

| Proposed Label | Reason for Rejection |
|----------------|---------------------|
| `epic.N` (per-epic labels) | GitHub milestones already serve this purpose; 40+ epic labels would create bloat; story files are the authoritative epic-to-issue link |
| `sprint.current` / `sprint.next` | ThreeDoors doesn't use fixed sprints; work is continuous; stories track scheduling |
| `effort.fibonacci` | Effort estimation is in story files, not issues; adding it to labels duplicates data |
| `os.macos` / `os.linux` | Single-platform app (macOS); not useful until cross-platform support |
| `version.X.Y` | No versioned backporting; single release channel (plus alpha) |
| `agent.pm` / `agent.architect` | BMAD agents are invoked, not persistent — they don't own issues long-term |
| `team.bmad` | Redundant with `process.party-mode`; all agents are BMAD |
| `needs-tests` | Testing is a universal requirement (per CLAUDE.md), not a label-worthy exception |
| `wip` | GitHub draft PRs serve this purpose |
| `breaking` | No public API; breaking changes are internal |

## 7. Migration Plan

### Phase 1: Rename existing scoped labels (`:` to `.`)
```
triage:new → triage.new
triage:in-progress → triage.in-progress
triage:needs-info → triage.needs-info
triage:complete → triage.complete
priority:P0 → priority.p0
priority:P1 → priority.p1
priority:P2 → priority.p2
priority:P3 → priority.p3
scope:in-scope → scope.in-scope
scope:out-of-scope → scope.out-of-scope
scope:needs-decision → scope.needs-decision
```

### Phase 2: Rename unscoped labels to scoped equivalents
```
bug → type.bug
enhancement → type.feature
documentation → type.docs
question → type.question
infrastructure → type.infra
ux → type.ux
duplicate → resolution.duplicate
needs-human-input → status.needs-human
multiclaude → agent.supervisor
fast-track → process.fast-track
```

### Phase 3: Create new labels
```
status.blocked
status.in-review
status.changes-requested
status.approved
status.stale
agent.envoy
agent.merge-queue
agent.pr-shepherd
agent.worker
contrib.good-first-issue
contrib.help-wanted
resolution.wontfix
resolution.fixed
process.party-mode
```

### Phase 4: Verify and clean up
- Confirm no issues/PRs lost labels during migration
- Delete old labels after migration
- Update envoy agent definition with new label names
- Update merge-queue scope checking to use new label names

## 8. Total Label Count

| Scope | Count |
|-------|-------|
| type.* | 6 |
| priority.* | 4 |
| triage.* | 4 |
| scope.* | 3 |
| status.* | 6 |
| agent.* | 5 |
| contrib.* | 2 |
| resolution.* | 3 |
| process.* | 2 |
| **Total** | **35** |

Current: 21 labels. Proposed: 35 labels (+14 net new, all serving identified consumers).

## 9. Color Scheme

Each scope gets a distinct color family for visual grouping in GitHub's label list:

| Scope | Color Family | Rationale |
|-------|-------------|-----------|
| type.* | Blue (#0075ca) | Neutral classification |
| priority.* | Red→Yellow gradient | Urgency heat map |
| triage.* | Cyan (#1D76B5) | Process flow |
| scope.* | Green/Red/Yellow | Traffic light: go/stop/caution |
| status.* | Purple (#7057ff) | State/phase |
| agent.* | Green (#0E8A16) | Active/alive |
| contrib.* | Teal (#006B75) | Inviting |
| resolution.* | Gray (#cfd3d7) | Terminal/closed |
| process.* | Light blue (#C5DEF5) | Informational |

---

## Sources

- [Addy Osmani — Agentic Engineering](https://addyosmani.com/blog/agentic-engineering/)
- [Anthropic — Building Effective Agents](https://www.anthropic.com/research/building-effective-agents)
- [Anthropic — Effective Harnesses for Long-Running Agents](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents)
- [Anthropic — Multi-Agent Research System](https://www.anthropic.com/engineering/multi-agent-research-system)
- [CPython Label Taxonomy](https://devguide.python.org/triage/labels/)
- [Sean Trane — Logical Colorful GitHub Labels](https://seantrane.com/posts/logical-colorful-github-labels-18230/)
- [GitLab Scoped Labels Discussion](https://github.com/orgs/community/discussions/16682)
- [Karpathy — Software 3.0 / Agentic Engineering](https://medium.com/generative-ai-revolution-ai-native-transformation/openai-cofounder-andrej-karpathy-signals-the-shift-from-vibe-coding-to-agentic-engineering-ea4bc364c4a1)
- [ACM DAI 2025 — Manager Agent as Unifying Research Challenge](https://dl.acm.org/doi/10.1145/3772429.3772439)
- ThreeDoors internal: SOUL.md, ROADMAP.md, envoy-rules-of-behavior-party-mode.md, BOARD.md
