# ThreeDoors justfile — local build, sign, and notarize workflow
# Loads .env.local automatically for signing secrets
set dotenv-load
set dotenv-filename := ".env.local"

THREEDOORS_DIR := env("THREEDOORS_DIR", env("HOME") / ".threedoors")
VERSION := env("VERSION", "dev")

# ─── Core recipes (mirror Makefile) ───────────────────────────────

# Build the binary
build:
    go build -ldflags "-X main.version={{VERSION}}" -o bin/threedoors ./cmd/threedoors

# Build and run
run: build
    ./bin/threedoors

# Remove build artifacts
clean:
    rm -rf bin/

# Format with gofumpt
fmt:
    gofumpt -w .

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run all tests
test:
    go test ./... -v

# Run shell-script analytics
analyze:
    @chmod +x scripts/*.sh
    @echo "=== Session Analysis ==="
    @./scripts/analyze_sessions.sh {{THREEDOORS_DIR}}/sessions.jsonl
    @echo ""
    @echo "=== Daily Completions ==="
    @./scripts/daily_completions.sh {{THREEDOORS_DIR}}/completed.txt
    @echo ""
    @echo "=== Validation Decision ==="
    @./scripts/validation_decision.sh {{THREEDOORS_DIR}}/sessions.jsonl

# Run script unit tests
test-scripts:
    @chmod +x scripts/*.sh
    @echo "Testing analyze_sessions.sh..."
    @./scripts/analyze_sessions.sh scripts/testdata/sessions.jsonl > /dev/null
    @echo "  PASS"
    @echo "Testing daily_completions.sh..."
    @./scripts/daily_completions.sh scripts/testdata/completed.txt > /dev/null
    @echo "  PASS"
    @echo "Testing validation_decision.sh..."
    @./scripts/validation_decision.sh scripts/testdata/sessions.jsonl > /dev/null
    @echo "  PASS"
    @echo "All script tests passed."

# Run distribution smoke tests
test-dist: build
    @echo "=== Distribution Tests ==="
    @echo "Testing --version flag..."
    @./bin/threedoors --version | grep -q "ThreeDoors" && echo "  PASS" || (echo "  FAIL" && exit 1)
    @echo "Testing Homebrew formula syntax..."
    @ruby -c Formula/threedoors.rb > /dev/null 2>&1 && echo "  PASS" || (echo "  FAIL" && exit 1)
    @echo "Testing shell script syntax..."
    @bash -n scripts/create-pkg.sh && echo "  PASS" || (echo "  FAIL" && exit 1)
    @echo "All distribution tests passed."

# ─── Cross-compile (mirrors CI) ──────────────────────────────────

# Build for all release targets
build-all:
    GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version={{VERSION}}" -o bin/threedoors-darwin-arm64 ./cmd/threedoors
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version={{VERSION}}" -o bin/threedoors-darwin-amd64 ./cmd/threedoors
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version={{VERSION}}" -o bin/threedoors-linux-amd64 ./cmd/threedoors

# ─── Signing & packaging ─────────────────────────────────────────

# Codesign the binary
sign: (_require-var "APPLE_SIGNING_IDENTITY" env("APPLE_SIGNING_IDENTITY", ""))
    codesign --force --options runtime --sign "${APPLE_SIGNING_IDENTITY}" --timestamp bin/threedoors

# Verify the codesign signature
verify:
    codesign --verify --verbose=2 bin/threedoors

# Create .app bundle from signed binary
app:
    @chmod +x scripts/create-app.sh
    ./scripts/create-app.sh bin/threedoors "{{VERSION}}" bin

# Codesign the .app bundle
sign-app: (_require-var "APPLE_SIGNING_IDENTITY" env("APPLE_SIGNING_IDENTITY", ""))
    codesign --force --deep --options runtime --sign "${APPLE_SIGNING_IDENTITY}" --timestamp bin/ThreeDoors.app

# Create .dmg disk image from .app bundle
dmg:
    @chmod +x scripts/create-dmg.sh
    ./scripts/create-dmg.sh bin/ThreeDoors.app "{{VERSION}}" bin/ThreeDoors.dmg

# Create a signed pkg installer
pkg: (_require-var "APPLE_INSTALLER_IDENTITY" env("APPLE_INSTALLER_IDENTITY", ""))
    @chmod +x scripts/create-pkg.sh
    ./scripts/create-pkg.sh bin/threedoors "{{VERSION}}" "${APPLE_INSTALLER_IDENTITY}" bin/threedoors.pkg

# Notarize an artifact (pkg, dmg, or zip)
notarize artifact: (_require-var "APPLE_NOTARIZATION_APPLE_ID" env("APPLE_NOTARIZATION_APPLE_ID", "")) (_require-var "APPLE_NOTARIZATION_PASSWORD" env("APPLE_NOTARIZATION_PASSWORD", "")) (_require-var "APPLE_NOTARIZATION_TEAM_ID" env("APPLE_NOTARIZATION_TEAM_ID", ""))
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Submitting {{artifact}} to Apple notarytool..."
    xcrun notarytool submit "{{artifact}}" \
        --apple-id "${APPLE_NOTARIZATION_APPLE_ID}" \
        --password "${APPLE_NOTARIZATION_PASSWORD}" \
        --team-id "${APPLE_NOTARIZATION_TEAM_ID}" \
        --wait --timeout 14400
    echo "Stapling notarization ticket..."
    xcrun stapler staple "{{artifact}}"
    echo "Notarization complete for {{artifact}}."

# Full local release: build all formats, sign, notarize
release-local: build sign verify app sign-app dmg pkg
    just notarize bin/threedoors.pkg
    just notarize bin/ThreeDoors.dmg
    @echo "Local release complete."

# ─── Diagnostics ─────────────────────────────────────────────────

# Show available codesigning identities
sign-check:
    security find-identity -v -p codesigning

# Check notarization status for a submission ID
notarize-status id: (_require-var "APPLE_NOTARIZATION_APPLE_ID" env("APPLE_NOTARIZATION_APPLE_ID", "")) (_require-var "APPLE_NOTARIZATION_PASSWORD" env("APPLE_NOTARIZATION_PASSWORD", "")) (_require-var "APPLE_NOTARIZATION_TEAM_ID" env("APPLE_NOTARIZATION_TEAM_ID", ""))
    xcrun notarytool info "{{id}}" \
        --apple-id "${APPLE_NOTARIZATION_APPLE_ID}" \
        --password "${APPLE_NOTARIZATION_PASSWORD}" \
        --team-id "${APPLE_NOTARIZATION_TEAM_ID}"

# Get Apple's detailed log for a submission ID
notarize-log id: (_require-var "APPLE_NOTARIZATION_APPLE_ID" env("APPLE_NOTARIZATION_APPLE_ID", "")) (_require-var "APPLE_NOTARIZATION_PASSWORD" env("APPLE_NOTARIZATION_PASSWORD", "")) (_require-var "APPLE_NOTARIZATION_TEAM_ID" env("APPLE_NOTARIZATION_TEAM_ID", ""))
    xcrun notarytool log "{{id}}" \
        --apple-id "${APPLE_NOTARIZATION_APPLE_ID}" \
        --password "${APPLE_NOTARIZATION_PASSWORD}" \
        --team-id "${APPLE_NOTARIZATION_TEAM_ID}"

# List recent notarization submissions
notarize-history: (_require-var "APPLE_NOTARIZATION_APPLE_ID" env("APPLE_NOTARIZATION_APPLE_ID", "")) (_require-var "APPLE_NOTARIZATION_PASSWORD" env("APPLE_NOTARIZATION_PASSWORD", "")) (_require-var "APPLE_NOTARIZATION_TEAM_ID" env("APPLE_NOTARIZATION_TEAM_ID", ""))
    xcrun notarytool history \
        --apple-id "${APPLE_NOTARIZATION_APPLE_ID}" \
        --password "${APPLE_NOTARIZATION_PASSWORD}" \
        --team-id "${APPLE_NOTARIZATION_TEAM_ID}"

# Run Gatekeeper assessment on the binary
gatekeeper-check:
    spctl --assess --verbose=2 bin/threedoors

# ─── Internal helpers ─────────────────────────────────────────────

# Check that an env var is set, print setup instructions if not
[private]
_require-var name value:
    #!/usr/bin/env bash
    if [ -z "{{value}}" ]; then
        echo "ERROR: {{name}} is not set."
        echo ""
        echo "To set up local signing:"
        echo "  1. Copy .env.local.example to .env.local"
        echo "  2. Fill in your values"
        echo "  3. Run 'just sign-check' to find your signing identities"
        echo ""
        echo ".env.local is gitignored (*.env.* pattern)."
        exit 1
    fi
