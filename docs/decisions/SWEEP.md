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

> **Note:** Research was consolidated from `docs/research/` into `_bmad-output/planning-artifacts/` (PR #324). All research, analysis, spike reports, and party mode outputs now live in a single flat directory with type suffixes for filtering.

### 3. Archive Aging

**Location:** `docs/decisions/BOARD.md` — Recently Decided and Recently Rejected sections

**Check:** For each entry in Recently Decided and Recently Rejected, compare the entry date against the current date. Entries older than 30 days should be moved to `ARCHIVE.md`.

**Why:** The active board should show only recent activity. Older entries clutter the dashboard and make it harder to spot items needing attention. The archive preserves the full history.

**Action:** Flag entries older than 30 days for archival. Include the entry ID, current section, date, and days since entry.

### 4. Stale Entries

**Criteria:**

| Section | Staleness Threshold | Marker | Rationale |
|---------|-------------------|--------|-----------|
| Under Investigation | > 2 weeks without update | `[STALE]` prefix on Topic | Research should conclude or report progress |
| Needs Decision | > 1 month without decision | `[STALE]` prefix on Recommendation | Recommendations awaiting sign-off shouldn't languish |

**Check:** Compare the date in each entry against the current date. Flag entries exceeding their threshold.

**Staleness markers:** When an entry exceeds its threshold, the sweep should recommend adding a `[STALE]` prefix to the entry's primary text column (Topic for Under Investigation, Recommendation for Needs Decision). This makes staleness visible at a glance without requiring date arithmetic.

### 5. ADR Candidates

**Criteria:** A decided item (D-NNN) is an ADR candidate if it has:
- Broad architectural impact (affects multiple packages or subsystems)
- Multiple stakeholders or consumers
- Irreversible or costly-to-reverse consequences
- No existing ADR link in its board entry

**Check:** Scan "Recently Decided" entries (and archived Decided entries) where the Link column does not reference an ADR file. Flag those matching the criteria above.

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
- Entries needing archival: N
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

## Entries Needing Archival

| ID | Section | Date | Days Old | Suggested Action |
|----|---------|------|----------|------------------|
| D-NNN | Recently Decided | YYYY-MM-DD | NN | Move to ARCHIVE.md |

## Stale Entries

| ID | Section | Date | Days Stale | Suggested Action |
|----|---------|------|------------|------------------|
| R-001 | Under Investigation | YYYY-MM-DD | NN | Add [STALE] marker, check progress or close |

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

1. Read `docs/decisions/BOARD.md` and `docs/decisions/ARCHIVE.md`
2. List files in `_bmad-output/planning-artifacts/`
3. Cross-reference each file against BOARD.md and ARCHIVE.md content
4. Check dates on Recently Decided and Recently Rejected entries for 30-day archival
5. Check dates on Under Investigation and Needs Decision entries for staleness markers
6. Scan Recently Decided entries (and archived Decided entries) for ADR candidates without ADR links
7. Produce the report in the format above
8. Send the report to the supervisor

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
| Unindexed research | Supervisor reviews | Add board entry with appropriate section |
| Entry needing archival | Supervisor approves | Move from Recently Decided/Rejected to ARCHIVE.md |
| Stale Under Investigation | Supervisor checks with owner | Update date, add [STALE] marker, close, or escalate |
| Stale Needs Decision | Supervisor makes decision | Move to Recently Decided or Recently Rejected |
| ADR candidate | Supervisor evaluates | Create ADR or note as not ADR-worthy |

The supervisor may delegate action (e.g., ask a worker to draft a board entry PR), but the decision to act always starts with the supervisor.

---

## Standing Order

> **Board hygiene sweep runs weekly** (or after major milestones). Any agent may run the sweep. Results go to the supervisor. No auto-creation of entries — the sweep only reports.

This standing order should be added to the supervisor's memory or standing orders list.
