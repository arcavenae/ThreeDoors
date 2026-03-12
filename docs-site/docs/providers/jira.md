# Jira

Pull tasks from Jira Cloud or Server via REST API with JQL filtering.

## Overview

The Jira provider imports issues matching a JQL query into ThreeDoors. This lets you work through your Jira backlog using the three doors approach, with configurable filtering and automatic polling for updates.

## Prerequisites

- A Jira Cloud or Server instance
- An API token (Cloud) or personal access token (Server)
- Network access to your Jira instance

## Setup

1. **Generate an API token:**
    - **Jira Cloud:** Go to [Atlassian Account Settings → API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens) and create a new token
    - **Jira Server:** Create a personal access token in your profile settings

2. **Add the Jira provider** to your configuration:

    ```yaml
    # ~/.threedoors/config.yaml
    providers:
      - name: jira
        settings:
          url: https://company.atlassian.net
          auth_type: basic
          email: your-email@company.com
          api_token: your-api-token
          jql: "assignee = currentUser() AND statusCategory != Done"
          max_results: "50"
          poll_interval: 30s
    ```

3. Launch ThreeDoors — matching Jira issues appear as doors

## Configuration

```yaml
providers:
  - name: jira
    settings:
      url: https://company.atlassian.net
      auth_type: basic
      email: your-email@company.com
      api_token: your-api-token
      jql: "assignee = currentUser() AND statusCategory != Done"
      max_results: "50"
      poll_interval: 30s
```

### Settings

| Setting | Required | Default | Description |
|---------|----------|---------|-------------|
| `url` | Yes | — | Base URL of your Jira instance |
| `auth_type` | No | `basic` | Authentication type (`basic`) |
| `email` | Yes | — | Your Jira account email |
| `api_token` | Yes | — | API token or personal access token |
| `jql` | No | `"assignee = currentUser() AND statusCategory != Done"` | JQL query to filter issues |
| `max_results` | No | `"50"` | Maximum number of issues to fetch per poll |
| `poll_interval` | No | `30s` | How often to refresh issues from Jira |

## Usage

### JQL Filtering

The `jql` setting accepts any valid JQL query. Common examples:

```yaml
# Your open tasks (default)
jql: "assignee = currentUser() AND statusCategory != Done"

# Current sprint for a specific project
jql: "project = PROJ AND sprint in openSprints()"

# Tagged tasks across all projects
jql: "labels = focus AND statusCategory != Done"

# High-priority items
jql: "assignee = currentUser() AND priority in (Highest, High) AND statusCategory != Done"

# Multiple projects
jql: "(project = PROJ1 OR project = PROJ2) AND assignee = currentUser() AND statusCategory != Done"
```

!!! tip
    Test your JQL query in the Jira web UI first (Issues → Search → Advanced) to make sure it returns the issues you expect.

### Polling

Jira tasks refresh on the configured interval (`poll_interval`, default 30 seconds). The adapter uses a circuit breaker — if Jira is repeatedly unreachable, polling backs off automatically to avoid flooding a degraded service.

### Field Mapping

Jira fields map to ThreeDoors task properties:

| Jira Field | ThreeDoors Property |
|------------|---------------------|
| Summary | Task text |
| Issue key (e.g., PROJ-123) | Task ID prefix |
| Status category | Task status |
| Priority | Effort hint |
| Labels | Task categories |

### Read-Only

The Jira provider is **read-only** — ThreeDoors imports issues but does not write status changes back to Jira. Completing a task in ThreeDoors removes it from your local pool but does not transition the Jira issue.

## Troubleshooting

**Jira tasks not loading**

- Verify `url`, `email`, and `api_token` are correct
- Test your JQL query in the Jira web UI first
- Ensure the API token has read access to the target projects
- Run `:health` to check Jira connectivity

**Authentication errors**

- For Jira Cloud, use your Atlassian account email and an API token (not your password)
- For Jira Server, ensure your personal access token hasn't expired
- Check that `auth_type` is set to `basic`

**Too many / too few results**

- Adjust `max_results` to control how many issues are fetched
- Refine your `jql` query to be more specific
- Remember that `max_results` is a string value in the config (e.g., `"50"`)

**Polling seems stuck**

- Check `:health` for Jira connectivity status
- If Jira is unreachable, the circuit breaker backs off automatically
- Pending operations replay when connectivity is restored
