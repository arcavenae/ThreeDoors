# BOARD.md Redesign Research

**Date:** 2026-03-15
**Type:** Research — Decision management improvement
**Scope:** Research only — no stories/epics created

---

## Problem Statement

BOARD.md has grown to 411 lines containing:
- **4 Open Questions** (all resolved inline but still in "Open" section)
- **1 Active Research** entry (resolved)
- **6 Pending Recommendations** (5 marked "Done", 1 awaiting sign-off)
- **181 Decided entries** (D-001 through D-181, with duplicate IDs: two D-059s, two D-074s, two D-112-D-117s, two D-128s, two D-137s, two D-153-D-156s)
- **118 Rejected entries** (X-001 through X-118, with duplicate IDs in X-050-X-072 range)
- **Epic Number Registry** (30 entries)
- **Two separate "Rejected" sections** (structural duplication)

### Specific Problems

1. **Resolved items linger in active sections.** All 4 Open Questions are marked "Resolved" inline but remain in the Open Questions table. All but one Pending Recommendation is "Done." The Active Research entry is "Resolved." These sections exist to surface items needing attention — resolved items are noise.

2. **Decided section is an undifferentiated wall.** 181 entries spanning architectural foundations (D-001: Go as primary language) to tactical micro-decisions (D-117: sort search results in filterTasks). No distinction between foundational, still-relevant decisions and historical decisions that are fully implemented and no longer need reference.

3. **Duplicate IDs.** Multiple ID collisions due to parallel agent work: D-059 (two different decisions), D-074 (two), D-112-117 (two sets), D-128 (two), D-137 (two), D-153-156 (two sets). This undermines the ID system's purpose.

4. **Rejected section is enormous and rarely consulted.** 118 entries. Valuable for preventing re-proposals, but most are micro-decisions (X-046 through X-049 are four rejected spacebar alternatives). The signal-to-noise ratio is low for everyday use.

5. **Two separate Rejected sections.** Lines 201-249 and 286-363 both contain rejected entries, indicating structural drift from accumulation.

6. **BOARD.md consumes ~29K tokens.** Every agent that reads it burns significant context. For a file whose purpose is "living dashboard," most of the content is historical and doesn't need to be in the active view.

7. **Epic Number Registry mixed with decisions.** The registry serves a completely different purpose (coordination mutex) but lives in BOARD.md, adding to its size and confusion.

---

## Current State Analysis

### Entry Classification

| Section | Count | Resolved/Done | Still Active |
|---------|-------|---------------|-------------|
| Open Questions | 4 | 4 (100%) | 0 |
| Active Research | 1 | 1 (100%) | 0 |
| Pending Recommendations | 6 | 5 (83%) | 1 (P-001: Justfile migration) |
| Decided | ~181 | All decided | ~181 (but most are historical) |
| Rejected | ~118 | All rejected | ~118 (but most are historical) |

**Key finding:** The "active" sections (Open Questions, Active Research, Pending Recommendations) are 100% or near-100% resolved — the board has zero items genuinely awaiting action except possibly P-001. The BOARD is not functioning as a live dialog; it's functioning as an append-only archive.

### Temporal Distribution of Decided Entries

| Period | Entries | Notes |
|--------|---------|-------|
| Nov 2025 | D-001 to D-009 | Foundational architectural decisions |
| Jan-Feb 2026 | D-010 to D-028 | Core infrastructure decisions |
| Mar 1-7, 2026 | D-029 to D-052 | Rapid growth phase |
| Mar 8-9, 2026 | D-053 to D-141 | Explosion (~90 decisions in 2 days) |
| Mar 10-15, 2026 | D-142 to D-181 | Continued high velocity |

The March 8-9 burst corresponds to the introduction of party mode and the BOARD itself (D-029), triggering a retroactive cataloging of decisions. Many of these entries document decisions that were already made — the board is functioning as a historical record rather than a decision-making tool.

---

## Best Practices Research

### Architectural Decision Records (ADRs) — Nygard/ThoughtWorks

- **Lightweight format:** Title, Status, Context, Decision, Consequences
- **Key insight:** ADRs are immutable once accepted. Superseded ADRs get a new status but the original text is preserved.
- **Weakness for ThreeDoors:** Too heavyweight for micro-decisions. Already rejected as sole format (X-002).

### Decision Log Pattern (Spotify/Netflix)

- **Tiered approach:** Strategic decisions get full writeups; tactical decisions get one-liners in a log
- **Relevance:** ThreeDoors already does this implicitly (ADRs for big decisions, BOARD entries for tactical ones)
- **Improvement opportunity:** Make the tier separation explicit and structural

### MADR (Markdown Any Decision Records)

- **Statuses:** proposed, accepted, deprecated, superseded
- **Key insight:** "deprecated" status for decisions that are still valid but no longer actively relevant
- **Applicable to ThreeDoors:** Could add "Implemented" as a status to distinguish live decisions from completed ones

### Living Document Anti-Pattern (Common in Agile)

- **Problem:** Append-only documents grow until they're unreadable, then get ignored
- **Solution:** Regular archival cadence — move resolved/completed items to an archive, keeping the active view focused
- **Key insight:** The value of a "living document" is that it reflects current state. If most content is historical, it's not living — it's a graveyard.

### Basecamp Shape Up — "Bets" Board

- **Active bets (this cycle) are prominent; past bets are archived**
- **Key insight:** Separate "what we're deciding now" from "what we decided before"

---

## Design Principles for Redesign

From SOUL.md:
- **"Show 3 tasks. Not 5. Not 'all of them with filters.'"** — The constraint is the feature. BOARD.md should show active decisions, not all decisions.
- **"The Director's Role"** — The human's job is decision-making. The board should surface decisions needing human input and minimize noise from completed decisions.
- **"Solo Human, Agent Team"** — The board is the primary human-AI decision dialog. It must be scannable in seconds, not minutes.
- **"Progress Over Perfection"** — The redesign should be incremental, not a big-bang rewrite.

---

## Proposed Redesign: Active Board + Decision Archive

### Approach: Two-File Split with Status-Based Sections

**BOARD.md** remains the living dashboard but is radically trimmed to contain ONLY items needing attention or recently decided. **ARCHIVE.md** holds the complete historical record.

### New BOARD.md Structure (Target: < 100 lines active content)

```markdown
# Knowledge Decisions Board

> Active decisions and open dialog for ThreeDoors.
> Full history: [ARCHIVE.md](ARCHIVE.md)

## Needs Decision (Human Input Required)

| ID | Question/Recommendation | Date | Source | Link |
|----|------------------------|------|--------|------|
| ... items awaiting human sign-off ... |

## Under Investigation

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| ... active research and open questions ... |

## Recently Decided (Last 30 Days)

| ID | Decision | Date | Rationale | Link |
|----|----------|------|-----------|------|
| ... decisions from the last 30 days ... |

## Recently Rejected (Last 30 Days)

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| ... rejections from the last 30 days ... |

## Epic Number Registry

(unchanged — or moved to its own file)
```

### New ARCHIVE.md Structure

```markdown
# Decision Archive

> Complete historical record of all ThreeDoors project decisions.
> Active board: [BOARD.md](BOARD.md)

## All Decided (Chronological)

| ID | Decision | Date | Rationale | Link |
|----|----------|------|-----------|------|
| D-001 | ... | ... | ... | ... |
...all 181+ entries...

## All Rejected (Chronological)

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-001 | ... | ... | ... | ... |
...all 118+ entries...

## Resolved Questions

| ID | Question | Date | Resolution |
|----|----------|------|------------|
...moved from Open Questions...

## Completed Research

| ID | Topic | Date | Outcome |
|----|-------|------|---------|
...moved from Active Research...

## Superseded

(unchanged)
```

### Key Design Decisions

1. **"Needs Decision" section replaces Open Questions + Pending Recommendations.** The user doesn't care whether something is a question vs. a recommendation — they care whether it needs their input.

2. **"Under Investigation" replaces Active Research.** Clearer label for the human reader.

3. **"Recently Decided/Rejected" with 30-day window.** Keeps recent decisions visible for context without accumulating forever. After 30 days, items age into the archive automatically during sweep.

4. **30-day window is a guideline, not a hard cutoff.** The sweep process moves items; agents don't need to compute dates. "Recently" means "since the last sweep moved things to archive."

5. **Archive preserves everything.** Nothing is lost. The archive is the complete record. The board is the focused view.

6. **Duplicate IDs get fixed during migration.** Re-number the second instance of each duplicate (e.g., the second D-059 becomes D-059b or gets the next available number).

7. **Epic Number Registry moves to its own file** (`EPIC_REGISTRY.md`) since it serves a different purpose (coordination mutex) and doesn't need to be in the decision dashboard.

### Staleness Markers

Instead of implicit staleness (entries sitting in sections for weeks), add explicit staleness:

- **Sweep adds `[STALE]` prefix** to entries in "Needs Decision" or "Under Investigation" that exceed thresholds (from SWEEP.md: 2 weeks for research, 1 month for questions/recommendations)
- **Stale items are visually distinct** — the `[STALE]` marker makes them scannable
- **Staleness is a sweep output**, not something agents manage inline

---

## Alternatives Considered

### Alternative A: Status Tags on Current Single File

Add `[ACTIVE]`, `[RESOLVED]`, `[IMPLEMENTED]` tags to every entry in the current BOARD.md. Keep everything in one file.

**Why rejected:** Doesn't address the core problem — the file is still 400+ lines and growing. Tags reduce noise slightly but don't create the "focused active view" the human needs. Also adds maintenance burden (every entry needs a tag) without structural improvement.

### Alternative B: Database-Backed Board (YAML/JSON + Generated Markdown)

Store decisions in structured YAML/JSON. Generate BOARD.md from data. Enable queries, filters, and views.

**Why rejected:** Over-engineering. Violates SOUL.md "simplest thing that works." Adds tooling dependency. The board is ~200 decisions — a text file with good structure is sufficient. Also, D-029 explicitly chose "zero infrastructure" for the board.

### Alternative C: Per-Epic Decision Files

Split decisions into per-epic files (`decisions/epic-39.md`, `decisions/epic-40.md`). Board becomes an index.

**Why rejected:** Scatters decisions. The board's value is having everything in one searchable place. Per-epic files make cross-cutting decisions hard to find. Also creates many small files (65+ epics) with high overhead.

### Alternative D: Aggressive Archival (Delete Old Entries)

Remove entries older than 90 days from BOARD.md entirely. Trust git history for archaeology.

**Why rejected:** Git history is the worst way to find a past decision. The archive file approach preserves searchability while keeping the active board clean. "Nothing lost" is a design requirement.

### Alternative E: Three-Tier System (Active / Reference / Archive)

Three files: BOARD.md (active), REFERENCE.md (decided but still relevant), ARCHIVE.md (historical).

**Why rejected:** The "still relevant" distinction is subjective and high-maintenance. Every decision would need periodic re-evaluation for relevance. Two tiers (active vs. archive) with a 30-day recently-decided window provides nearly the same benefit with much less maintenance. The REFERENCE tier would eventually suffer the same bloat problem as the current Decided section.

---

## Migration Plan (Implementation Guidance)

If this redesign is adopted, the migration would be:

1. **Create ARCHIVE.md** with all current Decided, Rejected, and Superseded entries
2. **Create EPIC_REGISTRY.md** moved from BOARD.md
3. **Trim BOARD.md** to new structure:
   - Move resolved Open Questions to archive
   - Move resolved Active Research to archive
   - Move "Done" Pending Recommendations to archive
   - Keep P-001 (Justfile) in "Needs Decision"
   - Keep last 30 days of Decided entries in "Recently Decided"
   - Keep last 30 days of Rejected entries in "Recently Rejected"
4. **Fix duplicate IDs** during migration
5. **Fix duplicate Rejected sections** (merge into one)
6. **Update README.md** to reflect new structure
7. **Update SWEEP.md** to include archive aging as a sweep responsibility
8. **Update CLAUDE.md** references if any point to BOARD.md sections by name

**Estimated resulting BOARD.md size:** ~80-120 lines (down from 411)
**Estimated ARCHIVE.md size:** ~350+ lines (but rarely read in full)

---

## Recommendations

1. **Adopt the two-file split** (Active Board + Decision Archive) — this is the core recommendation
2. **Move Epic Number Registry** to its own file
3. **Fix duplicate IDs** during migration
4. **Update sweep process** to include 30-day aging from board to archive
5. **Create a story** for the migration (single story, ~2-3 hours of agent work)

---

## Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| Two-file split: active BOARD.md + ARCHIVE.md | Adopted | Keeps active view focused (<100 lines) while preserving complete history; aligns with SOUL.md "show less" philosophy | Single file with tags (A: still bloated), database-backed (B: over-engineering), per-epic files (C: scatters decisions), aggressive deletion (D: loses searchability), three-tier (E: high maintenance) |
| "Needs Decision" section replaces Open Questions + Pending Recommendations | Adopted | User cares about "needs my input" not the lifecycle stage; reduces cognitive overhead | Keep separate sections (traditional ADR lifecycle — adds unnecessary process distinction) |
| 30-day recently-decided window on active board | Adopted | Provides recent context without accumulation; sweep handles aging | No window/show all (current problem), 7-day (too aggressive, loses useful context), 90-day (too much accumulation) |
| Move Epic Number Registry to own file | Adopted | Different purpose (coordination mutex), different audience (agents), different update frequency | Keep in BOARD.md (contributes to bloat, confuses purpose) |
| Fix duplicate IDs during migration | Adopted | IDs must be unique for references to work; migration is the natural time to fix | Ignore (breaks references), auto-renumber all (unnecessary churn) |
| Staleness via `[STALE]` marker added by sweep | Adopted | Makes staleness visible without requiring date math by readers; sweep is the natural staleness detector | Automatic date-based styling (requires tooling), no staleness markers (current problem of invisible rot) |
