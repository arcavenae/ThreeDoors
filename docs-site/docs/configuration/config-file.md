# Config File

ThreeDoors is configured via `~/.threedoors/config.yaml`. This page documents the complete schema.

---

## Simple Mode (Single Provider)

For a single task source, use the flat format:

```yaml
provider: textfile
note_title: ThreeDoors Tasks
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | `textfile` | Provider name: `textfile`, `applenotes`, `obsidian`, `jira`, `github`, `reminders`, `todoist` |
| `note_title` | string | — | Note title for Apple Notes provider |
| `theme` | string | `classic` | Door theme: `classic`, `modern`, `scifi`, `shoji` |

---

## Advanced Mode (Multiple Providers)

For multiple task sources, use the `providers` array:

```yaml
schema_version: 2
theme: classic

providers:
  - name: textfile
    settings:
      task_file: ~/.threedoors/tasks.yaml

  - name: applenotes
    settings:
      note_title: ThreeDoors Tasks
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `schema_version` | int | `2` | Config schema version |
| `theme` | string | `classic` | Door theme |
| `providers` | array | — | List of provider configurations |
| `providers[].name` | string | — | Provider identifier |
| `providers[].settings` | map | — | Provider-specific key-value settings |

---

## Provider Settings Reference

### `textfile`

```yaml
- name: textfile
  settings:
    task_file: ~/.threedoors/tasks.yaml
```

| Setting | Default | Description |
|---------|---------|-------------|
| `task_file` | `~/.threedoors/tasks.yaml` | Path to the YAML task file |

### `applenotes`

```yaml
- name: applenotes
  settings:
    note_title: ThreeDoors Tasks
```

| Setting | Default | Description |
|---------|---------|-------------|
| `note_title` | — | Exact title of the Apple Notes note to sync with |

### `obsidian`

```yaml
- name: obsidian
  settings:
    vault_path: /path/to/your/vault
    tasks_folder: tasks
    file_pattern: "*.md"
    daily_notes: true
    daily_notes_folder: Daily
    daily_notes_heading: "## Tasks"
    daily_notes_format: "2006-01-02.md"
```

| Setting | Default | Description |
|---------|---------|-------------|
| `vault_path` | — | Absolute path to Obsidian vault root |
| `tasks_folder` | `tasks` | Folder within vault to scan for tasks |
| `file_pattern` | `*.md` | Glob pattern for task files |
| `daily_notes` | `false` | Enable daily note integration |
| `daily_notes_folder` | `Daily` | Folder for daily notes |
| `daily_notes_heading` | `## Tasks` | Heading under which to add tasks |
| `daily_notes_format` | `2006-01-02.md` | Go time format for daily note filenames |

### `jira`

```yaml
- name: jira
  settings:
    url: https://company.atlassian.net
    auth_type: basic
    email: user@example.com
    api_token: your-api-token
    jql: "assignee = currentUser() AND statusCategory != Done"
    max_results: "50"
    poll_interval: 30s
```

| Setting | Default | Description |
|---------|---------|-------------|
| `url` | — | Jira Cloud or Server URL |
| `auth_type` | `basic` | Authentication type |
| `email` | — | Jira account email |
| `api_token` | — | Jira API token |
| `jql` | — | JQL query to filter issues |
| `max_results` | `50` | Maximum issues to fetch per poll |
| `poll_interval` | `30s` | How often to refresh from Jira |

### `github`

```yaml
- name: github
  settings:
    owner: your-username
    repo: your-repo
    token: ghp_your_token
```

| Setting | Default | Description |
|---------|---------|-------------|
| `owner` | — | GitHub repository owner |
| `repo` | — | GitHub repository name |
| `token` | — | GitHub personal access token with `repo` scope |

### `reminders`

```yaml
- name: reminders
  settings:
    lists: Work,Personal
    include_completed: false
```

| Setting | Default | Description |
|---------|---------|-------------|
| `lists` | — | Comma-separated Reminders list names |
| `include_completed` | `false` | Include completed reminders |

### `todoist`

```yaml
- name: todoist
  settings:
    api_token: your-todoist-api-token
    project_ids: "111222333, 444555666"
    filter: "today | overdue"
    poll_interval: 30s
```

| Setting | Default | Description |
|---------|---------|-------------|
| `api_token` | — | Todoist API token (or set `TODOIST_API_TOKEN` env var) |
| `project_ids` | — | Comma-separated project IDs to filter |
| `filter` | — | Todoist filter query |
| `poll_interval` | `30s` | Refresh interval |

!!! warning
    `project_ids` and `filter` are mutually exclusive — use one or the other. The `TODOIST_API_TOKEN` environment variable takes precedence over the config file value.

---

## Full Example

```yaml
schema_version: 2
theme: modern

providers:
  - name: textfile
    settings:
      task_file: ~/.threedoors/tasks.yaml

  - name: applenotes
    settings:
      note_title: ThreeDoors Tasks

  - name: obsidian
    settings:
      vault_path: /Users/me/Documents/MyVault
      tasks_folder: tasks
      file_pattern: "*.md"
      daily_notes: true
      daily_notes_folder: Daily
      daily_notes_heading: "## Tasks"

  - name: jira
    settings:
      url: https://company.atlassian.net
      auth_type: basic
      email: user@example.com
      api_token: your-api-token
      jql: "assignee = currentUser() AND statusCategory != Done"

  - name: github
    settings:
      owner: myuser
      repo: myrepo
      token: ghp_your_token

  - name: todoist
    settings:
      api_token: your-todoist-api-token
```

Tasks are aggregated and deduplicated across all configured providers. If a provider fails, ThreeDoors falls back to the next one automatically.

---

## Modifying Configuration

### Via CLI

```bash
threedoors config show               # View current config
threedoors config get theme           # Get a single value
threedoors config set theme scifi     # Set a value
```

### Via TUI

- `:theme` — Open the theme picker
- `:goals edit` — Edit values and goals

### Manual Editing

Edit `~/.threedoors/config.yaml` directly. Changes take effect on next launch.
