# Environment Variables

ThreeDoors reads the following environment variables. Environment variables take precedence over `config.yaml` values where applicable.

---

## Provider Credentials

| Variable | Description |
|----------|-------------|
| `TODOIST_API_TOKEN` | Todoist API token. Takes precedence over `api_token` in the Todoist provider config. |

!!! tip
    Store API tokens in environment variables rather than config files to avoid accidentally committing credentials. ThreeDoors checks environment variables first, falling back to config file values.

---

## Standard Go Variables

| Variable | Description |
|----------|-------------|
| `HOME` | User home directory. ThreeDoors stores data in `$HOME/.threedoors/`. |
| `XDG_DATA_HOME` | Not currently used — ThreeDoors always uses `~/.threedoors/`. |
| `NO_COLOR` | When set (any value), disables color output in CLI commands. Respected per the [no-color.org](https://no-color.org) convention. |
| `TERM` | Terminal type. ThreeDoors uses adaptive color profiles — it detects your terminal's color capabilities automatically via Lipgloss. |

---

## CI / Testing

| Variable | Description |
|----------|-------------|
| `CI` | When set to `true`, adjusts behavior for CI environments (e.g., disables interactive prompts, adjusts timeouts for flaky filesystem watchers). |
