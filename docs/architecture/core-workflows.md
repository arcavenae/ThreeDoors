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
    DV->>DV: Render Three Doors (horizontally)
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

    User->>DV: Press "W" (select Door 2)
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

---
