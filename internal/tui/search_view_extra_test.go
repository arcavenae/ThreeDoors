package tui

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestSearchView_SetSyncLog(t *testing.T) {
	t.Parallel()
	sv := NewSearchView(makePool("t1"), nil, nil, nil, nil)
	sv.SetSyncLog(nil)
	if sv.syncLog != nil {
		t.Error("expected nil sync log")
	}
}

func TestSearchView_RestoreState_IndexClamped(t *testing.T) {
	t.Parallel()
	pool := makePool("alpha")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	sv.RestoreState("alpha", 5)
	if sv.selectedIndex != 0 {
		t.Errorf("expected clamped index 0, got %d", sv.selectedIndex)
	}
}

func TestSearchView_FilterTasks_Empty(t *testing.T) {
	t.Parallel()
	pool := makePool("alpha", "beta")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	results := sv.filterTasks("")
	if results != nil {
		t.Error("expected nil for empty query")
	}
}

func TestSearchView_FilterTasks_Match(t *testing.T) {
	t.Parallel()
	pool := makePool("alpha", "beta", "alphabet")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	results := sv.filterTasks("alph")
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchView_CheckCommandMode(t *testing.T) {
	t.Parallel()
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	sv.textInput.SetValue(":help")
	sv.checkCommandMode()
	if !sv.isCommandMode {
		t.Error("expected command mode for ':help'")
	}
	sv.textInput.SetValue("search")
	sv.checkCommandMode()
	if sv.isCommandMode {
		t.Error("expected non-command mode for 'search'")
	}
}

func TestParseCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantCmd string
		wantArg string
	}{
		{":help", "help", ""},
		{":add my task", "add", "my task"},
		{":mood happy", "mood", "happy"},
		{":quit", "quit", ""},
		{":", "", ""},
	}

	for _, tt := range tests {
		cmd, args := parseCommand(tt.input)
		if cmd != tt.wantCmd {
			t.Errorf("parseCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.wantCmd)
		}
		if args != tt.wantArg {
			t.Errorf("parseCommand(%q) args = %q, want %q", tt.input, args, tt.wantArg)
		}
	}
}

func TestSearchView_ExecuteCommand_Add(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	// :add with no args → AddTaskPromptMsg
	sv.textInput.SetValue(":add")
	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal("expected non-nil cmd for :add")
		return
	}
	msg := cmd()
	if _, ok := msg.(AddTaskPromptMsg); !ok {
		t.Errorf("expected AddTaskPromptMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_AddWithText(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":add my new task")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(TaskAddedMsg); !ok {
		t.Errorf("expected TaskAddedMsg, got %T", msg)
	} else if m.Task.Text != "my new task" {
		t.Errorf("expected task text 'my new task', got %q", m.Task.Text)
	}
}

func TestSearchView_ExecuteCommand_AddWhy(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":add --why")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(AddTaskWithContextPromptMsg); !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Mood(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":mood")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowMoodMsg); !ok {
		t.Errorf("expected ShowMoodMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_MoodWithArg(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":mood happy")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(MoodCapturedMsg); !ok {
		t.Errorf("expected MoodCapturedMsg, got %T", msg)
	} else if m.Mood != "happy" {
		t.Errorf("expected mood 'happy', got %q", m.Mood)
	}
}

func TestSearchView_ExecuteCommand_Help(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":help")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowHelpMsg); !ok {
		t.Errorf("expected ShowHelpMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Quit(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":quit")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(RequestQuitMsg); !ok {
		t.Errorf("expected RequestQuitMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Unknown(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":foobar")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if m.Text == "" {
		t.Error("expected non-empty error text")
	}
}

func TestSearchView_ExecuteCommand_Dashboard(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":dashboard")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowInsightsMsg); !ok {
		t.Errorf("expected ShowInsightsMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Goals(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":goals")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowValuesSetupMsg); !ok {
		t.Errorf("expected ShowValuesSetupMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_GoalsEdit(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":goals edit")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowValuesEditMsg); !ok {
		t.Errorf("expected ShowValuesEditMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Tag(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":tag")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(ShowTagViewMsg); !ok {
		t.Errorf("expected ShowTagViewMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Stats_NoTracker(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":stats")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if m.Text == "" {
		t.Error("expected non-empty stats text")
	}
}

func TestSearchView_ExecuteCommand_Stats_WithTracker(t *testing.T) {
	pool := makePool("t1")
	tracker := core.NewSessionTracker()
	sv := NewSearchView(pool, tracker, nil, nil, nil)

	sv.textInput.SetValue(":stats")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if m.Text == "" {
		t.Error("expected non-empty stats text")
	}
}

func TestSearchView_ExecuteCommand_Health_NoChecker(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":health")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if m.Text == "" {
		t.Error("expected non-empty health text")
	}
}

func TestSearchView_ExecuteCommand_SyncLog_NoLog(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":synclog")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if m.Text == "" {
		t.Error("expected non-empty synclog text")
	}
}

func TestSearchView_ExecuteCommand_Insights_Mood(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":insights mood")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Insights_Avoidance(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":insights avoidance")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Insights_Default(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":insights all")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_AddCtx(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":add-ctx")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(AddTaskWithContextPromptMsg); !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_AddCtxWithArgs(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":add-ctx my context task")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(AddTaskWithContextPromptMsg); !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	} else if m.PrefilledText != "my context task" {
		t.Errorf("expected prefilled text 'my context task', got %q", m.PrefilledText)
	}
}

func TestSearchView_ExecuteCommand_Exit(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":exit")
	cmd := sv.executeCommand()
	msg := cmd()
	if _, ok := msg.(RequestQuitMsg); !ok {
		t.Errorf("expected RequestQuitMsg, got %T", msg)
	}
}

func TestSearchView_ExecuteCommand_Empty(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":")
	cmd := sv.executeCommand()
	if cmd != nil {
		t.Error("expected nil cmd for empty command")
	}
}

func TestSearchView_ExecuteCommand_AddWhyWithText(t *testing.T) {
	pool := makePool("t1")
	sv := NewSearchView(pool, nil, nil, nil, nil)

	sv.textInput.SetValue(":add --why my important task")
	cmd := sv.executeCommand()
	msg := cmd()
	if m, ok := msg.(AddTaskWithContextPromptMsg); !ok {
		t.Errorf("expected AddTaskWithContextPromptMsg, got %T", msg)
	} else if m.PrefilledText != "my important task" {
		t.Errorf("expected prefilled text 'my important task', got %q", m.PrefilledText)
	}
}

func TestSearchView_View(t *testing.T) {
	pool := makePool("alpha", "beta")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	sv.SetWidth(80)
	view := sv.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
