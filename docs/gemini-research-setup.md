# Gemini Research Setup Guide

This guide covers setting up the Gemini CLI for research queries in ThreeDoors.

## Prerequisites

- **Node.js/npm** — `brew install node` (if not already installed)
- **jq** — `brew install jq` (JSON parsing)
- **Google Account** — for OAuth authentication

## Installation

Install the Gemini CLI globally:

```bash
npm install -g @google/gemini-cli
```

Or use without global installation (the wrapper script uses this approach):

```bash
npx @google/gemini-cli
```

## OAuth Setup (First Time)

Run the Gemini CLI interactively to complete the OAuth flow:

```bash
gemini
```

This opens your browser for Google Account sign-in. After authenticating:
- The OAuth token is cached locally
- Tokens auto-refresh — no manual renewal needed
- No API key or environment variables required

## Verify Installation

Run the verification command:

```bash
gemini -p "Hello" --output-format json
```

Expected output: valid JSON with a `.response` field containing the model's reply.

## Using the Research Wrapper Script

The wrapper script at `scripts/gemini-research.sh` provides structured research capabilities.

### Basic Usage

```bash
# Quick lookup (uses Gemini Flash — fast, 1000 req/day)
./scripts/gemini-research.sh --depth quick --query "What is Bubbletea in Go?"

# Standard research (uses Gemini Pro — balanced, 50 req/day)
./scripts/gemini-research.sh --depth standard --query "Compare Go TUI frameworks"

# Deep analysis (uses Gemini Pro — thorough, 50 req/day)
./scripts/gemini-research.sh --depth deep --query "Architecture patterns for terminal task managers"
```

### Depth Levels

| Depth | Model | Daily Limit | Use Case |
|---|---|---|---|
| `quick` | gemini-2.5-flash | 1,000/day | Fact-checking, summaries, quick answers |
| `standard` | gemini-2.5-pro | 50/day | Balanced research, technology comparisons |
| `deep` | gemini-2.5-pro | 50/day | Architecture analysis, deep technology evaluation |

### Output

Each research query creates a timestamped directory in `_bmad-output/research-reports/`:

```
_bmad-output/research-reports/20260311-143022-what-is-bubbletea/
├── response.json   # Raw Gemini CLI JSON output
├── report.md       # Extracted response text
└── stderr.log      # CLI stderr (for debugging)
```

The report content is also printed to stdout.

### Budget Tracking

Usage is tracked in `_bmad-output/research-reports/budget.json`. The script:
- Warns at 80% of daily limit
- Blocks queries when the daily limit is reached
- Records success/failure for each query

## Troubleshooting

### Token Refresh Failure

If you see authentication errors after a period of inactivity:

```bash
# Re-authenticate by running gemini interactively
gemini
```

### Model Unavailable

If Gemini Pro is unavailable on your account, use `--depth quick` (Flash model) as a fallback.

### Missing Dependencies

```bash
# Install Node.js/npm
brew install node

# Install jq
brew install jq
```

### Version Pinning

The wrapper script pins Gemini CLI to version 0.32.1 via npx to prevent breaking changes. To update, modify the `GEMINI_CLI_VERSION` variable in `scripts/gemini-research.sh`.
