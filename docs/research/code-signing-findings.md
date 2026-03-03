# Code Signing Research Findings

**Date:** 2026-03-02
**Status:** All CI infrastructure is in place; secrets need to be configured to enable signing.

## Executive Summary

PR #30 (Story 5.1 - macOS Distribution & Packaging) added a comprehensive code signing and notarization pipeline to CI. The pipeline is **correctly designed to be fork-safe** — it skips signing when secrets are unavailable. Current alpha releases are unsigned because the required Apple Developer credentials have not been configured as GitHub Actions secrets/variables.

## Current State

### What Exists (from PR #30, merged 2026-03-02)

**CI Pipeline** (`.github/workflows/ci.yml`) has a 5-job structure:
1. `quality-gate` — lint, vet, test
2. `build-binaries` — cross-compile darwin-arm64, darwin-amd64, linux-amd64
3. `sign-and-notarize` — **gated by `vars.SIGNING_ENABLED == 'true'`**
4. `release` — creates GitHub Release, uses signed binaries if available, unsigned fallback
5. `update-homebrew` — updates Homebrew tap (also gated on `SIGNING_ENABLED`)

**Go packages** (`internal/dist/`):
- `CodeSigner` — wraps `codesign` with hardened runtime flags
- `Notarizer` — wraps `xcrun notarytool`
- `PkgBuilder` — wraps `pkgbuild`
- All use a `CommandRunner` interface for testability

**Scripts and Makefile**:
- `scripts/create-pkg.sh` — parameterized pkg installer creation
- `Makefile` targets: `sign`, `pkg`, `release-local`, `test-dist`

### Why Builds Are Unsigned

The `sign-and-notarize` job has this condition:
```yaml
if: github.event_name == 'push' && vars.SIGNING_ENABLED == 'true'
```

The repository variable `SIGNING_ENABLED` is **not set**, so the job is skipped on every push to main. The release job falls back to unsigned binaries.

Every alpha release confirms this:
> **Signed:** No (unsigned)

## What Is Missing

### 1. Repository Variable

| Variable | Value | Where |
|----------|-------|-------|
| `SIGNING_ENABLED` | `true` | GitHub repo Settings → Variables → Actions |

### 2. GitHub Actions Secrets

The CI workflow references these secrets:

| Secret | Description | How to Obtain |
|--------|-------------|---------------|
| `APPLE_CERTIFICATE_P12` | Base64-encoded Developer ID Application certificate (.p12) | Export from Keychain Access after enrollment |
| `APPLE_CERTIFICATE_PASSWORD` | Password for the .p12 file | Set during export |
| `APPLE_INSTALLER_CERTIFICATE_P12` | Base64-encoded Developer ID Installer certificate (.p12) | Export from Keychain Access |
| `APPLE_INSTALLER_CERTIFICATE_PASSWORD` | Password for the installer .p12 | Set during export |
| `APPLE_SIGNING_IDENTITY` | Certificate CN, e.g. `Developer ID Application: Your Name (TEAMID)` | `security find-identity -v -p codesigning` |
| `APPLE_INSTALLER_IDENTITY` | Installer cert CN, e.g. `Developer ID Installer: Your Name (TEAMID)` | `security find-identity -v` |
| `APPLE_ID` | Apple ID email used for notarization | Your Apple Developer account email |
| `APPLE_ID_PASSWORD` | App-specific password for notarization | Generate at appleid.apple.com → App-Specific Passwords |
| `APPLE_TEAM_ID` | 10-character Apple Developer Team ID | Apple Developer portal → Membership |
| `HOMEBREW_TAP_TOKEN` | GitHub PAT with repo scope for `arcaven/homebrew-tap` | GitHub Settings → Personal Access Tokens |

### 3. Apple Developer Program Enrollment

All of the above requires an active **Apple Developer Program** membership ($99/year). Without this, Developer ID certificates cannot be obtained.

## Signing Approaches

### Option A: Developer ID Signing + Notarization (Recommended)

This is what PR #30 implemented. It provides:
- Gatekeeper passes without user override
- Professional-grade trust chain
- Required for: Homebrew distribution, direct download without warnings

**Requirements:** Apple Developer Program ($99/year), all secrets above.

### Option B: Ad-hoc Signing (Minimal / Stopgap)

Ad-hoc signing uses `codesign --sign -` (dash identity) which:
- Does NOT require Apple Developer Program
- Does NOT pass Gatekeeper (users still get "unidentified developer" warning)
- Provides basic code integrity verification
- Does NOT support notarization

**Verdict:** Ad-hoc signing provides no user-facing benefit over unsigned binaries for distributed Go CLI tools. Not recommended as it adds CI complexity without solving the actual problem.

### Option C: Self-signed Certificate (Middle Ground)

Create a local certificate authority and self-signed code signing cert:
- Free, no Apple Developer enrollment needed
- Does NOT pass Gatekeeper
- Can be installed by users who trust the certificate
- Marginally better than ad-hoc for internal teams

**Verdict:** Only useful if distributing within a team that can install a trust profile. Not practical for open-source.

## Step-by-Step Plan to Enable Signing

### Prerequisites (Developer Must Provide)

1. **Enroll in Apple Developer Program** at developer.apple.com ($99/year)
2. **Create Developer ID certificates** in Xcode or Apple Developer portal:
   - Developer ID Application (for binary signing)
   - Developer ID Installer (for pkg signing)
3. **Generate an app-specific password** at appleid.apple.com for notarization

### Configuration Steps

1. **Export certificates as .p12 files:**
   ```bash
   # Open Keychain Access → My Certificates
   # Right-click "Developer ID Application: ..." → Export
   # Save as developer-id-app.p12 with a strong password
   # Repeat for "Developer ID Installer: ..."
   ```

2. **Base64-encode the .p12 files:**
   ```bash
   base64 -i developer-id-app.p12 | pbcopy
   # Paste as APPLE_CERTIFICATE_P12 secret
   base64 -i developer-id-installer.p12 | pbcopy
   # Paste as APPLE_INSTALLER_CERTIFICATE_P12 secret
   ```

3. **Find your signing identities:**
   ```bash
   security find-identity -v -p codesigning
   # Copy the full CN string for APPLE_SIGNING_IDENTITY
   security find-identity -v
   # Find the installer identity for APPLE_INSTALLER_IDENTITY
   ```

4. **Set GitHub Actions secrets** (Settings → Secrets and variables → Actions → Secrets):
   - `APPLE_CERTIFICATE_P12`
   - `APPLE_CERTIFICATE_PASSWORD`
   - `APPLE_INSTALLER_CERTIFICATE_P12`
   - `APPLE_INSTALLER_CERTIFICATE_PASSWORD`
   - `APPLE_SIGNING_IDENTITY`
   - `APPLE_INSTALLER_IDENTITY`
   - `APPLE_ID`
   - `APPLE_ID_PASSWORD`
   - `APPLE_TEAM_ID`

5. **Set GitHub Actions variable** (Settings → Secrets and variables → Actions → Variables):
   - `SIGNING_ENABLED` = `true`

6. **Create Homebrew tap repository** (optional, for `update-homebrew` job):
   - Create `arcaven/homebrew-tap` repository on GitHub
   - Generate a PAT with `repo` scope
   - Set as `HOMEBREW_TAP_TOKEN` secret

7. **Verify:** Push or merge to main and confirm the release shows "Signed: Yes (Apple Developer ID)"

### Local Signing (for testing)

```bash
# Set environment variables
export APPLE_SIGNING_IDENTITY="Developer ID Application: Your Name (TEAMID)"
export APPLE_INSTALLER_IDENTITY="Developer ID Installer: Your Name (TEAMID)"

# Build, sign, and create pkg
make release-local VERSION=0.1.0-test

# Verify
codesign --verify --deep --strict bin/threedoors
```

## Files Reference

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | CI pipeline with signing job (lines 88-200) |
| `internal/dist/codesign.go` | Go wrapper for codesign |
| `internal/dist/notarize.go` | Go wrapper for notarytool |
| `internal/dist/pkg_builder.go` | Go wrapper for pkgbuild |
| `scripts/create-pkg.sh` | Shell script for pkg creation |
| `Makefile` | Local `sign`, `pkg`, `release-local` targets |
| `Formula/threedoors.rb` | Homebrew formula template |

## Conclusion

No CI changes are needed. PR #30 built the complete signing pipeline correctly. The only action required is:

1. Enroll in Apple Developer Program
2. Configure 9 GitHub Actions secrets + 1 variable
3. Optionally create the Homebrew tap repository

Once `SIGNING_ENABLED=true` and secrets are configured, the next push to main will produce signed, notarized binaries and pkg installers automatically.
