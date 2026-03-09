# Research: Automated Task Pipeline for Multiclaude

**Date:** 2026-03-04
**Status:** Proposal

## 1. Current Multiclaude Capabilities

Multiclaude (v0.0.0-dev) is a repo-centric orchestrator managing Claude Code instances via tmux sessions. For ThreeDoors it runs:

- **Supervisor agent** — persistent, coordinates work, receives messages from workers
- **PR Shepherd agent** — persistent, keeps fork PRs rebased/green against upstream
- **Worker agents** — ephemeral, spawned per-task, create PRs, then signal completion via `multiclaude agent complete`

**Worker lifecycle (the key primitive):**

```
multiclaude worker create "<task description>" [--repo <repo>] [--branch <branch>]
```

Creates a git worktree, opens a tmux window, injects the worker prompt + task description. Worker receives full `CLAUDE.md` context, fork workflow instructions, and CLI reference. On completion, runs `multiclaude agent complete [--summary <text>]`.

**Inter-agent messaging:**

```
multiclaude message send <recipient> "<message>"
multiclaude message list
multiclaude message ack <message-id>
```

Messages stored in `~/.multiclaude/messages/ThreeDoors/<agent>/` and polled by recipients.

**What does NOT exist today:**

- No scheduling or cron capability
- No batch/pipeline dispatch command
- No story-file parser or backlog reader
- No automatic chaining (one task finishing does not trigger the next)
- No concurrency limits or cost controls built in
- No approval gates or human-in-the-loop hooks

The system is entirely **imperative** — a human or supervisor agent must explicitly call `multiclaude worker create` for each task.

---

## 2. What Makes a Task Suitable for Automation

**High Automation Suitability:**

- **Structured acceptance criteria** — Story files with numbered ACs (AC1-AC7) that are testable assertions
- **Clear file scope** — Stories specifying exactly which files/packages to create or modify
- **Quality gates defined** — Stories with explicit `make fmt`, `make lint`, `make test` checklists
- **Existing patterns to follow** — When the codebase already has similar implementations
- **Independent stories** — No cross-story data dependencies beyond declared "Blocks"/"Dependencies"
- **Test-first stories** — Stories enriched via the BMAD `tea` role with test specifications

**Low Automation Suitability (requires human judgment):**

- Ambiguous requirements without testable ACs
- Architecture decisions requiring approach selection
- Cross-cutting refactors touching 10+ files with subtle interdependencies
- External integration setup requiring API keys, OAuth flows, etc.
- Stories with partial in-progress work

**Scoring rubric for automation readiness:**

| Factor | Weight | Automatable if... |
|--------|--------|-------------------|
| Acceptance criteria count | 30% | 3+ numbered ACs |
| File scope specified | 20% | Story names specific files/packages |
| Quality gates listed | 15% | Has Pre-PR Checklist section |
| Dependencies resolved | 15% | All "Blocks" stories are `done` |
| Status = `draft` | 10% | Not `in-progress` (avoids conflicts) |
| Test patterns exist | 10% | Similar `*_test.go` files already in codebase |

---

## 3. Architecture Options

### Option A: Shell Script Scheduler (Lowest Effort)

A bash script that reads story files, filters for ready ones, and dispatches workers.

**Pros:** Works today with zero multiclaude changes. Simple to understand and debug.
**Cons:** No event-driven chaining. Crude dependency checking. No cost controls. Requires external cron.

### Option B: Event-Driven Pipeline via GitHub Actions

Use GitHub Actions to trigger worker dispatch when PRs merge.

**Pros:** Naturally chains — merge triggers next task. Leverages existing CI infrastructure.
**Cons:** Requires a self-hosted runner on the multiclaude machine. GHA is not designed for local process orchestration.

### Option C: Supervisor Agent Enhancement (Most Natural)

Extend the supervisor's prompt to auto-dispatch from a backlog. The supervisor already receives completion notifications.

**Pros:** Uses existing infrastructure. Supervisor already has the coordination role. No new commands needed.
**Cons:** Relies on supervisor's context window and reliability. No hard concurrency enforcement.

### Option D: New `multiclaude pipeline` Command (Most Robust)

A first-class pipeline command:

```
multiclaude pipeline run \
  --stories "docs/stories/17.*.story.md" \
  --max-workers 3 \
  --require-ci \
  --approval-gate pr-ready \
  --cost-limit 50.00
```

**Pros:** Full control. Proper dependency resolution. Built-in guardrails.
**Cons:** Requires multiclaude Go codebase changes. Significant development effort.

---

## 4. Safety Guardrails Required

**Concurrency Limits:**
- Hard cap on concurrent workers (recommend 3-5)
- Tracked in `state.json` — daemon already tracks active agents

**CI Gate:**
- Worker PR must pass CI before next dependent story dispatches
- Stories not marked "done" until PR is merged (not just created)

**Human Approval Gates:**
- **Pre-dispatch:** Optionally require `Automation: yes` field in story file
- **Pre-merge:** Fork mode means upstream maintainers provide the human gate
- **Emergency stop:** `multiclaude stop-all` halts everything

**Cost/Token Controls:**
- Wall-clock timeout per worker (e.g., 30 minutes)
- Maximum stories-per-run limit

**Rollback:**
- Each worker creates an isolated branch + PR
- Close PR and re-dispatch to retry
- Story status reverted from `in-progress` to `draft` to re-queue

---

## 5. Recommended Implementation Path

### Phase 1: Shell Script MVP (works today)

Build `scripts/pipeline.sh` that parses story files, checks dependencies, dispatches workers. Run manually or via `launchd`.

### Phase 2: Supervisor Enhancement (1-2 days)

Add backlog-awareness to supervisor prompt. On worker completion + CI pass, scan stories and dispatch next ready one.

### Phase 3: Story File Standardization (1 day)

Add `**Automation: yes|no**` field to story files for explicit human control over which stories auto-execute.

### Phase 4: `multiclaude pipeline` command (future)

Proper Go subcommand with DAG dependency resolution, cost limits, and concurrency controls. Shell script from Phase 1 serves as the spec.

**Estimated effort:** 2-3 days for Phases 1-3. Phase 4 is a multiclaude feature request.
