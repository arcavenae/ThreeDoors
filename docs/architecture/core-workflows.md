# Core Workflows

## Workflow 1: Application Startup & Three Doors Display

```mermaid
sequenceDiagram
    actor User
    participant Main as cmd/main.go
    participant FM as FileManager
    participant TP as TaskPool
    participant DV as DoorsView
    participant DS as DoorSelector

    User->>Main: Launch `threedoors`
    Main->>FM: LoadTasks()

    alt tasks.yaml exists
        FM->>FM: Read tasks.yaml
        FM->>FM: Unmarshal YAML
        FM->>TP: Build TaskPool from tasks
    else tasks.yaml missing
        FM->>FM: InitializeFiles()
        FM->>FM: Create sample tasks.yaml
        FM->>TP: Build TaskPool with samples
    end

    FM-->>Main: Return TaskPool
    Main->>DV: NewDoorsView(taskPool)
    DV->>DS: SelectDoors(taskPool, 3)
    DS->>TP: GetAvailableForDoors()
    TP-->>DS: [task1, task2, task3, ...]
    DS->>DS: Random selection (3 tasks)
    DS->>TP: MarkRecentlyShown(task IDs)
    DS-->>DV: DoorSelection{doors: [t1, t2, t3]}
    DV->>DV: Render Three Doors (horizontally, dynamically sized, no initial selection)
    DV-->>User: Display doors with status indicators
```

## Workflow 2: Select Door & Enter Task Detail View

```mermaid
sequenceDiagram
    actor User
    participant DV as DoorsView
    participant MM as MainModel
    participant TDV as TaskDetailView
    participant Task as Task Model

    User->>DV: Press "W" or "Up Arrow" (select Door 2)
    DV->>MM: Send SelectDoorMsg{doorIndex: 1}
    MM->>MM: Switch view mode to "detail"
    MM->>DV: GetTask(doorIndex: 1)
    DV-->>MM: Return task pointer
    MM->>TDV: NewTaskDetailView(task, taskPool)
    TDV->>Task: Read task data
    Task-->>TDV: Task details
    TDV->>TDV: Render detail view
    TDV-->>User: Display task details with options menu
```

## Workflow 3: Update Task Status

```mermaid
sequenceDiagram
    actor User
    participant TDV as TaskDetailView
    participant SM as StatusMenu
    participant SMgr as StatusManager
    participant Task as Task
    participant FM as FileManager

    User->>TDV: Press "S" (update status)
    TDV->>SM: NewStatusUpdateMenu(task.Status)
    SM->>SMgr: GetValidTransitions(currentStatus)
    SMgr-->>SM: [valid status options]
    SM-->>User: Display status options

    User->>SM: Select status + Enter
    SM->>SMgr: ValidateTransition(old, new)

    alt Valid transition
        SMgr-->>SM: OK
        SM-->>TDV: Return (newStatus, confirmed=true)
        TDV->>Task: UpdateStatus(newStatus, "")
        Task->>Task: Set status, update timestamp
        TDV->>FM: SaveTasks(taskPool)
        FM->>FM: Atomic write to tasks.yaml
        TDV-->>User: Show updated status
    else Invalid transition
        SMgr-->>SM: Error
        SM-->>User: Show error message
    end
```

## Workflow 4: Add Progress Note

```mermaid
sequenceDiagram
    actor User
    participant TDV as TaskDetailView
    participant NI as NotesInputView
    participant Task as Task
    participant FM as FileManager

    User->>TDV: Press "N" (add note)
    TDV->>NI: NewNotesInputView()
    NI-->>User: Display note input box

    User->>NI: Type note text
    NI-->>User: Show typed text + char count

    User->>NI: Press Ctrl+S (save)
    NI-->>TDV: Return (noteText, confirmed=true)
    TDV->>Task: AddNote(noteText)
    Task->>Task: Append to notes array
    TDV->>FM: SaveTasks(taskPool)
    FM-->>TDV: Success
    TDV-->>User: Show updated notes list
```

## Workflow 5: Complete Task & Return to Doors

```mermaid
sequenceDiagram
    actor User
    participant TDV as TaskDetailView
    participant Task as Task
    participant TP as TaskPool
    participant FM as FileManager
    participant MM as MainModel
    participant DV as DoorsView

    User->>TDV: Select status=COMPLETE
    TDV->>Task: UpdateStatus(StatusComplete, "")
    Task->>Task: Set status=complete, completedAt=now()

    TDV->>FM: AppendCompleted(task)
    FM->>FM: Atomic append to completed.txt

    TDV->>TP: RemoveTask(task.ID)
    Note over TP: Task removed from active pool

    TDV->>FM: SaveTasks(taskPool)
    FM-->>TDV: Success

    User->>TDV: Press ESC (return to doors)
    TDV->>MM: Send ReturnToDoorsMsg
    MM->>DV: Refresh doors
    DV-->>User: Show new three doors (completed task removed)
```

## Workflow 6: Mark Task as Blocked

```mermaid
sequenceDiagram
    actor User
    participant TDV as TaskDetailView
    participant BI as BlockerInput
    participant Task as Task
    participant FM as FileManager

    User->>TDV: Press "B" (mark blocked)
    TDV->>BI: NewBlockerInputView()
    BI-->>User: Display blocker input box

    User->>BI: Type blocker reason + Enter
    BI-->>TDV: Return (blockerText, confirmed=true)

    TDV->>Task: UpdateStatus(StatusBlocked, "")
    TDV->>Task: SetBlocker(blockerText)
    Task->>Task: Set blocker field, update timestamp

    TDV->>FM: SaveTasks(taskPool)
    FM-->>TDV: Success
    TDV-->>User: Show status=BLOCKED + blocker reason
```

## Post-MVP Workflows (Phase 2–3)

### Workflow 7: Multi-Provider Startup & Aggregation

```mermaid
sequenceDiagram
    actor User
    participant Main as cmd/main.go
    participant Cfg as ConfigLoader
    participant Reg as AdapterRegistry
    participant TP as TextFileAdapter
    participant OA as ObsidianAdapter
    participant Agg as MultiSourceAggregator
    participant DD as DuplicateDetector
    participant Pool as Unified TaskPool
    participant SE as SyncEngine

    User->>Main: Launch `threedoors`
    Main->>Cfg: LoadConfig(~/.threedoors/config.yaml)
    Cfg-->>Main: Config (providers, calendar, llm, sync)

    alt First run (no config)
        Main->>Main: Launch OnboardingView
        Note over Main: User completes onboarding wizard
    end

    Main->>Reg: NewAdapterRegistry(config)
    Reg->>TP: Initialize TextFileAdapter
    Reg->>OA: Initialize ObsidianAdapter

    loop For each enabled provider
        Reg->>Reg: provider.HealthCheck()
    end

    Main->>Agg: Aggregate()
    Agg->>Reg: LoadAll()
    Reg->>TP: Load()
    TP-->>Reg: [tasks from text file]
    Reg->>OA: Load()
    OA-->>Reg: [tasks from Obsidian]
    Reg-->>Agg: [all tasks with source tags]

    Agg->>DD: DetectDuplicates(tasks)
    DD-->>Agg: [tasks with dup flags]
    Agg->>Pool: Build unified TaskPool

    Main->>SE: NewSyncEngine(registry)
    SE->>SE: Replay any queued offline changes
    SE->>SE: Start background watch goroutines
```

### Workflow 8: Sync with Conflict Detection

```mermaid
sequenceDiagram
    participant SE as SyncEngine
    participant Q as OfflineQueue
    participant OA as ObsidianAdapter
    participant CR as ConflictResolver
    participant SL as SyncLog
    participant TUI as SyncStatusBar

    Note over SE: Background sync loop

    SE->>OA: Watch() - listen for external changes
    OA-->>SE: ChangeEvent{type: Updated, task: T1}

    SE->>Q: Check if T1 has pending local changes

    alt No local conflict
        SE->>SE: Apply remote change to TaskPool
        SE->>SL: Log("obsidian", "remote update applied", T1)
        SE->>TUI: SyncStatusMsg{provider: "obsidian", status: Connected}
    else Conflict detected
        SE->>CR: ResolveConflict(localChange, remoteChange)
        CR->>CR: Apply strategy (last-write-wins or surface to user)
        CR-->>SE: Resolution

        alt Auto-resolved
            SE->>SL: Log("obsidian", "conflict auto-resolved", T1)
        else Manual resolution needed
            SE->>TUI: ConflictMsg{task: T1, local: ..., remote: ...}
            Note over TUI: User sees conflict visualization
        end
    end
```

### Workflow 9: LLM Task Decomposition

```mermaid
sequenceDiagram
    actor User
    participant TDV as TaskDetailView
    participant LLM as LLMTaskDecomposer
    participant Backend as LLM Backend
    participant Git as Git Repository

    User->>TDV: Press "L" (decompose task)
    TDV->>TDV: Confirm action dialog

    TDV->>LLM: Decompose(task)
    LLM->>LLM: Build prompt (task text + context + instructions)
    LLM->>Backend: Complete(prompt)

    alt Local backend (Ollama)
        Backend->>Backend: POST http://localhost:11434/api/generate
    else Cloud backend (Anthropic)
        Backend->>Backend: POST api.anthropic.com/v1/messages
    end

    Backend-->>LLM: Generated story specs (structured output)
    LLM->>LLM: Parse and validate story specs

    LLM->>Git: OutputToGit(specs, repoPath)
    Git->>Git: Write story files to repo structure
    Git->>Git: git add + git commit

    LLM-->>TDV: [StorySpec, StorySpec, ...]
    TDV-->>User: Show decomposition results + git paths
```

### Workflow 10: Calendar-Aware Door Selection

```mermaid
sequenceDiagram
    participant DS as DoorSelector
    participant LE as LearningEngine
    participant CR as CalendarReader
    participant Pool as TaskPool

    DS->>CR: GetFreeBlocks(now, endOfDay)
    CR->>CR: Read macOS Calendar via AppleScript
    CR-->>DS: [TimeBlock{30min}, TimeBlock{2hr}]

    DS->>LE: GetPatterns(userHistory)
    LE-->>DS: UserPatterns{preferredTypes, moodCorrelations}

    DS->>Pool: GetAvailableForDoors()
    Pool-->>DS: [available tasks]

    DS->>DS: Score tasks by:<br/>1. Time fit (task duration vs available blocks)<br/>2. Historical preference (learning patterns)<br/>3. Mood context (current mood state)<br/>4. Diversity (vary task types)

    DS->>DS: Select top 3 diverse tasks
    DS-->>DS: DoorSelection{doors: [t1, t2, t3]}
```

## Future Task Management Workflows (Architectural Considerations)

The application now includes key bindings for several task management actions that are currently unimplemented: 'c' (complete), 'b' (blocked), 'i' (in progress), 'e' (expand), 'f' (fork), 'p' (procrastinate). These actions represent significant future functionality and will require dedicated workflows and architectural considerations.

**Architectural Implications:**

*   **State Management:** Each action will likely involve updating the state of a `Task` object (e.g., `StatusComplete`, `StatusBlocked`, `StatusInProgress`). This will require robust state transition logic, potentially leveraging the `StatusManager` component.
*   **Persistence:** Changes to task status or properties will need to be persisted. This will involve interactions with the `FileManager` to save updated `TaskPool` data.
*   **Task Expansion/Forking:** The 'e' (expand) and 'f' (fork) actions imply the creation of new tasks or the modification of existing ones. This will require logic to generate new task IDs, potentially split existing task content, and integrate these new tasks into the `TaskPool`. This could have implications for how tasks are uniquely identified and managed.
*   **Procrastination:** The 'p' (procrastinate) action might involve deferring a task, potentially moving it to a different pool or marking it with a future start date. This could introduce new scheduling or prioritization logic.
*   **User Feedback:** Implementing these actions will require clear visual feedback to the user about the success or failure of the operation.

These future workflows will need detailed design and story breakdown in subsequent development phases, with careful consideration of their impact on the existing `Task` model, `TaskPool`, and persistence mechanisms.

---
