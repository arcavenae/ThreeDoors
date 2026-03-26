package adapters_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// readOnlyProvider is a minimal TaskProvider that returns ErrReadOnly on all
// write operations. Used to exercise the skipIfReadOnly branches in contract tests.
type readOnlyProvider struct {
	tasks []*core.Task
}

func newReadOnlyProvider(tasks []*core.Task) *readOnlyProvider {
	cp := make([]*core.Task, len(tasks))
	copy(cp, tasks)
	return &readOnlyProvider{tasks: cp}
}

func (r *readOnlyProvider) Name() string { return "read-only-test" }

func (r *readOnlyProvider) LoadTasks() ([]*core.Task, error) {
	cp := make([]*core.Task, len(r.tasks))
	copy(cp, r.tasks)
	return cp, nil
}

func (r *readOnlyProvider) SaveTask(_ *core.Task) error    { return core.ErrReadOnly }
func (r *readOnlyProvider) SaveTasks(_ []*core.Task) error { return core.ErrReadOnly }
func (r *readOnlyProvider) DeleteTask(_ string) error      { return core.ErrReadOnly }
func (r *readOnlyProvider) MarkComplete(_ string) error    { return core.ErrReadOnly }
func (r *readOnlyProvider) Watch() <-chan core.ChangeEvent { return nil }
func (r *readOnlyProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{}
}

// TestReadOnlyProviderContract runs the contract suite against a read-only
// provider. Write tests should be skipped via skipIfReadOnly; read tests and
// the Name/Watch/HealthCheck tests should still pass.
func TestReadOnlyProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return newReadOnlyProvider([]*core.Task{
			core.NewTask("Pre-loaded task A"),
			core.NewTask("Pre-loaded task B"),
		})
	}

	adapters.RunContractTests(t, factory)
}

// TestReadOnlyProviderContract_EmptyTasks exercises the contract suite with
// a read-only provider that has no pre-loaded tasks.
func TestReadOnlyProviderContract_EmptyTasks(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return newReadOnlyProvider(nil)
	}

	adapters.RunContractTests(t, factory)
}

// watchingProvider exercises the Watch() path that returns a non-nil channel.
type watchingProvider struct {
	readOnlyProvider
	ch chan core.ChangeEvent
}

func (w *watchingProvider) Watch() <-chan core.ChangeEvent { return w.ch }

// TestWatchingProviderContract verifies the Watch contract test handles a
// provider that returns a real channel (exercises the non-nil Watch branch).
func TestWatchingProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		ch := make(chan core.ChangeEvent, 1)
		// Pre-send an event so the non-blocking read in the test receives it.
		ch <- core.ChangeEvent{Type: core.ChangeUpdated, TaskID: "test-1", Source: "watching-test"}
		return &watchingProvider{
			readOnlyProvider: *newReadOnlyProvider([]*core.Task{
				core.NewTask("Watched task"),
			}),
			ch: ch,
		}
	}

	adapters.RunContractTests(t, factory)
}

// healthyProvider exercises the HealthCheck contract test with non-empty Items.
type healthyProvider struct {
	readOnlyProvider
}

func (h *healthyProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{
		Items: []core.HealthCheckItem{
			{Name: "test-check", Status: core.HealthOK, Message: "all good"},
		},
		Overall: core.HealthOK,
	}
}

// TestHealthyProviderContract verifies the HealthCheck contract test with a
// provider that returns real health items.
func TestHealthyProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return &healthyProvider{
			readOnlyProvider: *newReadOnlyProvider([]*core.Task{
				core.NewTask("Healthy task"),
			}),
		}
	}

	adapters.RunContractTests(t, factory)
}

// inMemoryProvider is a fully functional in-memory provider that properly
// handles all CRUD operations, exercising success paths in contract tests
// that are not covered by read-only or file-based providers.
type inMemoryProvider struct {
	mu    sync.Mutex
	tasks map[string]*core.Task
}

func newInMemoryProvider() *inMemoryProvider {
	return &inMemoryProvider{tasks: make(map[string]*core.Task)}
}

func (m *inMemoryProvider) Name() string { return "in-memory-test" }

func (m *inMemoryProvider) LoadTasks() ([]*core.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*core.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		cp := *t
		result = append(result, &cp)
	}
	return result, nil
}

func (m *inMemoryProvider) SaveTask(task *core.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *task
	m.tasks[task.ID] = &cp
	return nil
}

func (m *inMemoryProvider) SaveTasks(tasks []*core.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = make(map[string]*core.Task, len(tasks))
	for _, t := range tasks {
		cp := *t
		m.tasks[t.ID] = &cp
	}
	return nil
}

func (m *inMemoryProvider) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, taskID)
	return nil
}

func (m *inMemoryProvider) MarkComplete(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[taskID]
	if !ok {
		return errors.New("task not found")
	}
	t.Status = core.StatusComplete
	return nil
}

func (m *inMemoryProvider) Watch() <-chan core.ChangeEvent { return nil }
func (m *inMemoryProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{}
}

// TestInMemoryProviderContract runs the full contract suite against an
// in-memory provider. This exercises all success paths that file-based
// providers cover, plus validates the contract works with non-file backends.
func TestInMemoryProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return newInMemoryProvider()
	}

	adapters.RunContractTests(t, factory)
}
