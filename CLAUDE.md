# CLAUDE.md — ThreeDoors

## Project Overview

ThreeDoors is a Go TUI application that reduces task management decision friction by showing only three tasks at a time. Built with Bubbletea (charmbracelet/bubbletea). See [SOUL.md](SOUL.md) for project philosophy and design values.

- **Language:** Go 1.25.4+
- **TUI Framework:** Bubbletea + Lipgloss + Bubbles
- **Data:** YAML task files, JSONL session logs
- **Build:** `just build` · `just test` · `just lint` · `just fmt`

## Project Structure

```
cmd/threedoors/       # Entry point
internal/tasks/       # Task domain: models, providers, persistence, analytics
internal/tui/         # Bubbletea views and UI components
docs/                 # Architecture, stories, PRD
scripts/              # Shell analysis scripts
```

Key interfaces: `TaskProvider` (internal/tasks/provider.go) — implement for new storage backends.

## Development Workflow

```bash
just fmt              # gofumpt formatting (run before every commit)
just lint             # golangci-lint — must pass with zero warnings
just test             # go test ./... -v
go test -race ./...   # Race detector — run before pushing
```

## CODEOWNERS Protection — MANDATORY

Governance-critical files are protected by `.github/CODEOWNERS` with `require_code_owner_review` enabled in the branch ruleset. PRs touching these files **require @skippy approval** before merge. PRs touching only unprotected files merge with CI-only gates (existing behavior).

**Protected files (require human review):**
- `SOUL.md` — project philosophy
- `CLAUDE.md` — agent instructions
- `.claude/` — agent rules, settings
- `ROADMAP.md` — scope control
- `docs/prd/epic-list.md` — epic definitions
- `docs/prd/epics-and-stories.md` — story breakdown
- `docs/decisions/BOARD.md` — architectural decisions
- `.github/` — CI/CD, CODEOWNERS itself
- `agents/` — agent behavior definitions

**Unprotected (AI agents can self-merge via merge-queue):**
- `internal/`, `cmd/`, `pkg/` — all application code
- `docs/stories/` — workers must update story status freely
- Test files, fixtures, scripts, build files

**Rules for workers:**
- Do NOT modify CODEOWNERS-protected files unless the story explicitly requires it
- merge-queue will skip PRs that touch protected files and label them `status.needs-human`
- If your story requires changes to protected files, the PR will need manual owner approval

## Git Safety — Hook-Enforced

A PreToolUse hook (`scripts/hooks/git-safety.sh`) mechanically enforces git safety rules via `.claude/settings.json`. This replaces prompt-level INC-002 guardrails with code-level enforcement that cannot be bypassed (Q-C-005).

**Blocked commands** (hook exits with code 2, tool call is rejected):
- `git fetch`, `git pull`, `git rebase`, `git merge` — worktrees are managed by multiclaude; manual sync causes mid-rebase conflicts (INC-002)
- `--no-gpg-sign` / `-c commit.gpgsign=false` — all commits must be signed
- `git push origin main/master` — use feature branches, never push directly to main
- `Co-Authored-By` trailers — forbidden per project policy

**Allowed:** `git add`, `git commit` (signed), `git push` (feature branches), `git status`, `git log`, `git diff`, `git branch`, `git checkout -b`, `git merge --abort`, `git stash`, etc.

## Story-Driven Development — MANDATORY

**DO NOT conduct work without a story.** Every implementation task must have a corresponding `docs/stories/X.Y.story.md` file before work begins. If work needs to get done, find or create the appropriate story first.

- Before implementing, check `docs/decisions/BOARD.md` for relevant prior decisions, rejected options, and active research
- Before implementing, verify the story file exists and read its acceptance criteria
- **DO NOT check in code without first updating the story file**, verifying that the ACs and tasks were met
- After implementation, update the story file status to `Done (PR #NNN)`. **`Done` means all acceptance criteria are met in code** — planning-only PRs (story creation, docs updates, research) do NOT qualify for `Done` status. `/plan-work` creates stories with status `Not Started`; only `/implement-story` sets `Done`.
- If no story exists for needed work, create one (or ask the supervisor/PM to create one) before writing code
- Research, spikes, and documentation tasks are exempt — but should still reference a story when possible

## Doc Maintenance — MANDATORY (D-162)

Do NOT edit `ROADMAP.md`, `docs/prd/epic-list.md`, or `docs/prd/epics-and-stories.md` unless you are running `/plan-work`, or you are project-watchdog or supervisor.

### Story File Updates (Implementation Workers)

- Update the story file status line: `Done (PR #NNN)` — this is the ONLY doc update implementation workers make
- Do NOT update planning docs — project-watchdog syncs them from story files

### Planning Doc Updates (project-watchdog / PM / /plan-work)

- project-watchdog initiates planning doc updates after story PRs merge, batching multiple updates into a single PR when possible (D-161)
- `/plan-work` workers create new epics/stories in all three planning docs as part of their pipeline
- When **creating a new epic or story**, request the epic number from project-watchdog — do NOT self-assign
- ROADMAP.md ownership belongs to the PM role
- These three planning docs plus the story files form the source-of-truth chain — story files are authoritative for individual story status; planning docs must be kept consistent

## Decision Recording — MANDATORY

When a party mode session, research spike, or architectural discussion produces a decision:

- Add an entry to `docs/decisions/BOARD.md` before the PR is submitted
- Record both the adopted approach AND rejected alternatives with rationale
- If a prior decision is being overridden, update the original entry rather than creating a duplicate

## Race Detector — MANDATORY for TUI and CLI

Any PR modifying files in `internal/tui/` or `internal/cli/` MUST pass `go test -race ./internal/tui/... ./internal/cli/...` before submission. This is not optional — concurrency bugs in these packages have caused production panics.

## Commit Message Format

Every commit message MUST reference the story being implemented:
- `feat: <description> (Story X.Y)`
- `fix: <description> (Story X.Y)`
- `docs: <description> (Story X.Y)`

Commits for infrastructure work without a story should reference the issue number: `fix: <description> (#NNN)`

## Go Quality Rules

### Idiomatic Go — MUST Follow

These rules prevent the most common AI-generated Go anti-patterns.

**1. Use `fmt.Fprintf` — never `WriteString` + `Sprintf`**
```go
// WRONG — allocates intermediate string
s.WriteString(fmt.Sprintf("Task: %s", name))

// RIGHT — writes directly to the writer
fmt.Fprintf(&s, "Task: %s", name)
```

**2. Never nil-check before `len`**
```go
// WRONG — len handles nil slices/maps (returns 0)
if tasks != nil && len(tasks) > 0 { ... }

// RIGHT
if len(tasks) > 0 { ... }
```

**3. Always check error returns**
```go
// WRONG — silently ignoring error
data, _ := json.Marshal(task)

// RIGHT — handle or propagate every error
data, err := json.Marshal(task)
if err != nil {
    return fmt.Errorf("marshal task %s: %w", task.ID, err)
}
```

**4. Wrap errors with context using `%w`**
```go
// WRONG — loses error chain
return fmt.Errorf("failed to save: %v", err)

// RIGHT — preserves chain for errors.Is/errors.As
return fmt.Errorf("save task %s: %w", id, err)
```

**5. Accept interfaces, return concrete types**
```go
// WRONG — returning interface hides implementation
func NewProvider() TaskProvider { ... }

// RIGHT — return the concrete type
func NewTextFileProvider(path string) *TextFileProvider { ... }
```

**6. `context.Context` is always the first parameter**
```go
// WRONG
func LoadTasks(path string, ctx context.Context) error

// RIGHT
func LoadTasks(ctx context.Context, path string) error
```

**7. Don't use `interface{}`/`any` without justification**
- Prefer specific types or generics over `any`
- If `any` is needed, document why in a comment

**8. Prefer value receivers unless mutation is needed**
```go
// Use pointer receiver only when:
// - The method mutates the receiver
// - The struct is large (>~64 bytes) and copying is expensive
// - Consistency: if one method needs pointer, all should use pointer
```

**9. No `init()` functions**
- Pass dependencies explicitly via constructors
- Configuration belongs in `main()` or factory functions

**10. Timestamps always in UTC**
```go
// WRONG
time.Now()

// RIGHT
time.Now().UTC()
```

### Error Handling

- Every exported function that can fail returns `error` as last return value
- Use `errors.Is()` and `errors.As()` for error inspection — never string matching
- Define sentinel errors as package-level `var` with documentation:
  ```go
  // ErrTaskNotFound is returned when a task ID doesn't exist in the pool.
  var ErrTaskNotFound = errors.New("task not found")
  ```
- No panics in user-facing code — Bubbletea `Update()` and `View()` must never panic

### Testing Standards

- **Table-driven tests** for any function with >2 test cases:
  ```go
  func TestValidateStatus(t *testing.T) {
      tests := []struct {
          name    string
          from    Status
          to      Status
          wantErr bool
      }{
          {"todo to active", StatusTodo, StatusActive, false},
          {"done to todo", StatusDone, StatusTodo, true},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              err := ValidateTransition(tt.from, tt.to)
              if (err != nil) != tt.wantErr {
                  t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
              }
          })
      }
  }
  ```
- **Use stdlib `testing`** — no testify. Use `t.Fatal`, `t.Errorf`, `t.Helper()`
- **Use `t.Helper()`** in test helper functions so failures report the caller's line
- **Use `t.Cleanup()`** instead of `defer` for test resource cleanup
- **Test files** live alongside source: `foo.go` → `foo_test.go`
- **Test fixtures** in `testdata/` directories
- Mark independent tests with `t.Parallel()` where safe

### Code Organization

- **Package naming:** lowercase, single word (`tasks`, `tui`) — no underscores, no camelCase
- **File naming:** lowercase snake_case (`task_pool.go`, `doors_view.go`)
- **One primary type per file** — `task.go` defines `Task`, `task_pool.go` defines `TaskPool`
- **Import order:** stdlib → external → internal (gofumpt enforces this)
- **Keep packages small** — split when a package exceeds ~10 files

### Design Patterns in This Project

- **Provider pattern** (`TaskProvider` interface) for storage backends — add new providers by implementing the interface
- **Factory functions** (`NewTaskPool()`, `NewTextFileProvider()`) — always use constructors, never raw struct literals for exported types
- **Atomic writes** for all file persistence — write to `.tmp`, sync, rename (see `docs/architecture/coding-standards.md`)
- **Bubbletea pattern** — all TUI output through `View()` methods, never `fmt.Println`

### Common AI Mistakes to Avoid

1. **Don't create unnecessary abstractions** — three similar lines are better than a premature helper
2. **Don't add unused parameters** "for future use" — YAGNI
3. **Don't shadow imports** — `var errors = ...` shadows the `errors` package
4. **Don't use `log.Fatal`/`os.Exit` outside `main()`** — let errors propagate
5. **Don't buffer channels without justification** — unbuffered is the default for a reason
6. **Don't use `sync.Mutex` when `atomic` suffices** for simple counters/flags
7. **Don't create `utils` or `helpers` packages** — put functions where they're used
8. **Don't add comments that restate the code** — only comment the "why", not the "what"
9. **Don't use `strings.Builder` then call `Sprintf` into it** — use `fmt.Fprintf` directly
10. **Don't return `bool, error` as a substitute for `error`** — if the bool just means "did it succeed", the error alone suffices

### Formatting & Linting

- **Formatter:** `gofumpt` (stricter than `gofmt`) — run via `just fmt`
- **Linter:** `golangci-lint run ./...` — must pass with zero warnings
- **Vet:** `go vet ./...` — runs as part of `golangci-lint`
- Never disable linter rules with `//nolint` without a justifying comment

### Go Proverbs to Follow

> The bigger the interface, the weaker the abstraction.

> Make the zero value useful.

> A little copying is better than a little dependency.

> Don't communicate by sharing memory; share memory by communicating.

> Errors are values — program with them.

> Don't just check errors, handle them gracefully.

## TUI-Specific Rules

- All user-visible output goes through Bubbletea `View()` — never `fmt.Println`
- Use Lipgloss for styling — never ANSI escape codes directly
- Keep `Update()` fast — no blocking I/O in the update loop
- Use `tea.Cmd` for async operations (file I/O, timers)
