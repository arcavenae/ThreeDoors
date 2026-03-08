# ADR-0031: Biweekly Sprint Cadence for Agent Team Collaboration

- **Status:** Accepted
- **Date:** 2026-03-08
- **Decision Makers:** Project founder
- **Related PRs:** #265
- **Related ADRs:** ADR-0025 (Story-Driven Development), ADR-0029 (Governance Phase Renaming)

## Context

ThreeDoors has no sprint concept. Work happens continuously with no natural checkpoints for doc freshness audits, retrospectives, or planning. The project is not a solo dev effort — it's a solo human directing a team of AI agents. This distinction matters: agent teams need rhythm and coordination checkpoints the same way human teams do.

Without cadence, the "build fast, catch up later" pattern emerges. Planning docs drift, retrospectives don't happen, and course corrections are reactive rather than proactive.

## Decision

Adopt biweekly (2-week) sprint boundaries with lightweight ceremonies:

### Sprint Start
- Brief scope statement: what epics/stories are in focus
- `/reconcile-docs` run to start with clean planning doc state

### Sprint End
- PM audit of doc freshness (`/reconcile-docs`)
- Retrospective trigger (`/bmad-bmm-retrospective`) if meaningful work was completed
- Review of Decision Board for any unrecorded decisions

### During Sprint
- Course corrections via `/course-correct` as needed
- No artificial urgency — if a sprint is quiet, ceremonies are brief

## Rationale

- Agent teams need coordination checkpoints just like human teams
- Biweekly is short enough to catch drift early, long enough to avoid ceremony overhead
- This is NOT Scrum — no velocity tracking, no standups, no burndown charts
- The human's role is directing and deciding, not coding — cadence helps the human review and course-correct at predictable intervals
- If no meaningful work happened in a sprint, ceremonies take 5 minutes

## Important Framing

This project is a solo human figuring out human-agent team collaboration. The sprint cadence serves the collaboration model, not traditional project management. As the agent governance model matures, the cadence may evolve.

## Rejected Alternatives

- **Epic-based sprints:** Uneven rhythm — some epics have 11 stories, others have 1. No predictable checkpoint.
- **PR-count sprints (every 25 PRs):** Arbitrary boundary with no semantic meaning. A burst of docs PRs triggers the same as feature PRs.
- **Event-triggered only:** Zero guaranteed rhythm. Easy to "forget" audits for long stretches. Doesn't serve the team collaboration model.
