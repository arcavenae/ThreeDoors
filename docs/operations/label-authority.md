# Label Authority Matrix & Triage Flow

> Operational reference for GitHub label authority, triage pipeline, and scoped label conventions.
> Authoritative source for which agents can set/remove which labels.

**Related documents:**
- Agent definitions: `agents/envoy.md`, `agents/merge-queue.md`
- Envoy operations: `docs/envoy-operations.md`
- Party mode deliberation: `_bmad-output/planning-artifacts/scoped-labels-party-mode.md`
- Envoy rules of behavior: `_bmad-output/planning-artifacts/envoy-rules-of-behavior-party-mode.md`
- BOARD decisions: D-106 through D-111, P-003, P-005

---

## Label Authority Matrix

All 27 labels in the ThreeDoors taxonomy, organized by scope. Each entry specifies who can set the label, who can remove it, and whether the action is autonomous or requires approval.

### type.* (5 labels)

Classification labels applied during triage. Mutually exclusive.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `type.bug` | Envoy | Yes | Envoy, supervisor | Something isn't working as expected |
| `type.feature` | Envoy | Yes | Envoy, supervisor | New feature or enhancement request |
| `type.docs` | Envoy | Yes | Envoy, supervisor | Documentation improvement needed |
| `type.question` | Envoy | Yes | Envoy, supervisor | Question about usage or behavior |
| `type.infra` | Envoy | Yes | Envoy, supervisor | CI/CD, tooling, or build system change |

### priority.* (4 labels)

Severity/urgency assessment. Mutually exclusive.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `priority.p0` | Envoy | Yes | Supervisor (override) | Blocks users — immediate response required |
| `priority.p1` | Envoy | Yes | Supervisor (override) | Important — address this sprint |
| `priority.p2` | Envoy | Yes | Supervisor (override) | Backlog — address when capacity allows |
| `priority.p3` | Envoy | Yes | Supervisor (override) | Low urgency — someday/maybe |

**Note:** Envoy sets priority autonomously based on triage assessment. Supervisor may override if the reporter's stated urgency differs from the envoy's assessment (e.g., reporter says P0, envoy assesses P2).

### triage.* (4 labels)

Triage pipeline state. Mutually exclusive — represents current triage stage.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `triage.new` | Envoy | Yes | Envoy | Issue acknowledged, entering triage |
| `triage.in-progress` | Envoy | Yes | Envoy | Envoy actively triaging this issue |
| `triage.needs-info` | Envoy | Yes | Envoy | Waiting on reporter for clarification |
| `triage.complete` | Envoy | Yes | Envoy | Triage finished — story may exist |

### scope.* (3 labels)

Roadmap alignment assessment. Mutually exclusive.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `scope.in-scope` | Envoy (propose), supervisor (decide) | Requires approval | Supervisor | Fits current roadmap |
| `scope.out-of-scope` | Envoy (propose), supervisor (decide) | Requires approval | Supervisor | Outside current project direction |
| `scope.needs-decision` | Envoy (propose), supervisor (decide) | Requires approval | Supervisor | Requires human evaluation for scope |

**Note:** Envoy may propose `scope.out-of-scope` for clearly misaligned requests per SOUL.md, but closure decisions always require supervisor approval.

### status.* (4 labels)

Issue/PR state signals. NOT mutually exclusive — an issue can be both `blocked` and `stale`.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `status.blocked` | Any agent | Yes | Resolving agent | Blocked on dependency or decision |
| `status.stale` | Envoy, PR-shepherd | Yes | Any agent on activity | No activity past staleness threshold |
| `status.needs-human` | Any agent | Yes | Human | Blocked on human action or decision |
| `status.do-not-merge` | Supervisor, human | Yes (supervisor) | Supervisor, human | Must not merge even if CI passes |

**Note:** `status.do-not-merge` is a hard stop for merge-queue. Merge-queue MUST NOT merge any PR with this label, regardless of CI status or review approvals.

### agent.* (2 labels)

Agent assignment visibility. NOT mutually exclusive — envoy triages, worker implements.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `agent.envoy` | Envoy | Yes | Envoy | Envoy is triaging or responsible |
| `agent.worker` | Supervisor | Yes | Supervisor, worker | Worker agent actively implementing |

**Note:** Agent labels serve human dashboard queries ("who is handling this?"). Agents should never rely on labels to determine their own workload — they use story files, messages, and GitHub API queries for coordination.

### contrib.* (2 labels)

Community contributor signals. NOT mutually exclusive.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `contrib.good-first-issue` | Envoy, supervisor | Yes | Envoy, supervisor | Good for newcomers to the project |
| `contrib.help-wanted` | Envoy, supervisor | Yes | Envoy, supervisor | Community contributions welcome |

### resolution.* (2 labels)

Closed-issue resolution classification. Mutually exclusive.

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `resolution.duplicate` | Envoy (propose), supervisor (confirm) | Requires approval | Supervisor | Duplicate of an existing issue |
| `resolution.wontfix` | Envoy (propose), supervisor (confirm) | Requires approval | Supervisor | Will not be addressed — see comment for reason |

**Note:** Envoy flags potential duplicates and proposes resolution labels but cannot close issues unilaterally (except spam). Supervisor confirms before closure.

### process.* (1 label)

Workflow modifier. Not applicable for mutual exclusivity (single label).

| Label | Set By | Autonomous? | Removed By | Notes |
|-------|--------|-------------|------------|-------|
| `process.fast-track` | Envoy | Yes | Envoy | Trivial fix — skip full triage pipeline |

---

## End-to-End Triage Flow

The standard triage pipeline processes every new issue through a defined sequence. The envoy drives steps 1-7; the supervisor decides the triage approach in step 6.

```
┌─────────────────────────────────────────────────────────┐
│  1. Issue Created                                       │
│     Reporter opens issue on GitHub                      │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  2. Envoy Detects on Patrol                             │
│     Poll: gh issue list --state open                    │
│     Compare against docs/issue-tracker.md               │
│     Apply: triage.new + type.* label                    │
│     Post acknowledgment comment to reporter             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  3. Layer 1: Deterministic Screening                    │
│     Gate 1.1: Spam detection                            │
│     Gate 1.2: Duplicate detection                       │
│     Gate 1.3: Already-fixed detection                   │
│     Gate 1.4: Previously-decided detection              │
│                                                         │
│     If screened out → notify supervisor, STOP           │
│     If passes all gates → continue to Layer 2           │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  4. Layer 2: Lightweight AI Screening                   │
│     Screen 2.1: SOUL.md alignment check                 │
│       - Clearly aligned → continue                     │
│       - Clearly misaligned → polite decline + STOP     │
│       - Gray area → escalate to supervisor             │
│     Screen 2.2: Authority tier routing                  │
│     Screen 2.3: Classification & labeling               │
│     Screen 2.4: Scope assessment (ROADMAP.md check)     │
│                                                         │
│     Apply: triage.in-progress                           │
│     If needs info → apply triage.needs-info, ask        │
│     reporter, WAIT                                      │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  5. Envoy Messages Supervisor                           │
│     Triage summary with:                                │
│       - Issue classification (type, priority)           │
│       - Scope assessment                                │
│       - Recommended approach                            │
│       - Layer 3 BMAD recommendation (if applicable)     │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  6. Supervisor Decides Triage Approach                  │
│     Options:                                            │
│       a) Fast-track → skip to story creation            │
│       b) BMAD pipeline → PM examination, party mode     │
│       c) Reject → envoy posts decline, closes issue     │
│       d) Hold → await more information                  │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  7. Triage Completion                                   │
│     Apply: triage.complete + scope.* + priority.*       │
│     Update docs/issue-tracker.md                        │
│     Post triage summary comment to reporter             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  8. Story Created                                       │
│     Story file: docs/stories/X.Y.story.md               │
│     Issue linked to story file in tracker               │
│     Reporter notified: "We've created a story for this" │
└─────────────────────────────────────────────────────────┘
```

### Label State Transitions During Triage

| Step | Labels Applied | Labels Removed |
|------|---------------|----------------|
| Issue detected | `triage.new`, `type.*` | — |
| Screening begins | `triage.in-progress` | `triage.new` |
| Needs reporter info | `triage.needs-info` | `triage.in-progress` |
| Reporter responds | `triage.in-progress` | `triage.needs-info` |
| Triage complete | `triage.complete`, `scope.*`, `priority.*` | `triage.in-progress` |
| Worker assigned | `agent.worker` | — |
| PR opened | — | — |
| PR merged / issue resolved | — | `agent.worker`, `agent.envoy` |

---

## Fast-Track Flow

The fast-track flow is a shortcut for trivial fixes that don't require the full BMAD pipeline. The envoy identifies candidates and labels them; the supervisor dispatches a worker directly.

```
┌─────────────────────────────────────────────────────────┐
│  1. Envoy Identifies Trivial Fix                        │
│     Criteria:                                           │
│       - Fix is obvious from the issue description       │
│       - Scope is small (typo, config, single-file)      │
│       - No architectural implications                   │
│       - No SOUL.md alignment concerns                   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  2. Envoy Applies process.fast-track                    │
│     Also applies: type.*, priority.*, triage.complete   │
│     Messages supervisor: "Fast-track candidate: [why]"  │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  3. Skip Full BMAD Pipeline                             │
│     No PM examination needed                            │
│     No party mode needed                                │
│     Story created directly from issue                   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  4. Supervisor Dispatches Worker                        │
│     Worker implements fix, creates PR                   │
│     Standard PR review and merge process                │
└─────────────────────────────────────────────────────────┘
```

### What Qualifies for Fast-Track

| Qualifies | Does NOT Qualify |
|-----------|-----------------|
| Typo fixes in docs or code | New features (even small ones) |
| Missing error messages | API changes |
| Config value corrections | Anything touching `internal/tui/` |
| Broken links in documentation | Changes requiring architectural decisions |
| Label/metadata updates | Issues from first-time contributors (give full triage as welcome) |
| Test fixture updates | Anything referenced in BOARD.md |

---

## Scoped Label Mutual Exclusivity Rules

GitHub does not enforce mutual exclusivity natively (unlike GitLab's scoped labels). Agents must enforce by convention: **remove the old label before applying a new one within an exclusive scope.**

### Mutually Exclusive Scopes

These scopes allow only ONE label per issue/PR at any time.

| Scope | Labels | Enforcement |
|-------|--------|-------------|
| `type.*` | bug, feature, docs, question, infra | An issue is one type. If reclassified, remove the old type label first. |
| `priority.*` | p0, p1, p2, p3 | An issue has one priority. If reprioritized, remove the old priority label first. |
| `triage.*` | new, in-progress, needs-info, complete | Triage is a state machine — only one state at a time. Transition removes the previous state. |
| `scope.*` | in-scope, out-of-scope, needs-decision | An issue has one scope determination. |
| `resolution.*` | duplicate, wontfix | A closed issue has one resolution reason. |

### Non-Exclusive Scopes

These scopes allow MULTIPLE labels per issue/PR simultaneously.

| Scope | Labels | Rationale |
|-------|--------|-----------|
| `status.*` | blocked, stale, needs-human, do-not-merge | An issue can be both `blocked` AND `stale`. A PR can be `do-not-merge` while also `needs-human`. |
| `agent.*` | envoy, worker | Envoy triages (sets `agent.envoy`), then supervisor assigns worker (adds `agent.worker`). Both labels valid simultaneously. |
| `contrib.*` | good-first-issue, help-wanted | An issue can be both a good first issue AND one where help is wanted. |

### Convention Enforcement Protocol

When applying a label in a mutually exclusive scope:

1. Query current labels on the issue: `gh issue view <number> --json labels`
2. Check if any label in the same scope is already applied
3. If yes: remove the existing label first: `gh issue edit <number> --remove-label <old-label>`
4. Apply the new label: `gh issue edit <number> --add-label <new-label>`

**Example — reprioritizing from P2 to P0:**
```bash
gh issue edit 42 --remove-label priority.p2
gh issue edit 42 --add-label priority.p0
```

**Future consideration:** A lightweight GitHub Action could enforce mutual exclusivity automatically (flag or auto-correct when two labels from an exclusive scope are applied). This is tracked as a potential future story per Murat's dissenting opinion in the scoped labels party mode (SL-013 discussion).

---

## Summary: Agent Label Authority

Quick reference showing which agents interact with which label scopes.

| Agent | Can Set | Can Remove | Notes |
|-------|---------|------------|-------|
| **Envoy** | type.*, priority.*, triage.*, agent.envoy, process.fast-track, contrib.*, status.stale, status.blocked, status.needs-human | type.*, triage.*, agent.envoy, process.fast-track, contrib.* | Primary label manager. Proposes scope.* and resolution.* but doesn't decide. |
| **Supervisor** | scope.*, resolution.*, status.do-not-merge, agent.worker, contrib.* | scope.*, priority.* (override), status.do-not-merge, agent.worker, resolution.*, contrib.* | Final authority on scope, resolution, and priority overrides. |
| **Merge-queue** | type.*, scope.in-scope, agent.worker | — | Applies type/scope/agent labels to PRs during merge validation. Reads `status.do-not-merge` as a hard stop. Also uses `broke-main` (outside the 27-label taxonomy). |
| **PR-shepherd** | status.stale | — | Can flag stale PRs. |
| **Workers** | — | agent.worker | Can remove their own agent label on completion. |
| **Human** | Any | Any | Full authority. Can override any agent's label decisions. |
