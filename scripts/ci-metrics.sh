#!/usr/bin/env bash
set -euo pipefail

# ci-metrics.sh — CI Efficiency Metrics
# Tracks CI runs per merged PR, churn ratio, and ADR-0030 re-entry gate.
#
# Usage:
#   ./scripts/ci-metrics.sh [--days N] [--json] [--repo OWNER/REPO]
#
# Options:
#   --days N        Measurement window in days (default: 7)
#   --json          Output machine-readable JSON line instead of human-readable summary
#   --repo OWNER/REPO  GitHub repository (default: arcaven/ThreeDoors)
#   --help, -h      Show this help message
#
# Requirements: gh CLI, jq
#
# Designed for manual runs or cron. Can be consumed by the retrospector agent
# (Story 51.8) once SLAES is operational.
#
# References:
#   - Baseline: docs/research/ci-churn-reduction-research.md
#   - ADR re-entry gate: docs/ADRs/ADR-0030-ci-churn-reduction.md (Phase 3)
#   - Story: docs/stories/0.37.story.md

REPO="${REPO:-arcaven/ThreeDoors}"
DAYS=7
JSON_OUTPUT=false

# Pre-optimization baseline (from CI churn reduction research, PR #233)
BASELINE_RUNS_PER_PR_LOW=5
BASELINE_RUNS_PER_PR_HIGH=10
BASELINE_CHURN_PCT=60
ADR_REENTRY_THRESHOLD=3

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --days) DAYS="$2"; shift 2 ;;
        --json) JSON_OUTPUT=true; shift ;;
        --repo) REPO="$2"; shift 2 ;;
        --help|-h)
            awk 'BEGIN{h=0} /^#[^!]/{sub(/^# ?/,""); print; h=1; next} h==1 && !/^#/{exit}' "$0"
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Validate days is a positive integer
if ! [[ "$DAYS" =~ ^[0-9]+$ ]] || [[ "$DAYS" -eq 0 ]]; then
    echo "Error: --days must be a positive integer, got '$DAYS'" >&2
    exit 1
fi

# Check for required tools
for cmd in gh jq; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: $cmd is required but not installed." >&2
        exit 1
    fi
done

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Calculate date range
if [[ "$(uname)" == "Darwin" ]]; then
    SINCE="$(date -u -v-"${DAYS}"d +%Y-%m-%dT%H:%M:%SZ)"
else
    SINCE="$(date -u -d "${DAYS} days ago" +%Y-%m-%dT%H:%M:%SZ)"
fi

# --- Fetch CI workflow runs ---
# Get all workflow runs in the time window
RUNS_JSON="$(gh api "repos/${REPO}/actions/runs?created=>=${SINCE}&per_page=100" \
    --jq '.workflow_runs' 2>/dev/null)" || {
    echo "Error: failed to query workflow runs. Check gh auth status and repo access." >&2
    exit 1
}

TOTAL_RUNS="$(echo "$RUNS_JSON" | jq 'length')"

# Count push-to-main runs and their failures
MAIN_PUSH_RUNS="$(echo "$RUNS_JSON" | jq '[.[] | select(.event == "push" and .head_branch == "main")] | length')"
MAIN_PUSH_FAILURES="$(echo "$RUNS_JSON" | jq '[.[] | select(.event == "push" and .head_branch == "main" and .conclusion == "failure")] | length')"

# Count docs-only skipped runs (conclusion == "skipped" or paths-filter cancelled them)
SKIPPED_RUNS="$(echo "$RUNS_JSON" | jq '[.[] | select(.conclusion == "skipped" or .conclusion == "cancelled")] | length')"

# --- Fetch merged PRs ---
MERGED_PRS_JSON="$(gh api "repos/${REPO}/pulls?state=closed&sort=updated&direction=desc&per_page=100" \
    --jq "[.[] | select(.merged_at != null and .merged_at >= \"${SINCE}\")]" 2>/dev/null)" || {
    echo "Error: failed to query merged PRs. Check gh auth status and repo access." >&2
    exit 1
}

MERGED_PRS="$(echo "$MERGED_PRS_JSON" | jq 'length')"

# --- Compute metrics ---
if [[ "$MERGED_PRS" -gt 0 ]]; then
    RUNS_PER_PR="$(echo "scale=2; $TOTAL_RUNS / $MERGED_PRS" | bc)"
else
    RUNS_PER_PR="N/A"
fi

# Baseline comparison
if [[ "$RUNS_PER_PR" != "N/A" ]]; then
    BASELINE_MID="$(echo "scale=1; ($BASELINE_RUNS_PER_PR_LOW + $BASELINE_RUNS_PER_PR_HIGH) / 2" | bc)"
    IMPROVEMENT_PCT="$(echo "scale=1; (1 - $RUNS_PER_PR / $BASELINE_MID) * 100" | bc)"
else
    BASELINE_MID="7.5"
    IMPROVEMENT_PCT="N/A"
fi

# ADR-0030 re-entry gate check
ADR_TRIGGERED=false
if [[ "$MAIN_PUSH_FAILURES" -gt "$ADR_REENTRY_THRESHOLD" ]]; then
    ADR_TRIGGERED=true
fi

# --- Output ---
if [[ "$JSON_OUTPUT" == "true" ]]; then
    jq -n \
        --arg timestamp "$TIMESTAMP" \
        --arg repo "$REPO" \
        --argjson days "$DAYS" \
        --arg since "$SINCE" \
        --argjson total_runs "$TOTAL_RUNS" \
        --argjson merged_prs "$MERGED_PRS" \
        --arg runs_per_pr "$RUNS_PER_PR" \
        --argjson skipped_runs "$SKIPPED_RUNS" \
        --argjson main_push_runs "$MAIN_PUSH_RUNS" \
        --argjson main_push_failures "$MAIN_PUSH_FAILURES" \
        --arg baseline_runs_per_pr "${BASELINE_RUNS_PER_PR_LOW}-${BASELINE_RUNS_PER_PR_HIGH}" \
        --arg improvement_pct "$IMPROVEMENT_PCT" \
        --argjson adr_0030_triggered "$ADR_TRIGGERED" \
        --argjson adr_0030_threshold "$ADR_REENTRY_THRESHOLD" \
        '{
            timestamp: $timestamp,
            repo: $repo,
            period_days: $days,
            since: $since,
            total_ci_runs: $total_runs,
            merged_prs: $merged_prs,
            runs_per_merged_pr: $runs_per_pr,
            skipped_runs: $skipped_runs,
            main_push_runs: $main_push_runs,
            main_push_failures: $main_push_failures,
            baseline_runs_per_pr: $baseline_runs_per_pr,
            improvement_pct: $improvement_pct,
            adr_0030_triggered: $adr_0030_triggered,
            adr_0030_failure_threshold: $adr_0030_threshold
        }'
    exit 0
fi

# Human-readable output
echo "╔══════════════════════════════════════════════════╗"
echo "║           CI Efficiency Metrics Report          ║"
echo "╠══════════════════════════════════════════════════╣"
echo "║  Repository:  ${REPO}"
echo "║  Period:      Last ${DAYS} days (since ${SINCE%%T*})"
echo "║  Generated:   ${TIMESTAMP}"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "── Core Metrics ──────────────────────────────────"
printf "  Total CI runs:           %s\n" "$TOTAL_RUNS"
printf "  Merged PRs:              %s\n" "$MERGED_PRS"
printf "  CI runs per merged PR:   %s\n" "$RUNS_PER_PR"
echo ""
echo "── Path Filter Effectiveness ─────────────────────"
printf "  Skipped/cancelled runs:  %s\n" "$SKIPPED_RUNS"
if [[ "$TOTAL_RUNS" -gt 0 ]]; then
    SKIP_PCT="$(echo "scale=1; $SKIPPED_RUNS * 100 / $TOTAL_RUNS" | bc)"
    printf "  Skip rate:               %s%%\n" "$SKIP_PCT"
fi
echo ""
echo "── Main Branch CI Health ─────────────────────────"
printf "  Push-to-main runs:       %s\n" "$MAIN_PUSH_RUNS"
printf "  Push-to-main failures:   %s\n" "$MAIN_PUSH_FAILURES"
echo ""
echo "── Baseline Comparison ─────────────────────────── "
printf "  Pre-optimization:        %s-%s runs/PR (%s%% churn)\n" \
    "$BASELINE_RUNS_PER_PR_LOW" "$BASELINE_RUNS_PER_PR_HIGH" "$BASELINE_CHURN_PCT"
printf "  Current:                 %s runs/PR\n" "$RUNS_PER_PR"
if [[ "$IMPROVEMENT_PCT" != "N/A" ]]; then
    printf "  Improvement:             %s%%\n" "$IMPROVEMENT_PCT"
fi

# ADR-0030 re-entry gate
echo ""
if [[ "$ADR_TRIGGERED" == "true" ]]; then
    echo "╔══════════════════════════════════════════════════╗"
    echo "║  ⚠  ADR-0030 RE-ENTRY GATE TRIGGERED ⚠         ║"
    echo "║                                                  ║"
    echo "║  Main branch CI failures (${MAIN_PUSH_FAILURES}) exceed threshold (${ADR_REENTRY_THRESHOLD}).  ║"
    echo "║  Consider GitHub Native Merge Queue.             ║"
    echo "║                                                  ║"
    echo "║  See: docs/ADRs/ADR-0030-ci-churn-reduction.md   ║"
    echo "╚══════════════════════════════════════════════════╝"
else
    echo "── ADR-0030 Re-Entry Gate ────────────────────────"
    printf "  Main CI failures:        %s / %s threshold\n" "$MAIN_PUSH_FAILURES" "$ADR_REENTRY_THRESHOLD"
    echo "  Status:                  OK — gate not triggered"
fi
