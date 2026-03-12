# GitHub Issues

Import issues from a GitHub repository into ThreeDoors.

## Overview

The GitHub Issues provider pulls issues assigned to you from a GitHub repository. Issue labels map to ThreeDoors task categories, making it easy to work through your GitHub backlog with the three doors approach.

## Prerequisites

- A GitHub account
- A personal access token with `repo` scope
- At least one repository with open issues

## Setup

1. **Create a personal access token:**
    - Go to [GitHub Settings → Developer Settings → Personal access tokens](https://github.com/settings/tokens)
    - Generate a new token with the `repo` scope
    - Copy the token (it won't be shown again)

2. **Add the GitHub provider** to your configuration:

    ```yaml
    # ~/.threedoors/config.yaml
    providers:
      - name: github
        settings:
          owner: your-username
          repo: your-repo
          token: ghp_your_token
    ```

3. Launch ThreeDoors — issues assigned to you appear as doors

## Configuration

```yaml
providers:
  - name: github
    settings:
      owner: your-username
      repo: your-repo
      token: ghp_your_token
```

### Settings

| Setting | Required | Description |
|---------|----------|-------------|
| `owner` | Yes | Repository owner (username or organization) |
| `repo` | Yes | Repository name |
| `token` | Yes | Personal access token with `repo` scope |

!!! tip
    You can also use a fine-grained personal access token with read-only access to issues for better security.

## Usage

### How It Works

- Issues assigned to you in the specified repository are imported as tasks
- Issue titles become task text
- Issue labels map to task categories
- Issue state (open/closed) maps to task status

### Label Mapping

GitHub issue labels are imported as task metadata:

| GitHub Label | ThreeDoors Mapping |
|--------------|-------------------|
| `bug` | Technical type |
| `enhancement` | Creative type |
| `documentation` | Administrative type |
| Other labels | Stored as task context |

### Repo Filtering

Each GitHub provider entry targets a single repository. To import from multiple repos, configure multiple provider entries:

```yaml
providers:
  - name: github
    settings:
      owner: your-username
      repo: project-one
      token: ghp_your_token
  - name: github
    settings:
      owner: your-org
      repo: project-two
      token: ghp_your_token
```

### Read-Only

The GitHub Issues provider is **read-only** — ThreeDoors imports issues but does not close or update them on GitHub. Completing a task in ThreeDoors removes it from your local pool but does not close the GitHub issue.

## Troubleshooting

**Issues not appearing**

- Verify the `owner` and `repo` values are correct
- Ensure your token has `repo` scope
- Check that you have issues assigned to you in the repository
- Run `:health` to check GitHub connectivity

**Authentication errors**

- Tokens expire — verify yours is still valid at GitHub Settings → Personal access tokens
- Classic tokens need `repo` scope; fine-grained tokens need Issues read access
- Make sure you're using the token value, not the token name

**Rate limiting**

- GitHub has API rate limits (5,000 requests/hour for authenticated requests)
- If you're hitting limits, reduce the number of configured GitHub provider entries or increase poll interval
