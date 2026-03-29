# Slack Bot as multiclaude Control Plane & Multi-Supervisor Coordinator

**Date:** 2026-03-29
**Type:** Research Spike
**Status:** Complete

---

## 1. Existing Bot Search Results

### What We Found

Searched all three GitHub organizations/users:
- **arcavenae** — No bot repos. Found `switchboard` (private, Go tmux session router — potentially related infrastructure but not a Slack bot)
- **arcaven** — No bot repos. 30 repos, mostly forks (Terraform, security tools, BMAD)
- **arcavenai** — 2 repos (ThreeDoors fork, `gt` private). No bot.

GitHub-wide search for "ourbot" across these orgs: **no matches**.

**Conclusion:** The "ourbot" repo was not found on GitHub under any of the user's known organizations. It may exist on:
- GitLab (not searchable with current `gh` CLI auth)
- A personal/work GitLab instance
- A different GitHub org/user not checked
- May have been deleted or renamed

**Recommendation:** Ask the user to check GitLab or provide the exact location. If starting fresh, the existing `switchboard` repo (arcavenae/switchboard) is interesting context — it's a Go-based tmux session router that could complement or inform the bot's multi-machine architecture.

---

## 2. Architecture Design

### 2.1 RBAC (Role-Based Access Control)

#### Recommended Approach: Slack User ID → Role Mapping with Channel Gates

```
┌─────────────────────────────────────────────┐
│                  RBAC Model                  │
├─────────────────────────────────────────────┤
│ Layer 1: Channel-Based Gating               │
│   #dark-factory  → control commands allowed │
│   #mc-status     → read-only queries only   │
│   #general       → general bot features     │
│                                             │
│ Layer 2: User-Role Mapping                  │
│   admin    → full control, config changes   │
│   operator → dispatch work, check status    │
│   viewer   → read-only status queries       │
│                                             │
│ Layer 3: Audit Trail                        │
│   Every command → timestamp, user, channel, │
│   command, target supervisor, result         │
└─────────────────────────────────────────────┘
```

**Implementation:**
- Store role mappings in a YAML config file (consistent with ThreeDoors patterns)
- Slack user IDs are immutable and reliable for identity
- Channel membership as first-pass filter (must be in #dark-factory to issue control commands)
- Per-command role check as second-pass
- Audit log as append-only JSONL file (mirrors ThreeDoors session log pattern, D-024)

**Role Matrix:**

| Action | Admin | Operator | Viewer |
|--------|-------|----------|--------|
| View supervisor status | Y | Y | Y |
| List workers | Y | Y | Y |
| View PR status | Y | Y | Y |
| Dispatch worker | Y | Y | N |
| Send agent message | Y | Y | N |
| Restart agent | Y | N | N |
| Stop supervisor | Y | N | N |
| Modify bot config | Y | N | N |
| View audit log | Y | Y | N |

**Rejected Alternatives:**
- **Slack workspace roles (admin/member/guest):** Too coarse — can't distinguish operator from viewer
- **OAuth scopes per user:** Over-engineered for team sizes < 20
- **External identity provider (LDAP/SAML):** Unnecessary complexity for a development tool

### 2.2 Multi-Supervisor Proxy

#### Architecture: Supervisor Registry with Heartbeat

```
┌──────────┐     ┌──────────────┐     ┌──────────────────┐
│  Slack   │────▶│   Slack Bot  │────▶│ Supervisor        │
│  User    │◀────│   (local)    │◀────│ Registry          │
└──────────┘     └──────────────┘     │                   │
                                      │ ThreeDoors: alive │
                                      │ OtherProj: alive  │
                                      │ FrontEnd:  stale  │
                                      └──────────────────┘
```

**Registration Protocol:**
1. Each bot instance registers its local supervisors on startup
2. Supervisors identified by: `{machine}:{repo}:{supervisor-name}`
3. Heartbeat: bot pings local multiclaude every 30 seconds, reports to Slack channel
4. Stale threshold: 2 missed heartbeats = warning, 5 = marked dead

**Command Routing:**
```
/mc-status ThreeDoors        → routes to ThreeDoors supervisor
/mc-dispatch ThreeDoors "implement story 42.1" → spawns worker
/mc-workers                  → aggregates from all supervisors
/mc-message ThreeDoors merge-queue "HEARTBEAT" → proxies message
```

**How it works locally:**
```bash
# Bot translates Slack command to multiclaude CLI
multiclaude worker list                              # /mc-workers
multiclaude message send supervisor "status request"  # /mc-status
multiclaude work "implement story 42.1"              # /mc-dispatch
```

### 2.3 Multi-Machine Support

#### Recommended: Bot-Per-Machine with Shared Slack App

```
┌─────────────────────────────────────────────────┐
│                 Slack Workspace                  │
│                                                  │
│  #dark-factory channel                           │
│  ┌─────────────────────────────────────────────┐ │
│  │ Bot Instance: skippy-mbp                    │ │
│  │   └─ ThreeDoors supervisor (alive)          │ │
│  │   └─ Switchboard supervisor (alive)         │ │
│  │                                             │ │
│  │ Bot Instance: skippy-workstation            │ │
│  │   └─ FrontendApp supervisor (alive)         │ │
│  │   └─ BackendAPI supervisor (stale)          │ │
│  └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

**Why bot-per-machine:**
- Each machine runs multiclaude locally — bot must be local to interact
- No SSH tunneling needed (simpler, more reliable, no key management)
- Each bot instance is a Slack Socket Mode client (no public URLs needed)
- Slack Socket Mode supports multiple connections to same app

**Conflict Avoidance:**
- Each bot instance registers with a unique machine ID on startup
- Commands that target a specific supervisor are routed only to the machine that owns it
- Global commands (e.g., `/mc-status`) — each bot responds with its local state, bot aggregates in a thread
- Use Slack's `app_id` + custom metadata to tag which instance handles which response

**Architecture Options Evaluated:**

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **A: Single bot + SSH** | Central control, one process | SSH key management, latency, firewall issues, single point of failure | Rejected |
| **B: Bot-per-machine, shared app** | Local execution, no network deps, resilient | Multiple processes, needs coordination | **Adopted** |
| **C: Bot-per-machine, separate channels** | Simple isolation | Channel sprawl, no unified view | Rejected |
| **D: Switchboard integration** | Leverages existing project | Switchboard is early-stage, adds dependency | Future option |

**Note on Switchboard (arcavenae/switchboard):** The tmux session router project could eventually provide the transport layer between machines, replacing Socket Mode for inter-bot coordination. But it's in early development (MVP scope is single-LAN E-router only). Worth revisiting when Switchboard matures.

### 2.4 General Bot Capabilities (Non-Dark-Factory)

#### Plugin Architecture

```
mc-slack-bot/
├── src/
│   ├── bot/              # Core bot framework
│   │   ├── __init__.py
│   │   ├── app.py        # Main Bolt app, routing
│   │   ├── rbac.py       # RBAC middleware
│   │   └── audit.py      # Audit logging (JSONL)
│   ├── plugins/          # Feature modules
│   │   ├── darkfactory/  # multiclaude control plane
│   │   │   ├── commands.py
│   │   │   ├── status.py
│   │   │   └── bridge.py
│   │   ├── chat/         # General conversation
│   │   ├── reminders/    # Team reminders
│   │   └── notifications/ # GitHub/CI notifications
│   └── config/           # Config models
│       ├── models.py     # Pydantic config models
│       └── loader.py     # YAML config loader
├── configs/
│   ├── roles.yaml
│   └── supervisors.yaml
├── pyproject.toml        # uv-managed project
└── Dockerfile
```

**Plugin Pattern (errbot-inspired, Bolt-native):**
```python
class Plugin:
    name: str
    def register(self, app: App) -> None: ...
    def commands(self) -> list[Command]: ...

class DarkFactoryPlugin(Plugin):
    name = "darkfactory"
    def register(self, app):
        app.command("/mc-status")(self.handle_status)
        app.command("/mc-dispatch")(self.handle_dispatch)
```

**Clean Separation:**
- Dark factory commands require RBAC (operator+) via Bolt middleware
- General features (chat, reminders) available to all workspace members
- Each plugin registers its own slash commands and event handlers
- Plugins loaded at startup from a registry list in config

---

## 3. Integration Architecture

### 3.1 Slack → multiclaude Bridge

**Slash Commands:**

| Command | Args | Role | Description |
|---------|------|------|-------------|
| `/mc-status` | `[supervisor]` | viewer | Show supervisor/agent status |
| `/mc-workers` | `[supervisor]` | viewer | List active workers |
| `/mc-dispatch` | `<supervisor> <task>` | operator | Spawn a worker |
| `/mc-message` | `<supervisor> <agent> <msg>` | operator | Send agent message |
| `/mc-agents` | `[supervisor]` | viewer | List persistent agents |
| `/mc-restart` | `<supervisor> <agent>` | admin | Restart an agent |
| `/mc-logs` | `<supervisor> [agent]` | operator | Tail agent logs |
| `/mc-pr` | `<supervisor>` | viewer | Show open PRs |

**Interactive Messages:**
- Worker completion → Slack message with "View PR" button
- CI failure → Slack message with "Dispatch Fix Worker" button
- Merge-queue holding → "Approve Merge" / "Block" buttons
- Human-gated decisions → @mention with approve/reject buttons

**Thread Model:**
- One thread per PR lifecycle (opened → CI → review → merge)
- One thread per worker lifecycle (spawned → working → PR created → completed)
- Summary thread for daily digest

### 3.2 multiclaude → Slack Bridge

**Event Routing:**

```
multiclaude event → bot bridge → Slack channel/thread
```

**Implementation Options:**

| Approach | How | Complexity | Verdict |
|----------|-----|------------|---------|
| **A: Log file watcher** | `fsnotify` on multiclaude JSONL logs | Low | **MVP** |
| **B: multiclaude hooks** | Shell hooks that call bot API | Medium | Phase 2 |
| **C: Named pipe / Unix socket** | Direct IPC | Medium | Phase 2 |
| **D: multiclaude native Slack** | PR to multiclaude | High | Future |

**MVP approach:** Bot watches multiclaude's JSONL session transcripts and state file for key events:
- Worker completion: detect `multiclaude agent complete` in logs
- PR creation: detect `gh pr create` output
- CI status: poll `gh run list` periodically
- Agent messages: poll `multiclaude message list`

**Events to Route:**

| Event | Channel | Priority |
|-------|---------|----------|
| PR merged | #mc-status | Normal |
| CI failed | #dark-factory | High (@mention) |
| Worker completed | #mc-status | Normal |
| Human decision needed | #dark-factory | High (@mention) |
| Agent crashed/stale | #dark-factory | High (@mention) |
| Daily summary | #mc-status | Low (scheduled) |

### 3.3 Security

- **Bot tokens:** Stored in macOS Keychain via Python `keyring` library (or env vars for Docker)
- **Slack tokens:** Socket Mode (no public webhook URLs, no inbound HTTP)
- **No secrets in messages:** Bot strips sensitive content from logs before posting
- **Inter-bot coordination:** Via Slack itself (no direct machine-to-machine needed in Phase 1)
- **Rate limiting:** Slack SDK handles rate limits; bot adds internal rate limit for multiclaude CLI calls (max 1 command/second per supervisor)
- **Audit trail:** All commands logged to `~/.slack-mc-bot/audit.jsonl` with Slack user ID, timestamp, command, result

---

## 4. Technology Recommendations

### Framework Evaluation

> **Note:** The user's prior "ourbot" was a simple Go bot experiment. Per human operator guidance, we evaluate mature frameworks across languages rather than defaulting to Go.

#### Option A: Python — Slack Bolt (Recommended)

**Framework:** [slack-bolt](https://github.com/slackapi/bolt-python) (Slack's official Python framework)

| Aspect | Details |
|--------|---------|
| **Maturity** | Official Slack SDK, actively maintained by Slack team |
| **Socket Mode** | First-class support, built-in |
| **Community** | Largest Slack bot ecosystem; most tutorials, examples, StackOverflow answers |
| **Plugin architecture** | `@app.command()`, `@app.action()`, `@app.event()` decorators make plugin modules trivial |
| **Async support** | Native asyncio support via `AsyncApp` |
| **LLM integration** | Best-in-class — Claude SDK, LangChain, etc. all Python-first |
| **Deployment** | `uv` for dependency management (per user's Python policy); Docker for production |
| **Weakness** | Not a single binary; requires Python runtime |

```python
# Example: how clean Bolt Python is
from slack_bolt import App
from slack_bolt.adapter.socket_mode import SocketModeHandler

app = App(token=os.environ["SLACK_BOT_TOKEN"])

@app.command("/mc-status")
def handle_status(ack, respond, command):
    ack()
    result = subprocess.run(["multiclaude", "status"], capture_output=True, text=True)
    respond(result.stdout)

SocketModeHandler(app, os.environ["SLACK_APP_TOKEN"]).start()
```

**Key libraries:**
- `slack-bolt` — Core framework
- `slack-sdk` — Lower-level Slack API access
- `anthropic` — Claude API for NL command parsing (Phase 3)
- `watchfiles` — File watching (Rust-backed, fast)
- `pydantic` — Config validation
- `sqlmodel` or `tinydb` — Lightweight storage if needed

#### Option B: JavaScript/TypeScript — Slack Bolt JS

**Framework:** [bolt-js](https://github.com/slackapi/bolt-js) (Slack's official JS framework)

| Aspect | Details |
|--------|---------|
| **Maturity** | Official Slack SDK, well-maintained |
| **Socket Mode** | Built-in |
| **Community** | Large; second to Python for Slack bots |
| **Plugin architecture** | Middleware-based, Express-like |
| **Async** | Native Promises/async-await |
| **LLM integration** | Good — Anthropic JS SDK, Vercel AI SDK |
| **Deployment** | Node.js runtime required; Docker common |
| **Weakness** | Node.js dependency management (npm/pnpm), heavier runtime |

#### Option C: Go — slack-go/slack

**Framework:** [slack-go/slack](https://github.com/slack-go/slack) (community-maintained)

| Aspect | Details |
|--------|---------|
| **Maturity** | Community project, well-maintained, 4.6k stars |
| **Socket Mode** | Supported but less ergonomic than Bolt |
| **Community** | Smaller than Python/JS for Slack specifically |
| **Plugin architecture** | Must build your own (no decorator pattern) |
| **Deployment** | Single binary — simplest deployment |
| **LLM integration** | Claude Go SDK exists but less mature |
| **Strength** | Ecosystem consistency with ThreeDoors, switchboard |
| **Weakness** | More boilerplate; no official Slack framework; bot-building patterns less established |

#### Option D: Open-Source Starter Bots

| Project | Language | Stars | Notes |
|---------|----------|-------|-------|
| [errbot](https://github.com/errbotio/errbot) | Python | 3k+ | Multi-platform bot framework (Slack, Discord, etc). Plugin system, RBAC, admin commands. Very mature but aging. |
| [Hubot](https://github.com/hubotio/hubot) | CoffeeScript/JS | 16k+ | GitHub's original bot framework. Venerable but showing age. |
| [opsdroid](https://github.com/opsdroid/opsdroid) | Python | 800+ | Event-driven with NLU support. Good for conversational bots. |
| [marvin](https://github.com/prefecthq/marvin) | Python | 5k+ | AI-first Python framework by Prefect. LLM-native but Slack is one of many connectors. |
| [dispatch](https://github.com/Netflix/dispatch) | Python | 4.5k+ | Netflix's incident management platform. Has Slack bot + RBAC + workflow orchestration. Heavy but proven at scale. |

**Best open-source starting point:** Netflix's Dispatch is the closest existing system to what's needed — it has RBAC, Slack integration, workflow orchestration, and plugin architecture. However, it's an incident management tool, not a bot framework, so it'd need significant adaptation. For a cleaner start, **Slack Bolt Python with errbot-inspired plugin patterns** gives the best foundation.

#### Recommendation: Python (Slack Bolt) with Plugin Architecture

**Why Python wins for this use case:**
1. **Official framework** — Slack Bolt Python is Slack's own; best-maintained, first to get new features
2. **LLM-native** — Phase 3 wants natural language commands; Python is the LLM ecosystem language
3. **Fastest to prototype** — Decorator-based handlers, built-in middleware, less boilerplate
4. **Plugin patterns proven** — errbot, Dispatch, opsdroid all demonstrate Python bot plugin architectures
5. **uv makes deployment clean** — `uv run bot` is a single command; `uvx` for development tools
6. **multiclaude integration is subprocess-based anyway** — Language doesn't matter for `exec("multiclaude ...")` calls

**Trade-off acknowledged:** Not a single binary like Go. Mitigated by Docker for production and `uv` for development.

### Database

**No database for MVP.** Use:
- YAML files for config and role mappings (Pydantic models for validation)
- JSONL for audit trail (append-only, consistent with ThreeDoors patterns)
- In-memory dict for supervisor registry (rebuilt on restart from config + heartbeat)

**Phase 2:** SQLite via `sqlmodel` or `tinydb` if query patterns emerge that need indexing.

### Deployment

```
Phase 1 (MVP):
  - uv-managed Python project
  - launchd plist on macOS (user-level, auto-restart)
  - Or: Docker container with uv

Phase 2:
  - Docker Compose for multi-service (bot + watchers)
  - systemd unit on Linux machines
  - Config sync via git repo (roles.yaml, supervisors.yaml)
```

---

## 5. Implementation Phases

### Phase 0: Proof of Concept (1-2 stories)
- Basic Slack bot in Go with Socket Mode
- `/mc-status` command that calls `multiclaude status` and posts result
- Single machine, single supervisor
- Hardcoded admin user

### Phase 1: Core Control Plane (3-4 stories)
- RBAC with YAML role config
- Full slash command set (/mc-status, /mc-workers, /mc-dispatch, /mc-message)
- Audit trail (JSONL)
- Log file watcher for basic event routing (worker completion, PR creation)
- Interactive messages for common actions

### Phase 2: Multi-Machine (2-3 stories)
- Multiple bot instances with shared Slack app
- Supervisor registry with heartbeat
- Command routing to specific machines
- Aggregated status views

### Phase 3: Rich Integration (2-3 stories)
- PR lifecycle threads
- CI status monitoring and notifications
- Daily/weekly summary posts
- General bot features (plugin architecture)

### Phase 4: Advanced (future)
- Switchboard integration for inter-bot transport
- Custom Slack workflows
- Slack canvas for live dashboards
- Mobile push for critical alerts

---

## 6. Open Questions (For Human Decision)

| ID | Question | Options | Recommendation |
|----|----------|---------|----------------|
| OQ-SB-1 | Where is the "ourbot" code? | GitLab? Other org? Fresh start? | User to confirm; fresh Go start recommended regardless |
| OQ-SB-2 | Should the bot live in its own repo or as a module in multiclaude? | Own repo (clean separation) vs multiclaude module (tight integration) | Own repo — bot has its own lifecycle and deployment |
| OQ-SB-3 | Should the bot use Slack Socket Mode or HTTP webhooks? | Socket Mode (no public URL) vs Webhooks (standard but needs endpoint) | Socket Mode — no infrastructure needed, works behind NAT |
| OQ-SB-4 | Which Slack workspace to use? | Existing workspace vs new dedicated workspace | Existing workspace, dedicated channels |
| OQ-SB-5 | Should Switchboard be considered for Phase 2 transport? | Yes (leverage existing Go project) vs No (keep separate) | Revisit when Switchboard has multi-hop support |
| OQ-SB-6 | Priority: should this be a ThreeDoors epic or a standalone project? | ThreeDoors epic (tracked in ROADMAP) vs Independent project | Independent project — it spans all multiclaude repos, not ThreeDoors-specific |
| OQ-SB-7 | Should the bot have Claude/LLM integration for natural language commands? | Yes (conversational control) vs No (structured commands only) | Phase 3+ — structured commands first, NL overlay later |

---

## 7. Decisions Summary

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| Language/Framework | Python (Slack Bolt) | Go (slack-go/slack), JavaScript (Bolt JS), Rust | Official Slack framework, LLM-native ecosystem, fastest to prototype, best plugin patterns (errbot, Dispatch) |
| Multi-machine architecture | Bot-per-machine, shared Slack app | Single bot + SSH, separate channels per machine | Local execution, no SSH key management, resilient |
| Slack connection | Socket Mode | HTTP webhooks | No public URL needed, works behind NAT, simpler |
| RBAC storage | YAML config files | Database, Slack workspace roles, OAuth | Simple, version-controllable, sufficient for team size |
| Audit trail | JSONL append-only | Database, Slack message history | Consistent with ThreeDoors patterns (D-024) |
| Event bridge (MVP) | Log file watcher (fsnotify) | multiclaude hooks, Unix socket, native integration | Lowest complexity, no multiclaude changes needed |
| Secret storage | macOS Keychain (keyring library) | Environment variables, dotenv files | Secure credential storage, cross-platform |
| Plugin architecture | Registry-based modules (errbot-inspired) | Dynamic plugins, separate processes | Clean separation, loaded at startup from config |
| Database (MVP) | None (YAML + JSONL) | SQLite, PostgreSQL | YAGNI; add SQLite in Phase 2 if needed |
| Deployment (MVP) | launchd plist (macOS) | Docker, systemd | Primary dev machine is macOS; expand later |
