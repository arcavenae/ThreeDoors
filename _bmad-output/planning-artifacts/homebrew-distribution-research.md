# Homebrew Public Distribution Research — ThreeDoors

**Date:** 2026-03-08
**Author:** proud-otter (worker agent)
**Status:** Complete

---

## Executive Summary

ThreeDoors can be distributed via Homebrew through two paths: a **custom tap** (immediate, self-hosted) and **homebrew-core** (official, curated, requires notability). The recommended strategy is to ship a custom tap now using GoReleaser, then graduate to homebrew-core when community adoption reaches the notability threshold (75+ GitHub stars for third-party submissions, 225+ for self-submissions).

ThreeDoors already has partial Homebrew infrastructure: a formula template (`Formula/threedoors.rb`), CI binary builds, and Apple code signing. The main gaps are: GoReleaser integration, a LICENSE file, semantic version tags, and supply chain signing.

---

## 1. Homebrew Distribution Models

### homebrew-core (Official)

**What it is:** The default Homebrew repository. Formulae here are installed with just `brew install <name>` — no tap needed.

**Acceptance criteria:**
- **Notability:** 75+ GitHub stars, OR 30+ forks, OR 30+ watchers (third-party submission). For self-submitted formulae (repo owner submits), thresholds are 3x: 225+ stars, 90+ forks, 90+ watchers
- **License:** Must have a DFSG-compatible license (MIT, Apache-2.0, BSD, GPL, etc.)
- **Stable release:** Must have at least one non-pre-release version
- **Builds from source:** No prebuilt binaries — formula must compile from source tarball
- **Cross-platform:** Must build on macOS (Intel + ARM) and Linux x86_64
- **Meaningful test block:** `test do` must verify actual functionality
- **Passes audit:** `brew audit --strict --new --online` must pass cleanly

**Review process:**
- Submit PR to `Homebrew/homebrew-core`
- BrewTestBot runs CI automatically (builds, tests, audits on all platforms)
- One maintainer approval required for merge
- Typical timeline: days to ~1 week if CI passes and no issues raised
- If no response after a week, post a bump comment

**Common rejection reasons:**
1. Below notability thresholds (most common)
2. No license or non-free license
3. Alpha/beta-only releases (no stable version)
4. Failing `brew audit --strict --new --online`
5. Missing or inadequate `test do` block
6. Formula doesn't build from source on all platforms
7. Duplicate functionality of existing formula

**Maintenance after acceptance:**
- Autobump: Homebrew detects new releases and auto-updates formula (default behavior)
- If build breaks on new OS/Go versions, submitter is expected to help fix
- Abandoned formulae may eventually be removed

### Custom Tap (Self-Hosted)

**What it is:** A third-party Homebrew repository hosted on GitHub (e.g., `arcaven/homebrew-threedoors`).

**How users install:**
```bash
brew tap arcaven/threedoors
brew install threedoors
```

**Advantages:**
- No notability requirements — ship immediately
- Full control over formula, release cadence, and distribution
- Can distribute prebuilt binaries (faster install) OR build from source
- No review process — publish when ready

**Disadvantages:**
- Users must know the tap name
- Less discovery — not in `brew search` results by default
- No Homebrew CI infrastructure (you run your own)
- Perception of being "unofficial"

**Tap repository structure:**
```
arcaven/homebrew-threedoors/
├── Formula/
│   └── threedoors.rb
└── README.md
```

### Comparison

| Aspect | Custom Tap | homebrew-core |
|--------|-----------|---------------|
| Entry bar | None | 75-225+ stars |
| Install command | `brew tap X && brew install Y` | `brew install Y` |
| Review process | None | Maintainer review |
| Build method | Binary or source | Source only |
| Autobump | Self-managed | Automatic |
| Discoverability | Low | High |
| Maintenance | Self-managed | Community-supported |

---

## 2. Technical Requirements

### Formula Structure for Go Projects

**Custom tap (binary distribution — current approach):**
```ruby
class Threedoors < Formula
  desc "TUI task manager showing only three tasks at a time"
  homepage "https://github.com/arcaven/ThreeDoors"
  version "1.0.0"
  license "MIT"

  on_arm do
    url "https://github.com/arcaven/ThreeDoors/releases/download/v1.0.0/threedoors-darwin-arm64"
    sha256 "abc123..."
  end

  on_intel do
    url "https://github.com/arcaven/ThreeDoors/releases/download/v1.0.0/threedoors-darwin-amd64"
    sha256 "def456..."
  end

  def install
    binary_name = Hardware::CPU.arm? ? "threedoors-darwin-arm64" : "threedoors-darwin-amd64"
    bin.install binary_name => "threedoors"
  end

  test do
    assert_match "ThreeDoors", shell_output("#{bin}/threedoors --version")
  end
end
```

**homebrew-core (source build — required for core):**
```ruby
class Threedoors < Formula
  desc "TUI task manager showing only three tasks at a time"
  homepage "https://github.com/arcaven/ThreeDoors"
  url "https://github.com/arcaven/ThreeDoors/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "abc123..."
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X github.com/arcaven/ThreeDoors/internal/cli.Version=#{version}
      -X github.com/arcaven/ThreeDoors/internal/cli.Commit=#{tap.user}
      -X github.com/arcaven/ThreeDoors/internal/cli.BuildDate=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/threedoors"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/threedoors --version")
    assert_match "ThreeDoors", shell_output("#{bin}/threedoors --help")
  end
end
```

### How Homebrew Builds Go Projects

- `depends_on "go" => :build` — Go is build-only, not a runtime dependency
- `std_go_args` helper sets `-trimpath`, `-ldflags "-s -w"`, and output path
- Go modules: dependencies fetched automatically via `go build` (no `go_resource` needed)
- Homebrew manages `GOPATH` and `GOCACHE`
- Result: statically linked binary, no dynamic library dependencies
- Bottles: `cellar :any_skip_relocation` (Homebrew CI builds bottles after PR merge)

### Signing Requirements

**homebrew-core:** No code signing required. Homebrew builds from source in its own CI. Bottles are signed by Homebrew's infrastructure.

**Custom tap:** No signing required, but recommended for trust:
- **Cosign** (keyless via GitHub OIDC): Sign checksums file, transitively covers all artifacts
- **Apple code signing:** Already implemented in our CI — separate from Homebrew signing
- **SLSA provenance:** Increasingly expected for security-conscious users

### Versioning Requirements

- **Semantic versioning mandatory:** `v1.0.0`, `v1.1.0`, `v2.0.0-beta.1`
- **Git tags:** Formula URL points to tag-based release tarball
- **No alpha-only projects in homebrew-core** — must have at least one stable release
- **Our current state:** Alpha auto-releases on every push (e.g., `alpha-20260308-abc1234`). Need to add semver tags for stable releases.

### License Requirements

- **MIT is fully compatible** — DFSG-compliant, widely accepted
- **BLOCKER: No LICENSE file exists in the repository.** Must add one.
- Formula declares `license "MIT"` — this must match the actual LICENSE file

### Current Project Compatibility

| Requirement | Status | Notes |
|------------|--------|-------|
| Go modules | ✅ | `go.mod` with all dependencies |
| Cross-platform build | ✅ | darwin-arm64, darwin-amd64, linux-amd64 in CI |
| `--version` flag | ✅ | Implemented via ldflags |
| `--help` flag | ✅ | Via cobra |
| Apple code signing | ✅ | CI pipeline exists |
| LICENSE file | ❌ | **Must add** |
| Semantic version tags | ❌ | **Must add** |
| GoReleaser | ❌ | **Must add** |
| Cosign signing | ❌ | Nice to have |
| SLSA provenance | ❌ | Nice to have |

---

## 3. CI/CD Pipeline for Homebrew

### GoReleaser — The Standard Tool

GoReleaser is the de facto standard for Go project releases. It:
1. Cross-compiles binaries for all target platforms
2. Creates tar.gz/zip archives with checksums
3. Creates GitHub Releases with changelogs
4. Generates and pushes Homebrew formulae to tap repositories
5. Optionally signs artifacts with cosign

**Integration with our existing CI:**
- Current: Manual binary build in `build-binaries` job → manual release → manual formula update
- With GoReleaser: Single `goreleaser release` command replaces all three steps
- Triggered by: pushing a semver git tag (e.g., `git tag v1.0.0 && git push --tags`)
- Workflow: separate from PR CI — runs only on tag pushes

### GitHub Release Workflow

```
Developer tags release (git tag v1.0.0)
    ↓
GitHub Actions triggered (on: push: tags: 'v*')
    ↓
GoReleaser runs:
    ├── Cross-compile binaries
    ├── Create archives + checksums
    ├── Sign checksums (cosign, optional)
    ├── Generate changelog
    ├── Create GitHub Release
    └── Push formula to tap repo
    ↓
Users run: brew update && brew upgrade threedoors
```

### SLSA Provenance / Supply Chain Security

- **SLSA Level 3:** Achievable via `slsa-github-generator` GitHub Actions
- Records: source repo, commit SHA, builder identity, build instructions
- Produces attestation file that users can verify
- **Not required** by homebrew-core, but increasingly expected
- GoReleaser integrates with `slsa-github-generator` via reusable workflow

### Reproducible Builds

- Use `-trimpath` in ldflags (removes local paths from binary) — `std_go_args` and GoReleaser both do this
- Pin Go version in `go.mod` (we have `go 1.25.4`)
- Avoid embedding build-time values that vary (timestamps) — or make them reproducible
- `CGO_ENABLED=0` for pure Go builds (no C dependencies)
- Our `modernc.org/sqlite` is pure Go — no CGO needed

### CI Gaps to Fill

| Gap | Priority | Phase |
|-----|----------|-------|
| GoReleaser workflow (tag-triggered) | P0 | Phase 1 |
| `brew audit` in CI | P1 | Phase 2 |
| Cosign checksum signing | P2 | Phase 2 |
| SLSA provenance | P2 | Phase 2 |
| `brew install --build-from-source` test | P1 | Phase 2 |
| `brew test` validation | P1 | Phase 2 |

---

## 4. Audit & Quality Standards

### `brew audit` Checks

`brew audit --strict --new --online` validates:
- Ruby style compliance of formula file
- Valid download URL (HTTPS, accessible)
- SHA256 checksum matches downloaded artifact
- License declared and valid
- Description present, properly formatted, not too long
- Homepage accessible
- `test do` block exists and is meaningful
- No deprecated Homebrew APIs
- Proper dependency declarations
- `--new` implies `--strict` and `--online`
- `--online` checks GitHub API for stars/activity (notability)

### `brew test` Expectations

- Runs in isolated temp directory
- Must verify actual functionality (not just file existence)
- Common patterns for CLI tools:
  - Check `--version` output matches expected version
  - Check `--help` output contains expected strings
  - For TUI apps: test non-interactive modes only
- Must pass on all supported platforms (macOS Intel, ARM, Linux)

### Our Current CI vs Homebrew Expectations

| Check | Our CI | Homebrew Expects | Gap? |
|-------|--------|-----------------|------|
| Code formatting | ✅ gofumpt | N/A | No |
| Linting | ✅ golangci-lint | N/A | No |
| Unit tests | ✅ with race detector | N/A | No |
| Coverage floor | ✅ 75% | N/A | No |
| Build verification | ✅ go build | ✅ `brew install --build-from-source` | **Yes** |
| Formula syntax | ✅ `ruby -c` | ✅ `brew audit` | **Yes** |
| Formula test | ❌ | ✅ `brew test` | **Yes** |
| Binary signing | ✅ Apple codesign | ❌ Not needed | No |
| Checksum signing | ❌ | ✅ Cosign (recommended) | **Nice to have** |
| Provenance | ❌ | ✅ SLSA (recommended) | **Nice to have** |

### Security Scanning

- homebrew-core doesn't mandate specific security scanning tools
- But formulae with known vulnerabilities are flagged and may be removed
- Our existing `go vet` + `golangci-lint` cover common Go security issues
- Adding `govulncheck` to CI would strengthen the security posture

---

## 5. Submission Process

### Step-by-Step: Current State → homebrew-core

#### Phase 1: Immediate (Custom Tap)

1. **Add LICENSE file** — MIT, copy from standard template. BLOCKER.
2. **Create tap repository** — `arcaven/homebrew-threedoors` on GitHub
3. **Add GoReleaser config** — `.goreleaser.yml` with Homebrew integration
4. **Add GoReleaser workflow** — `.github/workflows/release.yml`, triggered on `v*` tags
5. **Tag first release** — `git tag v0.1.0 && git push origin v0.1.0`
6. **Verify installation** — `brew tap arcaven/threedoors && brew install threedoors`
7. **Update README** — Add installation instructions for Homebrew

#### Phase 2: Quality Hardening

1. **Add `brew audit` to CI** — Install Homebrew on CI runner, run audit
2. **Add cosign signing** — Keyless via GitHub OIDC
3. **Add SLSA provenance** — Via `slsa-github-generator`
4. **Add `govulncheck`** — Go vulnerability scanner
5. **Write source-build formula variant** — For eventual homebrew-core submission

#### Phase 3: homebrew-core Submission (When Ready)

1. **Verify notability** — 75+ stars (third-party) or 225+ (self-submission)
2. **Prepare source-build formula:**
   ```bash
   export HOMEBREW_NO_INSTALL_FROM_API=1
   brew tap --force homebrew/core
   brew create https://github.com/arcaven/ThreeDoors/archive/refs/tags/vX.Y.Z.tar.gz --go
   ```
3. **Edit formula** — Add proper `install`, `test do`, metadata
4. **Validate locally:**
   ```bash
   brew install --build-from-source threedoors
   brew test threedoors
   brew audit --strict --new --online threedoors
   ```
5. **Submit PR** to `Homebrew/homebrew-core`:
   - Title: `threedoors X.Y.Z (new formula)`
   - Formula file at `Formula/t/threedoors.rb`
   - Brief description and homepage link
6. **Wait for BrewTestBot CI + maintainer review** (~days to 1 week)

### Maintenance After Acceptance

- **Autobump** is enabled by default — new releases auto-detected and formula updated
- **Build failures** on new platforms/Go versions are submitter's responsibility
- **No manual PRs needed** for routine version bumps
- **Can add `no_autobump!`** to formula if needed (rare)

### Who Needs to Be Involved

- **Repository owner** (arcaven) — creates tap repo, manages secrets
- **No special Homebrew account needed** — submit via GitHub PR
- **`HOMEBREW_TAP_TOKEN`** — already configured in CI secrets

---

## 6. Prototype Artifacts

### Draft Formula

See: `homebrew-formula-draft.rb` — source-build formula suitable for homebrew-core, with GoReleaser-compatible variant for the custom tap.

### Draft GoReleaser Config

See: `goreleaser-draft.yml` — complete GoReleaser configuration with Homebrew integration, cosign signing, and multi-platform builds.

### Required Changes to Build Process

| Change | File | Description |
|--------|------|-------------|
| Add LICENSE | `LICENSE` (new) | MIT license file |
| Add GoReleaser config | `.goreleaser.yml` (new) | Release automation |
| Add release workflow | `.github/workflows/release.yml` (new) | Tag-triggered releases |
| Create tap repo | `arcaven/homebrew-threedoors` (new repo) | Homebrew tap |
| Update README | `README.md` | Add Homebrew install instructions |
| Add `make release-tag` | `Makefile` | Helper to create semver tags |

### What Does NOT Change

- Existing CI pipeline (PR validation, quality gates)
- Existing alpha auto-release process
- Development workflow (branches, PRs, story-driven)
- Apple code signing pipeline (orthogonal to Homebrew)

---

## References

- [Homebrew Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Adding Software](https://docs.brew.sh/Adding-Software-to-Homebrew)
- [Homebrew How to Open a PR](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request)
- [Homebrew Bottles](https://docs.brew.sh/Bottles)
- [Homebrew Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [GoReleaser Homebrew Taps](https://goreleaser.com/customization/homebrew/)
- [GoReleaser Signing](https://goreleaser.com/customization/sign/)
- [GoReleaser SLSA Provenance](https://goreleaser.com/blog/slsa-generation-for-your-artifacts/)
- [Homebrew Notability PR #20981](https://github.com/Homebrew/brew/pull/20981)
- [Homebrew Build Provenance (Trail of Bits)](https://blog.trailofbits.com/2024/05/14/a-peek-into-build-provenance-for-homebrew/)
