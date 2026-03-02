# Data Models

## TaskStatus (Enum)

**Purpose:** Defines the lifecycle states a task can be in.

**Values:**
```go
type TaskStatus string

const (
    StatusTodo       TaskStatus = "todo"
    StatusBlocked    TaskStatus = "blocked"
    StatusInProgress TaskStatus = "in-progress"
    StatusInReview   TaskStatus = "in-review"
    StatusComplete   TaskStatus = "complete"
)
```

**Valid Transitions:**
```
todo → in-progress → in-review → complete
todo → blocked → in-progress
in-progress → blocked → in-progress
blocked → todo (unblock)
Any state → complete (force complete)
```

**State Machine Diagram:**

```mermaid
stateDiagram-v2
    [*] --> todo
    todo --> in-progress
    todo --> blocked
    todo --> complete
    blocked --> todo
    blocked --> in-progress
    blocked --> complete
    in-progress --> blocked
    in-progress --> in-review
    in-progress --> complete
    in-review --> in-progress
    in-review --> complete
    complete --> [*]
```

## Task

**Purpose:** Represents a single task with full lifecycle metadata including status, notes, and history.

**Key Attributes:**
- `id`: `string` - Unique identifier (UUID v4)
- `text`: `string` - Task description (1-500 chars, UTF-8)
- `status`: `TaskStatus` - Current lifecycle state (default: "todo")
- `notes`: `[]TaskNote` - Progress notes and updates
- `blocker`: `string` - Blocker description (only when status=blocked)
- `createdAt`: `time.Time` - When task was created
- `updatedAt`: `time.Time` - Last modification timestamp
- `completedAt`: `*time.Time` - When marked complete (nil if not complete)

**Validation Rules:**
1. `id`: Must be valid UUID v4
2. `text`: 1-500 chars, no newlines/tabs, trimmed whitespace
3. `status`: Must be one of valid TaskStatus values
4. `notes`: Can be empty array
5. `blocker`: Required non-empty when status=blocked, otherwise empty
6. `createdAt`: Required, cannot be zero
7. `updatedAt`: Required, >= createdAt
8. `completedAt`: Only set when status=complete

**YAML Storage Format (tasks.yaml):**

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

**Key Methods:**
- `NewTask(text string) *Task` - Create new task with defaults
- `Validate() error` - Validate all fields
- `UpdateStatus(newStatus TaskStatus, note string) error` - Change status with validation
- `AddNote(text string)` - Append progress note
- `SetBlocker(reason string) error` - Set blocker when status=blocked
- `IsValidTransition(newStatus TaskStatus) bool` - Check if transition is allowed

**UUID Generation:**
The task UUID is generated **immediately** in the `NewTask()` constructor using `uuid.New().String()`, ensuring the task has a unique identity before any persistence occurs. The UUID is immutable and never changes after task creation.

## TaskNote

**Purpose:** Captures progress updates and notes added to a task over time.

**Key Attributes:**
- `timestamp`: `time.Time` - When note was added (UTC)
- `text`: `string` - Note content (1-1000 chars)

**Design Decisions:**
- **Immutable:** Notes cannot be edited/deleted once added (append-only history)
- **Chronological:** Array order preserves time sequence
- **Longer text limit:** Notes can be more detailed than task text

## Config

**Purpose:** Centralized configuration for file paths and runtime limits.

**Key Attributes:**
- `baseDir`: `string` - Base directory for all ThreeDoors data (default: `~/.threedoors/`)
- `tasksPath`: `string` - Path to tasks file (default: `{baseDir}/tasks.yaml`)
- `completedPath`: `string` - Path to completed tasks file (default: `{baseDir}/completed.txt`)
- `maxTasks`: `int` - Maximum tasks to load (default: 1000)
- `maxTaskLength`: `int` - Maximum characters per task text (default: 500)
- `maxNoteLength`: `int` - Maximum characters per note (default: 1000)
- `recentlyShownWindow`: `int` - Ring buffer size (default: 10)

## TaskPool

**Purpose:** Manages in-memory collection of tasks filtered by status, with smart door selection.

**Key Attributes:**
- `tasks`: `map[string]*Task` - All tasks indexed by ID
- `recentlyShown`: `[]string` - Last N task IDs shown in doors (ring buffer)
- `recentlyShownIdx`: `int` - Current position in ring buffer
- `maxRecentlyShown`: `int` - Size of ring buffer (default: 10)

**Key Methods:**
- `AddTask(task *Task)` - Add task to pool
- `GetTask(id string) *Task` - Retrieve task by ID
- `UpdateTask(task *Task)` - Update existing task
- `RemoveTask(id string)` - Remove completed task from pool
- `GetTasksByStatus(status TaskStatus) []*Task` - Filter tasks by status
- `GetAvailableForDoors() []*Task` - Get tasks eligible for door selection
- `MarkRecentlyShown(taskID string)` - Add to recently shown buffer
- `IsRecentlyShown(taskID string) bool` - Check if in recently shown

**Door Selection Eligibility:**
- Status must be `todo`, `blocked`, or `in-progress` (exclude in-review and complete)
- Not in `recentlyShown` buffer (unless fewer than 3 tasks total)

## DoorSelection

**Purpose:** Represents the three tasks currently displayed as doors.

**Key Attributes:**
- `doors`: `[]*Task` - 0-3 tasks with full metadata
- `selectedIndex`: `int` - Which door selected (-1 if none)
- `selectedTask`: `*Task` - Pointer to selected task (nil if none)

## Data Model Diagram (Tech Demo)

```mermaid
erDiagram
    CONFIG {
        string baseDir
        string tasksPath
        string completedPath
        int maxTasks
        int maxTaskLength
        int maxNoteLength
        int recentlyShownWindow
    }

    TASK {
        string id PK
        string text
        TaskStatus status
        string blocker
        datetime createdAt
        datetime updatedAt
        datetime completedAt
    }

    TASK_NOTE {
        datetime timestamp
        string text
    }

    TASK_POOL {
        map tasks
        array recentlyShown
        int recentlyShownIdx
    }

    DOOR_SELECTION {
        array doors
        int selectedIndex
        Task selectedTask
    }

    COMPLETED_LOG {
        datetime timestamp
        string taskId
        string taskText
    }

    TASK ||--o{ TASK_NOTE : "has many"
    TASK_POOL ||--o{ TASK : "contains"
    DOOR_SELECTION ||--o{ TASK : "selects"
    TASK ||--o| COMPLETED_LOG : "archives to"
```

---

## Post-MVP Data Models (Phase 2–3)

### ProviderConfig

**Purpose:** Defines a configured task provider in `config.yaml`.

**Key Attributes:**
- `name`: `string` — Provider identifier (e.g., "textfile", "applenotes", "obsidian")
- `type`: `string` — Provider type (maps to adapter implementation)
- `enabled`: `bool` — Whether this provider is active
- `settings`: `map[string]any` — Provider-specific configuration

**Example (config.yaml):**
```yaml
providers:
  - name: textfile
    type: textfile
    enabled: true
    settings:
      path: ~/.threedoors/tasks.yaml

  - name: obsidian
    type: obsidian
    enabled: true
    settings:
      vault_path: ~/Documents/ObsidianVault
      tasks_folder: Tasks
      daily_notes: true
      daily_notes_folder: Daily

  - name: applenotes
    type: applenotes
    enabled: false
    settings:
      folder: ThreeDoors Tasks
```

### Task (Extended)

**Post-MVP additions to the Task model:**

- `source`: `string` — Provider name that owns this task (e.g., "obsidian", "applenotes")
- `sourceID`: `string` — Provider-specific identifier (e.g., Obsidian file path, Apple Notes record ID)
- `duplicateOf`: `*string` — If flagged as duplicate, points to canonical task ID
- `tags`: `[]string` — User or provider-assigned tags (for categorization)
- `estimatedDuration`: `*time.Duration` — Estimated time to complete (for calendar-aware selection)

### ChangeEvent

**Purpose:** Represents a change detected from a provider or queued locally.

**Key Attributes:**
- `id`: `string` — Unique event ID (UUID)
- `timestamp`: `time.Time` — When the change occurred
- `provider`: `string` — Source provider name
- `type`: `ChangeType` — Created, Updated, Deleted
- `taskID`: `string` — Affected task ID
- `task`: `*Task` — Full task data (nil for Deleted)
- `replayed`: `bool` — Whether this event has been synced

### SyncState

**Purpose:** Tracks per-provider sync status.

**Key Attributes:**
- `provider`: `string` — Provider name
- `lastSyncAt`: `time.Time` — Last successful sync timestamp
- `status`: `SyncStatus` — Connected, Syncing, Offline, Error
- `pendingChanges`: `int` — Number of queued changes awaiting replay
- `lastError`: `string` — Last error message (empty if healthy)

### CalendarEvent

**Purpose:** Represents a calendar event read from local sources.

**Key Attributes:**
- `title`: `string` — Event title
- `start`: `time.Time` — Start time
- `end`: `time.Time` — End time
- `allDay`: `bool` — Whether it's an all-day event

### TimeBlock

**Purpose:** Represents a free time block between calendar events.

**Key Attributes:**
- `start`: `time.Time` — Block start
- `end`: `time.Time` — Block end
- `duration`: `time.Duration` — Available time

### StorySpec (LLM Output)

**Purpose:** Represents an LLM-generated story/spec from task decomposition.

**Key Attributes:**
- `title`: `string` — Story title
- `description`: `string` — Story description with acceptance criteria
- `parentTaskID`: `string` — Task that was decomposed
- `filePath`: `string` — Output file path in git repo
- `format`: `string` — Output format (e.g., "bmad-story", "quick-spec")

### EnrichmentRecord

**Purpose:** Metadata stored in SQLite enrichment DB that cannot live in source systems.

**Key Attributes:**
- `taskID`: `string` — FK to task
- `provider`: `string` — Source provider
- `categories`: `[]string` — Categorization tags
- `learningPatterns`: `map[string]any` — Door selection history, mood correlations
- `dupDecisions`: `[]DupDecision` — Duplicate resolution history
- `crossRefs`: `[]string` — Related task IDs across providers

### Post-MVP Data Model Diagram

```mermaid
erDiagram
    CONFIG_YAML {
        array providers
        object onboarding
        object llm
        object calendar
    }

    PROVIDER_CONFIG {
        string name PK
        string type
        bool enabled
        map settings
    }

    TASK_EXTENDED {
        string id PK
        string text
        TaskStatus status
        string source
        string sourceID
        string duplicateOf
        array tags
        duration estimatedDuration
    }

    CHANGE_EVENT {
        string id PK
        datetime timestamp
        string provider
        ChangeType type
        string taskID
        bool replayed
    }

    SYNC_STATE {
        string provider PK
        datetime lastSyncAt
        SyncStatus status
        int pendingChanges
        string lastError
    }

    CALENDAR_EVENT {
        string title
        datetime start
        datetime end
        bool allDay
    }

    ENRICHMENT_RECORD {
        string taskID PK
        string provider
        array categories
        map learningPatterns
        array crossRefs
    }

    STORY_SPEC {
        string title
        string description
        string parentTaskID
        string filePath
    }

    CONFIG_YAML ||--o{ PROVIDER_CONFIG : "defines"
    PROVIDER_CONFIG ||--o{ TASK_EXTENDED : "provides"
    TASK_EXTENDED ||--o{ CHANGE_EVENT : "generates"
    PROVIDER_CONFIG ||--|| SYNC_STATE : "has"
    TASK_EXTENDED ||--o| ENRICHMENT_RECORD : "enriched by"
    TASK_EXTENDED ||--o{ STORY_SPEC : "decomposed into"
```

---
