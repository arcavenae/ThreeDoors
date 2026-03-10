package reminders

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockExecutor returns a CommandExecutor that returns canned responses.
type mockExecutor struct {
	output string
	err    error
	// capturedScript records the last script passed to Execute.
	capturedScript string
}

func (m *mockExecutor) Execute(_ context.Context, script string) (string, error) {
	m.capturedScript = script
	return m.output, m.err
}

func TestReadReminders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		listName  string
		output    string
		execErr   error
		wantCount int
		wantErr   bool
		checkFunc func(t *testing.T, reminders []ReminderJSON)
	}{
		{
			name:      "successful read with two reminders",
			listName:  "Work",
			output:    `[{"id":"x-apple-reminder://ABC123","name":"Buy groceries","body":"Milk, eggs, bread","dueDate":"2025-01-15T10:00:00.000Z","priority":1,"completed":false,"flagged":true,"creationDate":"2025-01-01T08:00:00.000Z","modificationDate":"2025-01-10T12:00:00.000Z"},{"id":"x-apple-reminder://DEF456","name":"Call dentist","body":"","dueDate":null,"priority":0,"completed":false,"flagged":false,"creationDate":"2025-01-02T09:00:00.000Z","modificationDate":"2025-01-02T09:00:00.000Z"}]`,
			wantCount: 2,
			checkFunc: func(t *testing.T, reminders []ReminderJSON) {
				t.Helper()
				if reminders[0].ID != "x-apple-reminder://ABC123" {
					t.Errorf("first reminder ID = %q, want %q", reminders[0].ID, "x-apple-reminder://ABC123")
				}
				if reminders[0].Name != "Buy groceries" {
					t.Errorf("first reminder name = %q, want %q", reminders[0].Name, "Buy groceries")
				}
				if reminders[0].Body != "Milk, eggs, bread" {
					t.Errorf("first reminder body = %q, want %q", reminders[0].Body, "Milk, eggs, bread")
				}
				if reminders[0].DueDate == nil || *reminders[0].DueDate != "2025-01-15T10:00:00.000Z" {
					t.Errorf("first reminder dueDate unexpected")
				}
				if reminders[0].Priority != 1 {
					t.Errorf("first reminder priority = %d, want 1", reminders[0].Priority)
				}
				if !reminders[0].Flagged {
					t.Error("first reminder flagged = false, want true")
				}
				if reminders[1].DueDate != nil {
					t.Errorf("second reminder dueDate = %v, want nil", reminders[1].DueDate)
				}
				if reminders[1].Body != "" {
					t.Errorf("second reminder body = %q, want empty", reminders[1].Body)
				}
			},
		},
		{
			name:      "empty list returns empty slice",
			listName:  "Empty",
			output:    `[]`,
			wantCount: 0,
		},
		{
			name:     "executor error propagates",
			listName: "Work",
			execErr:  errors.New("osascript: command not found"),
			wantErr:  true,
		},
		{
			name:     "invalid JSON returns error",
			listName: "Work",
			output:   `not valid json`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			reminders, err := ReadReminders(context.Background(), mock, tt.listName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadReminders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(reminders) != tt.wantCount {
				t.Fatalf("got %d reminders, want %d", len(reminders), tt.wantCount)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, reminders)
			}
		})
	}
}

func TestReadLists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		output    string
		execErr   error
		wantLists []string
		wantErr   bool
	}{
		{
			name:      "successful read with three lists",
			output:    `["Reminders","Work","Shopping"]`,
			wantLists: []string{"Reminders", "Work", "Shopping"},
		},
		{
			name:      "empty lists array",
			output:    `[]`,
			wantLists: []string{},
		},
		{
			name:    "executor error propagates",
			execErr: errors.New("permission denied"),
			wantErr: true,
		},
		{
			name:    "invalid JSON returns error",
			output:  `["unclosed`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			lists, err := ReadLists(context.Background(), mock)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadLists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(lists) != len(tt.wantLists) {
				t.Fatalf("got %d lists, want %d", len(lists), len(tt.wantLists))
			}
			for i, want := range tt.wantLists {
				if lists[i] != want {
					t.Errorf("lists[%d] = %q, want %q", i, lists[i], want)
				}
			}
		})
	}
}

func TestCompleteReminder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reminderID string
		output     string
		execErr    error
		wantErr    bool
	}{
		{
			name:       "successful completion",
			reminderID: "x-apple-reminder://ABC123",
			output:     `{"success":true}`,
		},
		{
			name:       "reminder not found",
			reminderID: "x-apple-reminder://NOTFOUND",
			output:     `{"success":false,"error":"reminder not found"}`,
			wantErr:    true,
		},
		{
			name:       "executor error propagates",
			reminderID: "x-apple-reminder://ABC123",
			execErr:    errors.New("timeout"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			err := CompleteReminder(context.Background(), mock, tt.reminderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteReminder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateReminder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rName    string
		body     string
		priority int
		listName string
		output   string
		execErr  error
		wantID   string
		wantErr  bool
	}{
		{
			name:     "successful creation",
			rName:    "New task",
			body:     "Some details",
			priority: 1,
			listName: "Work",
			output:   `{"success":true,"id":"x-apple-reminder://NEW789"}`,
			wantID:   "x-apple-reminder://NEW789",
		},
		{
			name:     "executor error propagates",
			rName:    "New task",
			body:     "",
			priority: 0,
			listName: "Work",
			execErr:  errors.New("access denied"),
			wantErr:  true,
		},
		{
			name:     "creation failure",
			rName:    "New task",
			body:     "",
			priority: 0,
			listName: "NonExistent",
			output:   `{"success":false,"error":"list not found"}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			id, err := CreateReminder(context.Background(), mock, tt.rName, tt.body, tt.priority, tt.listName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateReminder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if id != tt.wantID {
				t.Errorf("CreateReminder() id = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestUpdateReminder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reminderID string
		rName      string
		body       string
		priority   int
		output     string
		execErr    error
		wantErr    bool
	}{
		{
			name:       "successful update",
			reminderID: "x-apple-reminder://ABC123",
			rName:      "Updated task",
			body:       "New details",
			priority:   5,
			output:     `{"success":true}`,
		},
		{
			name:       "reminder not found",
			reminderID: "x-apple-reminder://GONE",
			rName:      "Updated task",
			body:       "",
			priority:   0,
			output:     `{"success":false,"error":"reminder not found"}`,
			wantErr:    true,
		},
		{
			name:       "executor error propagates",
			reminderID: "x-apple-reminder://ABC123",
			rName:      "Updated task",
			body:       "",
			priority:   0,
			execErr:    errors.New("timeout"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			err := UpdateReminder(context.Background(), mock, tt.reminderID, tt.rName, tt.body, tt.priority)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateReminder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteReminder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reminderID string
		output     string
		execErr    error
		wantErr    bool
	}{
		{
			name:       "successful deletion",
			reminderID: "x-apple-reminder://ABC123",
			output:     `{"success":true}`,
		},
		{
			name:       "reminder not found",
			reminderID: "x-apple-reminder://GONE",
			output:     `{"success":false,"error":"reminder not found"}`,
			wantErr:    true,
		},
		{
			name:       "executor error propagates",
			reminderID: "x-apple-reminder://ABC123",
			execErr:    errors.New("osascript crashed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{output: tt.output, err: tt.execErr}
			err := DeleteReminder(context.Background(), mock, tt.reminderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteReminder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEscapeJXA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "double quotes escaped",
			input: `say "hello"`,
			want:  `say \"hello\"`,
		},
		{
			name:  "backslashes escaped",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
		{
			name:  "newlines escaped",
			input: "line1\nline2",
			want:  `line1\nline2`,
		},
		{
			name:  "carriage returns escaped",
			input: "line1\rline2",
			want:  `line1\rline2`,
		},
		{
			name:  "injection attempt escaped",
			input: `"; app.delete(app.lists()); "`,
			want:  `\"; app.delete(app.lists()); \"`,
		},
		{
			name:  "combined special characters",
			input: "test\n\"quote\"\r\\slash",
			want:  `test\n\"quote\"\r\\slash`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := escapeJXA(tt.input)
			if got != tt.want {
				t.Errorf("escapeJXA(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScriptReadReminders_ContainsListName(t *testing.T) {
	t.Parallel()
	script := scriptReadReminders("Shopping")
	if !strings.Contains(script, `"Shopping"`) {
		t.Error("script does not contain list name")
	}
	if !strings.Contains(script, "JSON.stringify") {
		t.Error("script does not use JSON.stringify")
	}
	if !strings.Contains(script, "completed: false") {
		t.Error("script does not filter incomplete reminders")
	}
}

func TestScriptReadLists_UsesJSONStringify(t *testing.T) {
	t.Parallel()
	script := scriptReadLists()
	if !strings.Contains(script, "JSON.stringify") {
		t.Error("script does not use JSON.stringify")
	}
}

func TestScriptCompleteReminder_ContainsID(t *testing.T) {
	t.Parallel()
	script := scriptCompleteReminder("x-apple-reminder://ABC123")
	if !strings.Contains(script, "x-apple-reminder://ABC123") {
		t.Error("script does not contain reminder ID")
	}
	if !strings.Contains(script, "completed = true") {
		t.Error("script does not set completed = true")
	}
}

func TestScriptCreateReminder_ContainsAllFields(t *testing.T) {
	t.Parallel()
	script := scriptCreateReminder("Buy milk", "2% organic", 1, "Shopping")
	if !strings.Contains(script, "Buy milk") {
		t.Error("script does not contain reminder name")
	}
	if !strings.Contains(script, "2% organic") {
		t.Error("script does not contain body")
	}
	if !strings.Contains(script, "priority: 1") {
		t.Error("script does not contain priority")
	}
	if !strings.Contains(script, `"Shopping"`) {
		t.Error("script does not contain list name")
	}
}

func TestScriptCreateReminder_EscapesInput(t *testing.T) {
	t.Parallel()
	script := scriptCreateReminder(`Buy "special" milk`, `with "quotes"`, 0, `My "List"`)
	if strings.Contains(script, `""special""`) {
		t.Error("script contains unescaped double quotes in name")
	}
	if !strings.Contains(script, `\"special\"`) {
		t.Error("script does not properly escape quotes in name")
	}
}

func TestScriptUpdateReminder_ContainsFields(t *testing.T) {
	t.Parallel()
	script := scriptUpdateReminder("x-apple-reminder://ABC", "Buy milk", "2% organic", 5)
	if !strings.Contains(script, "x-apple-reminder://ABC") {
		t.Error("script does not contain reminder ID")
	}
	if !strings.Contains(script, "Buy milk") {
		t.Error("script does not contain updated name")
	}
	if !strings.Contains(script, "2% organic") {
		t.Error("script does not contain updated body")
	}
	if !strings.Contains(script, ".name = ") {
		t.Error("script does not assign name")
	}
	if !strings.Contains(script, ".body = ") {
		t.Error("script does not assign body")
	}
	if !strings.Contains(script, ".priority = 5") {
		t.Error("script does not assign priority")
	}
}

func TestScriptUpdateReminder_EscapesInput(t *testing.T) {
	t.Parallel()
	script := scriptUpdateReminder("x-apple-reminder://X", `Say "hello"`, "body\nline2", 0)
	if !strings.Contains(script, `\"hello\"`) {
		t.Error("script does not escape quotes in name")
	}
	if !strings.Contains(script, `body\nline2`) {
		t.Error("script does not escape newlines in body")
	}
}

func TestUpdateReminder_ErrorContextWrapped(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{err: errors.New("timeout")}
	err := UpdateReminder(context.Background(), mock, "x-apple-reminder://ABC", "name", "body", 0)
	if err == nil {
		t.Fatal("expected error")
		return
	}
	if !strings.Contains(err.Error(), "x-apple-reminder://ABC") {
		t.Errorf("error does not contain reminder ID: %v", err)
	}
	if !errors.Is(err, mock.err) {
		t.Errorf("error chain broken: %v does not wrap %v", err, mock.err)
	}
}

func TestScriptDeleteReminder_ContainsID(t *testing.T) {
	t.Parallel()
	script := scriptDeleteReminder("x-apple-reminder://DEF456")
	if !strings.Contains(script, "x-apple-reminder://DEF456") {
		t.Error("script does not contain reminder ID")
	}
	if !strings.Contains(script, "app.delete") {
		t.Error("script does not call app.delete")
	}
}

func TestReadReminders_ScriptPassedToExecutor(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{output: `[]`}
	_, _ = ReadReminders(context.Background(), mock, "Work")
	if !strings.Contains(mock.capturedScript, `"Work"`) {
		t.Error("executor did not receive script with list name")
	}
}

func TestCompleteReminder_ErrorContextWrapped(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{err: errors.New("timeout")}
	err := CompleteReminder(context.Background(), mock, "x-apple-reminder://ABC")
	if err == nil {
		t.Fatal("expected error")
		return
	}
	if !strings.Contains(err.Error(), "x-apple-reminder://ABC") {
		t.Errorf("error does not contain reminder ID: %v", err)
	}
	if !errors.Is(err, mock.err) {
		t.Errorf("error chain broken: %v does not wrap %v", err, mock.err)
	}
}
