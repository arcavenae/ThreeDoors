#!/usr/bin/env bash
#
# git-safety.sh — PreToolUse hook for Claude Code
#
# Blocks dangerous git commands in worker worktrees:
#   - git fetch, git pull, git rebase, git merge (INC-002 prevention)
#   - Unsigned commits (--no-gpg-sign, -c commit.gpgsign=false)
#   - Direct push to main/master
#   - Co-Authored-By trailers in commit messages
#
# Exit codes:
#   0 — allow the command
#   2 — block the command (stderr sent to Claude as error)
#
# Input: JSON on stdin with tool_input.command field
#
set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Extract the command string; exit 0 (allow) if not a Bash tool call or no command
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null) || exit 0
if [[ -z "$COMMAND" ]]; then
  exit 0
fi

# Only inspect commands that contain "git" — fast path for non-git commands
if [[ "$COMMAND" != *"git "* && "$COMMAND" != *"git"$'\t'* ]]; then
  exit 0
fi

# --- Blocked: git fetch, git pull, git rebase, git merge (INC-002) ---
# Match git followed by the dangerous subcommand, allowing flags between them.
# This catches: git fetch, git -C /path fetch, git fetch origin main, etc.
if echo "$COMMAND" | grep -qE '\bgit\b\s+(-[a-zA-Z]\s+\S+\s+)*\b(fetch|pull)\b'; then
  echo "BLOCKED: 'git fetch' and 'git pull' are forbidden in worker worktrees." >&2
  echo "Your worktree is managed by multiclaude and auto-synced. See INC-002." >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE '\bgit\b\s+(-[a-zA-Z]\s+\S+\s+)*\brebase\b'; then
  echo "BLOCKED: 'git rebase' is forbidden in worker worktrees." >&2
  echo "Your worktree is managed by multiclaude and auto-synced. See INC-002." >&2
  exit 2
fi

# Block git merge with remote refs or branch names (but allow git merge --abort)
if echo "$COMMAND" | grep -qE '\bgit\b\s+(-[a-zA-Z]\s+\S+\s+)*\bmerge\b'; then
  # Allow merge --abort and merge --continue (recovery commands)
  if echo "$COMMAND" | grep -qE '\bmerge\b\s+--(abort|continue)'; then
    : # allowed
  else
    echo "BLOCKED: 'git merge' is forbidden in worker worktrees." >&2
    echo "Your worktree is managed by multiclaude and auto-synced. See INC-002." >&2
    exit 2
  fi
fi

# --- Blocked: unsigned commits ---
if echo "$COMMAND" | grep -qE '\bgit\b.*\bcommit\b'; then
  if echo "$COMMAND" | grep -qE '\-\-no-gpg-sign'; then
    echo "BLOCKED: --no-gpg-sign is forbidden. All commits must be signed." >&2
    exit 2
  fi
  if echo "$COMMAND" | grep -qE '\-c\s+commit\.gpgsign=false'; then
    echo "BLOCKED: Disabling commit signing is forbidden. All commits must be signed." >&2
    exit 2
  fi

  # --- Blocked: Co-Authored-By trailers ---
  if echo "$COMMAND" | grep -qiE 'Co-Authored-By'; then
    echo "BLOCKED: Co-Authored-By trailers are forbidden per project policy." >&2
    exit 2
  fi
fi

# --- Blocked: push to main/master ---
if echo "$COMMAND" | grep -qE '\bgit\b\s+(-[a-zA-Z]\s+\S+\s+)*\bpush\b'; then
  # Block: git push origin main, git push origin master
  if echo "$COMMAND" | grep -qE '\bpush\b\s+\S+\s+(main|master)\b'; then
    echo "BLOCKED: Direct push to main/master is forbidden. Use feature branches." >&2
    exit 2
  fi
  # Block: git push origin HEAD:main, git push origin HEAD:refs/heads/main
  if echo "$COMMAND" | grep -qE '\bpush\b.*\bHEAD:(refs/heads/)?(main|master)\b'; then
    echo "BLOCKED: Direct push to main/master is forbidden. Use feature branches." >&2
    exit 2
  fi
fi

# All checks passed — allow
exit 0
