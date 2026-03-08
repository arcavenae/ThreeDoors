# ADR-0003: YAML for Task Persistence

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #6, #8
- **Related ADRs:** ADR-0008 (Atomic File Writes), ADR-0016 (Cancel SQLite Enrichment)

## Context

ThreeDoors needs a persistence format for storing tasks with metadata including status, notes, timestamps, effort levels, and categories. The storage must be:
- Human-readable and editable in any text editor
- Simple to implement without external database dependencies
- Capable of storing structured metadata per task
- Compatible with version control

## Considered Options

1. **Plain text** — Simple but no metadata support
2. **YAML** — Human-readable structured format
3. **JSON** — Machine-friendly structured format
4. **SQLite** — Embedded relational database
5. **TOML** — Configuration-oriented structured format

## Decision

Use **YAML** (`gopkg.in/yaml.v3`) for task storage in `~/.threedoors/tasks.yaml`.

Use **JSONL** (JSON Lines) for session metrics in `~/.threedoors/metrics.jsonl` (see ADR-0024).

## Rationale

- Human-readable — users can inspect and manually edit their task files
- Structured metadata — supports nested fields (status, notes, timestamps, categories)
- Familiar format — widely understood in the developer tools ecosystem
- No runtime dependencies — pure Go YAML parser, no CGO
- Git-friendly — reasonable diffs when tasks change

## Consequences

### Positive
- Users can manually edit tasks in any text editor
- No database setup or migration tooling required
- Easy to debug by reading the raw file
- Natural fit for Go struct marshaling/unmarshaling

### Negative
- No query capabilities — all tasks loaded into memory
- File locking is application-responsibility (solved via atomic writes)
- Performance scales linearly with task count (acceptable for personal task management)
- Schema evolution requires careful backward compatibility (see ADR-0010, ADR-0014)
