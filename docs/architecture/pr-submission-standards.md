# PR Submission Standards

## Purpose

This document codifies the code quality and PR submission standards for the ThreeDoors project. These standards were derived from a retrospective analysis of all 31 pull requests (#1-#31) and formalized as non-functional requirements NFR-CQ1 through NFR-CQ5 in the PRD.

## Rationale

Recurring patterns across PRs caused delays between submission and merge:

| Pattern | PRs Affected | Impact |
|---------|-------------|--------|
| gofumpt formatting not run before push | #9, #10, #23, #24 | Required fix-up commits |
| golangci-lint errors (errcheck, staticcheck) | #16 | 2 extra commits to resolve |
| Merge conflicts from stale branches | #3, #5, #19, #23 | PRs delayed or closed/recreated |
| Out-of-scope files committed | #5 | Included unrelated `agents/` directory |
| Dead code shipped | #28 | Unused `SortedDates()` method merged |
| Fix-up commit trails | ~15 PRs | Noisy git history |
| PRs closed & recreated | #3 superseded by #5, #14 by #13 | Wasted review effort |
| AppleScript injection risk | #17 | Security fix required post-merge |
| CI workflow token scope | #30 | PR blocked by missing OAuth scope |

## Standards

### NFR-CQ1: Code Formatting (gofumpt)

All Go code must pass `gofumpt` formatting before submission.

**Verification:**
```bash
gofumpt -l .
# Must produce no output. If files are listed:
gofumpt -w .
```

**Why gofumpt over gofmt:** The project uses `gofumpt` (a stricter superset of `gofmt`) as specified in TD-NFR1 and NFR1. It enforces additional formatting rules beyond the standard Go formatter.

### NFR-CQ2: Linting (golangci-lint)

All code must pass `golangci-lint run ./...` with zero issues before submission.

**Verification:**
```bash
golangci-lint run ./...
# Must report 0 issues
```

**Key linters to watch:**
- `errcheck` — unchecked error returns (most common finding)
- `staticcheck` — static analysis bugs and simplifications
- `govet` — suspicious constructs
- `unused` — unused code

### NFR-CQ3: Branch Freshness (Rebase)

All branches must be rebased onto `upstream/main` before PR creation.

**Verification:**
```bash
git fetch upstream main
git rebase upstream/main
# Resolve any conflicts before pushing
```

**Why rebase over merge:** Rebase produces a linear history and catches integration issues before the PR is created, rather than during review.

### NFR-CQ4: Scope Discipline (Diff Review)

All PRs must contain only in-scope changes. Review `git diff --stat` before pushing.

**Verification:**
```bash
git diff --stat upstream/main...HEAD
# Every file listed should be directly related to the story being implemented
```

**Common violations:**
- Unrelated directories accidentally staged (`git add .` without review)
- IDE-generated config files
- Unrelated formatting changes in files not touched by the story

### NFR-CQ5: Clean Commits (Squash Fix-ups)

All fix-up commits must be squashed before PR submission. PRs should contain a single clean commit or logically separated commits — not iterative fix-up trails.

**Verification:**
```bash
git log --oneline upstream/main..HEAD
# Should show clean, meaningful commit messages — not "fix lint", "fix formatting", etc.
```

**How to squash:**
```bash
# Interactive rebase to squash fix-ups into the main commit
git rebase -i upstream/main
# Mark fix-up commits as 'fixup' or 'squash'
```

## Pre-PR Submission Checklist

Every PR must complete this checklist before submission. This checklist is embedded in every story definition in `docs/prd/epics-and-stories.md` and every story file in `docs/stories/`.

- [ ] **Rebase onto latest main**: `git fetch upstream main && git rebase upstream/main`
- [ ] **Run gofumpt**: `gofumpt -l .` — verify no output
- [ ] **Run golangci-lint**: `golangci-lint run ./...` — verify 0 issues
- [ ] **Run all tests**: `go test ./... -count=1` — verify 0 failures
- [ ] **Check for dead code**: `go vet ./...` — review for unused functions, variables, or imports
- [ ] **Verify no out-of-scope files**: Review `git diff --stat` — only story-related files
- [ ] **Single clean commit preferred**: Squash fix-ups before pushing

### Story-Specific Additional Checks

Some stories require additional checks beyond the standard checklist:

| Story | Additional Check | Reason |
|-------|-----------------|--------|
| 2.4 (Apple Notes Write) | Verify AppleScript injection safety — ensure all note titles and task text are properly escaped | PR #17 required a security fix |
| 2.5 (Bidirectional Sync) | Verify AppleScript injection safety | Same AppleScript surface area as 2.4 |
| 5.1 (macOS Signing) | Validate CI workflow YAML syntax; verify no hardcoded secrets | PR #30 was blocked by missing `workflow` OAuth scope |

## Relationship to CI/CD

These standards are intended to be enforced both manually (via the checklist) and automatically (via CI):

- **gofumpt** and **golangci-lint** are run by the `just fmt` and `just lint` targets
- **Tests** are run by `just test`
- **CI pipeline** (GitHub Actions, Story 1.7) runs these checks on every push
- The manual checklist catches issues _before_ pushing, avoiding the CI round-trip delay

## References

- PRD: `docs/prd/requirements.md` — NFR-CQ1 through NFR-CQ5
- Story definitions: `docs/prd/epics-and-stories.md` — checklist on every story
- Story files: `docs/stories/*.story.md` — checklist on every story file
- Coding standards: `docs/architecture/coding-standards.md` — gofumpt and naming conventions
- Test strategy: `docs/architecture/test-strategy-and-standards.md` — testing conventions
- PR #32: Original retrospective that identified these patterns
