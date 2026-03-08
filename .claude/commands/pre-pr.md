# /pre-pr — Pre-PR Submission Validation

Run the full pre-PR checklist and report results. See CLAUDE.md for project coding standards and SOUL.md for project philosophy.

## Instructions

Run these checks sequentially and report pass/fail for each:

1. **Branch freshness:**
   ```bash
   git fetch origin main
   git log --oneline origin/main..HEAD | wc -l
   ```
   Warn if the branch is more than 5 commits behind origin/main.

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
   git diff --stat origin/main...HEAD
   ```
   Show the file list. Warn about any files outside `internal/`, `cmd/`, `docs/`, `.claude/`, or test files that seem unrelated to the current work.

7. **Commit cleanliness:**
   ```bash
   git log --oneline origin/main..HEAD
   ```
   Warn if commit messages contain "fix", "fixup", "wip", or "squash".

Report a summary table:

| Check | Status |
|-------|--------|
| Branch freshness | PASS/WARN |
| Formatting | PASS/FAIL |
| Linting | PASS/FAIL |
| Tests | PASS/FAIL |
| Dead code | PASS/WARN |
| Scope review | PASS/WARN |
| Commit cleanliness | PASS/WARN |

If all pass, say "Ready to push and create PR."
If any FAIL, say "Fix the above issues before submitting."
If only WARNs, say "Review warnings above, then push if acceptable."
