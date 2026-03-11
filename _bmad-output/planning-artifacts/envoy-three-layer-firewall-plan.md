# Envoy Three-Layer Firewall — Implementation Plan

**Date:** 2026-03-10
**Source:** BOARD.md P-002 (party mode, 8 sessions)
**Original Artifact:** `envoy-scope-and-firewall-design.md` (MISSING — confirmed by dead-research-audit)
**Reconstructed From:** `envoy-rules-of-behavior-party-mode.md`, `agents/envoy.md`, `docs/envoy-operations.md`, `docs/issue-tracker.md`, MEMORY.md reference ("3-layer firewall architecture: non-LLM gates, lightweight agents, BMAD deliberation")

---

## Context

The envoy agent currently operates as a "screen and relay" agent with a single-pass screening approach. The party mode research (8 sessions) recommended formalizing this into a three-layer firewall architecture where issues pass through progressively more sophisticated filters. The goal: reduce unnecessary LLM processing, create predictable decision paths, and clearly define when full BMAD deliberation is warranted.

## The Three Layers

### Layer 1: Non-LLM Gates (Deterministic Checks)

Fast, mechanical checks that don't require AI reasoning. These can be resolved with pattern matching, string comparison, and simple lookups.

**Gate 1.1 — Spam Detection:**
- Empty body or body < 10 characters
- Known advertising patterns (URLs to unrelated products, cryptocurrency spam)
- Gibberish detection (no recognizable English words in title)
- Action: Close + notify supervisor

**Gate 1.2 — Duplicate Detection:**
- Exact title match against open issues in tracker
- Fuzzy title match (>80% similarity via keyword overlap)
- Symptom keyword matching against recently resolved issues (90-day window)
- Action: Flag as potential duplicate, link original, do NOT close

**Gate 1.3 — Already-Fixed Detection:**
- Cross-reference issue description against PRs merged in last 30 days
- Match via `Fixes #N`, `Closes #N` patterns in merged PR descriptions
- Check if issue references a component/file that was recently modified
- Action: Comment linking the fix PR, suggest verification, recommend closure to supervisor

**Gate 1.4 — Previously-Decided Detection:**
- Search BOARD.md Decided section for matching keywords
- Search BOARD.md Pending Recommendations for related work
- Check SOUL.md exclusion patterns (see `docs/issue-tracker.md` alignment reference)
- Action: If decided against → polite decline citing the decision. If in-progress → link to existing work.

**Exit criteria for Layer 1:** If any gate resolves the issue (spam closed, duplicate flagged, already-fixed linked, previously-decided cited), processing stops. Otherwise, pass to Layer 2.

### Layer 2: Lightweight AI Screening

Quick AI-based assessment by the envoy agent. This is the envoy's core reasoning step — it reads the issue, understands intent, and makes classification decisions.

**Screen 2.1 — SOUL.md Alignment Classification:**
- Clearly Aligned → proceed to priority assessment
- Clearly Misaligned → polite decline with SOUL.md reference + notify supervisor with underlying need assessment
- Gray Area → escalate to supervisor (never reject unilaterally)

**Screen 2.2 — Authority Tier Routing:**
- Tier 1 (Owner): Skip misalignment check, highest priority, always escalate direction changes
- Tier 2 (Contributor): Enhanced priority, lower escalation threshold
- Tier 3 (Community): Standard processing

**Screen 2.3 — Issue Classification & Labeling:**
- Assign category label: `bug`, `enhancement`, `question`, `documentation`
- Assess priority: P0 (blocking), P1 (important), P2 (nice-to-have)
- Identify affected components (TUI, CLI, adapter, infrastructure)

**Screen 2.4 — Scope Assessment:**
- Check ROADMAP.md for related epics/stories
- Determine if the issue fits within an existing epic, needs a new story, or is out of scope
- In-scope → relay to supervisor with triage summary
- Out-of-scope → escalate to supervisor for scope decision

**Exit criteria for Layer 2:** Issue is either declined (misaligned), resolved (question answered), or relayed to supervisor with full triage context. Only complex issues requiring multi-agent deliberation proceed to Layer 3.

### Layer 3: BMAD Deliberation (Full Escalation)

Reserved for issues that require architectural review, multi-perspective analysis, or represent potential project evolution. The envoy doesn't run Layer 3 — it recommends it to the supervisor, who decides whether to invoke party mode.

**Criteria for Layer 3 recommendation:**
- Feature request that would require a new epic (>3 stories estimated)
- Request that could change project architecture or add new patterns
- Gray-area direction request from a contributor or owner
- Issue that reveals a systemic problem (not just a point fix)
- Bug report that suggests a fundamental design flaw (not just an implementation bug)
- Any issue where 3+ agents would have relevant perspectives

**What the envoy does:**
1. Complete Layer 2 assessment
2. Add a recommendation to the supervisor escalation: "Recommend BMAD party mode for this issue because: [specific reason]"
3. Suggest which agents should participate (e.g., "Architect + PM + Dev for architecture change" or "UX + PM + QA for user-facing feature")

**What the envoy does NOT do:**
- Invoke party mode directly
- Decide to skip party mode for complex issues
- Make architectural decisions

## Implementation Approach

The three-layer firewall is primarily a documentation and agent-definition change. No Go code is involved. The deliverables are:

1. Updated `agents/envoy.md` with explicit three-layer structure
2. Updated `docs/envoy-operations.md` with decision flowcharts per layer
3. Layer 1 gate specifications with detection criteria and response templates
4. Layer 3 BMAD escalation criteria

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| — | Three-layer vs single-pass | Three distinct layers with clear exit criteria | Monolithic screening | Reduces unnecessary processing, creates predictable paths |
| — | Layer 1 scope | Deterministic pattern matching only | AI-based spam detection | Keep Layer 1 fast and predictable; AI reasoning belongs in Layer 2 |
| — | Layer 3 invocation | Envoy recommends, supervisor decides | Envoy invokes party mode directly | Envoy is a screen and relay, not a decision-maker |
| — | Duplicate handling | Flag, never close | Auto-close obvious duplicates | Even "obvious" duplicates can be subtly different (party mode consensus) |

## Stories

See story files:
- 52.1: Envoy Agent Definition — Three-Layer Firewall Architecture
- 52.2: Operations Guide — Layer Decision Flowcharts
- 52.3: Layer 1 Gate Specifications
- 52.4: Layer 3 BMAD Escalation Criteria

**Note:** Epic number 52 is provisional — must be confirmed by project-watchdog.
