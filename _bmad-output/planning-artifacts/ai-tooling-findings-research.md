# AI Tooling & Quality Improvement Findings

**Date:** 2026-03-02
**Author:** zealous-rabbit (research agent)
**Scope:** CLAUDE.md, SOUL.md, Skills/Commands, DRY reduction, quality patterns

---

## Executive Summary

ThreeDoors currently has **no project-level CLAUDE.md** and relies entirely on BMAD agent prompts and per-story checklists to guide AI agents. This means every story must repeat coding standards, pre-PR checklists, and architectural patterns — creating maintenance burden and inconsistent enforcement. The project would benefit significantly from:

1. A **CLAUDE.md** that encodes project rules, Go idioms, and the pre-PR checklist once
2. A **SOUL.md** that captures the app's philosophy so agents make aligned decisions
3. **Custom skills/commands** for common workflows (pre-PR validation, adapter compliance, story creation)
4. **Removing ~40% of boilerplate** from every story file

---

## Finding 1: CLAUDE.md — Project-Level Rules

### Current State

No `CLAUDE.md` exists. The `.claude/` directory contains only BMad slash commands. Every story file embeds its own copy of coding standards and the pre-PR checklist. The result: 11 story files each contain nearly identical Pre-PR Submission Checklist blocks, and dev notes repeat the same patterns (atomic writes, error wrapping, MVU, etc.).

### Proposed CLAUDE.md Structure

```markdown
# ThreeDoors — Project Rules for AI Agents

## Quick Reference
- Language: Go 1.25.4 | TUI: Bubbletea (MVU) | Storage: YAML + text files
- Format: `gofumpt -w .` | Lint: `golangci-lint run ./...` | Test: `go test ./... -count=1`
- Build: `make build` | All checks: `make fmt && make lint && make test`

## Pre-PR Checklist (MANDATORY — do not skip any step)
1. `git fetch upstream main && git rebase upstream/main`
2. `gofumpt -l .` — must produce no output
3. `golangci-lint run ./...` — must report 0 issues
4. `go test ./... -count=1` — must report 0 failures
5. `go vet ./...` — no dead code, unused vars, or suspicious constructs
6. `git diff --stat upstream/main...HEAD` — only story-related files
7. Squash fix-up commits into clean, meaningful commits

## Go Idioms & Coding Standards

### Naming
- Packages: lowercase single word (`tui`, `tasks`, `dist`)
- Files: `snake_case.go` (e.g., `task_pool.go`, `doors_view.go`)
- Exported types/funcs: PascalCase; unexported: camelCase
- Constants: PascalCase (`StatusTodo`, `MaxTasks`)

### Imports — always in this order, separated by blank lines
1. Standard library
2. External packages (`github.com/charmbracelet/...`)
3. Internal packages (`github.com/arcavenae/ThreeDoors/internal/...`)

### Error Handling
- Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Never ignore errors — check every return with `if err != nil`
- No panics in user-facing code (especially `Update()` and `View()` methods)
- User-facing messages should be friendly; log developer details to stderr

### Design Patterns (follow these consistently)

#### Adapter Pattern — TaskProvider Interface
All task storage backends implement `TaskProvider`:
```go
type TaskProvider interface {
    LoadTasks() ([]*Task, error)
    SaveTask(task *Task) error
    SaveTasks(tasks []*Task) error
    DeleteTask(taskID string) error
    MarkComplete(taskID string) error
}
```
Current adapters: `TextFileProvider`, `AppleNotesProvider`, `FallbackProvider`.
New storage backends MUST implement this interface. Do not add provider-specific
methods to calling code — use the interface.

#### Factory Pattern — Provider Creation
Use `NewProviderFromConfig()` to create providers. Add new providers to the
switch statement there, never instantiate providers directly in `main.go`.

#### Strategy Pattern — Door Selection
`DoorSelector` uses a selection strategy to pick 3 tasks. New selection
algorithms plug in here without changing the callers.

#### Decorator Pattern — FallbackProvider
`FallbackProvider` wraps a primary provider with a fallback, implementing
the same `TaskProvider` interface. Use this pattern for cross-cutting concerns
(retry, caching, logging) without modifying the underlying provider.

#### State Machine — Task Status
Status transitions are validated by `IsValidTransition()`. Never set
`task.Status` directly — always validate the transition first.
Valid states: todo → in-progress → in-review → complete (with blocked as a side state).

#### MVU (Model-View-Update) — Bubbletea TUI
- `Init()`: setup, return initial commands
- `Update(msg)`: handle messages, return updated model + commands
- `View()`: pure render function, NO side effects
- Never use `fmt.Println` for TUI output — only `View()` return strings
- All view switching goes through `MainModel` via message passing

#### Repository Pattern — File I/O
`FileManager` handles all YAML persistence. It owns the file format and
atomic write logic. Never read/write task files directly — go through
`FileManager` or `TaskProvider`.

#### Observer Pattern — Bubbletea Messages
Views communicate via typed messages (`SelectDoorMsg`, `TaskUpdatedMsg`, etc.)
defined in `messages.go`. Add new message types there. Views should never
call each other's methods directly.

### Atomic Write Pattern (CRITICAL for data safety)
```go
tempPath := targetPath + ".tmp"
if err := os.WriteFile(tempPath, data, 0644); err != nil {
    return fmt.Errorf("write temp: %w", err)
}
f, _ := os.Open(tempPath)
f.Sync()
f.Close()
os.Rename(tempPath, targetPath) // atomic on POSIX
```

### Testing Standards
- Table-driven tests preferred
- Test files co-located: `foo.go` → `foo_test.go`
- Shared test helpers in `test_helpers_test.go`
- No external mocking frameworks — use interfaces + manual test doubles
- Test names: `TestFunctionName_scenario` (e.g., `TestLoadTasks_emptyFile`)
- Target 70%+ coverage on `internal/tasks/`, pragmatic coverage on `internal/tui/`

### What NOT to Do
- No `fmt.Println` in TUI code — use `View()` methods only
- No panics in `Update()` or `View()`
- No direct file writes — use atomic write pattern
- No mutable task IDs — UUIDs assigned at creation are immutable
- No local timezone — always `time.Now().UTC()`
- No YAML tags without `omitempty` where fields are optional
- No `git add .` without reviewing staged files
- No merge commits — always rebase

## Architecture at a Glance
```
cmd/threedoors/main.go       → Entry point, wiring
internal/tasks/              → Domain layer (providers, models, business logic)
internal/tui/                → Presentation layer (Bubbletea views)
internal/dist/               → Distribution helpers (signing, packaging)
docs/stories/                → Story specs (implementation instructions)
docs/architecture/           → Architecture decision records
```

## Data Directory: ~/.threedoors/
- `tasks.yaml` — Active tasks
- `completed.txt` — Completed task log (append-only)
- `sessions.jsonl` — Session metrics
- `improvements.txt` — Session improvement notes
- `config.yaml` — Provider configuration

## Security
- AppleScript strings MUST be escaped to prevent injection
- No PII in logs or metrics files
- Input validation at system boundaries (file I/O, AppleScript)
```

### Impact Assessment

Putting this in CLAUDE.md means:
- **Every AI agent** sees these rules automatically without story-level repetition
- Pre-PR checklist runs from CLAUDE.md — stories no longer need to embed it
- Design patterns are documented once — agents follow them without per-story reminders
- Estimated **~30-50 lines removed per story** from Definition of Done, Dev Notes, and Checklist sections

---

## Finding 2: SOUL.md — Project Philosophy

### Why a SOUL Document?

Stories currently embed philosophical context like "progress over perfection" and "works with human psychology." A SOUL.md captures the *why* behind ThreeDoors so agents make aligned decisions when stories don't specify every detail.

### Proposed SOUL.md

```markdown
# ThreeDoors — Soul Document

## What Is ThreeDoors?
A personal achievement partner disguised as a todo app. It doesn't manage tasks —
it helps humans *do* things by working with their psychology, not against it.

## Core Philosophy

### Progress Over Perfection
Imperfect action beats perfect planning. The whole point of showing three doors
is to get someone moving — not to optimize their todo list. Every design decision
should reduce friction to starting, not add friction for "correctness."

### Work With Human Nature
Choice paralysis is real. Overwhelm is real. The "I'll do it later" trap is real.
ThreeDoors exists because traditional todo apps ignore these realities. Features
should acknowledge that humans are not rational task-processing machines.

### Three Doors, Not Three Hundred
Show 3 tasks. Not 5. Not "all of them with filters." The constraint IS the feature.
When in doubt, show less. Resist the urge to add "just one more option."

### Local-First, Privacy-Always
- Data stays on the user's machine. Period.
- No telemetry, no analytics, no phone-home.
- No accounts, no sign-up, no cloud sync (unless the user explicitly configures it).
- Apple Notes integration is local AppleScript, not a cloud API.

### Meet Users Where They Are
People already have tasks in Apple Notes, Jira, Linear, text files. ThreeDoors
integrates with existing tools — it doesn't ask users to migrate. The adapter
pattern exists precisely for this: plug into what people already use.

### Solo Dev Reality
This is built by one person in 2-4 hours per week. Every feature must justify
its complexity. Prefer the simple solution that works today over the elegant
solution that takes three sprints. If a feature requires more than one story
to be useful, reconsider the decomposition.

## Design Principles for AI Agents

When implementing a story and you face a decision not covered by the spec:

1. **Would this reduce friction?** If yes, do it. If it adds a step for the user, don't.
2. **Is this the simplest thing that works?** Ship that. Refactor later if needed.
3. **Does this respect the user's data?** No silent writes, no data loss, atomic operations.
4. **Does this follow existing patterns?** Check how similar things are done in the codebase.
   Don't invent new patterns when existing ones work.
5. **Would the user notice this?** If it's an internal refactor with no visible change,
   keep the scope minimal. If it's user-facing, make it feel effortless.

## What ThreeDoors Is NOT
- Not a project management tool (no Gantt charts, no sprints, no team features)
- Not a habit tracker (no streaks, no gamification, no guilt)
- Not a second brain (no knowledge graph, no linking, no tagging taxonomy)
- Not trying to be everything to everyone — it's a personal tool for one person at a time

## The Feeling We're Going For
Opening ThreeDoors should feel like a friend saying: "Hey, here are three things
you could do right now. Pick one. Any one. Let's go."

Not: "You have 47 overdue tasks. Here's a productivity report."
```

---

## Finding 3: Skills & Custom Commands

### Current State

The `.claude/commands/` directory contains 50+ BMad slash commands but **zero ThreeDoors-specific commands**. There are no project-level skills for common development workflows.

### Proposed Custom Skills

#### `/pre-pr` — Pre-PR Validation

Automates the 7-step checklist that is currently copy-pasted into every story. This is the single highest-impact skill.

```markdown
# /pre-pr — Pre-PR Submission Validation

Run the full pre-PR checklist and report results.

## Instructions

Run these checks sequentially and report pass/fail for each:

1. **Branch freshness:**
   ```bash
   git fetch upstream main
   git log --oneline upstream/main..HEAD | head -20
   ```
   Warn if the branch is more than 5 commits behind upstream/main.

2. **Formatting:**
   ```bash
   gofumpt -l .
   ```
   PASS if no output. FAIL if files listed (show them).

3. **Linting:**
   ```bash
   golangci-lint run ./...
   ```
   PASS if 0 issues. FAIL if any issues (show them).

4. **Tests:**
   ```bash
   go test ./... -count=1
   ```
   PASS if all pass. FAIL if any fail (show failures).

5. **Dead code check:**
   ```bash
   go vet ./...
   ```
   PASS if clean. WARN if issues found.

6. **Scope review:**
   ```bash
   git diff --stat upstream/main...HEAD
   ```
   Show the file list. Warn about any files outside `internal/`, `cmd/`, `docs/stories/`, or test files.

7. **Commit cleanliness:**
   ```bash
   git log --oneline upstream/main..HEAD
   ```
   Warn if commit messages contain "fix", "fixup", "wip", or "squash".

Report a summary table:
| Check | Status |
|-------|--------|

If all pass, say "Ready to push and create PR."
If any fail, say "Fix the above issues before submitting."
```

#### `/validate-adapter` — TaskProvider Compliance Check

```markdown
# /validate-adapter — Validate TaskProvider Implementation

Check that a TaskProvider implementation is complete and correct.

## Instructions

1. Read `internal/tasks/provider.go` to get the current `TaskProvider` interface.

2. Find all types that implement `TaskProvider`:
   ```bash
   grep -rn "func (.*) LoadTasks" internal/tasks/
   ```

3. For each implementing type, verify:
   - All 5 interface methods are implemented (LoadTasks, SaveTask, SaveTasks, DeleteTask, MarkComplete)
   - Error wrapping uses `fmt.Errorf("...: %w", err)` pattern
   - The type is registered in `provider_factory.go`'s switch statement
   - A corresponding `_test.go` file exists with test coverage
   - No direct file I/O outside the atomic write pattern (for file-based providers)

4. Report compliance for each adapter in a table:
   | Adapter | Methods | Error Wrapping | Factory | Tests | Atomic Writes |
   |---------|---------|---------------|---------|-------|--------------|

5. Flag any issues or missing implementations.
```

#### `/new-story` — Story Template Generator

```markdown
# /new-story — Generate Story Template

Create a new story file from the standard template.

## Instructions

1. Ask for the story identifier (e.g., "3.5") and title.

2. Read `docs/stories/` to find the latest story format (use the most recent file as reference).

3. Create `docs/stories/{id}.story.md` with the standard structure:
   - YAML frontmatter (title, status: "Draft")
   - User story format (As a / I want / So that)
   - Acceptance Criteria section (empty, numbered)
   - NOT In Scope section
   - Definition of Done (reference CLAUDE.md, don't duplicate checklist)
   - Architecture & Design section
   - Key Files to Create/Modify table
   - Testing section
   - Tasks / Subtasks section
   - Dev Agent Record section (empty)
   - QA Results section (empty)

4. IMPORTANT: Do NOT include the Pre-PR Submission Checklist — that lives in CLAUDE.md.

5. Do NOT include Dev Notes that duplicate CLAUDE.md rules (coding standards, patterns, etc.).
   Only include story-SPECIFIC dev notes.
```

#### `/check-patterns` — Design Pattern Compliance

```markdown
# /check-patterns — Verify Design Pattern Compliance

Scan the codebase for common pattern violations.

## Instructions

Check these patterns and report violations:

1. **Direct status mutation** — `task.Status =` without `IsValidTransition()`:
   ```bash
   grep -rn '\.Status\s*=' internal/ --include='*.go' | grep -v '_test.go' | grep -v 'task_status.go'
   ```

2. **Direct file writes** (bypassing atomic write pattern):
   ```bash
   grep -rn 'os.WriteFile\|ioutil.WriteFile' internal/ --include='*.go' | grep -v '_test.go' | grep -v '.tmp'
   ```

3. **fmt.Println in TUI code**:
   ```bash
   grep -rn 'fmt.Print' internal/tui/ --include='*.go' | grep -v '_test.go'
   ```

4. **Panics in user code**:
   ```bash
   grep -rn 'panic(' internal/ --include='*.go' | grep -v '_test.go'
   ```

5. **Provider instantiation outside factory**:
   ```bash
   grep -rn 'NewTextFileProvider\|NewAppleNotesProvider' --include='*.go' | grep -v 'provider_factory.go' | grep -v '_test.go'
   ```

6. **Missing error wrapping** (bare error returns without %w):
   ```bash
   grep -rn 'return.*err$' internal/ --include='*.go' | grep -v '_test.go' | grep -v '%w'
   ```

Report findings grouped by pattern with file:line references.
```

---

## Finding 4: Reducing Repetition (DRY Analysis)

### What's Repeated Across All 11 Story Files

| Repeated Element | Occurrences | Lines per Story | Action |
|-----------------|-------------|-----------------|--------|
| Pre-PR Submission Checklist | 11/11 stories | ~10 lines | Move to CLAUDE.md, reference only |
| Definition of Done (make build, make test, gofumpt, golangci-lint) | 11/11 stories | ~6 lines | Move standard DoD to CLAUDE.md |
| Dev Notes: atomic write pattern | ~5 stories | ~8 lines | In CLAUDE.md |
| Dev Notes: error wrapping pattern | ~6 stories | ~3 lines | In CLAUDE.md |
| Dev Notes: MVU pattern reminder | ~6 stories | ~5 lines | In CLAUDE.md |
| Dev Notes: "follow existing patterns in X" | ~8 stories | ~2 lines | In CLAUDE.md |
| Historical PR references (e.g., "PR #9, #10 had gofumpt issues") | 11/11 stories | ~5 lines | In pr-submission-standards.md (already there) |

### Estimated Savings

- **~40-50 lines per story** can be removed
- Across 11 existing stories: **~500 lines** of duplicated content
- For every future story: authors write ~40 fewer lines
- Maintenance: updating a standard (e.g., adding a new lint rule) requires editing **1 file instead of N+1 files**

### Proposed Story File Structure (After DRY)

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

Gone:
- No embedded Pre-PR Submission Checklist (it's in CLAUDE.md)
- No repeated coding standards in Dev Notes (they're in CLAUDE.md)
- No pattern reminders (atomic writes, MVU, etc. — in CLAUDE.md)
- No historical PR references (they're in pr-submission-standards.md)

---

## Finding 5: Quality Improvements — Preventing CI Failures

### Root Cause Analysis of Historical Issues

| Issue Category | PRs Affected | Root Cause | Prevention |
|---------------|-------------|------------|------------|
| gofumpt not run | #9, #10, #23, #24 | Manual process, easy to forget | CLAUDE.md rule + `/pre-pr` skill |
| golangci-lint errors | #16 | errcheck violations not caught locally | CLAUDE.md rule + `/pre-pr` skill |
| Merge conflicts | #3, #5, #19, #23 | Stale branches, not rebasing | CLAUDE.md rule (rebase first) |
| Out-of-scope files | #5 | `git add .` without review | CLAUDE.md rule (no `git add .`) |
| Dead code shipped | #28 | No dead code check in workflow | `go vet` in `/pre-pr` |
| Fix-up commit trails | ~15 PRs | No squash discipline | CLAUDE.md rule + NFR-CQ5 |
| Security (AppleScript injection) | #17 | No input sanitization rule | CLAUDE.md security section |

### Proposed CLAUDE.md Rules That Would Have Prevented These

1. **"Always run `make fmt && make lint && make test` before committing"** — prevents 4 PRs worth of gofumpt/lint issues
2. **"Never use `git add .` or `git add -A` — always add specific files"** — prevents out-of-scope file commits
3. **"Rebase onto `upstream/main` before pushing"** — prevents merge conflicts
4. **"Squash fix-up commits before PR creation"** — prevents noisy history
5. **"All AppleScript strings must be escaped — see security.md"** — prevents injection

### Additional Quality Rules for CLAUDE.md

```markdown
## Git Workflow
- Always create a feature branch — never commit to main
- Rebase, don't merge: `git fetch upstream main && git rebase upstream/main`
- Stage specific files: `git add internal/tasks/foo.go` — never `git add .`
- Squash fix-ups before push
- One logical change per commit

## Before Every Commit
Run in this order:
1. `gofumpt -w .`
2. `golangci-lint run ./...`
3. `go test ./... -count=1`
If any step fails, fix before committing. Do not use --no-verify.
```

---

## Finding 6: Idiomatic Go & Design Pattern Enforcement

### Current Pattern Usage (Good)

The codebase already uses several patterns well:

| Pattern | Where | Quality |
|---------|-------|---------|
| Adapter (interface) | `TaskProvider` → `TextFileProvider`, `AppleNotesProvider` | Good — clean interface |
| Factory | `NewProviderFromConfig()` | Good — single creation point |
| Decorator | `FallbackProvider` wrapping primary+fallback | Good — transparent wrapping |
| State Machine | `TaskStatus` transitions with `IsValidTransition()` | Good — explicit validation |
| MVU | Bubbletea `Init/Update/View` | Good — framework-enforced |
| Observer | Bubbletea message types in `messages.go` | Good — typed messages |
| Repository | `FileManager` for YAML I/O | Good — encapsulated persistence |

### Patterns to Codify in CLAUDE.md

These should be documented as rules so every agent follows them:

1. **Interface-first design**: Define the interface before the implementation. New storage backends, selection algorithms, or sync engines should start with an interface in a separate file.

2. **Constructor functions**: Every exported type gets a `NewXxx()` constructor. Never initialize structs with bare literals outside tests.

3. **Accept interfaces, return structs**: Functions should accept `TaskProvider` (interface) but return `*TextFileProvider` (concrete). This follows the Go proverb.

4. **Small interfaces**: Prefer many small interfaces over one large one. `TaskProvider` has 5 methods — if a consumer only needs `LoadTasks`, consider whether a `TaskLoader` interface would be cleaner.

5. **Table-driven tests**: All test functions with multiple cases should use table-driven format with `t.Run()` subtests.

6. **Package by feature, not by layer**: The current `internal/tasks/` and `internal/tui/` split is good — it's by feature domain, not by technical layer. Don't create packages like `internal/models/` or `internal/utils/`.

7. **No `utils` or `helpers` packages**: If a function is general-purpose, put it where it's used. If it's used everywhere, it belongs in the package that owns the concept.

8. **Errors are values**: Use sentinel errors for expected conditions (`var ErrTaskNotFound = errors.New("task not found")`). Wrap with `%w` for unexpected errors. Check with `errors.Is()` and `errors.As()`.

9. **Zero values are useful**: Design types so the zero value is valid and useful. For example, an empty `TaskPool` should be usable without initialization.

### Patterns to Watch For (Anti-Patterns)

Add to CLAUDE.md's "What NOT to Do" section:

```markdown
## Go Anti-Patterns to Avoid
- No `interface{}` or `any` unless absolutely necessary (e.g., YAML unmarshalling)
- No init() functions — explicit initialization in main() or constructors
- No global mutable state — pass dependencies explicitly
- No premature abstraction — don't create an interface until you have 2+ implementations
- No deep package nesting — `internal/tasks/` is fine, `internal/tasks/models/types/` is not
- No getters for exported fields — use exported fields directly
- No `sync.Mutex` unless you've proven a race exists (use `go test -race`)
```

---

## Finding 7: Implementation Recommendations

### Priority Order

| Priority | Item | Impact | Effort |
|----------|------|--------|--------|
| P0 | Create CLAUDE.md | Eliminates repetition in every future story, prevents CI failures | Medium (mostly writing) |
| P0 | Create `/pre-pr` skill | Automates the most-violated checklist | Small |
| P1 | Create SOUL.md | Guides agent decision-making on ambiguous cases | Small (mostly writing) |
| P1 | Slim down existing story files | Remove duplicated content from 11 files | Medium |
| P2 | Create `/validate-adapter` skill | Prevents adapter compliance issues | Small |
| P2 | Create `/check-patterns` skill | Catches anti-patterns before PR | Small |
| P3 | Create `/new-story` skill | Faster story creation, consistent format | Small |

### File Locations

```
CLAUDE.md                              ← Project root (auto-loaded by Claude Code)
SOUL.md                                ← Project root (referenced from CLAUDE.md)
.claude/commands/pre-pr.md             ← /pre-pr skill
.claude/commands/validate-adapter.md   ← /validate-adapter skill
.claude/commands/check-patterns.md     ← /check-patterns skill
.claude/commands/new-story.md          ← /new-story skill
```

### What NOT to Change

- **Don't remove docs/architecture/pr-submission-standards.md** — it provides rationale and history. CLAUDE.md references it but doesn't replace it.
- **Don't remove docs/architecture/coding-standards.md** — it's the authoritative source. CLAUDE.md summarizes the essentials.
- **Don't remove story-specific Dev Notes** — only remove the generic ones that duplicate CLAUDE.md.
- **Don't over-engineer CLAUDE.md** — keep it scannable. Link to detailed docs for deep dives.

---

## Appendix: Cross-Reference of Repeated Content

### Pre-PR Checklist appears in:
1. `docs/architecture/pr-submission-standards.md` (canonical source)
2. `docs/stories/1.1.story.md`
3. `docs/stories/1.2.story.md`
4. `docs/stories/1.3.story.md`
5. `docs/stories/1.7.story.md`
6. `docs/stories/1.8.story.md`
7. `docs/stories/2.2.story.md`
8. `docs/stories/2.4.story.md`
9. `docs/stories/3.3.story.md`
10. `docs/stories/3.4.story.md`
11. `docs/stories/3.6.story.md`
12. `docs/stories/5.1.story.md`

### Coding standards repeated in stories:
- "gofumpt -d . produces no output" — all 11 stories (Definition of Done)
- "golangci-lint run ./... passes" — all 11 stories
- "make build succeeds" — all 11 stories
- "make test passes" — all 11 stories
- Atomic write pattern — stories 1.2, 1.3, 2.2, 2.4, 3.1
- Error wrapping with %w — stories 1.2, 1.3, 2.2, 2.4, 2.6, 3.1
- MVU pattern — stories 1.2, 1.3, 3.1, 3.2, 3.3, 3.4
