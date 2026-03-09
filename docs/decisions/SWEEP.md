# Board Hygiene Sweep Process

## Purpose

The Knowledge Decisions Board (`BOARD.md`) drifts as new artifacts, research, and decisions accumulate across the project. The sweep process identifies gaps — unindexed artifacts, missing research entries, stale items — and reports them to the supervisor for action.

**Philosophy:** The sweep **reports findings** but **never auto-creates** board entries or ADRs. Entry creation requires judgment about significance, framing, and categorization that only a human or supervisor can provide.

---

## Scan Targets

### 1. Unindexed Planning Artifacts

**Location:** `_bmad-output/planning-artifacts/`

**Check:** For each file in this directory, verify that `BOARD.md` contains a reference (filename or identifiable substring). Files without any board reference are flagged.

**Why:** Party mode artifacts contain decisions that should be discoverable from the board. An unindexed artifact means decisions are invisible to workers checking the board.

### 2. Unindexed Research

**Location:** `_bmad-output/planning-artifacts/` (research files use the `-research` naming suffix)

**Check:** For each `*-research.md` file in planning-artifacts, verify that `BOARD.md` contains a reference to it. Research without a board entry is flagged.

**Why:** Completed research often contains recommendations that need a decision. Without a board entry, research sits orphaned with no follow-up mechanism.

> **Note:** Research was consolidated from `docs/research/` into `_bmad-output/planning-artifacts/` (PR #TBD). All research, analysis, spike reports, and party mode outputs now live in a single flat directory with type suffixes for filtering.

### 3. Stale Entries

**Criteria:**

| Column | Staleness Threshold | Rationale |
|--------|-------------------|-----------|
| Active Research | > 2 weeks without update | Research should conclude or report progress |
| Pending Recommendations | > 1 month without decision | Recommendations awaiting sign-off shouldn't languish |
| Open Questions | > 1 month without movement | Questions should either start research or be closed |

**Check:** Compare the date in each entry against the current date. Flag entries exceeding their threshold.

### 4. ADR Candidates

**Criteria:** A decided item (D-NNN) is an ADR candidate if it has:
- Broad architectural impact (affects multiple packages or subsystems)
- Multiple stakeholders or consumers
- Irreversible or costly-to-reverse consequences
- No existing ADR link in its board entry

**Check:** Scan "Decided" entries where the Link column does not reference an ADR file. Flag those matching the criteria above.

---

## Report Format

The sweep produces a markdown report with findings organized by category:

```markdown
# Board Hygiene Sweep Report

**Date:** YYYY-MM-DD
**Sweep run by:** [agent name or manual]

## Summary

- Unindexed artifacts: N
- Unindexed research: N
- Stale entries: N
- ADR candidates: N
- **Total findings: N**

## Unindexed Planning Artifacts

| File | Created/Modified | Suggested Action |
|------|-----------------|------------------|
| `artifact-name.md` | YYYY-MM-DD | Review for board-worthy decisions |

## Unindexed Research

| File | Status | Suggested Action |
|------|--------|------------------|
| `research-topic.md` | Complete/In Progress | Add board entry in [column] |

## Stale Entries

| ID | Column | Date | Days Stale | Suggested Action |
|----|--------|------|------------|------------------|
| R-001 | Active Research | YYYY-MM-DD | NN | Check progress or close |

## ADR Candidates

| ID | Decision | Why ADR-Worthy |
|----|----------|---------------|
| D-NNN | Decision description | Broad impact / irreversible / multi-stakeholder |

## Recommended Actions

1. [Prioritized list of recommended actions for the supervisor]
```

The report is addressed to the supervisor, who decides which findings to act on and which to dismiss.

---

## Running the Sweep

### Manual Execution

Any agent or the supervisor can run the sweep by following these steps:

1. Read `docs/decisions/BOARD.md`
2. List files in `_bmad-output/planning-artifacts/`
3. Cross-reference each file against BOARD.md content
4. Check dates on Active Research, Pending Recommendations, and Open Questions entries
5. Scan Decided entries for ADR candidates without ADR links
6. Produce the report in the format above
7. Send the report to the supervisor

### Via `/loop` Skill

Set up a recurring sweep:

```
/loop 7d [run board hygiene sweep per docs/decisions/SWEEP.md and report findings]
```

This runs the sweep weekly. Adjust the interval based on project activity:
- **High activity** (multiple PRs/day): Every 3-4 days
- **Normal activity**: Weekly
- **Low activity**: Every 2 weeks

### Trigger-Based Sweeps

Run an ad-hoc sweep after:
- A major milestone or epic completion
- A batch of PRs merging (e.g., after a sprint)
- Noticing outdated board entries during normal work

---

## Escalation Paths

| Finding Type | Escalation | Expected Response |
|-------------|-----------|-------------------|
| Unindexed artifact | Supervisor reviews | Add board entry or note as not board-worthy |
| Unindexed research | Supervisor reviews | Add board entry with appropriate column |
| Stale Active Research | Supervisor checks with owner | Update date, close, or escalate |
| Stale Pending Recommendation | Supervisor makes decision | Move to Decided or Rejected |
| Stale Open Question | Supervisor triages | Start research, close as won't-do, or re-assign |
| ADR candidate | Supervisor evaluates | Create ADR or note as not ADR-worthy |

The supervisor may delegate action (e.g., ask a worker to draft a board entry PR), but the decision to act always starts with the supervisor.

---

## Standing Order

> **Board hygiene sweep runs weekly** (or after major milestones). Any agent may run the sweep. Results go to the supervisor. No auto-creation of entries — the sweep only reports.

This standing order should be added to the supervisor's memory or standing orders list.
