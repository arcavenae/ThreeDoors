# CI Path Filtering Improvements — Research

**Date:** 2026-03-29
**Trigger:** PR #866 (Story 73.5) runs full Go Quality Gate despite only adding shell scripts + docs
**Autonomy:** L3 (AI-autonomous research)

## Problem

PR #866 adds `scripts/quota-status.sh`, `scripts/test-quota-status.sh`, `docs/operations/quota-monitoring.md`, updates `docs/stories/73.5.story.md`, and adds 4 lines to `justfile`. No Go code was changed, yet the full Quality Gate runs (~2-3 min of Go tests, linting, vet, govulncheck, coverage).

### Root Cause

The `dorny/paths-filter` `code` filter in `.github/workflows/ci.yml` includes `justfile`:

```yaml
code:
  - '**.go'
  - 'go.mod'
  - 'go.sum'
  - 'justfile'          # <-- this triggers on ANY justfile change
  - '.golangci.yml'
  - '.github/workflows/**'
```

PR #866 modified `justfile` to add `quota-status` targets, which set `code: true`, which triggered Quality Gate. The shell script and docs changes alone would NOT have triggered it.

### Secondary Issue: No Script Linting

There are 20+ shell scripts in `scripts/` with no CI linting. These scripts handle:
- Session analytics (`analyze_sessions.sh`, `daily_completions.sh`)
- Release packaging (`create-app.sh`, `create-dmg.sh`, `create-pkg.sh`)
- Operational monitoring (`quota-status.sh`, `ci-metrics.sh`)
- Agent coordination (`rollcall.sh`, `shift-*.sh`, `handover-*.sh`)

Shell bugs in release scripts (`create-pkg.sh`, `create-dmg.sh`) could break the signing/notarization pipeline.

## Required Status Checks Context

The branch ruleset requires these checks to pass:
- **Quality Gate** (required)
- **Performance Benchmarks** (required)
- **Docker E2E Tests** (required)

GitHub Actions treats jobs skipped via `if:` conditions as passing for required check purposes. The existing `docs-pass` no-op job is **not** in the required checks list — it's cosmetic. Scripts-only PRs (without justfile changes) already skip Quality Gate successfully; the issue is specifically `justfile` being in the `code` filter.

## Recommendations

### 1. Remove `justfile` from the `code` filter

**Impact:** Low risk. Justfile changes to build/test recipes are almost always paired with `.go` file changes (new build targets need new Go code). The rare case of a justfile-only change to Go targets missing CI is acceptable — the developer would catch it locally via `just test`.

```yaml
code:
  - '**.go'
  - 'go.mod'
  - 'go.sum'
  - '.golangci.yml'
  - '.github/workflows/**'
```

### 2. Add a `scripts` filter and `shellcheck` job

Add a new filter output and a lightweight job:

```yaml
# In the changes job filters:
scripts:
  - 'scripts/**'
  - 'justfile'
```

```yaml
shellcheck:
  name: ShellCheck
  needs: changes
  if: needs.changes.outputs.scripts == 'true' || github.event_name == 'push'
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v6
    - name: Install shellcheck
      run: sudo apt-get install -y shellcheck
    - name: Run shellcheck on scripts
      run: shellcheck scripts/*.sh
```

**Why shellcheck is worthwhile:**
- 20+ shell scripts, several in the release-critical path
- ShellCheck catches real bugs: unquoted variables, missing error handling, bashisms in `#!/bin/sh` scripts
- Runs in <10 seconds — negligible CI cost
- The `scripts/create-pkg.sh` and `scripts/create-dmg.sh` scripts are used in the CI signing pipeline itself

**Note:** ShellCheck does NOT need to be a required status check. Making it informational (warning-only) avoids blocking PRs while still surfacing issues. It can be promoted to required later once existing scripts are clean.

### 3. Rename `docs-pass` to `skip-pass` (cosmetic)

The current name suggests it only applies to docs PRs, but it also covers scripts-only, story-file-only, and other non-code PRs. A more accurate name:

```yaml
skip-pass:
  name: CI Skipped (no code changes)
  needs: changes
  if: github.event_name == 'pull_request' && needs.changes.outputs.code != 'true' && needs.changes.outputs.docker != 'true'
  runs-on: ubuntu-latest
  steps:
    - run: echo "No code changes — Go CI skipped."
```

### 4. Add `scripts/hooks/**` to the `code` filter (safety)

The git safety hooks in `scripts/hooks/` affect CI behavior and should trigger Quality Gate when changed:

```yaml
code:
  - '**.go'
  - 'go.mod'
  - 'go.sum'
  - '.golangci.yml'
  - '.github/workflows/**'
  - 'scripts/hooks/**'    # CI hooks should trigger full suite
```

## Rejected Alternatives

| Alternative | Why Rejected |
|---|---|
| Keep `justfile` in `code` filter | Causes false positives like PR #866. Justfile-only changes affecting Go are rare and caught locally. |
| Split justfile into Go-specific and script-specific sections | Justfile is a single file — `dorny/paths-filter` can't filter by content, only by path. Not feasible. |
| Make shellcheck a required check immediately | Existing scripts likely have shellcheck warnings. Would block PRs until all 20+ scripts are cleaned up. Better to start informational. |
| Use `hadolint` instead of shellcheck for scripts | Hadolint is for Dockerfiles, not shell scripts. Different tool for different purpose. |

## Implementation Estimate

- Filter changes: ~10 lines in `ci.yml`
- shellcheck job: ~15 lines in `ci.yml`
- Initial shellcheck cleanup: separate story (fix existing warnings before promoting to required)

## Next Steps

1. Create a story for implementing these CI changes
2. Optionally: run `shellcheck scripts/*.sh` locally to assess current warning count before deciding on required vs informational
