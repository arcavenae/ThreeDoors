# Todoist

Sync tasks from Todoist via the REST API.

## Overview

The Todoist provider imports tasks from your Todoist account into ThreeDoors. Filter by project or use Todoist's filter queries to control which tasks appear as doors.

## Prerequisites

- A Todoist account (free or premium)
- A Todoist API token

## Setup

1. **Get your API token:**
    - Go to [Todoist Settings → Integrations → Developer](https://todoist.com/app/settings/integrations/developer)
    - Copy your API token

2. **Add the Todoist provider** to your configuration:

    ```yaml
    # ~/.threedoors/config.yaml
    providers:
      - name: todoist
        settings:
          api_token: your-todoist-api-token
    ```

3. Launch ThreeDoors — your Todoist tasks appear as doors

!!! tip
    You can also set the `TODOIST_API_TOKEN` environment variable instead of putting the token in your config file. The environment variable takes precedence over the config value.

## Configuration

```yaml
providers:
  - name: todoist
    settings:
      api_token: your-todoist-api-token
      project_ids: "111222333, 444555666"
      filter: "today | overdue"
      poll_interval: 30s
```

### Settings

| Setting | Required | Default | Description |
|---------|----------|---------|-------------|
| `api_token` | Yes* | — | Todoist API token (*or set `TODOIST_API_TOKEN` env var) |
| `project_ids` | No | — | Comma-separated project IDs to filter |
| `filter` | No | — | Todoist filter query |
| `poll_interval` | No | `30s` | How often to refresh tasks |

!!! warning "Mutually Exclusive Filters"
    `project_ids` and `filter` cannot be used together — choose one or the other.

## Usage

### Filtering by Project

To import tasks from specific projects only, use `project_ids`:

```yaml
project_ids: "111222333, 444555666"
```

Find project IDs in the Todoist web app — they appear in the URL when viewing a project.

### Filtering by Query

Use Todoist's filter syntax to control which tasks appear:

```yaml
# Tasks due today or overdue
filter: "today | overdue"

# High-priority tasks
filter: "p1 | p2"

# Tasks with a specific label
filter: "@focus"

# Tasks due this week
filter: "7 days"
```

See [Todoist's filter documentation](https://todoist.com/help/articles/introduction-to-filters-V98wIH) for the full filter syntax.

### Project Mapping

Todoist projects map to task context in ThreeDoors. When viewing a task from Todoist, you'll see the project name as additional context.

### Polling

Tasks refresh every 30 seconds by default. Customize with `poll_interval`:

```yaml
# Refresh every 2 minutes
poll_interval: 2m

# Refresh every 10 seconds
poll_interval: 10s
```

### Read-Only

The Todoist provider is **read-only** — ThreeDoors imports tasks but does not complete or update them in Todoist. Completing a task in ThreeDoors removes it from your local pool but does not mark it done in Todoist.

## Troubleshooting

**Todoist tasks not loading**

- Verify `api_token` is set correctly (in config or via `TODOIST_API_TOKEN` env var)
- Check that `project_ids` and `filter` are not both set (they're mutually exclusive)
- Run `:health` to check Todoist connectivity

**Wrong tasks appearing**

- Review your `filter` or `project_ids` settings
- Test your filter in the Todoist web app first
- Without any filter, all active tasks from all projects are imported

**Token not working**

- Regenerate your API token at Todoist Settings → Integrations → Developer
- If using the environment variable, verify it's set: `echo $TODOIST_API_TOKEN`
- The environment variable takes precedence over the config file value
