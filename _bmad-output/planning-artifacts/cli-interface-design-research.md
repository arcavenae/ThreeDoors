# CLI Interface Design Research

**Date:** 2026-03-06
**Author:** cool-rabbit (research agent)
**Method:** Analyst deep-dive + 2 rounds of BMAD party mode ideation
**Scope:** Complete non-TUI CLI alternative for humans and LLMs/Claude

---

## Executive Summary

ThreeDoors needs a CLI interface that serves two audiences: **human power users** who want scriptable task management without leaving their terminal, and **LLM agents** (Claude Code, multiclaude workers) who need a programmatic interface to read/write tasks. The existing codebase is exceptionally well-positioned for this: `internal/core` has zero Bubbletea dependencies, and all domain types already carry JSON struct tags.

The recommended approach: **Cobra-based subcommand CLI** in a new `internal/cli/` package, sharing `internal/core` with the existing TUI. `threedoors` with no args launches the TUI (backward compatible); any subcommand routes to CLI handlers. A global `--json` flag switches output from human-readable to structured JSON for LLM consumption.

---

## 1. Codebase Readiness Assessment

### Domain Logic: ~90% CLI-Ready

The `internal/core` package is cleanly separated from the TUI layer. All critical operations have zero Bubbletea dependencies:

| Component | File | CLI-Ready? | Notes |
|-----------|------|------------|-------|
| `TaskPool` (CRUD) | `task_pool.go` | Yes | `AddTask`, `GetTask`, `UpdateTask`, `RemoveTask`, `GetAllTasks`, `GetTasksByStatus` |
| `Task` model | `task.go` | Yes | Full JSON tags on all fields, `UpdateStatus()`, `AddNote()`, `SetBlocker()`, `Validate()` |
| `SelectDoors()` | `door_selector.go` | Yes | Pure function: `SelectDoors(pool, count)` returns `[]*Task` |
| `SelectDoorsWithMood()` | `mood_selector.go` | Yes | Mood-aware selection, also pure |
| `SessionTracker` | `session_tracker.go` | Yes | `RecordMood()`, `RecordTaskCompleted()`, `Finalize()` |
| `HealthChecker` | `health_checker.go` | Yes | `RunAll()` returns `HealthCheckResult` (JSON-serializable) |
| `PatternAnalyzer` | `pattern_analyzer.go` | Yes | `Analyze()`, `GetDailyCompletions()`, `GetWeekOverWeek()`, `GetMoodCorrelations()` |
| `Registry` | `registry.go` | Yes | `ListProviders()`, `InitProvider()`, `GetProvider()` |
| `ProviderConfig` | `provider_config.go` | Yes | `LoadProviderConfig()`, `SaveProviderConfig()` |
| `MetricsWriter` | `metrics_writer.go` | Yes | `AppendSession()` |
| `metrics.Reader` | `metrics/reader.go` | Yes | `ReadAll()`, `ReadSince()`, `ReadLast()` |
| Task status transitions | `task_status.go` | Yes | `IsValidTransition()`, `GetValidTransitions()`, `ValidateStatus()` |

### TUI-Coupled Components (NOT needed for CLI)

| Component | Why TUI-Only |
|-----------|-------------|
| `internal/tui/*.go` | Bubbletea `tea.Model` implementations, `tea.Msg` types |
| Door themes (`theme_picker.go`) | Visual rendering only |
| Lipgloss styles (`styles.go`) | Terminal styling |
| View transitions (`messages.go`) | Inter-view communication |

### Key Advantage: JSON Tags Already Present

```go
// internal/core/task.go — already has dual tags
type Task struct {
    ID             string       `yaml:"id" json:"id"`
    Text           string       `yaml:"text" json:"text"`
    Status         TaskStatus   `yaml:"status" json:"status"`
    Type           TaskType     `yaml:"type,omitempty" json:"type,omitempty"`
    Effort         TaskEffort   `yaml:"effort,omitempty" json:"effort,omitempty"`
    // ... all fields have json tags
}
```

Same applies to `SessionMetrics`, `PatternReport`, `HealthCheckResult`, `HealthCheckItem`, `MoodCorrelation`, `AvoidanceEntry`, `DoorPositionStats`, etc.

---

## 2. Architecture Decision: Layered CLI/TUI Coexistence

### Package Structure

```
cmd/threedoors/main.go    --> detect subcommand --> route to cli or tui
internal/cli/             --> Cobra commands, output formatters (NEW)
internal/core/            --> Domain logic (SHARED, unchanged)
internal/tui/             --> Bubbletea views (TUI only, unchanged)
```

### Routing Logic in main.go

```go
func main() {
    // If no args or only flags, launch TUI (backward compatible)
    // If first arg is a known subcommand, route to Cobra CLI
    if len(os.Args) > 1 && isSubcommand(os.Args[1]) {
        cli.Execute() // Cobra root command
    } else {
        launchTUI()   // Existing Bubbletea program
    }
}
```

This preserves 100% backward compatibility: `threedoors` alone launches the TUI exactly as today. `threedoors task list` routes to CLI.

### Dependency Graph

```
cmd/threedoors/main.go
    |
    +---> internal/cli/   (imports core, NOT tui)
    |         |
    +---> internal/tui/   (imports core, NOT cli)
              |
              +---> internal/core/  (imported by both, imports neither)
```

`internal/core` remains the shared domain layer. No circular dependencies.

---

## 3. Framework Recommendation: Cobra

### Why Cobra

| Criteria | Cobra | stdlib `flag` | Kong | urfave/cli |
|----------|-------|---------------|------|------------|
| Subcommand support | Native | Manual | Native | Native |
| Shell completions | Built-in | None | Plugin | Plugin |
| Help generation | Automatic | Basic | Automatic | Automatic |
| Go ecosystem adoption | kubectl, docker, gh, hugo | - | Smaller | Moderate |
| Viper config integration | Native | None | None | None |
| Man page generation | Built-in | None | None | None |

### Why NOT Alternatives

- **stdlib `flag`**: No subcommand routing. Would require manual dispatch table.
- **Kong**: Type-driven design fights Go interface conventions. Smaller ecosystem.
- **urfave/cli**: Less mature, smaller community. No Viper integration.

### Viper for Config

The project already uses `gopkg.in/yaml.v3` for `config.yaml`. Viper reads the same format and adds:
- Environment variable overrides (`THREEDOORS_PROVIDER=obsidian`)
- Config precedence: flags > env vars > config file > defaults
- Single-value access: `viper.GetString("provider")`

This is optional — Cobra works fine without Viper. The existing `LoadProviderConfig()` / `SaveProviderConfig()` functions may be sufficient.

---

## 4. Command Taxonomy

### Design Principles

1. **Noun-verb pattern**: `threedoors task add` (like `gh issue create`)
2. **Signature command**: `threedoors doors` is the CLI equivalent of launching the TUI
3. **Consistency**: Every resource gets `list`, `show`, `add`/`set` verbs where applicable
4. **Batch support**: Variadic IDs for bulk operations
5. **Discoverability**: `threedoors --help` shows command groups

### Complete Command Reference

#### Core Experience

```
threedoors                              # Launch TUI (no args)
threedoors doors                        # Get 3 random tasks (human-readable)
threedoors doors --json                 # Get 3 random tasks (JSON for LLMs)
threedoors doors --pick 1               # Select door 1 (scripting)
threedoors doors --interactive          # Interactive pick-a-door prompt
```

#### Task Management

```
threedoors task list                    # List all active tasks
threedoors task list --status todo      # Filter by status
threedoors task list --type code        # Filter by type
threedoors task list --effort small     # Filter by effort
threedoors task list --status todo --type code  # Compose filters
threedoors task list --json             # JSON output

threedoors task show <id>               # Show task detail (supports prefix match)
threedoors task show <id> --json        # JSON detail

threedoors task add "Buy groceries"     # Add task with text
threedoors task add --stdin             # Read task text from stdin
threedoors task add --context "Need for dinner"  # Add with context
echo "Buy milk" | threedoors task add   # Pipe task text

threedoors task edit <id> --text "New text"      # Edit task text
threedoors task edit <id> --context "Updated why" # Edit context

threedoors task delete <id>             # Delete task
threedoors task delete <id1> <id2>      # Batch delete

threedoors task complete <id>           # Mark complete
threedoors task complete <id1> <id2>    # Batch complete

threedoors task block <id> --reason "Waiting on API"  # Mark blocked
threedoors task unblock <id>            # Unblock (-> todo)

threedoors task status <id> in-progress # Change status
threedoors task status <id> in-review   # Any valid transition

threedoors task expand <id> --subtask "Write tests"   # Create subtask
threedoors task fork <id>               # Clone/fork task

threedoors task note <id> "Progress update"  # Add note
threedoors task search "query"          # Search tasks
threedoors task search "query" --json   # JSON search results

threedoors task tag <id> --type code --effort medium --location home  # Set categories
```

#### Mood Tracking

```
threedoors mood set <mood>              # Record mood (focused, energized, etc.)
threedoors mood set custom "Feeling creative"  # Custom mood
threedoors mood history                 # Show mood entries
threedoors mood history --json          # JSON mood history
```

#### Session Metrics

```
threedoors stats                        # Summary dashboard
threedoors stats --daily                # Daily completions (last 7 days)
threedoors stats --weekly               # Week-over-week comparison
threedoors stats --patterns             # Full pattern analysis
threedoors stats --json                 # JSON metrics
```

#### Configuration

```
threedoors config show                  # Dump full config
threedoors config show --json           # JSON config
threedoors config get <key>             # Get single value
threedoors config set <key> <value>     # Set value
threedoors config set theme modern      # Example: change theme
threedoors config set provider obsidian # Example: change provider
```

#### Provider Management

```
threedoors provider list                # List registered providers
threedoors provider health              # Health check all providers
threedoors provider health --json       # JSON health results
threedoors provider sync                # Trigger sync cycle
```

#### System

```
threedoors health                       # Full system health check
threedoors health --json                # JSON health
threedoors version                      # Version info
threedoors version --json               # JSON version
```

---

## 5. Output Format Design

### Human-Readable (Default)

Clean table output using a lightweight table formatter (no heavy dependency — `text/tabwriter` from stdlib or `tablewriter`):

```
$ threedoors task list --status todo

ID        STATUS  TYPE   EFFORT  TEXT
abc123    todo    code   medium  Write CLI interface research doc
def456    todo    admin  small   Reply to email thread
ghi789    todo    code   large   Refactor sync engine

3 tasks found
```

### JSON Output (`--json` flag)

Structured envelope with schema version for forward compatibility:

```json
{
  "schema_version": 1,
  "command": "task.list",
  "data": [
    {
      "id": "abc123-def456-...",
      "text": "Write CLI interface research doc",
      "status": "todo",
      "type": "code",
      "effort": "medium",
      "created_at": "2026-03-06T10:00:00Z",
      "updated_at": "2026-03-06T10:00:00Z"
    }
  ],
  "metadata": {
    "total": 42,
    "filtered": 3,
    "filters_applied": {"status": "todo"}
  }
}
```

### Design Decisions

1. **Schema versioning**: `schema_version: 1` in every JSON response. Breaking changes bump the version. LLMs can check compatibility.
2. **Command field**: Tells consumers what produced the output (useful for logging/debugging).
3. **Metadata**: Total count, filter info, pagination data when applicable.
4. **No CSV/YAML output initially**: JSON + human-readable covers 99% of use cases. Add `--format csv` later if needed.
5. **Consistent envelope**: Even single-object responses (like `task show`) use the same structure with `data` as a single object (not array).

---

## 6. ID Prefix Matching

UUIDs are hostile in CLIs. Support **prefix matching** like git's short SHAs:

```
$ threedoors task show abc
# Matches abc123-def456-789...

$ threedoors task show a
# Error: ambiguous prefix "a": 5 matches (use more characters)

$ threedoors task complete abc def ghi
# Batch with prefixes
```

### Implementation

Add to `TaskPool`:

```go
func (tp *TaskPool) FindByPrefix(prefix string) ([]*Task, error) {
    var matches []*Task
    for id, t := range tp.tasks {
        if strings.HasPrefix(id, prefix) {
            matches = append(matches, t)
        }
    }
    if len(matches) == 0 {
        return nil, ErrTaskNotFound
    }
    if len(matches) > 1 {
        return matches, fmt.Errorf("ambiguous prefix %q: %d matches", prefix, len(matches))
    }
    return matches, nil
}
```

LLMs always use full IDs (from JSON output). Humans use short prefixes.

---

## 7. Exit Code Scheme

| Code | Meaning | Example |
|------|---------|---------|
| 0 | Success | Task created, listed, completed |
| 1 | General error | Provider failure, I/O error |
| 2 | Not found | Task ID doesn't exist |
| 3 | Validation error | Invalid status transition, bad input |
| 4 | Provider error | Provider health check failed |
| 5 | Ambiguous input | ID prefix matches multiple tasks |

For batch operations: exit 0 if all succeed, exit 1 if any fail (but still process all items). Per-item results in both human and JSON output.

---

## 8. Stdin and Pipe Patterns

### Reading from Stdin

```go
// Detect if stdin is a pipe/redirect (not a terminal)
if !term.IsTerminal(int(os.Stdin.Fd())) {
    // Read task text from stdin
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        text := scanner.Text()
        // Create task from each line
    }
}
```

### Supported Patterns

```bash
# Pipe task text
echo "Buy groceries" | threedoors task add

# Pipe multiple tasks (one per line)
cat tasks.txt | threedoors task add --stdin

# Pipe to jq for processing
threedoors task list --json | jq '.data[] | select(.status=="blocked") | .id'

# Batch complete from filtered list
threedoors task list --json --status blocked | \
  jq -r '.data[].id' | \
  xargs threedoors task complete

# Chain commands
threedoors doors --json | jq -r '.data[0].id' | xargs threedoors task complete
```

### Non-Interactive by Default

- `threedoors doors` prints 3 tasks and exits (non-interactive)
- `threedoors doors --interactive` prompts for door selection (human use)
- Auto-detect: if stdout is not a terminal (piped), force non-interactive even with `--interactive`

---

## 9. LLM Integration Surface

### Primary Interface: CLI with `--json`

The `--json` flag is the LLM integration surface. No separate API server needed. LLMs invoke CLI commands via `subprocess`/`exec` and parse JSON responses.

### Typical LLM Workflow

```bash
# 1. Get three doors
DOORS=$(threedoors doors --json)

# 2. Pick a task (LLM decides based on context)
TASK_ID=$(echo $DOORS | jq -r '.data[0].id')

# 3. Work on task...

# 4. Mark complete
threedoors task complete $TASK_ID --json

# 5. Record mood
threedoors mood set focused

# 6. Check stats
threedoors stats --json
```

### Future: MCP Server (Out of Scope)

MCP (Model Context Protocol) is the emerging standard for LLM tool integration. A future `threedoors mcp serve` command could expose all CLI operations as MCP tools, enabling native integration with Claude, ChatGPT, and other MCP-compatible agents. This is noted as a high-leverage future opportunity but is out of scope for the initial CLI implementation.

The CLI `--json` interface serves as the foundation for any future MCP integration — the JSON schemas would be reused directly.

---

## 10. Implementation Phasing

### Phase 1: Minimum Viable CLI (Recommended First Epic)

Core commands that enable both human and LLM usage:

1. `threedoors doors` / `threedoors doors --json`
2. `threedoors task list` (with `--status`, `--json` filters)
3. `threedoors task show <id>`
4. `threedoors task add "text"`
5. `threedoors task complete <id>`
6. `threedoors task block <id> --reason "..."`
7. `threedoors health` / `threedoors health --json`
8. `threedoors version`

Dependencies: `github.com/spf13/cobra`

**New files:**
- `internal/cli/root.go` — Cobra root command, `--json` persistent flag
- `internal/cli/task.go` — task subcommands
- `internal/cli/doors.go` — doors command
- `internal/cli/health.go` — health command
- `internal/cli/output.go` — JSON/table output formatter
- `internal/cli/output_test.go`
- `internal/cli/task_test.go`
- `internal/cli/doors_test.go`

**Modified files:**
- `cmd/threedoors/main.go` — add subcommand detection and CLI routing
- `internal/core/task_pool.go` — add `FindByPrefix()` method
- `go.mod` — add Cobra dependency

### Phase 2: Extended CLI

- `threedoors task edit`, `delete`, `status`, `note`, `search`, `tag`, `expand`, `fork`
- `threedoors mood set/history`
- `threedoors stats` (daily, weekly, patterns)
- `threedoors config get/set/show`
- Stdin support for `task add`
- `--interactive` flag for `doors`
- Shell completions (Cobra built-in)

### Phase 3: Advanced Features

- `threedoors provider list/health/sync`
- `threedoors task list --format csv`
- `threedoors mcp serve` (MCP server)
- Batch operations with per-item error reporting
- `threedoors completion bash/zsh/fish` (shell completion scripts)

---

## 11. Testing Strategy

### Unit Tests (internal/cli/)

- Table-driven tests for each command handler
- Mock `TaskProvider` via interface (already exists in `core`)
- Assert JSON output structure and human-readable formatting
- Verify exit codes for all error paths

### Integration Tests

- Real `TextFileProvider` in temp directory
- Run CLI commands via `exec.Command`
- Verify file state changes after operations

### Golden File Tests

- Same pattern as `internal/tui/golden_test.go`
- Capture expected output for human-readable and JSON modes
- Compare against stored reference files
- Test at multiple terminal widths for table formatting

### Exit Code Tests

- Every error path returns the documented code
- Batch operations return correct composite codes

### Stdin/Pipe Tests

- Pipe test data through `exec.Command` stdin
- Verify non-interactive behavior when stdout is redirected

---

## 12. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| JSON schema changes break LLM consumers | High | Schema versioning (`schema_version: 1`) in all responses |
| UUID IDs hostile for humans | Medium | Prefix matching with clear error on ambiguity |
| Cobra dependency bloat | Low | Cobra is lightweight (~2MB binary size increase) |
| CLI and TUI logic drift | Medium | Both import `internal/core` — domain logic is shared, not duplicated |
| Interactive mode confuses LLMs | Medium | Non-interactive by default; auto-detect TTY |
| Config changes via CLI conflict with TUI | Low | Both use `LoadProviderConfig`/`SaveProviderConfig` with atomic writes |

---

## 13. Party Mode Consensus Summary

Two rounds of BMAD multi-agent ideation produced strong consensus on:

1. **Cobra framework** (unanimous among architect, dev, QA)
2. **`internal/cli/` package** alongside `internal/tui/` — both import `internal/core` (unanimous)
3. **`--json` flag** as the LLM integration surface (unanimous)
4. **Noun-verb command taxonomy**: `threedoors task <verb>` (strong consensus)
5. **`threedoors doors`** as the signature CLI command (unanimous)
6. **Non-interactive by default**, `--interactive` opt-in (consensus)
7. **Short ID prefix matching** for human usability (consensus)
8. **JSON schema versioning** for LLM stability (consensus from QA + architect)
9. **Exit codes 0-5** for machine-parseable error handling (consensus)
10. **Phase implementation** — ship 8 commands first, expand based on usage (pragmatic consensus)

### Noted Future Opportunities (Out of Scope)

- MCP server for native LLM tool integration
- GraphQL-style field selection (`--fields id,text,status`)
- Watch mode (`threedoors task list --watch`)
- REPL mode (`threedoors shell`)
- Remote CLI access via SSH

---

## Appendix A: Comparable CLI Tools

| Tool | Pattern | JSON Support | Relevant Lesson |
|------|---------|-------------|-----------------|
| `gh` (GitHub CLI) | `gh <noun> <verb>` | `--json` flag | Best-in-class CLI/JSON dual output |
| `kubectl` | `kubectl <verb> <noun>` | `-o json` | Resource-oriented, good for APIs |
| `task` (Taskwarrior) | `task <verb> [filter]` | JSON export | Powerful filtering, complex syntax |
| `todo.txt-cli` | `todo.sh <verb>` | None | Simple but limited |
| `jira-cli` | `jira issue list` | `--json` | Enterprise integration patterns |

The `gh` CLI is the closest model for ThreeDoors: noun-verb commands, `--json` for machine consumption, interactive prompts for humans when needed.

---

## Appendix B: JSON Schema Examples

### doors response

```json
{
  "schema_version": 1,
  "command": "doors",
  "data": [
    {
      "door": 1,
      "task": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "text": "Write CLI interface research doc",
        "status": "todo",
        "type": "code",
        "effort": "medium",
        "location": "",
        "created_at": "2026-03-06T10:00:00Z",
        "updated_at": "2026-03-06T10:00:00Z"
      }
    },
    {"door": 2, "task": {"...": "..."}},
    {"door": 3, "task": {"...": "..."}}
  ],
  "metadata": {
    "total_available": 42,
    "selection_method": "diversity"
  }
}
```

### task.list response

```json
{
  "schema_version": 1,
  "command": "task.list",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "text": "Write CLI interface research doc",
      "status": "todo",
      "type": "code",
      "effort": "medium",
      "notes": [],
      "created_at": "2026-03-06T10:00:00Z",
      "updated_at": "2026-03-06T10:00:00Z"
    }
  ],
  "metadata": {
    "total": 42,
    "filtered": 3,
    "filters": {"status": "todo"}
  }
}
```

### health response

```json
{
  "schema_version": 1,
  "command": "health",
  "data": {
    "overall": "OK",
    "duration_ms": 45,
    "checks": [
      {"name": "Task File", "status": "OK", "message": "Task file exists and is writable"},
      {"name": "Database", "status": "OK", "message": "42 tasks loaded successfully"},
      {"name": "Sync Status", "status": "WARN", "message": "Last sync: 2 hours ago"}
    ]
  }
}
```

### Error response

```json
{
  "schema_version": 1,
  "command": "task.show",
  "error": {
    "code": 2,
    "message": "task not found",
    "detail": "no task matches prefix \"xyz\""
  }
}
```
