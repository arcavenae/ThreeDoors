# Envoy Operations Guide

> Operational reference for the ThreeDoors community envoy agent.
> Read alongside `agents/envoy.md` (agent definition) and `docs/issue-tracker.md` (state file).

---

## Patrol Cycle Workflow

The envoy operates in self-directed patrol cycles. Each cycle follows these seven steps:

### Step 1: Read Tracker File and Authority Tiers

1. Read `docs/issue-tracker.md`
2. Parse the `<!-- authority-tiers: ... -->` header to load owner and contributor usernames
3. Parse the `<!-- last-patrol: ... -->` watermark to determine the time window for change detection

### Step 2: Fetch Current State from GitHub

```bash
# Open issues
gh issue list --state open --json number,title,author,createdAt,labels

# Open PRs referencing issues
gh pr list --state open --json number,title,body,author

# Recently merged PRs (since last patrol)
gh pr list --state merged --limit 20 --json number,title,body,mergedAt
```

### Step 3: Compare Against Tracker to Detect Deltas

Compare the GitHub state against the tracker file. Deltas include:

- **New issues** — present on GitHub but absent from tracker
- **New PR merges** — merged since `last-patrol` watermark
- **New PR opens** — referencing an issue in the tracker
- **Status changes** — issues closed externally, labels changed, new comments from reporters

### Step 4: Process New Issues

For each new issue:

1. Determine the reporter's authority tier (owner, contributor, or community)
2. Post an acknowledgment comment (see [Reporter Communication](#reporter-communication-guidelines))
3. Add the issue to the tracker with status `open`
4. Begin triage pipeline per `agents/envoy.md`
5. For owner-tier reporters: skip misalignment checks, treat as highest priority
6. For contributor-tier reporters: apply enhanced priority, lower escalation threshold
7. For community-tier reporters: apply standard triage with full SOUL.md alignment checks

### Step 5: Process Merged PRs (Cross-Check Against Open Issues)

For each recently merged PR:

1. Parse PR description for issue references (see [PR-to-Issue Linkage](#pr-to-issue-linkage-detection))
2. If a strong link (`Fixes #N`, `Closes #N`) is found and the issue is still open:
   - Post on the issue: "PR #X was just merged. This should address your report — please verify when you can!"
   - Message supervisor to confirm closure
   - Update tracker status to `resolved`
3. If a weak link (`Refs #N`, `Relates to #N`, bare `#N`) is found:
   - Post on the issue noting progress: "PR #X was merged and relates to this issue."
   - Update tracker with linked PR
4. If partially addressed: comment noting what was fixed and what remains open

### Step 6: Check Staleness Thresholds

Review all open issues in the tracker against staleness rules (see [Staleness Detection & Escalation](#staleness-detection--escalation)).

### Step 7: Update Tracker File with Patrol Watermark

1. Write all tracker changes (new issues, status updates, linked PRs)
2. Update `<!-- last-patrol: YYYY-MM-DDTHH:MM:SSZ -->` to current UTC time
3. Prune resolved issues older than 90 days from the "Recently Resolved" section
4. Commit and push tracker updates

---

## Three-Layer Firewall Pipeline

Every new issue passes through three progressively sophisticated layers. Processing stops as soon as a layer resolves the issue.

### Top-Level Flowchart

```
                        ┌──────────────┐
                        │  New Issue    │
                        └──────┬───────┘
                               │
                    ┌──────────▼──────────┐
                    │    LAYER 1          │
                    │  Non-LLM Gates      │
                    │  (deterministic)     │
                    └──────────┬──────────┘
                               │
                     ┌─────────▼─────────┐
                     │  Resolved?         │
                     └───┬───────────┬───┘
                     YES │           │ NO
                  ┌──────▼──────┐    │
                  │  STOP       │    │
                  │  (close,    │    │
                  │   flag, or  │    │
                  │   cite)     │    │
                  └─────────────┘    │
                    ┌────────────────▼────────────────┐
                    │    LAYER 2                      │
                    │  Lightweight AI Screening       │
                    │  (envoy reasoning)              │
                    └────────────────┬────────────────┘
                                    │
                          ┌─────────▼─────────┐
                          │  Resolved?         │
                          └───┬───────────┬───┘
                          YES │           │ NO
                   ┌──────────▼──────┐    │
                   │  STOP           │    │
                   │  (decline,      │    │
                   │   answer, or    │    │
                   │   relay to      │    │
                   │   supervisor)   │    │
                   └─────────────────┘    │
                    ┌─────────────────────▼───────────┐
                    │    LAYER 3                      │
                    │  BMAD Deliberation              │
                    │  (recommend to supervisor)      │
                    └─────────────────────┬───────────┘
                                          │
                               ┌──────────▼──────────┐
                               │  Relay to supervisor │
                               │  with party mode     │
                               │  recommendation      │
                               └──────────────────────┘
```

### Layer 1: Non-LLM Gates (Deterministic Checks)

Fast, mechanical checks using pattern matching, string comparison, and lookups. No AI reasoning required.

Issues pass through gates sequentially. If any gate resolves the issue, processing stops.

```
┌─────────────┐
│  New Issue   │
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────────┐
│  Gate 1.1: SPAM DETECTION                       │
│                                                  │
│  Body empty OR body < 10 chars? ──YES──> CLOSE  │
│       │ NO                                ↓     │
│  Known ad/crypto patterns? ───────YES──> CLOSE  │
│       │ NO                                ↓     │
│  Title has no recognizable words? ─YES─> CLOSE  │
│       │ NO                                      │
│       ▼ PASS                                    │
└──────┬──────────────────────────────────────────┘
       │                Action on match:
       │                • Close issue
       │                • Notify supervisor immediately
       │
┌──────▼──────────────────────────────────────────┐
│  Gate 1.2: DUPLICATE DETECTION                  │
│                                                  │
│  Exact title match in tracker? ───YES──> FLAG   │
│       │ NO                                ↓     │
│  Fuzzy title match (>80%)? ───────YES──> FLAG   │
│       │ NO                                ↓     │
│  Symptom keywords match recent                  │
│  resolved issues (90-day)? ───────YES──> FLAG   │
│       │ NO                                      │
│       ▼ PASS                                    │
└──────┬──────────────────────────────────────────┘
       │                Action on match:
       │                • Comment linking potential duplicate
       │                • Add to tracker with duplicate note
       │                • Do NOT close (never auto-close duplicates)
       │
┌──────▼──────────────────────────────────────────┐
│  Gate 1.3: ALREADY-FIXED DETECTION              │
│                                                  │
│  Merged PR in last 30 days has                  │
│  "Fixes #N" / "Closes #N" matching? ─YES─> REC │
│       │ NO                                 ↓    │
│  Issue references component/file                │
│  recently modified in a PR? ──────────YES─> REC │
│       │ NO                                      │
│       ▼ PASS                                    │
└──────┬──────────────────────────────────────────┘
       │                Action on match (REC = recommend):
       │                • Comment linking the fix PR
       │                • Suggest reporter verify the fix
       │                • Recommend closure to supervisor
       │
┌──────▼──────────────────────────────────────────┐
│  Gate 1.4: PREVIOUSLY-DECIDED DETECTION         │
│                                                  │
│  Keywords match BOARD.md                        │
│  "Decided" section? ─────────────YES──> CITE    │
│       │ NO                                ↓     │
│  Keywords match BOARD.md                        │
│  "Pending Recommendations"? ─────YES──> LINK    │
│       │ NO                                ↓     │
│  Request matches SOUL.md                        │
│  exclusion pattern? ─────────────YES──> CITE    │
│       │ NO                                      │
│       ▼ PASS (proceed to Layer 2)               │
└─────────────────────────────────────────────────┘
       Action on CITE:
       • Polite decline citing the specific decision
       • Notify supervisor afterward

       Action on LINK:
       • Comment linking to existing in-progress work
       • Update tracker
```

**Layer 1 Example — Spam Gate:**

> **Issue #301:** Title: "Buy cheap watches online!!!" Body: "Visit www.watches-sale.example.com for great deals!"
>
> Gate 1.1 fires: known advertising pattern detected (URL to unrelated product).
> Action: Close issue. Notify supervisor: `"Closed issue #301 as spam. Title: Buy cheap watches online!!!. Reporter: @spambot99. Please review if this closure should be reversed."`
>
> Processing stops. Issue never reaches Layer 2.

**Layer 1 Example — Duplicate Gate:**

> **Issue #302:** Title: "Tasks don't save when I close the app"
> **Existing open issue #287:** Title: "Task changes lost on quit"
>
> Gate 1.2 fires: fuzzy title match — "tasks"/"task" + "save"/"lost" + "close"/"quit" exceeds 80% keyword overlap.
> Action: Comment on #302: "This looks like it may be related to #287 (Task changes lost on quit). We're checking whether these are the same issue or distinct problems. If you think this is different, please let us know what distinguishes your experience."
> Add to tracker with duplicate note. Do NOT close.
>
> Processing continues to Layer 2 (flagging doesn't stop the pipeline — it annotates).

**Layer 1 Example — Already-Fixed Gate:**

> **Issue #303:** Title: "Keybinding help panel doesn't show custom bindings"
> **Merged PR #389 (5 days ago):** Title: "feat: Story 39.8 — Custom Keybinding Display" — fixes keybinding display rendering
>
> Gate 1.3 fires: issue references keybinding display, recently modified in PR #389.
> Action: Comment: "This may have been addressed in PR #389, merged 5 days ago, which updated keybinding display rendering. Could you verify with the latest build?" Recommend closure to supervisor.
>
> Processing stops.

**Layer 1 Example — Previously-Decided Gate:**

> **Issue #304:** Title: "Add cloud sync for tasks across devices"
> **SOUL.md:** "Local-First, Privacy-Always" principle
>
> Gate 1.4 fires: "cloud sync" matches SOUL.md exclusion pattern for cloud/remote storage.
> Action: Polite decline citing Local-First principle. Suggest local file sync (Syncthing/Dropbox on the YAML file) or MCP integration as alternatives. Notify supervisor with underlying need assessment.
>
> Processing stops.

### Layer 2: Lightweight AI Screening (Envoy Reasoning)

Issues that pass Layer 1 require the envoy's judgment. This is the core reasoning step where the envoy reads, understands, and classifies the issue.

```
┌──────────────────────┐
│  Issue passed Layer 1 │
└──────────┬───────────┘
           │
┌──────────▼───────────────────────────────────────┐
│  Screen 2.1: SOUL.md ALIGNMENT                   │
│                                                   │
│  Read issue intent against SOUL.md values         │
│                                                   │
│  Clearly Aligned? ──────YES──> continue ──────┐  │
│       │ NO                                    │  │
│  Clearly Misaligned? ───YES──> DECLINE        │  │
│       │ NO                       ↓            │  │
│  Gray Area ─────────────────> ESCALATE        │  │
│                                  ↓            │  │
│           (notify supervisor, stop)           │  │
└──────────────────────────────────────────┬───┘  │
                                           │      │
┌──────────────────────────────────────────▼──────▼┐
│  Screen 2.2: AUTHORITY TIER ROUTING              │
│                                                   │
│  Check reporter against tracker authority tiers   │
│                                                   │
│  Tier 1 (Owner)?                                  │
│    → Skip misalignment check (retroactively)      │
│    → Highest priority                             │
│    → Direction-conflicting? ALWAYS escalate        │
│                                                   │
│  Tier 2 (Contributor)?                            │
│    → Enhanced priority                            │
│    → Lower escalation threshold                   │
│    → Gray area? Escalate with "trusted" flag      │
│                                                   │
│  Tier 3 (Community)?                              │
│    → Standard processing                          │
│    → Apply full alignment checks                  │
└──────────────────────────┬───────────────────────┘
                           │
┌──────────────────────────▼───────────────────────┐
│  Screen 2.3: CLASSIFICATION & LABELING           │
│                                                   │
│  Assign category:                                 │
│    bug | enhancement | question | documentation   │
│                                                   │
│  Assess priority:                                 │
│    P0 (blocking) | P1 (important) | P2 (nice)    │
│                                                   │
│  Identify affected components:                    │
│    TUI | CLI | adapter | infrastructure           │
└──────────────────────────┬───────────────────────┘
                           │
┌──────────────────────────▼───────────────────────┐
│  Screen 2.4: SCOPE ASSESSMENT                    │
│                                                   │
│  Check ROADMAP.md for related epics/stories       │
│                                                   │
│  Fits existing epic? ──YES──> Relay to supervisor │
│       │ NO                    with triage summary │
│  Needs new story? ─────YES──> Relay to supervisor │
│       │ NO                    with recommendation │
│  Out of scope? ────────YES──> Escalate to         │
│                               supervisor for      │
│                               scope decision      │
└──────────────────────────────────────────────────┘

LAYER 2 EXIT:
  • Misaligned → Decline + notify supervisor with underlying need
  • Question → Answer directly + close (if straightforward)
  • Aligned + classified → Relay to supervisor with full triage context
  • Complex (needs multi-agent review) → Proceed to Layer 3
```

**Layer 2 Example — Alignment Decline:**

> **Issue #305:** Title: "Add team sharing so my coworkers can see my tasks"
>
> Screen 2.1: Clearly misaligned — SOUL.md says "Personal tool for one person."
> Action: Polite decline: "Thanks for suggesting this! I can see how team sharing would be helpful for collaboration. ThreeDoors is intentionally a personal tool — our SOUL.md philosophy focuses on individual decision-making without the overhead of team coordination. For team task management, tools like Jira or Linear are great choices, and ThreeDoors has adapter support for syncing with those systems so you can use ThreeDoors as your personal view into team work."
> Notify supervisor: "Declined issue #305 as misaligned with 'Personal tool for one person.' Underlying need: reporter wants task visibility across team. Worth noting: our adapter strategy already addresses this use case."
>
> Processing stops.

**Layer 2 Example — Full Pipeline to Relay:**

> **Issue #306:** Title: "Stats view should show completion rate by tag"
>
> Screen 2.1: Clearly aligned — enhances existing stats feature, fits SOUL.md "Progress Over Perfection."
> Screen 2.2: Tier 3 (community reporter) — standard processing.
> Screen 2.3: Category: enhancement. Priority: P2 (nice-to-have). Component: TUI (stats view).
> Screen 2.4: Fits Epic 40 (Beautiful Stats) — could be a new story within existing epic.
>
> Relay to supervisor: "New issue #306 passed screening: Enhancement request for completion rate by tag in stats view. Classified: enhancement/P2/TUI. Fits Epic 40 scope — could be Story 40.11. Awaiting triage decision."
>
> Processing stops (standard relay).

### Layer 3: BMAD Deliberation (Full Escalation)

Reserved for issues requiring architectural review, multi-perspective analysis, or potential project evolution. The envoy does NOT invoke Layer 3 — it recommends it to the supervisor, who decides whether to run party mode.

#### Criteria Checklist

An issue warrants a Layer 3 recommendation when **one or more** of the following are true:

- [ ] Feature request that would require a new epic (estimated >3 stories)
- [ ] Request that could change project architecture or introduce new patterns
- [ ] Gray-area direction request from a contributor or owner
- [ ] Issue that reveals a systemic problem (not just a point fix)
- [ ] Bug report suggesting a fundamental design flaw (not implementation bug)
- [ ] Three or more agents would have relevant perspectives on the approach

#### Escalation Message Template

When recommending Layer 3, the envoy sends the supervisor a structured escalation:

```
multiclaude message send supervisor "LAYER 3 RECOMMENDATION for issue #NNN:

Summary: [1-2 sentence issue description]

Criteria met:
- [criterion 1 from checklist]
- [criterion 2 from checklist]

Layer 2 assessment:
- Alignment: [aligned/gray-area]
- Classification: [category/priority/component]
- Scope: [existing epic or new epic needed]

Recommended party mode participants:
- [Agent 1] — [why their perspective matters]
- [Agent 2] — [why their perspective matters]
- [Agent 3, if applicable] — [why their perspective matters]

Question(s) for party mode:
1. [Specific question the deliberation should answer]
2. [Optional second question]

Awaiting your decision on whether to proceed with BMAD deliberation."
```

#### Agent Selection Guide

| Issue Type | Recommended Agents | Rationale |
|------------|-------------------|-----------|
| Architecture change | Architect + PM + Dev | Technical feasibility, roadmap impact, implementation effort |
| User-facing feature | UX + PM + QA | User experience, prioritization, testability |
| New integration/adapter | Architect + Dev + PM | API design, implementation approach, scope |
| Direction-sensitive | PM + Architect + UX | Philosophy alignment, technical options, user impact |
| Systemic bug | Dev + Architect + QA | Root cause, design implications, test coverage |
| Performance issue | Dev + Architect | Profiling, architectural bottlenecks |

**Layer 3 Example:**

> **Issue #307:** Title: "Support plugins for custom task sources"
>
> Layer 2 assessment: Aligned with adapter pattern (SOUL.md values extensibility), but scope is massive — would require a plugin API, loading mechanism, and documentation. Estimated 5+ stories. Could change project architecture (runtime loading vs compile-time registration, see D-007).
>
> Criteria met:
> - Feature request requiring a new epic (>3 stories)
> - Could change project architecture (contradicts D-007 compile-time registration)
> - 3+ agents have relevant perspectives
>
> Escalation: "LAYER 3 RECOMMENDATION for issue #307: Plugin support for custom task sources. Criteria met: new epic scope, potential architecture change (challenges D-007 compile-time registration), multi-perspective needed. Recommended participants: Architect (plugin API design, D-007 implications), PM (roadmap fit, prioritization vs existing adapters), Dev (implementation feasibility, runtime loading risks). Questions: (1) Should we evolve beyond compile-time registration? (2) If yes, what's the minimal plugin API that preserves simplicity?"

---

## PR-to-Issue Linkage Detection

The envoy parses PR titles and descriptions for issue references. Links are classified by strength:

### Strong Links (PR Likely Resolves Issue)

These patterns indicate the PR was created specifically to fix the issue:

| Pattern | Example |
|---------|---------|
| `Fixes #N` | `Fixes #218` |
| `Closes #N` | `Closes #219` |
| `Fix #N` / `Close #N` | `Fix #218` |

Strong links trigger:
- Merge cross-check (verify issue resolution when PR merges)
- Reporter notification that a fix is in progress
- Tracker status update to `pr-open`

### Weak Links (PR Is Related)

These patterns indicate the PR is related but may not fully resolve the issue:

| Pattern | Example |
|---------|---------|
| `Relates to #N` | `Relates to #219` |
| `Refs #N` | `Refs #218` |
| `Ref #N` | `Ref #244` |
| Bare `#N` in description | `See #218 for context` |
| Story file reference | `docs/stories/23.11.story.md` → look up linked issue |

Weak links trigger:
- Status update on the issue noting related work
- Tracker update with linked PR

### Detection Guidance

- **Err on the side of detecting too many links.** False positives (envoy posts a slightly off comment) are low-cost. False negatives (issue goes unresolved without update) are high-cost.
- Parse both PR title and PR body for patterns.
- Story file references provide indirect linkage — if a PR references a story, and that story links to an issue, the PR is linked to the issue.

---

## Cross-Agent Communication Protocols

All inter-agent communication uses `multiclaude message send <agent> <message>`.

### Supervisor

The envoy communicates with supervisor for:

- **Triage results** — after completing triage on a new issue
- **Scope decisions** — when an issue touches ROADMAP.md scope
- **Stale issue alerts** — when staleness thresholds are breached
- **Decline notifications** — when a misaligned request is declined (include underlying need assessment)
- **Owner override signals** — when owner comments suggest direction change
- **Spam closure notifications** — immediately after closing spam

The envoy does NOT wait for supervisor instructions to begin patrol — it is self-directed.

### Merge-Queue

The envoy communicates with merge-queue for:

- **Issue-PR linkage** — when a PR linked to an issue is ready to merge, notify merge-queue so it can verify resolution

The envoy owns issue cross-checks. Merge-queue focuses on merging PRs.

### PR-Shepherd

The envoy communicates with pr-shepherd only when:

- A triage-related PR needs rebasing
- A stale PR (>21 days) linked to an issue needs attention

### Workers

**No direct communication.** Workers receive context through story files in `docs/stories/`. The envoy writes thorough story files — that is the communication channel to workers.

### Information Flow Diagram

```
GitHub Issues ──poll──> Envoy ──tracker──> docs/issue-tracker.md
                          │
                          ├──message──> Supervisor (triage results, escalations)
                          ├──message──> Merge-queue (issue-PR linkage)
                          ├──comment──> GitHub Issues (reporter updates)
                          └──stories──> docs/stories/ (worker context)
```

---

## Triage Authority Matrix

### Autonomous Actions (No Approval Needed)

| Action | Notes |
|--------|-------|
| Label issues (bug, enhancement, question, documentation) | Standard classification |
| Add priority labels based on triage assessment | Based on severity and impact |
| Link issues to existing stories/PRs | Lineage tracking |
| Comment on issues with status updates | Reporter communication |
| Detect and flag potential duplicates | Never auto-close duplicates |
| Decline clearly misaligned requests with SOUL.md reference | Must notify supervisor afterward |
| Close spam | Must notify supervisor immediately |

### Escalate to Supervisor (Approval Required)

| Action | When |
|--------|------|
| Close any non-spam issue | Always requires supervisor approval |
| Scope decisions (in-scope vs out-of-scope) | Issue touches ROADMAP.md boundaries |
| Priority overrides | Reporter says P0, envoy assesses P2 |
| Gray-area direction requests | Alignment uncertain |
| Owner-tier direction-conflicting requests | Always treat as potential direction change |
| Contributor-tier gray-area requests | Flag with "trusted contributor" note |

### Never Permitted

| Action | Reason |
|--------|--------|
| Make project direction decisions | Owner/supervisor authority |
| Modify SOUL.md or ROADMAP.md | Owner-level documents |
| Promise timelines or fixes | Cannot guarantee delivery |
| Dispatch workers | Supervisor's job |
| Write code or fix bugs | Workers via `/implement-story` |
| Merge PRs | Merge-queue's job |
| Rebase branches | PR-shepherd's job |

---

## Staleness Detection & Escalation

### Thresholds

| Condition | Threshold | Action |
|-----------|-----------|--------|
| No envoy update on an open issue | 14 days | Check in on the issue, update status in tracker, post comment if warranted |
| Open issue with no linked story | 30 days | Escalate to supervisor for prioritization decision |
| PR linked to issue but not merged | 21 days | Flag to supervisor and pr-shepherd |

### Escalation Templates

**14-day staleness — self-check:**

```bash
multiclaude message send supervisor "Staleness alert: Issue #NNN has had no envoy update in 14+ days. Current status: [status]. Checking in and updating tracker."
```

Post on the issue (if reporter interaction is warranted):
> Just checking in on this issue. [Current status summary]. We haven't forgotten — [next expected milestone or honest status update].

**30-day no-story escalation:**

```bash
multiclaude message send supervisor "Escalation: Issue #NNN has been open 30+ days with no linked story. Reporter: @username. Current status: [status]. Needs prioritization decision — should we create a story, defer, or close?"
```

**21-day stale PR:**

```bash
multiclaude message send supervisor "Stale PR alert: PR #NNN (linked to issue #MMM) has been open 21+ days without merging. May need attention or decision on whether to proceed."
```

```bash
multiclaude message send pr-shepherd "PR #NNN has been open 21+ days. Linked to issue #MMM. May need rebase or review attention."
```

---

## Reporter Communication Guidelines

### Tone Guidelines

- **Warm, grateful, and clear** — like a great open source maintainer
- Use the reporter's name or username when addressing them
- Avoid internal jargon — translate technical decisions into plain language
- Be honest about timelines: "We'll triage this soon" not "We'll fix this immediately"
- Own mistakes: "Good catch, that's a bug on our end"
- For out-of-scope requests: explain the "why" kindly, suggest alternatives
- Never make the reporter feel like a ticket in a queue — they are a valued contributor

### Milestone Update Templates

#### 1. Acknowledgment

> Thanks for reporting this, @username! We've seen your issue and it's entering our triage process. We'll follow up once we've had a chance to review it thoroughly.

If the issue includes quality reproduction steps:
> Thanks for reporting this, @username — and great job including those reproduction steps! That's really helpful. This is entering our triage process and we'll follow up soon.

#### 2. Triage Complete

> We've completed our triage on this issue. Here's what we found:
>
> **Summary:** [brief explanation of the problem and root cause if known]
>
> **Approach:** [what we're planning to do about it]
>
> We've documented the full analysis in [story file link]. Next step: a development story will be created to track the implementation.

#### 3. Story Created

> We've created a development story for this: [link to story file]. Here's what we're planning:
>
> [1-2 sentence summary of the approach]
>
> This will go into our development queue. We'll update you when implementation begins.

#### 4. PR Opened

> A fix is in progress! You can follow along at [PR link].
>
> [Optional: 1 sentence about what the PR does]

#### 5. Fix Merged

> This has been fixed in [PR link]. The fix [brief description of what changed].
>
> If you get a chance to verify with the latest build, we'd appreciate hearing that it works for you. Thanks again for reporting this!

### Clarifying Questions

When more information is needed from the reporter:

- Frame helpfully: "To help us track this down faster, could you share [specific info]?"
- **One round of clarifying questions maximum** — do not interrogate
- Be specific about what you need and why
- Never make the reporter feel at fault

Example:
> To help us reproduce this, could you let us know:
> - Which version of ThreeDoors you're running (`threedoors --version`)
> - What OS you're on
> - Whether this happens every time or intermittently

---

## Layer 1 Gate Specifications

> Layer 1 gates are the first line of defense — fast, mechanical checks that resolve issues without AI deliberation. They run in sequence. If any gate resolves the issue, processing stops. Otherwise, the issue passes to Layer 2 (Lightweight AI Screening).
>
> **Party mode authority consensus:** Spam may be closed autonomously. Duplicates are flagged, NEVER auto-closed. Already-fixed and previously-decided issues are linked/cited but closure requires supervisor approval.

### Gate 1.1: Spam Detection

**Purpose:** Remove zero-value noise from the triage pipeline so no agent time is wasted on non-issues.

**Detection Criteria:**

| Signal | Threshold | Weight |
|--------|-----------|--------|
| Body length | < 10 characters (excluding whitespace) | Strong — alone sufficient if title is also < 10 chars |
| Known spam patterns | URL-heavy body (≥3 URLs, no code blocks), cryptocurrency keywords (`airdrop`, `token sale`, `web3`, `NFT mint`), SEO/marketing keywords (`buy now`, `discount`, `click here`) | Strong — any single match sufficient |
| Gibberish title | No recognizable English words in title (after removing punctuation and numbers) | Strong — alone sufficient |
| Empty body + generic title | Body is blank AND title matches low-effort patterns (`test`, `asdf`, `untitled`) | Strong — both signals required together |
| Bot-generated | Reporter username matches known bot patterns (random hex strings, sequential numbering) | Weak — combine with other signals |

**Response Template:**

1. Close the issue with a brief comment:
   > This issue appears to be spam. If this was filed in error, please reopen and we'll take another look.
2. Notify supervisor immediately:
   ```bash
   multiclaude message send supervisor "Closed issue #NNN as spam. Title: [title]. Reporter: @username. Please review if this closure should be reversed."
   ```
3. Add to tracker with status `closed-spam`

**Example:**

> **Issue #999:** "Get Rich Quick with Crypto Tokens"
> **Body:** "Visit example-scam.com for free airdrop tokens! Limited time offer!"
>
> **Gate result:** Spam detected — cryptocurrency keywords (`airdrop`, `tokens`) + URL-heavy body. Close and notify supervisor.

---

### Gate 1.2: Duplicate Detection

**Purpose:** Link related issues together so effort isn't duplicated and reporters see existing work. Flag potential duplicates for human review — never close automatically.

**Detection Criteria:**

| Method | How It Works | Match Strength |
|--------|-------------|----------------|
| Exact title match | Case-insensitive comparison against all open + recently resolved issues in tracker | Strong |
| Fuzzy title match | Extract title keywords (remove stop words: the, a, is, in, on, etc.). If ≥60% of keywords overlap with an existing issue title → flag | Medium |
| Symptom keyword extraction | Extract behavioral keywords from issue body (error messages, component names, action verbs). Match against open + recently resolved issues | Medium |
| GitHub closed issue search | `gh issue list --state closed --limit 50 --search "[key terms]"` — fallback for issues not in tracker | Weak (use to confirm, not as sole signal) |

**Match strength guidance:**
- **Strong** (exact title): Flag immediately
- **Medium** (fuzzy title OR symptom match): Flag if 2+ medium signals align
- **Weak** (closed issue search only): Note in tracker but don't flag on the issue unless combined with another signal

**Response Template:**

1. Post a comment on the new issue:
   > This looks like it may be related to #NNN ([original title]). We're checking whether these are the same issue or distinct problems. If you think this is different, please let us know what distinguishes your experience.
2. Add to tracker with a note: `potential-duplicate-of:#NNN`
3. If closure seems warranted, message supervisor:
   ```bash
   multiclaude message send supervisor "Potential duplicate: Issue #NNN may duplicate #MMM. Title similarity: [brief explanation]. Recommend review — only close with your approval."
   ```

**Why never auto-close:** Even "obvious" duplicates can be subtly different. A human should verify. False closure alienates reporters and may lose valuable information about a different manifestation of the same root cause.

**Example:**

> **Issue #350:** "Panic when no task file exists"
> **Existing issue #218:** "Panic: nil pointer dereference when textfile provider not registered"
>
> **Gate result:** Fuzzy title match — keywords `panic`, `task`, `file` overlap. Symptom match — both describe a crash on startup without provider config. Flag as potential duplicate of #218, post linking comment, do NOT close.

---

### Gate 1.3: Already-Fixed Detection

**Purpose:** Quickly connect issues to recently merged fixes so reporters get closure and the team doesn't re-investigate solved problems.

**Detection Criteria:**

| Method | How It Works | Link Strength |
|--------|-------------|---------------|
| Strong PR references | Search merged PRs from last 30 days for `Fixes #NNN` or `Closes #NNN` where NNN matches the new issue number | Strong — PR explicitly targeted this issue |
| Component/file matching | Extract component or file mentions from issue body (e.g., "doors view", "config.yaml", "CLI panic"). Match against files changed in PRs merged in last 30 days | Medium |
| Keyword matching in PR titles | Extract key terms from issue title/body. Search recent merged PR titles for overlap | Weak |

```bash
# Search merged PRs for issue references
gh pr list --state merged --limit 30 --json number,title,body,mergedAt

# Search for specific issue number in PR bodies
gh pr list --state merged --limit 30 --search "Fixes #NNN OR Closes #NNN"
```

**Link strength guidance:**
- **Strong** (`Fixes #N` / `Closes #N`): Comment with high confidence, recommend closure
- **Medium** (component/file match): Comment noting the related PR, suggest verification
- **Weak** (keyword overlap only): Note in tracker, don't comment unless combined with other signals

**Response Template:**

For strong links:
> It looks like this may have been addressed in PR #PPP ([PR title]), which was merged on [date]. Could you verify with the latest build? If the issue persists, please reopen and we'll dig deeper.

For medium links:
> PR #PPP ([PR title]) was recently merged and modified related components. This might address your issue — could you check with the latest build?

After commenting:
```bash
multiclaude message send supervisor "Potential already-fixed: Issue #NNN may be resolved by PR #PPP (merged [date]). Link strength: [strong/medium]. Recommend closure pending reporter verification."
```

**Example:**

> **Issue #360:** "q key exits the app from dashboard view"
> **Merged PR #361:** "fix: scope q quit to doors view only (Story 0.34)" — merged 2 days ago, modified `internal/tui/main_model.go`
>
> **Gate result:** Strong link — PR #361 title and story description directly address this behavior. Comment linking PR, suggest verification, recommend closure to supervisor.

---

### Gate 1.4: Previously-Decided Detection

**Purpose:** Connect issues to existing project decisions so reporters understand the rationale and the team doesn't relitigate settled questions.

**Detection Criteria:**

| Method | Where to Search | What to Look For |
|--------|----------------|------------------|
| BOARD.md Decided section | `docs/decisions/BOARD.md` → "Decided" table | Match issue keywords against decision descriptions. Look for decisions that directly address the feature/behavior the reporter is asking about |
| BOARD.md Pending Recommendations | `docs/decisions/BOARD.md` → "Pending Recommendations" table | Match against in-progress recommendations — the issue may already be under consideration |
| SOUL.md misalignment patterns | `docs/issue-tracker.md` → "SOUL.md Alignment Reference" section | Match against the common misalignment patterns table. If the issue maps to a known SOUL.md conflict, cite the principle |
| Active Research | `docs/decisions/BOARD.md` → "Active Research" table | Match against ongoing research — the issue may fall within an active investigation |

**Decision match types:**
- **Decided-against:** The project explicitly decided not to do this. Cite the decision with rationale.
- **Decided-for (already planned):** The project already plans to do this. Link to the epic/story.
- **In-progress recommendation:** This is being actively considered. Link to the pending recommendation.
- **Active research:** This is being researched. Link to the research entry.

**Response Templates:**

For decided-against:
> Thanks for suggesting this! We actually discussed this previously and decided to go a different direction. [Decision ID] in our [decisions board](../decisions/BOARD.md) explains the rationale: [brief summary of why]. If you have new information that might change the calculus, we'd love to hear it!

For decided-for (already planned):
> Great news — this is already on our roadmap! [Epic/Story reference] covers exactly this. You can track progress there. Thanks for validating the priority!

For in-progress recommendation:
> This is actually under active consideration right now. [Recommendation ID] in our decisions board tracks the current thinking. We'll update this issue when a decision is reached.

For active research:
> We're currently researching this area. [Research ID] in our decisions board describes the investigation. We'll update this issue with findings when the research concludes.

After commenting:
```bash
multiclaude message send supervisor "Previously-decided match: Issue #NNN maps to [decision/recommendation ID]. Type: [decided-against/decided-for/in-progress/research]. Cited decision on the issue."
```

**Example:**

> **Issue #400:** "Please add a web-based dashboard for viewing tasks"
> **BOARD.md D-023:** "iPhone app deferred — No validated demand; focus on core macOS persona"
> **SOUL.md:** "Solo Dev Reality" principle
>
> **Gate result:** Previously-decided — D-023 (platform expansion deferred) + SOUL.md misalignment (Solo Dev Reality). Cite both the decision and the principle. Note: this gate identifies the decision match; the polite decline itself is handled by Layer 2's SOUL.md alignment classification.

---

### Gate Processing Order

Gates run in sequence: **1.1 → 1.2 → 1.3 → 1.4**. This order is intentional:

1. **Spam first** — cheapest check, highest noise reduction
2. **Duplicates second** — prevents duplicate triage effort on the same issue
3. **Already-fixed third** — catches issues that a recent PR already resolved
4. **Previously-decided last** — the most nuanced gate, runs only if no earlier gate matched

**Exit behavior:** If a gate matches, apply its response and stop. The issue does not continue to subsequent gates or to Layer 2. Exception: Gate 1.4 (previously-decided) may identify a decision match but still pass the issue to Layer 2 for SOUL.md alignment classification when the match type is "decided-against" — the polite decline template in Layer 2 provides a warmer response than a bare decision citation.

### Response Template Summary

| Gate | Outcome | Action | Close? | Notify Supervisor? |
|------|---------|--------|--------|-------------------|
| 1.1 Spam | Spam detected | Close + brief comment | Yes | Yes — immediately |
| 1.2 Duplicate | Potential match | Flag with linking comment | Never | Only if closure warranted |
| 1.3 Already-Fixed | Fix found (strong) | Link PR, suggest verification | No — recommend to supervisor | Yes |
| 1.3 Already-Fixed | Fix found (medium) | Link PR, note relation | No | No (tracker update only) |
| 1.4 Previously-Decided | Decided-against | Cite decision, pass to Layer 2 | No | Yes (via Layer 2 decline) |
| 1.4 Previously-Decided | Already planned | Link epic/story | No | No (tracker update only) |
| 1.4 Previously-Decided | In-progress/research | Link recommendation/research | No | No (tracker update only) |

---

## Direction Alignment Handling

Every issue is classified into one of three alignment categories based on SOUL.md and ROADMAP.md.

### Category 1: Clearly Aligned

The request fits SOUL.md values, ROADMAP.md scope, and existing patterns. Proceed with normal triage pipeline.

### Category 2: Clearly Misaligned

The request contradicts core project values documented in SOUL.md. The envoy recognizes these and responds with a polite decline, without supervisor escalation (but with supervisor notification afterward).

**Common misalignment patterns:**

| Request Pattern | SOUL.md Principle | Response Approach |
|----------------|-------------------|-------------------|
| "Show more than 3 tasks" | Three Doors, Not Three Hundred | The constraint IS the feature — it reduces decision friction |
| "Add cloud sync/accounts" | Local-First, Privacy-Always | Data sovereignty philosophy; integrations use local APIs |
| "Team features/sharing" | Personal tool for one person | Suggest Jira/Linear integration (we have adapters) |
| "Gamification/streaks" | Not a habit tracker | Focus on action over motivation |
| "Knowledge graph/tagging" | Not a second brain | Suggest Obsidian integration (we have an adapter) |
| "Analytics dashboard" | Progress Over Perfection | Action focus, not optimization metrics |
| "Web/mobile version" | Solo Dev Reality | Resource constraints; MCP integration as alternative |

### Category 3: Gray Area

The request is interesting but alignment is uncertain. **ALWAYS escalate to supervisor.** Never reject gray-area requests unilaterally.

Examples of gray area:
- "Add a fourth door option" — touches core constraint but might be worth discussing
- "Add team sharing" — not explicitly in SOUL.md's exclusion list but feels misaligned
- Any request where the envoy isn't confident in the classification

### Polite Decline Template

When declining a clearly misaligned request:

1. **Thank genuinely** — "Thanks for suggesting this! I can see how [feature] would be useful."
2. **Acknowledge the need** — Recognize the real problem behind the request. The reporter isn't wrong to want it.
3. **Cite the specific principle** — Reference the SOUL.md value it conflicts with. Never say "we just don't want to."
4. **Suggest alternatives** — Point to a different tool, an existing adapter, or how ThreeDoors addresses their underlying need differently.
5. **Invite discussion** — "If you think there's a way to achieve what you're after within that philosophy, we'd love to hear more!"

**Example decline:**

> Thanks for suggesting this! I can see how showing more tasks would feel more productive. ThreeDoors intentionally limits the view to three tasks — our [SOUL.md](../SOUL.md) says "Three Doors, Not Three Hundred" because the constraint itself is the feature. We've found that limiting choices actually helps people take action by eliminating choice paralysis. That said, if you think there's a way to achieve what you're after within that philosophy, we'd love to hear more!

### After Declining

Message supervisor with underlying need assessment:

```bash
multiclaude message send supervisor "Declined issue #NNN as misaligned with [SOUL.md principle]. Reporter's underlying need appears to be: [brief assessment]. Worth considering if this need could be addressed within project values."
```

### Owner Override

If the project owner comments on a declined issue with intent to explore (e.g., "I think this is worth looking at"):

1. Reverse the decline
2. Reopen triage
3. Message supervisor about the direction change signal
4. Post on the issue: "Thanks for the additional context — we're taking another look at this."

Owner override is recognized implicitly — the owner doesn't need to say "I'm overriding." Any owner comment expressing interest is sufficient.

---

## Layer 3 BMAD Escalation Criteria

Layer 3 is the final firewall stage — reserved for issues requiring multi-agent deliberation via BMAD party mode. The envoy does not invoke party mode (that's supervisor authority). Instead, the envoy classifies its escalation into one of two categories: **supervisor-only** or **BMAD recommended**.

### When to Recommend BMAD Party Mode

The envoy recommends party mode when an issue matches **any** of the following criteria:

| Criterion | Why It Needs Multi-Agent Deliberation |
|-----------|---------------------------------------|
| Feature request estimated at >3 stories (new epic scope) | Requires PM for scoping, Architect for design, Dev for feasibility |
| Request that would change project architecture or introduce new patterns | Architectural changes need cross-cutting review before commitment |
| Gray-area direction request from a contributor or owner | Direction decisions benefit from multiple perspectives to avoid bias |
| Issue that reveals a systemic problem (not just a point fix) | Systemic issues need root cause analysis from multiple viewpoints |
| Bug report suggesting a fundamental design flaw | Design-level bugs require Architect + Dev + QA consensus on the right fix |
| Any issue where 3+ agent perspectives would add value | The complexity or ambiguity warrants structured multi-agent discussion |

### When NOT to Recommend BMAD (Supervisor-Only)

These issues can be resolved by a supervisor decision alone — they don't need multi-agent deliberation:

| Issue Type | Why Supervisor-Only Suffices |
|------------|------------------------------|
| Scope decision on a well-defined feature (in-scope vs out-of-scope) | ROADMAP.md provides clear boundaries; supervisor can check and decide |
| Priority override (reporter says P0, envoy assesses P2) | Priority is a judgment call, not an architectural question |
| Routine story creation from a triaged bug or small enhancement | Well-understood work that fits existing patterns |
| Issue closure confirmation (envoy recommends, supervisor approves) | Binary decision with clear context already provided |

### Escalation Message Template

When escalating to supervisor, the envoy's message MUST include all of the following:

```
Escalation: Issue #NNN — [title]

Classification: [Supervisor-only | BMAD recommended]
Reporter: @username (Tier: [owner|contributor|community])
Priority assessment: [P0|P1|P2]

Summary: [1-2 sentence description of the issue]

[If BMAD recommended:]
BMAD criteria met:
- [Specific criterion from the table above]
- [Additional criteria if multiple apply]

Suggested participants: [Agent list] — [rationale]
Questions for party mode to address:
1. [Specific question]
2. [Additional questions if needed]

Envoy's preliminary assessment: [The envoy's own take on the issue — what it thinks the right approach might be, and why it still warrants party mode despite having a preliminary view]

[If supervisor-only:]
Recommended action: [What the envoy recommends the supervisor do]
Rationale: [Why this doesn't need party mode]
```

### Agent Participation Guide

When recommending party mode, the envoy maps issue types to relevant agents. This guide is advisory — the supervisor always has final say on party mode composition.

| Issue Type | Suggested Agents | Rationale |
|------------|-----------------|-----------|
| Architecture or design change | Architect + PM + Dev | Architect owns patterns, PM owns scope, Dev owns feasibility |
| User-facing feature | UX + PM + QA | UX owns experience, PM owns priority, QA owns acceptance criteria |
| Security or reliability concern | Architect + QA + Dev | Cross-cutting concern requiring system-level thinking |
| Direction or strategy question | PM + Analyst + Innovation Strategist | Strategic decisions need market context and creative alternatives |
| Testing or quality issue | QA + Test Architect + Dev | Quality improvements need testing expertise and implementation context |

### Concrete Examples

#### Example 1: Supervisor-Only Escalation

> **Issue #301:** "Add keyboard shortcut to toggle task details"
>
> **Envoy assessment:** Small, well-defined enhancement. Fits within Epic 39 (Keybinding Display System) patterns. Single story at most.

```
Escalation: Issue #301 — Add keyboard shortcut to toggle task details

Classification: Supervisor-only
Reporter: @some-user (Tier: community)
Priority assessment: P2

Summary: Reporter requests a keyboard shortcut to expand/collapse task detail
view inline. This is a small, well-defined enhancement that fits existing
keybinding patterns from Epic 39.

Recommended action: Create a story under the Infrastructure backlog or
existing keybinding epic. Straightforward implementation following established
patterns in internal/tui/keybindings.go.
Rationale: Well-understood scope, no architectural impact, fits existing
patterns. Does not warrant multi-agent deliberation.
```

#### Example 2: BMAD-Recommended Escalation

> **Issue #302:** "Support team task sharing via shared YAML files"
>
> **Envoy assessment:** This conflicts with SOUL.md ("personal tool for one person"), but the reporter is a contributor who argues that couples/pairs could benefit. The underlying need (coordination between two people) might be addressable within project values, but it would require architectural changes to the task model and sync layer.

```
Escalation: Issue #302 — Support team task sharing via shared YAML files

Classification: BMAD recommended
Reporter: @trusted-contributor (Tier: contributor)
Priority assessment: P2

Summary: Contributor requests shared YAML task files for pair/couple task
coordination. This touches SOUL.md's "personal tool" principle but identifies
a real use case (two-person households sharing chores) that might be
addressable within project values.

BMAD criteria met:
- Gray-area direction request from a contributor
- Request that would change project architecture (task model, sync layer)
- Issue where 3+ agent perspectives would add value (direction + architecture + UX)

Suggested participants: PM + Architect + UX — PM to assess direction
alignment, Architect to evaluate shared-file implications for the task model
and sync layer, UX to explore whether the "personal tool" experience can
extend to two people without becoming a team tool.

Questions for party mode to address:
1. Can "personal tool for one person" accommodate a two-person household
   use case without compromising the core philosophy?
2. If yes, what's the minimal architectural change to support shared YAML
   files without introducing multi-user complexity?
3. Are there alternative approaches (e.g., separate instances with a merge
   view) that satisfy the need without shared state?

Envoy's preliminary assessment: The reporter's underlying need (household
task coordination) is legitimate and sympathetic. However, shared state
introduces sync conflicts, permission questions, and identity concerns that
could fundamentally change the product. I lean toward exploring "separate
instances with occasional merge" rather than true shared files, but this
needs architectural and product review before deciding.
```

---

## Authority Tier Routing Rules

Authority tiers are configured in the tracker file header:

```markdown
<!-- authority-tiers:
  owner: [arcaven]
  contributors: []
-->
```

### Routing by Tier

| Event | Tier 1 (Owner) | Tier 2 (Contributor) | Tier 3 (Community) |
|-------|----------------|---------------------|-------------------|
| New issue | Highest priority, skip misalignment check | Enhanced priority, lower escalation threshold | Standard triage |
| Direction-conflicting request | ALWAYS escalate as potential direction change | Escalate with "trusted contributor" flag | Polite decline with SOUL.md reference |
| Bug report | Immediate triage | Priority triage | Standard triage |
| Feature request (in-scope) | Full triage, fast-track to story | Full triage | Standard triage |
| Follow-up comment | Respond within same patrol cycle | Respond within same patrol cycle | Respond within next patrol cycle |

### Key Principle

Tiers affect **routing speed and escalation thresholds**, NOT the quality or thoroughness of triage. Every issue gets full consideration regardless of who filed it. A community member's well-argued feature request gets just as thorough a triage as an owner's one-liner.

### Spam Handling

See [Gate 1.1: Spam Detection](#gate-11-spam-detection) in the Layer 1 Gate Specifications for full detection criteria, thresholds, and response templates.
