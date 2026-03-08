---
stepsCompleted: ["step-01-init", "step-02-context", "step-03-decisions", "step-04-components", "step-05-review"]
inputDocuments:
  - docs/prd/requirements.md (NFR-DX1 through NFR-DX5)
  - docs/research/ai-tooling-findings.md (7 findings, research spike)
  - docs/prd/epic-list.md (Epic 34 added)
  - CLAUDE.md (existing coding standards)
  - _bmad-output/planning-artifacts/sprint-change-proposal-2026-03-08-soul-skills.md
workflowType: 'architecture'
project_name: 'ThreeDoors'
user_name: 'arcaven'
date: '2026-03-08'
---

# Architecture Decision Document: Epic 34 — SOUL.md + Custom Development Skills

## 1. Overview

Epic 34 adds two developer-facing capabilities to ThreeDoors:

1. **SOUL.md** — A project philosophy document that captures the "why" behind ThreeDoors, complementing CLAUDE.md which captures the "how"
2. **Custom Claude Code Slash Commands** — Four project-specific commands (`/pre-pr`, `/validate-adapter`, `/check-patterns`, `/new-story`) that automate common development workflows

**Key constraint:** This is a documentation/configuration-only epic. No Go code changes, no new packages, no architectural changes to the application runtime.

## 2. Project Context

### Current State

| Artifact | Status | Role |
|----------|--------|------|
| CLAUDE.md | Exists | Coding standards, pre-PR checklist, Go idioms, design patterns |
| SOUL.md | Does not exist | (Proposed) Project philosophy and AI agent behavioral guidelines |
| .claude/commands/ | 50+ BMAD commands | BMAD framework skills (planning, analysis, QA, etc.) |
| ThreeDoors-specific commands | 0 | (Proposed) /pre-pr, /validate-adapter, /check-patterns, /new-story |

### Problem

- AI agents make decisions without consistent philosophical alignment
- 11 story files contain ~500 lines of duplicated content (checklists, standards, patterns)
- 12+ PRs had preventable CI failures that a `/pre-pr` command would have caught
- No project-specific development automation despite extensive BMAD infrastructure

## 3. Architectural Decisions

### AD-34.1: SOUL.md Placement and Relationship to CLAUDE.md

**Decision:** Place SOUL.md at project root alongside CLAUDE.md. Establish a clear separation:
- **CLAUDE.md** = "How" — coding standards, technical rules, pre-PR checklists
- **SOUL.md** = "Why" — project philosophy, design principles, behavioral guidelines

**Rationale:** Claude Code automatically loads CLAUDE.md. SOUL.md is referenced from CLAUDE.md so agents see both. The separation keeps each document focused and scannable. CLAUDE.md stays action-oriented (rules to follow); SOUL.md stays values-oriented (principles for ambiguous decisions).

**Integration:** CLAUDE.md will include a reference: `See SOUL.md for project philosophy and design principles.`

### AD-34.2: Slash Command Location and Naming

**Decision:** Place all custom commands in `.claude/commands/` as standard Claude Code slash commands with descriptive names:
- `.claude/commands/pre-pr.md`
- `.claude/commands/validate-adapter.md`
- `.claude/commands/check-patterns.md`
- `.claude/commands/new-story.md`

**Rationale:** This is the standard Claude Code convention. Commands placed here are automatically available as `/pre-pr`, `/validate-adapter`, etc. They work in any Claude Code session (standalone or multiclaude). No special infrastructure needed.

**Note:** These coexist with 50+ existing BMAD commands in the same directory. The naming convention (no `bmad-` prefix) distinguishes ThreeDoors-specific commands from BMAD framework commands.

### AD-34.3: Remote Reference Convention

**Decision:** All commands reference `origin/main` (not `upstream/main`).

**Rationale:** ThreeDoors switched from fork workflow to direct push (2026-03-07). The `/pre-pr` command must use `origin/main` for branch freshness checks, scope review, and rebase validation.

### AD-34.4: Living Documentation — Specs Must Reflect Code Reality

**Decision:** Completed story files and specs MUST be updated retroactively when code improvements, architectural changes, or lessons learned diverge from what the specs describe. The `/new-story` template applies to future stories, AND existing completed stories should be updated to remove duplicated content and reflect current project standards.

**Rationale:** Learning and improvements captured only in code — not reflected back into specs — is an anti-pattern. The spec is the authoritative description of what the system does and why. If you deleted all code and rebuilt from specs alone, the result should be a *better* program, not a regression. Forward-only policies create spec drift: over time, completed stories become misleading historical artifacts rather than useful documentation. Git blame noise is a trivial cost compared to specs that no longer describe reality.

**Implications:**
- When a code improvement changes behavior described in a completed story, update the story
- When DRYing content (e.g., removing embedded checklists in favor of CLAUDE.md references), update all stories — not just new ones
- Story status remains "Done (PR #NNN)" — the implementation is complete, the documentation is living
- Retroactive updates are a separate PR from the original implementation (clean git history)

### AD-34.5: MCP Prompt Template Integration Point

**Decision:** Document SOUL.md as an input source for Epic 24.8 (MCP Prompt Templates & Advanced Interaction Patterns). SOUL.md's philosophy — particularly the coaching tone ("a friend saying pick one and go") — should inform MCP prompt template design.

**Rationale:** MCP prompts need personality alignment. SOUL.md provides the authoritative source for ThreeDoors' voice and interaction philosophy. Story 24.8 can reference SOUL.md directly rather than embedding philosophy ad hoc.

**No action required in Epic 34.** This is a documented integration point for future work.

## 4. Component Design

### 4.1: SOUL.md Content Structure

```
SOUL.md
├── What Is ThreeDoors? (one-paragraph identity)
├── Core Philosophy
│   ├── Progress Over Perfection
│   ├── Work With Human Nature
│   ├── Three Doors, Not Three Hundred
│   ├── Local-First, Privacy-Always
│   ├── Meet Users Where They Are
│   └── Solo Dev Reality
├── Design Principles for AI Agents (5 decision heuristics)
├── What ThreeDoors Is NOT (anti-patterns)
└── The Feeling We're Going For (emotional target)
```

**Source:** Content derived from docs/research/ai-tooling-findings.md Finding 2 (proposed SOUL.md). Reviewed and approved in party mode.

### 4.2: /pre-pr Command Design

**8-step sequential validation:**

| Step | Check | Tool | Pass Condition |
|------|-------|------|---------------|
| 1 | Branch freshness | `git fetch origin main` + log | ≤5 commits behind |
| 2 | Formatting | `gofumpt -l .` | No output |
| 3 | Linting | `golangci-lint run ./...` | 0 issues |
| 4 | Tests | `go test ./... -count=1` | All pass |
| 5 | Race detection | `go test -race ./...` | No races |
| 6 | Dead code | `go vet ./...` | Clean |
| 7 | Scope review | `git diff --stat origin/main...HEAD` | Show file list, warn on out-of-scope |
| 8 | Commit cleanliness | `git log --oneline origin/main..HEAD` | No fixup/wip/squash messages |

**Output:** Summary table with pass/fail per step. Final verdict: "Ready to push" or "Fix issues above."

**Design note:** Step 5 (`go test -race`) added per QA agent recommendation in party mode. This catches race conditions that standard `go test` misses.

### 4.3: /validate-adapter Command Design

**Validation sequence:**

1. Read `internal/tasks/provider.go` for current `TaskProvider` interface
2. Find all implementing types via method signature grep
3. For each implementation, verify:
   - All interface methods implemented
   - Error wrapping uses `%w` pattern
   - Type registered in factory switch statement
   - Corresponding `_test.go` file exists
   - File-based providers use atomic write pattern

**Output:** Compliance table per adapter.

### 4.4: /check-patterns Command Design

**6 pattern violation categories:**

| Category | Detection Method |
|----------|-----------------|
| Direct status mutation | Grep for `.Status =` outside `task_status.go` |
| Non-atomic file writes | Grep for `os.WriteFile`/`ioutil.WriteFile` not targeting `.tmp` |
| fmt.Println in TUI | Grep for `fmt.Print` in `internal/tui/` |
| Panics in user code | Grep for `panic(` in `internal/` |
| Factory bypass | Grep for direct provider construction outside factory/tests |
| Missing error wrapping | Grep for bare `return.*err$` without `%w` |

**Output:** Findings grouped by category with file:line references.

### 4.5: /new-story Command Design

**Story template structure:**

```markdown
<!--
title: "X.Y: Story Title"
status: "Draft"
-->

# Story X.Y: Title

**As a** user,
**I want** ...,
**so that** ...

## Acceptance Criteria
(numbered ACs with Given/When/Then)

## NOT In Scope

## Definition of Done
- All standard checks pass (see CLAUDE.md Pre-PR Checklist)
- [Story-specific DoD items only]

## Architecture & Design
(data models, view modes, files table)

## Key Files to Create/Modify
| File | Action |

## Testing
(Story-specific test requirements)

## Tasks / Subtasks
- [ ] Task 1...

## Dev Agent Record
## QA Results
```

**Key change from current template:** No embedded Pre-PR Submission Checklist, no repeated coding standards, no pattern reminders. All reference CLAUDE.md.

## 5. File Inventory

| File | Action | Description |
|------|--------|-------------|
| `SOUL.md` | Create | Project philosophy document |
| `.claude/commands/pre-pr.md` | Create | Pre-PR validation command |
| `.claude/commands/validate-adapter.md` | Create | Adapter compliance command |
| `.claude/commands/check-patterns.md` | Create | Pattern violation scanner |
| `.claude/commands/new-story.md` | Create | Story template generator |
| `CLAUDE.md` | Edit | Add reference to SOUL.md |
| `docs/stories/*.story.md` | Edit | Retroactive DRY — remove embedded checklists, align with code reality |

**Total files:** 5 new, ~12 edited (CLAUDE.md + completed stories). All markdown. Zero Go code changes.

## 6. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Skill bash commands have bugs | Medium | Low | Fix in subsequent PR; no production impact |
| SOUL.md contradicts CLAUDE.md | Low | Low | Review both documents for consistency |
| Agents ignore SOUL.md | Low | Low | Reference from CLAUDE.md ensures loading |
| /pre-pr produces false failures | Medium | Low | Run manually, fix command definitions |

**Overall risk: Low.** No code changes means zero regression risk. Worst case: a slash command has a typo in a bash snippet.

## 7. NFR Traceability

| NFR | Architectural Decision | Component |
|-----|----------------------|-----------|
| NFR-DX1 | AD-34.1 (SOUL.md placement) | SOUL.md |
| NFR-DX2 | AD-34.2 (command location) | /pre-pr |
| NFR-DX3 | AD-34.2 (command location) | /validate-adapter |
| NFR-DX4 | AD-34.2 (command location) | /check-patterns |
| NFR-DX5 | AD-34.2 (command location) | /new-story |
| NFR-DX6 | AD-34.4 (living documentation) | Retroactive story updates |

## 8. Integration Points (Future Epics)

| Epic | Integration | Status |
|------|------------|--------|
| Epic 24.8 (MCP Prompt Templates) | SOUL.md philosophy informs coaching tone in prompts | Documented, not yet started |
| Epic 25-26, 30 (New Integrations) | /validate-adapter checks new adapter implementations | Available immediately |
| All future stories | /new-story generates DRY templates; /pre-pr catches CI issues | Available immediately |
