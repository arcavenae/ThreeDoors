# Infrastructure Enforcement Research — Issue #907

**Date:** 2026-03-31
**Source:** Party mode (3 rounds: Architect, Dev, PM, QA, SM)
**Issue:** [#907](https://github.com/ArcavenAE/ThreeDoors/issues/907) — Replace advisory enforcement with infrastructure for epic allocation, messaging, and checkout isolation
**Status:** Research complete — recommendations ready for story creation
**Provenance:** L3 (AI-autonomous research with human PR review)

---

## Executive Summary

All four documented incidents (INC-001 through INC-004) share a root cause: **soft constraints (prose in agent definitions, memory entries, markdown registries) assumed to enforce hard properties (mutual exclusion, message delivery, filesystem isolation).** The kos knowledge graph identifies this as bedrock finding `df-codify-not-instruct`.

The existing `git-safety.sh` PreToolUse hook (INC-002 fix) proves the pattern works: mechanical enforcement via Claude Code hooks has had zero incidents since deployment. This research evaluates extending that pattern to the three unresolved enforcement gaps.

**Verdict:** Extend the hook pattern for messaging (INC-004). Use process + CI backstop for epic allocation (INC-003). Defer platform-level changes for checkout isolation (INC-001). No new services or databases needed — hooks and CI checks are the right granularity for ThreeDoors.

---

## Incident Pattern Analysis

| Incident | Soft Constraint | Hard Property Needed | Advisory Fix Applied | Did It Hold? |
|----------|----------------|---------------------|---------------------|--------------|
| INC-001 | Agent definition: "don't checkout" | Filesystem isolation | Definition rewrite + git-safety hook (partial) | Yes — no recurrence (worktrees for workers, hook blocks dangerous ops) |
| INC-002 | Memory: "sync git first" | Trust platform abstractions | Removed instruction, documented worktree model | **Resolved** — hook mechanically enforces |
| INC-003 | Markdown registry: "reserve first" | Mutual exclusion on epic numbers | project-watchdog declared as mutex in memory | **Failed** — memory policy invisible to workers |
| INC-004 | Definition: "use multiclaude message" | Reliable message delivery | Added warning to all 11 agent definitions | Partially — agents still occasionally use SendMessage |

**Key insight:** Advisory constraints have a **0% success rate under concurrent agent load** in this project. This is a category problem, not a sample size problem. Prose instructions are suggestions; hooks are enforcement.

---

## Recommendations by Tier

### Tier 1 — Ship This Week (Hours of Work, Massive ROI)

#### 1A. SendMessage PreToolUse Hook (INC-004)

**Approach:** Add a new PreToolUse hook entry in `.claude/settings.json` with matcher `"SendMessage"`. The hook script reads the JSON input, extracts `tool_input.to`, checks against a list of known multiclaude agent names, and blocks with exit code 2 if matched.

**Implementation details:**
- New file: `scripts/hooks/message-safety.sh` (~25 lines)
- Agent name list: `scripts/hooks/multiclaude-agents.txt` (one name per line, generated from `agents/*.md` directory)
- New entry in `.claude/settings.json`:
  ```json
  {
    "matcher": "SendMessage",
    "hooks": [{
      "type": "command",
      "command": "bash \"$CLAUDE_PROJECT_DIR/scripts/hooks/message-safety.sh\""
    }]
  }
  ```
- Error message should teach the correct action:
  ```
  BLOCKED: SendMessage does not route through multiclaude messaging (INC-004).
  Use: multiclaude message send <recipient> "<message>"
  Run via Bash tool, not SendMessage.
  ```

**Why this over R-016 SQLite queue:** The hook is 2-4 hours of work and eliminates the entire class of outbound message drops. The SQLite queue (R-016) is the right *eventual* answer but requires weeks of multiclaude platform work. They're phased, not competing.

**Trade-offs:**
- (+) Extends proven pattern (git-safety.sh)
- (+) Zero new dependencies
- (+) Testable with shell unit tests
- (-) Agent name list requires maintenance when agents added/removed
- (-) Only fixes outbound drops; inbound delivery via tmux paste-buffer remains fragile

**Rejected alternative:** ToolSearch blocking (prevent agents from even fetching the SendMessage schema). Too aggressive — breaks legitimate subagent use within a single Claude process. The `to` field check is the right granularity.

#### 1B. Pre-Assign Epic Numbers in Task Descriptions (INC-003)

**Approach:** Supervisor pre-assigns epic numbers in the worker task description before dispatch. "You are creating Epic N. Do not change this number."

**This is process, not infrastructure** — and that's the correct choice here. The supervisor is already a serialization point (only one supervisor dispatches workers). You don't need a mutex when you have a single writer. The race in INC-003 occurred because workers self-assigned from a shared registry, not because the supervisor dispatched conflicting numbers.

**Codification points (triple redundancy):**
1. `/plan-work` slash command template — bake "epic number: N" into the template
2. `agents/worker.md` — "Workers MUST NOT self-assign epic numbers" (already partially present)
3. Supervisor MEMORY.md dispatch checklist — "Pre-assign epic number before dispatch" (already present)

**Why triple redundancy:** INC-003's secondary cause was that the mutex policy existed in supervisor memory but workers didn't know about it. Three codification points cover three failure modes: new supervisor session (memory), new worker template (slash command), agent definition reload (worker.md).

**Rejected alternatives:**
- Lock file mechanism — over-engineered for a problem with a single-writer solution
- project-watchdog as allocation service — requires reliable messaging (circular dependency with INC-004)
- Atomic file-based registry — still races under concurrent access without OS-level locking

### Tier 2 — Ship Next Sprint (Days of Work, Good ROI)

#### 2A. Epic Number Collision CI Check (INC-003 Backstop)

**Approach:** GitHub Actions workflow triggered on PRs touching `docs/stories/*.story.md`. The workflow:
1. Lists all open PRs: `gh pr list --json number,files`
2. Extracts epic numbers from story file paths matching `docs/stories/(\d+)\.\d+\.story\.md`
3. Compares current PR's epic numbers against all other open PRs
4. Fails with clear message naming the conflicting PR if overlap detected

**Why this is Tier 2, not Tier 1:** It's detection at PR time, not prevention at work time. The process fix (Tier 1B) prevents the race at the source. This CI check is defense-in-depth for when the process fails (new supervisor, forgotten checklist item, manually dispatched worker).

**Implementation:** ~40 lines of bash in a workflow file. ~1 day including testing.

**Trade-offs:**
- (+) Catches collisions before merge (unlike merge-queue's post-hoc comment in INC-003)
- (+) Automated, no human discipline required
- (-) Races against itself: two PRs created simultaneously may not see each other until next workflow run
- (-) Fires on every story file PR, including status updates (needs path filtering)

#### 2B. Hook Agent List Maintenance (Supporting 1A)

**Approach:** Generate `scripts/hooks/multiclaude-agents.txt` from `agents/*.md` directory. Add a CI check that verifies the hook's agent list matches the directory listing — drift detection.

**Implementation:** `ls agents/*.md | sed 's|agents/||;s|\.md||' | sort` compared against the config file. Fails CI if they differ.

### Tier 3 — Defer to Platform Evolution (Weeks, Strategic)

#### 3A. SQLite Message Queue (R-016)

The R-016 research recommends SQLite-backed queue with priority levels, message TTL, heartbeat deduplication, and MCP server delivery. This is the right architecture for Marvel and dark factory scenarios. **Not appropriate for ThreeDoors today** — the hook (Tier 1A) provides 80% of the value at 5% of the cost.

**When to build:** When multiclaude is extracted as a standalone platform (P-014) or when dark factory needs reliable multi-factory messaging.

#### 3B. Per-Agent Persistent Worktrees (INC-001)

multiclaude already provides worktrees for workers. Persistent agents still share the main checkout. INC-001 has not recurred since the agent definition updates + git-safety hook deployment — the combination of advisory + partial infrastructure has been sufficient.

**When to build:** When multiclaude upstream supports per-agent worktree lifecycle management, or when dark factory requires persistent agents operating on different repositories.

**Rejected for now because:** It's a multiclaude platform change (worktree creation/cleanup in daemon lifecycle), not a ThreeDoors change. The current state works because the hook blocks the most dangerous operations and agent definitions cover the rest.

#### 3C. Full Allocation Service (project-watchdog API)

A proper allocation service where workers request epic numbers via API call and receive serialized responses. Architecturally clean but requires reliable messaging (circular dependency with INC-004) and significant multiclaude platform work.

**When to build:** When the system scales beyond one supervisor dispatching workers sequentially. Multi-supervisor or multi-repo scenarios need this.

---

## SOUL.md Alignment Analysis

| SOUL Principle | Alignment |
|---------------|-----------|
| *"Prefer the simple solution that works today"* | **Fully aligned.** Hooks are the simplest enforcement mechanism. 25 lines of bash vs weeks of platform work. |
| *"Agent infrastructure should minimize the decisions that require human attention"* | **Fully aligned.** Every hook that mechanically blocks a bad action is one fewer incident the human diagnoses. |
| *"The agent governance infrastructure is as much a part of the project as the TUI itself"* | **Fully aligned.** This research treats infrastructure enforcement as a first-class project concern. |
| *"Lessons learned here inform how future projects are structured"* | **Fully aligned.** The hook pattern is portable to any Claude Code project. The tier model (hooks → CI → platform) is a reusable framework. |

**No SOUL tension found.** Infrastructure enforcement is the most SOUL-aligned approach to governance.

---

## Architectural Principle (Confirmed)

> **Hooks and CI checks are the right granularity for ThreeDoors. Services and databases are the right granularity for Marvel.**

This principle should be recorded as a decision on BOARD.md. It resolves the "how much infrastructure is enough" question for ThreeDoors-scale projects and prevents over-engineering governance mechanisms.

---

## Implementation Roadmap

```
Week 1 (Tier 1):
├── Story: SendMessage PreToolUse hook
│   ├── scripts/hooks/message-safety.sh
│   ├── scripts/hooks/multiclaude-agents.txt
│   ├── .claude/settings.json update
│   └── Shell unit tests
├── Process: Pre-assign epic numbers
│   ├── Update /plan-work template
│   ├── Verify worker.md codification
│   └── Verify MEMORY.md checklist
└── Restart all persistent agents (definitions updated)

Week 2-3 (Tier 2):
├── Story: Epic collision CI check
│   ├── .github/workflows/epic-collision-check.yml
│   └── Agent list drift detection
└── Story: Hook test suite
    └── Automated tests for git-safety.sh + message-safety.sh

Later (Tier 3 — Marvel/platform):
├── SQLite message queue (R-016)
├── Per-agent persistent worktrees
└── Allocation service API
```

---

## Open Questions for Human Decision

1. **Should the SendMessage hook block ALL SendMessage calls or only those targeting known multiclaude agent names?** Recommendation: only known names — allows legitimate subagent use within a single Claude process.

2. **Should the epic collision CI check be a required status check (blocking merge) or advisory (warning only)?** Recommendation: required — the whole point is preventing merge of colliding PRs.

3. **Should this research produce stories immediately, or should the supervisor/PM create them through the normal planning pipeline?** This artifact is research only per the task description.

---

## Party Mode Participants

| Agent | Role | Key Contribution |
|-------|------|-----------------|
| 🏗️ Winston (Architect) | System architecture | Tiered approach, hook-vs-service granularity principle |
| 💻 Amelia (Dev) | Implementation | Hook implementation details, CI workflow design, agent list maintenance |
| 📋 John (PM) | Priority/cost-benefit | Stack-ranked tiers, SOUL.md alignment analysis |
| 🧪 Murat (QA/Test Architect) | Risk assessment | Validation matrix, test strategy, 0% success rate finding |
| 🏃 Bob (SM) | Process clarity | Triple-redundancy codification, acceptance criteria |
