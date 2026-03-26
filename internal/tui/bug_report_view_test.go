package tui

import (
	"errors"
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

func enterPreviewState(t *testing.T, v *BugReportView) {
	t.Helper()
	v.textArea.SetValue("test bug description")
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.state != bugReportPreview {
		t.Fatal("failed to enter preview state")
	}
}

func TestBugReportView_PreviewShowsSubmissionKeys(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	output := v.View()

	if !strings.Contains(output, "[b] Open in browser") {
		t.Error("preview should show browser option")
	}
	if !strings.Contains(output, "[f] Save to file") {
		t.Error("preview should show file save option")
	}
}

func TestBugReportView_PreviewShowsAPIOption_WhenTokenSet(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.hasToken = true
	enterPreviewState(t, v)

	output := v.View()

	if !strings.Contains(output, "[s] Submit via GitHub") {
		t.Error("preview should show API submit option when token is set")
	}
}

func TestBugReportView_PreviewHidesAPIOption_WhenNoToken(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.hasToken = false
	enterPreviewState(t, v)

	output := v.View()

	if strings.Contains(output, "[s] Submit via GitHub") {
		t.Error("preview should hide API submit option when no token")
	}
}

func TestBugReportView_BKeyFromPreview_TriggersCmd(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'b' in preview")
	}
	if v.state != bugReportSubmitting {
		t.Errorf("state = %d, want bugReportSubmitting", v.state)
	}
}

func TestBugReportView_FKeyFromPreview_TriggersCmd(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'f' in preview")
	}
	if v.state != bugReportSubmitting {
		t.Errorf("state = %d, want bugReportSubmitting", v.state)
	}
}

func TestBugReportView_SKeyFromPreview_NoOpWithoutToken(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.hasToken = false
	enterPreviewState(t, v)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if cmd != nil {
		t.Error("expected nil cmd for 's' when no token")
	}
	if v.state != bugReportPreview {
		t.Errorf("state should remain bugReportPreview, got %d", v.state)
	}
}

func TestBugReportView_SuccessMsg_TransitionsToSuccess(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	v.Update(BugReportSubmittedMsg{Method: "browser", Details: "https://example.com"})

	if v.state != bugReportSuccess {
		t.Errorf("state = %d, want bugReportSuccess", v.state)
	}

	output := v.View()
	if !strings.Contains(output, "Opening GitHub in your browser...") {
		t.Error("success view should show browser confirmation message")
	}
	if !strings.Contains(output, "Press any key to return") {
		t.Error("success view should show return hint")
	}
}

func TestBugReportView_SuccessMsg_APIShowsURL(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	v.Update(BugReportSubmittedMsg{Method: "api", Details: "https://github.com/arcavenae/ThreeDoors/issues/42"})

	if v.state != bugReportSuccess {
		t.Errorf("state = %d, want bugReportSuccess", v.state)
	}

	output := v.View()
	if !strings.Contains(output, "Issue created:") {
		t.Error("success view should show issue URL for API submission")
	}
	if !strings.Contains(output, "issues/42") {
		t.Error("success view should include issue URL")
	}
}

func TestBugReportView_SuccessMsg_FileShowsPath(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	v.Update(BugReportSubmittedMsg{Method: "file", Details: "~/.threedoors/bug-reports/bug-2025-01-15T10-00-30Z.md"})

	output := v.View()
	if !strings.Contains(output, "Report saved to") {
		t.Error("success view should show file path")
	}
}

func TestBugReportView_AnyKeyFromSuccess_Returns(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)
	v.Update(BugReportSubmittedMsg{Method: "file", Details: "/path"})

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if cmd == nil {
		t.Fatal("expected non-nil cmd from success state")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestBugReportView_ErrorMsg_TransitionsToError(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)

	v.Update(BugReportErrorMsg{Method: "browser", Err: errors.New("failed to open")})

	if v.state != bugReportError {
		t.Errorf("state = %d, want bugReportError", v.state)
	}
	if v.failedMethod != "browser" {
		t.Errorf("failedMethod = %q, want %q", v.failedMethod, "browser")
	}
}

func TestBugReportView_ErrorView_BrowserFallback(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)
	v.Update(BugReportErrorMsg{Method: "browser", Err: errors.New("failed")})

	output := v.View()
	if !strings.Contains(output, "Could not submit via browser") {
		t.Error("error view should show error message")
	}
	if !strings.Contains(output, "[b] Copy URL to clipboard") {
		t.Error("error view should offer clipboard fallback for browser failure")
	}
	if !strings.Contains(output, "[f] Save to file") {
		t.Error("error view should offer file save fallback")
	}
}

func TestBugReportView_ErrorView_APIFallback(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)
	v.Update(BugReportErrorMsg{Method: "api", Err: errors.New("unauthorized")})

	output := v.View()
	if !strings.Contains(output, "[b] Try browser instead") {
		t.Error("error view should offer browser fallback for API failure")
	}
	if !strings.Contains(output, "[f] Save to file") {
		t.Error("error view should offer file save fallback")
	}
}

func TestBugReportView_ErrorView_EscCancels(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)
	v.Update(BugReportErrorMsg{Method: "file", Err: errors.New("disk full")})

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc in error state")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestBugReportView_ErrorView_FKeyFallbackToFile(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	enterPreviewState(t, v)
	v.Update(BugReportErrorMsg{Method: "api", Err: errors.New("unauthorized")})

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'f' in error state")
	}
	if v.state != bugReportSubmitting {
		t.Errorf("state = %d, want bugReportSubmitting", v.state)
	}
}

func TestBugReportView_SubmittingView(t *testing.T) {
	t.Parallel()

	v := newTestBugReportView()
	v.state = bugReportSubmitting

	output := v.View()
	if !strings.Contains(output, "Submitting your report...") {
		t.Error("submitting view should show progress message")
	}
}
