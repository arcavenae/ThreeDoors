# Signing & Notarization Failure Timeline

**Date:** 2026-03-03
**Author:** zealous-rabbit (automated analysis)
**Status:** All 32 signing runs have failed. Zero successful notarizations exist.

---

## Executive Summary

Every Sign & Notarize CI run has failed. Code signing and signature verification succeed every time — the failure is exclusively in Apple's notarization processing. The `xcrun notarytool submit --wait` command uploads the binary successfully, but Apple's service never transitions the submission from "In Progress" to a terminal state ("Accepted" or "Invalid"). The timeout has been escalated from 15 minutes to 4 hours across three PRs with no improvement.

**Root cause hypothesis:** The Apple Developer account or notarization credentials have an issue preventing processing (e.g., pending agreement acceptance, account enrollment issue, or invalid app-specific password permissions). This is NOT a timeout calibration problem.

---

## Timeline

### Phase 1: Pipeline Creation (Signing Disabled)

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| PR #26 merged | 2026-03-02 13:54 | Added macOS distribution epic to PRD |
| PR #30 merged | 2026-03-02 17:45 | Created 5-job CI pipeline in `.github/workflows/ci.yml`. Added `internal/dist/` package with `CodeSigner`, `Notarizer`, `PkgBuilder`. Sign & Notarize gated by `vars.SIGNING_ENABLED == 'true'` |
| Runs 22607604680, 22608362550, 22609292125 | 2026-03-03 03:57–05:10 | Sign & Notarize **skipped** (SIGNING_ENABLED not set) |

### Phase 2: Signing Enabled — First Failures (15-min timeout)

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| PR #61 merged | 2026-03-03 05:40 | Fixed secret name mismatches: `APPLE_ID` → `APPLE_NOTARIZATION_APPLE_ID`, `APPLE_ID_PASSWORD` → `APPLE_NOTARIZATION_PASSWORD`, `APPLE_TEAM_ID` → `APPLE_NOTARIZATION_TEAM_ID` |
| **Run 22610007818** (first signing run) | 2026-03-03 05:40 | Certificate import: SUCCESS. Code signing: SUCCESS. Signature verification: SUCCESS. **Notarization upload: SUCCESS** (submission `3b099c4d-fcc2-4dc5-9565-4cbcb9a62c9f`). **Notarization wait: TIMEOUT** at 900s. Exit code 124. |
| Runs 22610342104–22610427593 | 2026-03-03 05:53–05:58 | All timeout at 900s with exit code 124 |

### Phase 3: Timeout Escalation — 30 Minutes

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| PR #67 merged | 2026-03-03 06:04 | Bumped timeout from 900s → 1800s. Rationale: "First submissions from a new Apple Developer account undergo extra validation." |
| **Run 22610622193** | 2026-03-03 06:04 | Submission `16ceff8b-778a-4b21-9e5a-f973c1e596bf`. Upload SUCCESS. **Timeout at 1800s.** Exit code 124. |
| Runs 22610948630–22611522206 | 2026-03-03 06:14–06:40 | All timeout at 1800s |

### Phase 4: Timeout Escalation — 1 Hour

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| PR #76 merged | 2026-03-03 06:52 | Bumped timeout from 1800s → 3600s. Rationale: "A full hour gives plenty of margin." |
| **Run 22611826698** | 2026-03-03 06:52 | Submission `8d62caa4-0b23-4fe4-8d35-da8235e0d0ca`. Upload SUCCESS. **Timeout at 3600s** (polled 60 minutes). Exit code 124. |
| Runs 22612010696–22612785683 | 2026-03-03 06:59–07:27 | All timeout at 3600s |

### Phase 5: Timeout Escalation — 4 Hours

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| PR #88 merged | 2026-03-03 07:53 | Bumped timeout from 3600s → 14400s. Rationale: "Giving a full 4-hour window to ensure completion." |
| **Run 22613556272** | 2026-03-03 07:53 | Submission `46ba72c3-b2a6-4ac0-a845-12e028ae55a2`. Upload SUCCESS. **Timeout at 14400s** (polled 4 hours). Exit code 124. |
| Runs 22613409974–22615673784 | 2026-03-03 07:48–08:59 | All timeout at 14400s |

### Phase 6: Network Failure (Most Recent Completed Run)

| Event | Date/Time (UTC) | Details |
|-------|-----------------|---------|
| **Run 22617464268** | 2026-03-03 09:50 | Submission `f1182072-88c8-492f-a2fb-c991f6a5edce`. Upload SUCCESS. Polled "In Progress" for ~53 minutes. **New failure mode: network error.** Exit code 1. |

**Exact error from Run 22617464268:**
```
Error: HTTPError(statusCode: nil, error: Error Domain=NSURLErrorDomain Code=-1009
"The Internet connection appears to be offline."
UserInfo={
  _kCFStreamErrorCodeKey=50,
  NSUnderlyingError=0x600002ac1170 {
    Error Domain=kCFErrorDomainCFNetwork Code=-1009 "(null)"
    UserInfo={
      _NSURLErrorNWResolutionReportKey=Resolved 0 endpoints in 1ms using unknown from cache,
      _NSURLErrorNWPathKey=unsatisfied (No network route)
    }
  },
  NSLocalizedDescription=The Internet connection appears to be offline.,
  NSErrorFailingURLStringKey=https://appstoreconnect.apple.com/notary/v2/submissions/f1182072-88c8-492f-a2fb-c991f6a5edce?
}
```

This suggests the GitHub Actions macOS runner lost network connectivity after ~53 minutes of polling, possibly due to runner recycling or network policy limits on long-running connections.

---

## What Succeeds vs. What Fails

### Always succeeds (every run)

1. **Certificate import** — "1 identity imported." for both application and installer certs
2. **Code signing** — `codesign --force --options runtime --sign` completes without error for both arm64 and amd64
3. **Signature verification** — `codesign --verify --deep --strict` passes for both architectures
4. **Notarization upload** — `xcrun notarytool submit` uploads successfully ("Successfully uploaded file")

### Always fails (every run)

5. **Notarization processing** — Apple's server accepts the binary but never transitions from "In Progress" to "Accepted" or "Invalid". The `--wait` flag polls until timeout.

---

## Notarization Submission IDs

| Run ID | Submission UUID | Timeout | Outcome |
|--------|----------------|---------|---------|
| 22610007818 | `3b099c4d-fcc2-4dc5-9565-4cbcb9a62c9f` | 900s | Timeout (exit 124) |
| 22610622193 | `16ceff8b-778a-4b21-9e5a-f973c1e596bf` | 1800s | Timeout (exit 124) |
| 22611826698 | `8d62caa4-0b23-4fe4-8d35-da8235e0d0ca` | 3600s | Timeout (exit 124) |
| 22613556272 | `46ba72c3-b2a6-4ac0-a845-12e028ae55a2` | 14400s | Timeout (exit 124) |
| 22617464268 | `f1182072-88c8-492f-a2fb-c991f6a5edce` | 14400s | Network error (exit 1) |

---

## Comparison: Expected vs. Actual Notarization Behavior

### What Apple documentation says a successful notarization looks like

Per Apple's `xcrun notarytool` documentation, a successful flow produces:

```
Successfully uploaded file
  id: <uuid>
  path: /path/to/binary.zip

Waiting for processing to complete.
Current status: Accepted..............
Processing complete
  id: <uuid>
  status: Accepted
```

After "Accepted", the binary can be stapled with `xcrun stapler staple`.

### What we are getting

```
Successfully uploaded file
  id: <uuid>
  path: /path/to/binary.zip

Waiting for processing to complete.
Current status: In Progress..........[repeats for hours]
Timeout of <N> second(s) was reached before processing completed.
```

The submission never transitions to "Accepted" or "Invalid". It remains perpetually "In Progress".

### What a rejected notarization looks like (for reference)

```
Processing complete
  id: <uuid>
  status: Invalid

[notarytool log <uuid>] shows specific rejection reasons
```

We are not even getting "Invalid" — the service simply never completes processing.

---

## Analysis: Why Notarization Never Completes

### Ruled out

- **Invalid certificates** — Code signing and verification succeed, proving the Developer ID certificate is valid
- **Timeout too short** — Tested from 15 minutes to 4 hours with identical results
- **Upload failure** — Every submission uploads successfully with a unique UUID
- **Binary format issue** — If the binary were malformed, Apple would return "Invalid", not perpetual "In Progress"

### Likely causes (in order of probability)

1. **Apple Developer account issue** — The account may have a pending agreement, unverified enrollment, or billing hold that prevents notarization processing while still allowing uploads
2. **App-specific password restrictions** — The app-specific password used for `APPLE_NOTARIZATION_PASSWORD` may not have the required permissions for notarization, or may have expired
3. **Team ID mismatch** — The `APPLE_NOTARIZATION_TEAM_ID` may not match the team associated with the signing certificate
4. **Apple-side service issue** — Less likely given the duration (8+ hours across 32 runs), but Apple's notarization service has had outages before
5. **Rate limiting** — The 30+ rapid-fire submissions from the same account may have triggered a processing queue block (though Apple typically returns errors for rate limits rather than silent stalls)

---

## Recommended Actions

### Immediate (manual investigation required)

1. **Check Apple Developer account status** — Log into developer.apple.com, verify:
   - Enrollment is active
   - All agreements (especially the updated Developer Program License Agreement) are accepted
   - No billing holds
   - Notarization is enabled for the team

2. **Query a stuck submission manually:**
   ```bash
   xcrun notarytool info 3b099c4d-fcc2-4dc5-9565-4cbcb9a62c9f \
     --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
     --password "$APPLE_NOTARIZATION_PASSWORD" \
     --team-id "$APPLE_NOTARIZATION_TEAM_ID"
   ```

3. **Retrieve the notarization log** (may reveal rejection reasons not shown in status):
   ```bash
   xcrun notarytool log 3b099c4d-fcc2-4dc5-9565-4cbcb9a62c9f \
     --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
     --password "$APPLE_NOTARIZATION_PASSWORD" \
     --team-id "$APPLE_NOTARIZATION_TEAM_ID"
   ```

4. **Verify the app-specific password** is valid — generate a fresh one at appleid.apple.com and update the `APPLE_NOTARIZATION_PASSWORD` secret

### CI improvements (after root cause is resolved)

5. **Stop increasing the timeout** — 4 hours is already extreme. The problem is not timeout duration.

6. **Add `notarytool log` retrieval on failure** — Modify the CI step to fetch and print the notarization log when `--wait` times out, providing diagnostic data automatically:
   ```yaml
   - name: Notarize binaries
     run: |
       for BINARY in threedoors-darwin-arm64 threedoors-darwin-amd64; do
         zip "${BINARY}.zip" "$BINARY"
         SUBMISSION_ID=$(xcrun notarytool submit "${BINARY}.zip" \
           --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
           --password "$APPLE_NOTARIZATION_PASSWORD" \
           --team-id "$APPLE_NOTARIZATION_TEAM_ID" \
           --wait --timeout 1800 2>&1 | grep "id:" | head -1 | awk '{print $2}')
         if [ $? -ne 0 ]; then
           echo "::warning::Notarization timed out. Fetching log..."
           xcrun notarytool log "$SUBMISSION_ID" \
             --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
             --password "$APPLE_NOTARIZATION_PASSWORD" \
             --team-id "$APPLE_NOTARIZATION_TEAM_ID" || true
         fi
         rm "${BINARY}.zip"
       done
   ```

7. **Consider separating notarization into an async workflow** — Submit without `--wait`, store the submission ID, and check status in a separate workflow or manual step.

---

## Impact on Releases

The CI pipeline has a resilient fallback: the release job checks `needs.sign-and-notarize.result != 'success'` and downloads unsigned binaries instead. Every release has been created successfully with unsigned binaries despite the signing failures. Users can still download and run the binary — they will see a Gatekeeper warning on first launch that can be bypassed with right-click → Open.

---

## Run Count Summary

| Timeout | Run Count | Result |
|---------|-----------|--------|
| 900s (15 min) | ~6 runs | All timeout |
| 1800s (30 min) | ~5 runs | All timeout |
| 3600s (1 hour) | ~7 runs | All timeout |
| 14400s (4 hours) | ~14 runs | All timeout or network error |
| **Total** | **32 runs** | **Zero successes** |
