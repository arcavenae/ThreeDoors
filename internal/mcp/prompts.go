package mcp

// promptTemplate holds a prompt template with its metadata.
type promptTemplate struct {
	Description string
	Template    string
}

// promptTemplates maps prompt names to their templates.
var promptTemplates = map[string]promptTemplate{
	"daily_summary": {
		Description: "Generate a daily productivity summary by querying session and analytics data.",
		Template: `You are a productivity coach for a ThreeDoors user. Compose a daily summary by calling these tools in order:

1. Call get_session(type: "current") to get today's session metrics
2. Call get_mood_correlation(from: "<today_start>", to: "<now>") to see mood-productivity patterns
3. Call get_productivity_profile(from: "<today_start>", to: "<now>") for time-of-day analysis
4. Call burnout_risk() to check current burnout indicators

Then synthesize a brief, encouraging daily summary that includes:
- Tasks completed today vs recent average
- Current mood trend and its correlation with productivity
- Peak productivity time today
- Any burnout risk signals to watch
- One actionable suggestion for tomorrow`,
	},
	"weekly_retrospective": {
		Description: "Generate a weekly retrospective by querying multi-day analytics.",
		Template: `You are a productivity coach for a ThreeDoors user. Compose a weekly retrospective by calling these tools in order:

1. Call get_completions(from: "<week_start>", to: "<week_end>", group_by: "day", include_mood: true, include_patterns: true) for daily breakdown
2. Call get_mood_correlation(from: "<week_start>", to: "<week_end>") for mood-productivity patterns
3. Call get_productivity_profile(from: "<week_start>", to: "<week_end>") for time-of-day trends
4. Call burnout_risk() for overall health check

Then compose a weekly retrospective that includes:
- Total velocity (tasks completed) and comparison to prior week
- Best and worst days with context on why
- Mood-productivity correlation insights
- Streak status and consistency
- Burnout risk assessment
- Top 3 recommendations for next week`,
	},
	"blocked_tasks": {
		Description: "Query blocked tasks, show blockers, and suggest unblocking actions.",
		Template: `You are a productivity coach for a ThreeDoors user. Identify and resolve blocked tasks by calling these tools in order:

1. Call query_tasks(status: "blocked") to get all blocked tasks
2. For each blocked task, call walk_graph(task_id: "<task_id>", depth: 2, direction: "incoming", edge_types: ["blocks"]) to find what blocks it
3. Call prioritize_tasks() to see which blockers are highest priority to resolve

Then report:
- List of blocked tasks with their blocker chains
- Which blockers, if resolved, would unblock the most tasks
- Specific suggested actions to unblock each task
- Quick wins: blockers that are quick-win effort level`,
	},
	"task_deep_dive": {
		Description: "Deep dive into a specific task: full context, relationships, history, and suggestions.",
		Template: `You are a productivity coach for a ThreeDoors user. Perform a deep dive on a specific task by calling these tools in order:

1. Call get_task(task_id: "<task_id>") for full task detail and enrichment data
2. Call walk_graph(task_id: "<task_id>", depth: 2) to see the full relationship graph
3. Call get_session(type: "history") to check if this task appears in session history
4. Call what_if(complete_task_ids: ["<task_id>"]) to model the impact of completing it

Then provide a comprehensive report:
- Full task context, notes, and metadata
- Relationship map: what it blocks, what blocks it, related tasks
- Historical engagement: when was it last worked on, how many sessions
- Impact analysis: what completing it would unblock
- Actionable suggestions: next steps, estimated effort, best time to tackle it`,
	},
}

// promptDefinitions returns the list of available MCP prompts.
func promptDefinitions() []PromptItem {
	return []PromptItem{
		{
			Name:        "daily_summary",
			Description: "Generate a daily productivity summary by querying session and analytics data.",
		},
		{
			Name:        "weekly_retrospective",
			Description: "Generate a weekly retrospective by querying multi-day analytics.",
		},
		{
			Name:        "blocked_tasks",
			Description: "Query blocked tasks, show blockers, and suggest unblocking actions.",
		},
		{
			Name:        "task_deep_dive",
			Description: "Deep dive into a specific task: full context, relationships, history, and suggestions.",
		},
	}
}
