package dispatch

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockStoryGenerator implements StoryGenerator for testing.
type mockStoryGenerator struct {
	result *StoryResult
	err    error
	called bool
	taskIn string
}

func (m *mockStoryGenerator) GenerateStories(_ context.Context, taskDescription string) (*StoryResult, error) {
	m.called = true
	m.taskIn = taskDescription
	return m.result, m.err
}

// mockDispatcher implements Dispatcher for testing the engine.
type mockDispatcher struct {
	workerName string
	createErr  error
	taskIn     string
}

func (m *mockDispatcher) CreateWorker(_ context.Context, task string) (string, error) {
	m.taskIn = task
	return m.workerName, m.createErr
}

func (m *mockDispatcher) ListWorkers(_ context.Context) ([]WorkerInfo, error) {
	return nil, nil
}

func (m *mockDispatcher) GetHistory(_ context.Context, _ int) ([]HistoryEntry, error) {
	return nil, nil
}

func (m *mockDispatcher) RemoveWorker(_ context.Context, _ string) error {
	return nil
}

func (m *mockDispatcher) CheckAvailable(_ context.Context) error {
	return nil
}

func TestDispatchEngine_RequireStoryFalse_SkipsGeneration(t *testing.T) {
	t.Parallel()

	sg := &mockStoryGenerator{result: &StoryResult{Branch: "stories/test"}}
	disp := &mockDispatcher{workerName: "happy-fox"}
	engine := NewDispatchEngine(disp, sg, DispatchConfig{RequireStory: false})

	item := QueueItem{ID: "dq-1234", TaskText: "implement auth"}
	result, err := engine.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if sg.called {
		t.Error("story generator should not be called when RequireStory is false")
	}
	if result.WorkerName != "happy-fox" {
		t.Errorf("WorkerName = %q, want %q", result.WorkerName, "happy-fox")
	}
	if result.StoryResult != nil {
		t.Error("StoryResult should be nil when RequireStory is false")
	}
	if result.StoryErr != "" {
		t.Errorf("StoryErr should be empty, got %q", result.StoryErr)
	}
}

func TestDispatchEngine_RequireStoryTrue_CallsGenerator(t *testing.T) {
	t.Parallel()

	storyResult := &StoryResult{
		Branch:    "stories/dq-1234",
		FilePaths: []string{"docs/stories/14.1.story.md", "docs/stories/14.2.story.md"},
	}
	sg := &mockStoryGenerator{result: storyResult}
	disp := &mockDispatcher{workerName: "brave-lion"}
	engine := NewDispatchEngine(disp, sg, DispatchConfig{RequireStory: true})

	item := QueueItem{ID: "dq-1234", TaskText: "implement auth", Context: "uses JWT"}
	result, err := engine.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !sg.called {
		t.Error("story generator should be called when RequireStory is true")
	}
	if result.WorkerName != "brave-lion" {
		t.Errorf("WorkerName = %q, want %q", result.WorkerName, "brave-lion")
	}
	if result.StoryResult == nil {
		t.Fatal("StoryResult should not be nil")
		return
	}
	if result.StoryResult.Branch != "stories/dq-1234" {
		t.Errorf("Branch = %q, want %q", result.StoryResult.Branch, "stories/dq-1234")
	}
	if len(result.StoryResult.FilePaths) != 2 {
		t.Errorf("FilePaths len = %d, want 2", len(result.StoryResult.FilePaths))
	}

	// Verify task description includes story references.
	if !strings.Contains(disp.taskIn, "stories/dq-1234") {
		t.Error("task description should reference the story branch")
	}
	if !strings.Contains(disp.taskIn, "14.1.story.md") {
		t.Error("task description should reference story file paths")
	}
}

func TestDispatchEngine_StoryGenerationFailure_NonFatal(t *testing.T) {
	t.Parallel()

	sg := &mockStoryGenerator{err: errors.New("LLM timeout")}
	disp := &mockDispatcher{workerName: "calm-bear"}
	engine := NewDispatchEngine(disp, sg, DispatchConfig{RequireStory: true})

	item := QueueItem{ID: "dq-5678", TaskText: "fix bug"}
	result, err := engine.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() should succeed even when story generation fails, got: %v", err)
	}

	if result.WorkerName != "calm-bear" {
		t.Errorf("WorkerName = %q, want %q", result.WorkerName, "calm-bear")
	}
	if result.StoryResult != nil {
		t.Error("StoryResult should be nil when generation fails")
	}
	if !strings.Contains(result.StoryErr, "story generation failed") {
		t.Errorf("StoryErr should contain failure message, got %q", result.StoryErr)
	}
	if !strings.Contains(result.StoryErr, "LLM timeout") {
		t.Errorf("StoryErr should contain original error, got %q", result.StoryErr)
	}

	// Worker should still receive the raw task description.
	if !strings.Contains(disp.taskIn, "Task: fix bug") {
		t.Error("worker should receive raw task description on story failure")
	}
}

func TestDispatchEngine_NilGenerator_NonFatal(t *testing.T) {
	t.Parallel()

	disp := &mockDispatcher{workerName: "swift-hawk"}
	engine := NewDispatchEngine(disp, nil, DispatchConfig{RequireStory: true})

	item := QueueItem{ID: "dq-9999", TaskText: "add feature"}
	result, err := engine.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() should succeed with nil generator, got: %v", err)
	}

	if result.WorkerName != "swift-hawk" {
		t.Errorf("WorkerName = %q, want %q", result.WorkerName, "swift-hawk")
	}
	if !strings.Contains(result.StoryErr, "not configured") {
		t.Errorf("StoryErr should mention not configured, got %q", result.StoryErr)
	}
}

func TestDispatchEngine_WorkerCreationFailure_Fatal(t *testing.T) {
	t.Parallel()

	disp := &mockDispatcher{createErr: errors.New("worker limit reached")}
	engine := NewDispatchEngine(disp, nil, DispatchConfig{RequireStory: false})

	item := QueueItem{ID: "dq-0001", TaskText: "task"}
	_, err := engine.Dispatch(context.Background(), item)
	if err == nil {
		t.Fatal("Dispatch() should fail when worker creation fails")
	}
	if !errors.Is(err, disp.createErr) {
		t.Errorf("error should wrap worker creation error, got: %v", err)
	}
}

func TestDispatchEngine_ContextPassedToGenerator(t *testing.T) {
	t.Parallel()

	sg := &mockStoryGenerator{result: &StoryResult{Branch: "stories/test"}}
	disp := &mockDispatcher{workerName: "test-fox"}
	engine := NewDispatchEngine(disp, sg, DispatchConfig{RequireStory: true})

	item := QueueItem{ID: "dq-ctx", TaskText: "do thing", Context: "extra context"}
	_, err := engine.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !strings.Contains(sg.taskIn, "do thing") {
		t.Error("generator should receive task text")
	}
	if !strings.Contains(sg.taskIn, "extra context") {
		t.Error("generator should receive item context")
	}
}

func TestBuildTaskDescriptionWithStories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     QueueItem
		stories  *StoryResult
		contains []string
		excludes []string
	}{
		{
			name:    "nil stories falls back to BuildTaskDescription",
			item:    QueueItem{TaskText: "fix bug"},
			stories: nil,
			contains: []string{
				"Task: fix bug",
				"Sign all commits",
			},
			excludes: []string{
				"Generated Story Files",
				"Check out branch",
			},
		},
		{
			name: "with stories includes branch and file references",
			item: QueueItem{
				TaskText:           "implement auth",
				Context:            "uses JWT",
				AcceptanceCriteria: []string{"login works"},
				Scope:              "internal/auth",
			},
			stories: &StoryResult{
				Branch:    "stories/dq-1234",
				FilePaths: []string{"docs/stories/14.1.story.md"},
			},
			contains: []string{
				"Task: implement auth",
				"Context: uses JWT",
				"Generated Story Files (branch: stories/dq-1234)",
				"docs/stories/14.1.story.md",
				"Check out branch",
				"- login works",
				"Scope: internal/auth",
				"Sign all commits",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildTaskDescriptionWithStories(tt.item, tt.stories)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("missing %q in:\n%s", s, got)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("should not contain %q in:\n%s", s, got)
				}
			}
		})
	}
}

func TestDispatchConfig_DefaultRequireStoryFalse(t *testing.T) {
	t.Parallel()

	var cfg DispatchConfig
	if cfg.RequireStory {
		t.Error("RequireStory should default to false")
	}
}
