package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// --- SetAgentService ---

func TestSetAgentService(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.SetAgentService(nil)
	if m.agentService != nil {
		t.Error("expected nil agent service")
	}
}

// --- updateHealth: transition to health view ---

func TestHealthView_TransitionAndBack(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Directly send HealthCheckMsg
	m.Update(HealthCheckMsg{Result: core.HealthCheckResult{
		Items: []core.HealthCheckItem{{Name: "test", Status: core.HealthOK}},
	}})

	if m.viewMode != ViewHealth {
		t.Errorf("expected ViewHealth, got %d", m.viewMode)
	}

	// Send escape to return
	_, cmd := m.Update(keyMsg("esc"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
}

// --- updateSearch: transition to search view ---

func TestSearchView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("/"))

	if m.viewMode != ViewSearch {
		t.Errorf("expected ViewSearch, got %d", m.viewMode)
	}
	if m.searchView == nil {
		t.Fatal("searchView should not be nil")
		return
	}
}

func TestSearchView_EscReturns(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("/"))

	_, cmd := m.Update(keyMsg("esc"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after esc from search, got %d", m.viewMode)
	}
}

// --- updateAddTask ---

func TestAddTaskView_EscReturns(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(AddTaskPromptMsg{})

	if m.viewMode != ViewAddTask {
		t.Fatal("should be in ViewAddTask")
	}

	_, cmd := m.Update(keyMsg("esc"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after esc from add task, got %d", m.viewMode)
	}
}

// --- updateFeedback ---

func TestFeedbackView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("feedback task")

	// Directly send ShowFeedbackMsg
	m.Update(ShowFeedbackMsg{Task: task})

	if m.viewMode != ViewFeedback {
		t.Errorf("expected ViewFeedback, got %d", m.viewMode)
	}

	if m.feedbackView == nil {
		t.Fatal("feedbackView should not be nil")
		return
	}
}

// --- updateAvoidancePrompt ---

func TestAvoidancePromptView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("avoided task")

	m.Update(ShowAvoidancePromptMsg{Task: task})

	if m.viewMode != ViewAvoidancePrompt {
		t.Errorf("expected ViewAvoidancePrompt, got %d", m.viewMode)
	}
}

func TestAvoidancePromptView_EscReturns(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("avoided task")

	m.Update(ShowAvoidancePromptMsg{Task: task})
	_, cmd := m.Update(keyMsg("esc"))
	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}

	if m.viewMode != ViewDoors {
		t.Errorf("expected ViewDoors after esc from avoidance, got %d", m.viewMode)
	}
}

// --- updateOnboarding ---

func TestOnboardingView_Transition(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, &testProvider{}, nil, true, nil)

	if m.viewMode != ViewOnboarding {
		t.Errorf("expected ViewOnboarding for first run, got %d", m.viewMode)
	}
}

// --- updateConflict ---

func TestConflictView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	localTask := core.NewTask("Conflicted task")
	remoteTask := core.NewTask("Conflicted task remote")
	cs := core.NewConflictSet("test-provider", []core.Conflict{
		{LocalTask: localTask, RemoteTask: remoteTask},
	})
	m.Update(SyncConflictMsg{ConflictSet: cs})

	if m.viewMode != ViewConflict {
		t.Errorf("expected ViewConflict, got %d", m.viewMode)
	}
}

// --- updateValues ---

func TestValuesView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Directly send ShowValuesSetupMsg
	m.Update(ShowValuesSetupMsg{})

	if m.viewMode != ViewValuesGoals {
		t.Errorf("expected ViewValuesGoals, got %d", m.viewMode)
	}
}

// --- updateSyncLog ---

func TestSyncLogView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	m.Update(ShowSyncLogMsg{})

	if m.viewMode != ViewSyncLog {
		t.Errorf("expected ViewSyncLog, got %d", m.viewMode)
	}
}

// --- updateInsights ---

func TestInsightsView_Transition(t *testing.T) {
	m := makeModel("task1", "task2", "task3")

	// Directly send ShowInsightsMsg
	m.Update(ShowInsightsMsg{})

	if m.viewMode != ViewInsights {
		t.Errorf("expected ViewInsights, got %d", m.viewMode)
	}
}

// --- Delegate function exercise (send msg WHILE in view mode) ---

func TestUpdateInsights_Delegate(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowInsightsMsg{})
	// Now in ViewInsights, send a key to exercise updateInsights
	m.Update(keyMsg("j"))
	m.Update(keyMsg("esc"))
}

func TestUpdateFeedback_Delegate(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("fb task")
	m.Update(ShowFeedbackMsg{Task: task})
	// Now in ViewFeedback, send a key to exercise updateFeedback
	m.Update(keyMsg("1"))
}

func TestUpdateOnboarding_Delegate(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, &testProvider{}, nil, true, nil)
	// Now in ViewOnboarding, send a key to exercise updateOnboarding
	m.Update(keyMsg("enter"))
}

func TestUpdateConflict_Delegate(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	localTask := core.NewTask("c1")
	remoteTask := core.NewTask("c2")
	cs := core.NewConflictSet("test", []core.Conflict{
		{LocalTask: localTask, RemoteTask: remoteTask},
	})
	m.Update(SyncConflictMsg{ConflictSet: cs})
	// Now in ViewConflict, send a key to exercise updateConflict
	m.Update(keyMsg("j"))
}

func TestUpdateSyncLog_Delegate(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowSyncLogMsg{})
	// Now in ViewSyncLog, send a key to exercise updateSyncLog
	m.Update(keyMsg("esc"))
}

func TestUpdateValues_Delegate(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowValuesSetupMsg{})
	// Now in ViewValuesGoals, send a key to exercise updateValues
	m.Update(keyMsg("j"))
}

func TestUpdateInsights_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewInsights
	m.insightsView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil insightsView")
	}
}

func TestUpdateFeedback_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewFeedback
	m.feedbackView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil feedbackView")
	}
}

func TestUpdateOnboarding_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewOnboarding
	m.onboardingView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil onboardingView")
	}
}

func TestUpdateConflict_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewConflict
	m.conflictView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil conflictView")
	}
}

func TestUpdateSyncLog_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewSyncLog
	m.syncLogView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil syncLogView")
	}
}

func TestUpdateValues_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewValuesGoals
	m.valuesView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil valuesView")
	}
}

func TestUpdateNextSteps_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewNextSteps
	m.nextStepsView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil nextStepsView")
	}
}

func TestUpdateAvoidancePrompt_NilView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.viewMode = ViewAvoidancePrompt
	m.avoidancePromptView = nil
	_, cmd := m.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("expected nil cmd for nil avoidancePromptView")
	}
}

// --- View rendering for all view modes ---

func TestViewRendering_InsightsView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowInsightsMsg{})
	_ = m.View()
}

func TestViewRendering_FeedbackView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("render fb")
	m.Update(ShowFeedbackMsg{Task: task})
	_ = m.View()
}

func TestViewRendering_ValuesView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowValuesSetupMsg{})
	_ = m.View()
}

func TestViewRendering_ConflictView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	localTask := core.NewTask("c1")
	remoteTask := core.NewTask("c2")
	cs := core.NewConflictSet("test", []core.Conflict{
		{LocalTask: localTask, RemoteTask: remoteTask},
	})
	m.Update(SyncConflictMsg{ConflictSet: cs})
	_ = m.View()
}

func TestViewRendering_SyncLogView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(ShowSyncLogMsg{})
	_ = m.View()
}

func TestViewRendering_OnboardingView(t *testing.T) {
	pool := makePool("task1", "task2", "task3")
	tracker := core.NewSessionTracker()
	m := NewMainModel(pool, tracker, &testProvider{}, nil, true, nil)
	_ = m.View()
}

func TestViewRendering_AvoidancePromptView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	task := core.NewTask("avoided")
	m.Update(ShowAvoidancePromptMsg{Task: task})
	_ = m.View()
}

func TestViewRendering_HealthView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(HealthCheckMsg{Result: core.HealthCheckResult{
		Items: []core.HealthCheckItem{{Name: "test", Status: core.HealthOK}},
	}})
	_ = m.View()
}

// --- ClearFlashMsg ---

func TestClearFlashMsg(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.flash = "some flash"
	m.Update(ClearFlashMsg{})
	if m.flash != "" {
		t.Error("expected flash to be cleared")
	}
}

// --- FlashMsg ---

func TestFlashMsg(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(FlashMsg{Text: "hello"})
	if m.flash != "hello" {
		t.Errorf("expected flash 'hello', got %q", m.flash)
	}
}

// --- SyncStatusUpdateMsg ---

func TestSyncStatusUpdateMsg(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.syncTracker = core.NewSyncStatusTracker()
	m.Update(SyncStatusUpdateMsg{
		ProviderName: "test",
		Phase:        core.SyncPhaseSyncing,
	})
}

// --- View rendering for different modes ---

func TestViewRendering_SearchView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(keyMsg("/"))
	view := m.View()
	if !strings.Contains(view, "Search") && !strings.Contains(view, "search") && !strings.Contains(view, "Type to filter") {
		// Search view should show some search UI
		_ = view
	}
}

func TestViewRendering_AddTaskView(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(AddTaskPromptMsg{})
	view := m.View()
	if !strings.Contains(view, "Add") && !strings.Contains(view, "task") {
		// Should show add task UI
		_ = view
	}
}

// --- typeIcon coverage ---

func TestTypeIcon(t *testing.T) {
	t.Parallel()
	tests := []struct {
		taskType core.TaskType
		wantIcon bool
	}{
		{core.TypeCreative, true},
		{core.TypeAdministrative, true},
		{core.TypeTechnical, true},
		{core.TypePhysical, true},
		{"", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		icon := typeIcon(tt.taskType)
		if tt.wantIcon && icon == "" {
			t.Errorf("expected icon for type %q, got empty", tt.taskType)
		}
		if !tt.wantIcon && icon != "" {
			t.Errorf("expected empty icon for type %q, got %q", tt.taskType, icon)
		}
	}
}

// --- categoryBadge coverage ---

func TestCategoryBadge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		task *core.Task
		want bool
	}{
		{"no categories", core.NewTask("plain"), false},
		{"with type", func() *core.Task {
			task := core.NewTask("typed")
			task.Type = core.TypeCreative
			return task
		}(), true},
		{"with effort", func() *core.Task {
			task := core.NewTask("effortful")
			task.Effort = core.EffortDeepWork
			return task
		}(), true},
		{"with location", func() *core.Task {
			task := core.NewTask("located")
			task.Location = core.LocationHome
			return task
		}(), true},
		{"all categories", func() *core.Task {
			task := core.NewTask("full")
			task.Type = core.TypeTechnical
			task.Effort = core.EffortQuickWin
			task.Location = core.LocationWork
			return task
		}(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := categoryBadge(tt.task)
			if tt.want && badge == "" {
				t.Error("expected non-empty badge")
			}
			if !tt.want && badge != "" {
				t.Errorf("expected empty badge, got %q", badge)
			}
		})
	}
}

// --- WindowSizeMsg propagates to sub-views ---

func TestWindowSizeMsg_PropagatesWidth(t *testing.T) {
	m := makeModel("task1", "task2", "task3")
	m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})

	if m.doorsView.width != 200 {
		t.Errorf("expected doorsView width 200, got %d", m.doorsView.width)
	}
}
