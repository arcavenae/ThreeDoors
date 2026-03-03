package llm

import (
	"strings"
	"testing"
)

func TestStorySpec_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    StorySpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid spec",
			spec: StorySpec{
				StoryID:   "14.1",
				Title:     "Test Story",
				UserStory: "As a dev, I want tests, So that I ship quality",
				ACs:       []string{"AC1"},
				Tasks:     []string{"Task1"},
			},
			wantErr: false,
		},
		{
			name: "missing story_id",
			spec: StorySpec{
				Title:     "Test",
				UserStory: "Story",
				ACs:       []string{"AC1"},
				Tasks:     []string{"Task1"},
			},
			wantErr: true,
			errMsg:  "story_id is required",
		},
		{
			name: "missing title",
			spec: StorySpec{
				StoryID:   "1.1",
				UserStory: "Story",
				ACs:       []string{"AC1"},
				Tasks:     []string{"Task1"},
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing user_story",
			spec: StorySpec{
				StoryID: "1.1",
				Title:   "Test",
				ACs:     []string{"AC1"},
				Tasks:   []string{"Task1"},
			},
			wantErr: true,
			errMsg:  "user_story is required",
		},
		{
			name: "empty acceptance criteria",
			spec: StorySpec{
				StoryID:   "1.1",
				Title:     "Test",
				UserStory: "Story",
				ACs:       []string{},
				Tasks:     []string{"Task1"},
			},
			wantErr: true,
			errMsg:  "acceptance criterion",
		},
		{
			name: "empty tasks",
			spec: StorySpec{
				StoryID:   "1.1",
				Title:     "Test",
				UserStory: "Story",
				ACs:       []string{"AC1"},
				Tasks:     []string{},
			},
			wantErr: true,
			errMsg:  "task is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestStorySpec_ToMarkdown(t *testing.T) {
	t.Parallel()

	spec := StorySpec{
		StoryID:   "14.1",
		Title:     "LLM Decomposition",
		UserStory: "As a dev, I want LLM decomposition, So that tasks break down automatically",
		ACs:       []string{"Given a task, When decomposed, Then stories are generated"},
		Tasks:     []string{"Implement backend", "Write tests"},
		DevNotes:  "Use HTTP mocks for testing",
	}

	md := spec.ToMarkdown()

	checks := []string{
		"# Story 14.1: LLM Decomposition",
		"## Story",
		"As a dev, I want LLM decomposition",
		"## Status",
		"Ready for Dev",
		"## Acceptance Criteria",
		"- Given a task",
		"## Tasks",
		"1. Implement backend",
		"2. Write tests",
		"## Dev Notes",
		"Use HTTP mocks",
	}

	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("ToMarkdown() missing %q", check)
		}
	}
}

func TestParseStorySpecs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "json array",
			input:     `[{"story_id":"1.1","title":"Test","user_story":"US","acceptance_criteria":["AC"],"tasks":["T"]}]`,
			wantCount: 1,
		},
		{
			name:      "single object",
			input:     `{"story_id":"1.1","title":"Test","user_story":"US","acceptance_criteria":["AC"],"tasks":["T"]}`,
			wantCount: 1,
		},
		{
			name:      "decomposition result wrapper",
			input:     `{"source_task":"test","stories":[{"story_id":"1.1","title":"Test","user_story":"US","acceptance_criteria":["AC"],"tasks":["T"]}]}`,
			wantCount: 1,
		},
		{
			name:    "invalid json",
			input:   "this is not json",
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   "[]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			specs, err := ParseStorySpecs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStorySpecs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(specs) != tt.wantCount {
				t.Errorf("ParseStorySpecs() got %d specs, want %d", len(specs), tt.wantCount)
			}
		})
	}
}
