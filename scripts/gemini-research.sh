#!/usr/bin/env bash
# scripts/gemini-research.sh — Wrapper for Gemini CLI research queries
#
# Usage:
#   ./scripts/gemini-research.sh --depth quick --query "What is the state of Go TUI frameworks?"
#   ./scripts/gemini-research.sh --depth deep --query "Architecture analysis of Bubbletea apps"
#   ./scripts/gemini-research.sh --depth standard --query "Compare task management approaches"
#
# Requires:
#   - Gemini CLI authenticated via OAuth (run `gemini` once interactively first)
#   - jq for JSON parsing (brew install jq)
#
# Depth levels:
#   quick    → gemini-2.5-flash (fast, 1000 req/day free tier)
#   standard → gemini-2.5-pro   (balanced, 50 req/day free tier)
#   deep     → gemini-2.5-pro   (thorough, 50 req/day free tier)

set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────
GEMINI_CLI_VERSION="0.32.1"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPORTS_DIR="${PROJECT_ROOT}/_bmad-output/research-reports"
BUDGET_FILE="${REPORTS_DIR}/budget.json"

# ─── Dependency checks ─────────────────────────────────────────────────────
check_dependencies() {
    if ! command -v npx &>/dev/null; then
        echo "Error: npx is required but not found. Install Node.js: brew install node" >&2
        exit 1
    fi

    if ! command -v jq &>/dev/null; then
        echo "Error: jq is required but not found. Install with: brew install jq" >&2
        exit 1
    fi
}

# ─── Argument parsing ──────────────────────────────────────────────────────
DEPTH=""
QUERY=""

usage() {
    echo "Usage: $0 --depth <quick|standard|deep> --query \"<research question>\""
    echo ""
    echo "Options:"
    echo "  --depth    Research depth: quick (Flash), standard (Pro), deep (Pro)"
    echo "  --query    The research question to investigate"
    echo "  --help     Show this help message"
    exit 1
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --depth)
                DEPTH="$2"
                shift 2
                ;;
            --query)
                QUERY="$2"
                shift 2
                ;;
            --help|-h)
                usage
                ;;
            *)
                echo "Error: Unknown argument: $1" >&2
                usage
                ;;
        esac
    done

    if [[ -z "$DEPTH" ]]; then
        echo "Error: --depth is required" >&2
        usage
    fi

    if [[ -z "$QUERY" ]]; then
        echo "Error: --query is required" >&2
        usage
    fi

    case "$DEPTH" in
        quick|standard|deep) ;;
        *)
            echo "Error: --depth must be one of: quick, standard, deep" >&2
            usage
            ;;
    esac
}

# ─── Model selection ────────────────────────────────────────────────────────
select_model() {
    case "$DEPTH" in
        quick)    echo "gemini-2.5-flash" ;;
        standard) echo "gemini-2.5-pro" ;;
        deep)     echo "gemini-2.5-pro" ;;
    esac
}

# ─── Output directory ──────────────────────────────────────────────────────
create_output_dir() {
    local timestamp
    timestamp="$(date -u +%Y%m%d-%H%M%S)"
    local slug
    slug="$(echo "$QUERY" | tr '[:upper:]' '[:lower:]' | tr -cs '[:alnum:]' '-' | head -c 40 | sed 's/-$//')"
    local output_dir="${REPORTS_DIR}/${timestamp}-${slug}"
    mkdir -p "$output_dir"
    echo "$output_dir"
}

# ─── Budget tracking ───────────────────────────────────────────────────────
init_budget() {
    if [[ ! -f "$BUDGET_FILE" ]]; then
        cat > "$BUDGET_FILE" <<'BUDGET_JSON'
{
  "daily_limits": {
    "pro": 50,
    "flash": 1000
  },
  "usage": []
}
BUDGET_JSON
    fi
}

record_usage() {
    local model="$1"
    local depth="$2"
    local status="$3"
    local today
    today="$(date -u +%Y-%m-%d)"

    local tier="flash"
    if [[ "$model" == *"pro"* ]]; then
        tier="pro"
    fi

    local entry
    entry=$(jq -n \
        --arg date "$today" \
        --arg tier "$tier" \
        --arg depth "$depth" \
        --arg status "$status" \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        '{date: $date, tier: $tier, depth: $depth, status: $status, timestamp: $timestamp}')

    local tmp_file="${BUDGET_FILE}.tmp"
    jq --argjson entry "$entry" '.usage += [$entry]' "$BUDGET_FILE" > "$tmp_file"
    mv "$tmp_file" "$BUDGET_FILE"
}

check_budget() {
    local model="$1"
    local today
    today="$(date -u +%Y-%m-%d)"

    local tier="flash"
    local limit=1000
    if [[ "$model" == *"pro"* ]]; then
        tier="pro"
        limit=50
    fi

    local used
    used=$(jq --arg date "$today" --arg tier "$tier" \
        '[.usage[] | select(.date == $date and .tier == $tier and .status == "success")] | length' \
        "$BUDGET_FILE")

    if [[ "$used" -ge "$limit" ]]; then
        echo "Error: Daily $tier limit reached ($used/$limit). Try again tomorrow or use a different depth." >&2
        exit 1
    fi

    local threshold=$(( limit * 80 / 100 ))
    if [[ "$used" -ge "$threshold" ]]; then
        echo "Warning: $tier usage at ${used}/${limit} (80% threshold reached)" >&2
    fi
}

# ─── Main execution ────────────────────────────────────────────────────────
main() {
    parse_args "$@"
    check_dependencies

    local model
    model="$(select_model)"

    init_budget
    check_budget "$model"

    local output_dir
    output_dir="$(create_output_dir)"

    echo "Research query: $QUERY" >&2
    echo "Model: $model (depth: $DEPTH)" >&2
    echo "Output: $output_dir" >&2

    # Invoke Gemini CLI in headless mode with pinned version
    local exit_code=0
    npx "@google/gemini-cli@${GEMINI_CLI_VERSION}" \
        -m "$model" \
        -p "$QUERY" \
        --output-format json \
        > "$output_dir/response.json" 2>"$output_dir/stderr.log" || exit_code=$?

    if [[ "$exit_code" -ne 0 ]]; then
        # Check for auth failure
        if grep -qi "auth\|oauth\|sign.in\|credentials\|unauthorized\|401" "$output_dir/stderr.log" 2>/dev/null; then
            echo "Error: Authentication failed. Re-authenticate by running: gemini" >&2
            echo "This opens a browser for Google Account OAuth sign-in." >&2
        else
            echo "Error: Gemini CLI exited with code $exit_code" >&2
            echo "See: $output_dir/stderr.log" >&2
        fi
        record_usage "$model" "$DEPTH" "failure"
        exit "$exit_code"
    fi

    # Extract response text into report
    if jq -e '.response' "$output_dir/response.json" &>/dev/null; then
        jq -r '.response' "$output_dir/response.json" > "$output_dir/report.md"
    else
        echo "Warning: No .response field in output. Raw JSON saved." >&2
        cp "$output_dir/response.json" "$output_dir/report.md"
    fi

    record_usage "$model" "$DEPTH" "success"

    echo "Research complete: $output_dir/report.md" >&2
    # Output the report to stdout for callers
    cat "$output_dir/report.md"
}

main "$@"
