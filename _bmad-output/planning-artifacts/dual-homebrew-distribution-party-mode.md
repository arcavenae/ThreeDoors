# Party Mode: Dual Homebrew Distribution (Stable + Alpha)

**Date:** 2026-03-08
**Participants:** PM (John), Architect (Winston), Dev (Amelia), QA (Quinn), SM (Bob), Test Architect (Murat)
**Topic:** Dual Homebrew distribution strategy: `threedoors` (stable, GoReleaser, v* tags) + `threedoors-a` (alpha, every main push)

---

## Discussion Summary

### The Core Question: Binary Naming and Formula Strategy

📋 **John (PM):** The user requirement is clear: two parallel Homebrew distribution channels. `brew install arcavenae/tap/threedoors` gets you stable releases. `brew install arcavenae/tap/threedoors-a` gets you bleeding-edge alpha builds updated on every push to main. The critical design question is: should the alpha binary itself be named `threedoors-a` or `threedoors`? If both install a binary called `threedoors`, they conflict. If the alpha binary is `threedoors-a`, users get both installed simultaneously with zero friction.

🏗️ **Winston (Architect):** The binary MUST be named differently. Here's why:

1. **Homebrew conflict prevention:** If two formulae install the same binary name to `/usr/local/bin`, Homebrew will either refuse to link one (`keg_only`) or require `conflicts_with`. Neither is clean UX. Different binary names means both formulae can be installed and linked simultaneously.

2. **Apple code signing:** Apple codesign uses a bundle identifier (passed via `--identifier`), not the filename. The binary can be named anything. We already sign `threedoors-darwin-arm64` and `threedoors-darwin-amd64` -- the filename is irrelevant to notarization. No constraint here.

3. **User clarity:** When a user types `threedoors-a`, they know they're running the alpha channel. No version confusion. No "wait, which one did I install?" moments.

💻 **Amelia (Dev):** Concrete proposal: the alpha build produces a binary named `threedoors-a`. The Go build just needs `go build -o threedoors-a ./cmd/threedoors`. The version string already includes alpha markers (`0.1.0-alpha.20260308.abc1234`). The binary name change is the only code-level difference -- same source, same entrypoint, different output name.

### Version Scheme

📋 **John (PM):** Two version tracks:
- **Stable:** Semantic versions from git tags: `v0.1.0`, `v1.0.0`, `v1.1.0`. GoReleaser handles this.
- **Alpha:** `0.1.0-alpha.YYYYMMDD.SHA7` -- already implemented in CI. The date+SHA gives monotonic ordering and traceability.

🏗️ **Winston (Architect):** The alpha version scheme is already solid. The CI generates `0.1.0-alpha.$(date -u +%Y%m%d).${GITHUB_SHA::7}` and tags releases as `alpha-YYYYMMDD-SHA7`. For the Homebrew formula, we need the version to be parseable. Homebrew's version comparison works with pre-release segments. `0.1.0-alpha.20260308.abc1234` sorts correctly.

One refinement: the alpha formula should use `version` explicitly rather than inferring from the archive name, since alpha archives won't follow GoReleaser's naming convention.

### Tap Structure

🏗️ **Winston (Architect):** The tap repo is `arcavenae/homebrew-tap`. Structure:

```
arcavenae/homebrew-tap/
├── Formula/
│   ├── threedoors.rb          # Stable: GoReleaser auto-updates on v* tags
│   └── threedoors-a.rb        # Alpha: CI auto-updates on every main push
└── README.md
```

Both formulae in the same tap. Users install with:
```bash
brew install arcavenae/tap/threedoors      # stable
brew install arcavenae/tap/threedoors-a    # alpha
```

No `conflicts_with` needed because the binaries have different names. Both can be installed simultaneously. `brew upgrade` works independently for each.

💻 **Amelia (Dev):** GoReleaser already pushes `threedoors.rb` to the tap on v* tags. For the alpha formula, we need a CI step in the existing `ci.yml` workflow (the `release` job) that:

1. Builds the binary as `threedoors-a-{os}-{arch}`
2. Creates/updates the GitHub release (already happens)
3. Updates `threedoors-a.rb` in the tap repo with new version, URL, and SHA256

This is a `gh api` or direct git push to the tap repo, using `HOMEBREW_TAP_TOKEN`.

### CI Workflow Changes

💻 **Amelia (Dev):** Current alpha flow in `ci.yml`:

```
push to main → build-binaries → sign-and-notarize → release (GitHub pre-release)
```

Changes needed:

1. **build-binaries job:** Add `threedoors-a` binary builds alongside existing `threedoors` builds:
   ```bash
   GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o threedoors-a-darwin-arm64 ./cmd/threedoors
   GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o threedoors-a-darwin-amd64 ./cmd/threedoors
   GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o threedoors-a-linux-amd64 ./cmd/threedoors
   ```

2. **sign-and-notarize job:** Sign the `threedoors-a-*` binaries with a distinct identifier (e.g., `com.arcavenae.threedoors-a`).

3. **release job:** After creating the GitHub release, update `threedoors-a.rb` in the tap:
   - Compute SHA256 for each binary
   - Generate formula with correct URLs, versions, and checksums
   - Push to `arcavenae/homebrew-tap`

🧪 **Murat (Test Architect):** The formula update step needs to be atomic and idempotent. If CI fails mid-update, the tap should not be left in a broken state. Recommendation: generate the full formula file, then push as a single commit. If the push fails, the old formula remains valid.

Also: add a CI validation step that runs `ruby -c Formula/threedoors-a.rb` after generation to catch syntax errors before pushing.

### Conflict Prevention

🏗️ **Winston (Architect):** With different binary names (`threedoors` vs `threedoors-a`), there are zero conflicts:

| Scenario | Result |
|----------|--------|
| Install only stable | `/usr/local/bin/threedoors` |
| Install only alpha | `/usr/local/bin/threedoors-a` |
| Install both | Both binaries coexist, no conflict |
| Upgrade stable | Only `threedoors` updated |
| Upgrade alpha | Only `threedoors-a` updated |
| Uninstall one | Other remains unaffected |

No `keg_only`, no `conflicts_with`, no `brew link/unlink` gymnastics. This is the cleanest possible UX.

📋 **John (PM):** This matches how other projects handle it. `firefox` and `firefox-nightly` are separate formulae with separate binaries. `go` and `go@1.21` use `keg_only` because they share a binary name -- we avoid that complexity entirely.

### The `--version` Output Question

🧪 **Quinn (QA):** Both binaries share the same `cmd/threedoors` source. The `--version` flag outputs the version string injected via ldflags. For `threedoors-a`, the version will be `0.1.0-alpha.20260308.abc1234` -- the alpha marker is embedded in the version, not the binary name. This is correct behavior. Users can identify which channel they're running via `threedoors-a --version`.

Should the `--version` output also include the binary name? E.g., `ThreeDoors Alpha 0.1.0-alpha.20260308.abc1234`?

📋 **John (PM):** Yes, good call. The version output should indicate the channel. Inject an additional ldflag like `-X main.channel=alpha` for the alpha build. The `--version` command can then display `ThreeDoors (alpha) v0.1.0-alpha.20260308.abc1234`. Low effort, high clarity.

### Alternative Considered: Same Binary Name, Different Versions

🏗️ **Winston (Architect):** We considered keeping the binary name as `threedoors` for both formulae and using `conflicts_with` to prevent simultaneous installation. Rejected because:

1. Users who want to test alpha while keeping stable as fallback can't.
2. `brew switch` / `brew link --overwrite` is poor UX.
3. Scripts that reference `threedoors` would break when switching channels.

The different binary name approach (`threedoors-a`) is strictly superior for our use case.

### Alternative Considered: `threedoors@alpha` Naming

📋 **John (PM):** Homebrew's `@` convention (e.g., `node@20`) is for versioned formulae of the same tool. `threedoors@alpha` would imply it's an older/specific version, not a parallel channel. It also triggers `keg_only :versioned_formula` expectations. The `-a` suffix is cleaner for a channel distinction and doesn't carry versioned-formula baggage.

💻 **Amelia (Dev):** Also, `@` in formula names has special handling in Homebrew's Ruby code. Avoiding it for a custom tap keeps things simpler.

---

## Decisions Summary

| ID | Decision | Rationale |
|----|----------|-----------|
| DD-1 | Alpha binary named `threedoors-a` (not `threedoors`) | Prevents Homebrew conflicts; allows simultaneous installation; clear channel identity |
| DD-2 | Alpha formula named `threedoors-a.rb` in same tap | Single tap (`arcavenae/homebrew-tap`); consistent install UX |
| DD-3 | Alpha version scheme: `0.1.0-alpha.YYYYMMDD.SHA7` | Already implemented; monotonic; traceable to commit |
| DD-4 | CI updates tap formula on every push to main | Extends existing release job; uses HOMEBREW_TAP_TOKEN |
| DD-5 | No `conflicts_with` or `keg_only` needed | Different binary names = zero conflicts |
| DD-6 | `--version` output includes channel identifier | `ThreeDoors (alpha) v0.1.0-alpha...` for clarity |
| DD-7 | Reject `threedoors@alpha` naming | `@` convention is for versioned formulae, not channels |

## Rejected Options

### Option A: Same Binary Name with `conflicts_with`
**Rejected because:** Prevents simultaneous installation; poor UX for users wanting both channels; `brew link/unlink` friction.

### Option B: `threedoors@alpha` Formula Naming
**Rejected because:** `@` is Homebrew's versioned-formula convention; implies specific version, not rolling channel; triggers `keg_only :versioned_formula` expectations.

### Option C: Separate Tap for Alpha (`arcavenae/homebrew-tap-alpha`)
**Rejected because:** Unnecessary complexity; a single tap with two formulae is standard practice; extra `brew tap` command for users.

### Option D: Alpha via `brew install --HEAD`
**Rejected because:** `--HEAD` builds from source on user's machine; slow; requires Go toolchain installed; no prebuilt binary distribution; can't auto-update.
