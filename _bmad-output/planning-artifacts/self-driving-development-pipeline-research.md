# Self-Driving Development Pipeline: Research Document

## 1. Executive Summary

This research explores a "meta" feature where ThreeDoors tasks directly trigger development work on the ThreeDoors codebase via multiclaude. The user marks a task in the TUI, that task flows into a queue, multiclaude consumes it and spawns a worker agent, and results (PRs, CI status) flow back into ThreeDoors as task updates. This creates a closed loop: the app manages its own development.

The recommended approach is **Option B: TUI-native dispatch with file-based queue**, which builds on existing infrastructure (the `AgentService` in `internal/intelligence/`, the multiclaude `worker create` CLI, and the YAML task file format) with minimal new moving parts.

---

## 2. Task Designation Mechanism

### Options Evaluated

| Mechanism | Description | Pros | Cons |
|-----------|-------------|------|------|
| **Tag prefix** (`#dev`, `@implement`) | User adds a tag to the task text | Zero schema change, works with current text format | Fragile parsing, pollutes task text, no validation |
| **New TaskStatus** (e.g., `dispatched`) | Add to the state machine | Type-safe, fits existing transition validation | Status conflation: "dispatched" is an action, not a state; tasks can be dispatched from `todo` or `in-progress` |
| **New Task field** (`DevQueue bool`) | Boolean field on `Task` struct | Clean separation of dispatch state from lifecycle state | Adds a field that most tasks never use |
| **Dedicated key binding** ('x' in TUI) | TUI action that marks + enqueues | Explicit user intent, discoverable | Requires TUI changes |
| **Command palette** (`:dispatch`) | Colon command in TUI | Consistent with existing `:add`, `:edit`, `:mood` commands | Slightly more keystrokes |

### Recommendation: Compound approach -- new field + key binding + command

Add a `DevDispatch` metadata struct to the `Task` type:

```go
type DevDispatch struct {
    Queued      bool       `yaml:"queued,omitempty" json:"queued,omitempty"`
    QueuedAt    *time.Time `yaml:"queued_at,omitempty" json:"queued_at,omitempty"`
    WorkerName  string     `yaml:"worker_name,omitempty" json:"worker_name,omitempty"`
    PRNumber    int        `yaml:"pr_number,omitempty" json:"pr_number,omitempty"`
    PRStatus    string     `yaml:"pr_status,omitempty" json:"pr_status,omitempty"`
    DispatchErr string     `yaml:"dispatch_err,omitempty" json:"dispatch_err,omitempty"`
}
```

This keeps dispatch orthogonal to the task lifecycle. A task can be `in-progress` AND dispatched, or `todo` AND dispatched. The TUI gets:
- **'x' key** in detail view: dispatch selected task to dev queue
- **`:dispatch`** command: dispatch from command palette
- Visual badge on doors view showing dispatched tasks (e.g., a gear icon or `[DEV]` tag)

---

## 3. Idea Queue Design

### Data Flow

```
+--------------------+     +------------------+     +-------------------+
|   ThreeDoors TUI   |     |   Dev Queue      |     |   multiclaude     |
|                    |     |   (YAML file)    |     |   daemon          |
|  User presses 'x' +---->+ ~/.threedoors/   +---->+ worker create     |
|  on a task         |     |  dev-queue.yaml  |     |  "<task desc>"    |
|                    |     |                  |     |                   |
|  Task updated with |<----+ Queue item       |<----+ Worker completes  |
|  PR link, status   |     |  status updated  |     |  PR created       |
+--------------------+     +------------------+     +-------------------+
```

### Queue File Format

Location: `~/.threedoors/dev-queue.yaml`

```yaml
queue:
  - id: "dq-a855835c"
    task_id: "a855835c-88f0-470b-9a23-2c5a851521fa"
    task_text: "Add keyboard shortcut help overlay"
    context: "Users need to see available key bindings without leaving the TUI"
    status: "pending"           # pending | dispatched | completed | failed
    priority: 1                 # 1=high, 2=medium, 3=low
    scope: "internal/tui/"      # constrains agent to relevant directories
    acceptance_criteria:
      - "Help overlay appears on '?' keypress"
      - "Shows all key bindings grouped by context"
    queued_at: "2026-03-04T16:00:00Z"
    dispatched_at: null
    completed_at: null
    worker_name: ""
    pr_number: 0
    pr_url: ""
    error: ""
    
settings:
  max_concurrent: 2            # never more than 2 workers at once
  auto_dispatch: false         # require manual approval before spawning
  require_story: true          # auto-generate story file before dispatching
  repo_path: "/Users/skippy/.multiclaude/repos/ThreeDoors"
```

### Why file-based over alternatives

| Approach | Verdict | Reasoning |
|----------|---------|-----------|
| **File-based YAML** | RECOMMENDED | Consistent with ThreeDoors data model (tasks.yaml, completed.txt), easy to inspect/edit, atomic write pattern already exists, works offline |
| **Git-based (issues)** | Too coupled | Requires GitHub connectivity, mixes task management with issue tracking, adds API dependency |
| **Named pipe/socket** | Over-engineered | Requires daemon on both sides, no persistence if either process crashes, harder to debug |
| **SQLite** | Future option | Could be used if enrichment layer (Epic 6) adds SQLite; premature now |

---

## 4. multiclaude Integration

### Dispatch Flow (Detailed)

```
                        ThreeDoors Process
                        ==================
                        
User presses 'x'
        |
        v
+-------------------+
| Validate task:    |
| - Has text        |
| - Not already     |
|   dispatched      |
| - Status is todo  |
|   or in-progress  |
+-------------------+
        |
        v
+-------------------+
| Write queue item  |
| to dev-queue.yaml |
| (atomic write)    |
+-------------------+
        |
        v
+-------------------+     +-----------------------+
| If auto_dispatch: |     | If !auto_dispatch:    |
| spawn worker now  |     | mark "pending" and    |
|                   |     | wait for approval     |
+-------------------+     +-----------------------+
        |                           |
        v                           v
+-------------------+     +-----------------------+
| Shell out:        |     | User reviews queue    |
| multiclaude       |     | via ':devqueue' cmd   |
| worker create     |     | and approves          |
| "<task + context>"|     +-----------------------+
+-------------------+               |
        |                           v
        +------ same path ----------+
        |
        v
+-------------------+
| Update queue item |
| status: dispatched|
| worker_name: X    |
+-------------------+
        |
        v
+-------------------+
| Poll / watch for  |
| worker completion |
| via multiclaude   |
| history or fsnotify|
+-------------------+
        |
        v
+-------------------+
| Update queue item |
| + original task   |
| with PR number,   |
| status, link      |
+-------------------+
```

### Task-to-Worker Translation

The translation from a ThreeDoors task to a multiclaude worker command needs to produce a rich prompt. The existing `multiclaude worker create` takes a task description string. The system should construct this string by combining:

1. **Task text** (from `Task.Text`)
2. **Task context** (from `Task.Context`)
3. **Acceptance criteria** (from the queue item, if user provided them)
4. **Scope constraints** (directories the agent should focus on)
5. **Standard suffix** (sign commits, no co-authored-by, fork workflow instructions)

Example generated command:

```bash
multiclaude worker create \
  "Implement keyboard shortcut help overlay. Context: Users need to see available key bindings without leaving the TUI. Acceptance criteria: 1) Help overlay appears on '?' keypress 2) Shows all key bindings grouped by context. Scope: internal/tui/. IMPORTANT: Sign all commits (git commit -S). Do NOT add Co-Authored-By for AI. This is a fork workflow — PR targets upstream arcaven/ThreeDoors."
```

### Story File Generation

The existing `AgentService` (`/Users/skippy/.multiclaude/repos/ThreeDoors/internal/intelligence/agent_service.go`) already has `DecomposeAndWrite` which uses an LLM to break a task into BMAD story specs and write them to git. The pipeline should optionally:

1. Call `AgentService.DecomposeAndWrite()` to generate story files
2. Commit the story files to a branch
3. THEN dispatch the worker with instructions to implement the stories

This two-phase approach matches the project's story-driven development rule. Configuration option `require_story: true` controls whether this step is mandatory.

### Polling vs Event-Driven

| Approach | Implementation | Latency | Complexity |
|----------|---------------|---------|------------|
| **Polling** (check `multiclaude history` every 30s) | `tea.Tick` command in Bubbletea | 0-30s delay | Low |
| **File watching** (fsnotify on multiclaude state files) | fsnotify library | Near-instant | Medium |
| **multiclaude message** (inter-agent messaging) | `multiclaude message send` | Near-instant | Medium, requires agent identity |

**Recommendation: Polling via `tea.Tick`** for MVP. The TUI already uses Bubbletea commands for async operations. A 30-second tick that runs `multiclaude repo history --status completed -n 10` and parses the output is simple, reliable, and consistent with the architecture. File watching can be added later as an optimization.

### Safety Guardrails

| Guardrail | Implementation | Priority |
|-----------|---------------|----------|
| **Max concurrent workers** | Check `multiclaude worker list` count before dispatch; reject if >= `max_concurrent` | P0 |
| **Approval gate** | `auto_dispatch: false` requires user to explicitly approve each dispatch from a queue view | P0 |
| **Scope constraint** | Worker task description includes explicit directory scope; CLAUDE.md rules prevent out-of-scope changes | P1 |
| **Cost awareness** | Log estimated token usage per dispatch; warn if daily total exceeds threshold | P1 |
| **Rate limiting** | Minimum 5-minute cooldown between dispatches to the same task | P1 |
| **Dry run mode** | `:dispatch --dry-run` shows what would be dispatched without executing | P2 |
| **Kill switch** | 'K' key in dev queue view runs `multiclaude worker rm <name>` | P0 |

---

## 5. Recommendation Loop

### PR Results Flowing Back to Tasks

```
+-------------------+     +------------------+     +-------------------+
| multiclaude       |     | Dev Queue Poller |     | ThreeDoors Task   |
| worker completes  |     | (30s tick)       |     | Pool              |
|                   |     |                  |     |                   |
| PR #134 created   +---->+ Parses history,  +---->+ Task.DevDispatch  |
| CI: passing       |     | extracts PR #,   |     |   .PRNumber = 134 |
| Status: merged    |     | status, URL      |     |   .PRStatus =     |
|                   |     |                  |     |     "merged"       |
+-------------------+     +------------------+     +-------------------+
                                  |
                                  v
                          +------------------+
                          | Generate new     |
                          | ThreeDoors tasks:|
                          | "Review PR #134" |
                          | "Test feature X" |
                          +------------------+
```

### Task Status Mapping

| multiclaude Worker Status | Queue Item Status | ThreeDoors Task Action |
|--------------------------|-------------------|----------------------|
| `running` | `dispatched` | Task badge shows "Agent working..." |
| `no-pr` (completed, no PR) | `failed` | Task note added: "Agent completed without PR" |
| `open` (PR created) | `completed` | Task.DevDispatch.PRNumber set; new task created: "Review PR #N" |
| `merged` (PR merged) | `completed` | Task status transitions to `complete` automatically |
| Worker removed/crashed | `failed` | Task.DevDispatch.DispatchErr set; badge shows error |

### Auto-Generated Tasks

When a PR is created, the system generates follow-up tasks:
- "Review PR #N: <title>" (status: `todo`, type: `administrative`)
- If CI fails: "Fix CI on PR #N: <failure summary>" (status: `todo`, type: `technical`)
- If PR gets review comments: "Address review comments on PR #N" (status: `todo`)

These generated tasks appear in the normal door rotation, creating a natural review cadence.

---

## 6. Architecture Options

### Option A: File-Based Queue with External Daemon

```
+------------------+          +------------------+          +------------------+
|  ThreeDoors TUI  |  writes  |  dev-queue.yaml  |  reads   |  Queue Daemon    |
|                  +--------->+                  +--------->+  (separate proc) |
|  User dispatches |          |  Pending items   |          |  Polls queue,    |
|  via 'x' key     |          |                  |          |  spawns workers  |
|                  |<---------+  Updated status  |<---------+  via multiclaude |
|  Reads results   |  reads   |                  |  writes  |  CLI             |
+------------------+          +------------------+          +------------------+
```

**Pros:**
- ThreeDoors TUI stays simple -- no subprocess management
- Daemon can run independently (cron, launchd, or persistent process)
- Queue file is the single source of truth
- Daemon can be restarted without affecting TUI

**Cons:**
- Another process to manage (daemon lifecycle, crash recovery)
- Two writers to the queue file (needs file locking or separate status file)
- User must set up the daemon separately
- More complex installation/setup

**Complexity:** Medium-High
**Security:** Daemon needs same filesystem access as multiclaude; file locking prevents corruption

### Option B: TUI-Native Dispatch (RECOMMENDED)

```
+------------------------------------------------------------------+
|                        ThreeDoors TUI Process                     |
|                                                                   |
|  +-------------+    +--------------+    +---------------------+   |
|  | Doors View  |    | Dev Queue    |    | Dispatch Engine     |   |
|  |             |    | View         |    |                     |   |
|  | 'x' key ---+---->| Shows queue  |    | - Builds task desc  |   |
|  |             |    | 'y' approves +---->| - Calls multiclaude |   |
|  |             |    | 'K' kills    |    |   worker create     |   |
|  | Badge shows |    |              |    | - Writes queue YAML |   |
|  | [DEV] status|    +--------------+    | - Polls for results |   |
|  +-------------+                        +---------------------+   |
|                                                                   |
|  +-------------------------------------------------------------+ |
|  | tea.Tick (30s): check multiclaude history, update queue      | |
|  +-------------------------------------------------------------+ |
+------------------------------------------------------------------+
         |                                          ^
         | exec: multiclaude worker create ...      | exec: multiclaude repo history
         v                                          |
+------------------------------------------------------------------+
|                     multiclaude daemon                            |
|  Manages workers, git worktrees, tmux sessions                   |
+------------------------------------------------------------------+
```

**Pros:**
- Single process -- no separate daemon to manage
- Natural Bubbletea integration (views, commands, messages)
- User sees everything in one interface
- Approval gate built into the TUI flow (dev queue view)
- Leverages existing multiclaude daemon (already running)
- Consistent with existing TUI patterns (mood view, health view, sync view)

**Cons:**
- TUI process must be running for polling to work (tasks dispatched while TUI is closed won't get status updates until next launch)
- Subprocess execution (`os/exec`) in a TUI needs careful error handling
- TUI becomes more complex (another view, more key bindings)

**Complexity:** Medium
**Security:** Inherits multiclaude's security model; subprocess calls are constrained to `multiclaude` CLI; no direct git operations

### Option C: Git-Native (GitHub Issues as Queue)

```
+------------------+          +------------------+          +------------------+
|  ThreeDoors TUI  |  gh CLI  |  GitHub Issues   |  watches |  multiclaude     |
|                  +--------->+                  +--------->+  (issue watcher) |
|  ':dispatch'     |          |  Label: dev-auto |          |  Spawns worker   |
|  creates issue   |          |  Body: task desc |          |  per issue       |
|                  |<---------+  Issue closed    |<---------+  Closes issue    |
|  Reads issue     |  gh CLI  |  with PR link   |  gh CLI  |  when PR merged  |
|  status          |          |                  |          |                  |
+------------------+          +------------------+          +------------------+
```

**Pros:**
- GitHub is the single source of truth for both tasks and PRs
- Natural PR-to-issue linking
- Works when TUI is not running (GitHub persists state)
- CI status naturally visible on issues
- multiclaude already has PR awareness (PR shepherd, review agents)

**Cons:**
- Requires GitHub connectivity (violates local-first/offline-first principles)
- Requires new multiclaude feature (issue watching)
- Conflates ThreeDoors tasks with GitHub issues (different audiences)
- API rate limits could be a concern
- Adds GitHub as a hard dependency for a local-first app
- Task management split across two systems

**Complexity:** High
**Security:** Requires GitHub token with issue+PR permissions; API calls expose task content to GitHub

### Comparison Matrix

| Criterion | Option A (Daemon) | Option B (TUI-Native) | Option C (Git-Native) |
|-----------|-------------------|----------------------|----------------------|
| Complexity | Medium-High | **Medium** | High |
| Offline support | Yes | **Yes** | No |
| User experience | Split (TUI + daemon) | **Unified** | Split (TUI + GitHub) |
| Setup effort | High (daemon config) | **Low** (built-in) | High (GitHub config) |
| Reliability | Good (separate process) | Good (TUI must be running) | Dependent on GitHub |
| Local-first | **Yes** | **Yes** | No |
| Existing infra reuse | Partial | **High** (multiclaude CLI, tea.Tick) | Partial |

---

## 7. Bootstrapping / Dogfooding

### The Meta-Pipeline: Building This Feature With Itself

The minimum viable version that could bootstrap itself:

**MVP-0 (Manual bridge, zero code changes):**
Right now, a human can:
1. Look at a ThreeDoors task
2. Manually run `multiclaude worker create "<task text>"`
3. Manually check `multiclaude repo history` for results
4. Manually update the task

This is the current workflow. Everything below automates steps of this workflow.

**MVP-1 (Queue file + dispatch command, ~2 stories):**
1. Add `:dispatch` command to TUI that writes a queue entry to `~/.threedoors/dev-queue.yaml`
2. Add a shell script `scripts/dispatch-worker.sh` that reads the queue and runs `multiclaude worker create`
3. User manually runs the script, or sets up a cron job

At this point, the feature could dispatch its own stories:
- Task: "Add dev queue view to TUI" --> `:dispatch` --> script runs --> worker creates PR
- Task: "Add polling for worker status" --> `:dispatch` --> script runs --> worker creates PR

**MVP-2 (TUI-integrated dispatch + polling, ~3 stories):**
1. TUI calls `multiclaude worker create` directly via `os/exec` wrapped in a `tea.Cmd`
2. 30-second `tea.Tick` polls `multiclaude repo history` and updates queue/task state
3. Dev queue view shows pending/dispatched/completed items

**MVP-3 (Full loop, ~2 stories):**
1. PR results auto-update task status
2. Auto-generated review tasks
3. Story file generation via existing `AgentService`

### Dogfooding Sequence

| Step | What Gets Built | Built By |
|------|----------------|----------|
| 1 | Story files for MVP-1 | Human (or LLM decomposer) |
| 2 | MVP-1 implementation | `multiclaude worker create` (manually triggered) |
| 3 | Story files for MVP-2 | MVP-1 (dispatched via `:dispatch` + script) |
| 4 | MVP-2 implementation | MVP-1 pipeline |
| 5 | Story files for MVP-3 | MVP-2 (dispatched from TUI) |
| 6 | MVP-3 implementation | MVP-2 pipeline |

---

## 8. Existing Infrastructure to Leverage

### Already Built

| Component | Location | Relevance |
|-----------|----------|-----------|
| `AgentService` | `/Users/skippy/.multiclaude/repos/ThreeDoors/internal/intelligence/agent_service.go` | LLM task decomposition + git story writing |
| `TaskDecomposer` | `/Users/skippy/.multiclaude/repos/ThreeDoors/internal/intelligence/llm/decomposer.go` | Breaks task text into BMAD story specs |
| `GitOutputWriter` | `/Users/skippy/.multiclaude/repos/ThreeDoors/internal/intelligence/llm/git_writer.go` | Writes story files to repo branches |
| `Task` model | `/Users/skippy/.multiclaude/repos/ThreeDoors/internal/core/task.go` | Task struct with YAML serialization, extensible with new fields |
| `TaskStatus` state machine | `/Users/skippy/.multiclaude/repos/ThreeDoors/internal/core/task_status.go` | Validated transitions; dispatch is orthogonal, not a new status |
| multiclaude CLI | `multiclaude worker create`, `repo history`, `worker list`, `message send` | Full worker lifecycle management |
| multiclaude agents | worker, reviewer, pr-shepherd, merge-queue | Existing agent templates for the pipeline |
| Atomic file writes | Pattern used throughout adapters | Safe queue file persistence |
| `tea.Tick` pattern | Used in sync status polling | 30-second poll for worker status |
| Command palette | `:add`, `:edit`, `:mood`, `:stats`, `:help` | Natural home for `:dispatch`, `:devqueue` |

### multiclaude CLI Capabilities (Observed)

```
multiclaude worker create <task>    # Spawn a worker with a task description
multiclaude worker list             # List active workers (name, status, branch, task)
multiclaude worker rm <name>        # Remove/kill a worker
multiclaude repo history            # Show completed tasks with PR numbers, summaries
multiclaude message send <to> <msg> # Inter-agent messaging
multiclaude review <pr-url>         # Spawn a review agent for a PR
multiclaude config                  # Fork mode, merge queue, PR shepherd settings
```

---

## 9. Security and Safety Considerations

### Threat Model

| Threat | Likelihood | Impact | Mitigation |
|--------|-----------|--------|------------|
| Runaway agents (100 workers spawned) | Medium | High (API cost, system resources) | `max_concurrent` limit (default: 2), rate limiting (5 min cooldown) |
| Agent modifies files outside scope | Medium | Medium (unintended changes) | Scope constraint in task description; CLAUDE.md rules; PR review before merge |
| Agent pushes directly to main | Low | High (broken main branch) | multiclaude fork workflow enforces feature branches; user CLAUDE.md forbids main push |
| Queue file corruption | Low | Medium (lost dispatch state) | Atomic write pattern; queue backup before writes |
| Sensitive data in task descriptions | Low | Medium (exposed in PR/commit) | Warn if task text contains patterns matching secrets (API keys, passwords) |
| Cost explosion from LLM decomposition | Medium | Medium (API charges) | Daily cost cap in settings; decomposition is optional; prefer local Ollama |
| Agent creates malicious code | Very Low | High | All changes go through PR; CI gates; human review required for merge |

### Mandatory Guardrails (Non-Negotiable)

1. **Human-in-the-loop for merge**: The maintainer (arcaven) must merge PRs. multiclaude cannot auto-merge. This is already enforced by the fork workflow.

2. **No auto-dispatch by default**: `auto_dispatch: false` means the user must explicitly approve each dispatch from the dev queue view. This prevents accidental mass-dispatch.

3. **Worker count cap**: Hard limit of `max_concurrent` workers (default: 2). The system checks `multiclaude worker list` before every dispatch and refuses if at capacity.

4. **Commit signing**: All commits must be signed (`git commit -S`). This is enforced by the global gitconfig and CLAUDE.md rules.

5. **No force push**: Workers cannot force-push. multiclaude agents follow CLAUDE.md which forbids `--force` to main/develop.

### Recommended Guardrails (Configurable)

6. **Scope validation**: Before dispatch, validate that the task description references only in-scope directories. Warn if the task seems to affect areas outside `internal/`, `cmd/`, `docs/`.

7. **Daily dispatch limit**: Default 10 dispatches per day. Configurable in settings.

8. **Cooldown period**: Minimum 5 minutes between dispatches. Prevents rapid-fire experimentation that wastes resources.

9. **Dry run mode**: `:dispatch --dry-run` shows the full multiclaude command that would be executed, without executing it.

10. **Audit log**: Every dispatch, completion, and failure is logged to `~/.threedoors/dev-dispatch.log` in JSONL format (consistent with `sessions.jsonl`).

---

## 10. Detailed Component Design

### New Files Required

```
internal/core/dev_dispatch.go          # DevDispatch struct, queue item model
internal/core/dev_dispatch_test.go     # Tests for dispatch model
internal/core/dispatch_engine.go       # Queue management, multiclaude CLI wrapper
internal/core/dispatch_engine_test.go  # Tests with mock CommandRunner
internal/tui/devqueue_view.go          # Bubbletea view for dev queue
internal/tui/devqueue_view_test.go     # TUI tests
internal/tui/messages.go               # Add DispatchTaskMsg, WorkerStatusMsg
```

### Modified Files

```
internal/core/task.go                  # Add DevDispatch field to Task struct
internal/tui/detail_view.go            # Add 'x' key binding for dispatch
internal/tui/doors_view.go             # Add [DEV] badge rendering
internal/tui/main_model.go             # Add dev queue view routing, tick polling
```

### Key Interface

```go
// Dispatcher abstracts the multiclaude CLI for testability.
type Dispatcher interface {
    CreateWorker(ctx context.Context, task string) (workerName string, err error)
    ListWorkers(ctx context.Context) ([]WorkerInfo, error)
    GetHistory(ctx context.Context, limit int) ([]HistoryEntry, error)
    RemoveWorker(ctx context.Context, name string) error
}
```

This follows the project's "accept interfaces, return concrete types" pattern. The concrete implementation wraps `os/exec` calls to the multiclaude CLI. Tests use a mock implementation.

---

## 11. End-to-End Data Flow (ASCII Diagram)

```
USER                    ThreeDoors TUI                    multiclaude
====                    ==============                    ===========

Sees task in door
  "Add help overlay"
        |
        |  presses 'x'
        v
                    +---> Validate task
                    |     (not already dispatched,
                    |      valid status)
                    |
                    +---> Show dispatch confirmation
                    |     "Dispatch to dev queue?"
                    |     [y/n]
                    |
        |  presses 'y'
        v
                    +---> Write to dev-queue.yaml
                    |     status: "pending"
                    |
                    +---> Update Task.DevDispatch
                    |     Queued: true
                    |     QueuedAt: now()
                    |
                    +---> If auto_dispatch:
                    |       exec: multiclaude worker
                    |             create "<task desc>"
                    |       status: "dispatched"
                    |
                    |     If !auto_dispatch:
                    |       Show in dev queue view
                    |       User approves manually
                    |
                    |                                     Worker spawned
                    |                                     in git worktree
                    |                                     |
                    |                                     v
                    |                                     Reads CLAUDE.md
                    |                                     Writes code
                    |                                     Runs make test
                    |                                     Creates PR
                    |                                     |
                    |  30s tick                            v
                    +---> exec: multiclaude               Worker status:
                    |     repo history                    "merged" / "open"
                    |     |                               PR #134
                    |     v
                    +---> Parse history output
                    |     Match worker_name to queue item
                    |
                    +---> Update dev-queue.yaml
                    |     status: "completed"
                    |     pr_number: 134
                    |
                    +---> Update Task.DevDispatch
                    |     PRNumber: 134
                    |     PRStatus: "open"
                    |
                    +---> Create new task:
                          "Review PR #134: Add help overlay"
                          status: todo
                          type: administrative

Sees new "Review PR"
task in door rotation
```

---

## 12. Story Breakdown for Implementation

### Epic: Self-Driving Development Pipeline

| Story | Title | Dependencies | Effort |
|-------|-------|-------------|--------|
| SD.1 | Dev dispatch data model and queue persistence | None | Small |
| SD.2 | Dispatch engine with multiclaude CLI wrapper | SD.1 | Medium |
| SD.3 | TUI dispatch key binding and confirmation flow | SD.1 | Small |
| SD.4 | Dev queue view (list, approve, kill) | SD.1, SD.2 | Medium |
| SD.5 | Worker status polling and task update loop | SD.2 | Medium |
| SD.6 | Auto-generated review/follow-up tasks | SD.5 | Small |
| SD.7 | Optional story file generation via AgentService | SD.2 | Medium |
| SD.8 | Safety guardrails (rate limiting, cost caps, audit log) | SD.2 | Small |

---

## 13. Open Questions

1. **Should the dispatch engine live in `internal/core/` or a new `internal/dispatch/` package?** Given the project's package-by-feature convention and the "keep packages small" rule, a new `internal/dispatch/` package is justified since it has its own concerns (queue management, CLI wrapping, polling).

2. **Should queue items persist across TUI restarts?** Yes. The file-based queue survives process restarts. On TUI launch, the dispatch engine reads the queue and resumes polling for any items in `dispatched` status.

3. **What happens if multiclaude is not installed or the daemon is not running?** The dispatch engine should check for `multiclaude` on PATH during initialization. If not found, the `:dispatch` command and 'x' key binding should be hidden/disabled with a message: "multiclaude not found -- dev dispatch unavailable."

4. **Should this feature be gated behind a config flag?** Yes. Add `dev_dispatch_enabled: true` to `~/.threedoors/config.yaml`. Disabled by default. This prevents accidental exposure and keeps the TUI clean for users who do not use multiclaude.

5. **How should the feature handle the fork workflow?** The existing multiclaude fork configuration (`Fork of: arcaven/ThreeDoors`) means workers push to the fork and create PRs against upstream. The dispatch engine does not need to manage this -- multiclaude handles it. However, the task description suffix should include fork workflow instructions (as shown in the current worker history).

---

## 14. Recommendation Summary

| Decision | Recommendation | Rationale |
|----------|---------------|-----------|
| Architecture | **Option B: TUI-Native Dispatch** | Unified UX, leverages existing infra, local-first |
| Task designation | **New `DevDispatch` field + 'x' key + `:dispatch` command** | Orthogonal to task lifecycle, explicit user intent |
| Queue format | **File-based YAML at `~/.threedoors/dev-queue.yaml`** | Consistent with data model, offline-capable, inspectable |
| Worker status | **30-second `tea.Tick` polling via `multiclaude repo history`** | Simple, reliable, matches Bubbletea patterns |
| Story generation | **Optional, via existing `AgentService`** | Already built, configurable per dispatch |
| Safety | **No auto-dispatch by default, max 2 concurrent, human merge required** | Conservative defaults, user can relax constraints |
| First milestone | **MVP-1: `:dispatch` command + queue file + manual script** | Can ship in 1-2 stories, enables dogfooding immediately |