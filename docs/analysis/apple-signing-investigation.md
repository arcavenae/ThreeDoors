# Apple Code Signing & Notarization Investigation

**Date:** 2026-03-03
**Status:** Signing succeeds; notarization has NEVER completed successfully.

## Executive Summary

Apple code signing and notarization infrastructure is fully built and all 10 GitHub Actions secrets plus the `SIGNING_ENABLED=true` variable are configured. Code signing works perfectly. **Notarization fails on every single CI run** — Apple accepts the submission and begins processing, but the status remains "In Progress" indefinitely (2+ hours) until the GitHub Actions macOS runner loses network connectivity, crashing with `NSURLErrorDomain Code=-1009 "The Internet connection appears to be offline."`.

This is not a timeout issue (despite PRs #67, #76, #88 all increasing timeout). It is either an Apple-side processing hang or a credential/account issue that causes Apple to silently fail to process the submission.

## Evidence: CI Run Analysis

### Failure Pattern (100% consistent across all runs)

Every push to `main` triggers the `sign-and-notarize` job. The step-level results are always:

| Step | Result |
|------|--------|
| Import certificates | SUCCESS |
| Sign darwin binaries | SUCCESS |
| Verify signatures | SUCCESS |
| **Notarize binaries** | **FAILURE** |
| Verify Gatekeeper assessment | SKIPPED |
| Build pkg installers | SKIPPED |
| Upload signed binaries | SKIPPED |
| Cleanup keychain | SUCCESS |

Checked runs (all on `main`, all with identical pattern):
- `22617464268` (2026-03-03 09:50 UTC) — Notarize binaries: FAILURE
- `22615673784` (2026-03-03 08:59 UTC) — Notarize binaries: FAILURE
- `22615326402` (2026-03-03 08:49 UTC) — Notarize binaries: FAILURE
- `22614969460` (2026-03-03 08:38 UTC) — Notarize binaries: FAILURE
- `22614679944` (2026-03-03 08:29 UTC) — Notarize binaries: FAILURE
- `22613698919` (2026-03-03 07:58 UTC) — Notarize binaries: FAILURE
- `22613683473` (2026-03-03 07:57 UTC) — Notarize binaries: FAILURE

### Detailed Failure Sequence (from run 22615673784)

1. **Submission accepted** — Apple returns a submission ID:
   ```
   Submission ID received
     id: aeaa7107-4876-4943-9213-6747d99da505
   Successfully uploaded file
   Waiting for processing to complete. Wait timeout is set to 14400.0 second(s).
   ```

2. **Indefinite "In Progress"** — polls every ~20 seconds for 2+ hours:
   ```
   Current status: In Progress...
   Current status: In Progress....
   Current status: In Progress.....
   [... continues for ~2 hours ...]
   ```

3. **Network failure** — macOS runner loses connectivity after extended polling:
   ```
   Error: HTTPError(statusCode: nil, error: Error Domain=NSURLErrorDomain Code=-1009
   "The Internet connection appears to be offline."
   UserInfo={NSErrorFailingURLStringKey=https://appstoreconnect.apple.com/notary/v2/asp?...})
   ```

4. Job exits with code 1.

### Timeline

- Job starts at ~13:14 UTC
- Notarization submission at ~13:14 UTC
- Continuous "In Progress" polling from 13:14 to 15:10 (~1h 56m)
- Network error at 15:10 UTC
- Job cleanup at 14:38 UTC (latest run) / 15:10 UTC (earlier run)

## PR History

| PR | Title | What Changed | Result |
|----|-------|--------------|--------|
| #30 | Story 5.1 - macOS Distribution & Packaging | Built entire signing pipeline | Infrastructure complete, not yet enabled |
| #46 | Code signing research findings | Documented that secrets were missing | Correctly identified missing config |
| #61 | Align CI secret names | Fixed 3 secret name mismatches (`APPLE_ID` → `APPLE_NOTARIZATION_APPLE_ID`, etc.) | Names now match configured secrets |
| #67 | Bump notarization timeout to 30 minutes | Timeout: 900s → 1800s | Did not help — problem is not timeout |
| #76 | Increase notarization timeout to 1 hour | Timeout: 1800s → 3600s | Did not help |
| #88 | Increase notarization timeout to 4 hours | Timeout: 3600s → 14400s | Made it worse — now hangs longer before network dies |

## Secrets Configuration Audit

All required secrets and variables are configured in GitHub Actions:

| Secret/Variable | Status | Purpose |
|----------------|--------|---------|
| `SIGNING_ENABLED` (variable) | **Set to `true`** | Gates the sign-and-notarize job |
| `APPLE_CERTIFICATE_P12` | Configured | Developer ID Application cert (base64) |
| `APPLE_CERTIFICATE_PASSWORD` | Configured | P12 password |
| `APPLE_INSTALLER_CERTIFICATE_P12` | Configured | Developer ID Installer cert (base64) |
| `APPLE_INSTALLER_CERTIFICATE_PASSWORD` | Configured | Installer P12 password |
| `APPLE_SIGNING_IDENTITY` | Configured | Signing identity CN string |
| `APPLE_INSTALLER_IDENTITY` | Configured | Installer identity CN string |
| `APPLE_NOTARIZATION_APPLE_ID` | Configured | Apple ID email for notarytool |
| `APPLE_NOTARIZATION_PASSWORD` | Configured | App-specific password |
| `APPLE_NOTARIZATION_TEAM_ID` | Configured | 10-char team ID |
| `HOMEBREW_TAP_TOKEN` | Configured | GitHub PAT for homebrew-tap repo |

**Note:** We cannot inspect secret values from the API — we can only confirm they exist. The secrets could contain incorrect values.

## Root Cause Analysis

### What Works

1. **Certificate import** — Both application and installer certs import successfully into the CI keychain ("1 identity imported" x2).
2. **Code signing** — `codesign --force --options runtime --sign` succeeds for both arm64 and amd64 binaries with hardened runtime and timestamp.
3. **Signature verification** — `codesign --verify --deep --strict` passes.

### What Fails

4. **Notarization** — `xcrun notarytool submit --wait` accepts the upload, gets a submission ID, but processing never completes.

### Possible Root Causes (ranked by likelihood)

#### 1. MOST LIKELY: Invalid or Mismatched Notarization Credentials

The notarization submission is accepted (Apple returns a submission ID), but processing never completes. This pattern is consistent with:

- **Wrong app-specific password** — Apple may accept the initial API auth but the notarization service internally fails to process. The `--wait` flag just polls the status endpoint which returns "In Progress" indefinitely because the submission is in a failed/limbo state.
- **Team ID mismatch** — The `APPLE_NOTARIZATION_TEAM_ID` may not match the team that owns the Developer ID certificate used for signing. Apple accepts the submission but can't validate the signing chain.
- **Apple ID not associated with the team** — The `APPLE_NOTARIZATION_APPLE_ID` account may not be a member of the team specified by `APPLE_NOTARIZATION_TEAM_ID`.

**Diagnostic:** Run `xcrun notarytool log <submission-id>` to retrieve the notarization log from Apple. This will show whether Apple actually rejected the submission or is genuinely still processing.

#### 2. LIKELY: New Apple Developer Account Throttling

Apple's documentation notes that first-time submissions from new Developer Program enrollments may take longer. However:
- Normal "longer" is 15-30 minutes, not 2+ hours
- The submissions have been running for days with no success
- This alone doesn't explain indefinite "In Progress"

#### 3. POSSIBLE: Binary Requires Entitlements

Go CLI binaries with hardened runtime (`--options runtime`) may need an entitlements file to function with notarization. Without entitlements, Apple may:
- Accept the submission
- Begin scanning
- Hit an issue with the binary structure and stall

Typical entitlements needed for Go binaries:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
</dict>
</plist>
```

Go's runtime uses memory mapping and JIT-like techniques that may require `com.apple.security.cs.allow-unsigned-executable-memory`. The `--options runtime` flag enables hardened runtime which restricts these capabilities unless explicitly entitled.

**Note:** Many Go CLI tools (e.g., `gh`, `terraform`, `hugo`) are successfully notarized without entitlements, so this may not be the issue. But it's worth testing.

#### 4. UNLIKELY: GitHub Actions macOS Runner Network Issues

The network error is a symptom, not the cause. The runner loses connectivity after ~2 hours of continuous polling, but the real question is why Apple never finishes processing. Healthy notarization completes in 1-15 minutes.

#### 5. UNLIKELY: Binary Format Issues

Go cross-compiled binaries (built on Linux with `GOOS=darwin`) produce valid Mach-O binaries, but there could be subtle format issues. However, `codesign --verify` passes, which strongly suggests the binary format is fine.

## CI Workflow Analysis

### Current Notarization Step (ci.yml)

```yaml
- name: Notarize binaries
  env:
    APPLE_NOTARIZATION_APPLE_ID: ${{ secrets.APPLE_NOTARIZATION_APPLE_ID }}
    APPLE_NOTARIZATION_PASSWORD: ${{ secrets.APPLE_NOTARIZATION_PASSWORD }}
    APPLE_NOTARIZATION_TEAM_ID: ${{ secrets.APPLE_NOTARIZATION_TEAM_ID }}
  run: |
    for BINARY in threedoors-darwin-arm64 threedoors-darwin-amd64; do
      echo "Notarizing $BINARY..."
      zip "${BINARY}.zip" "$BINARY"
      xcrun notarytool submit "${BINARY}.zip" \
        --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
        --password "$APPLE_NOTARIZATION_PASSWORD" \
        --team-id "$APPLE_NOTARIZATION_TEAM_ID" \
        --wait --timeout 14400
      rm "${BINARY}.zip"
    done
```

### Issues Identified

1. **No log retrieval on failure** — When notarization fails, the CI does not call `xcrun notarytool log <id>` to fetch Apple's detailed processing log. This is the single most important missing diagnostic.

2. **4-hour timeout is counterproductive** — Normal notarization takes 1-15 minutes. A 4-hour timeout means:
   - macOS runner costs ($0.08/min = ~$19.20 per 4-hour run)
   - Runner network dies before timeout is reached anyway
   - No useful information gained from waiting longer

3. **Sequential notarization** — Both binaries are notarized sequentially. If the first one hangs, the second never gets attempted.

4. **No `--output-format json`** — The notarytool output is not parsed for the submission ID, making it impossible to fetch logs programmatically after failure.

5. **Missing entitlements** — No `.entitlements` file is passed to `codesign`. While not always required for CLI tools, it should be tested.

## Apple Developer Program Requirements

For successful notarization, the following must all be true:

1. **Active Apple Developer Program membership** ($99/year) — certificates and notarization require this
2. **Developer ID Application certificate** — must be created in the Apple Developer portal, NOT an iOS distribution cert
3. **The Apple ID used for notarytool** must be a member of the team that owns the Developer ID cert
4. **App-specific password** must be generated for the specific Apple ID, not a regular password
5. **The binary must be signed** with the Developer ID cert (not ad-hoc, not self-signed)
6. **Hardened runtime must be enabled** (`--options runtime`) — already done
7. **Timestamp must be included** (`--timestamp`) — already done
8. **Binary must be zipped** for submission — already done

### Go-Specific Considerations

- Go binaries are statically linked and don't use Apple frameworks, which simplifies notarization
- No Info.plist is needed for CLI tools (only for .app bundles)
- No entitlements are strictly required for most Go CLI tools, but may be needed if the Go runtime's memory management triggers hardened runtime restrictions
- Cross-compilation (`GOOS=darwin GOARCH=arm64` on Linux) produces valid Mach-O binaries that Apple accepts for notarization

## Recommended Next Steps

### Immediate (no CI changes needed)

1. **Retrieve the notarization log** — This is the #1 priority. From a Mac with the Apple Developer credentials:
   ```bash
   xcrun notarytool log aeaa7107-4876-4943-9213-6747d99da505 \
     --apple-id "$APPLE_ID" \
     --password "$APP_SPECIFIC_PASSWORD" \
     --team-id "$TEAM_ID"
   ```
   This will show Apple's detailed assessment including any rejection reasons.

2. **Check submission history** — List all past submissions to see if any were actually rejected:
   ```bash
   xcrun notarytool history \
     --apple-id "$APPLE_ID" \
     --password "$APP_SPECIFIC_PASSWORD" \
     --team-id "$TEAM_ID"
   ```

3. **Verify credentials locally** — Test that the notarization credentials work:
   ```bash
   xcrun notarytool store-credentials "threedoors" \
     --apple-id "$APPLE_ID" \
     --password "$APP_SPECIFIC_PASSWORD" \
     --team-id "$TEAM_ID"
   ```

### CI Improvements (for future PR, once root cause is found)

1. **Add log retrieval on failure** — Capture the notarytool submission ID and fetch logs on failure
2. **Reduce timeout to 15 minutes** — If notarization doesn't complete in 15 minutes, something is wrong; fetch logs and fail fast
3. **Add `--output-format json`** — Parse the submission ID programmatically
4. **Test with entitlements** — Create a minimal `.entitlements` file and pass it to codesign with `--entitlements`
5. **Add retry logic** — Submit once, if "In Progress" stalls for >10 minutes, fetch log and report

### Account Verification Checklist

- [ ] Confirm Apple Developer Program membership is active (not expired)
- [ ] Confirm Developer ID Application certificate is valid and not revoked
- [ ] Confirm the Apple ID used for notarization is a member of the correct team
- [ ] Confirm the app-specific password is still valid (they can be revoked)
- [ ] Confirm the Team ID matches the team that owns the Developer ID cert
- [ ] Run `xcrun notarytool history` to see all past submission statuses
- [ ] Run `xcrun notarytool log <submission-id>` to get Apple's detailed assessment

## Conclusion

The signing infrastructure is correctly built and code signing works. The notarization failure is almost certainly a credential or account configuration issue — not a timeout issue. The previous approach of increasing timeout (PRs #67, #76, #88) was treating the symptom rather than the disease.

**The single most important action is retrieving the notarization log from Apple** (`xcrun notarytool log <submission-id>`), which will definitively show why Apple is not completing the notarization.
