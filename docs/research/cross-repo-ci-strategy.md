# Cross-Repo CI Strategy: ThreeDoors ↔ Shared Homebrew Tap

**Date:** 2026-03-08
**Status:** Research complete
**Related:** [Homebrew Distribution Research](homebrew-distribution-research.md), [GoReleaser Draft](goreleaser-draft.yml)

---

## Problem Statement

When a ThreeDoors release is cut via GoReleaser, it pushes a formula update to the **shared** tap repo `arcaven/homebrew-tap`. This tap serves multiple arcaven projects (currently ThreeDoors and switchboard). A release is not complete until the tap's CI passes for the ThreeDoors formula specifically. Today, there is no mechanism to detect or respond to tap CI failures from the ThreeDoors side.

## Current State

### Shared Tap: `arcaven/homebrew-tap`

| Aspect | Current State |
|--------|--------------|
| Formulas | `threedoors.rb`, `switchboard.rb` |
| CI | `brew audit --strict` + `brew style` for all formulas (single job, loop) |
| Distribution | Prebuilt binaries (darwin arm64 + amd64) |
| Push model | Direct commit to main (GoReleaser `goreleaserbot` author) |
| Isolation | None — any formula failure breaks the whole CI run |

### Release Lifecycle (Current)

```
Tag push (ThreeDoors) → GoReleaser builds → GitHub Release → Formula commit to tap → Tap CI → ???
```

The "???" is the gap: no one monitors whether tap CI passes after a release.

### Release Lifecycle (Target)

```
Tag push (ThreeDoors)
  → GoReleaser builds binaries + archives
  → GitHub Release created in ThreeDoors
  → GoReleaser pushes formula update to arcaven/homebrew-tap
  → Tap CI runs (per-formula matrix)
  → release-verify workflow detects tap CI result
  → Commit status posted to ThreeDoors release tag
  → If failed: GitHub issue opened with failure details
  → Release is "complete" when both repos are green
```

## Failure Points

| # | Failure | Where | Detection | Recovery |
|---|---------|-------|-----------|----------|
| 1 | GoReleaser build fails | ThreeDoors CI | Immediate (workflow fails) | Fix code, re-tag |
| 2 | GoReleaser can't push to tap | ThreeDoors CI | Immediate (GoReleaser step fails) | Refresh `HOMEBREW_TAP_TOKEN` |
| 3 | `brew audit` fails for threedoors | Tap CI | release-verify workflow | Fix `.goreleaser.yml` template, patch release |
| 4 | `brew install` fails | Tap CI | release-verify workflow | Fix archive/checksum config, patch release |
| 5 | `brew test` fails | Tap CI | release-verify workflow | Fix version ldflags, patch release |
| 6 | Tap CI fails for *other* formula | Tap CI | release-verify ignores (matrix isolation) | Not ThreeDoors's concern |
| 7 | Tap CI doesn't run (GitHub outage) | Timeout | release-verify timeout (30 min) | Retry manually |

## Recommended Architecture

### 1. Tap CI Enhancement (Changes to `arcaven/homebrew-tap`)

Convert the current single-job loop to a matrix strategy for formula-level isolation:

```yaml
jobs:
  audit:
    strategy:
      fail-fast: false
      matrix:
        formula:
          - threedoors
          - switchboard
    name: "Audit ${{ matrix.formula }}"
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v6
      - name: Tap local checkout
        run: |
          mkdir -p "$(brew --repository)/Library/Taps/arcaven"
          ln -sf "$GITHUB_WORKSPACE" "$(brew --repository)/Library/Taps/arcaven/homebrew-tap"
      - name: Audit
        run: brew audit --strict "arcaven/tap/${{ matrix.formula }}"
      - name: Style
        run: brew style "arcaven/tap/${{ matrix.formula }}"
      - name: Install
        run: brew install "arcaven/tap/${{ matrix.formula }}"
      - name: Test
        run: brew test "arcaven/tap/${{ matrix.formula }}"
      - name: Cleanup
        if: always()
        run: brew uninstall "arcaven/tap/${{ matrix.formula }}" || true
```

**Benefits:** Each formula gets independent pass/fail status. switchboard failures don't mask threedoors issues and vice versa. `fail-fast: false` ensures all formulas are tested even if one fails.

**Note:** This change benefits ALL projects using the shared tap.

### 2. Release Verification Workflow (New in ThreeDoors)

New file: `.github/workflows/release-verify.yml`

```yaml
name: Release Verify
on:
  workflow_run:
    workflows: ["Release"]  # Triggers after GoReleaser completes
    types: [completed]

jobs:
  verify-tap:
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    runs-on: ubuntu-latest
    timeout-minutes: 30
    steps:
      - name: Find GoReleaser tap commit
        id: find-commit
        env:
          GH_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
        run: |
          # Wait for GoReleaser commit to appear in tap repo
          for i in $(seq 1 12); do
            SHA=$(gh api repos/arcaven/homebrew-tap/commits \
              --jq 'map(select(.commit.message | startswith("chore(formula): update threedoors"))) | .[0].sha // empty' \
              2>/dev/null)
            if [ -n "$SHA" ]; then
              echo "sha=$SHA" >> "$GITHUB_OUTPUT"
              break
            fi
            sleep 30
          done
          if [ -z "$SHA" ]; then
            echo "::error::GoReleaser tap commit not found after 6 minutes"
            exit 1
          fi

      - name: Wait for tap CI
        env:
          GH_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
        run: |
          SHA="${{ steps.find-commit.outputs.sha }}"
          for i in $(seq 1 30); do
            STATUS=$(gh api repos/arcaven/homebrew-tap/commits/$SHA/status \
              --jq '.state' 2>/dev/null)
            case "$STATUS" in
              success) echo "Tap CI passed"; exit 0 ;;
              failure|error) echo "::error::Tap CI failed"; exit 1 ;;
              pending) echo "Waiting... ($i/30)"; sleep 60 ;;
            esac
          done
          echo "::error::Tap CI timed out after 30 minutes"
          exit 1

      - name: Report failure
        if: failure()
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh issue create \
            --title "Release verification failed: tap CI" \
            --body "The homebrew tap CI failed after the latest release. Check https://github.com/arcaven/homebrew-tap/actions for details." \
            --label "bug,release"
```

### 3. GoReleaser Config Update

The existing draft at `docs/research/goreleaser-draft.yml` targets `arcaven/homebrew-threedoors`. Update for shared tap:

```yaml
brews:
  - repository:
      owner: arcaven
      name: homebrew-tap  # Shared tap, not dedicated
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
    commit_msg_template: "chore(formula): update threedoors to {{ .Tag }}"
    directory: Formula
    homepage: "https://github.com/arcaven/ThreeDoors"
    description: "TUI task manager that reduces decision friction by showing only three tasks"
    license: "MIT"
    install: |
      bin.install "threedoors"
    test: |
      assert_match version.to_s, shell_output("#{bin}/threedoors --version 2>&1")
```

Key change: `name: homebrew-tap` instead of `homebrew-threedoors`.

### 4. Agent Model

No persistent release-manager agent. Instead:

- **GitHub Actions** handle automated detection (deterministic, reliable)
- **Spawnable agent prompt** (`agents/release-manager.md`) captures release domain knowledge
- On failure: supervisor spawns a worker with release-manager context to investigate

### 5. Multiclaude Integration

pr-shepherd does NOT monitor cross-repo releases. Its scope remains branch/PR health within ThreeDoors.

When tap CI fails:
1. `release-verify` workflow opens a GitHub issue
2. Supervisor sees the issue (via merge-queue's open issue cross-check or manual monitoring)
3. Supervisor spawns a worker with release-manager agent context to fix

## Multi-Formula Considerations

Since `arcaven/homebrew-tap` is shared:

1. **GoReleaser commit messages must be formula-specific** — `chore(formula): update threedoors to vX.Y.Z` (not generic). This allows filtering by formula in monitoring.
2. **Tap CI matrix isolates failures** — switchboard failures don't affect ThreeDoors verification.
3. **Token scope** — `HOMEBREW_TAP_TOKEN` needs write access to `arcaven/homebrew-tap`. This is the same token used by all arcaven projects that publish to the tap.
4. **Concurrent releases** — if ThreeDoors and switchboard release simultaneously, both push commits to the tap. CI runs for each. The release-verify workflow must find the correct commit by message pattern, not just "latest commit."
5. **Formula naming** — each project manages its own formula template via GoReleaser. No coordination needed unless formulas conflict (they won't — different names).

## What This Research Does NOT Cover

- Implementing the tap CI changes (that's in the tap repo, not ThreeDoors)
- Creating the `release-verify` workflow (implementation story)
- Writing the spawnable release-manager agent prompt
- homebrew-core submission strategy (covered in [homebrew-distribution-research.md](homebrew-distribution-research.md))

## References

- [arcaven/homebrew-tap](https://github.com/arcaven/homebrew-tap) — shared tap repo
- [GoReleaser Homebrew Taps](https://goreleaser.com/customization/homebrew/)
- [GitHub Actions workflow_run trigger](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#workflow_run)
- [GitHub Commit Status API](https://docs.github.com/en/rest/commits/statuses)
- [Homebrew Distribution Research](homebrew-distribution-research.md)
- [GoReleaser Draft Config](goreleaser-draft.yml)
