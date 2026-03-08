# /check-patterns — Verify Design Pattern Compliance

Scan the codebase for common pattern violations. See CLAUDE.md for the project's design patterns and coding standards.

## Instructions

Check these patterns and report violations:

1. **Direct status mutation** — `.Status =` without `IsValidTransition()`:
   ```bash
   grep -rn '\.Status\s*=' internal/ --include='*.go' | grep -v '_test.go' | grep -v 'task_status.go'
   ```
   Exclude legitimate assignments inside the status transition logic itself.

2. **Direct file writes** (bypassing atomic write pattern):
   ```bash
   grep -rn 'os.WriteFile\|ioutil.WriteFile' internal/ --include='*.go' | grep -v '_test.go' | grep -v '.tmp'
   ```
   All file writes in production code should use the atomic write pattern (write to .tmp, sync, rename).

3. **fmt.Println in TUI code**:
   ```bash
   grep -rn 'fmt.Print' internal/tui/ --include='*.go' | grep -v '_test.go'
   ```
   TUI output must go through Bubbletea `View()` methods only.

4. **Panics in user-facing code**:
   ```bash
   grep -rn 'panic(' internal/ --include='*.go' | grep -v '_test.go'
   ```
   No panics in production code — errors should propagate, not panic.

5. **Provider instantiation outside factory**:
   ```bash
   grep -rn 'NewTextFileProvider\|NewAppleNotesProvider\|NewObsidianProvider\|NewJiraProvider\|NewGitHubProvider' --include='*.go' | grep -v 'provider_factory\|_test.go'
   ```
   Providers should be created through the factory, not instantiated directly in application code.

6. **Missing error wrapping** (bare error returns without context):
   ```bash
   grep -rn 'return.*err$' internal/ --include='*.go' | grep -v '_test.go' | grep -v 'fmt.Errorf'
   ```
   Errors should be wrapped with `fmt.Errorf("context: %w", err)` for debuggability.

Report findings grouped by pattern with file:line references.

If no violations found, say "All pattern checks pass — no violations detected."
If violations found, list them with severity (MUST FIX vs SHOULD FIX) and suggested corrections.
