# DFCP Schema Reference (v1.0)

The `.dfcp.yaml` file declares a repository's Dark Factory Control Pattern profile. It is the single source of truth for governance rules, autonomy levels, and CI requirements. Tooling (CI workflows, the multiclaude daemon, and AI agents) reads this file to determine what is permitted.

## Profiles

| Profile | Description | Human Review | Max Autonomy |
|---------|-------------|-------------|--------------|
| `golden` | Human-gated governance files, AI handles implementation | Required for governance files | L3 |
| `factory` | Full AI autonomy, CI-only gates | None | L4 |
| `custom` | Mix of golden and factory rules, configured per-field | Per configuration | Per configuration |

**Golden repos** are the source of truth — they contain the project's philosophy, architecture, and governance. AI agents implement features but humans review all governance changes.

**Factory repos** are fully AI-managed variants spawned from a golden repo. They operate within a variant budget and require no human review. CI gates are the only quality control.

**Custom repos** allow fine-grained control when neither golden nor factory fits. Each field is configured independently.

## Top-Level Structure

```yaml
dfcp:
  version: "1.0"
  profile: golden
  autonomy: { ... }
  governance: { ... }
  ci: { ... }
  factory: { ... }
```

All fields are under the `dfcp` top-level key.

## Field Reference

### `dfcp.version`

- **Type:** string
- **Required:** yes
- **Values:** `"1.0"`
- **Purpose:** Schema version for forward compatibility. Tooling should reject unknown versions rather than guessing.

### `dfcp.profile`

- **Type:** string
- **Required:** yes
- **Values:** `golden` | `factory` | `custom`
- **Purpose:** Declares the overall governance posture. Tooling uses this as a quick classifier before reading individual fields.

### `dfcp.autonomy`

Controls what autonomy levels are permitted for AI work in this repo.

#### `dfcp.autonomy.default`

- **Type:** string
- **Required:** yes
- **Values:** `L0` | `L1` | `L2` | `L3` | `L4`
- **Purpose:** The autonomy level applied to AI work when not explicitly specified. Workers tag their commits and PRs with this level unless overridden.

#### `dfcp.autonomy.max`

- **Type:** string
- **Required:** yes
- **Values:** `L0` | `L1` | `L2` | `L3` | `L4`
- **Purpose:** The maximum autonomy level permitted. Tooling should reject work tagged above this level. Golden repos cap at L3 (human PR review required). Factory repos allow L4.

### `dfcp.governance`

Rules governing how changes flow through the repo.

#### `dfcp.governance.require_human_review`

- **Type:** list of glob patterns
- **Required:** yes (may be empty `[]` for factory repos)
- **Purpose:** Files matching these patterns require human review before merge. These patterns should match the paths declared in `.github/CODEOWNERS`. The validation script checks this consistency.

#### `dfcp.governance.require_story_reference`

- **Type:** boolean
- **Required:** yes
- **Purpose:** When `true`, all commits must reference a story (e.g., `Story X.Y`). The CI scope-check workflow enforces this.

#### `dfcp.governance.require_provenance`

- **Type:** boolean
- **Required:** yes
- **Purpose:** When `true`, AI-generated work must carry provenance tags (autonomy level in commits, PRs, and story files). See `docs/operations/provenance.md`.

#### `dfcp.governance.require_signed_commits`

- **Type:** boolean
- **Required:** yes
- **Purpose:** When `true`, all commits must be GPG-signed. Enforced by git hooks and branch protection rules.

### `dfcp.ci`

CI/CD configuration requirements.

#### `dfcp.ci.scope_check`

- **Type:** string
- **Required:** yes
- **Values:** `warn` | `block` | `off`
- **Purpose:** How the CI scope-check workflow handles out-of-scope changes. `warn` adds a comment but doesn't block merge. `block` fails the check. `off` disables scope checking entirely (typical for factory repos).

#### `dfcp.ci.required_checks`

- **Type:** list of strings
- **Required:** yes
- **Purpose:** GitHub status check names that must pass before merge. These must match the actual `name:` fields in `.github/workflows/` files. The validation script verifies this.

### `dfcp.factory`

Dark factory configuration. Relevant even in golden repos to declare factory eligibility.

#### `dfcp.factory.enabled`

- **Type:** boolean
- **Required:** yes
- **Purpose:** Whether this repo can be used as a dark factory source (spawning variants). Golden repos typically set this to `false`.

#### `dfcp.factory.variant_budget`

- **Type:** integer or `null`
- **Required:** yes
- **Purpose:** Maximum number of variants this factory can produce per cycle. `null` means no budget (not a factory). Only meaningful when `factory.enabled` is `true`.

#### `dfcp.factory.repo_visibility`

- **Type:** string
- **Required:** yes
- **Values:** `public` | `private` | `internal`
- **Purpose:** Required visibility for this repo and any variants it spawns. Factory repos should always be `private`.

## How Tooling Should Read .dfcp.yaml

### General Pattern

1. Parse the YAML file from the repo root
2. Check `dfcp.version` — reject if unknown
3. Read `dfcp.profile` for quick classification
4. Read specific fields as needed for the operation

### Agent Integration

#### merge-queue

- Read `dfcp.autonomy.max` — reject PRs with provenance tags above this level
- Read `dfcp.governance.require_human_review` — skip auto-merge for PRs touching these paths (label as `status.needs-human`)
- Read `dfcp.ci.required_checks` — verify all listed checks passed before merging

#### Workers

- Read `dfcp.governance.require_human_review` — avoid modifying these files unless the story explicitly requires it
- Read `dfcp.autonomy.default` — use this level in provenance tags
- Read `dfcp.governance.require_story_reference` — ensure commits include story references
- Read `dfcp.governance.require_provenance` — include provenance trailers in commits

#### arch-watchdog

- Read `dfcp.profile` — validate that changes are consistent with the repo's governance posture
- Read `dfcp.governance.require_human_review` — flag PRs that touch protected paths without justification
- Read `dfcp.ci.required_checks` — alert if CI workflows are modified in ways that remove required checks

#### pr-shepherd

- Read `dfcp.governance.require_human_review` — when rebasing PRs, preserve the `status.needs-human` label if protected files are touched
- Read `dfcp.ci.required_checks` — verify checks are passing after rebase

### Creating a Factory Profile

To create a `.dfcp.yaml` for a dark factory repo:

1. Start from the `.dfcp.yaml.factory-example` in this repo
2. Set `profile: factory`
3. Set `autonomy.default` and `autonomy.max` to `L4`
4. Set `governance.require_human_review` to `[]`
5. Set `governance.require_story_reference` to `false`
6. Keep `governance.require_provenance` and `governance.require_signed_commits` as `true` — traceability is always required
7. Set `ci.scope_check` to `off`
8. Set `factory.enabled` to `true` and configure `variant_budget`
9. Validate with `scripts/validate-dfcp.sh`
