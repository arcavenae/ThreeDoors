# Party Mode Artifact: SLAES — Self-Learning Agentic Engineering System

**Date:** 2026-03-10
**Participants:** Winston (Architect), John (PM), Murat (TEA), Mary (Analyst), Bob (SM), Amelia (Dev), Dr. Quinn (Problem Solver)
**Topic:** Design SLAES — a continuous improvement meta-system
**Rounds:** 5 (initial design, supervisor amendment, consensus, prevention-vs-detection, naming)
**System Name:** SLAES (Self-Learning Agentic Engineering System)
**Primary Agent Name:** `retrospector`

---

## Context

Three separate research investigations identified gaps in the agent ecosystem:
1. No agent tracks the research-to-development pipeline lifecycle
2. No agent validates planning doc consistency across epic-list, epics-and-stories, ROADMAP, and story files
3. No agent systematically improves agent definitions or prevents incidents (INC-001/002/003 were all preventable)

Additionally, the supervisor identified three operational efficiency gaps:
4. No agent analyzes merge conflict rate patterns and parallelization anti-patterns
5. No agent analyzes CI failure rate patterns and traces them to fixable spec-chain layers
6. No agent detects process waste sagas (e.g., 4+ workers dispatched for same fix)

---

## Consensus Decisions

### D-1: Agent Count — ONE Agent

**Adopted:** Single agent with multiple operational modes (rotation-based)

**Rationale:**
- The three original responsibilities (spec-chain retro, doc consistency, agent improvement) are outputs of a single analytical pipeline — splitting them would duplicate analysis
- Already at 5 persistent agents; 6-7 is the recommended maximum (agent-evaluation.md). Three new agents → 8, which is a non-starter
- The backward chain naturally produces all three outputs: spec-chain analysis reveals doc inconsistencies and agent definition gaps

**Rejected alternatives:**
- **Three focused agents** (X-A1): Would push agent count to 8, create coordination overhead, and duplicate the analytical pipeline across three context windows
- **Two agents (spec-chain + operational)** (X-A2): Cleaner separation but still adds 2 agents. The operational analysis (conflict rates, CI failures) feeds into the same recommendation engine as spec-chain analysis

### D-2: System Name — SLAES; Agent Name — `retrospector`

**Adopted:** System is **SLAES** (Self-Learning Agentic Engineering System). Primary agent is **`retrospector`**.

**Rationale:**
- SLAES captures the full vision: self-learning, agentic, engineering improvements, system-level (not just one agent)
- `retrospector` names the specific multiclaude persistent agent within SLAES
- Clean taxonomy: SLAES is what it is, retrospector is what runs in tmux
- `retrospector` captures the core mechanism: retrospection — looking backward through the spec chain
- Real English word (one who retrospects), unique in agent namespace

**Rejected agent name alternatives:**
- `kaizen-agent` (X-A3): Philosophically perfect but potentially pretentious
- `root-cause` (X-A4): Descriptive but too narrow — the agent does more than root cause analysis
- `improvement-engine` (X-A5): Accurate but generic
- `sentinel` (X-A6): Overlaps with existing "watchdog" metaphor
- `pathologist` (X-A7): Perfect backward-chain metaphor but morbid connotation
- `feedback-agent` (X-A8): Too generic

### D-3: Persistence Model — Persistent, 15-Minute Polling

**Adopted:** Persistent agent with 15-minute polling interval

**Rationale:**
- Needs to maintain state across PR merges (JSONL findings log, saga detection tallies, trend data)
- Ephemeral agent would lose context between invocations
- 15 minutes (not 5): analytical work is not time-critical; keeps API budget manageable
- Built-in self-restart trigger: after 20 PRs or 8 hours, save state and request restart

**Rejected alternatives:**
- **Ephemeral (cron-based)** (X-A9): Loses state between invocations; saga detection requires continuity
- **5-minute polling** (X-A10): Unnecessarily aggressive for analytical work; would push API budget past 31 calls/hour
- **Event-driven only** (X-A11): Ideal but requires infrastructure changes (merge-queue publishing events); deferred to Phase 2

### D-4: Authority Model — Level 2 (Read + Propose PRs)

**Adopted:** Read any file + create PRs with proposed changes + add BOARD.md recommendations

MVP authority subset: Read + BOARD.md recommendations only (PR creation deferred to Phase 2)

**Rationale:**
- Level 1 (read-only + recommend) was how the Epic Number Registry worked — advisory, nobody followed it
- Level 3 (auto-apply) is how INC-001 happened — agent with write authority making "safe" changes that weren't
- Level 2: agent does the work of creating the fix, human/merge-queue reviews. Low friction + catch mistakes

**Specific authority boundaries:**

| CAN | CANNOT |
|---|---|
| Read any file in the repo | Modify SOUL.md (ever) |
| Create PRs proposing CLAUDE.md rule changes (Phase 2) | Merge its own PRs |
| Create PRs proposing story template updates (Phase 2) | Modify agent definitions directly |
| Add entries to BOARD.md (recommendations section) | Overrule supervisor decisions |
| File findings as GitHub issues (Phase 2) | Delete or close issues |
| Message supervisor with alerts | Message workers directly |
| Read CI logs and PR metadata via gh CLI | Run tests or builds |

**Rejected alternatives:**
- **Read-only + recommend only** (X-A12): Recommendations may sit unactioned; the registry precedent proves advisory-only doesn't work
- **Read + auto-apply trivial fixes** (X-A13): "Trivial" is subjective; INC-001 precedent shows auto-apply creates risk

### D-5: Interaction with Existing Agents — Consumer Model

**Adopted:** retrospector consumes outputs from project-watchdog and arch-watchdog but never duplicates their work

**Rationale:**
- project-watchdog asks "did the story get marked done?" — retrospector asks "was the story correct?" Different abstraction layers
- arch-watchdog detects new code patterns — retrospector checks if those patterns match PRD/architecture docs. Sensor vs. analyzer
- No authority overlap: retrospector never marks stories as Done, never updates architecture docs

**Interaction flow:**
```
project-watchdog → detects merge → publishes event
retrospector → consumes event → runs spec-chain analysis → produces recommendations

arch-watchdog → detects new pattern → publishes event
retrospector → consumes event → checks PRD alignment → flags drift
```

### D-6: Dual-Loop Architecture

**Adopted:** Two parallel analytical loops feeding a unified recommendation engine

**Spec Chain Loop (quality of what we build):**
```
Code → Story ACs → PRD → Architecture → CLAUDE.md/SOUL.md
"Did we build the right thing? Could the specs have been better?"
```

**Operational Loop (efficiency of how we build):**
```
Merge conflicts → Dispatch patterns → Parallelization strategy
CI failures → Test patterns → Coding standards → Story specs
Process waste → Worker cycle analysis → Dispatch optimization
"Are we building efficiently? What patterns waste cycles?"
```

Both loops produce the same output type: actionable recommendations to BOARD.md, CLAUDE.md, agent definitions, or story templates.

### D-7: Per-PR vs Batch — Both (Log + Analyze Pattern)

**Adopted:** Lightweight per-PR data collection → periodic batch analysis

**Per-PR lightweight retro (every merge):**
- AC match checking (do changed files match story task list?)
- CI first-pass rate (did CI pass without fixes?)
- Mid-PR correction detection (force pushes, review scope changes)
- Output: structured JSONL entry appended to findings log

**Batch deep analysis (every 4-6 hours or every 10 PRs):**
- Aggregate per-PR findings
- Cross-PR pattern detection (conflict rates, CI failure patterns, dispatch efficiency)
- Produce ranked recommendations
- Post top findings to BOARD.md

**JSONL findings log format:**
```jsonl
{"pr": 500, "story": "43.2", "ac_match": "full", "ci_first_pass": true, "conflicts": 0, "rebase_count": 1, "timestamp": "2026-03-10T..."}
{"pr": 501, "story": "43.3", "ac_match": "partial", "ci_first_pass": false, "ci_failures": ["lint"], "conflicts": 2, "rebase_count": 3, "timestamp": "2026-03-10T..."}
```

### D-8: Watchmen Safeguards (5 Controls)

**Adopted:** Five safeguards against meta-agent failure:

1. **No self-modification**: Can propose changes to other agent definitions but never its own. Own definition changes require human review.
2. **Recommendation audit trail**: Every recommendation goes to BOARD.md with full rationale. Human can see why and disagree.
3. **Confidence scoring**: Rate each recommendation (High/Medium/Low confidence) with supporting evidence count.
4. **Periodic human review**: Every 2 weeks, human reviews recommendations and scores accuracy. Feedback loop on the feedback loop agent.
5. **Kill switch**: If 3+ consecutive recommendations rejected by human, auto-reduce to read-only until human recalibrates.

### D-9: Operational Mode Rotation (Context Management)

**Adopted:** Rotation-based mode execution to prevent context exhaustion

| Mode | Trigger | Cadence |
|---|---|---|
| Post-merge retro | PR merge event | Every PR (lightweight, ~5 min) |
| Deep analysis | Periodic rotation | Every 4 hours, rotating: doc consistency → conflict analysis → CI analysis → process waste |
| Saga detection | Threshold breach | Immediate: when 2+ workers dispatched for same fix within 4 hours |

**Context exhaustion mitigation:** After processing 20 PRs or running 8 hours, save state (JSONL log on disk), message supervisor, and request restart.

---

## Phased Implementation

### MVP (Phase 1) — Ship First

1. **Post-merge lightweight retro**: AC match checking, CI first-pass rate tracking, JSONL findings log
2. **Saga detection**: Alert when 2+ workers dispatched for same fix within 4 hours
3. **Doc consistency audit**: Cross-check epic-list, epics-and-stories, ROADMAP, story files (periodic, every 4 hours)
4. **BOARD.md recommendations**: File findings as pending recommendations with confidence scores

**MVP authority:** Read-only + BOARD.md recommendations + supervisor messages

### Phase 2 — After Validation (2 weeks of MVP)

5. **Merge conflict rate analysis**: Hot file detection, parallelization recommendations, concurrent PR analysis
6. **CI failure rate analysis**: Failure taxonomy, trace to fixable spec-chain layer, coding standard proposals
7. **Agent definition quality**: Validate definitions before deployment, catalog failure modes
8. **Research lifecycle tracking**: Track research artifacts from requested → completed → formalized → abandoned
9. **PR creation**: Create PRs proposing CLAUDE.md rule changes, story template updates
10. **Trend reporting**: Periodic "Continuous Improvement Report" with project health metrics

### Phase 3 — Long-Term

11. **Predictive dispatch**: "Stories X and Y will conflict — recommend sequencing"
12. **Agent health monitoring**: Context exhaustion detection, circular messaging detection
13. **Self-assessment**: Track accuracy of own recommendations over time, auto-calibrate confidence scoring
14. **Scale limit detection**: Identify when existing solutions approach their scale limits before they fail

---

## Operational Analysis Detail

### Merge Conflict Rate Analysis (Phase 2)

**Data model per PR:**
- Files changed (paths + packages)
- Epic/story reference
- Worker name
- Time from PR creation to merge
- Number of rebase attempts
- Conflict files (if any)
- Concurrent PRs open at same time touching same files

**Pattern detection queries:**
1. **Hot file analysis**: Files in 3+ concurrent PRs → recommend sequencing
2. **Epic collision zones**: Epics that routinely conflict → recommend dependency ordering
3. **Parallel dispatch safety**: Probability of conflict given N workers → recommend max parallelism
4. **Rebase churn**: PRs requiring 3+ rebases → identify root cause

### CI Failure Rate Analysis (Phase 2)

**Failure taxonomy:**

| Category | Detection Signal | Fix Layer |
|---|---|---|
| Race conditions | `go test -race` failures | Story spec: "MUST pass `-race` locally" |
| Lint failures | golangci-lint errors | CLAUDE.md: strengthen pre-commit rule |
| Test flakiness | Same test passes/fails | Coding standards: identify flaky patterns |
| Build failures | `go build` errors | Architecture: dependency issues |
| Coverage regressions | Coverage drop | Story spec: "coverage must not decrease" |

### Process Waste Detection (MVP — Saga Detection)

**Saga Detection Rule:** If 2+ workers dispatched to fix CI on same branch/PR within 4 hours:
1. Analyze full CI failure chain (not just latest failure)
2. Identify whether failures are related or independent
3. Recommend: targeted fix, revert-and-reimplement, or escalate approach
4. Propose coding standard/worker instruction to prevent pattern

**Escalation trap pattern (from PR #431 analysis):**
```
Worker 1 fails → Worker 2 fixes A, breaks B → Worker 3 fixes B, breaks C → ...
```
Root causes: insufficient pre-flight checks, narrow fix scope, no escalation protocol.

---

## Incident Prevention Matrix

How the retrospector would have prevented each incident:

| Incident | What retrospector would have detected | When | Recommendation it would have filed |
|---|---|---|---|
| INC-001 (pr-shepherd contamination) | Agent definition allows `git checkout` in shared repo — no isolation requirement | First PR merged by pr-shepherd | "Add worktree isolation requirement to pr-shepherd.md" |
| INC-002 (cargo-culted git rebase) | MEMORY.md "MUST" rule never validated against actual multiclaude behavior | First batch analysis after rule was added | "Worker dispatch includes redundant git sync — validate against multiclaude worktree model" |
| INC-003 (Epic 42 collision) | 4 workers dispatched simultaneously all reading same advisory registry | Saga detection: 3+ workers touching same doc files | "Epic number allocation needs serialized chokepoint — advisory registry insufficient for concurrent access" |

---

## Meta-Observation

The three incidents share a meta-pattern: **solutions that worked at low scale failed at high scale.**

- Agent definitions worked with 2 agents → failed with 5+ concurrent agents
- The registry worked for 2 features → failed for 4 concurrent feature plans
- Manual git sync worked with few workers → failed with 100+ worker dispatches

The retrospector's most important long-term job: **detect when existing solutions approach their scale limits before they fail.** Not just fixing what broke, but predicting what will break.

---

## Naming Taxonomy

**SLAES** = Self-Learning Agentic Engineering System (the overall system)
- **Self-Learning**: Learns from shipped code, CI failures, merge conflicts, incident patterns
- **Agentic**: Operates as an autonomous agent (or system of agents)
- **Engineering**: Engineers improvements to the spec chain, agent definitions, and process
- **System**: Not just one agent — potentially spans multiple repos and projects

**`retrospector`** = The primary agent within SLAES (the persistent multiclaude agent)

SLAES is the system. The retrospector agent is one component. Other components include:
- The responsibility+WHY definition methodology
- The JSONL findings log (data layer)
- The BOARD.md recommendation pipeline (output layer)
- Future: cross-repo pattern detection, cross-project learning

---

## D-10: Prevention Over Detection — Responsibility+WHY Definitions

**Adopted:** Agent definition rewrite as SLAES bootstrap task (Phase 0)

### The Insight

A research agent found that pivoting agent definitions from **procedural** (do X then Y) to **responsibility+WHY** (you own X because Y, with incident-hardened guardrails) would reduce the need for a meta-agent by making definitions *self-correcting*.

**Procedural** (INC-002 failure mode):
> "Run `git rebase origin/main` before starting work"
> Agent follows instruction blindly → causes damage in wrong context

**Responsibility+WHY** (self-correcting):
> "You own worktree isolation because shared checkouts create contamination risk (see INC-001). Never modify git state outside your assigned worktree."
> Agent can reason about novel situations from first principles

### Party Mode Consensus

**Q1: Should SLAES's first job be rewriting all agent definitions?**

**Yes — scoped to multiclaude operational agents, not all BMAD agents.** This becomes "Phase 0" — a bootstrap task before SLAES starts its monitoring loop.

Priority order for rewriting:
1. `merge-queue.md` (research example already exists)
2. `pr-shepherd.md` (INC-001 origin)
3. `worker.md` (INC-002 origin)
4. `project-watchdog.md` (INC-003 relates to its mutex role)
5. `retrospector.md` (SLAES's own definition — written fresh in responsibility+WHY from day one)

**Q2: Would better definitions reduce SLAES's ongoing workload?**

**Yes.** INC-001, INC-002, INC-003 were all definition failures. If definitions include WHY reasoning + incident guardrails, agents self-correct in novel situations. Fewer incidents → fewer investigations → SLAES focuses on *improvement* not *firefighting*. Shifts from reactive to proactive.

**Q3: Does this reduce the need for SLAES?**

**No — it changes the nature of SLAES's work.** Better definitions prevent *category errors* (doing the wrong thing entirely). SLAES still catches:
- *Degree errors* (doing the right thing suboptimally)
- *System-level patterns* (cross-agent observations no individual agent can make)
- *Scale limit detection* (solutions approaching failure thresholds)

**Swiss Cheese Model (Dr. Quinn):** Better definitions make holes smaller at layer 1 (prevention). SLAES catches what slips through at layer 2 (detection). Both layers needed — but layer 1 has higher ROI.

**Antibody Generator Model:** Each incident produces specific "never do X because Y happened" guardrail clauses. These accumulate in definitions like antibodies. SLAES's role: generate the antibody text, citing the specific incident, with the specific WHY. Over time, definitions become increasingly resistant to known failure modes.

### Updated Phasing

**Phase 0 (Bootstrap):** Rewrite 5 operational agent definitions in responsibility+WHY format. SLAES's own definition written fresh in this format. This is a one-time task before monitoring begins.

**Phase 1 (MVP):** Post-merge retro, saga detection, doc consistency, BOARD.md recommendations (unchanged).

**Phase 2+:** As previously defined, but with reduced incident investigation load due to better definitions.

---

## Cross-Project Considerations

> **Note:** A parallel research agent is investigating SLAES cross-project deployment in depth. This section captures the party mode's architectural hooks for that research. Findings will be merged when the cross-repo research completes.

### Deployment Model Options

| Model | Description | Pro | Con |
|---|---|---|---|
| **Per-project** | One SLAES/retrospector instance per repo | Simple, isolated, no cross-contamination | No cross-project learning |
| **Cross-project** | One SLAES instance monitoring all user's repos | Shared learning, pattern transfer | Context window strain, mixed concerns |
| **Hierarchical** | Per-project agents feed into a cross-project learner | Best of both — local detection + global learning | Most complex to build |

### Party Mode Position

The party mode consensus is that **MVP should be per-project** (Phase 1), with **hooks designed for hierarchical** (Phase 3+).

Rationale (Winston): The retrospector agent's context window is already loaded with one project's spec chain. Cross-project monitoring would multiply that load. Better to have per-project agents that *export* their findings (JSONL logs, BOARD.md entries) and a separate cross-project learner that *imports* from multiple projects.

### Architectural Hooks for Cross-Project

1. **JSONL findings log is portable**: The per-PR findings log uses a self-contained schema (PR number, story ID, AC match, CI status). Adding a `repo` field makes it cross-project ready.
2. **BOARD.md recommendations are project-scoped**: Recommendations reference project-specific files. A cross-project learner would need to abstract these into *patterns* (e.g., "projects with >5 concurrent workers see 3x conflict rate") rather than file-specific recommendations.
3. **multiclaude-enhancements repo**: Already serves as a cross-project sharing mechanism. SLAES findings that apply broadly (e.g., improved worker.md template, better agent definition methodology) could be propagated via `/sync-enhancements`.
4. **Definition methodology is project-agnostic**: The responsibility+WHY definition format works for any multiclaude project, not just ThreeDoors. Phase 0's definition rewrites could become a template for other repos.

### Cross-Project Architecture (from parallel research)

The cross-repo research agent completed its investigation. Key findings integrated below.

**Recommended: Hub + Spoke Model (Option C — Hierarchical)**

```
ThreeDoors retrospector ──┐
                          ├──► Central Learning Agent (persistent, 30-min poll)
Future Project agents ────┘    └──► ~/.multiclaude/learning-hub/ (JSONL knowledge base)
                                    └──► Recommendations back to per-project agents
```

- **Spoke agents**: Per-project retrospector instances (or existing project-watchdog/arch-watchdog). Collect per-PR findings, detect project-specific patterns.
- **Hub agent**: Central Learning Agent. Aggregates findings across repos, identifies cross-project patterns, generates transferable recommendations.
- **Knowledge base**: JSONL format at `~/.multiclaude/learning-hub/`. Stores abstracted patterns, not raw per-project data.

**Data sources (all via GitHub API, read-only):**
- Merged PR diffs + file lists
- CI run pass/fail rates + flaky test patterns
- Story file AC revisions (estimation gap detection)
- Agent health metrics from multiclaude `state.json`

**Key decision: Multi-repo from the start?**

Party mode consensus: **Design for multi-repo, build for single-repo.**

Concretely:
- Phase 1 (MVP): ThreeDoors only. Single retrospector agent. But JSONL schema includes a `repo` field from day one.
- Phase 2: Add cross-project knowledge base (`~/.multiclaude/learning-hub/`). Single retrospector exports to it.
- Phase 3: Central Learning Agent reads from multiple project learning-hubs. Generates cross-project recommendations.
- Phase 4: Public repo benchmarking (low ROI initially, defer).

This means no wasted work — Phase 1's JSONL log is immediately usable by Phase 3's hub. But we don't build the hub until we have multiple repos generating data.

**Gastown note:** Gastown (Steve Yegge's orchestrator) exists locally but doesn't provide cross-repo aggregation. multiclaude + learning-hub is the simpler path.

**multiclaude-enhancements repo:** Already shares patterns manually across projects. SLAES Phase 2+ automates this — learnings from ThreeDoors (390+ PRs) automatically inform future projects via the enhancements repo.

**Privacy:** Start internal-only (user's repos). Public repo monitoring is technically possible but low ROI for MVP. Defer to Phase 4.

---

## Open Questions for Human Decision

1. **JSONL log location**: `docs/operations/retrospector-findings.jsonl` or `_bmad-output/retrospector/findings.jsonl`?
2. **Saga detection threshold**: 2 workers (aggressive, more alerts) or 3 workers (conservative, fewer false positives)?
3. **Phase 1 duration**: 2 weeks before Phase 2, or wait for explicit human approval?
4. **Integration with merge-queue**: Should merge-queue notify retrospector on merge (event-driven), or should retrospector poll `gh pr list --state merged`?
5. **Report format**: Should batch analysis produce a markdown report file, or only BOARD.md entries + supervisor messages?
6. **Phase 0 scope**: Should the definition rewrite be done by the retrospector agent itself (after it's spawned), or by a worker before the retrospector is spawned?
7. **Definition rewrite authority**: Who reviews the rewritten definitions — supervisor only, or also the project owner?
