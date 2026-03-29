# Provenance Tagging Specification

## Overview

Provenance tagging tracks who did what and at what autonomy level. Every AI-generated artifact (story, commit, PR) carries an autonomy level tag. This is the foundation for dark factory traceability (R-003).

**Decision:** Q-C-007 — Mandatory for AI-generated work, opt-in for human.

## Autonomy Levels (L0-L4)

| Level | Name | Description | Example |
|-------|------|-------------|---------|
| L0 | Human-only | No AI involvement | Human writes code in editor |
| L1 | AI-assisted | Human directs, AI helps | Human asks Claude for suggestions, reviews every change |
| L2 | AI-paired | Collaborative | `/implement-story` with active human oversight |
| L3 | AI-autonomous | AI works independently, human reviews PR | Worker implements story, human merges |
| L4 | AI-full | AI end-to-end, no human review | Dark factory variant (future) |

## Where Provenance Appears

### 1. Story Files

Story files include a `## Provenance` section after implementation:

```markdown
## Provenance
- **Autonomy Level:** L3 (AI-autonomous)
- **Implementation Agent:** worker/<name>
- **Review:** Human PR review required
```

This section is added by the implementing worker after completing the story.

### 2. Commit Messages

AI-generated commits include a `Provenance:` trailer on its own line:

```
feat: implement provenance tagging (Story 74.3)

Provenance: L3
```

The trailer follows git convention and is added by the worker agent at commit time, not post-hoc.

### 3. PR Labels

GitHub labels classify PRs by autonomy level:

| Label | Color | Description |
|-------|-------|-------------|
| `provenance.L0` | Blue (#2B67C6) | Human-only: no AI involvement |
| `provenance.L1` | Green (#1D8348) | AI-assisted: human directs, AI helps |
| `provenance.L2` | Orange (#F39C12) | AI-paired: collaborative human-AI work |
| `provenance.L3` | Purple (#8E44AD) | AI-autonomous: AI works independently, human reviews |
| `provenance.L4` | Red (#E74C3C) | AI-full: end-to-end AI, no human review |

Workers apply the appropriate label when creating a PR. merge-queue warns if a `work/*` branch PR is missing a provenance label.

## Mandatory vs Opt-in

| Source | Requirement |
|--------|-------------|
| Worker agents (`work/*` branches) | **Mandatory** — must tag story file, commits, and PR |
| `/plan-work` output | **Mandatory** — L2 (AI-paired) |
| Human commits | **Opt-in** — absence means "unknown/human" |
| Persistent agents (merge-queue, pr-shepherd) | Not applicable — infrastructure, not implementation |

## Common Autonomy Levels by Activity

| Activity | Typical Level |
|----------|---------------|
| Worker implements story via `/implement-story` | L3 |
| `/plan-work` creates stories and planning docs | L2 |
| Human writes code with Claude suggestions | L1 |
| Human writes code without AI | L0 |
| Dark factory fully automated pipeline (future) | L4 |

## Dark Factory Usage

Provenance data enables dark factory (R-003) to:

1. **Track AI contribution ratio** — what percentage of work is AI vs human
2. **Measure autonomy progression** — are we moving from L2 to L3 to L4 over time
3. **Audit trail** — every artifact traces back to its origin for compliance
4. **Performance analysis** — compare cycle time, defect rate by autonomy level
5. **Trust calibration** — identify which autonomy levels produce reliable output

## Implementation Notes

- Provenance tags are immutable once set — do not retroactively change them
- If a human significantly reworks an AI PR before merge, the provenance still reflects the original creation level (the human review is captured in the PR review process)
- The commit trailer format uses a single line: `Provenance: L3` (no parenthetical description needed in commits)
- PR labels use the dotted format: `provenance.L3`
