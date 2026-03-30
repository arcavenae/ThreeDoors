#!/usr/bin/env bash
# validate-dfcp.sh — Validate .dfcp.yaml against CODEOWNERS and CI workflows.
# Checks: YAML validity, CODEOWNERS consistency, autonomy level consistency,
# and CI required checks match actual workflow names.
#
# Usage: scripts/validate-dfcp.sh [path-to-repo-root]
# Exit codes: 0 = pass, 1 = validation failure

set -euo pipefail

REPO_ROOT="${1:-.}"
DFCP_FILE="$REPO_ROOT/.dfcp.yaml"
CODEOWNERS_FILE="$REPO_ROOT/.github/CODEOWNERS"
WORKFLOWS_DIR="$REPO_ROOT/.github/workflows"

errors=0
warnings=0

err() { echo "ERROR: $1" >&2; errors=$((errors + 1)); }
warn() { echo "WARN:  $1" >&2; warnings=$((warnings + 1)); }
info() { echo "INFO:  $1"; }

# --- Pre-flight checks ---

if ! command -v yq &>/dev/null; then
    echo "FATAL: yq is required but not installed. Install with: brew install yq" >&2
    exit 1
fi

if [[ ! -f "$DFCP_FILE" ]]; then
    echo "FATAL: $DFCP_FILE not found" >&2
    exit 1
fi

# --- 1. YAML validity ---

info "Checking YAML validity..."
if ! yq eval '.' "$DFCP_FILE" >/dev/null 2>&1; then
    err ".dfcp.yaml is not valid YAML"
    exit 1
fi
info "  YAML is valid"

# --- 2. Required fields ---

info "Checking required fields..."

version=$(yq eval '.dfcp.version' "$DFCP_FILE")
profile=$(yq eval '.dfcp.profile' "$DFCP_FILE")
autonomy_default=$(yq eval '.dfcp.autonomy.default' "$DFCP_FILE")
autonomy_max=$(yq eval '.dfcp.autonomy.max' "$DFCP_FILE")

if [[ "$version" == "null" ]]; then err "Missing field: dfcp.version"; fi
if [[ "$profile" == "null" ]]; then err "Missing field: dfcp.profile"; fi
if [[ "$autonomy_default" == "null" ]]; then err "Missing field: dfcp.autonomy.default"; fi
if [[ "$autonomy_max" == "null" ]]; then err "Missing field: dfcp.autonomy.max"; fi

# --- 3. Version check ---

if [[ "$version" != "null" && "$version" != "1.0" ]]; then
    err "Unknown schema version: $version (expected 1.0)"
fi

# --- 4. Profile validation ---

info "Checking profile..."
case "$profile" in
    golden|factory|custom) info "  Profile: $profile" ;;
    null) ;; # already reported as missing
    *) err "Invalid profile: $profile (must be golden, factory, or custom)" ;;
esac

# --- 5. Autonomy level consistency ---

info "Checking autonomy levels..."

level_to_int() {
    case "$1" in
        L0) echo 0 ;; L1) echo 1 ;; L2) echo 2 ;;
        L3) echo 3 ;; L4) echo 4 ;; *) echo -1 ;;
    esac
}

if [[ "$autonomy_default" != "null" && "$autonomy_max" != "null" ]]; then
    default_int=$(level_to_int "$autonomy_default")
    max_int=$(level_to_int "$autonomy_max")

    if [[ $default_int -eq -1 ]]; then
        err "Invalid autonomy.default: $autonomy_default (must be L0-L4)"
    fi
    if [[ $max_int -eq -1 ]]; then
        err "Invalid autonomy.max: $autonomy_max (must be L0-L4)"
    fi
    if [[ $default_int -gt $max_int ]]; then
        err "autonomy.default ($autonomy_default) exceeds autonomy.max ($autonomy_max)"
    fi
fi

# Profile-specific autonomy checks
if [[ "$profile" == "golden" && "$autonomy_max" == "L4" ]]; then
    err "Golden profile cannot have autonomy.max = L4 (that's factory-level)"
fi
if [[ "$profile" == "factory" && "$autonomy_max" != "L4" ]]; then
    warn "Factory profile typically has autonomy.max = L4 (currently $autonomy_max)"
fi

# --- 6. CODEOWNERS consistency ---

info "Checking CODEOWNERS consistency..."

if [[ -f "$CODEOWNERS_FILE" ]]; then
    # Extract protected paths from .dfcp.yaml
    dfcp_paths=$(yq eval '.dfcp.governance.require_human_review[]' "$DFCP_FILE" 2>/dev/null || true)

    # Extract paths from CODEOWNERS (skip comments, blank lines; take first field)
    codeowners_paths=$(grep -v '^\s*#' "$CODEOWNERS_FILE" | grep -v '^\s*$' | awk '{print $1}' | sed 's|^/||')

    # Normalize a path: strip leading / and trailing /**, reduce to base dir
    normalize_path() {
        local p="$1"
        p="${p#/}"        # strip leading /
        p="${p%/\*\*}"    # strip trailing /**
        p="${p%\*\*}"     # strip trailing ** (no slash prefix)
        p="${p%/}"        # strip trailing /
        echo "$p"
    }

    # Check each CODEOWNERS path is represented in DFCP
    while IFS= read -r co_path; do
        [[ -z "$co_path" ]] && continue
        co_norm=$(normalize_path "$co_path")
        found=false
        while IFS= read -r dfcp_path; do
            [[ -z "$dfcp_path" ]] && continue
            dfcp_norm=$(normalize_path "$dfcp_path")
            if [[ "$co_norm" == "$dfcp_norm" ]]; then
                found=true
                break
            fi
        done <<< "$dfcp_paths"
        if [[ "$found" == "false" ]]; then
            warn "CODEOWNERS path '$co_path' not found in .dfcp.yaml governance.require_human_review"
        fi
    done <<< "$codeowners_paths"

    # Check each DFCP protected path has a CODEOWNERS entry
    while IFS= read -r dfcp_path; do
        [[ -z "$dfcp_path" ]] && continue
        # .dfcp.yaml itself won't be in CODEOWNERS yet — skip self-reference check
        if [[ "$dfcp_path" == ".dfcp.yaml" ]]; then
            continue
        fi
        dfcp_norm=$(normalize_path "$dfcp_path")
        found=false
        while IFS= read -r co_path; do
            [[ -z "$co_path" ]] && continue
            co_norm=$(normalize_path "$co_path")
            if [[ "$dfcp_norm" == "$co_norm" ]]; then
                found=true
                break
            fi
        done <<< "$codeowners_paths"
        if [[ "$found" == "false" ]]; then
            warn "DFCP protected path '$dfcp_path' has no matching CODEOWNERS entry"
        fi
    done <<< "$dfcp_paths"

    info "  CODEOWNERS cross-check complete"
else
    warn "CODEOWNERS file not found at $CODEOWNERS_FILE — skipping consistency check"
fi

# --- 7. CI required checks ---

info "Checking CI required checks..."

if [[ -d "$WORKFLOWS_DIR" ]]; then
    # Collect all workflow job names
    workflow_names=""
    for wf in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        [[ -f "$wf" ]] || continue
        # Extract job-level name fields (these become GitHub status check names)
        job_names=$(yq eval '.jobs[].name // ""' "$wf" 2>/dev/null || true)
        workflow_names="$workflow_names"$'\n'"$job_names"
    done

    required_checks=$(yq eval '.dfcp.ci.required_checks[]' "$DFCP_FILE" 2>/dev/null || true)
    while IFS= read -r check; do
        [[ -z "$check" ]] && continue
        if ! echo "$workflow_names" | grep -qF "$check"; then
            err "Required CI check '$check' not found in any workflow job name"
        else
            info "  Found: $check"
        fi
    done <<< "$required_checks"
else
    warn "Workflows directory not found at $WORKFLOWS_DIR — skipping CI check validation"
fi

# --- 8. Boolean field validation ---

info "Checking boolean fields..."
for field in \
    ".dfcp.governance.require_story_reference" \
    ".dfcp.governance.require_provenance" \
    ".dfcp.governance.require_signed_commits" \
    ".dfcp.factory.enabled"; do
    val=$(yq eval "$field" "$DFCP_FILE")
    if [[ "$val" != "true" && "$val" != "false" ]]; then
        err "$field must be true or false (got: $val)"
    fi
done

# --- 9. Scope check mode ---

scope_check=$(yq eval '.dfcp.ci.scope_check' "$DFCP_FILE")
case "$scope_check" in
    warn|block|off) ;;
    null) err "Missing field: dfcp.ci.scope_check" ;;
    *) err "Invalid ci.scope_check: $scope_check (must be warn, block, or off)" ;;
esac

# --- Summary ---

echo ""
echo "=== Validation Summary ==="
echo "Errors:   $errors"
echo "Warnings: $warnings"

if [[ $errors -gt 0 ]]; then
    echo "FAIL — fix errors above"
    exit 1
fi

if [[ $warnings -gt 0 ]]; then
    echo "PASS (with warnings)"
else
    echo "PASS"
fi
exit 0
