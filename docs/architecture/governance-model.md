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

## Decision Tiers

Not all decisions need the same rigor:

| Tier | Examples | Process |
|------|----------|---------|
| **Trivial** | Variable naming, import order, test structure | Agent decides autonomously |
| **Local** | Implementation approach within a story, error handling strategy | Agent decides, documents in PR |
| **Architectural** | New interfaces, data model changes, dependency additions | Party mode → Decision Board → Story |
| **Strategic** | New epics, roadmap changes, methodology changes | Human decides after party mode input |

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
