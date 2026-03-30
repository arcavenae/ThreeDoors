#!/usr/bin/env bash
#
# git-safety_test.sh — Tests for the git-safety PreToolUse hook
#
# Tests cover two contexts:
#   1. Worker worktree (CLAUDE_PROJECT_DIR=~/.multiclaude/wts/...): full restrictions
#   2. Main checkout (CLAUDE_PROJECT_DIR=/path/to/repo): sync commands allowed
#
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
HOOK="$SCRIPT_DIR/git-safety.sh"
PASS=0
FAIL=0

# Simulated paths for context detection
WORKER_PATH="$HOME/.multiclaude/wts/ThreeDoors/test-worker"
MAIN_PATH="/Users/skippy/repos/ThreeDoors"

# Helper: run the hook with a simulated Bash tool input, expect a specific exit code
# Accepts an optional 4th arg for CLAUDE_PROJECT_DIR override
run_test() {
  local description="$1"
  local command="$2"
  local expected_exit="$3"
  local context_path="${4:-}"

  local input
  input=$(jq -n --arg cmd "$command" '{tool_input: {command: $cmd}}')

  local actual_exit=0
  if [[ -n "$context_path" ]]; then
    echo "$input" | CLAUDE_PROJECT_DIR="$context_path" bash "$HOOK" >/dev/null 2>/dev/null || actual_exit=$?
  else
    echo "$input" | bash "$HOOK" >/dev/null 2>/dev/null || actual_exit=$?
  fi

  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "  PASS: $description"
    ((PASS++))
  else
    echo "  FAIL: $description (expected exit $expected_exit, got $actual_exit)"
    echo "        command: $command"
    ((FAIL++))
  fi
}

echo "========================================"
echo "WORKER WORKTREE CONTEXT"
echo "(CLAUDE_PROJECT_DIR=$WORKER_PATH)"
echo "========================================"

echo ""
echo "=== BLOCKED: Sync commands in worker worktree (expect exit 2) ==="

run_test "git fetch (worker)" \
  "git fetch origin main" 2 "$WORKER_PATH"

run_test "git fetch no args (worker)" \
  "git fetch" 2 "$WORKER_PATH"

run_test "git pull (worker)" \
  "git pull origin main" 2 "$WORKER_PATH"

run_test "git pull no args (worker)" \
  "git pull" 2 "$WORKER_PATH"

run_test "git rebase (worker)" \
  "git rebase origin/main" 2 "$WORKER_PATH"

run_test "git rebase interactive (worker)" \
  "git rebase -i HEAD~3" 2 "$WORKER_PATH"

run_test "git merge branch (worker)" \
  "git merge feature-branch" 2 "$WORKER_PATH"

run_test "git merge origin/main (worker)" \
  "git merge origin/main" 2 "$WORKER_PATH"

echo ""
echo "=== BLOCKED: Universal protections in worker worktree (expect exit 2) ==="

run_test "unsigned commit --no-gpg-sign (worker)" \
  "git commit --no-gpg-sign -m 'test'" 2 "$WORKER_PATH"

run_test "unsigned commit -c gpgsign=false (worker)" \
  "git commit -c commit.gpgsign=false -m 'test'" 2 "$WORKER_PATH"

run_test "push to main (worker)" \
  "git push origin main" 2 "$WORKER_PATH"

run_test "push to master (worker)" \
  "git push origin master" 2 "$WORKER_PATH"

run_test "push HEAD:main (worker)" \
  "git push origin HEAD:main" 2 "$WORKER_PATH"

run_test "push HEAD:refs/heads/main (worker)" \
  "git push origin HEAD:refs/heads/main" 2 "$WORKER_PATH"

run_test "push HEAD:master (worker)" \
  "git push origin HEAD:master" 2 "$WORKER_PATH"

run_test "Co-Authored-By in commit (worker)" \
  "git commit -m 'feat: something\n\nCo-Authored-By: AI <ai@example.com>'" 2 "$WORKER_PATH"

run_test "co-authored-by lowercase (worker)" \
  "git commit -m 'feat: something\n\nco-authored-by: AI <ai@example.com>'" 2 "$WORKER_PATH"

echo ""
echo "=== ALLOWED: Safe commands in worker worktree (expect exit 0) ==="

run_test "git add (worker)" \
  "git add ." 0 "$WORKER_PATH"

run_test "git commit signed (worker)" \
  "git commit -S -m 'feat: add feature (Story 73.8)'" 0 "$WORKER_PATH"

run_test "git commit default signing (worker)" \
  "git commit -m 'feat: add feature (Story 73.8)'" 0 "$WORKER_PATH"

run_test "git push feature branch (worker)" \
  "git push origin work/eager-rabbit" 0 "$WORKER_PATH"

run_test "git push -u feature branch (worker)" \
  "git push -u origin work/eager-rabbit" 0 "$WORKER_PATH"

run_test "git status (worker)" \
  "git status" 0 "$WORKER_PATH"

run_test "git log (worker)" \
  "git log --oneline -5" 0 "$WORKER_PATH"

run_test "git diff (worker)" \
  "git diff HEAD" 0 "$WORKER_PATH"

run_test "git branch (worker)" \
  "git branch -a" 0 "$WORKER_PATH"

run_test "git checkout -b (worker)" \
  "git checkout -b feature/new-thing" 0 "$WORKER_PATH"

run_test "git merge --abort (worker)" \
  "git merge --abort" 0 "$WORKER_PATH"

run_test "git merge --continue (worker)" \
  "git merge --continue" 0 "$WORKER_PATH"

run_test "git stash (worker)" \
  "git stash" 0 "$WORKER_PATH"

run_test "git stash pop (worker)" \
  "git stash pop" 0 "$WORKER_PATH"

echo ""
echo "========================================"
echo "MAIN CHECKOUT CONTEXT"
echo "(CLAUDE_PROJECT_DIR=$MAIN_PATH)"
echo "========================================"

echo ""
echo "=== ALLOWED: Sync commands in main checkout (expect exit 0) ==="

run_test "git fetch (main checkout)" \
  "git fetch origin main" 0 "$MAIN_PATH"

run_test "git fetch no args (main checkout)" \
  "git fetch" 0 "$MAIN_PATH"

run_test "git pull (main checkout)" \
  "git pull origin main" 0 "$MAIN_PATH"

run_test "git pull no args (main checkout)" \
  "git pull" 0 "$MAIN_PATH"

run_test "git rebase (main checkout)" \
  "git rebase origin/main" 0 "$MAIN_PATH"

run_test "git rebase interactive (main checkout)" \
  "git rebase -i HEAD~3" 0 "$MAIN_PATH"

run_test "git merge branch (main checkout)" \
  "git merge feature-branch" 0 "$MAIN_PATH"

run_test "git merge origin/main (main checkout)" \
  "git merge origin/main" 0 "$MAIN_PATH"

echo ""
echo "=== BLOCKED: Universal protections in main checkout (expect exit 2) ==="

run_test "unsigned commit --no-gpg-sign (main checkout)" \
  "git commit --no-gpg-sign -m 'test'" 2 "$MAIN_PATH"

run_test "unsigned commit -c gpgsign=false (main checkout)" \
  "git commit -c commit.gpgsign=false -m 'test'" 2 "$MAIN_PATH"

run_test "push to main (main checkout)" \
  "git push origin main" 2 "$MAIN_PATH"

run_test "push to master (main checkout)" \
  "git push origin master" 2 "$MAIN_PATH"

run_test "push HEAD:main (main checkout)" \
  "git push origin HEAD:main" 2 "$MAIN_PATH"

run_test "push HEAD:refs/heads/main (main checkout)" \
  "git push origin HEAD:refs/heads/main" 2 "$MAIN_PATH"

run_test "Co-Authored-By in commit (main checkout)" \
  "git commit -m 'feat: something\n\nCo-Authored-By: AI <ai@example.com>'" 2 "$MAIN_PATH"

echo ""
echo "=== ALLOWED: Safe commands in main checkout (expect exit 0) ==="

run_test "git add (main checkout)" \
  "git add ." 0 "$MAIN_PATH"

run_test "git commit signed (main checkout)" \
  "git commit -S -m 'feat: add feature'" 0 "$MAIN_PATH"

run_test "git push feature branch (main checkout)" \
  "git push origin work/some-worker" 0 "$MAIN_PATH"

run_test "git status (main checkout)" \
  "git status" 0 "$MAIN_PATH"

run_test "non-git command (main checkout)" \
  "go test ./..." 0 "$MAIN_PATH"

echo ""
echo "========================================"
echo "EDGE CASES"
echo "========================================"

echo ""
echo "=== Worker worktree edge cases ==="

run_test "git fetch in pipe (worker, blocked)" \
  "git fetch origin main && git log" 2 "$WORKER_PATH"

run_test "git pull in subshell (worker, blocked)" \
  "echo start && git pull origin main" 2 "$WORKER_PATH"

run_test "git push to main in chain (worker, blocked)" \
  "git add . && git commit -m 'test' && git push origin main" 2 "$WORKER_PATH"

run_test "git cherry-pick (worker, allowed)" \
  "git cherry-pick abc123" 0 "$WORKER_PATH"

run_test "git tag (worker, allowed)" \
  "git tag v1.0.0" 0 "$WORKER_PATH"

run_test "git remote -v (worker, allowed)" \
  "git remote -v" 0 "$WORKER_PATH"

run_test "git show (worker, allowed)" \
  "git show HEAD" 0 "$WORKER_PATH"

echo ""
echo "=== Main checkout edge cases ==="

run_test "git fetch in pipe (main, allowed)" \
  "git fetch origin main && git log" 0 "$MAIN_PATH"

run_test "git rebase in chain (main, allowed)" \
  "git fetch origin main && git rebase origin/main" 0 "$MAIN_PATH"

run_test "git push to main in chain (main, blocked)" \
  "git add . && git commit -m 'test' && git push origin main" 2 "$MAIN_PATH"

echo ""
echo "=== Context detection edge cases ==="

run_test "nested wts path (worker context)" \
  "git fetch" 2 "$HOME/.multiclaude/wts/SomeRepo/busy-fox"

run_test "path containing wts but not multiclaude (not worker)" \
  "git fetch" 0 "/tmp/wts/something"

run_test "non-git command (worker, allowed)" \
  "go test ./..." 0 "$WORKER_PATH"

run_test "non-git command (main, allowed)" \
  "just fmt" 0 "$MAIN_PATH"

run_test "empty command (worker)" \
  "" 0 "$WORKER_PATH"

run_test "empty command (main)" \
  "" 0 "$MAIN_PATH"

run_test "cat .gitignore (worker, allowed)" \
  "cat .gitignore" 0 "$WORKER_PATH"

echo ""
echo "========================================"
echo "RESULTS"
echo "========================================"
echo "Passed: $PASS"
echo "Failed: $FAIL"
TOTAL=$((PASS + FAIL))
echo "Total:  $TOTAL"

if [[ "$FAIL" -gt 0 ]]; then
  echo ""
  echo "SOME TESTS FAILED"
  exit 1
else
  echo ""
  echo "ALL TESTS PASSED"
  exit 0
fi
