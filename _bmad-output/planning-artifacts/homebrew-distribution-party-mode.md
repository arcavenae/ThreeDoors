# Party Mode: Homebrew Public Distribution Strategy

**Date:** 2026-03-08
**Participants:** PM (John), Architect (Winston), Dev (Amelia), QA (Quinn), SM (Bob), Test Architect (Murat)
**Topic:** Distributing ThreeDoors via Homebrew publicly — strategy, quality gates, risks, process changes

---

## Discussion Summary

### Strategy: Custom Tap First, Then Graduate

📋 **John (PM):** The data is clear — homebrew-core requires 75+ stars for third-party submissions, or 225+ for self-submissions. We have ~0 stars today. Custom tap is the only viable path right now, and it's the right one. Ship via `brew tap arcavenae/threedoors && brew install threedoors` immediately. Graduate to homebrew-core when we hit notability thresholds. This is the standard pattern — even major tools like `goreleaser` started with custom taps.

🏗️ **Winston (Architect):** Agreed. The custom tap approach aligns with our existing CI. We already have a `Formula/threedoors.rb` template and a `update-homebrew` CI job. The gap is that our current formula distributes **prebuilt binaries** — homebrew-core requires **build-from-source**. We need two formula variants: one for our tap (binary distribution, faster installs) and one designed for homebrew-core submission (source build with `depends_on "go" => :build`). GoReleaser bridges this gap cleanly.

💻 **Amelia (Dev):** Current state: `Formula/threedoors.rb` uses `VERSION_PLACEHOLDER`/`SHA256_*_PLACEHOLDER` with prebuilt binaries. `Makefile` has `LDFLAGS` for version injection. CI builds darwin-arm64, darwin-amd64, linux-amd64. The missing piece is GoReleaser — it replaces our manual binary-build CI job, generates checksums, creates GitHub releases, and auto-updates the tap formula. One `.goreleaser.yml` replaces ~50 lines of CI config.

### Quality Gates for CI

🧪 **Quinn (QA):** For homebrew-core, `brew audit --strict --new --online` is the acceptance gate. Our current CI covers: gofumpt, golangci-lint, go vet, tests with race detector, 75% coverage floor, benchmarks, Docker E2E. What we're missing: `brew audit` validation in CI, formula syntax checking beyond `ruby -c`, and `brew test` simulation.

🧪 **Murat (Test Architect):** Risk calculation: the CI gaps are low-risk for the custom tap phase but blocking for homebrew-core. I'd add these quality gates in phases:

**Phase 1 (Custom Tap):**
- GoReleaser dry-run in CI (validates build config)
- Formula syntax check (already have `ruby -c` in `test-dist`)
- Checksum verification post-release

**Phase 2 (Pre-homebrew-core):**
- `brew audit --strict --new --online` in CI (requires Homebrew installed in CI runner)
- `brew install --build-from-source` test
- `brew test` validation
- SLSA provenance generation (increasingly expected)
- Cosign signing of checksums

### Process Changes

🏃 **Bob (SM):** Release cadence impact: today we auto-release alpha builds on every push to main. For Homebrew, we need **semantic versioned tags** (v1.0.0, v1.1.0). This means:
1. Define a release process: when do we tag? Per-epic? Per-sprint?
2. GoReleaser triggers on tags, not pushes — separate release workflow
3. CHANGELOG.md generation (GoReleaser handles this)
4. Decision: keep alpha auto-releases AND add tagged stable releases? (Recommended yes)

📋 **John (PM):** The cadence question is critical. I'd recommend: alpha auto-releases continue for internal testing. Stable releases (semver tags) happen when meaningful user-facing changes land — roughly per-epic completion. This gives us both fast iteration and stable distribution.

🏗️ **Winston (Architect):** Process-wise, the tap repository (`arcavenae/homebrew-threedoors`) needs to exist. GoReleaser pushes formula updates there on each tagged release. We need a `HOMEBREW_TAP_TOKEN` (personal access token with repo write access to the tap repo) — we already have this secret configured in CI.

### Risks

🏗️ **Winston (Architect):** Key risks:

1. **Security: Supply chain attacks** — Our binaries are unsigned for Homebrew (Apple signing is separate). Cosign keyless signing via GitHub OIDC is the mitigation. Low effort, high trust signal.

2. **Maintenance burden** — Once in homebrew-core, build failures on new Go versions or macOS versions are our problem. The tap has no SLA. homebrew-core has implicit expectations.

3. **Version management** — We're on Go 1.25.4 which is bleeding edge. homebrew-core's Go version may lag. We need to ensure builds work with Homebrew's Go version (currently ~1.24.x in homebrew-core). May need to relax our `go.mod` version.

4. **CGO_ENABLED=0** — We use `modernc.org/sqlite` which is pure Go. But if we ever need CGO, Homebrew handles it differently (needs `uses_from_macos "sqlite"` etc). Current state is fine.

5. **Missing LICENSE file** — **BLOCKER**. We have `license "MIT"` in the formula but no LICENSE file in the repo. Must add one.

💻 **Amelia (Dev):** Additional technical risk: our `--version` flag injects via `-X main.version`. GoReleaser uses `{{.Version}}` template which maps to the git tag. Need to verify the ldflags alignment. Also: we have two binaries (`threedoors` and `threedoors-mcp`) — do we distribute both via Homebrew? Recommend: `threedoors` only for now, `threedoors-mcp` is a separate concern.

🧪 **Murat (Test Architect):** Test risk: our `test do` block only checks `--version`. For homebrew-core, reviewers may want more robust testing. Recommend adding a `--help` check and a basic task-file-load test (create temp file, verify it reads). But for the custom tap, `--version` is sufficient.

### Impact on Development Workflow

🏃 **Bob (SM):** Minimal impact on daily workflow. Changes:
1. New `make release-tag` target to create semver tags
2. New GitHub Actions workflow triggered on tags (GoReleaser)
3. Keep existing CI as-is for PR validation
4. New responsibility: don't break the formula (test `brew install` periodically)

---

## Adopted Approach

**Phase 1: Custom Tap with GoReleaser (Immediate)**
- Add LICENSE file (MIT)
- Create `arcavenae/homebrew-threedoors` tap repository
- Add `.goreleaser.yml` configuration
- Add GoReleaser GitHub Actions workflow (triggered on semver tags)
- Tag first stable release (v0.1.0 or v1.0.0)
- Verify `brew tap arcavenae/threedoors && brew install threedoors` works

**Phase 2: CI Hardening (Before homebrew-core submission)**
- Add `brew audit` to CI
- Add cosign signing + SLSA provenance
- Add `brew install --build-from-source` test
- Reproducible build verification (`-trimpath`, pinned Go version)

**Phase 3: homebrew-core Submission (When notability thresholds met)**
- Write source-build formula (not binary distribution)
- Pass `brew audit --strict --new --online`
- Submit PR to Homebrew/homebrew-core
- Set up autobump maintenance

**Rationale:** Custom tap gives us immediate distribution with minimal process change. GoReleaser automates the release pipeline. We build quality gates incrementally — no need to over-engineer for a custom tap. homebrew-core submission is gated on community adoption (stars), not technical readiness.

## Rejected Options

### Option A: Skip Custom Tap, Submit Directly to homebrew-core
**Rejected because:** We don't meet the notability thresholds (need 75+ or 225+ stars). Would be immediately rejected. Waste of effort.

### Option B: Manual Formula Management (No GoReleaser)
**Rejected because:** We already have manual CI for binary builds, but it's fragile (placeholder substitution via sed). GoReleaser is the industry standard for Go projects, handles checksums, changelogs, and Homebrew formula generation. The maintenance cost of not using GoReleaser is higher.

### Option C: Distribute Both `threedoors` and `threedoors-mcp` via Homebrew
**Rejected because:** MCP server is a developer/integration tool, not an end-user tool. Adding complexity to the formula for a niche use case. Can be added later as a separate formula or cask if demand exists.

### Option D: Wait for homebrew-core Before Any Homebrew Distribution
**Rejected because:** Custom tap provides real user value now. Waiting for stars is circular — we need distribution to get users to get stars. Ship the tap, iterate, graduate.

### Option E: Use Homebrew Cask Instead of Formula
**Rejected because:** Casks are for GUI apps (`.app` bundles, `.dmg`, `.pkg`). ThreeDoors is a CLI/TUI tool — formula is the correct distribution mechanism. We do have DMG/PKG builds but those are alternative distribution, not the primary Homebrew path.
