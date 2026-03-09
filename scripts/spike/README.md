# Story 2.2: Apple Notes Integration Spike

## Purpose

Evaluate three Apple Notes integration approaches for the ThreeDoors app:
1. **AppleScript/osascript** — Shell out to `osascript` for Notes automation
2. **Direct SQLite** — Read Apple Notes' CoreData SQLite database directly
3. **MCP Server** — Use a Model Context Protocol server for Apple Notes

## Prerequisites

- macOS Sonoma 14+ (required for all PoCs)
- Apple Notes app installed (comes with macOS)
- A test note titled "ThreeDoors Tasks" (run `./setup_test_note.sh` or create manually)

## Running PoCs

```bash
# Set up test note (first time only)
./scripts/spike/setup_test_note.sh

# Run individual PoCs
./scripts/spike/poc_applescript_read.sh
./scripts/spike/poc_applescript_write.sh
./scripts/spike/poc_sqlite_read.sh
./scripts/spike/poc_mcp.sh

# Run Go PoCs
cd scripts/spike && go run poc_applescript.go
cd scripts/spike && go run poc_sqlite.go
```

## Running Tests

```bash
# Smoke tests (any platform)
bash scripts/spike/tests/smoke_test.sh

# Go unit tests
cd scripts/spike && go test -v -cover ./...

# Validate spike report completeness
bash scripts/spike/tests/validate_report.sh
```

## Output

- Spike report: `../../_bmad-output/planning-artifacts/2.2-apple-notes-integration.md`
- Benchmark data: `scripts/spike/benchmarks/*.json`
- MCP evaluation: `scripts/spike/mcp_evaluation.md`
