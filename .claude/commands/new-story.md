# /new-story — Generate Story Template

Create a new story file from the standard template. See SOUL.md for project philosophy and CLAUDE.md for coding standards that apply to all stories.

## Instructions

1. Ask for the story identifier (e.g., "27.3") and title.

2. Read `docs/stories/` to find the latest story format (use the most recent completed story file as reference for structure).

3. Create `docs/stories/{id}.story.md` with this structure:

```markdown
# Story {id}: {title}

**Status:** Draft
**Epic:** {epic number} — {epic title}
**Depends On:** {dependencies or "None"}
**Blocks:** {blocked stories or "None"}

## User Story

As a {role},
I want {capability},
So that {benefit}.

## Acceptance Criteria

**Given** {context}
**When** {action}
**Then:**

- AC1: {criterion}
- AC2: {criterion}

## NOT In Scope

- {exclusion}

## Quality Gate

- All standard checks pass (see CLAUDE.md)
- {story-specific quality items}

## Architecture & Design

{Data models, view modes, key design decisions}

## Key Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| | | |

## Testing

{Story-specific test requirements — table-driven tests, edge cases to cover}

## Dependencies

- {dependency or "None"}

## Tasks

1. {task}
2. {task}
```

4. **IMPORTANT:** Do NOT include a Pre-PR Submission Checklist — that lives in CLAUDE.md and applies to all stories automatically.

5. Do NOT include Dev Notes that duplicate CLAUDE.md rules (coding standards, error handling patterns, atomic writes, etc.). Only include story-SPECIFIC implementation notes that aren't covered by CLAUDE.md.

6. After creating the file, confirm the path and remind the developer to fill in the placeholder sections before implementation begins.
