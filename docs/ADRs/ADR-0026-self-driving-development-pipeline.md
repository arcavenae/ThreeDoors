# ADR-0026: Self-Driving Development Pipeline

- **Status:** Accepted
- **Date:** 2026-03-01
- **Decision Makers:** Design decision H7, architecture review
- **Related PRs:** #135, #141, #149, #152, #159-#164
- **Related ADRs:** ADR-0025 (Story-Driven Development)

## Context

ThreeDoors uses AI agents (multiclaude workers) for story implementation. Manual story dispatch is a bottleneck. Epic 22 explored automating the dispatch-implement-review cycle.

## Considered Options

1. **Shell script MVP** — Bash script parsing story files and dispatching `multiclaude worker create`
2. **GitHub Actions** — Trigger workers from CI on story file changes
3. **Supervisor enhancement** — Extend multiclaude supervisor with story awareness
4. **`multiclaude pipeline` command** — First-class pipeline support in multiclaude

## Decision

**Shell script MVP** (Option A) as the immediate approach, with the TUI providing dispatch and monitoring capabilities.

## Implementation (Epic 22 — 8 stories)

| Story | Component | PR |
|-------|-----------|-----|
| 22.1 | Dispatch data model and queue persistence | #149 |
| 22.2 | Dispatch engine with multiclaude CLI wrapper | #152 |
| 22.3 | TUI dispatch key binding and confirmation | #163 |
| 22.4 | Dev queue view (list, approve, kill) | #162 |
| 22.5 | Worker status polling and task update loop | #161 |
| 22.6 | Auto-generated review and follow-up tasks | #164 |
| 22.7 | Optional story file generation | #159 |
| 22.8 | Safety guardrails (rate limiting, cost caps, audit) | #160 |

## Rationale

- Shell script approach works immediately with existing multiclaude infrastructure
- TUI integration provides human-in-the-loop oversight
- Safety guardrails (rate limiting, cost caps, audit logging) prevent runaway costs
- Dev queue view gives visibility into active workers

## Consequences

### Positive
- Reduced manual overhead for story dispatch
- Human approval required before dispatch (safety)
- Cost and rate limiting prevent accidental overuse
- Audit log provides full history of automated actions

### Negative
- Shell script approach is fragile for complex workflows
- Requires multiclaude to be running and configured
- Worker failures need manual investigation
- Cost caps are advisory — actual API costs depend on worker behavior
