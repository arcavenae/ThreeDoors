# Infrastructure and Deployment

## Infrastructure as Code

**Tool:** GitHub Actions

**Approach:** ThreeDoors uses GitHub Actions for CI/CD on public runners. The application runs locally with no cloud infrastructure for runtime.

## Deployment Strategy

**Strategy:** Direct Binary Distribution

**Build Process:**
```bash
just build    # Compiles to bin/threedoors
```

**Installation:**
```bash
# Option 1: Manual install
cp bin/threedoors /usr/local/bin/

# Option 2: Run from project directory
just run

# Option 3 (Future): Homebrew tap
brew install arcavenae/tap/threedoors
```

**CI/CD Platform:** GitHub Actions (`.github/workflows/ci.yml`)

**PR Quality Gates** (runs on every `pull_request` to `main`):
- `gofumpt` formatting check
- `go vet` correctness check
- `golangci-lint` static analysis
- `go test` unit tests
- `go build` build validation

**Alpha Release** (runs on `push` to `main`, i.e., PR merge):
- Cross-compiles binaries for darwin/arm64, linux/amd64
- Uploads as GitHub Actions workflow artifacts (14-day retention)

**Recommended:** Enable branch protection on `main` requiring the quality-gate job to pass before merge.

## Environments

**Development:**
- Purpose: Local development and testing
- Location: Developer's macOS machine
- Data: `~/.threedoors/` (can be deleted/reset)

**Production (User Environment):**
- Purpose: End-user execution
- Location: User's macOS machine
- Data: `~/.threedoors/` (user's actual task data)

## Rollback Strategy

**Primary Method:** User keeps previous binary

**Rollback Process:**
```bash
# User manually switches to previous version
cp threedoors.old /usr/local/bin/threedoors
```

**Data Compatibility:**
- YAML schema must remain backward compatible
- Forward migrations add fields with defaults
- Never break existing tasks.yaml format

## Post-MVP Deployment Considerations (Phase 2–3)

**Additional Installation Method (Epic 5):**

```bash
# Homebrew tap (code-signed + notarized)
brew install arcavenae/tap/threedoors

# DMG/pkg installer (alternative)
# Download from GitHub Releases
```

**Code Signing & Notarization (Epic 5):**
- Apple Developer certificate for binary signing
- Apple notarization service for Gatekeeper approval
- Automated in CI pipeline (GitHub Actions)
- Cross-compiled for darwin/arm64, linux/amd64

**Runtime Dependencies (Post-MVP):**
- No new external service dependencies at runtime
- SQLite (embedded via modernc.org/sqlite) — no separate server
- LLM backends are opt-in and user-configured
- All calendar integration is local-only (no cloud APIs)

**Data Migration Path:**
- Phase 1 → Phase 2: `tasks.yaml` format preserved; new `config.yaml` created on first run with defaults
- Phase 2 → Phase 3: `enrichment.db` and `sync-state/` created on demand
- All migrations are additive (new files/tables), never destructive

---
