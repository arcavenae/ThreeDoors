# Sessions & Analytics

## Session Tracking

Every session automatically records:

| Metric | Description |
|--------|-------------|
| Session ID | Unique identifier per session |
| Start/end time | UTC timestamps |
| Duration | Session length in seconds |
| Tasks completed | Count of completed tasks |
| Doors viewed | How many door sets you saw |
| Refreshes used | Number of times you pressed ++s++ / ++down++ |
| Detail views | How many tasks you opened |
| Door selections | Which position (left/center/right) and which task |
| Task bypasses | Which tasks were shown but not selected |
| Mood entries | Timestamped mood logs |
| Door feedback | Feedback given via ++n++ key |
| Time to first door | Seconds from launch to first door selection |

## Viewing Stats

**In-app:** Type `:stats` in the command palette for a quick session summary.

**Insights dashboard:** Type `:dashboard` or `:insights` for detailed analytics.

!!! info
    Insights require 3+ completed sessions for meaningful data. Each app launch and quit counts as one session.

## Insights Dashboard

Access via `:dashboard` or `:insights`. Shows:

- **Completion trends** — last 7 days with sparkline visualization
- **Streaks** — current and longest consecutive completion days
- **Mood and productivity** — average completions per mood state
- **Door position preferences** — left/center/right selection percentages

### Filtered Views

| Command | Shows |
|---------|-------|
| `:insights mood` | Mood & productivity correlations |
| `:insights avoidance` | Avoidance pattern analysis |

## Completion Tracking

After completing a task, ThreeDoors shows:

- How many tasks you've completed today
- Comparison with yesterday's count
- Your current streak (consecutive days with at least one completion)

## Mood Correlation

Over time, ThreeDoors correlates mood logs with task completion patterns:

- Which moods lead to the most completions
- Which task types you prefer in different moods
- How mood affects your door position preference

This data powers [mood-aware door selection](doors-interaction.md#mood-aware-selection).

## Data Files

Session data is stored in `~/.threedoors/`:

| File | Purpose |
|------|---------|
| `sessions.jsonl` | One JSON object per line, one per session |
| `patterns.json` | Cached analysis report (regenerated when new sessions arrive) |
| `completed.txt` | Append-only completion log |

## Analysis Scripts

Run `just analyze` to execute the analysis scripts in `scripts/` against your session data.
