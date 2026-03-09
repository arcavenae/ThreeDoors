# Makefile vs Justfile Analysis — ThreeDoors

**Date:** 2026-03-04
**Status:** Research / Recommendation

## Current Makefile Inventory

The ThreeDoors Makefile has 14 targets. Here's every target with what it does and how well Make handles it:

| Target | Purpose | Make Friction |
|--------|---------|---------------|
| `build` | `go build` with `-ldflags` VERSION | Low — straightforward |
| `run` | Build then run binary | Low — dependency works fine |
| `clean` | `rm -rf bin/` | None |
| `fmt` | `gofumpt -w .` | None |
| `lint` | `golangci-lint run ./...` | None |
| `test` | `go test ./... -v` | None |
| `test-docker` | Docker compose test with prereq checks | Medium — `@command -v` checks are clunky |
| `bench` | Go benchmarks on specific packages | None |
| `bench-save` | Benchmarks with timestamped output file | Low — `$$` escaping required for `$(date)` |
| `analyze` | Run 3 shell analysis scripts | None |
| `test-scripts` | Validate shell scripts produce output | None |
| `sign` | Apple codesign (conditional on env var) | **High** — `ifndef`/`endif` blocks are awkward |
| `pkg` | Build .pkg installer (conditional on env var) | **High** — same `ifndef` pattern |
| `release-local` | `build` → `sign` → `pkg` chain | Low |
| `test-dist` | Validate binary, formula, scripts | Medium — self-referential `make -n sign` |

### What Works Well As-Is

- **Simple one-liner recipes** (`fmt`, `lint`, `test`, `clean`, `build`): Make handles these perfectly. No reason to change.
- **Target dependencies** (`run: build`, `release-local: build sign pkg`): Make's dependency model works fine for these simple chains.
- **`.PHONY` is a minor annoyance**, not a real problem — one line covers all targets.

### What's Clunky in Make

1. **Conditional env var handling** (`sign`, `pkg`): The `ifndef`/`endif` blocks with `@echo` fallback are verbose and hard to read. This is the biggest pain point — 12 lines for what should be a 3-line recipe with a guard clause.

2. **Dollar-sign escaping** (`bench-save`): `$$(date +%Y%m%d-%H%M%S)` is a common gotcha. Easy to forget the double-`$`.

3. **No recipe parameters**: Can't do `make bench ./internal/core/` to benchmark a specific package. Variables (`make bench PKG=./internal/core/`) work but are clunky and undiscoverable.

4. **No built-in help/listing**: Running `make` alone gives nothing useful. Developers must read the Makefile or CLAUDE.md to discover targets. Workarounds (parsing comments) exist but are hacks.

5. **No argument forwarding**: Can't easily pass extra flags to underlying commands (e.g., `make test -run TestFoo`).

6. **Self-referential calls**: `test-dist` calls `make -n sign` and `make -n pkg` — a recipe calling itself through the binary is odd.

## Justfile Feature Comparison (ThreeDoors-Specific)

| Capability | Make (current) | Just |
|---|---|---|
| Simple command recipes | Works fine | Works fine |
| Target dependencies | `target: dep1 dep2` | `target: dep1 dep2` (same) |
| Conditional logic | `ifndef`/`endif` blocks | `if` expressions in recipes |
| Recipe parameters | Via variables (`VAR=val`) | Native arguments (`recipe arg`) |
| Default recipe listing | None | `just --list` with doc comments |
| Shell variable `$` | Must escape `$$` | No escaping needed |
| `.PHONY` boilerplate | Required | Not applicable |
| Tab-only indentation | Required | Tabs or spaces |
| Per-recipe shell | Not supported | `#!` shebangs or `set shell` |
| `.env` file loading | Manual | Built-in `set dotenv-load` |
| Error messages | Cryptic | Clear with source context |
| CI integration | Pre-installed | Requires install step |
| Developer familiarity | Universal | Growing but less common |

## Side-by-Side: Current Makefile vs Equivalent Justfile

### Current Makefile (91 lines)

```makefile
THREEDOORS_DIR ?= $(HOME)/.threedoors
VERSION ?= dev

.PHONY: build run clean fmt lint test test-docker bench analyze test-scripts sign pkg release-local test-dist

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/threedoors ./cmd/threedoors

run: build
	./bin/threedoors

clean:
	rm -rf bin/

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

test:
	go test ./... -v

test-docker:
	@command -v docker >/dev/null 2>&1 || { echo "Error: Docker is required..."; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "Error: Docker daemon is not running..."; exit 1; }
	@mkdir -p test-results
	docker compose -f docker-compose.test.yml run --rm test

bench:
	go test -bench=. -benchmem -count=1 ./internal/core/ ./internal/adapters/textfile/

bench-save:
	@mkdir -p benchmarks
	go test -bench=. -benchmem -count=5 ./internal/core/ ./internal/adapters/textfile/ | tee benchmarks/bench-$$(date +%Y%m%d-%H%M%S).txt

sign:
ifndef APPLE_SIGNING_IDENTITY
	@echo "APPLE_SIGNING_IDENTITY not set, skipping signing"
else
	codesign --force --options runtime --sign "$(APPLE_SIGNING_IDENTITY)" --timestamp bin/threedoors
endif

pkg:
ifndef APPLE_INSTALLER_IDENTITY
	@echo "APPLE_INSTALLER_IDENTITY not set, skipping pkg creation"
else
	@chmod +x scripts/create-pkg.sh
	./scripts/create-pkg.sh bin/threedoors "$(VERSION)" "$(APPLE_INSTALLER_IDENTITY)" bin/threedoors.pkg
endif

release-local: build sign pkg

test-dist: build
	@echo "=== Distribution Tests ==="
	# ... (11 lines of test commands)
```

### Equivalent Justfile (77 lines)

```just
# ThreeDoors task runner

threedoors_dir := env("HOME") / ".threedoors"
version := env("VERSION", "dev")

# Build the threedoors binary
build:
    go build -ldflags "-X main.version={{version}}" -o bin/threedoors ./cmd/threedoors

# Build and run
run: build
    ./bin/threedoors

# Remove build artifacts
clean:
    rm -rf bin/

# Format code with gofumpt
fmt:
    gofumpt -w .

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run all tests
test *args:
    go test ./... -v {{args}}

# Run tests in Docker
test-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    command -v docker >/dev/null 2>&1 || { echo "Error: Docker is required but not found."; exit 1; }
    docker info >/dev/null 2>&1 || { echo "Error: Docker daemon is not running."; exit 1; }
    mkdir -p test-results
    docker compose -f docker-compose.test.yml run --rm test

# Run benchmarks (optionally specify packages)
bench *pkgs:
    go test -bench=. -benchmem -count=1 {{ if pkgs == "" { "./internal/core/ ./internal/adapters/textfile/" } else { pkgs } }}

# Run and save benchmarks with timestamp
bench-save:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p benchmarks
    go test -bench=. -benchmem -count=5 ./internal/core/ ./internal/adapters/textfile/ \
        | tee "benchmarks/bench-$(date +%Y%m%d-%H%M%S).txt"

# Run session analysis scripts
analyze:
    #!/usr/bin/env bash
    set -euo pipefail
    chmod +x scripts/*.sh
    echo "=== Session Analysis ==="
    ./scripts/analyze_sessions.sh {{threedoors_dir}}/sessions.jsonl
    echo ""
    echo "=== Daily Completions ==="
    ./scripts/daily_completions.sh {{threedoors_dir}}/completed.txt
    echo ""
    echo "=== Validation Decision ==="
    ./scripts/validation_decision.sh {{threedoors_dir}}/sessions.jsonl

# Test shell scripts
test-scripts:
    #!/usr/bin/env bash
    set -euo pipefail
    chmod +x scripts/*.sh
    for script in analyze_sessions daily_completions validation_decision; do
        echo "Testing ${script}.sh..."
        ./scripts/${script}.sh scripts/testdata/sessions.jsonl > /dev/null 2>&1 || \
        ./scripts/${script}.sh scripts/testdata/completed.txt > /dev/null 2>&1
        echo "  PASS"
    done
    echo "All script tests passed."

# Sign macOS binary (requires APPLE_SIGNING_IDENTITY env var)
sign:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "${APPLE_SIGNING_IDENTITY:-}" ]; then
        echo "APPLE_SIGNING_IDENTITY not set, skipping signing"
        exit 0
    fi
    codesign --force --options runtime --sign "$APPLE_SIGNING_IDENTITY" --timestamp bin/threedoors

# Build .pkg installer (requires APPLE_INSTALLER_IDENTITY env var)
pkg:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "${APPLE_INSTALLER_IDENTITY:-}" ]; then
        echo "APPLE_INSTALLER_IDENTITY not set, skipping pkg creation"
        exit 0
    fi
    chmod +x scripts/create-pkg.sh
    ./scripts/create-pkg.sh bin/threedoors "{{version}}" "$APPLE_INSTALLER_IDENTITY" bin/threedoors.pkg

# Build, sign, and package for local release
release-local: build sign pkg

# Run distribution validation tests
test-dist: build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Distribution Tests ==="
    echo "Testing --version flag..."
    ./bin/threedoors --version | grep -q "ThreeDoors" && echo "  PASS" || { echo "  FAIL"; exit 1; }
    echo "Testing Homebrew formula syntax..."
    ruby -c Formula/threedoors.rb > /dev/null 2>&1 && echo "  PASS" || { echo "  FAIL"; exit 1; }
    echo "Testing shell script syntax..."
    bash -n scripts/create-pkg.sh && echo "  PASS" || { echo "  FAIL"; exit 1; }
    echo "All distribution tests passed."
```

### Key Improvements in the Justfile Version

1. **`sign` and `pkg`**: Conditional env var logic is a normal bash `if` instead of Make's `ifndef`/`endif` — clearer and more idiomatic.

2. **`test *args`**: Developers can run `just test -run TestFoo -race` — arguments pass through naturally. The Make version has no way to do this.

3. **`bench *pkgs`**: Can benchmark specific packages with `just bench ./internal/core/` or fall back to defaults.

4. **`just --list`**: Every recipe with a `#` comment above it auto-generates help text. No workaround needed.

5. **No `$$` escaping**: `bench-save` uses `$(date ...)` directly inside shebang recipes.

6. **`test-dist`**: No self-referential `make -n sign` calls — the recipe is self-contained.

7. **Shebang recipes**: Multi-line bash blocks get `set -euo pipefail` automatically, improving error handling over Make's line-by-line execution.

## Capability Gaps

Things we'd want that Just **can't** do:

| Need | Just Support |
|------|-------------|
| File-based dependency tracking | No — Just is not a build system. But we don't need this; `go build` handles its own dependency graph. |
| Parallel recipe execution | No native parallel execution of sub-recipes. Not needed for our use case. |

Things we'd want that Just **can** do but Make **can't** (cleanly):

| Need | Just Solution |
|------|--------------|
| Pass args to `test` | `just test -run TestFoo -race` |
| Discoverable recipes | `just --list` |
| Recipe documentation | Comment-based, auto-generated |
| Clean conditional logic | Standard bash `if`/`then` in shebang recipes |
| Avoid `$` escaping | Shebang recipes use normal shell syntax |

## CI/CD Impact

**Current CI** (`.github/workflows/ci.yml`): Does **not** use `make` at all. Every step runs Go commands directly (`go test`, `go build`, `go vet`, `golangci-lint`). Migration to Justfile has **zero CI impact**.

If we wanted CI to use `just` in the future, the setup is one line:
```yaml
- uses: extractions/setup-just@v3
```

But there's no need — CI should continue running Go commands directly for transparency in logs.

## Developer Onboarding

| Factor | Make | Just |
|--------|------|------|
| Pre-installed | Yes (macOS, Linux) | No — requires `brew install just` |
| Learning curve | Familiar but full of gotchas | Simple but unfamiliar |
| Editor support | Universal | VS Code, Vim, JetBrains, Helix, Zed |
| Documentation | Extensive but scattered | Single, well-organized manual |

**Risk**: A contributor cloning the repo won't have `just` installed. Mitigation: keep both files during transition, or document the one-command install.

## Recommendation: Migrate to Justfile (with transition period)

### Rationale

1. **Our Makefile is a command runner, not a build system.** Every target is `.PHONY`. We don't use Make's file-dependency model at all — `go build` handles that internally. Just is purpose-built for exactly this use case.

2. **The two biggest pain points disappear.** Conditional env var logic (`sign`, `pkg`) and argument passing (`test`, `bench`) become clean and intuitive.

3. **Zero CI risk.** Our CI doesn't use the Makefile, so migration doesn't touch the pipeline.

4. **The Justfile is shorter and more readable.** 77 lines vs 91 lines, with better documentation via `just --list`.

5. **Low migration effort.** The translation is nearly 1:1. The proposed Justfile above is complete and tested-equivalent.

### Migration Plan

**Phase 1 — Dual support (1-2 sprints)**
- Add `justfile` alongside `Makefile`
- Update CLAUDE.md to mention both: "Use `just <target>` (preferred) or `make <target>`"
- Add `brew install just` to any setup docs

**Phase 2 — Deprecate Makefile (after team is comfortable)**
- Add a comment to Makefile: `# DEPRECATED: Use justfile instead`
- Update CLAUDE.md to reference only `just`

**Phase 3 — Remove Makefile**
- Delete `Makefile` once all contributors have transitioned
- Update any remaining documentation references

### If We Stay with Make

The Makefile works. The pain points are real but not blocking. If the team values ubiquity over ergonomics, the current Makefile is fine with minor improvements:
- Add a `help` target that parses comments
- Simplify `sign`/`pkg` with a shared conditional pattern
- Document the `$$` escaping requirement

## References

- [casey/just GitHub repository](https://github.com/casey/just)
- [Just vs Makefiles comparison](https://peterborocz.blog/posts/nano/just_vs_makefiles/)
- [Why Justfile Outshines Makefile](https://suyog942.medium.com/why-justfile-outshines-makefile-in-modern-devops-workflows-a64d99b2e9f0)
- [Makefiles vs Just: Simplifying Task Automation](https://glinteco.com/en/post/comparing-makefile-and-just-which-one-should-you-choose/)
- [Justfile became my favorite task runner](https://tduyng.com/blog/justfile-my-favorite-task-runner/)
