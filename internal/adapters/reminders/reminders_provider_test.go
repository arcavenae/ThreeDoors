package reminders

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestRemindersProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewRemindersProvider(&mockExecutor{output: "[]"}, nil)
	if got := p.Name(); got != "reminders" {
		t.Errorf("Name() = %q, want %q", got, "reminders")
	}
}

func TestRemindersProvider_SaveTask_Create(t *testing.T) {
	t.Parallel()

	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			if strings.Contains(script, "app.Reminder") {
				return `{"success":true,"id":"x-apple-reminder://NEW1"}`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, []string{"Work"})

	task := &core.Task{Text: "New task", Effort: core.EffortDeepWork}
	if err := p.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}
	if task.ID != "x-apple-reminder://NEW1" {
		t.Errorf("task.ID = %q, want %q", task.ID, "x-apple-reminder://NEW1")
	}
}

func TestRemindersProvider_SaveTask_Update(t *testing.T) {
	t.Parallel()

	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			if strings.Contains(script, ".name = ") {
				return `{"success":true}`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, []string{"Work"})

	task := &core.Task{ID: "x-apple-reminder://EXIST", Text: "Updated"}
	if err := p.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}
}

func TestRemindersProvider_SaveTask_UpdateNotFound_FallsBackToCreate(t *testing.T) {
	t.Parallel()

	callCount := 0
	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			callCount++
			if strings.Contains(script, ".name = ") {
				return `{"success":false,"error":"reminder not found"}`, nil
			}
			if strings.Contains(script, "app.Reminder") {
				return `{"success":true,"id":"x-apple-reminder://CREATED"}`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, []string{"Work"})

	task := &core.Task{ID: "nonexistent-uuid", Text: "New task"}
	if err := p.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}
	if task.ID != "x-apple-reminder://CREATED" {
		t.Errorf("task.ID = %q, want x-apple-reminder://CREATED", task.ID)
	}
}

func TestRemindersProvider_DeleteTask(t *testing.T) {
	t.Parallel()

	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			if strings.Contains(script, "app.delete") {
				return `{"success":true}`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, []string{"Work"})

	if err := p.DeleteTask("x-apple-reminder://ABC"); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}
}

func TestRemindersProvider_MarkComplete(t *testing.T) {
	t.Parallel()

	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			if strings.Contains(script, "completed = true") {
				return `{"success":true}`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, []string{"Work"})

	if err := p.MarkComplete("x-apple-reminder://ABC"); err != nil {
		t.Fatalf("MarkComplete() error: %v", err)
	}
}

func TestRemindersProvider_DefaultList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		lists       []string
		wantDefault string
	}{
		{"no lists defaults to Reminders", nil, "Reminders"},
		{"empty slice defaults to Reminders", []string{}, "Reminders"},
		{"uses first configured list", []string{"Work", "Personal"}, "Work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewRemindersProvider(&mockExecutor{output: "[]"}, tt.lists)
			if p.defaultList != tt.wantDefault {
				t.Errorf("defaultList = %q, want %q", p.defaultList, tt.wantDefault)
			}
		})
	}
}

func TestRemindersProvider_Watch(t *testing.T) {
	t.Parallel()
	p := NewRemindersProvider(&mockExecutor{output: "[]"}, nil)
	if ch := p.Watch(); ch != nil {
		t.Error("Watch() should return nil")
	}
}

func TestRemindersProvider_LoadTasks_SpecificLists(t *testing.T) {
	t.Parallel()

	mock := &mockExecutor{
		output: `[{"id":"x-apple-reminder://ABC","name":"Buy milk","body":"2% organic","dueDate":"2026-03-10T10:00:00.000Z","priority":1,"completed":false,"flagged":false,"creationDate":"2026-01-01T08:00:00.000Z","modificationDate":"2026-01-05T12:00:00.000Z"}]`,
	}
	p := NewRemindersProvider(mock, []string{"Shopping"})

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}

	task := tasks[0]
	if task.ID != "x-apple-reminder://ABC" {
		t.Errorf("ID = %q, want %q", task.ID, "x-apple-reminder://ABC")
	}
	if task.Text != "Buy milk" {
		t.Errorf("Text = %q, want %q", task.Text, "Buy milk")
	}
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}
	if task.Effort != core.EffortDeepWork {
		t.Errorf("Effort = %q, want %q (priority 1)", task.Effort, core.EffortDeepWork)
	}
	if task.Context != "due:2026-03-10" {
		t.Errorf("Context = %q, want %q", task.Context, "due:2026-03-10")
	}
	if len(task.Notes) != 1 || task.Notes[0].Text != "2% organic" {
		t.Errorf("Notes = %v, want single note with %q", task.Notes, "2% organic")
	}
	if task.SourceProvider != "reminders:Shopping" {
		t.Errorf("SourceProvider = %q, want %q", task.SourceProvider, "reminders:Shopping")
	}
	if len(task.SourceRefs) != 1 || task.SourceRefs[0].Provider != "reminders:Shopping" {
		t.Errorf("SourceRefs = %v, want reminders:Shopping", task.SourceRefs)
	}
}

func TestRemindersProvider_LoadTasks_AllLists(t *testing.T) {
	t.Parallel()

	callCount := 0
	mock := &scriptDispatcher{
		dispatch: func(script string) (string, error) {
			callCount++
			if callCount == 1 {
				return `["Work","Personal"]`, nil
			}
			return `[]`, nil
		},
	}
	p := NewRemindersProvider(mock, nil)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	// 1 call for ReadLists + 2 calls for ReadReminders (Work, Personal)
	if callCount != 3 {
		t.Errorf("executor called %d times, want 3", callCount)
	}
}

func TestRemindersProvider_LoadTasks_ExecutorError(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{err: errors.New("access denied")}
	p := NewRemindersProvider(mock, nil)

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("expected error from LoadTasks()")
	}
}

func TestRemindersProvider_HealthCheck_OK(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{output: `["Reminders"]`}
	p := NewRemindersProvider(mock, nil)

	result := p.HealthCheck()
	if result.Overall != core.HealthOK {
		t.Errorf("Overall = %q, want %q", result.Overall, core.HealthOK)
	}
	if len(result.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(result.Items))
	}
	if result.Items[0].Name != "Apple Reminders" {
		t.Errorf("item name = %q, want %q", result.Items[0].Name, "Apple Reminders")
	}
}

func TestRemindersProvider_HealthCheck_TCCDenied(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errMsg string
	}{
		{"not allowed", "not allowed assistant access"},
		{"denied", "access denied"},
		{"error code 1002", "error 1002: Reminders"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{err: errors.New(tt.errMsg)}
			p := NewRemindersProvider(mock, nil)

			result := p.HealthCheck()
			if result.Overall != core.HealthFail {
				t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
			}
			item := result.Items[0]
			if item.Suggestion == "" {
				t.Error("expected non-empty Suggestion for TCC denial")
			}
			if item.Message != "Reminders access denied by macOS privacy settings" {
				t.Errorf("Message = %q, want TCC message", item.Message)
			}
		})
	}
}

func TestRemindersProvider_HealthCheck_GenericError(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{err: errors.New("osascript: command not found")}
	p := NewRemindersProvider(mock, nil)

	result := p.HealthCheck()
	if result.Overall != core.HealthFail {
		t.Errorf("Overall = %q, want %q", result.Overall, core.HealthFail)
	}
}

func TestMapPriorityToEffort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority int
		want     core.TaskEffort
	}{
		{"priority 0 (none)", 0, ""},
		{"priority 1 (highest)", 1, core.EffortDeepWork},
		{"priority 2", 2, core.EffortDeepWork},
		{"priority 3", 3, core.EffortDeepWork},
		{"priority 4", 4, core.EffortDeepWork},
		{"priority 5 (medium)", 5, core.EffortMedium},
		{"priority 6", 6, core.EffortQuickWin},
		{"priority 7", 7, core.EffortQuickWin},
		{"priority 8", 8, core.EffortQuickWin},
		{"priority 9 (lowest)", 9, core.EffortQuickWin},
		{"negative priority", -1, ""},
		{"out of range", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapPriorityToEffort(tt.priority)
			if got != tt.want {
				t.Errorf("mapPriorityToEffort(%d) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestMapReminderToTask_FieldMapping(t *testing.T) {
	t.Parallel()

	dueDate := "2026-03-15T14:00:00.000Z"

	tests := []struct {
		name     string
		reminder ReminderJSON
		list     string
		check    func(t *testing.T, task *core.Task)
	}{
		{
			name: "complete mapping with all fields",
			reminder: ReminderJSON{
				ID:               "x-apple-reminder://FULL",
				Name:             "Full task",
				Body:             "Details here",
				DueDate:          &dueDate,
				Priority:         5,
				Completed:        false,
				CreationDate:     "2026-01-01T00:00:00.000Z",
				ModificationDate: "2026-02-01T00:00:00.000Z",
			},
			list: "Work",
			check: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.ID != "x-apple-reminder://FULL" {
					t.Errorf("ID = %q", task.ID)
				}
				if task.Text != "Full task" {
					t.Errorf("Text = %q", task.Text)
				}
				if task.Context != "due:2026-03-15" {
					t.Errorf("Context = %q", task.Context)
				}
				if task.Status != core.StatusTodo {
					t.Errorf("Status = %q", task.Status)
				}
				if task.Effort != core.EffortMedium {
					t.Errorf("Effort = %q", task.Effort)
				}
				if len(task.Notes) != 1 || task.Notes[0].Text != "Details here" {
					t.Errorf("Notes = %v", task.Notes)
				}
				if task.CompletedAt != nil {
					t.Error("CompletedAt should be nil for incomplete task")
				}
				if task.SourceProvider != "reminders:Work" {
					t.Errorf("SourceProvider = %q", task.SourceProvider)
				}
			},
		},
		{
			name: "completed reminder maps to complete status",
			reminder: ReminderJSON{
				ID:               "x-apple-reminder://DONE",
				Name:             "Done task",
				Completed:        true,
				CreationDate:     "2026-01-01T00:00:00.000Z",
				ModificationDate: "2026-01-02T00:00:00.000Z",
			},
			list: "Personal",
			check: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Status != core.StatusComplete {
					t.Errorf("Status = %q, want complete", task.Status)
				}
				if task.CompletedAt == nil {
					t.Error("CompletedAt should be set for completed task")
				}
			},
		},
		{
			name: "empty body produces no notes",
			reminder: ReminderJSON{
				ID:               "x-apple-reminder://NOBODY",
				Name:             "No body",
				Body:             "",
				CreationDate:     "2026-01-01T00:00:00.000Z",
				ModificationDate: "2026-01-01T00:00:00.000Z",
			},
			list: "Work",
			check: func(t *testing.T, task *core.Task) {
				t.Helper()
				if len(task.Notes) != 0 {
					t.Errorf("Notes should be empty, got %d", len(task.Notes))
				}
			},
		},
		{
			name: "nil due date produces empty context",
			reminder: ReminderJSON{
				ID:               "x-apple-reminder://NODUE",
				Name:             "No due date",
				DueDate:          nil,
				CreationDate:     "2026-01-01T00:00:00.000Z",
				ModificationDate: "2026-01-01T00:00:00.000Z",
			},
			list: "Work",
			check: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Context != "" {
					t.Errorf("Context = %q, want empty", task.Context)
				}
			},
		},
		{
			name: "priority 0 maps to empty effort",
			reminder: ReminderJSON{
				ID:               "x-apple-reminder://NOPRI",
				Name:             "No priority",
				Priority:         0,
				CreationDate:     "2026-01-01T00:00:00.000Z",
				ModificationDate: "2026-01-01T00:00:00.000Z",
			},
			list: "Work",
			check: func(t *testing.T, task *core.Task) {
				t.Helper()
				if task.Effort != "" {
					t.Errorf("Effort = %q, want empty", task.Effort)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := mapReminderToTask(tt.reminder, tt.list)
			tt.check(t, task)
		})
	}
}

func TestParseISOTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantUTC time.Time
	}{
		{
			name:    "millisecond format",
			input:   "2026-01-15T10:00:00.000Z",
			wantUTC: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:    "RFC3339 format",
			input:   "2026-01-15T10:00:00Z",
			wantUTC: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid returns zero",
			input:   "not-a-date",
			wantUTC: time.Time{},
		},
		{
			name:    "empty returns zero",
			input:   "",
			wantUTC: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseISOTime(tt.input)
			if !got.Equal(tt.wantUTC) {
				t.Errorf("parseISOTime(%q) = %v, want %v", tt.input, got, tt.wantUTC)
			}
		})
	}
}

func TestMapEffortToPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		effort core.TaskEffort
		want   int
	}{
		{"deep-work maps to 1", core.EffortDeepWork, 1},
		{"medium maps to 5", core.EffortMedium, 5},
		{"quick-win maps to 9", core.EffortQuickWin, 9},
		{"empty maps to 0", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapEffortToPriority(tt.effort)
			if got != tt.want {
				t.Errorf("mapEffortToPriority(%q) = %d, want %d", tt.effort, got, tt.want)
			}
		})
	}
}

func TestMapTaskToReminderFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		task         *core.Task
		wantName     string
		wantBody     string
		wantPriority int
	}{
		{
			name:         "full mapping",
			task:         &core.Task{Text: "Buy milk", Notes: []core.TaskNote{{Text: "2% organic"}}, Effort: core.EffortDeepWork},
			wantName:     "Buy milk",
			wantBody:     "2% organic",
			wantPriority: 1,
		},
		{
			name:         "no notes produces empty body",
			task:         &core.Task{Text: "Simple task", Effort: core.EffortMedium},
			wantName:     "Simple task",
			wantBody:     "",
			wantPriority: 5,
		},
		{
			name:         "no effort produces priority 0",
			task:         &core.Task{Text: "No effort"},
			wantName:     "No effort",
			wantBody:     "",
			wantPriority: 0,
		},
		{
			name:         "quick-win maps to priority 9",
			task:         &core.Task{Text: "Quick", Effort: core.EffortQuickWin},
			wantName:     "Quick",
			wantBody:     "",
			wantPriority: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			name, body, priority := mapTaskToReminderFields(tt.task)
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if priority != tt.wantPriority {
				t.Errorf("priority = %d, want %d", priority, tt.wantPriority)
			}
		})
	}
}

func TestPriorityEffortRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		effort   core.TaskEffort
		priority int
	}{
		{core.EffortDeepWork, 1},
		{core.EffortMedium, 5},
		{core.EffortQuickWin, 9},
	}

	for _, tt := range tests {
		t.Run(string(tt.effort), func(t *testing.T) {
			t.Parallel()
			gotPriority := mapEffortToPriority(tt.effort)
			if gotPriority != tt.priority {
				t.Errorf("effort→priority: %q → %d, want %d", tt.effort, gotPriority, tt.priority)
			}
			gotEffort := mapPriorityToEffort(tt.priority)
			if gotEffort != tt.effort {
				t.Errorf("priority→effort: %d → %q, want %q", tt.priority, gotEffort, tt.effort)
			}
		})
	}
}

// scriptDispatcher allows custom per-call dispatch logic in tests.
type scriptDispatcher struct {
	dispatch func(script string) (string, error)
}

func (d *scriptDispatcher) Execute(_ context.Context, script string) (string, error) {
	return d.dispatch(script)
}
