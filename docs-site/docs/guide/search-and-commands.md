# Search & Commands

## Search Mode

Press ++slash++ to search tasks by text. Results filter in real-time as you type.

| Key | Action |
|-----|--------|
| ++j++ or ++down++ | Navigate to next result |
| ++k++ or ++up++ | Navigate to previous result |
| ++enter++ | Open selected task in detail view |
| ++esc++ | Exit search |

!!! tip
    After viewing a task from search, you return to search results (not the doors view) so you can continue browsing.

## Command Palette

Press ++colon++ to enter command mode. Type a command and press ++enter++.

### Task Commands

| Command | Description |
|---------|-------------|
| `:add <text>` | Add a new task. Without text, opens a prompt. |
| `:add-ctx <text>` | Add a task with context — prompts for task text, then why/context. |
| `:add --why <text>` | Same as `:add-ctx` — captures task and reason. |
| `:tag` | Categorize the selected task (type, effort, location). |
| `:deferred` | Show deferred/snoozed tasks. |

### Insights & Analytics

| Command | Description |
|---------|-------------|
| `:stats` | Show session metrics (tasks completed, duration, refreshes). |
| `:dashboard` | Open the insights dashboard. |
| `:insights` | Same as `:dashboard`. Accepts optional filter: `:insights mood` or `:insights avoidance`. |
| `:mood [mood]` | Log mood. Without argument, opens the mood selector. |

### Configuration & Data

| Command | Description |
|---------|-------------|
| `:goals` | View your values and goals. |
| `:goals edit` | Edit your values and goals. |
| `:theme` | Open theme picker. |
| `:seasonal` | Open seasonal theme picker. |
| `:health` | Run system health checks on providers and data files. |
| `:synclog` | Show sync history. |
| `:connect` | Connect a new data source. |
| `:sources` | View connected data sources. |

### Other

| Command | Description |
|---------|-------------|
| `:suggestions` | Browse LLM task proposals. |
| `:devqueue` | Open dev dispatch queue. |
| `:plan` | Start daily planning mode. |
| `:help` | Display all available commands. |
| `:quit` / `:exit` | Exit the application. |
