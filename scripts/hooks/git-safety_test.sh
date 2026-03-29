#!/usr/bin/env bash
#
# git-safety_test.sh — Tests for the git-safety PreToolUse hook
#
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
HOOK="$SCRIPT_DIR/git-safety.sh"
PASS=0
FAIL=0

# Helper: run the hook with a simulated Bash tool input, expect a specific exit code
run_test() {
  local description="$1"
  local command="$2"
  local expected_exit="$3"

  local input
  input=$(jq -n --arg cmd "$command" '{tool_input: {command: $cmd}}')

  local actual_exit=0
  echo "$input" | bash "$HOOK" >/dev/null 2>/dev/null || actual_exit=$?

  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "  PASS: $description"
    ((PASS++))
  else
    echo "  FAIL: $description (expected exit $expected_exit, got $actual_exit)"
    echo "        command: $command"
    ((FAIL++))
  fi
}

echo "=== BLOCKED COMMANDS (expect exit 2) ==="

run_test "git fetch" \
  "git fetch origin main" 2

run_test "git fetch (no args)" \
  "git fetch" 2

run_test "git pull" \
  "git pull origin main" 2

run_test "git pull (no args)" \
  "git pull" 2

run_test "git rebase" \
  "git rebase origin/main" 2

run_test "git rebase interactive" \
  "git rebase -i HEAD~3" 2

run_test "git merge branch" \
  "git merge feature-branch" 2

run_test "git merge origin/main" \
  "git merge origin/main" 2

run_test "unsigned commit --no-gpg-sign" \
  "git commit --no-gpg-sign -m 'test'" 2

run_test "unsigned commit -c gpgsign=false" \
  "git commit -c commit.gpgsign=false -m 'test'" 2

run_test "push to main" \
  "git push origin main" 2

run_test "push to master" \
  "git push origin master" 2

run_test "push HEAD:main" \
  "git push origin HEAD:main" 2

run_test "push HEAD:refs/heads/main" \
  "git push origin HEAD:refs/heads/main" 2

run_test "push HEAD:master" \
  "git push origin HEAD:master" 2

run_test "Co-Authored-By in commit" \
  "git commit -m 'feat: something\n\nCo-Authored-By: AI <ai@example.com>'" 2

run_test "co-authored-by (lowercase)" \
  "git commit -m 'feat: something\n\nco-authored-by: AI <ai@example.com>'" 2

echo ""
echo "=== ALLOWED COMMANDS (expect exit 0) ==="

run_test "git add" \
  "git add ." 0

run_test "git add specific file" \
  "git add internal/tasks/pool.go" 0

run_test "git commit -S (signed)" \
  "git commit -S -m 'feat: add feature (Story 73.3)'" 0

run_test "git commit (default signing via gitconfig)" \
  "git commit -m 'feat: add feature (Story 73.3)'" 0

run_test "git push to feature branch" \
  "git push origin work/eager-rabbit" 0

run_test "git push -u feature branch" \
  "git push -u origin work/eager-rabbit" 0

run_test "git status" \
  "git status" 0

run_test "git log" \
  "git log --oneline -5" 0

run_test "git diff" \
  "git diff HEAD" 0

run_test "git branch" \
  "git branch -a" 0

run_test "git checkout -b new branch" \
  "git checkout -b feature/new-thing" 0

run_test "git merge --abort (recovery)" \
  "git merge --abort" 0

run_test "git merge --continue (recovery)" \
  "git merge --continue" 0

run_test "non-git command" \
  "go test ./..." 0

run_test "non-git command with git in path" \
  "cat .gitignore" 0

run_test "just fmt" \
  "just fmt" 0

run_test "empty command" \
  "" 0

echo ""
echo "=== EDGE CASES ==="

run_test "git fetch in pipe (still blocked)" \
  "git fetch origin main && git log" 2

run_test "git pull in subshell (still blocked)" \
  "echo start && git pull origin main" 2

run_test "git push to main in chain (still blocked)" \
  "git add . && git commit -m 'test' && git push origin main" 2

run_test "git stash (allowed)" \
  "git stash" 0

run_test "git stash pop (allowed)" \
  "git stash pop" 0

run_test "git cherry-pick (allowed)" \
  "git cherry-pick abc123" 0

run_test "git tag (allowed)" \
  "git tag v1.0.0" 0

run_test "git remote -v (allowed)" \
  "git remote -v" 0

run_test "git show (allowed)" \
  "git show HEAD" 0

echo ""
echo "=== RESULTS ==="
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
