# Dual Homebrew Distribution Research — Stable + Alpha Channels

**Date:** 2026-03-08
**Status:** Complete

---

## Executive Summary

ThreeDoors needs two parallel Homebrew distribution channels: a **stable** channel (`threedoors`) updated on semantic version tags via GoReleaser, and an **alpha** channel (`threedoors-a`) updated on every push to main. The recommended approach uses different binary names to avoid all Homebrew conflicts, with both formulae hosted in the existing `arcavenae/homebrew-tap` repository.

---

## 1. Problem Statement

PR #262 introduced GoReleaser for tagged `v*` releases, producing stable builds distributed via `brew install arcavenae/tap/threedoors`. The existing CI also builds alpha releases on every push to main (`0.1.0-alpha.YYYYMMDD.SHA7`). These alpha builds are available as GitHub releases but not via Homebrew.

Users should be able to:
- Install the stable release: `brew install arcavenae/tap/threedoors`
- Install the alpha release: `brew install arcavenae/tap/threedoors-a`
- Install both simultaneously without conflicts
- Upgrade each independently via `brew upgrade`

---

## 2. Homebrew Multi-Formula Research

### Can Two Formulae Coexist in One Tap?

**Yes.** A Homebrew tap is just a Git repository with a `Formula/` directory. Multiple formulae can coexist. Examples from official taps:
- `go` and `go@1.21`, `go@1.22` in homebrew-core
- `node` and `node@18`, `node@20`, `node@22` in homebrew-core
- `python@3.12` and `python@3.13` in homebrew-core

Custom taps have even fewer restrictions than homebrew-core.

### Binary Name Conflicts

When two formulae install a binary with the same name, Homebrew requires one of:
1. **`conflicts_with`** — prevents simultaneous installation entirely
2. **`keg_only`** — installs to the Cellar but doesn't symlink to `/usr/local/bin`
3. **Different binary names** — both can be installed and linked simultaneously

Option 3 (different binary names) is the cleanest solution. No special Homebrew directives needed. Both formulae install and upgrade independently.

### The `@` Convention

Homebrew's `@` naming (e.g., `node@20`) is specifically for **versioned formulae** — older or specific versions of the same tool. Using `threedoors@alpha` would be semantically incorrect because:
- `@` implies a pinned version, not a rolling channel
- Versioned formulae are typically `keg_only :versioned_formula`
- homebrew-core explicitly states "unstable versions (alpha, beta, development) are not acceptable for versioned formulae"

The `-a` suffix (or `-alpha`) is more appropriate for a parallel channel.

### How Other Projects Handle This

| Project | Stable | Nightly/Dev | Strategy |
|---------|--------|-------------|----------|
| Firefox | `firefox` | `firefox@nightly` (cask) | Separate cask, different app name |
| Go | `go` | `go@1.21` | Versioned, `keg_only` |
| Node | `node` | `node@18` | Versioned, `keg_only` |
| Rust | `rust` | `rustup` | Different tool entirely |
| GoReleaser | `goreleaser` | `goreleaser/tap/goreleaser-pro` | Different formula name |

The pattern that most closely matches our use case is GoReleaser's own approach: a separate formula with a distinct name in the same tap.

---

## 3. Apple Code Signing Constraints

### Does the Binary Name Need to Match the Formula Name?

**No.** Apple code signing uses a **bundle identifier** (e.g., `com.arcavenae.threedoors-a`), not the filename. The `codesign` command accepts an `--identifier` flag:

```bash
codesign --force --options runtime --sign "$IDENTITY" --identifier "com.arcavenae.threedoors-a" --timestamp threedoors-a-darwin-arm64
```

Notarization verifies the signature, hardened runtime, and team ID — not the binary name. Our existing CI already signs binaries with names like `threedoors-darwin-arm64` that don't match the installed binary name `threedoors`. No changes needed to the signing pipeline beyond adding the new binary names.

### Binary Naming Recommendations

| Binary | Identifier | Formula |
|--------|-----------|---------|
| `threedoors` | `com.arcavenae.threedoors` | `threedoors.rb` |
| `threedoors-a` | `com.arcavenae.threedoors-a` | `threedoors-a.rb` |

---

## 4. Recommended Design

### Tap Structure

```
arcavenae/homebrew-tap/
├── Formula/
│   ├── threedoors.rb          # Stable — GoReleaser auto-updates on v* tags
│   └── threedoors-a.rb        # Alpha — CI auto-updates on every main push
└── README.md
```

### Formula: `threedoors.rb` (Stable)

Already managed by GoReleaser. No changes needed. Updated automatically when a `v*` tag is pushed.

### Formula: `threedoors-a.rb` (Alpha)

```ruby
class ThreedoorsA < Formula
  desc "TUI task manager — alpha channel (updated on every main push)"
  homepage "https://github.com/arcavenae/ThreeDoors"
  version "0.1.0-alpha.20260308.abc1234"  # Updated by CI
  license "MIT"

  on_arm do
    url "https://github.com/arcavenae/ThreeDoors/releases/download/alpha-20260308-abc1234/threedoors-a-darwin-arm64"
    sha256 "PLACEHOLDER"
  end

  on_intel do
    url "https://github.com/arcavenae/ThreeDoors/releases/download/alpha-20260308-abc1234/threedoors-a-darwin-amd64"
    sha256 "PLACEHOLDER"
  end

  on_linux do
    url "https://github.com/arcavenae/ThreeDoors/releases/download/alpha-20260308-abc1234/threedoors-a-linux-amd64"
    sha256 "PLACEHOLDER"
  end

  def install
    if OS.mac? && Hardware::CPU.arm?
      bin.install "threedoors-a-darwin-arm64" => "threedoors-a"
    elsif OS.mac?
      bin.install "threedoors-a-darwin-amd64" => "threedoors-a"
    else
      bin.install "threedoors-a-linux-amd64" => "threedoors-a"
    end
  end

  test do
    assert_match "ThreeDoors", shell_output("#{bin}/threedoors-a --version 2>&1")
  end
end
```

### Version Scheme

| Channel | Version Format | Example | Source |
|---------|---------------|---------|--------|
| Stable | Semantic version from git tag | `1.0.0` | `v1.0.0` tag |
| Alpha | `0.1.0-alpha.YYYYMMDD.SHA7` | `0.1.0-alpha.20260308.abc1234` | Push to main |

The alpha version scheme is already implemented in CI. The date component ensures monotonic ordering; the SHA provides commit traceability.

### Binary Build Changes

The alpha CI build step needs to produce binaries named `threedoors-a-*` instead of (or in addition to) `threedoors-*`:

```bash
# In ci.yml build-binaries job
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o threedoors-a-darwin-arm64 ./cmd/threedoors
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o threedoors-a-darwin-amd64 ./cmd/threedoors
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o threedoors-a-linux-amd64 ./cmd/threedoors
```

Same source, same entrypoint (`./cmd/threedoors`), different output name. The ldflags already inject the alpha version string.

### Channel Identifier in `--version`

Add an ldflag for the channel:

```bash
LDFLAGS="-X main.version=$VERSION -X main.channel=alpha"
```

The `--version` output becomes:
```
ThreeDoors (alpha) v0.1.0-alpha.20260308.abc1234
```

For stable builds, `main.channel` would be empty, producing:
```
ThreeDoors v1.0.0
```

### CI Workflow: Tap Update Step

After the alpha release is created, a new CI step pushes the updated formula to the tap:

```yaml
- name: Update alpha Homebrew formula
  env:
    HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
  run: |
    VERSION="${{ steps.version.outputs.version }}"
    TAG="${{ steps.version.outputs.tag }}"

    # Compute SHA256 for each binary
    SHA_ARM64=$(shasum -a 256 threedoors-a-darwin-arm64 | cut -d' ' -f1)
    SHA_AMD64=$(shasum -a 256 threedoors-a-darwin-amd64 | cut -d' ' -f1)
    SHA_LINUX=$(shasum -a 256 threedoors-a-linux-amd64 | cut -d' ' -f1)

    # Generate formula
    cat > threedoors-a.rb <<FORMULA
    class ThreedoorsA < Formula
      desc "TUI task manager — alpha channel"
      homepage "https://github.com/arcavenae/ThreeDoors"
      version "$VERSION"
      license "MIT"
      # ... (full formula with correct URLs and checksums)
    end
    FORMULA

    # Validate syntax
    ruby -c threedoors-a.rb

    # Push to tap repo
    # (clone tap, copy formula, commit, push)
```

### Conflict Matrix

| Installation State | `/usr/local/bin/threedoors` | `/usr/local/bin/threedoors-a` | Conflicts? |
|---|---|---|---|
| Only stable installed | Present | Absent | No |
| Only alpha installed | Absent | Present | No |
| Both installed | Present | Present | **No** |
| Upgrade stable | Updated | Unchanged | No |
| Upgrade alpha | Unchanged | Updated | No |

Zero conflicts in any combination.

---

## 5. Implementation Checklist

### Phase 1: Alpha Formula Infrastructure

- [ ] Modify `ci.yml` `build-binaries` job to produce `threedoors-a-*` binaries
- [ ] Update `sign-and-notarize` job for `threedoors-a-*` with distinct identifier
- [ ] Add CI step to generate and push `threedoors-a.rb` to tap repo
- [ ] Add `main.channel` ldflag support to `cmd/threedoors`
- [ ] Add `ruby -c` validation before pushing formula
- [ ] Verify `brew install arcavenae/tap/threedoors-a` works end-to-end

### Phase 2: Polish

- [ ] Update README with dual install instructions
- [ ] Add `brew upgrade` verification to CI (smoke test)
- [ ] Consider keeping old `threedoors-*` alpha binaries in release for backward compat
- [ ] Document the dual-channel strategy in architecture docs

---

## 6. Open Questions

1. **Should old alpha releases be cleaned up?** GitHub releases accumulate fast (one per push to main). Consider a retention policy (e.g., keep last 30 alpha releases).
2. **Should the alpha formula include Linux support?** Current stable formula (via GoReleaser) includes Linux. Alpha should match.
3. **Should we keep producing unsigned `threedoors-*` binaries in alpha releases?** The existing alpha release already produces these. They could be kept for backward compatibility or removed to simplify.

---

## References

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Formulae Versions](https://docs.brew.sh/Versions)
- [Homebrew Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Homebrew Discussion: Two formulae with conflicting binary names](https://github.com/orgs/Homebrew/discussions/98)
- [Homebrew Discussion: Two formulae using the same command](https://github.com/orgs/Homebrew/discussions/4908)
- [GoReleaser Homebrew Taps](https://goreleaser.com/customization/homebrew/)
- [Apple Resolving Common Notarization Issues](https://developer.apple.com/documentation/security/resolving-common-notarization-issues)
- [Existing Homebrew Distribution Research](homebrew-distribution-research.md)
- [Existing Homebrew Party Mode](../../_bmad-output/planning-artifacts/homebrew-distribution-party-mode.md)
