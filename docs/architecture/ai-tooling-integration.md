# AI Tooling Integration Notes

How SOUL.md, CLAUDE.md, custom slash commands, and story files work together to guide development in ThreeDoors.

## Audience

This document is for any developer or AI agent working on ThreeDoors. Read this to understand where project guidance lives and how to write lean story files.

## Layered Information Model

ThreeDoors uses a four-layer information architecture. Each layer has a distinct purpose — do not duplicate content across layers.

| Layer | File(s) | Purpose | Contains |
|-------|---------|---------|----------|
| **Why** | `SOUL.md` | Philosophy and values | Design principles, what ThreeDoors is/isn't, the intended user experience |
| **How** | `CLAUDE.md` | Technical standards and rules | Go idioms, error handling, testing standards, design patterns, formatting/linting rules |
| **Do** | `.claude/commands/*.md` | Workflow automation | Pre-PR validation, pattern checks, adapter compliance, story generation |
| **What** | `docs/stories/*.story.md` | Task-specific context | Acceptance criteria, architecture decisions, files to modify, story-specific testing |

**Key principle:** Information flows downward. Story files can reference CLAUDE.md and SOUL.md but should never repeat their content. Slash commands encode the workflows described in CLAUDE.md into executable steps.

## Custom Slash Commands

Four ThreeDoors-specific slash commands live in `.claude/commands/`:

| Command | File | Purpose | When to Use |
|---------|------|---------|-------------|
| `/pre-pr` | `pre-pr.md` | Runs the full pre-PR validation checklist | Before pushing a branch or creating a PR |
| `/validate-adapter` | `validate-adapter.md` | Checks TaskProvider implementations for completeness | After creating or modifying a storage adapter |
| `/check-patterns` | `check-patterns.md` | Scans for design pattern violations | During code review or before PR submission |
| `/new-story` | `new-story.md` | Generates a story file from the standard template | When starting a new story |

These commands reference CLAUDE.md and SOUL.md without duplicating their content. Each command is a markdown instruction file that Claude Code executes as a workflow.

## The DRY Story Approach

### What Goes in Story Files

Story files contain **only story-specific information**:

- User story (As a / I want / So that)
- Acceptance criteria specific to this story
- What's NOT in scope
- Architecture and design decisions unique to this story
- Files to create or modify
- Story-specific testing requirements
- Task breakdown

### What Stories Inherit from CLAUDE.md

Stories do **not** need to include:

- Pre-PR submission checklist (use `/pre-pr` instead)
- Coding standards (Go idioms, error handling patterns, etc.)
- Design pattern reminders (atomic writes, MVU, factory pattern, etc.)
- Testing standards (table-driven tests, t.Helper(), etc.)
- Formatting and linting rules (gofumpt, golangci-lint)
- Historical PR references or lessons learned

The Quality Gate section in stories should say "All standard checks pass (see CLAUDE.md)" plus any story-specific quality items.

### What Stories Reference from SOUL.md

When a story involves a design decision where philosophy matters (e.g., "should we add this option?"), reference SOUL.md principles rather than restating them. For example: "Per SOUL.md's 'Three Doors, Not Three Hundred' principle, this view should not expose additional configuration."

## For Story Authors

When writing a new story (or using `/new-story`):

1. **Start with the user story** — who benefits and why
2. **Write specific acceptance criteria** — testable, unambiguous
3. **Define what's NOT in scope** — prevent scope creep
4. **Keep Quality Gate lean** — reference CLAUDE.md, add only story-specific items
5. **Focus Architecture & Design on what's new** — don't restate patterns from CLAUDE.md
6. **Testing section = story-specific only** — CLAUDE.md covers the standards

If you find yourself writing something that applies to all stories (not just this one), it probably belongs in CLAUDE.md instead.

## Related Documents

- [SOUL.md](../../SOUL.md) — Project philosophy and design values
- [CLAUDE.md](../../CLAUDE.md) — Technical standards and project rules
- [Coding Standards](coding-standards.md) — Detailed coding standards reference
- [PR Submission Standards](pr-submission-standards.md) — Detailed PR process and history
- [AI Tooling Findings](../research/ai-tooling-findings.md) — Research that informed this architecture
