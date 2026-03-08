# Governance Model

## Overview

ThreeDoors uses a multi-agent development system (BMAD + Multiclaude) where AI agents handle the majority of implementation. This document defines how decisions flow, who has authority to act, and what safeguards prevent drift.

## Decision Pipeline

```
Research / Issue Report
  → Party Mode Deliberation (multi-agent consensus)
    → Sprint Change Proposal (if mid-sprint)
    → Decision Board Entry (docs/decisions/BOARD.md)
      → Story Creation (docs/stories/)
        → Implementation (/implement-story)
```

Every decision that affects architecture, scope, or project direction flows through this pipeline. Ad-hoc decisions made during implementation must be recorded retroactively.

## Agent Authority Tiers

Each agent has an `## Authority` section defining three tiers:

| Tier | Meaning |
|------|---------|
| **CAN** (Autonomous) | Agent acts without asking. These are routine operations within the agent's domain. |
| **CANNOT** (Forbidden) | Agent must never do these, even if asked by another agent. Only a human can override. |
| **ESCALATE** (Requires Human) | Agent must stop and ask before proceeding. These are judgment calls that need human input. |

See individual agent files in `agents/` for specific authority definitions.

### Persistent Agents

| Agent | Domain | Key Authority |
|-------|--------|---------------|
| merge-queue | PR merging | Merge when CI green + no blockers; halt on roadmap violations |
| pr-shepherd | Fork PR management | Rebase and fix CI; cannot merge |
| envoy | Community triage | Welcome, triage, create stories; cannot make scope decisions |
| arch-watchdog | Architecture compliance | Flag violations; cannot override human decisions |
| project-watchdog | Planning doc health | Detect drift; cannot modify story files |

### Ephemeral Agents

| Agent | Domain | Key Authority |
|-------|--------|---------------|
| worker | Implementation | Build within assigned scope; cannot expand scope |
| reviewer | Code review | Block for security/bugs; cannot block for style |
| release-manager | Release recovery | Fix and patch-release; cannot bump major/minor |

## Decision Tiers (ADR-0030)

Not all decisions need the same rigor. Three tiers control the cost of deliberation:

### Tier 1: Quick Decision
- **Process:** Single agent recommends, human approves
- **When:** Config/tooling changes, naming decisions, minor refactors, story task ordering
- **Artifact:** Brief note in PR description or Decision Board entry
- **Examples:** "Should we use `snake_case` for this field?", "Which linter rule to enable?"

### Tier 2: Standard Decision
- **Process:** 3-agent party mode (relevant domain experts)
- **When:** Feature design, API contracts, integration approaches, multi-story coordination
- **Artifact:** Party mode artifact in `_bmad-output/planning-artifacts/`
- **Examples:** "How should the Todoist adapter handle field mapping?", "What's the UX for snooze?"

### Tier 3: Full Decision
- **Process:** Full party mode (6+ agents), extensive deliberation
- **When:** Architecture changes, philosophy/SOUL.md changes, epic scoping, methodology changes
- **Artifact:** Full party mode artifact + Decision Board entry + ADR
- **Examples:** "Should we adopt a new persistence layer?", "How should agent governance work?"

## Sprint Cadence (ADR-0031)

Biweekly (2-week) sprints provide rhythm without heavy process:

- **Sprint start:** Scope statement + `/reconcile-docs` for clean baseline
- **Sprint end:** PM audit (`/reconcile-docs`) + retrospective if meaningful work completed
- **During sprint:** Course corrections via `/course-correct` as needed

This is NOT Scrum. No velocity tracking, no standups, no burndown charts. The cadence serves human-agent team collaboration — giving the human predictable checkpoints to review and course-correct.

## Work Tracking (ADR-0032)

- **BMAD epic/story files are the source of truth** for work tracking
- **ROADMAP.md is a derived view** — synced from BMAD files for multiclaude scope checking
- **`/reconcile-docs`** handles the sync between BMAD files and ROADMAP.md
- Story files live in `docs/stories/`, epic definitions in `docs/prd/`

## Doc Maintenance Rules

Planning documents drift is the #1 governance problem at scale. These rules prevent it:

1. **Every PR updates its own story file status** — do not batch updates across PRs
2. **Every PR updates ROADMAP.md** if it completes an epic
3. **Decision Board entries before PR submission** — not after
4. **Story files are the source of truth** — planning docs sync to them, not vice versa
5. **Run `/reconcile-docs` periodically** to detect and fix drift

## Course Correction Pattern

When something goes wrong mid-sprint:

1. **Detect** — agent or human identifies the problem
2. **Analyze** — root cause, blast radius, repeat check
3. **Propose** — sprint change proposal (`/course-correct`)
4. **Validate** — optional party mode review
5. **Execute** — create stories, update planning docs
6. **Codify** — if the problem was systemic, add a rule to prevent recurrence

This pattern has been used 5+ times in 6 days of active development. The key insight: **codify as infrastructure** (commands, checks, CI rules) rather than as instructions (markdown rules agents may forget).

## Executable Over Instructional

Rules encoded as:
- CI checks → always enforced
- Custom commands (`/reconcile-docs`, `/course-correct`) → easy to run consistently
- Agent authority sections → clear boundaries

Are more reliable than rules encoded as:
- Markdown paragraphs in CLAUDE.md → agents may miss or misinterpret
- Verbal agreements → lost across context windows
- PR description templates → easy to skip

When adding a new governance rule, prefer making it executable (a check, a command, a gate) over making it instructional (a paragraph in a doc).
