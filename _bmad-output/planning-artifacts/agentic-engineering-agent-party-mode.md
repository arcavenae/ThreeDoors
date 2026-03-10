# Party Mode Artifact: Agentic Engineering Agent Design

**Date:** 2026-03-10
**Participants:** Winston (Architect), John (PM), Murat (TEA), Mary (Analyst), Bob (SM), Amelia (Dev), Dr. Quinn (Problem Solver)
**Topic:** Design the Agentic Engineering Agent — a continuous improvement meta-agent
**Rounds:** 3 (initial design, supervisor amendment, consensus)

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

### D-2: Agent Name — `retrospector`

**Adopted:** `retrospector`

**Rationale:**
- Captures the core mechanism: retrospection — looking backward through the spec chain from shipped code to fundamentals
- Real English word (one who retrospects), unique in agent namespace
- Not as dry as "auditor," not as pretentious as "kaizen," not as narrow as "retro-agent"

**Rejected alternatives:**
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

## Open Questions for Human Decision

1. **JSONL log location**: `docs/operations/retrospector-findings.jsonl` or `_bmad-output/retrospector/findings.jsonl`?
2. **Saga detection threshold**: 2 workers (aggressive, more alerts) or 3 workers (conservative, fewer false positives)?
3. **Phase 1 duration**: 2 weeks before Phase 2, or wait for explicit human approval?
4. **Integration with merge-queue**: Should merge-queue notify retrospector on merge (event-driven), or should retrospector poll `gh pr list --state merged`?
5. **Report format**: Should batch analysis produce a markdown report file, or only BOARD.md entries + supervisor messages?
