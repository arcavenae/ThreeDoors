# Data Storage Schema

## tasks.yaml Schema

**Location:** `~/.threedoors/tasks.yaml`

**Format:** YAML with strict schema validation

**Root Structure:**
```yaml
tasks:
  - # Array of Task objects
```

**Task Object Schema:**

```yaml
id: string                 # UUID v4, required
text: string              # 1-500 chars, required
status: string            # Enum: todo|blocked|in-progress|in-review|complete, required
notes:                    # Array of TaskNote objects, can be empty
  - timestamp: datetime   # RFC3339 format, required
    text: string          # 1-1000 chars, required
blocker: string           # Empty or 1-500 chars, required when status=blocked
created_at: datetime      # RFC3339 format, required
updated_at: datetime      # RFC3339 format, required, >= created_at
completed_at: datetime    # RFC3339 format, nullable, only when status=complete
```

**Example:**
```yaml
tasks:
  - id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
    text: Write architecture document for ThreeDoors
    status: in-progress
    notes:
      - timestamp: 2025-11-07T14:15:00Z
        text: Started with high-level overview
      - timestamp: 2025-11-07T14:45:00Z
        text: Completed data models section
    blocker: ""
    created_at: 2025-11-07T10:00:00Z
    updated_at: 2025-11-07T14:45:00Z
    completed_at: null
```

**Validation Rules:**
1. All timestamps in UTC (RFC3339 format)
2. Task IDs must be unique across all tasks
3. Empty blocker field when status != blocked
4. completedAt must be null unless status == complete
5. notes array preserves chronological order (newest last)

## completed.txt Schema

**Location:** `~/.threedoors/completed.txt`

**Format:** Plain text, append-only log

**Line Format:**
```
[YYYY-MM-DD HH:MM:SS] task_id | task_text
```

**Example:**
```
[2025-11-07 14:32:15] a1b2c3d4-e5f6-7890-abcd-ef1234567890 | Write architecture document for ThreeDoors
[2025-11-07 14:45:03] b2c3d4e5-f6a7-8901-bcde-f12345678901 | Implement Story 1.1 - Project Setup
```

---

## Post-MVP Storage (Phase 2–3)

### config.yaml Schema

**Location:** `~/.threedoors/config.yaml`

**Format:** YAML configuration file

**Schema:**
```yaml
# Provider configuration
providers:
  - name: string           # Unique provider name, required
    type: string           # Provider type: textfile|applenotes|obsidian, required
    enabled: bool          # Whether active, required
    settings:              # Provider-specific settings, varies by type
      # textfile settings:
      path: string         # Path to tasks.yaml

      # obsidian settings:
      vault_path: string   # Path to Obsidian vault
      tasks_folder: string # Folder within vault for tasks
      daily_notes: bool    # Enable daily note integration
      daily_notes_folder: string # Daily notes folder name

      # applenotes settings:
      folder: string       # Apple Notes folder name

# Onboarding state
onboarding:
  complete: bool           # Whether onboarding has been completed
  values: [string]         # User-defined values
  goals: [string]          # User-defined goals

# Calendar configuration
calendar:
  enabled: bool            # Whether calendar awareness is active
  sources:                 # Calendar sources (local-only, no OAuth)
    - type: string         # applescript|ics|caldav_cache
      path: string         # File path (for ics/caldav_cache types)

# LLM configuration
llm:
  enabled: bool            # Whether LLM decomposition is available
  backend: string          # local|cloud
  local:
    endpoint: string       # Ollama/llama.cpp endpoint (e.g., http://localhost:11434)
    model: string          # Model name (e.g., llama3.2)
  cloud:
    provider: string       # anthropic|openai
    model: string          # Model ID (e.g., claude-sonnet-4-20250514)
    # API key stored in OS keychain, NOT in config file
  output:
    repo_path: string      # Git repo path for story output
    format: string         # bmad-story|quick-spec

# Sync configuration
sync:
  conflict_strategy: string # last-write-wins|manual
  log_max_size_mb: int     # Max sync log size before rotation (default: 10)
```

### enrichment.db Schema (SQLite)

**Location:** `~/.threedoors/enrichment.db`

**Format:** SQLite database for metadata that cannot live in source systems

**Tables:**

```sql
-- Cross-provider enrichment metadata
CREATE TABLE enrichment (
    task_id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    categories TEXT,          -- JSON array of category strings
    learning_data TEXT,       -- JSON object of pattern data
    cross_refs TEXT,          -- JSON array of related task IDs
    created_at TEXT NOT NULL, -- RFC3339
    updated_at TEXT NOT NULL  -- RFC3339
);

-- Duplicate detection decisions
CREATE TABLE dup_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id_a TEXT NOT NULL,
    task_id_b TEXT NOT NULL,
    decision TEXT NOT NULL,   -- 'duplicate' | 'distinct'
    decided_at TEXT NOT NULL, -- RFC3339
    UNIQUE(task_id_a, task_id_b)
);

-- Sync state per provider
CREATE TABLE sync_state (
    provider TEXT PRIMARY KEY,
    last_sync_at TEXT,        -- RFC3339
    status TEXT NOT NULL,     -- 'connected' | 'syncing' | 'offline' | 'error'
    pending_changes INTEGER DEFAULT 0,
    last_error TEXT
);
```

### sync-state/ Directory

**Location:** `~/.threedoors/sync-state/`

**Files:**

- `queue.jsonl` — Offline change queue (append-only, entries removed after replay)
  ```jsonl
  {"id":"uuid","timestamp":"2026-03-02T10:00:00Z","provider":"obsidian","type":"updated","task_id":"abc123","task":{...}}
  ```

- `sync.log` — Human-readable sync debug log (rotating, max 10MB)
  ```
  [2026-03-02 10:00:00] [obsidian] Sync started
  [2026-03-02 10:00:01] [obsidian] 3 tasks loaded, 1 change detected
  [2026-03-02 10:00:01] [obsidian] Sync complete (1.2s)
  ```

### Updated User Data Directory Structure

```
~/.threedoors/
├── config.yaml              # User configuration (providers, calendar, LLM, sync)
├── tasks.yaml               # Text file adapter task storage (Phase 1+)
├── completed.txt            # Completed task log (Phase 1+)
├── metrics.jsonl            # Session metrics (Phase 1+)
├── enrichment.db            # SQLite enrichment database (Phase 2+)
└── sync-state/              # Sync engine state (Phase 3+)
    ├── queue.jsonl           # Offline change queue
    └── sync.log              # Sync debug log
```

---
