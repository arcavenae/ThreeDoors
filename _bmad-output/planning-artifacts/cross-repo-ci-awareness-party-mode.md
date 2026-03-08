# Party Mode: Cross-Repo CI Awareness for Shared Homebrew Tap

**Date:** 2026-03-08
**Participants:** PM (John), Architect (Winston), Dev (Amelia), QA (Quinn/Murat), SM (Bob)
**Topic:** Cross-repo CI awareness between ThreeDoors and the shared `arcaven/homebrew-tap` repo

---

## Context

- **Shared tap repo:** `arcaven/homebrew-tap` (already exists, public)
- **Current formulas:** `threedoors.rb`, `switchboard.rb` (and potentially more in future)
- **Current CI:** Audits and style-checks ALL formulas on push/PR to main
- **Release flow:** GoReleaser pushes formula update commit to shared tap → tap CI runs
- **Key constraint:** Multiple arcaven projects update this tap — CI failures may be from *other* formulas, not ThreeDoors

---

## Discussion Summary

### Agent Ownership: pr-shepherd vs. release-manager

📋 **John (PM):** The shared tap changes everything. With a dedicated tap repo, a release-manager agent made sense — it owned the whole lifecycle. With a *shared* tap, monitoring is per-formula, and failures might not even be ours. The question isn't "who monitors the tap repo" — it's "who monitors ThreeDoors releases end-to-end."

pr-shepherd's job is branch hygiene and PR health. A release is not a PR. But a GoReleaser-generated *commit* to the tap repo could be a PR (if we configure GoReleaser to create PRs instead of direct pushes). If it's a PR, pr-shepherd could naturally monitor it. If it's a direct commit, it's outside pr-shepherd's domain.

🏗️ **Winston (Architect):** The shared tap introduces a **blast radius problem**. If switchboard's formula breaks `brew audit`, tap CI fails, and it *looks* like ThreeDoors's release is broken too — but it's not our fault. We need formula-level isolation in monitoring.

Two architectural approaches:

1. **Formula-level CI** — tap CI runs each formula independently, reports per-formula status. ThreeDoors only cares about its own formula's pass/fail.
2. **Cross-repo commit status** — after GoReleaser pushes to the tap, a GitHub Action in ThreeDoors polls for the specific commit's CI result, filtered to threedoors-related checks.

I recommend approach 1 (formula-level CI) because it benefits ALL projects using the shared tap, not just ThreeDoors.

💻 **Amelia (Dev):** Current tap CI loops over all formulas in a single job:
```yaml
for f in Formula/*.rb; do
  brew audit --strict "arcaven/tap/$name"
done
```
If *any* formula fails, the whole workflow fails. We can't tell which formula broke. Fix: use a **matrix strategy** — one job per formula. Each formula gets its own pass/fail status. GoReleaser's commit to the tap triggers CI, and we check only the threedoors matrix job.

Concretely:
```yaml
jobs:
  audit:
    strategy:
      matrix:
        formula: [threedoors, switchboard]
    steps:
      - brew audit --strict "arcaven/tap/${{ matrix.formula }}"
```

This way, `threedoors` can pass even if `switchboard` fails.

🧪 **Murat (Test Architect):** Formula-level isolation is essential for accurate failure attribution. But there's a second layer: **the commit that triggered CI**. GoReleaser pushes a commit that only modifies `Formula/threedoors.rb`. The tap CI should ideally detect *which formulas changed* and only test those (plus a periodic full audit). This is optimization — the matrix approach works first.

For ThreeDoors's cross-repo monitoring, the verification workflow should:
1. Wait for tap CI to start (GoReleaser commit triggers it)
2. Poll for the specific workflow run triggered by that commit
3. Check the `threedoors` matrix job specifically
4. Report pass/fail back to ThreeDoors (commit status or issue)

🏃 **Bob (SM):** Process question: should GoReleaser push directly to `main` on the tap repo, or create a PR? If it creates a PR:
- pr-shepherd can monitor it naturally (it's a PR!)
- CI runs on the PR before merge
- Human can review the formula change
- But: adds friction to every release

If it pushes directly:
- Faster releases (no PR review needed)
- But: broken formulas land on main immediately
- Need the separate verification workflow

### Technical Monitoring Approach

💻 **Amelia (Dev):** For the ThreeDoors side, here's the concrete design:

**Option A: `release-verify.yml` workflow in ThreeDoors** (recommended)
- Triggers on `release: published`
- Waits for GoReleaser to push to tap (GoReleaser runs in the same release workflow, so the tap commit happens before this workflow triggers — use `workflow_run` trigger instead)
- Polls `gh run list --repo arcaven/homebrew-tap --branch main --limit 1` to find the CI run triggered by GoReleaser's commit
- Checks result, posts commit status to release tag
- On failure: opens GitHub issue with details

**Option B: `repository_dispatch` from tap to ThreeDoors**
- Tap CI sends event to ThreeDoors on completion
- Cleanest event-driven approach
- Requires PAT with cross-repo dispatch permissions
- More complex setup

**Option C: GoReleaser creates PRs to tap (enables pr-shepherd)**
- GoReleaser config: `pull_request: { enabled: true }`
- pr-shepherd monitors the PR naturally
- Adds human review step (could be good or bad)
- Most natural fit with existing multiclaude architecture

I'd recommend **Option A for automation + Option C as future consideration**. Option A is simplest to implement now. Option C aligns with the multiclaude agent model but adds release friction.

🏗️ **Winston (Architect):** Option A with one refinement: the polling needs to identify the *specific* CI run triggered by GoReleaser's commit, not just "the latest run." GoReleaser's commit has a known message format (`chore(formula): update threedoors to vX.Y.Z`). Use `gh api` to find the commit SHA, then find the workflow run for that SHA.

```bash
# Find the GoReleaser commit in tap repo
gh api repos/arcaven/homebrew-tap/commits \
  --jq '.[0] | select(.commit.message | startswith("chore(formula): update threedoors"))' \
  | jq -r '.sha'

# Find CI run for that commit
gh run list --repo arcaven/homebrew-tap --commit $SHA --json status,conclusion
```

### Failure Recovery

🧪 **Murat (Test Architect):** Failure scenarios for ThreeDoors in the shared tap:

| Failure | Cause | Detection | Recovery |
|---------|-------|-----------|----------|
| `brew audit` fails for threedoors | Formula syntax/standards | Matrix job fails | Fix `.goreleaser.yml` template, re-release |
| `brew install` fails | Bad URL, checksum mismatch | Install job fails | Verify GoReleaser archive config, re-release |
| `brew test` fails | `--version` output mismatch | Test job fails | Fix ldflags alignment, re-release |
| Tap CI fails for *other* formula | Not our problem | Matrix isolates it | No action needed (if matrix CI is implemented) |
| GoReleaser can't push to tap | Token expired/revoked | GoReleaser step fails | Refresh `HOMEBREW_TAP_TOKEN` secret |

Recovery is always: fix the root cause in ThreeDoors, tag a patch release (e.g., `v1.0.1`), GoReleaser regenerates the formula. Never manually edit the tap formula — it's auto-generated.

📋 **John (PM):** One more thing — the `release-verify` workflow should have a **timeout**. If tap CI doesn't start within 10 minutes or doesn't complete within 30 minutes, something is wrong. Alert via GitHub issue. Don't let it hang forever.

### Tap CI Recommendations

🧪 **Murat (Test Architect):** The current tap CI only does `brew audit --strict` and `brew style`. It should also:

**P0 (must have):**
- `brew install --formula` — verify the formula actually installs
- `brew test` — run the test block
- Matrix strategy for per-formula isolation

**P1 (should have):**
- macOS arm64 runner (current `macos-latest` is arm64, so this is covered)
- `brew uninstall` cleanup
- Diff detection (only test changed formulas on PRs)

**P2 (nice to have):**
- macOS Intel runner (for x86_64 coverage)
- Linux runner (if we distribute Linux bottles via the tap)
- `brew install --build-from-source` (for homebrew-core readiness)

---

## Adopted Approach

### 1. Tap CI Enhancement (in `arcaven/homebrew-tap`)
- Convert CI to **matrix strategy** — one job per formula for isolated pass/fail
- Add `brew install` and `brew test` steps (currently only audit/style)
- This benefits ALL arcaven projects, not just ThreeDoors

### 2. Release Verification Workflow (in ThreeDoors)
- New `.github/workflows/release-verify.yml`
- Triggers on `workflow_run` (after GoReleaser release workflow completes)
- Identifies GoReleaser's commit to tap repo by commit message pattern
- Polls tap CI for that specific commit's result (threedoors matrix job)
- Posts commit status to ThreeDoors release tag
- On failure: opens GitHub issue with failure details and recovery steps
- Timeout: 30 minutes max

### 3. Agent Model: No Persistent Release-Manager
- No new persistent agent — releases are infrequent
- GitHub Action handles detection (automated, deterministic)
- On failure: supervisor spawns a worker to investigate/fix
- Document release knowledge in a **spawnable agent prompt** (`agents/release-manager.md`) for when workers need release-specific context

### 4. GoReleaser Config Update
- Target shared repo: `arcaven/homebrew-tap` (not `arcaven/homebrew-threedoors`)
- Formula path: `Formula/threedoors.rb`
- Commit message pattern: `chore(formula): update threedoors to {{ .Tag }}`
- Token: existing `HOMEBREW_TAP_TOKEN` secret

### Rationale
The shared tap means formula-level CI isolation is critical. A persistent release-manager agent is overkill for infrequent releases — GitHub Actions handle the happy path, workers handle failures. The matrix CI strategy benefits all arcaven projects.

---

## Rejected Options

### Option A: Dedicated `arcaven/homebrew-threedoors` Tap Repo
**Rejected because:** The shared tap `arcaven/homebrew-tap` already exists with ThreeDoors and switchboard formulas. Creating a separate tap would fragment the distribution. Users would need multiple `brew tap` commands for different arcaven tools. The shared tap is the standard pattern for organizations with multiple CLI tools.

### Option B: pr-shepherd Owns Cross-Repo Monitoring
**Rejected because:** pr-shepherd's domain is branch hygiene and PR health. Release lifecycle is a different concern with different failure modes, different recovery patterns, and different cadence. Overloading pr-shepherd dilutes its responsibility. The shared tap adds complexity (multi-formula, multi-project) that's outside pr-shepherd's scope.

### Option C: Persistent Release-Manager Agent
**Rejected because:** Releases happen infrequently (once per sprint or less). A persistent agent sitting idle 99% of the time is waste. GitHub Actions handle automated detection deterministically. Only spawn a worker when human/agent intervention is needed for failure recovery.

### Option D: GoReleaser Creates PRs to Tap (pr-shepherd Integration)
**Rejected for now because:** Adds friction to every release (PR review step). The current tap CI is lightweight enough that direct-push-to-main is safe, especially with matrix CI for isolation. Could be reconsidered if formula quality issues become frequent. Noted as a future option.

### Option E: `repository_dispatch` Event-Driven Monitoring
**Rejected because:** More complex setup than polling. Requires PAT with cross-repo dispatch scope. The polling approach (`gh run list`) is simpler, well-understood, and sufficient for the infrequent release cadence. Event-driven adds unnecessary infrastructure.

### Option F: GitHub Webhooks to Notify ThreeDoors
**Rejected because:** Requires webhook server infrastructure or a GitHub App. Massive overengineering for checking CI status a few times per month. `gh run list` polling in a GitHub Action is simpler by orders of magnitude.
