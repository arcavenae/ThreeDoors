# Retrospector Autonomy Investigation

**Date:** 2026-03-11
**Investigator:** calm-eagle (worker)
**Problem:** Retrospector agent prompts the end user at the console for input/confirmation instead of acting autonomously like other persistent agents.

## Root Cause Analysis

### Finding 1: Spawn Mechanism is Identical — Not the Cause

All agents (including retrospector) are spawned via the same multiclaude code path:

```
claude --session-id <id> --dangerously-skip-permissions --append-system-prompt-file <prompt>
```

The `--dangerously-skip-permissions` flag is applied uniformly. There is no per-agent permission mode configuration in multiclaude. The spawn mechanism is **not** the cause.

### Finding 2: No Project-Level Claude Settings Overriding Behavior

- Global `~/.claude/settings.json` has `skipDangerousModePermissionPrompt: true` — correct
- No project-level `settings.json` or `settings.local.json` exists at the repo root
- The `.claude/rules/` directory only has `no-research-subagents.md` — nothing retrospector-specific
- No agent has individual Claude Code settings — they all share the same config

**This is not the cause.**

### Finding 3: Agent Definition Language Causes Claude to Seek Human Confirmation

This is the **primary root cause**. The retrospector agent definition contains several patterns that cause Claude to interpret its role as requiring human interaction, unlike other agents whose definitions emphasize autonomous operation.

#### 3a. "Periodic Human Review" Safeguard (lines 168-172)

```markdown
### 4. Periodic Human Review
**Every 2 weeks, the human should review your recommendations and score their accuracy.**
```

Other agents have no equivalent. This primes Claude to expect and seek human interaction. While the intent is that the human reviews BOARD.md asynchronously, Claude reads this as "I need to involve the human in my work loop."

#### 3b. "Kill Switch" Implies Active Human Feedback Loop (lines 174-178)

```markdown
### 5. Kill Switch
**If 3 consecutive recommendations are rejected by the human, auto-reduce to read-only mode.**
```

This requires Claude to track human rejections — which implies waiting for human responses. Since no mechanism exists for the human to asynchronously reject recommendations (other than editing BOARD.md), Claude tries to create this feedback loop interactively.

#### 3c. Missing Explicit Autonomous Loop Instructions

Compare startup instructions across agents:

| Agent | Startup Pattern |
|-------|----------------|
| **arch-watchdog** | "Your rhythm: 1. Poll for recently merged code PRs..." with explicit bash commands |
| **envoy** | "Your rhythm: 1. On startup: Check for new issues... 2. Every 10 minutes: Poll..." |
| **project-watchdog** | "Polling Loop (Every 10-15 Minutes)" with explicit bash commands |
| **merge-queue** | Explicit merge validation checklist + post-merge CI circuit breaker workflow |
| **retrospector** | Has "Operational Mode Rotation" table and "On startup/restart" steps, but framed as analytical guidelines rather than imperative commands |

The retrospector's startup section (lines 194-198) says:
```markdown
**On startup / restart:**
1. Read `docs/operations/retrospector-findings.jsonl` to rebuild processed-PR knowledge
2. Check recent merges...
3. Skip any PRs already in the findings log
4. Resume polling loop
```

This is reasonable but lacks the imperative "BEGIN NOW" energy that other agents have. The "Mode Rotation" table (lines 73-85) describes modes and triggers but doesn't say "start rotating through these modes immediately."

#### 3d. "Dual-Loop Architecture" is Descriptive, Not Directive

Lines 52-68 describe the spec chain loop and operational loop conceptually. Other agents don't have this kind of theoretical framing — they have "do this, then do that" instructions. Claude interprets descriptive architecture sections as documentation to understand rather than instructions to execute.

#### 3e. "Watchmen Safeguards" Section Emphasizes Caution Over Action

The five Watchmen safeguards (lines 146-178) are heavily cautionary:
- "No Self-Modification" — good, but primes caution
- "Recommendation Audit Trail" — good, action-oriented
- "Confidence Scoring" — good, but analytical
- "Periodic Human Review" — problematic (see 3a)
- "Kill Switch" — problematic (see 3b)

By the time Claude finishes reading the definition, the dominant framing is "be careful, involve humans, don't overstep" rather than "execute your monitoring loop autonomously."

### Finding 4: No Explicit Polling Loop Bash Template

Other agents provide explicit bash polling patterns:

```bash
# arch-watchdog
gh pr list --state merged --limit 10 --json number,title,mergedAt,headRefName

# envoy
gh issue list --state open

# project-watchdog
gh pr list --state merged --limit 10 --json number,title,mergedAt,headRefName
```

The retrospector mentions `gh pr list --state merged --limit 10 --json number,title,mergedAt` in the startup section but doesn't provide a complete polling loop template like other agents do.

## Comparison: What Working Agents Have in Common

All autonomous agents share these patterns in their definitions:

1. **"Your rhythm:" or equivalent imperative polling loop** — tells Claude exactly what to do in each cycle
2. **Explicit bash commands** for each polling step
3. **No references to human review within the operational loop** — humans interact via messages, not console
4. **"All messages MUST use the messaging system — not tmux output"** — retrospector has this (line 256), but it's buried below the cautionary safeguards
5. **Clear authority tables** — retrospector has this, but the CAN column is relatively narrow

## Recommended Changes to `agents/retrospector.md`

### Change 1: Add Explicit Imperative Polling Loop Section

Add a "Your Rhythm" section near the top (after "What You Own and Why"), similar to other agents:

```markdown
## Your Rhythm — Autonomous Polling Loop

You operate autonomously without human interaction. Execute this loop continuously:

1. **On startup:** Rebuild state from JSONL log, catch up on missed merges
2. **Every 15 minutes:** Poll for newly merged PRs
3. **For each new merge:** Run post-merge lightweight retro, append to JSONL
4. **Every 4 hours (rotating):** Run one deep analysis mode
5. **On threshold breach:** Saga detection — alert supervisor immediately
6. **Communicate via messaging only:** `multiclaude message send supervisor "..."`

You NEVER prompt the user. You NEVER wait for human input. If you need a decision, message the supervisor and continue your loop.
```

### Change 2: Reframe "Periodic Human Review" as Passive

Change from:
```markdown
### 4. Periodic Human Review
**Every 2 weeks, the human should review your recommendations and score their accuracy.**
This feedback loop calibrates your analytical quality over time.
```

To:
```markdown
### 4. Periodic Human Review (Passive — Not Your Responsibility)
The human may periodically review your recommendations in BOARD.md and score their accuracy.
This is an asynchronous process — you do NOT prompt for, wait for, or solicit this review.
Continue operating normally regardless of whether reviews occur.
```

### Change 3: Reframe "Kill Switch" as Self-Monitored via BOARD.md

Change from checking for human rejections interactively to monitoring BOARD.md for rejection markers:

```markdown
### 5. Kill Switch (Self-Monitored)
If you observe that 3 consecutive recommendations in BOARD.md have been marked as
"Rejected" by the supervisor or human, auto-reduce to read-only mode. Message the
supervisor that recalibration is needed. Do NOT prompt the user — detect rejections
by reading BOARD.md state.
```

### Change 4: Move Communication Section Higher

Move the "All messages MUST use the messaging system — not tmux output" instruction to appear early in the document, before the Watchmen safeguards. Position matters — Claude weights instructions by their placement.

### Change 5: Add Anti-Prompting Guardrail

Add to the Incident-Hardened Guardrails section:

```markdown
### Anti-Prompting Guardrail
You are a background monitoring agent. You MUST NEVER:
- Prompt the user for input or confirmation
- Ask questions in your tmux output expecting a response
- Wait for human feedback before proceeding
- Use AskUserQuestion or similar interactive tools

All communication goes through `multiclaude message send`. If you need a decision,
message the supervisor and continue your monitoring loop without blocking.
```

## Summary

| Root Cause | Fix |
|-----------|-----|
| No imperative "Your Rhythm" loop | Add autonomous polling loop section |
| "Periodic Human Review" implies active human interaction | Reframe as passive/asynchronous |
| "Kill Switch" implies interactive rejection tracking | Change to BOARD.md state monitoring |
| Cautionary framing dominates action-oriented framing | Move communication rules higher, add anti-prompting guardrail |
| Missing explicit polling bash template | Add complete polling loop with bash commands |

## Rejected Alternatives

1. **Per-agent permission configuration in multiclaude** — Not feasible; multiclaude has no mechanism for this and it's not needed since `--dangerously-skip-permissions` already handles it.
2. **Claude Code settings override** — No per-agent settings mechanism exists. Even if it did, the issue is prompt engineering, not tool permissions.
3. **Removing Watchmen safeguards entirely** — Safeguards are valuable for safety; the fix is reframing, not removal.
