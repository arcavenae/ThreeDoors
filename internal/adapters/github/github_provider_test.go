package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// mockIssueLister implements IssueLister for testing.
type mockIssueLister struct {
	issues     map[string][]*GitHubIssue // keyed by "owner/repo"
	user       string
	listErr    error
	userErr    error
	closeErr   error
	closeCalls []closeCall
}

type closeCall struct {
	owner  string
	repo   string
	number int
}

func (m *mockIssueLister) ListIssues(_ context.Context, owner, repo, _ string) ([]*GitHubIssue, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	key := owner + "/" + repo
	return m.issues[key], nil
}

func (m *mockIssueLister) GetAuthenticatedUser(_ context.Context) (string, error) {
	if m.userErr != nil {
		return "", m.userErr
	}
	return m.user, nil
}

func (m *mockIssueLister) CloseIssue(_ context.Context, owner, repo string, issueNumber int) error {
	m.closeCalls = append(m.closeCalls, closeCall{owner, repo, issueNumber})
	return m.closeErr
}

func newTestConfig(repos ...string) *GitHubConfig {
	return &GitHubConfig{
		Token:           "test-token",
		Repos:           repos,
		Assignee:        "testuser",
		PollInterval:    5 * time.Minute,
		InProgressLabel: "in-progress",
	}
}

func newTestProvider(lister *mockIssueLister, repos ...string) *GitHubProvider {
	return NewGitHubProvider(lister, newTestConfig(repos...))
}

func makeGitHubIssue(number int, title, state, repo string, labels []string) *GitHubIssue {
	now := time.Now().UTC()
	return &GitHubIssue{
		Number:    number,
		Title:     title,
		Body:      "Issue body for " + title,
		State:     state,
		Labels:    labels,
		Assignee:  "testuser",
		CreatedAt: now,
		UpdatedAt: now,
		HTMLURL:   "https://github.com/" + repo + "/issues/" + fmt.Sprintf("%d", number),
		Repo:      repo,
	}
}

func findTaskByID(tasks []*core.Task, id string) *core.Task {
	for _, t := range tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// --- Contract Tests ---

func TestGitHubProviderContract(t *testing.T) {
	t.Parallel()
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		lister := &mockIssueLister{user: "testuser"}
		return newTestProvider(lister, "owner/repo")
	}
	adapters.RunContractTests(t, factory)
}

// --- Name ---

func TestName(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")
	if p.Name() != "github" {
		t.Errorf("Name() = %q, want %q", p.Name(), "github")
	}
}

// --- Read-only methods (AC7: Save/Delete remain ErrReadOnly) ---

func TestReadOnlyMethods(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"SaveTask", func() error { return p.SaveTask(core.NewTask("test")) }},
		{"SaveTasks", func() error { return p.SaveTasks([]*core.Task{core.NewTask("test")}) }},
		{"DeleteTask", func() error { return p.DeleteTask("test-id") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if !errors.Is(err, core.ErrReadOnly) {
				t.Errorf("%s() error = %v, want ErrReadOnly", tt.name, err)
			}
		})
	}
}

// --- MarkComplete (AC1, AC2) ---

func TestMarkComplete_Success(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{user: "testuser"}
	p := newTestProvider(lister, "owner/repo")

	err := p.MarkComplete("github:owner/repo#42")
	if err != nil {
		t.Fatalf("MarkComplete() error: %v", err)
	}

	if len(lister.closeCalls) != 1 {
		t.Fatalf("expected 1 CloseIssue call, got %d", len(lister.closeCalls))
	}
	call := lister.closeCalls[0]
	if call.owner != "owner" || call.repo != "repo" || call.number != 42 {
		t.Errorf("CloseIssue called with (%q, %q, %d), want (owner, repo, 42)", call.owner, call.repo, call.number)
	}
}

func TestMarkComplete_InvalidTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		taskID string
	}{
		{"missing prefix", "owner/repo#42"},
		{"missing hash", "github:owner/repo42"},
		{"missing owner", "github:/repo#42"},
		{"missing repo", "github:owner/#42"},
		{"bad number", "github:owner/repo#abc"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")
			err := p.MarkComplete(tt.taskID)
			if err == nil {
				t.Errorf("MarkComplete(%q) expected error, got nil", tt.taskID)
			}
		})
	}
}

func TestMarkComplete_APIError(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user:     "testuser",
		closeErr: errors.New("not found"),
	}
	p := newTestProvider(lister, "owner/repo")

	err := p.MarkComplete("github:owner/repo#42")
	if err == nil {
		t.Fatal("MarkComplete() expected error, got nil")
	}
}

// --- Rate limit handling (AC5) ---

func TestMarkComplete_RateLimitRetry(t *testing.T) {
	t.Parallel()

	callCount := 0
	lister := &rateLimitIssueLister{
		inner:        &mockIssueLister{user: "testuser"},
		failCount:    2, // fail twice with rate limit, then succeed
		callCountPtr: &callCount,
	}

	p := NewGitHubProvider(lister, newTestConfig("owner/repo"))
	p.sleepFn = func(_ time.Duration) {} // no-op sleep for tests

	err := p.MarkComplete("github:owner/repo#42")
	if err != nil {
		t.Fatalf("MarkComplete() error: %v (expected success after retry)", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 CloseIssue calls (2 rate-limited + 1 success), got %d", callCount)
	}
}

func TestMarkComplete_RateLimitExhausted(t *testing.T) {
	t.Parallel()

	callCount := 0
	lister := &rateLimitIssueLister{
		inner:        &mockIssueLister{user: "testuser"},
		failCount:    10, // always fail
		callCountPtr: &callCount,
	}

	p := NewGitHubProvider(lister, newTestConfig("owner/repo"))
	p.sleepFn = func(_ time.Duration) {}

	err := p.MarkComplete("github:owner/repo#42")
	if err == nil {
		t.Fatal("MarkComplete() expected error after exhausted retries, got nil")
	}

	if callCount != maxRateLimitRetries {
		t.Errorf("expected %d CloseIssue calls, got %d", maxRateLimitRetries, callCount)
	}
}

// --- Circuit breaker (AC6) ---

func TestMarkComplete_CircuitBreakerTrips(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user:     "testuser",
		closeErr: errors.New("server error"),
	}

	p := newTestProvider(lister, "owner/repo")

	// Trip the circuit breaker with 5 consecutive failures
	for range 5 {
		_ = p.MarkComplete("github:owner/repo#1")
	}

	// Next call should be fast-failed by circuit breaker
	err := p.MarkComplete("github:owner/repo#1")
	if !errors.Is(err, core.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen after 5 failures, got: %v", err)
	}
}

// --- WAL integration (AC3, AC4) ---

func TestMarkComplete_WALFallback(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lister := &mockIssueLister{
		user:     "testuser",
		closeErr: errors.New("network error"),
	}

	inner := newTestProvider(lister, "owner/repo")
	walProvider := core.NewWALProvider(inner, tmpDir)

	// MarkComplete should queue in WAL when API fails
	err := walProvider.MarkComplete("github:owner/repo#42")
	if err != nil {
		t.Fatalf("WAL MarkComplete() error: %v (should queue, not fail)", err)
	}

	if walProvider.PendingCount() != 1 {
		t.Errorf("PendingCount() = %d, want 1", walProvider.PendingCount())
	}
}

func TestMarkComplete_WALReplay(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lister := &mockIssueLister{user: "testuser"}

	// Start with failing API to queue in WAL
	lister.closeErr = errors.New("network error")
	inner := newTestProvider(lister, "owner/repo")
	walProvider := core.NewWALProvider(inner, tmpDir)

	_ = walProvider.MarkComplete("github:owner/repo#42")
	if walProvider.PendingCount() != 1 {
		t.Fatalf("PendingCount() = %d, want 1", walProvider.PendingCount())
	}

	// Fix the API, reset call tracking, and replay
	lister.closeErr = nil
	lister.closeCalls = nil
	errs := walProvider.ReplayPending()
	if len(errs) > 0 {
		t.Fatalf("ReplayPending() errors: %v", errs)
	}

	if walProvider.PendingCount() != 0 {
		t.Errorf("PendingCount() after replay = %d, want 0", walProvider.PendingCount())
	}

	if len(lister.closeCalls) != 1 {
		t.Errorf("expected 1 CloseIssue call after replay, got %d", len(lister.closeCalls))
	}
}

// --- parseTaskID ---

func TestParseTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskID     string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{"valid", "github:owner/repo#42", "owner", "repo", 42, false},
		{"org with dash", "github:my-org/my-repo#1", "my-org", "my-repo", 1, false},
		{"missing prefix", "owner/repo#42", "", "", 0, true},
		{"missing hash", "github:owner/repo", "", "", 0, true},
		{"bad number", "github:owner/repo#abc", "", "", 0, true},
		{"empty owner", "github:/repo#42", "", "", 0, true},
		{"empty repo", "github:owner/#42", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, number, err := parseTaskID(tt.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTaskID(%q) error = %v, wantErr %v", tt.taskID, err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
			if number != tt.wantNumber {
				t.Errorf("number = %d, want %d", number, tt.wantNumber)
			}
		})
	}
}

// --- Helpers ---

// rateLimitIssueLister wraps a mockIssueLister and returns rate limit errors
// for the first N calls to CloseIssue, then delegates to the inner mock.
type rateLimitIssueLister struct {
	inner        *mockIssueLister
	failCount    int
	callCountPtr *int
}

func (r *rateLimitIssueLister) ListIssues(ctx context.Context, owner, repo, assignee string) ([]*GitHubIssue, error) {
	return r.inner.ListIssues(ctx, owner, repo, assignee)
}

func (r *rateLimitIssueLister) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return r.inner.GetAuthenticatedUser(ctx)
}

func (r *rateLimitIssueLister) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) error {
	*r.callCountPtr++
	if *r.callCountPtr <= r.failCount {
		return &RateLimitError{RetryAfter: 1 * time.Second}
	}
	return r.inner.CloseIssue(ctx, owner, repo, issueNumber)
}

// --- LoadTasks ---

func TestLoadTasks_SingleRepo(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "First issue", "open", "owner/repo", []string{"bug"}),
				makeGitHubIssue(2, "Second issue", "open", "owner/repo", []string{"feature"}),
			},
		},
	}

	p := newTestProvider(lister, "owner/repo")
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("LoadTasks() returned %d tasks, want 2", len(tasks))
	}

	task1 := findTaskByID(tasks, "github:owner/repo#1")
	if task1 == nil {
		t.Fatal("task github:owner/repo#1 not found")
		return
	}
	if task1.Text != "First issue" {
		t.Errorf("task1.Text = %q, want %q", task1.Text, "First issue")
	}
	if task1.Status != core.StatusTodo {
		t.Errorf("task1.Status = %q, want %q", task1.Status, core.StatusTodo)
	}
	if task1.SourceProvider != "github" {
		t.Errorf("task1.SourceProvider = %q, want %q", task1.SourceProvider, "github")
	}
}

// AC11: Multi-repo aggregation
func TestLoadTasks_MultiRepo(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"org/repo-a": {
				makeGitHubIssue(1, "Issue A", "open", "org/repo-a", nil),
			},
			"org/repo-b": {
				makeGitHubIssue(5, "Issue B", "open", "org/repo-b", nil),
			},
		},
	}

	cfg := newTestConfig("org/repo-a", "org/repo-b")
	p := NewGitHubProvider(lister, cfg)
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("LoadTasks() returned %d tasks, want 2", len(tasks))
	}

	taskA := findTaskByID(tasks, "github:org/repo-a#1")
	if taskA == nil {
		t.Error("task from repo-a not found")
	}
	taskB := findTaskByID(tasks, "github:org/repo-b#5")
	if taskB == nil {
		t.Error("task from repo-b not found")
	}
}

func TestLoadTasks_EmptyResults(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user:   "testuser",
		issues: map[string][]*GitHubIssue{},
	}

	p := newTestProvider(lister, "owner/repo")
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("LoadTasks() returned %d tasks, want 0", len(tasks))
	}
}

func TestLoadTasks_APIError(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user:    "testuser",
		listErr: errors.New("connection refused"),
	}

	p := newTestProvider(lister, "owner/repo")
	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error, got nil")
	}
}

// AC3: Status mapping
func TestMapStatus(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	tests := []struct {
		name       string
		state      string
		labels     []string
		wantStatus core.TaskStatus
	}{
		{"open maps to todo", "open", nil, core.StatusTodo},
		{"closed maps to complete", "closed", nil, core.StatusComplete},
		{"open with in-progress label", "open", []string{"in-progress"}, core.StatusInProgress},
		{"closed with in-progress label still complete", "closed", []string{"in-progress"}, core.StatusComplete},
		{"open with other labels", "open", []string{"bug", "feature"}, core.StatusTodo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &GitHubIssue{State: tt.state, Labels: tt.labels}
			got := p.mapStatus(issue)
			if got != tt.wantStatus {
				t.Errorf("mapStatus() = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

// AC4: Label-to-effort mapping
func TestMapEffort(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	tests := []struct {
		name       string
		labels     []string
		wantEffort core.TaskEffort
	}{
		{"priority:critical -> deep-work", []string{"priority:critical"}, core.EffortDeepWork},
		{"priority:high -> deep-work", []string{"priority:high"}, core.EffortDeepWork},
		{"priority:medium -> medium", []string{"priority:medium"}, core.EffortMedium},
		{"priority:low -> quick-win", []string{"priority:low"}, core.EffortQuickWin},
		{"no matching label -> quick-win default", []string{"bug"}, core.EffortQuickWin},
		{"no labels -> quick-win default", nil, core.EffortQuickWin},
		{"first match wins", []string{"priority:low", "priority:critical"}, core.EffortQuickWin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := p.mapEffort(tt.labels)
			if got != tt.wantEffort {
				t.Errorf("mapEffort(%v) = %q, want %q", tt.labels, got, tt.wantEffort)
			}
		})
	}
}

// AC4: Custom priority labels
func TestMapEffort_CustomLabels(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig("owner/repo")
	cfg.PriorityLabels = map[string]string{
		"urgent": "deep-work",
		"minor":  "quick-win",
	}
	p := NewGitHubProvider(&mockIssueLister{user: "testuser"}, cfg)

	got := p.mapEffort([]string{"urgent"})
	if got != core.EffortDeepWork {
		t.Errorf("mapEffort([urgent]) = %q, want %q", got, core.EffortDeepWork)
	}

	got = p.mapEffort([]string{"minor"})
	if got != core.EffortQuickWin {
		t.Errorf("mapEffort([minor]) = %q, want %q", got, core.EffortQuickWin)
	}
}

// AC2: Field mapping
func TestMapIssueToTask(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	issue := makeGitHubIssue(42, "Fix the bug", "open", "owner/repo", []string{"bug", "priority:high"})
	task := p.mapIssueToTask(issue)

	if task.ID != "github:owner/repo#42" {
		t.Errorf("task.ID = %q, want %q", task.ID, "github:owner/repo#42")
	}
	if task.Text != "Fix the bug" {
		t.Errorf("task.Text = %q, want %q", task.Text, "Fix the bug")
	}
	if task.Context != "Issue body for Fix the bug" {
		t.Errorf("task.Context = %q, want body text", task.Context)
	}
	if task.Status != core.StatusTodo {
		t.Errorf("task.Status = %q, want %q", task.Status, core.StatusTodo)
	}
	if task.Effort != core.EffortDeepWork {
		t.Errorf("task.Effort = %q, want %q", task.Effort, core.EffortDeepWork)
	}
	if task.SourceProvider != "github" {
		t.Errorf("task.SourceProvider = %q, want %q", task.SourceProvider, "github")
	}
	if len(task.SourceRefs) != 1 || task.SourceRefs[0].NativeID != "github:owner/repo#42" {
		t.Errorf("task.SourceRefs = %v, want [{github github:owner/repo#42}]", task.SourceRefs)
	}
	if string(task.Location) != "owner/repo#42" {
		t.Errorf("task.Location = %q, want %q", task.Location, "owner/repo#42")
	}
}

// AC5: Milestone due date mapping
func TestMapIssueToTask_MilestoneDueDate(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	dueDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	issue := makeGitHubIssue(10, "Milestone task", "open", "owner/repo", nil)
	issue.MilestoneDueOn = &dueDate

	task := p.mapIssueToTask(issue)

	wantContext := "Issue body for Milestone task\n\nDue: 2026-04-15"
	if task.Context != wantContext {
		t.Errorf("task.Context = %q, want %q", task.Context, wantContext)
	}
}

func TestMapIssueToTask_NoMilestone(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	issue := makeGitHubIssue(10, "No milestone", "open", "owner/repo", nil)
	task := p.mapIssueToTask(issue)

	if task.Context != "Issue body for No milestone" {
		t.Errorf("task.Context = %q, want body only", task.Context)
	}
}

// AC8: Name and source badge
func TestSourceBadge(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	issue := makeGitHubIssue(1, "Test", "open", "owner/repo", nil)
	task := p.mapIssueToTask(issue)

	if task.SourceProvider != "github" {
		t.Errorf("task.SourceProvider = %q, want %q", task.SourceProvider, "github")
	}
}

// @me assignee resolution
func TestLoadTasks_AtMeResolution(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user: "realuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "My issue", "open", "owner/repo", nil),
			},
		},
	}

	cfg := &GitHubConfig{
		Token:           "test-token",
		Repos:           []string{"owner/repo"},
		Assignee:        "@me",
		PollInterval:    5 * time.Minute,
		InProgressLabel: "in-progress",
	}
	p := NewGitHubProvider(lister, cfg)
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}
}

func TestLoadTasks_AtMeResolutionError(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		userErr: errors.New("auth failed"),
	}

	cfg := &GitHubConfig{
		Token:           "test-token",
		Repos:           []string{"owner/repo"},
		Assignee:        "@me",
		PollInterval:    5 * time.Minute,
		InProgressLabel: "in-progress",
	}
	p := NewGitHubProvider(lister, cfg)
	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error for failed @me resolution, got nil")
	}
}

// --- HealthCheck (AC7) ---

func TestHealthCheck_Healthy(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")
	result := p.HealthCheck()

	if result.Overall != core.HealthOK {
		t.Errorf("HealthCheck().Overall = %q, want %q", result.Overall, core.HealthOK)
	}
	if len(result.Items) == 0 {
		t.Error("HealthCheck().Items is empty, want at least 1 item")
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{userErr: errors.New("unauthorized")}
	p := newTestProvider(lister, "owner/repo")
	result := p.HealthCheck()

	if result.Overall != core.HealthFail {
		t.Errorf("HealthCheck().Overall = %q, want %q", result.Overall, core.HealthFail)
	}
}

// --- Watch (AC13) ---

func TestWatch_ReturnsChannel(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	ch := p.Watch()
	if ch == nil {
		t.Error("Watch() returned nil, want channel")
	}

	// Signal stop to clean up goroutine
	close(p.stopCh)
}

// --- Cache (AC9) ---

func TestLoadTasks_WritesCacheOnSuccess(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "Cached issue", "open", "owner/repo", nil),
			},
		},
	}

	p := newTestProvider(lister, "owner/repo")
	p.SetCachePath(tmpDir)

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}

	cachePath := filepath.Join(tmpDir, cacheFileName)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}
}

func TestLoadTasks_FallsBackToCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// First: successful load to populate cache
	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "Cached issue", "open", "owner/repo", nil),
			},
		},
	}

	p := newTestProvider(lister, "owner/repo")
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("initial LoadTasks() error: %v", err)
	}

	// Second: API failure should use disk cache
	failLister := &mockIssueLister{
		user:    "testuser",
		listErr: errors.New("connection refused"),
	}
	p2 := newTestProvider(failLister, "owner/repo")
	p2.SetCachePath(tmpDir)

	tasks, err := p2.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() with cache fallback error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d cached tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "github:owner/repo#1" {
		t.Errorf("cached task ID = %q, want %q", tasks[0].ID, "github:owner/repo#1")
	}
}

func TestLoadTasks_FailsWithoutCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lister := &mockIssueLister{
		user:    "testuser",
		listErr: errors.New("connection refused"),
	}
	p := newTestProvider(lister, "owner/repo")
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error when API fails and no cache, got nil")
	}
}

func TestLoadTasks_UsesInMemoryCacheWithinTTL(t *testing.T) {
	t.Parallel()

	callCount := 0
	lister := &countingIssueLister{
		inner: &mockIssueLister{
			user: "testuser",
			issues: map[string][]*GitHubIssue{
				"owner/repo": {
					makeGitHubIssue(1, "Cached", "open", "owner/repo", nil),
				},
			},
		},
		listCount: &callCount,
	}

	p := NewGitHubProvider(lister, newTestConfig("owner/repo"))

	// First call should hit API
	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("first LoadTasks() error: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 API call, got %d", callCount)
	}

	// Second call within TTL should use in-memory cache
	_, err = p.LoadTasks()
	if err != nil {
		t.Fatalf("second LoadTasks() error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}

func TestCacheAtomicWrite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "Task", "open", "owner/repo", nil),
			},
		},
	}

	p := newTestProvider(lister, "owner/repo")
	p.SetCachePath(tmpDir)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	tmpPath := filepath.Join(tmpDir, cacheFileName+".tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful write")
	}
}

// --- Factory (AC10) ---

func TestFactory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]string
		wantErr  bool
	}{
		{
			name: "valid config",
			settings: map[string]string{
				"token": "ghp_test123",
				"repos": "owner/repo",
			},
			wantErr: false,
		},
		{
			name:     "missing repos",
			settings: map[string]string{"token": "ghp_test123"},
			wantErr:  true,
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := &core.ProviderConfig{
				Providers: []core.ProviderEntry{
					{Name: "github", Settings: tt.settings},
				},
			}
			provider, err := Factory(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Factory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && provider == nil {
				t.Error("Factory() returned nil provider")
			}
		})
	}
}

func TestFactory_NoGitHubSettings(t *testing.T) {
	t.Parallel()
	config := &core.ProviderConfig{
		Providers: []core.ProviderEntry{
			{Name: "jira", Settings: map[string]string{}},
		},
	}
	_, err := Factory(config)
	if err == nil {
		t.Fatal("Factory() expected error for missing github settings, got nil")
	}
}

// --- splitOwnerRepo ---

func TestSplitOwnerRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"valid", "owner/repo", "owner", "repo", false},
		{"missing owner", "/repo", "", "", true},
		{"missing repo", "owner/", "", "", true},
		{"no slash", "ownerrepo", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, err := splitOwnerRepo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitOwnerRepo(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

// AC3: Custom in-progress label name via config
func TestMapStatus_CustomInProgressLabel(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig("owner/repo")
	cfg.InProgressLabel = "wip"
	p := NewGitHubProvider(&mockIssueLister{user: "testuser"}, cfg)

	tests := []struct {
		name       string
		state      string
		labels     []string
		wantStatus core.TaskStatus
	}{
		{"open with custom wip label", "open", []string{"wip"}, core.StatusInProgress},
		{"open with default in-progress not matched", "open", []string{"in-progress"}, core.StatusTodo},
		{"closed ignores custom label", "closed", []string{"wip"}, core.StatusComplete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &GitHubIssue{State: tt.state, Labels: tt.labels}
			got := p.mapStatus(issue)
			if got != tt.wantStatus {
				t.Errorf("mapStatus() = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

// AC7: Special character handling
func TestMapIssueToTask_SpecialCharacters(t *testing.T) {
	t.Parallel()
	p := newTestProvider(&mockIssueLister{user: "testuser"}, "owner/repo")

	tests := []struct {
		name     string
		title    string
		body     string
		labels   []string
		wantText string
		wantCtx  string
	}{
		{
			name:     "unicode title",
			title:    "Fix 日本語 rendering in übersicht",
			body:     "Details here",
			labels:   nil,
			wantText: "Fix 日本語 rendering in übersicht",
			wantCtx:  "Details here",
		},
		{
			name:     "emoji in title",
			title:    "🐛 Bug: crash on startup",
			body:     "Repro steps",
			labels:   nil,
			wantText: "🐛 Bug: crash on startup",
			wantCtx:  "Repro steps",
		},
		{
			name:     "markdown body with code blocks",
			title:    "Add feature",
			body:     "## Steps\n\n```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n- [x] Done",
			labels:   nil,
			wantText: "Add feature",
			wantCtx:  "## Steps\n\n```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n- [x] Done",
		},
		{
			name:     "labels with special characters",
			title:    "Test",
			body:     "",
			labels:   []string{"priority:high", "type/bug-fix", "area:日本語"},
			wantText: "Test",
			wantCtx:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			now := time.Now().UTC()
			issue := &GitHubIssue{
				Number:    1,
				Title:     tt.title,
				Body:      tt.body,
				State:     "open",
				Labels:    tt.labels,
				Assignee:  "testuser",
				CreatedAt: now,
				UpdatedAt: now,
				HTMLURL:   "https://github.com/owner/repo/issues/1",
				Repo:      "owner/repo",
			}
			task := p.mapIssueToTask(issue)
			if task.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", task.Text, tt.wantText)
			}
			if task.Context != tt.wantCtx {
				t.Errorf("Context = %q, want %q", task.Context, tt.wantCtx)
			}
		})
	}
}

// AC7: Special characters in LoadTasks round-trip
func TestLoadTasks_SpecialCharacters(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"owner/repo": {
				makeGitHubIssue(1, "🔥 Urgent: café résumé", "open", "owner/repo", []string{"priority:high", "type/日本語"}),
			},
		},
	}

	p := newTestProvider(lister, "owner/repo")
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].Text != "🔥 Urgent: café résumé" {
		t.Errorf("task.Text = %q, want unicode text preserved", tasks[0].Text)
	}
}

// AC8: Assignee filtering — only configured user's issues returned
func TestLoadTasks_AssigneeFiltering(t *testing.T) {
	t.Parallel()

	capturedAssignee := ""
	lister := &assigneeCapturingLister{
		inner: &mockIssueLister{
			user: "testuser",
			issues: map[string][]*GitHubIssue{
				"owner/repo": {
					makeGitHubIssue(1, "My issue", "open", "owner/repo", nil),
				},
			},
		},
		capturedAssignee: &capturedAssignee,
	}

	cfg := newTestConfig("owner/repo")
	cfg.Assignee = "specificuser"
	p := NewGitHubProvider(lister, cfg)

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if capturedAssignee != "specificuser" {
		t.Errorf("assignee passed to API = %q, want %q", capturedAssignee, "specificuser")
	}
}

// AC4: Multi-repo unique ID verification
func TestLoadTasks_MultiRepo_UniqueIDs(t *testing.T) {
	t.Parallel()
	lister := &mockIssueLister{
		user: "testuser",
		issues: map[string][]*GitHubIssue{
			"org/repo-a": {
				makeGitHubIssue(1, "Issue A-1", "open", "org/repo-a", nil),
				makeGitHubIssue(2, "Issue A-2", "open", "org/repo-a", nil),
			},
			"org/repo-b": {
				makeGitHubIssue(1, "Issue B-1", "open", "org/repo-b", nil),
			},
		},
	}

	cfg := newTestConfig("org/repo-a", "org/repo-b")
	p := NewGitHubProvider(lister, cfg)
	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("LoadTasks() returned %d tasks, want 3", len(tasks))
	}

	// Verify all IDs are unique
	ids := make(map[string]bool)
	for _, task := range tasks {
		if ids[task.ID] {
			t.Errorf("duplicate task ID: %s", task.ID)
		}
		ids[task.ID] = true
	}

	// Same issue number in different repos should produce different IDs
	if findTaskByID(tasks, "github:org/repo-a#1") == nil {
		t.Error("missing task github:org/repo-a#1")
	}
	if findTaskByID(tasks, "github:org/repo-b#1") == nil {
		t.Error("missing task github:org/repo-b#1")
	}
}

// --- Helpers ---

// assigneeCapturingLister wraps a mockIssueLister and captures the assignee param.
type assigneeCapturingLister struct {
	inner            *mockIssueLister
	capturedAssignee *string
}

func (a *assigneeCapturingLister) ListIssues(ctx context.Context, owner, repo, assignee string) ([]*GitHubIssue, error) {
	*a.capturedAssignee = assignee
	return a.inner.ListIssues(ctx, owner, repo, assignee)
}

func (a *assigneeCapturingLister) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return a.inner.GetAuthenticatedUser(ctx)
}

func (a *assigneeCapturingLister) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) error {
	return a.inner.CloseIssue(ctx, owner, repo, issueNumber)
}

// countingIssueLister wraps a mockIssueLister and counts ListIssues calls.
type countingIssueLister struct {
	inner     *mockIssueLister
	listCount *int
}

func (c *countingIssueLister) ListIssues(ctx context.Context, owner, repo, assignee string) ([]*GitHubIssue, error) {
	*c.listCount++
	return c.inner.ListIssues(ctx, owner, repo, assignee)
}

func (c *countingIssueLister) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return c.inner.GetAuthenticatedUser(ctx)
}

func (c *countingIssueLister) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) error {
	return c.inner.CloseIssue(ctx, owner, repo, issueNumber)
}
