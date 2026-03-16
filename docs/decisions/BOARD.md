# Knowledge Decisions Board

> The living dashboard for all ThreeDoors project decisions. See [README.md](README.md) for how this board works.
>
> **Historical decisions** (Decided, Rejected, Resolved Questions, Completed Research, Superseded) have been moved to [ARCHIVE.md](ARCHIVE.md).
> **Epic Number Registry** has been moved to [EPIC_REGISTRY.md](EPIC_REGISTRY.md).

---

## Open Questions

*All questions resolved — see [ARCHIVE.md](ARCHIVE.md) for resolved questions.*

## Active Research

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| R-002 | PRD Post-Reconstruction Quality Audit — 3 HIGH issues, 4 MEDIUM issues identified in formal BMAD validation | 2026-03-15 | PM Validation | [Report](../../_bmad-output/planning-artifacts/prd-validation-report-2026-03-15.md) — **Open.** HIGH: stale next-steps.md, chaotic phase numbering in product-scope.md, missing v2.0 change log entry. MEDIUM: stale BOARD.md epic registry, incomplete user journeys, no YAML frontmatter, stale checklist-results-report.md. |

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-006 | In-app bug reporting via `:bug` command — browser URL primary, PAT upgrade, file fallback | 2026-03-09 | Party mode (4 rounds: PM, Architect, UX, Dev) | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md), [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) | **Done** — Epic 50 created (3 stories) |
| P-007 | PRD post-reconstruction cleanup: update next-steps.md, consolidate phase numbering, add v2.0 change log, sync BOARD.md epic registry, add user journeys | 2026-03-15 | PRD Validation (R-002) + Party mode (6 rounds: PM, Architect, Analyst, UX) | [Validation Report](../../_bmad-output/planning-artifacts/prd-validation-report-2026-03-15.md), [Party Mode](../../_bmad-output/planning-artifacts/prd-remediation-party-mode-2026-03-15.md) | **Done** — Stories 0.61 (cleanup sprint), 0.62 (phase consolidation), 0.63 (user journey expansion) |
| P-001 | Migrate from Makefile to Justfile | 2026-03-04 | Research spike | [Analysis](../../_bmad-output/planning-artifacts/makefile-vs-justfile-analysis.md) | Owner sign-off |
| P-002 | Envoy three-layer firewall implementation | 2026-03-08 | Party mode (8 sessions) | [Plan](../../_bmad-output/planning-artifacts/envoy-three-layer-firewall-plan.md), [Party Mode](../../_bmad-output/planning-artifacts/envoy-rules-of-behavior-party-mode.md) | **Done** — Epic 52 created (4 stories). Original artifact missing; plan reconstructed from available sources. Epic number provisional. |
| P-003 | GitHub issue labeling taxonomy and triage flow | 2026-03-08 | Party mode (5 sessions) | [Artifact](../../_bmad-output/planning-artifacts/issue-labeling-and-triage-strategy.md) | **Done** — Story 0.46 (triage flow docs) |
| P-004 | Update pr-shepherd definition to remove fork references | 2026-03-08 | Investigation | [Research](../../_bmad-output/planning-artifacts/persistent-agent-communication-research.md) | **Done** — fork references already removed during Story 51.2 agent definition rewrite (PR #460) |
| P-005 | Scoped label taxonomy: 27 labels with `.` separator, migration plan | 2026-03-08 | Party mode (3 rounds) + research spike | [Party Mode](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md), [Research](../../_bmad-output/planning-artifacts/scoped-labels-research.md) | **Done** — Stories 0.44 (migration), 0.45 (agent defs), 0.46 (authority docs) |

## Decided

*See [ARCHIVE.md](ARCHIVE.md) for all decided entries (D-001 through D-181).*

## Rejected

*See [ARCHIVE.md](ARCHIVE.md) for all rejected entries (X-001 through X-118).*

## Epic Number Registry

*Moved to [EPIC_REGISTRY.md](EPIC_REGISTRY.md).*

## Superseded

*See [ARCHIVE.md](ARCHIVE.md) for superseded entries.*
