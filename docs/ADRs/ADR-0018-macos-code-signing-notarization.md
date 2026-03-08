# ADR-0018: macOS Code Signing and Notarization

- **Status:** Accepted
- **Date:** 2025-11-09
- **Decision Makers:** Project founder
- **Related PRs:** #30, #46, #61, #67, #76, #88, #97, #99, #100, #101, #110, #111, #113
- **Related ADRs:** ADR-0019 (Docker E2E Testing)

## Context

macOS Gatekeeper blocks unsigned binaries downloaded from the internet. For ThreeDoors to be distributable to users without requiring manual security bypass, the binary must be code-signed and notarized by Apple.

## Decision

Implement **CI-based code signing and notarization** via GitHub Actions:

1. Binary signed with Apple Developer ID certificate
2. Notarized via `notarytool` in CI pipeline
3. `.app` bundle and `.dmg` installer for GUI distribution
4. Signed `.pkg` installer for command-line installation
5. Homebrew formula for developer-oriented installation

## Rationale

- Required for user trust — unsigned binaries trigger Gatekeeper warnings
- CI-based ensures every release is signed consistently
- Multiple distribution formats serve different user preferences
- Apple's notarization provides malware scanning assurance

## Implementation Timeline

This was one of the most challenging infrastructure efforts, requiring 13+ PRs:

| PR | Issue Addressed |
|----|-----------------|
| #30 | Initial packaging setup |
| #46 | Code signing research |
| #61 | CI secret alignment |
| #67, #76, #88 | Notarization timeout increases (5min → 30min → 1hr → 4hr) |
| #97 | Validation gate decision documentation |
| #99-#101 | Apple intermediate certificate chain issues |
| #110 | `.app` and `.dmg` packaging |
| #111, #113 | Installer certificate keychain access |

## Consequences

### Positive
- Clean installation experience for end users
- No Gatekeeper bypass required
- Automated in CI — no manual signing steps
- Multiple distribution formats (binary, .app, .dmg, .pkg, Homebrew)

### Negative
- Apple Developer Program membership required ($99/year)
- Notarization adds 2-10 minutes to CI pipeline
- Certificate management complexity in CI secrets
- Apple intermediate certificate changes require manual updates
