# Cron Configuration — SM Sprint Health & QA Coverage Audit

Setup instructions for the two automated monitoring jobs defined in Story 37.2.

## SM Sprint Health Check (every 4 hours)

Queries open PRs for staleness, checks for blocked stories, and summarizes worker status.

### What it checks

- **Stale PRs:** Open PRs with no activity for >24 hours (configurable via `--stale-hours`)
- **Blocked stories:** Story files in `docs/stories/` with `Status: Blocked`
- **Worker status:** Active worker count via `multiclaude worker list`

### Running manually

```bash
./scripts/sm-sprint-health.sh
./scripts/sm-sprint-health.sh --stale-hours 12     # more aggressive staleness threshold
./scripts/sm-sprint-health.sh --report             # also send to supervisor
```

### Automated via multiclaude loop

The recommended approach uses the `/loop` skill within a supervisor session:

```
/loop 4h /bmad-bmm-sprint-status
```

This runs the SM sprint health check every 4 hours within the supervisor's context.

### Automated via system cron

If a standalone cron job is preferred:

```bash
crontab -e
```

Add the following entry (runs every 4 hours):

```cron
0 */4 * * * cd /path/to/ThreeDoors && ./scripts/sm-sprint-health.sh --report >> /tmp/sm-sprint-health.log 2>&1
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REPO` | `arcaven/ThreeDoors` | GitHub repo for PR queries |
| `STALE_HOURS` | `24` | Hours without activity before a PR is stale |
| `STORIES_DIR` | `docs/stories` | Path to story files |
| `REPORT_TO_SUPERVISOR` | `false` | Send report via `multiclaude message send` |

### Dependencies

- `gh` (GitHub CLI) — authenticated with repo access
- `jq` — JSON processing

## QA Coverage Audit (weekly, Monday 9am)

Runs `go test -cover`, compares per-package coverage against the stored baseline, and flags regressions.

### What it checks

- Per-package test coverage percentages
- Coverage drops exceeding 5 percentage points (configurable via `--threshold`)
- New packages without baseline entries
- Removed packages that were in the previous baseline

### Running manually

```bash
./scripts/qa-coverage-audit.sh                     # report only, no baseline update
./scripts/qa-coverage-audit.sh --update            # report and update baseline
./scripts/qa-coverage-audit.sh --threshold 3       # stricter regression threshold
./scripts/qa-coverage-audit.sh --update --report   # full audit with supervisor notification
```

### Automated via system cron

```bash
crontab -e
```

Add the following entry (runs Monday at 9am local time):

```cron
0 9 * * 1 cd /path/to/ThreeDoors && ./scripts/qa-coverage-audit.sh --update --report >> /tmp/qa-coverage-audit.log 2>&1
```

### Automated via multiclaude

```bash
multiclaude work "Run QA coverage audit: ./scripts/qa-coverage-audit.sh --update --report" --repo ThreeDoors
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASELINE_FILE` | `docs/quality/coverage-baseline.json` | Path to coverage baseline |
| `THRESHOLD` | `5` | Percentage points drop that triggers a regression flag |
| `UPDATE_BASELINE` | `false` | Update baseline after reporting |
| `REPORT_TO_SUPERVISOR` | `false` | Send report via `multiclaude message send` |

### Dependencies

- `go` — Go toolchain for running tests
- `jq` — JSON processing

## Coverage Baseline

The coverage baseline is stored at `docs/quality/coverage-baseline.json`. Format:

```json
{
  "updated": "2026-03-08",
  "packages": {
    "github.com/arcaven/ThreeDoors/internal/tasks": {
      "coverage": 78.5,
      "updated": "2026-03-08"
    }
  }
}
```

- The first run of `qa-coverage-audit.sh` establishes the initial baseline (no regressions flagged)
- Subsequent runs compare against the stored baseline
- Use `--update` to persist the new coverage numbers after each audit
- The baseline is checked into version control so the team shares a common reference
