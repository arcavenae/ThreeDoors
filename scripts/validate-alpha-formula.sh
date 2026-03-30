#!/usr/bin/env bash
# validate-alpha-formula.sh — Extract the alpha Homebrew formula template
# from ci.yml, substitute placeholder values, and validate Ruby syntax.
# Catches invalid DSL before it reaches CI.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CI_YML="${SCRIPT_DIR}/../.github/workflows/ci.yml"

if [ ! -f "$CI_YML" ]; then
  echo "ERROR: ci.yml not found at $CI_YML" >&2
  exit 1
fi

# Extract the formula template between FORMULA heredoc markers,
# remove leading whitespace (matching the sed in ci.yml), and
# substitute shell variables with placeholder values.
FORMULA=$(sed -n '/cat > threedoors-a.rb <<FORMULA/,/^[[:space:]]*FORMULA$/p' "$CI_YML" \
  | tail -n +2 \
  | sed '/^[[:space:]]*FORMULA$/d' \
  | sed 's/^          //')

# Substitute Ruby template interpolations (literal ${...}) with valid placeholders
# shellcheck disable=SC2016 # Single quotes intentional — these are literal template tokens, not shell vars
FORMULA=$(echo "$FORMULA" \
  | sed 's/${VERSION}/0.1.0-alpha.20260101.abcdef0/g' \
  | sed 's/${BASE_URL}/https:\/\/github.com\/arcavenae\/ThreeDoors\/releases\/download\/alpha-20260101-abcdef0/g' \
  | sed 's/${SHA_ARM64}/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/g' \
  | sed 's/${SHA_AMD64}/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/g' \
  | sed 's/${SHA_LINUX}/cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc/g')

TMPFILE=$(mktemp /tmp/threedoors-a-validate.XXXXXX.rb)
echo "$FORMULA" > "$TMPFILE"

# Validate Ruby syntax
if ruby -c "$TMPFILE" 2>&1; then
  echo "Alpha formula template: Ruby syntax OK"
else
  echo "ERROR: Alpha formula template has invalid Ruby syntax" >&2
  echo "Generated formula:"
  cat "$TMPFILE"
  rm -f "$TMPFILE"
  exit 1
fi

# Check for invalid Homebrew DSL patterns
if grep -qE '\bon_arm\b|\bon_intel\b|\bon_linux\b' "$TMPFILE"; then
  echo "ERROR: Formula uses on_arm/on_intel/on_linux blocks (invalid for url/sha256)" >&2
  grep -nE '\bon_arm\b|\bon_intel\b|\bon_linux\b' "$TMPFILE"
  rm -f "$TMPFILE"
  exit 1
fi

echo "Alpha formula template: DSL pattern OK (no on_arm/on_intel/on_linux)"

# Verify the conditional pattern is present
if grep -q 'if OS.mac? && Hardware::CPU.arm?' "$TMPFILE"; then
  echo "Alpha formula template: Correct conditional pattern found"
else
  echo "ERROR: Missing 'if OS.mac? && Hardware::CPU.arm?' conditional" >&2
  rm -f "$TMPFILE"
  exit 1
fi

rm -f "$TMPFILE"
echo "All alpha formula template validations passed."
