package services

import "fmt"

// extractionPromptTemplate is the prompt sent to the LLM for task extraction.
const extractionPromptTemplate = `You are a task extraction assistant. Given the following text, identify all actionable tasks, to-dos, and commitments.

For each task, extract:
- text: A clear, actionable description in imperative form (e.g., "Email Sarah" not "I should email Sarah")
- effort: Estimated effort 1-5 (1=tiny quick action, 2=small task, 3=moderate task, 4=significant work, 5=major project)
- tags: Relevant categories as an array of strings

Return ONLY a JSON array. No explanation, no markdown fencing, no additional text.

Example output:
[
  {"text": "Email Sarah about Q2 budget proposal", "effort": 1, "tags": ["communication", "finance"]},
  {"text": "Prep slides for client demo", "effort": 3, "tags": ["presentation", "client"]}
]

If no tasks are found, return an empty array: []

Text to analyze:
---
%s
---`

// retryPromptTemplate is a stricter prompt used when the first response fails to parse.
const retryPromptTemplate = `Your previous response could not be parsed as JSON. Please try again.

Return ONLY a valid JSON array of task objects. Nothing else — no markdown, no explanation.
Each object must have exactly these fields:
- "text" (string): imperative task description
- "effort" (integer 1-5): effort estimate
- "tags" (array of strings): categories

Example: [{"text": "Do the thing", "effort": 2, "tags": ["work"]}]

If no tasks found, return: []

Text to analyze:
---
%s
---`

// buildExtractionPrompt constructs the extraction prompt for the given text.
func buildExtractionPrompt(text string) string {
	return fmt.Sprintf(extractionPromptTemplate, text)
}

// buildRetryPrompt constructs a stricter retry prompt for the given text.
func buildRetryPrompt(text string) string {
	return fmt.Sprintf(retryPromptTemplate, text)
}

// enrichmentPromptTemplate is the prompt sent to the LLM for task enrichment.
const enrichmentPromptTemplate = `You are a task enrichment assistant. Given a brief task description, produce an enriched version with additional context, tags, and an effort estimate.

TASK:
%s

OUTPUT FORMAT:
Return a single JSON object with these fields:
- "enriched_text": A clearer, more actionable version of the task (1-2 sentences max). Keep the original intent.
- "tags": An array of 1-5 lowercase tags relevant to this task (e.g., ["finance", "personal", "urgent"]).
- "effort": An integer 1-5 representing estimated effort (1=trivial, 2=quick, 3=moderate, 4=substantial, 5=major).
- "context": A brief sentence explaining why this task matters or what it involves. This helps the user understand the task better.

RULES:
- Keep enriched_text concise — enhance, don't overwrite the user's intent.
- Tags should be specific and useful for filtering, not generic ("task", "todo").
- Effort should reflect real-world time/complexity.
- Context should add information the user might not have written down.
- Return ONLY valid JSON, no markdown formatting, no explanation outside the JSON.`

// buildEnrichmentPrompt creates the prompt for task enrichment.
func buildEnrichmentPrompt(taskText string) string {
	return fmt.Sprintf(enrichmentPromptTemplate, taskText)
}
