# Scoped Labels Taxonomy — Party Mode Deliberation

**Date:** 2026-03-08
**Participants:** PM (John), SM (Bob), Architect (Winston), TEA/Test Architect (Murat)
**Topic:** Finalizing the scoped label taxonomy for ThreeDoors
**Input:** [Scoped Labels Research](scoped-labels-research.md)
**Facilitator:** BMad Master
**Rounds:** 3 (Taxonomy Review, Consumer Challenge, Final Consensus)

---

## ROUND 1: Taxonomy Review — Are Categories Complete? Any Gaps?

### Opening Position Statements

📋 **John (PM):** "The research identifies 9 scopes with 35 total labels. That's a 67% increase from our current 21. My first question: is every new scope earning its keep? Let me walk through each.

- `type.*` — Yes, absolutely. We already have these unscoped; scoping them is pure cleanup.
- `priority.*` — Yes, just reformatting the separator.
- `triage.*` — Yes, same.
- `scope.*` — Yes, same.
- `status.*` — This is the biggest new addition (6 labels). I'm concerned about overlap with GitHub's native PR review states. GitHub already tracks 'changes requested' and 'approved' in its review system. Do we need labels duplicating that?
- `agent.*` — Novel and interesting. This is the agentic engineering differentiator. But I want to challenge whether 5 agent labels are needed.
- `contrib.*` — Standard open source practice. Two labels is minimal and appropriate.
- `resolution.*` — Useful for closed-issue archaeology. Three is fine.
- `process.*` — Only 2. Fine."

🏗️ **Winston (Architect):** "I broadly agree with the taxonomy structure. The `.` separator is the right call — it's the most universally understood namespace separator. My concerns:

1. **`status.*` overlap with GitHub native states:** GitHub Reviews already track approved/changes-requested at the review level. But labels serve a DIFFERENT audience — agents querying `gh issue list --label status.approved` can't easily query review status. The labels are for agent consumption, not human consumption. I'd keep them.

2. **`agent.*` granularity:** Five agent labels maps 1:1 to our persistent agents. But what happens when we add a new persistent agent? We'd need a new label. This is acceptable IF we commit to the pattern that every persistent agent gets a label. It's a small, bounded set.

3. **Missing: `status.do-not-merge`:** CPython has this. We should too. Merge-queue needs a way to know 'this PR is open intentionally but should NOT be merged even if CI passes.' Draft PRs partially serve this, but not all 'do not merge' situations are drafts."

🏃 **Bob (SM):** "My review from a workflow perspective:

**Gaps I see:**
1. **No `status.ready` or `status.merge-ready`** — How does merge-queue know a PR is ready? Currently it checks CI + approval. But what about PRs that are approved and green but should wait for a dependency? `status.blocked` covers the negative case, but there's no positive 'go ahead and merge this' signal beyond the absence of blockers.

2. **Staleness automation:** `status.stale` is good, but who sets it? The envoy checks issues on 14/30/21 day thresholds. PR-shepherd checks PR staleness. Both should be able to set this label. Is there a conflict?

3. **`process.party-mode` — is this actually useful?** When would you label an issue 'needs party mode' BEFORE running party mode? Party mode is a decision-making process, not a state an issue sits in. I'd cut this.

**No gaps in:**
- `type.*` — complete for our project
- `priority.*` — complete
- `triage.*` — complete, matches envoy's workflow perfectly
- `scope.*` — complete"

🧪 **Murat (TEA):** "From a testing and quality perspective:

1. **`status.changes-requested` is valuable for PR-shepherd.** When changes are requested, PR-shepherd should deprioritize rebasing that branch — no point keeping it up-to-date with main if the code needs rework first.

2. **I support cutting `process.party-mode`.** Bob is right — party mode is a process step, not an issue state. If we need to track 'this requires deliberation,' that's captured by `scope.needs-decision` already.

3. **Missing: a signal for 'CI is failing' that persists beyond the check run.** Currently if CI fails, the check mark turns red, but there's no label to query. For tracking patterns ('how many PRs had CI failures this sprint?'), a `status.ci-failing` label would be useful. But this might be over-engineering — we can query check runs directly.

4. **No gap in resolution labels.** Three is right: duplicate, wontfix, fixed. 'Invalid' could be a fourth, but we'd use `scope.out-of-scope` + `resolution.wontfix` for that."

### Round 1 Synthesis

**Agreed changes:**
- Add `status.do-not-merge` (Winston's point, Bob agrees, unanimous)
- Cut `process.party-mode` (Bob's point, Murat agrees, John agrees — "scope.needs-decision covers it")
- Keep `status.changes-requested` despite GitHub native state (Winston's ACI argument: agents query labels, not review states)

**Deferred to Round 2:**
- Whether `status.merge-ready` is needed (Bob's positive signal question)
- Whether `status.ci-failing` is needed (Murat's observability question)
- Agent label granularity challenge (Winston)

---

## ROUND 2: Consumer Challenge — Who Consumes Each Label? What Action Does It Trigger?

### The Challenge Protocol

Each label must survive this test:
1. **Who reads it?** (specific agent name or human role)
2. **What action does the reader take?** (specific behavior change)
3. **What happens if this label doesn't exist?** (is there an alternative signal?)

### type.* Labels

| Label | Reader | Action | Without It? |
|-------|--------|--------|-------------|
| `type.bug` | Envoy | Prioritizes triage; checks for reproduction steps | Issue body text serves same purpose — but slower to parse |
| `type.feature` | Envoy, supervisor | Routes to PM examination; checks ROADMAP alignment | Envoy reads issue body each time — label is a cache |
| `type.docs` | Envoy | Routes to tech writer workflow | Could infer from title/body — but unreliable |
| `type.question` | Envoy | May not need full triage; can often answer directly | Envoy must read full issue to decide — slower |
| `type.infra` | Envoy, merge-queue | Signals CI/tooling scope; merge-queue may apply different rules | Merge-queue can't distinguish infra PRs from feature PRs |
| `type.ux` | Envoy | Routes to UX review if available | Inferrable from content — marginal value |

📋 **John (PM):** "I challenge `type.ux`. We don't have a persistent UX agent. UX concerns are handled during party mode for any story. What action does this label trigger that `type.feature` doesn't?"

🏗️ **Winston (Architect):** "Fair challenge. The envoy's triage pipeline doesn't have a separate UX branch. `type.feature` covers it. I'd absorb `type.ux` into `type.feature`."

🏃 **Bob (SM):** "Agreed. Cut `type.ux`. If we ever get a persistent UX agent, we can add it back."

**Decision: Cut `type.ux`.** Type scope drops to 5 labels.

### priority.* Labels

| Label | Reader | Action | Without It? |
|-------|--------|--------|-------------|
| `priority.p0` | All agents | Immediate attention; supervisor dispatches worker ASAP | Could use issue title prefix — but less queryable |
| `priority.p1` | Supervisor, PR-shepherd | Address this sprint; PR-shepherd prioritizes rebase | Story files track priority — but not queryable via GitHub API |
| `priority.p2` | Supervisor | Backlog; no immediate action | Absence of P0/P1 implies P2+ — but explicit is better |
| `priority.p3` | — | Someday/maybe; essentially no active consumer | Could omit — P3 is "everything else" |

🧪 **Murat (TEA):** "I challenge `priority.p3`. What agent action does it trigger? None. It's the default state — anything without a priority label is implicitly P3."

📋 **John (PM):** "Counter-argument: explicit P3 communicates to reporters that we've SEEN their issue and assessed it as low priority. Without it, 'no priority label' is ambiguous — did we assess it or just not get to it?"

🏃 **Bob (SM):** "John's right. P3 is a communication signal to reporters, not an agent signal. The envoy uses it to say 'we heard you, this is just not urgent.' Keep it."

**Decision: Keep `priority.p3`.** It serves the reporter communication audience even though no agent acts on it directly.

### triage.* Labels

All four pass the consumer challenge. Envoy is the sole producer. Supervisor and human are consumers. Each maps to a specific step in the envoy's triage pipeline. No changes.

### scope.* Labels

All three pass. Merge-queue uses `scope.in-scope` as a merge gate. Envoy sets all three. Human resolves `scope.needs-decision`. No changes.

### status.* Labels

| Label | Reader | Action | Without It? |
|-------|--------|--------|-------------|
| `status.blocked` | PR-shepherd, supervisor | PR-shepherd skips rebase; supervisor investigates blocker | PR sits in queue getting rebased pointlessly |
| `status.in-review` | Human, merge-queue | Human knows to review; merge-queue knows PR is in review pipeline | GitHub's "review requested" partially covers this |
| `status.changes-requested` | PR-shepherd, workers | PR-shepherd deprioritizes rebase; worker knows to address feedback | GitHub review state covers this for humans — but not for agent queries |
| `status.approved` | Merge-queue | Merge-queue knows PR is approved and can proceed | GitHub review state — but requires separate API call |
| `status.stale` | Envoy, PR-shepherd, supervisor | Envoy posts update on issue; PR-shepherd may close PR; supervisor reviews | Staleness detection happens per-agent already — label is coordination signal |
| `status.needs-human` | Supervisor, human | Human knows action required; supervisor routes for attention | Could use comment/mention — but label is queryable |
| `status.do-not-merge` | Merge-queue | Merge-queue MUST NOT merge even if CI passes | Draft PR state — but not all DNM cases are drafts |

📋 **John (PM):** "I want to challenge `status.in-review` and `status.approved`. GitHub's review system already tracks these. The value is ONLY for agents that can't easily query review status. Is that actually true?"

🏗️ **Winston (Architect):** "Let me check. `gh pr view --json reviews` gives review status. `gh pr list --search 'review:approved'` filters by review state. So yes, agents CAN query review status without labels. The label adds no information that isn't already available via the API."

🏃 **Bob (SM):** "Then cut `status.in-review` and `status.approved`. They duplicate GitHub-native state. Our agents already use `gh pr` commands and can read review status directly."

🧪 **Murat (TEA):** "Agreed. Keep only the status labels that signal something GitHub doesn't track natively: blocked, changes-requested (debatable), stale, needs-human, do-not-merge."

📋 **John (PM):** "Actually, I'd cut `status.changes-requested` too. Same argument — `gh pr list --search 'review:changes_requested'` works. Three cuts from status."

**Decision: Cut `status.in-review`, `status.approved`, `status.changes-requested`.** Status scope drops to 4 labels: `blocked`, `stale`, `needs-human`, `do-not-merge`.

### agent.* Labels

| Label | Reader | Action | Without It? |
|-------|--------|--------|-------------|
| `agent.envoy` | Envoy, supervisor | Envoy filters its workload; supervisor knows who's handling | Could infer from triage.* labels — if triaging, envoy owns it |
| `agent.merge-queue` | Merge-queue, supervisor | — | Merge-queue doesn't need a label to know it owns merge |
| `agent.pr-shepherd` | PR-shepherd, supervisor | — | PR-shepherd already knows which PRs it manages |
| `agent.worker` | Supervisor | Supervisor knows a worker is assigned | Could track via story file status — but label is queryable |
| `agent.supervisor` | Supervisor, human | Signals supervisor is directly handling | Replaces `multiclaude` — more specific |

🏗️ **Winston (Architect):** "I'm going to challenge the whole `agent.*` scope. Each persistent agent already knows what it owns based on its role definition. Merge-queue owns all open approved PRs. PR-shepherd owns all PRs needing rebase. Envoy owns all open issues in triage. Workers own their assigned story. The label doesn't add information the agent doesn't already have.

The ONE case where agent labels add value: **human dashboard queries.** `label:agent.envoy` lets the human see 'what is envoy currently handling?' at a glance. That's a human-consumption use case, not an agent-consumption use case."

📋 **John (PM):** "That human-dashboard use case is real and valuable. The project owner needs to see workload distribution. But do we need all 5? I'd keep `agent.envoy` (most issues go through envoy) and `agent.worker` (shows a story is assigned). The other three agents' workloads are visible through other means:
- Merge-queue: `gh pr list --label status.approved` (if we kept it) or just `gh pr list --state open`
- PR-shepherd: PRs needing rebase are visible via `gh pr checks`
- Supervisor: the supervisor IS the human in most cases"

🏃 **Bob (SM):** "Compromise: keep `agent.envoy` and drop the rest. Envoy is the only agent where 'which issues is it handling?' isn't obvious from other signals. Workers operate through story files, not issues. Merge-queue and PR-shepherd operate on PRs, not issues."

🧪 **Murat (TEA):** "I support Bob's compromise. But I'd also keep `agent.worker` as a signal that 'someone is actively implementing this.' It answers a common question from reporters: 'is anyone working on this?'"

**Decision: Keep `agent.envoy` and `agent.worker`. Cut `agent.merge-queue`, `agent.pr-shepherd`, `agent.supervisor`.** Agent scope drops to 2 labels.

### contrib.* Labels

Both pass easily. `contrib.good-first-issue` and `contrib.help-wanted` are GitHub conventions that enable contributor discovery (GitHub surfaces them on the /contribute page). No changes.

### resolution.* Labels

| Label | Reader | Action | Without It? |
|-------|--------|--------|-------------|
| `resolution.duplicate` | Envoy | Flags for supervisor review before closing | Currently `duplicate` label — just renaming |
| `resolution.wontfix` | Envoy, supervisor | Communicates to reporter that issue won't be addressed | Currently implicit in closing + comment |
| `resolution.fixed` | Envoy | Cross-references with merged PR | GitHub's "Fixes #N" auto-close provides this |

📋 **John (PM):** "Challenge `resolution.fixed`. GitHub already closes issues automatically when a PR with 'Fixes #N' merges. The label is redundant with GitHub's native behavior."

🏗️ **Winston (Architect):** "Agreed. And for issues closed without 'Fixes #N' (e.g., envoy cross-check closures), the closing comment provides the context. The label adds nothing."

**Decision: Cut `resolution.fixed`.** Resolution scope drops to 2 labels.

### process.* Labels

`process.party-mode` was already cut in Round 1. Only `process.fast-track` remains.

📋 **John (PM):** "Does `process.fast-track` pass? The envoy uses it to skip the full triage pipeline for trivial fixes. That's a real workflow shortcut."

🏃 **Bob (SM):** "It passes. Fast-track means: skip party mode, create story directly, dispatch worker. Clear action trigger."

**Decision: Keep `process.fast-track`.** Process scope has 1 label.

### Round 2 — Status: `status.merge-ready` and `status.ci-failing`

🏃 **Bob (SM):** "Revisiting my Round 1 question about `status.merge-ready`. After the cuts to `status.in-review` and `status.approved`, merge-queue's positive signal for 'this is ready' is just: CI green + reviews approved + no `status.blocked` + no `status.do-not-merge`. That's a composite signal from multiple sources. A single `status.merge-ready` label could simplify merge-queue's logic."

🏗️ **Winston (Architect):** "But who SETS `status.merge-ready`? It would need to be derived from the composite check. That means either: (a) a bot/action that checks all conditions and applies the label, or (b) a human manually applying it. Both add complexity. I say let merge-queue evaluate the composite conditions directly — it already does."

**Decision: Don't add `status.merge-ready`.** Merge-queue evaluates conditions directly.

🧪 **Murat (TEA):** "And `status.ci-failing` — same conclusion. Agents can query `gh pr checks` directly. The label would need a bot to keep it in sync. Not worth the complexity."

**Decision: Don't add `status.ci-failing`.** Query check runs directly.

---

## ROUND 3: Final Consensus — Ordering, Colors, Descriptions, Migration Plan

### Final Taxonomy (Post-Challenges)

| Scope | Labels | Count |
|-------|--------|-------|
| type.* | bug, feature, docs, question, infra | 5 |
| priority.* | p0, p1, p2, p3 | 4 |
| triage.* | new, in-progress, needs-info, complete | 4 |
| scope.* | in-scope, out-of-scope, needs-decision | 3 |
| status.* | blocked, stale, needs-human, do-not-merge | 4 |
| agent.* | envoy, worker | 2 |
| contrib.* | good-first-issue, help-wanted | 2 |
| resolution.* | duplicate, wontfix | 2 |
| process.* | fast-track | 1 |
| **Total** | | **27** |

Started at 35, cut 8 labels that failed the consumer challenge. Net change from current 21: +6 genuinely new labels, 21 renamed/restructured.

### Color Assignments

📋 **John (PM):** "Colors should make scopes visually distinct in GitHub's label list. Each scope gets one base color with slight variations within."

🏗️ **Winston (Architect):** "I'd use a simple rule: each scope gets a single hex color. No variations within scope — that's over-design. The scope prefix already distinguishes labels within a group."

| Scope | Hex Color | Visual | Rationale |
|-------|-----------|--------|-----------|
| type.* | `#0075ca` | Blue | Neutral, informational |
| priority.p0 | `#B60205` | Dark red | Danger/urgent |
| priority.p1 | `#D93F0B` | Orange-red | Important |
| priority.p2 | `#FBCA04` | Yellow | Moderate |
| priority.p3 | `#D4C5F9` | Lavender | Low/calm |
| triage.* | `#1D76B5` | Steel blue | Process/workflow |
| scope.in-scope | `#0E8A16` | Green | Go |
| scope.out-of-scope | `#E11D48` | Red | Stop |
| scope.needs-decision | `#FEF2C0` | Cream/yellow | Caution |
| status.blocked | `#D93F0B` | Orange | Warning/attention |
| status.stale | `#BFD4F2` | Light blue | Faded/inactive |
| status.needs-human | `#d93f0b` | Orange-red | Urgent human attention |
| status.do-not-merge | `#B60205` | Dark red | Hard stop |
| agent.envoy | `#0E8A16` | Green | Active agent |
| agent.worker | `#0E8A16` | Green | Active agent |
| contrib.* | `#7057ff` | Purple | Community/inviting |
| resolution.* | `#cfd3d7` | Gray | Terminal/closed |
| process.fast-track | `#C5DEF5` | Light blue | Informational |

🏃 **Bob (SM):** "Priority labels get individual colors (heat map: red→yellow→lavender). Everything else gets one color per scope. Clean."

### Description Standards

🧪 **Murat (TEA):** "Every label description should be a single sentence answering: 'When should this label be applied?' Not 'what does this mean?' — 'when do I use it?'"

**Adopted description format:** Action-oriented, present tense, specifying the trigger condition.

### Final Label Specifications

| Label | Description | Color |
|-------|------------|-------|
| `type.bug` | Something isn't working as expected | `#0075ca` |
| `type.feature` | New feature or enhancement request | `#0075ca` |
| `type.docs` | Documentation improvement needed | `#0075ca` |
| `type.question` | Question about usage or behavior | `#0075ca` |
| `type.infra` | CI/CD, tooling, or build system change | `#0075ca` |
| `priority.p0` | Blocks users — immediate response required | `#B60205` |
| `priority.p1` | Important — address this sprint | `#D93F0B` |
| `priority.p2` | Backlog — address when capacity allows | `#FBCA04` |
| `priority.p3` | Low urgency — someday/maybe | `#D4C5F9` |
| `triage.new` | Issue acknowledged, entering triage | `#C2E0F4` |
| `triage.in-progress` | Envoy actively triaging this issue | `#6FBAED` |
| `triage.needs-info` | Waiting on reporter for clarification | `#BFD4F2` |
| `triage.complete` | Triage finished — story may exist | `#1D76B5` |
| `scope.in-scope` | Fits current roadmap | `#0E8A16` |
| `scope.out-of-scope` | Outside current project direction | `#E11D48` |
| `scope.needs-decision` | Requires human evaluation for scope | `#FEF2C0` |
| `status.blocked` | Blocked on dependency or decision | `#D93F0B` |
| `status.stale` | No activity past staleness threshold | `#BFD4F2` |
| `status.needs-human` | Blocked on human action or decision | `#d93f0b` |
| `status.do-not-merge` | Must not merge even if CI passes | `#B60205` |
| `agent.envoy` | Envoy is triaging or responsible | `#0E8A16` |
| `agent.worker` | Worker agent actively implementing | `#0E8A16` |
| `contrib.good-first-issue` | Good for newcomers to the project | `#7057ff` |
| `contrib.help-wanted` | Community contributions welcome | `#7057ff` |
| `resolution.duplicate` | Duplicate of an existing issue | `#cfd3d7` |
| `resolution.wontfix` | Will not be addressed — see comment for reason | `#cfd3d7` |
| `process.fast-track` | Trivial fix — skip full triage pipeline | `#C5DEF5` |

### Migration Plan Review

🏗️ **Winston (Architect):** "The migration needs to be atomic from the agents' perspective. Rename existing labels FIRST (GitHub preserves label-issue associations on rename), then create new labels. Order:

1. **Rename colon-separated labels** — `triage:new` → `triage.new`, etc. (11 renames)
2. **Rename unscoped labels** — `bug` → `type.bug`, etc. (10 renames)
3. **Create new labels** — `status.blocked`, `status.stale`, etc. (6 new)
4. **Update agent definitions** — envoy.md, merge-queue.md with new label names
5. **Delete obsolete labels** — only after confirming no references remain
6. **Verify** — `gh label list` shows exactly 27 labels"

📋 **John (PM):** "Critical: this is a research-and-recommendation PR. The actual migration should be a separate story. This PR delivers the research doc and party mode artifact only."

🏃 **Bob (SM):** "Agreed. Implementation story should be created after this PR is reviewed and approved. The story would include: label migration script, agent definition updates, and verification."

### Authority Updates

🏃 **Bob (SM):** "Documenting which agents can set which labels:

| Label Scope | Set By | Removed By |
|------------|--------|------------|
| type.* | Envoy (autonomous) | Envoy, supervisor |
| priority.* | Envoy (autonomous) | Supervisor (override) |
| triage.* | Envoy (autonomous) | Envoy |
| scope.* | Envoy (propose), supervisor (decide) | Supervisor |
| status.blocked | Any agent | Resolving agent |
| status.stale | Envoy, PR-shepherd | Any agent on activity |
| status.needs-human | Any agent | Human |
| status.do-not-merge | Supervisor, human | Supervisor, human |
| agent.envoy | Envoy | Envoy |
| agent.worker | Supervisor | Supervisor, worker |
| contrib.* | Envoy, supervisor | Envoy, supervisor |
| resolution.* | Envoy (propose), supervisor (confirm) | Supervisor |
| process.fast-track | Envoy | Envoy |

Note: Envoy can PROPOSE `scope.out-of-scope` for clearly misaligned requests per the envoy rules of behavior, but closure decisions require supervisor approval."

---

## Decisions Summary

| ID | Decision | Confidence | Rationale |
|----|----------|------------|-----------|
| SL-001 | Use `.` as label scope separator | High | Universal namespace separator; clean on GitHub; avoids `:` ambiguity |
| SL-002 | 9 scopes, 27 total labels | High | Each survived consumer challenge; 6 net new vs current 21 |
| SL-003 | Cut `type.ux` | High | No persistent UX agent; `type.feature` covers it |
| SL-004 | Keep `priority.p3` | High | Reporter communication signal even without agent consumer |
| SL-005 | Cut `status.in-review`, `status.approved`, `status.changes-requested` | High | GitHub review states queryable via API; labels duplicate native state |
| SL-006 | Cut `agent.merge-queue`, `agent.pr-shepherd`, `agent.supervisor` | High | Agent workloads visible through other signals; only envoy + worker need labels |
| SL-007 | Cut `resolution.fixed` | High | GitHub auto-close via "Fixes #N" and envoy comments cover this |
| SL-008 | Cut `process.party-mode` | High | `scope.needs-decision` covers the signal; party mode is process, not state |
| SL-009 | Add `status.do-not-merge` | High | Merge-queue needs hard stop signal beyond draft PRs |
| SL-010 | Don't add `status.merge-ready` or `status.ci-failing` | High | Merge-queue evaluates composite conditions; labels would need sync bot |
| SL-011 | Priority colors as heat map (red→lavender) | Medium | Intuitive urgency gradient |
| SL-012 | Migration via rename-first strategy | High | Preserves label-issue associations |
| SL-013 | Implementation as separate story (not this PR) | Unanimous | Research PR should not apply changes |

## Rejected Options

| Option | Reason for Rejection |
|--------|---------------------|
| `epic.N` per-epic labels | GitHub milestones serve this; 40+ labels is bloat |
| `sprint.*` labels | No fixed sprints in ThreeDoors |
| `effort.*` Fibonacci labels | Effort tracked in story files, not issues |
| `os.*` platform labels | Single-platform app |
| `version.*` labels | No versioned backporting |
| Agent labels for BMAD roles | BMAD agents are invoked, not persistent |
| `status.merge-ready` | Composite signal evaluated by merge-queue directly |
| `status.ci-failing` | Check runs queryable via API |
| `process.party-mode` | Covered by `scope.needs-decision` |
| `type.ux` | No persistent UX agent; covered by `type.feature` |
| `resolution.fixed` | GitHub auto-close covers this |
| Full 5-agent `agent.*` set | Only envoy and worker need visibility labels |
| `::` as separator (GitLab style) | Looks unusual on GitHub |
| `/` as separator | Conflicts with path references |

---

## Dissenting Opinions

### Winston on Agent Labels

🏗️ **Winston:** "I still think agent labels are primarily a human-consumption feature, not an agent-coordination mechanism. Our agents coordinate through story files, messages, and GitHub API queries — not labels. The `agent.envoy` and `agent.worker` labels survived because they answer a legitimate human question ('who is handling this?'), but agents should never rely on labels to know their own workload. If we catch ourselves writing agent code that queries `label:agent.envoy`, we've over-coupled to the label system."

### John on Label Count

📋 **John:** "27 labels feels high for a solo-dev project with <50 open issues at any time. But the research correctly identifies that labels serve MULTIPLE audiences (agents + human + public), and each label survived the consumer test. If we find ourselves not using certain labels after 3 months of operation, we should prune aggressively."

### Murat on Automation

🧪 **Murat:** "Without GitHub Actions to enforce scoped label mutual exclusivity, we're relying on convention. An issue could have both `priority.p0` and `priority.p2` applied simultaneously. The research mentions GitLab's native scoped label support — we lack this. Recommend a lightweight GitHub Action in a future story to enforce 'only one label per scope' when labels are applied. Not needed for this PR, but should be tracked."

---

*Party mode concluded after three rounds. All participants aligned on the final 27-label taxonomy. Research document and this deliberation artifact are the deliverables for the PR.*
