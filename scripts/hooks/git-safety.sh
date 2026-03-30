#!/usr/bin/env bash
#
# git-safety.sh — PreToolUse hook for Claude Code
#
# Scoped protections (worker worktrees only — Story 73.8):
#   - git fetch, git pull, git rebase, git merge (INC-002 prevention)
#
# Universal protections (all contexts):
#   - Unsigned commits (--no-gpg-sign, -c commit.gpgsign=false)
#   - Direct push to main/master
#   - Co-Authored-By trailers in commit messages
#
# Worktree detection: commands run from inside ~/.multiclaude/wts/ are
# treated as worker context. All other paths (main checkout, persistent
# agents, multiclaude CLI) are exempt from sync restrictions.
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

# --- Worktree detection (Story 73.8) ---
# Determine if we're in a multiclaude worker worktree.
# Workers run from ~/.multiclaude/wts/<repo>/<worker-name>/.
# Persistent agents, supervisor, and multiclaude CLI run from the main checkout.
# Use CLAUDE_PROJECT_DIR if set (Claude Code provides it), fall back to PWD.
HOOK_CWD="${CLAUDE_PROJECT_DIR:-$PWD}"
IS_WORKER_WORKTREE=false
if [[ "$HOOK_CWD" == */.multiclaude/wts/* ]]; then
  IS_WORKER_WORKTREE=true
fi

# --- Scoped blocks: git fetch, git pull, git rebase, git merge (INC-002) ---
# Only enforced in worker worktrees. Persistent agents and multiclaude CLI
# need these operations for branch management and worktree creation.
if [[ "$IS_WORKER_WORKTREE" == true ]]; then
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
