# Coding Standards

**⚠️ MANDATORY for AI Agents:** These standards directly control code generation behavior.

## Core Standards

**Languages & Runtimes:**
- Go 1.25.4+ strictly
- No external languages in codebase

**Style & Linting:**
- **Formatting:** `gofumpt` - run before every commit
- **Linting:** `golangci-lint run ./...` - must pass with zero warnings
- **Import ordering:** Standard library → external → internal (auto-formatted)

**Test Organization:**
- Test files: `*_test.go` alongside source files
- Table-driven tests preferred
- Test fixtures: `testdata/` directory

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| **Packages** | Lowercase, single word | `tui`, `tasks` |
| **Files** | Lowercase, snake_case | `task_pool.go`, `doors_view.go` |
| **Types (exported)** | PascalCase | `TaskPool`, `DoorSelection` |
| **Types (private)** | camelCase | `internalState` |
| **Functions (exported)** | PascalCase | `NewTaskPool`, `SelectDoors` |
| **Functions (private)** | camelCase | `validateTask`, `renderDoor` |
| **Constants** | PascalCase | `StatusTodo`, `MaxTasks` |

## Critical Rules

**MUST Follow:**

1. **Never use fmt.Println for user output in TUI code**
   - TUI output goes through Bubbletea View() methods only
   - Logging goes through log.Printf() to stderr

2. **All file writes must use atomic write pattern**
   - Write to `.tmp` file
   - Sync to disk
   - Atomic rename
   - Cleanup temp on error

3. **Always validate status transitions before applying**
   - Call StatusManager.ValidateTransition() first
   - Never allow direct Task.Status field assignment from UI

4. **Errors must be wrapped with context**
   - Use `%w` verb: `fmt.Errorf("operation failed: %w", err)`
   - Preserves error chain for errors.Is() and errors.As()

5. **No panics in user-facing code**
   - Bubbletea Update() and View() must never panic
   - Return error values, handle gracefully

6. **Task IDs are immutable**
   - UUID assigned at creation
   - Never modify Task.ID after creation

7. **Timestamps always stored in UTC**
   - Use `time.Now().UTC()` not `time.Now()`
   - Convert to local timezone only for display

8. **YAML field tags match schema exactly**
   - Use `yaml:"field_name"` tags
   - Use `omitempty` for nullable fields

## Atomic Write Pattern Checklist

**CRITICAL:** Every file write operation MUST follow this exact pattern to prevent data corruption:

```
✅ Step 1: Create temp path
   tempPath := targetPath + ".tmp"

✅ Step 2: Write to temp file
   if err := os.WriteFile(tempPath, data, 0644); err != nil {
       return fmt.Errorf("failed to write temp file: %w", err)
   }

✅ Step 3: Sync to disk (flush buffers)
   f, err := os.OpenFile(tempPath, os.O_RDWR, 0644)
   if err == nil {
       f.Sync()
       f.Close()
   }

✅ Step 4: Atomic rename
   if err := os.Rename(tempPath, targetPath); err != nil {
       os.Remove(tempPath)  // Cleanup on failure
       return fmt.Errorf("failed to commit changes: %w", err)
   }

✅ Step 5: Success - temp file now atomically replaces target
```

**Why This Matters:**
- Prevents partial writes (crash during write leaves original intact)
- Prevents corruption (temp file discarded if write fails)
- Atomic rename is OS-level operation (succeeds or fails completely)

**Reference Implementation:** See `FileManager.SaveTasks()` in Section 5 (Components)

## PR-Analysis-Derived Rules (Mandatory)

> These rules are derived from analysis of all 49 PRs (#1–#49). Each rule prevents a specific class of defect that recurred across multiple PRs.

### Rule 9: Always use fmt.Fprintf, never WriteString+Sprintf

**MUST Follow:**

```go
// ❌ WRONG — triggers staticcheck QF1012, creates unnecessary allocation
buf.WriteString(fmt.Sprintf("Score: %d", score))

// ✅ CORRECT — direct formatted write
fmt.Fprintf(buf, "Score: %d", score)
```

This applies to ALL `*bytes.Buffer`, `*strings.Builder`, and any `io.Writer`. No exceptions.

*Evidence: 11+ violations across PRs #42, #44, #45 required 5 fix-up commits. This was the single most recurring lint failure in the project.*

### Rule 10: Check ALL error return values — including Close, Remove, WriteFile

**MUST Follow:**

```go
// ❌ WRONG — errcheck violation
defer f.Close()

// ✅ CORRECT — check error on writable file handles
defer func() {
    if cerr := f.Close(); cerr != nil && err == nil {
        err = fmt.Errorf("closing file: %w", cerr)
    }
}()

// ❌ WRONG — ignoring cleanup error
os.Remove(tempPath)

// ✅ CORRECT — check or explicitly document why ignored
if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
    log.Printf("warning: failed to clean up temp file: %v", err)
}
```

In test code, use `t.Helper()` patterns or `require.NoError()` — do not assign errors to `_`.

*Evidence: 18+ errcheck violations across PRs #16, #42, #43. `f.Close()` was the most common offender (6 instances).*

### Rule 11: Escape all user input in AppleScript/shell interpolation

**MUST Follow:**

```go
// ❌ WRONG — injection vulnerability
script := fmt.Sprintf(`tell app "Notes" to show note "%s"`, noteTitle)

// ✅ CORRECT — escape for AppleScript string context
escaped := strings.ReplaceAll(noteTitle, `\`, `\\`)
escaped = strings.ReplaceAll(escaped, `"`, `\"`)
script := fmt.Sprintf(`tell app "Notes" to show note "%s"`, escaped)
```

Every dynamic command construction MUST have a corresponding test with special characters.

*Evidence: PR #17 had an AppleScript injection vulnerability via unescaped note titles.*

### Rule 12: Call time.Now() once per operation, reuse the result

**MUST Follow:**

```go
// ❌ WRONG — inconsistent timestamps across loop iterations
for _, task := range tasks {
    task.ParsedAt = time.Now().UTC()
}

// ✅ CORRECT — single timestamp for the batch
now := time.Now().UTC()
for _, task := range tasks {
    task.ParsedAt = now
}
```

*Evidence: PR #17 called time.Now() inside a parseNoteBody loop.*

### Rule 13: Fix lint categories by sweeping the entire codebase

**MUST Follow:**

When CI reports a lint violation (e.g., QF1012 on line 117), search the ENTIRE codebase for all instances of that pattern and fix them all in a single commit. Do NOT fix only the reported lines.

```bash
# After fixing a QF1012 violation, verify no others exist:
grep -rn "WriteString(fmt.Sprintf" internal/ cmd/ --include="*.go"
# Must produce zero results
```

*Evidence: PR #42 fixed QF1012 incrementally across 3 separate commits, each revealing new instances on different lines.*

## Pre-PR Verification Checklist

Run this sequence before every PR submission. All checks MUST pass:

```bash
# 1. Format
gofumpt -l -w .

# 2. Lint (zero issues required)
golangci-lint run ./...

# 3. Tests (all must pass)
go test ./...

# 4. Verify no WriteString+Sprintf anti-pattern
! grep -rn "WriteString(fmt.Sprintf" internal/ cmd/ --include="*.go"

# 5. Verify no unchecked Close on writable files
# (manual review — look for bare `defer f.Close()` on files opened for writing)

# 6. Rebase onto latest main
git fetch upstream main && git rebase upstream/main

# 7. Re-run format after rebase (rebase can introduce drift)
gofumpt -l -w .

# 8. Scope check — only story-related files changed
git diff --stat upstream/main...HEAD
```

---
