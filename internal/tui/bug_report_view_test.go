package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestBugReportView() *BugReportView {
	env := EnvironmentInfo{
		Version:         "1.0.0",
		Commit:          "abc1234",
		BuildDate:       "2025-01-15",
		GoVersion:       "go1.25.4",
		OS:              "darwin",
		Arch:            "arm64",
		TerminalWidth:   80,
		TerminalHeight:  24,
		CurrentView:     "Doors",
		ThemeName:       "modern",
		TaskCount:       5,
		ProviderCount:   2,
		SessionDuration: 10 * time.Minute,
	}
	return NewBugReportView(env, "breadcrumb trail here")
}

func TestBugReportView_InitialState(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()

	if v.state != bugReportInput {
		t.Errorf("initial state = %d, want bugReportInput (0)", v.state)
	}
}

func TestBugReportView_ViewInput_ShowsEnvironmentSummary(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	output := v.View()

	checks := []string{
		"Bug Report",
		"Environment data that will be included",
		"1.0.0",
		"abc1234",
		"darwin/arm64",
		"80x24",
		"modern",
		"No task names, content, or personal data will be included",
		"[Enter] Preview",
		"[Esc] Cancel",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("input view missing %q", check)
		}
	}
}

func TestBugReportView_EscFromInput_Cancels(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestBugReportView_EnterWithEmptyDescription_NoOp(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected nil cmd when description is empty")
	}
	if v.state != bugReportInput {
		t.Error("should remain in input state")
	}
}

func TestBugReportView_EnterWithDescription_TransitionsToPreview(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("The app froze when I selected a door")

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected nil cmd for state transition")
	}
	if v.state != bugReportPreview {
		t.Errorf("state = %d, want bugReportPreview (1)", v.state)
	}
	if v.report == nil {
		t.Fatal("report should be set after entering preview")
	}
	if v.report.Description != "The app froze when I selected a door" {
		t.Errorf("description = %q, want original text", v.report.Description)
	}
}

func TestBugReportView_PreviewShowsMarkdown(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("Something broke")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	output := v.View()

	if !strings.Contains(output, "Bug Report Preview") {
		t.Error("preview should show 'Bug Report Preview' header")
	}
	if !strings.Contains(output, "[Esc] Back to edit") {
		t.Error("preview should show Esc hint")
	}
}

func TestBugReportView_EscFromPreview_ReturnsToInput(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("Something broke")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.state != bugReportPreview {
		t.Fatal("should be in preview state")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if v.state != bugReportInput {
		t.Errorf("state = %d, want bugReportInput (0) after Esc from preview", v.state)
	}
}

func TestBugReportView_QFromPreview_Cancels(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("Something broke")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for q in preview")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestBugReportView_SetWidth(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.SetWidth(100)

	if v.width != 100 {
		t.Errorf("width = %d, want 100", v.width)
	}
}

func TestBugReportView_SetHeight(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.SetHeight(50)

	if v.height != 50 {
		t.Errorf("height = %d, want 50", v.height)
	}
}

func TestBugReportView_ReportTimestamp_IsUTC(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("test")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.report.Timestamp.Location() != time.UTC {
		t.Error("report timestamp should be UTC")
	}
}

func TestBugReportView_PreviewIncludesBreadcrumbs(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.textArea.SetValue("test bug")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.report.Breadcrumbs != "breadcrumb trail here" {
		t.Errorf("breadcrumbs = %q, want %q", v.report.Breadcrumbs, "breadcrumb trail here")
	}
}

func TestBugCommand_ProducesShowBugReportMsg(t *testing.T) {
	t.Parallel()

	sv := NewSearchView(nil, nil, nil, nil, nil)
	sv.textInput.SetValue(":bug")
	sv.isCommandMode = true
	cmd := sv.executeCommand()

	if cmd == nil {
		t.Fatal(":bug should produce a command")
	}

	msg := cmd()
	if _, ok := msg.(ShowBugReportMsg); !ok {
		t.Errorf("expected ShowBugReportMsg, got %T", msg)
	}
}
