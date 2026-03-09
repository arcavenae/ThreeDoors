# Dead Research Audit — 2026-03-09 (Follow-up)

**Auditor:** happy-bear (worker agent)
**Scope:** All files in `_bmad-output/planning-artifacts/`, cross-referenced against `docs/decisions/BOARD.md`
**Prior audit:** `dead-research-audit-2026-03-08.md` — this audit builds on it

---

## Summary

| Finding | Count |
|---------|-------|
| Artifacts created since prior audit | 5 |
| New untracked artifacts with decisions | 2 |
| Supporting docs (no standalone decisions) | 3 |
| Broken links in BOARD.md | 2 |
| **BOARD.md entries added** | **16** |

---

## New Untracked Artifacts (Post Prior Audit)

### 1. Scoped Labels — Untracked, Contains Decisions

| Artifact | Status | Action Taken |
|----------|--------|-------------|
| `scoped-labels-party-mode.md` | 13 decisions + 13 rejected options, NONE in BOARD.md | Added P-005, D-106 through D-111, X-050 through X-058 |
| `scoped-labels-research.md` | Research input for above party mode | Linked from P-005 |

**Context:** Party mode (3 rounds, 4 participants) produced a complete 27-label scoped taxonomy with `.` separator, migration plan, and authority matrix. Implementation awaits a separate story (SL-013/D-111). The overall recommendation is tracked as P-005.

### 2. Beautiful Stats — Supporting Docs, Already Tracked

| Artifact | Status | Action Taken |
|----------|--------|-------------|
| `architecture-beautiful-stats.md` | Implementation guide for Epic 40 | No action — decisions captured via D-096 through D-104 |
| `beautiful-stats-research.md` | Research input for Epic 40 party mode | No action — supporting doc |
| `beautiful-stats-ux-review.md` | UX review for Epic 40 | No action — supporting doc |

These three artifacts fed into the Beautiful Stats party mode, whose decisions are already tracked as D-096 through D-104. No standalone decisions exist beyond what's captured.

---

## Broken Links in BOARD.md

| Entry | Referenced File | Issue |
|-------|----------------|-------|
| P-002 | `envoy-scope-and-firewall-design.md` | File was never committed; prior audit identified it as "uncommitted" |
| P-003 | `issue-labeling-and-triage-strategy.md` | File was never committed; prior audit identified it as "uncommitted" |

**Action taken:** Updated P-002 and P-003 link text to note artifacts are missing. The recommendations themselves remain valid (backed by other evidence), but the referenced artifacts would need to be recreated or located from worktree history.

---

## BOARD.md Entries Added

### Pending Recommendations (1 new)

| ID | Recommendation |
|----|----------------|
| P-005 | Scoped label taxonomy: 27 labels with `.` separator, migration plan |

### Decided (6 new)

| ID | Decision |
|----|----------|
| D-106 | `.` as label scope separator |
| D-107 | 9 scopes, 27 total labels (trimmed from 35) |
| D-108 | `status.do-not-merge` label for merge-queue |
| D-109 | Agent labels limited to envoy + worker only |
| D-110 | Label migration via rename-first strategy |
| D-111 | Implementation as separate story from research |

### Rejected (9 new)

| ID | Option |
|----|--------|
| X-050 | `epic.N` per-epic labels |
| X-051 | `sprint.*` labels |
| X-052 | `effort.*` Fibonacci labels |
| X-053 | Full 5-agent `agent.*` set |
| X-054 | `status.in-review/approved/changes-requested` |
| X-055 | `status.merge-ready` and `status.ci-failing` |
| X-056 | `::` (GitLab-style) separator |
| X-057 | `/` as separator |
| X-058 | `process.party-mode` label |

---

## Prior Audit Status

The prior audit (2026-03-08) identified 68 untracked items and added 17 Decided entries + 1 Rejected + 3 Pending Recommendations. All of those entries are confirmed present in the current BOARD.md. The prior audit's work remains valid and complete.

---

## Remaining Gaps

1. **P-002 and P-003 artifacts need recreation** — The envoy firewall design and issue labeling strategy artifacts were never committed. The recommendations are still valid but lack their source artifacts.

2. **Q-001 and Q-002 (Jira questions)** — Flagged by prior audit as potentially resolved during Epic 19 implementation. Still open in BOARD.md. Recommend verifying against Jira adapter implementation and closing if resolved.

3. **No other dead research found** — All 73 artifacts in `_bmad-output/planning-artifacts/` are now either tracked in BOARD.md or classified as supporting/process documents that don't require standalone entries.
