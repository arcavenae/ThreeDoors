# MCP Server

ThreeDoors ships a separate MCP (Model Context Protocol) server binary that exposes tasks and analytics to LLM agents like Claude.

---

## Running the Server

### stdio Transport (default)

For Claude Desktop, Cursor, and other desktop LLM clients:

```bash
threedoors-mcp
```

### SSE Transport

For web-based clients:

```bash
threedoors-mcp --transport sse --port 8080
```

---

## Available Tools

| Tool | Description |
|------|-------------|
| `query_tasks` | Query tasks with filters (status, type, effort, provider, text, date range) |
| `get_task` | Get full task details with enrichment data |
| `search_tasks` | Full-text search with relevance scoring |
| `list_providers` | List configured providers with health/sync status |
| `get_session` | Current or historical session metrics |
| `get_mood_correlation` | Mood vs. productivity correlation analysis |
| `get_productivity_profile` | Time-of-day productivity analysis |
| `burnout_risk` | Burnout risk assessment (0-1 score) |
| `walk_graph` | Traverse task relationship graph (BFS) |
| `find_paths` | Find paths between two tasks in the graph |
| `get_critical_path` | Longest dependency chain |
| `get_orphans` | Find tasks with no relationships |
| `get_completions` | Completion data with grouping options |
| `get_clusters` | Discover related task groups |
| `get_provider_overlap` | Find tasks shared between providers |

---

## Client Configuration

### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "threedoors": {
      "command": "threedoors-mcp",
      "args": []
    }
  }
}
```

### Cursor

Add to your Cursor MCP settings:

```json
{
  "mcpServers": {
    "threedoors": {
      "command": "threedoors-mcp"
    }
  }
}
```

### SSE Clients

For web-based or remote clients, start the server with SSE transport and connect to the endpoint:

```bash
threedoors-mcp --transport sse --port 8080
```

The SSE endpoint will be available at `http://localhost:8080`.

---

## Example Prompts

Once the MCP server is connected, you can ask your LLM agent questions like:

- "What tasks do I have in progress?"
- "Show me my productivity patterns for this week"
- "Which tasks am I avoiding?"
- "What's my burnout risk score?"
- "Find all tasks related to the auth module"
- "What's the critical path through my task dependencies?"
- "How does my mood affect my task completion rate?"

---

## Security

The MCP server includes security middleware that:

- Validates all tool inputs against schemas
- Enforces size limits on request payloads
- Sanitizes output to prevent injection
- Runs with the same file permissions as the ThreeDoors data directory

The server only accesses your local ThreeDoors data — it does not make network requests or expose data externally (unless you explicitly use SSE transport on a network interface).
