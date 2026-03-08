# ADR-0025: Story-Driven Development Process

- **Status:** Accepted
- **Date:** 2026-01-15
- **Decision Makers:** Project founder
- **Related PRs:** #48, #130, #146
- **Related ADRs:** ADR-0026 (Self-Driving Development Pipeline)

## Context

As the project grew beyond the initial tech demo, development needed structure to ensure changes were traceable, testable, and reviewable. The project uses AI agents (multiclaude workers) for implementation, making clear specifications essential.

## Decision

Adopt **mandatory story-driven development**:

1. Every implementation task must have a `docs/stories/X.Y.story.md` file before work begins
2. Story files contain acceptance criteria, technical approach, testing requirements, and dependencies
3. Code cannot be committed without verifying ACs and updating story status
4. After implementation, story status is updated to `Done (PR #NNN)`
5. Research, spikes, and documentation are exempt but should reference stories when possible

## Supporting Tools

- `/implement-story` reusable workflow command (PR #48) — standardized implementation flow
- Pre-PR submission checklist added to all story files (PR #32, #33)
- Story-driven development rule codified in CLAUDE.md (PR #130)
- Baseline test phase added to implementation workflow (PR #146)

## Rationale

- AI workers need clear, unambiguous specifications
- Acceptance criteria provide binary pass/fail validation
- Story files serve as living documentation of what was built and why
- PR-story mapping enables full traceability from requirement to implementation
- Status tracking in story files provides sprint visibility

## Consequences

### Positive
- Clear contract between specification and implementation
- AI workers can autonomously implement stories with confidence
- Full traceability: PRD → Epic → Story → PR → Code
- Sprint status derivable from story file statuses

### Negative
- Story creation overhead for small changes
- Story files must be maintained even after implementation
- Risk of story/code drift if updates are forgotten
