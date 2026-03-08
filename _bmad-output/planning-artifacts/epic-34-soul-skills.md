---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/requirements.md (NFR-DX1 through NFR-DX5)
  - docs/prd/epic-list.md (Epic 34 entry)
  - _bmad-output/planning-artifacts/architecture-soul-skills.md
  - _bmad-output/planning-artifacts/sprint-change-proposal-2026-03-08-soul-skills.md
  - docs/research/ai-tooling-findings.md
  - CLAUDE.md
---

# ThreeDoors - Epic 34 Breakdown: SOUL.md + Custom Development Skills

## Overview

This document provides the epic and story breakdown for Epic 34, which adds project philosophy documentation (SOUL.md) and four custom Claude Code slash commands to improve AI agent alignment and developer workflow. Based on research findings (docs/research/ai-tooling-findings.md) and party mode consensus (2026-03-08).

**Key constraint:** Documentation/configuration only. Zero Go code changes.

## Requirements Inventory

### Non-Functional Requirements (Developer Experience)

- **NFR-DX1:** The project shall maintain a SOUL.md document at the project root defining the project's philosophy, design principles, and behavioral guidelines for AI agents — ensuring consistent decision-making aligned with ThreeDoors values (progress over perfection, work with human nature, three doors not three hundred, local-first privacy-always, meet users where they are)
- **NFR-DX2:** The project shall provide a `/pre-pr` Claude Code slash command that automates an 8-step pre-PR validation checklist (branch freshness, formatting via `gofumpt`, linting via `golangci-lint`, tests via `go test`, race detection via `go test -race`, dead code via `go vet`, scope review via `git diff`, commit cleanliness check)
- **NFR-DX3:** The project shall provide a `/validate-adapter` Claude Code slash command that checks TaskProvider implementations for interface compliance, error wrapping patterns, factory registration, test coverage, and atomic write usage
- **NFR-DX4:** The project shall provide a `/check-patterns` Claude Code slash command that scans the codebase for design pattern violations
- **NFR-DX5:** The project shall provide a `/new-story` Claude Code slash command that generates story files from a standard template, referencing CLAUDE.md for coding standards

### Additional Requirements (from Architecture)

- SOUL.md placed at project root, referenced from CLAUDE.md (AD-34.1)
- All commands in `.claude/commands/` following Claude Code convention (AD-34.2)
- Use `origin/main` not `upstream/main` in all commands (AD-34.3)
- Living documentation — completed stories MUST be updated when code diverges from specs (AD-34.4)
- Document MCP integration point for Epic 24.8 (AD-34.5)

### NFR Coverage Map

| NFR | Story | Coverage |
|-----|-------|----------|
| NFR-DX1 | 34.1 | SOUL.md creation and CLAUDE.md reference |
| NFR-DX2 | 34.2 | /pre-pr command with 8 validation steps |
| NFR-DX3 | 34.2 | /validate-adapter command |
| NFR-DX4 | 34.2 | /check-patterns command |
| NFR-DX5 | 34.2 | /new-story command with DRY template |
| NFR-DX6 | 34.4 | Retroactive story updates and spec alignment |

## Epic List

| Epic | Title | Stories | Priority | Prerequisites |
|------|-------|---------|----------|---------------|
| 34 | SOUL.md + Custom Development Skills | 4 | P1 | None |

---

## Epic 34: SOUL.md + Custom Development Skills

**Goal:** Create SOUL.md project philosophy document and 4 custom Claude Code slash commands (/pre-pr, /validate-adapter, /check-patterns, /new-story) to improve AI agent alignment and developer workflow automation.

**Rationale:** Research identified ~500 lines of duplicated content across 11 stories, 12+ PRs with preventable CI failures, and zero ThreeDoors-specific development commands despite 50+ BMAD commands. This epic addresses all three gaps with documentation/configuration files only — no code changes.

**Prerequisites:** None (CLAUDE.md already exists with coding standards)

**Estimated Effort:** 1-2 days

---

### Story 34.1: Create SOUL.md Project Philosophy Document

**As a** developer or AI agent working on ThreeDoors,
**I want** a SOUL.md document that captures the project's philosophy and design principles,
**So that** I can make aligned decisions on ambiguous design choices without per-story guidance.

**Acceptance Criteria:**

**AC 34.1.1:**
**Given** the project root directory,
**When** I look for SOUL.md,
**Then** it exists with the following sections: "What Is ThreeDoors?", "Core Philosophy" (with 6 sub-principles), "Design Principles for AI Agents" (with 5 decision heuristics), "What ThreeDoors Is NOT", and "The Feeling We're Going For"

**AC 34.1.2:**
**Given** SOUL.md exists,
**When** I read the Core Philosophy section,
**Then** it contains principles for: Progress Over Perfection, Work With Human Nature, Three Doors Not Three Hundred, Local-First Privacy-Always, Meet Users Where They Are, and Solo Dev Reality

**AC 34.1.3:**
**Given** SOUL.md exists,
**When** I read the Design Principles for AI Agents section,
**Then** it contains 5 decision heuristics: friction reduction, simplicity, data respect, pattern following, and user-visibility check

**AC 34.1.4:**
**Given** SOUL.md exists at project root,
**When** I read CLAUDE.md,
**Then** CLAUDE.md contains a reference to SOUL.md (e.g., "See SOUL.md for project philosophy and design principles")

**AC 34.1.5:**
**Given** SOUL.md content,
**When** compared against the PRD executive summary and product brief,
**Then** the philosophy is consistent with — not contradicting — existing product documentation

**Tasks:**
- [ ] Create SOUL.md at project root with all sections from architecture doc (AD-34.1)
- [ ] Content sourced from docs/research/ai-tooling-findings.md Finding 2
- [ ] Add SOUL.md reference to CLAUDE.md
- [ ] Verify consistency with PRD goals and product brief

**Key Files:**
| File | Action |
|------|--------|
| `SOUL.md` | Create |
| `CLAUDE.md` | Edit (add reference) |

**NFRs covered:** NFR-DX1

---

### Story 34.2: Create Custom Claude Code Slash Commands

**As a** developer or AI agent working on ThreeDoors,
**I want** project-specific Claude Code slash commands for common workflows,
**So that** I can automate pre-PR validation, adapter compliance checks, pattern enforcement, and story creation.

**Acceptance Criteria:**

**AC 34.2.1 — /pre-pr command:**
**Given** the `.claude/commands/pre-pr.md` file exists,
**When** an agent invokes `/pre-pr`,
**Then** it runs 8 sequential checks: (1) branch freshness via `git fetch origin main`, (2) formatting via `gofumpt -l .`, (3) linting via `golangci-lint run ./...`, (4) tests via `go test ./... -count=1`, (5) race detection via `go test -race ./...`, (6) dead code via `go vet ./...`, (7) scope review via `git diff --stat origin/main...HEAD`, (8) commit cleanliness via `git log --oneline origin/main..HEAD`
**And** reports a summary table with pass/fail per step
**And** provides a final verdict: "Ready to push" or "Fix issues above"

**AC 34.2.2 — /validate-adapter command:**
**Given** the `.claude/commands/validate-adapter.md` file exists,
**When** an agent invokes `/validate-adapter`,
**Then** it reads the TaskProvider interface from `internal/tasks/provider.go`, finds all implementing types, and verifies for each: all interface methods implemented, error wrapping uses `%w`, type registered in factory, `_test.go` file exists, and file-based providers use atomic writes
**And** reports a compliance table per adapter

**AC 34.2.3 — /check-patterns command:**
**Given** the `.claude/commands/check-patterns.md` file exists,
**When** an agent invokes `/check-patterns`,
**Then** it scans for 6 violation categories: direct status mutation (`.Status =` outside task_status.go), non-atomic file writes, `fmt.Println` in TUI code, panics in user code, provider instantiation outside factory/tests, and bare error returns without `%w`
**And** reports findings grouped by category with file:line references

**AC 34.2.4 — /new-story command:**
**Given** the `.claude/commands/new-story.md` file exists,
**When** an agent invokes `/new-story`,
**Then** it prompts for story ID and title, uses the latest story format as reference, creates `docs/stories/{id}.story.md` with standard structure (frontmatter, user story, ACs, NOT In Scope, DoD referencing CLAUDE.md, Architecture, Key Files, Testing, Tasks, Dev Agent Record, QA Results)
**And** the generated template does NOT include an embedded Pre-PR Submission Checklist
**And** the generated template does NOT include duplicated coding standards or pattern reminders

**AC 34.2.5 — Remote references:**
**Given** any slash command that references a git remote,
**When** comparing remote URLs in the command definitions,
**Then** all commands use `origin/main` (not `upstream/main`)

**Tasks:**
- [ ] Create `.claude/commands/pre-pr.md` with 8-step validation (including -race per party mode)
- [ ] Create `.claude/commands/validate-adapter.md` with adapter compliance checks
- [ ] Create `.claude/commands/check-patterns.md` with 6 pattern violation categories
- [ ] Create `.claude/commands/new-story.md` with DRY story template
- [ ] Verify all commands use `origin/main` (AD-34.3)

**Key Files:**
| File | Action |
|------|--------|
| `.claude/commands/pre-pr.md` | Create |
| `.claude/commands/validate-adapter.md` | Create |
| `.claude/commands/check-patterns.md` | Create |
| `.claude/commands/new-story.md` | Create |

**NFRs covered:** NFR-DX2, NFR-DX3, NFR-DX4, NFR-DX5

---

### Story 34.3: Story Template Update and Integration Notes

**As a** developer or AI agent creating new stories,
**I want** the story template to reference CLAUDE.md instead of embedding standards,
**So that** future stories are leaner and maintainable from a single source of truth.

**Acceptance Criteria:**

**AC 34.3.1:**
**Given** the `/new-story` command from Story 34.2,
**When** comparing its output template to completed story files,
**Then** the template's Definition of Done section contains only "All standard checks pass (see CLAUDE.md Pre-PR Checklist)" plus story-specific items — no embedded Pre-PR Submission Checklist block

**AC 34.3.2:**
**Given** the `/new-story` command template,
**When** reviewing for duplicated content,
**Then** the template does NOT contain repeated Dev Notes about: atomic write pattern, error wrapping with %w, MVU pattern, "follow existing patterns in X", or historical PR references — all of which are already in CLAUDE.md

**AC 34.3.3:**
**Given** Epic 24.8 (MCP Prompt Templates) is pending,
**When** reviewing Epic 34 deliverables,
**Then** the architecture document (_bmad-output/planning-artifacts/architecture-soul-skills.md) contains an integration note documenting that SOUL.md philosophy should inform MCP prompt template design (AD-34.5)

**Tasks:**
- [ ] Verify /new-story template from 34.2 meets DRY criteria (AC 34.3.1, 34.3.2)
- [ ] Verify architecture doc contains MCP integration note (AC 34.3.3)

**Key Files:**
| File | Action |
|------|--------|
| `.claude/commands/new-story.md` | Verify (created in 34.2) |
| `_bmad-output/planning-artifacts/architecture-soul-skills.md` | Verify (created in architecture step) |

**NFRs covered:** NFR-DX5 (validation)

---

### Story 34.4: Retroactive Story DRY & Spec Alignment

**As a** developer or AI agent reading completed stories for context,
**I want** completed story files updated to remove duplicated content and reflect current project standards,
**So that** specs remain an accurate, authoritative description of the system — enabling a "delete all code, rebuild from specs" workflow that produces a better program.

**Acceptance Criteria:**

**AC 34.4.1:**
**Given** completed story files in `docs/stories/*.story.md`,
**When** reviewing for embedded Pre-PR Submission Checklists,
**Then** all embedded checklists are replaced with a reference to CLAUDE.md (e.g., "See CLAUDE.md Pre-PR Checklist")

**AC 34.4.2:**
**Given** completed story files,
**When** reviewing for duplicated coding standards content (atomic write pattern descriptions, error wrapping reminders, MVU pattern explanations, "follow existing patterns in X"),
**Then** duplicated content is removed and replaced with CLAUDE.md references where the content is already covered

**AC 34.4.3:**
**Given** completed story files,
**When** comparing story descriptions to the current codebase behavior,
**Then** any story whose described behavior has been superseded by later improvements is updated to reflect current reality (acceptance criteria, architecture notes, key files)

**AC 34.4.4:**
**Given** a completed story file that has been updated,
**When** reviewing its status field,
**Then** the status still reads "Done (PR #NNN)" — retroactive spec updates do not change completion status

**AC 34.4.5:**
**Given** the retroactive updates,
**When** reviewing the git history,
**Then** retroactive spec updates are in a separate PR from original implementation PRs (clean git history)

**Tasks:**
- [ ] Audit all completed story files for embedded Pre-PR Checklists (~11 files, ~500 lines)
- [ ] Replace embedded checklists with CLAUDE.md references
- [ ] Remove duplicated coding standards content
- [ ] Review stories for code-reality divergence and update descriptions
- [ ] Verify story statuses remain unchanged
- [ ] Create PR with retroactive updates

**Key Files:**
| File | Action |
|------|--------|
| `docs/stories/*.story.md` | Edit (all completed stories with duplicated content) |

**NFRs covered:** NFR-DX6 (living documentation)

---

## Story Dependencies

```
34.1 (SOUL.md) ──────────────── No dependencies
34.2 (Custom Skills) ────────── No dependencies (can run in parallel with 34.1)
34.3 (Template Verification) ── Depends on 34.2 (verifies /new-story output)
34.4 (Retroactive DRY) ──────── Depends on 34.2 (needs /new-story template as reference for target format)
```

**Parallelism:** Stories 34.1 and 34.2 can be implemented in parallel. Stories 34.3 and 34.4 run after 34.2. Stories 34.3 and 34.4 can run in parallel with each other.

## Implementation Notes

1. **No code changes** — All deliverables are markdown files
2. **No test infrastructure** — Slash commands are tested by running them manually
3. **Living documentation** — Completed stories MUST be updated when code diverges from specs. Specs are the authoritative system description. If you could delete all code and rebuild from specs alone, the result must be a better program, not a regression. Learning captured only in code is an anti-pattern.
4. **Remote convention** — Use `origin/main` throughout (direct push, not fork)
5. **MCP integration** — SOUL.md is a documented future input for Epic 24.8 prompt templates
