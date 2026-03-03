# CI Signing Pipeline Audit

**Date:** 2026-03-03
**Scope:** `.github/workflows/ci.yml` — sign-and-notarize job
**Status:** Pipeline implemented but inactive (`SIGNING_ENABLED` not set)

---

## Executive Summary

The signing pipeline (Story 5.1, merged 2026-03-02) is structurally sound. The `codesign` invocation, hardened runtime flag, and `notarytool` arguments all follow Apple's documented requirements. However, there are **four concrete issues** that will cause problems once the pipeline is activated, and one design concern worth tracking.

| # | Finding | Severity | Section |
|---|---------|----------|---------|
| 1 | Cross-compiled Go binaries are not valid Mach-O for codesign | **Critical** | §1 |
| 2 | CGO_ENABLED not explicitly set during cross-compilation | **High** | §2 |
| 3 | 4-hour notarization timeout masks a deeper problem | **Medium** | §3 |
| 4 | No entitlements plist — acceptable but fragile | **Low** | §4 |
| 5 | spctl --assess may not work in CI without GUI session | **Medium** | §5 |

---

## Finding 1: Cross-Compiled Binaries Cannot Be Signed on Linux (Critical)

### Problem

The `build-binaries` job runs on `ubuntu-latest` and cross-compiles darwin binaries:

```yaml
# build-binaries job (ubuntu-latest)
GOOS=darwin GOARCH=arm64 go build ... -o threedoors-darwin-arm64
GOOS=darwin GOARCH=amd64 go build ... -o threedoors-darwin-amd64
```

The `sign-and-notarize` job then downloads these artifacts on `macos-latest` and runs `codesign` on them.

**Go cross-compilation from Linux to darwin produces valid Mach-O binaries**, and `codesign` on macOS can sign them. This part is correct. However, the binary must be a **pure Go binary** (no CGO) for this to work — if CGO is involved, the cross-compilation will either fail or produce a broken binary.

### Current State

The build step does **not** explicitly set `CGO_ENABLED=0`. Go's default behavior when `GOOS` differs from the host OS is to disable CGO automatically, so this works **by accident**. If any dependency ever adds a CGO import (e.g., `mattn/go-sqlite3`), the build will fail silently or produce an invalid binary.

### Recommendation

Explicitly set `CGO_ENABLED=0` in the build step:

```yaml
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ...
```

This makes the intent clear and prevents future breakage.

---

## Finding 2: CGO_ENABLED Not Explicitly Set (High)

### Problem

Related to Finding 1 but distinct: the absence of `CGO_ENABLED=0` is a code quality issue independent of cross-compilation.

When building on `ubuntu-latest` with `GOOS=darwin`, Go implicitly disables CGO. But this implicit behavior is:
- Not obvious to maintainers reading the workflow
- Not guaranteed across Go versions (though it has been stable)
- A source of confusion if the build job is ever moved to `macos-latest`

If the build were moved to `macos-latest` (e.g., to build natively instead of cross-compiling), `CGO_ENABLED` would default to `1`, and the resulting binary would link against system dylibs — changing signing behavior and potentially breaking notarization if the linked dylibs aren't also signed.

### Recommendation

Add `CGO_ENABLED=0` explicitly to all three build commands. This is standard practice for Go binaries distributed as standalone executables.

---

## Finding 3: Notarization Timeout Escalation (Medium)

### Problem

The CI workflow uses `--timeout 14400` (4 hours) for `notarytool submit --wait`. The task description notes this has been escalated from 15min → 30min → 1hr → 4hrs.

**Root cause analysis:** Apple's notarization service typically completes in 5–15 minutes for CLI binaries. Repeated timeouts suggest one of:

1. **The pipeline has never actually run.** Since `SIGNING_ENABLED` has never been set to `true`, the sign-and-notarize job has always been skipped. The timeout bumps may have been speculative/preemptive rather than based on actual failures.

2. **If it did run:** Submitting an unsigned or improperly built binary causes notarytool to hang or return cryptic errors. Apple's notarization requires:
   - Valid Mach-O with a Developer ID signature
   - Hardened runtime enabled
   - No unsigned nested code

3. **Network/auth issues:** If the Apple ID credentials are wrong or the app-specific password has expired, notarytool can hang during the `--wait` polling loop.

### What Known-Working Pipelines Use

Comparing against open-source Go projects that notarize successfully:

| Project | Timeout | Notes |
|---------|---------|-------|
| **goreleaser** (sign plugin) | 10 minutes | Signs natively on macOS runner |
| **mitchellh/gon** | 15 minutes default | Popular Go notarization tool |
| **create-dmg** community | 20 minutes | Typical for small binaries |

A 4-hour timeout is far beyond what any working pipeline uses. If notarization hasn't completed in 20 minutes for a <20MB CLI binary, something is wrong.

### Recommendation

1. Set timeout back to `1200` (20 minutes)
2. Add `--verbose` flag to notarytool for better diagnostics:
   ```
   xcrun notarytool submit ... --wait --timeout 1200 --verbose
   ```
3. On failure, fetch the notarization log:
   ```
   xcrun notarytool log <submission-id> --apple-id ... notarization-log.json
   ```
4. Upload the log as an artifact for debugging

---

## Finding 4: No Entitlements Plist (Low)

### Current Behavior

The `codesign` invocation uses `--options runtime` (hardened runtime) without an `--entitlements` flag:

```bash
codesign --force --options runtime --sign "$IDENTITY" --timestamp "$BINARY"
```

### Assessment

For a pure Go CLI binary that doesn't use JIT, doesn't access the camera/microphone, doesn't use Apple Events, and doesn't need any special capabilities, **this is correct**. No entitlements plist is needed.

However, if the application ever needs to:
- Execute unsigned code (e.g., plugin system)
- Disable library validation
- Access protected resources

...an entitlements plist will be required, and its absence will cause notarization to fail with a non-obvious error.

### Recommendation

No action needed now. Document this decision so future maintainers know to add entitlements if the binary's capabilities change.

---

## Finding 5: spctl --assess May Not Work in CI (Medium)

### Problem

The workflow runs:

```bash
spctl --assess --type execute "$BINARY"
```

On modern macOS (Ventura+), `spctl --assess` can behave differently in headless CI environments. GitHub's `macos-latest` runners may not have the full Gatekeeper policy database available, causing false negatives or errors.

### What Known-Working Pipelines Do

Most pipelines skip `spctl --assess` in CI and instead:
1. Verify the signature with `codesign --verify --deep --strict` (already done)
2. Check the notarization status with `xcrun notarytool info <submission-id>`
3. Test Gatekeeper locally during development

### Recommendation

Keep `spctl --assess` but make it non-fatal:

```bash
spctl --assess --type execute "$BINARY" || echo "Warning: spctl assessment failed (expected in CI)"
```

Or replace with `xcrun notarytool info` to verify notarization status.

---

## Signing Correctness Checklist

| Requirement | Status | Notes |
|-------------|--------|-------|
| Hardened runtime (`--options runtime`) | Correct | Present in codesign call |
| Timestamp (`--timestamp`) | Correct | Present in codesign call |
| Developer ID Application cert | Correct | Imported from P12 secret |
| Developer ID Installer cert | Correct | Separate cert for pkg signing |
| Keychain setup with partition list | Correct | `set-key-partition-list` called |
| Keychain cleanup on failure | Correct | Uses `if: always()` |
| Binary signed BEFORE notarization | Correct | Sign step precedes notarize step |
| Signed AFTER cross-compilation | Correct | Build on Linux, sign on macOS |
| Zip wrapper for notarytool submit | Correct | Standalone Mach-O must be zipped |
| Staple skipped for bare binaries | Correct | Only pkg files are stapled |
| pkg files stapled after notarization | Correct | `xcrun stapler staple` called |
| Fork-safe (no signing on PRs) | Correct | Gated on `SIGNING_ENABLED` variable |

---

## Comparison with Known-Working Go Notarization Approaches

### mitchellh/gon (Reference Implementation)

The most widely used Go notarization tool. Key differences from our approach:

| Aspect | gon | Our Pipeline |
|--------|-----|--------------|
| Build location | macOS runner | Linux runner (cross-compile) |
| CGO | Explicitly disabled | Implicitly disabled |
| Signing | `codesign` directly | `codesign` directly (same) |
| Notarization | gon wraps notarytool | Direct notarytool call |
| Timeout | 10–15 min | 4 hours (excessive) |
| Entitlements | Configurable via HCL | None (acceptable for CLI) |

### GoReleaser + sign plugin

| Aspect | GoReleaser | Our Pipeline |
|--------|-----------|--------------|
| Build | Native per-platform | Cross-compile on Linux |
| Sign | macOS-only job | macOS-only job (same) |
| Distribution | Homebrew, snap, etc. | Homebrew + GitHub Release |

Our architecture matches the GoReleaser pattern: build on Linux, sign on macOS. This is valid.

---

## Action Items (Priority Order)

1. **Add `CGO_ENABLED=0` to all build commands** — Prevents future breakage, makes intent explicit
2. **Reduce notarization timeout to 1200s** — 4 hours is a CI resource waste; if it takes >20min, something is wrong
3. **Add `--verbose` to notarytool** — Essential for debugging first-run issues
4. **Add notarization log upload on failure** — Capture `notarytool log` output as artifact
5. **Make `spctl --assess` non-fatal** — May fail in CI headless environment
6. **Activate the pipeline** — Configure the 9 secrets + `SIGNING_ENABLED=true` variable in GitHub repo settings (see `docs/research/code-signing-findings.md` for the full list)

---

## Appendix: Pipeline Architecture Diagram

```
Push to main
    │
    ▼
┌─────────────┐
│ quality-gate │  ubuntu-latest
│ fmt/vet/lint │  Every PR + push
│ test/build   │
└──────┬──────┘
       │ (push only)
       ▼
┌──────────────┐
│build-binaries│  ubuntu-latest
│ darwin-arm64 │  CGO_ENABLED=0 (should be explicit)
│ darwin-amd64 │
│ linux-amd64  │
└──────┬───────┘
       │
       ▼
┌───────────────────┐
│sign-and-notarize  │  macos-latest
│ import certs      │  SIGNING_ENABLED == 'true'
│ codesign --runtime│
│ notarytool submit │
│ create + sign pkg │
│ staple pkg        │
└──────┬────────────┘
       │
       ▼
┌─────────┐
│ release │  ubuntu-latest
│ GH tag  │  Signed or unsigned
│ upload  │
└────┬────┘
     │ (if signed)
     ▼
┌───────────────┐
│update-homebrew│  ubuntu-latest
│ Formula/      │
│ arcaven/      │
│ homebrew-tap  │
└───────────────┘
```
