# Party Mode Session 4: Docker and E2E Test Acceleration

**Participants:** Architect, TEA/QA, Dev, PM, SM
**Date:** 2026-03-11
**Topic:** Optimize Docker E2E tests and the overall E2E testing strategy

---

## Current Docker E2E Setup

### Dockerfile.test Analysis

```dockerfile
FROM golang:1.25-bookworm
# Installs gofumpt + golangci-lint (redundant with quality-gate)
# Copies go.mod/go.sum for layer caching
# Runs as non-root testuser
# Default: go test ./... -v -count=1 -timeout 5m
```

**Key observations:**
1. **Full test suite rerun**: Docker E2E runs the entire `go test ./...` — identical to Quality Gate minus race detection
2. **CGO_ENABLED=0**: Tests run in pure-Go mode. Quality Gate uses default (CGO enabled with race detector)
3. **Source mounted read-only**: Good — code changes don't require image rebuild
4. **Layer caching**: Docker Buildx cache is configured, but the base image is large (golang:1.25-bookworm is ~1.2GB)
5. **Tool version skew**: Dockerfile pins `golangci-lint v2.1.6` while CI uses `v2.10.1` — potential inconsistency

### docker-compose.test.yml

```yaml
volumes:
  - .:/app:ro        # Source code (read-only)
  - ./test-results:/app/test-results:rw  # Results output
environment:
  - CGO_ENABLED=0
  - GOFLAGS=-count=1
  - GOPATH=/go
```

### CI Docker E2E Job Timing Breakdown

| Step | Estimated Duration |
|------|-------------------|
| Checkout | ~3s |
| Setup Docker Buildx | ~5s |
| Restore Docker layer cache | ~10s |
| Build Docker test image | ~30s (cache hit) / ~90s (cold) |
| Create test-results directory | ~1s |
| Run Docker E2E tests | ~90s |
| Check golden file diffs | ~2s |
| Upload test results | ~5s |
| Rotate cache | ~3s |
| **Total** | **~2m30s (warm) / ~3m30s (cold)** |

---

## Analysis: Is Docker E2E Providing Value?

### What Docker E2E catches that Quality Gate doesn't:

| Scenario | Docker E2E? | Quality Gate? | Likelihood |
|----------|-------------|---------------|------------|
| Linux-specific file path bugs | Yes | No (runs on ubuntu too) | Very low — CI is already Linux |
| CGO=0 build issues | Yes | No (CGO enabled) | Very low — project has no CGO deps |
| Different Go toolchain version | Possible | No | Only if Dockerfile.test lags |
| Container-specific behavior | Yes | No | N/A — not a containerized app |
| Golden file rendering differences | Yes | Also yes | Identical environments |

**Key insight:** Both Quality Gate and Docker E2E run on `ubuntu-latest`. The Docker E2E just adds a container layer. Since ThreeDoors is a local TUI app (not deployed in containers), the Docker layer provides negligible additional validation.

### Architect's Recommendation

The Docker E2E job was created for Story 18.4 — establishing reproducible E2E testing. At that time, it served as the canonical "runs the same everywhere" environment. Now that Quality Gate is mature and runs the same tests with race detection, the Docker E2E is redundant for PR validation.

**Recommendation:** Keep Docker E2E for push-to-main only (defense in depth) or convert to a nightly cron.

---

## Docker Build Optimization (if retained)

### 1. Slim Base Image

Replace `golang:1.25-bookworm` (~1.2GB) with `golang:1.25-alpine` (~350MB):

```dockerfile
FROM golang:1.25-alpine
RUN apk add --no-cache git gcc musl-dev
```

**Expected impact:** ~30% faster image pull on cold cache. Build step from ~90s → ~60s cold.

**Risk:** Alpine uses musl libc, which can cause subtle test differences. For CGO_ENABLED=0, this doesn't matter.

### 2. Fix Tool Version Skew

Dockerfile.test pins `golangci-lint v2.1.6`. CI uses `v2.10.1`. This is a bug waiting to happen.

```dockerfile
ARG GOLANGCI_LINT_VERSION=v2.10.1  # Match ci.yml
```

### 3. Remove Redundant Tools from Docker Image

Since Docker E2E runs tests only (no linting), remove `gofumpt` and `golangci-lint` from the Dockerfile:

```dockerfile
FROM golang:1.25-bookworm
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["go", "test", "./...", "-v", "-count=1", "-timeout", "5m"]
```

**Expected impact:** Faster image build (~15s saved). Smaller image (~200MB less).

---

## TUI E2E Test Strategy

### Current Approach

TUI E2E tests use the `teatest` package from Bubbletea. Each test:
1. Creates a `teatest.NewModel(m)` with the app's root model
2. Sends key events via `tm.Send(keyMsg)`
3. Waits via `time.Sleep(200ms)` for the model to process
4. Reads output via `tm.FinalOutput()` or `tm.Output()`

### The Sleep Problem

Each `time.Sleep(200ms)` between keystrokes is:
- **Conservative**: Most events process in <10ms
- **Necessary**: Without it, events queue and tests race
- **Cumulative**: 58 sleeps × 200ms = 11.6s just waiting

### Optimization: WaitFor Pattern

Teatest provides `teatest.WaitFor` which polls for expected output:

```go
// Polls every 10ms (default), fails after 2s timeout
teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
    return bytes.Contains(bts, []byte("expected text"))
}, teatest.WithDuration(2*time.Second))
```

**Benefits:**
- Resolves in 10-50ms instead of fixed 200ms
- Self-documenting: makes expected behavior explicit
- More stable: doesn't rely on timing assumptions
- Faster: 4-20x faster per wait point

**Migration strategy:**
1. Start with the slowest tests (TestWorkflow_MultipleRerolls: 9s)
2. Replace each `time.Sleep` with the appropriate `WaitFor` condition
3. Run with `-race` after each migration to catch new races
4. Preserve assertion semantics — same checks, faster waits

### Optimization: Reduce Teatest Model Initialization

Some E2E tests create the full app model with all providers loaded. If tests only exercise TUI behavior, a mock provider is faster:

```go
// Instead of loading real YAML files
provider := tasks.NewMockProvider(testTasks)
m := NewModel(WithProvider(provider))
```

This saves ~50ms per test on provider initialization.

---

## Consensus Recommendations

### Immediate (no code changes):
1. **Move Docker E2E to push-only** — saves CI runner time, no wall-clock impact for PRs
2. **Fix golangci-lint version skew** in Dockerfile.test

### Short-term (code changes, high impact):
3. **Migrate TUI E2E sleeps to WaitFor** — the single highest-impact optimization
4. **Strip unnecessary tools from Dockerfile.test** — if retaining Docker E2E

### Medium-term (architecture changes):
5. **Consider replacing Docker E2E with nightly cron** — if post-push validation is sufficient
6. **Evaluate teatest.WithTimeout for global test timeout** — prevents individual test hangs from blocking the suite

### Rejected:
- **Running Docker E2E in a separate workflow** — overcomplicates required checks
- **Using Docker-in-Docker for parallel test execution** — overkill for this project size
- **Pre-building test binaries in Docker** — complexity doesn't justify gain
