package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/google/uuid"
)

func testPayload(t *testing.T, text string) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
		return nil
	}
	return data
}

func testPool(t *testing.T, tasks ...*core.Task) *core.TaskPool {
	t.Helper()
	pool := core.NewTaskPool()
	for _, task := range tasks {
		pool.AddTask(task)
	}
	return pool
}

func testTask(id, text string) *core.Task {
	now := time.Now().UTC()
	return &core.Task{
		ID:        id,
		Text:      text,
		Status:    core.StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testStore(t *testing.T, pool *core.TaskPool) *ProposalStore {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "proposals.jsonl")
	store, err := NewProposalStore(path, pool)
	if err != nil {
		t.Fatalf("create test store: %v", err)
		return nil
	}
	return store
}

func TestNewProposal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pType   ProposalType
		taskID  string
		payload json.RawMessage
		wantErr bool
	}{
		{"valid enrich-metadata", ProposalEnrichMetadata, "task-1", json.RawMessage(`{"text":"test"}`), false},
		{"valid add-note", ProposalAddNote, "task-1", json.RawMessage(`{"note":"hello"}`), false},
		{"invalid type", ProposalType("invalid"), "task-1", json.RawMessage(`{}`), true},
		{"empty task_id", ProposalEnrichMetadata, "", json.RawMessage(`{}`), true},
		{"empty payload", ProposalEnrichMetadata, "task-1", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := NewProposal(tt.pType, tt.taskID, time.Now().UTC(), tt.payload, "mcp:test", "test rationale")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProposal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if p.Status != ProposalPending {
					t.Errorf("expected pending status, got %s", p.Status)
				}
				if p.ExpiresAt.Before(p.CreatedAt) {
					t.Error("ExpiresAt should be after CreatedAt")
				}
			}
		})
	}
}

func TestProposalIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   ProposalStatus
		terminal bool
	}{
		{ProposalPending, false},
		{ProposalStale, false},
		{ProposalApproved, true},
		{ProposalRejected, true},
		{ProposalExpired, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			p := &Proposal{Status: tt.status}
			if p.IsTerminal() != tt.terminal {
				t.Errorf("IsTerminal() = %v, want %v", p.IsTerminal(), tt.terminal)
			}
		})
	}
}

func TestProposalStore_CRUD(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "buy groceries"))
	store := testStore(t, pool)

	// Create — use the task's UpdatedAt as BaseVersion for concurrency check.
	task := pool.GetTask(taskID)
	p, err := NewProposal(ProposalAddNote, taskID, task.UpdatedAt, testPayload(t, "remember milk"), "mcp:test", "helpful note")
	if err != nil {
		t.Fatalf("NewProposal: %v", err)
		return
	}
	if err := store.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get
	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
		return
	}
	if got.ID != p.ID {
		t.Errorf("Get ID = %s, want %s", got.ID, p.ID)
	}
	if got.Status != ProposalPending {
		t.Errorf("Get Status = %s, want pending", got.Status)
	}

	// Get not found
	_, err = store.Get("nonexistent")
	if !errors.Is(err, ErrProposalNotFound) {
		t.Errorf("Get nonexistent: expected ErrProposalNotFound, got %v", err)
	}

	// List
	all := store.List(ProposalFilter{})
	if len(all) != 1 {
		t.Errorf("List all: got %d, want 1", len(all))
	}

	// List with filter
	byTask := store.List(ProposalFilter{TaskID: taskID})
	if len(byTask) != 1 {
		t.Errorf("List by task: got %d, want 1", len(byTask))
	}
	byOther := store.List(ProposalFilter{TaskID: "other"})
	if len(byOther) != 0 {
		t.Errorf("List by other task: got %d, want 0", len(byOther))
	}

	// UpdateStatus
	reviewedAt := time.Now().UTC()
	if err := store.UpdateStatus(p.ID, ProposalApproved, reviewedAt); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = store.Get(p.ID)
	if got.Status != ProposalApproved {
		t.Errorf("UpdateStatus: got %s, want approved", got.Status)
	}
	if got.ReviewedAt == nil {
		t.Error("ReviewedAt should be set after UpdateStatus")
	}
}

func TestProposalStore_UpdateStatus_NotFound(t *testing.T) {
	t.Parallel()

	store := testStore(t, testPool(t))
	err := store.UpdateStatus("nonexistent", ProposalApproved, time.Now().UTC())
	if !errors.Is(err, ErrProposalNotFound) {
		t.Errorf("expected ErrProposalNotFound, got %v", err)
	}
}

func TestProposalStore_UpdateStatus_TerminalState(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "test task"))
	store := testStore(t, pool)

	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), testPayload(t, "note"), "mcp:test", "reason")
	_ = store.Create(p)
	_ = store.UpdateStatus(p.ID, ProposalRejected, time.Now().UTC())

	err := store.UpdateStatus(p.ID, ProposalApproved, time.Now().UTC())
	if err == nil {
		t.Error("expected error updating terminal proposal, got nil")
	}
}

func TestProposalStore_OptimisticConcurrency(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	task := testTask(taskID, "original task")
	pool := testPool(t, task)
	store := testStore(t, pool)

	// Create proposal with current baseVersion.
	p, _ := NewProposal(ProposalAddNote, taskID, task.UpdatedAt, testPayload(t, "note"), "mcp:test", "reason")
	_ = store.Create(p)

	// Simulate task modification by updating the task's UpdatedAt.
	task.UpdatedAt = time.Now().UTC().Add(time.Hour)
	pool.UpdateTask(task)

	// Approve should detect stale and mark as stale.
	err := store.UpdateStatus(p.ID, ProposalApproved, time.Now().UTC())
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
		return
	}

	got, _ := store.Get(p.ID)
	if got.Status != ProposalStale {
		t.Errorf("expected stale status, got %s", got.Status)
	}
}

func TestProposalStore_PerTaskCap(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "capped task"))
	store := testStore(t, pool)

	// Fill up to cap.
	for i := range MaxPendingPerTask {
		payload := testPayload(t, "unique note "+uuid.New().String())
		p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
		if err := store.Create(p); err != nil {
			t.Fatalf("Create proposal %d: %v", i, err)
		}
	}

	// Next one should fail.
	payload := testPayload(t, "one too many "+uuid.New().String())
	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	err := store.Create(p)
	if !errors.Is(err, ErrPerTaskCapReached) {
		t.Errorf("expected ErrPerTaskCapReached, got %v", err)
	}
}

func TestProposalStore_DedupPendingProposals(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "unique task text that will not match"))
	store := testStore(t, pool)

	payload := testPayload(t, "add milk to the shopping list")
	p1, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	if err := store.Create(p1); err != nil {
		t.Fatalf("Create first: %v", err)
	}

	// Very similar payload should be rejected.
	payload2 := testPayload(t, "add milk to the shopping list")
	p2, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload2, "mcp:test", "reason")
	err := store.Create(p2)
	if !errors.Is(err, ErrDuplicateProposal) {
		t.Errorf("expected ErrDuplicateProposal, got %v", err)
	}
}

func TestProposalStore_DedupAgainstExistingTasks(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "buy groceries from the store"))
	store := testStore(t, pool)

	// Proposal with text very similar to existing task.
	payload := testPayload(t, "buy groceries from the store")
	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	err := store.Create(p)
	if !errors.Is(err, ErrDuplicateProposal) {
		t.Errorf("expected ErrDuplicateProposal, got %v", err)
	}
}

func TestProposalStore_DedupRecentlyRejected(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "completely different task name"))
	store := testStore(t, pool)

	// Create and reject a proposal.
	payload := testPayload(t, "rejected enrichment idea")
	p1, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	_ = store.Create(p1)
	_ = store.UpdateStatus(p1.ID, ProposalRejected, time.Now().UTC())

	// Try to create a similar one — should be rejected due to recent rejection.
	payload2 := testPayload(t, "rejected enrichment idea")
	p2, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload2, "mcp:test", "reason")
	err := store.Create(p2)
	if !errors.Is(err, ErrDuplicateProposal) {
		t.Errorf("expected ErrDuplicateProposal for recently rejected, got %v", err)
	}
}

func TestProposalStore_Expiration(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "expirable task"))
	store := testStore(t, pool)

	// Create a proposal and manually set it to expire in the past.
	payload := testPayload(t, "will expire "+uuid.New().String())
	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	p.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	_ = store.Create(p)

	// Run expiration sweep.
	expired := store.ExpireSweep()
	if expired != 1 {
		t.Errorf("ExpireSweep: got %d expired, want 1", expired)
	}

	got, _ := store.Get(p.ID)
	if got.Status != ProposalExpired {
		t.Errorf("expected expired status, got %s", got.Status)
	}
}

func TestProposalStore_Persistence(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	dir := t.TempDir()
	path := filepath.Join(dir, "proposals.jsonl")
	pool := testPool(t, testTask(taskID, "persistent task"))

	// Create store and add a proposal.
	store1, err := NewProposalStore(path, pool)
	if err != nil {
		t.Fatalf("create store: %v", err)
		return
	}
	payload := testPayload(t, "persistent note "+uuid.New().String())
	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	_ = store1.Create(p)

	// Reload from disk.
	store2, err := NewProposalStore(path, pool)
	if err != nil {
		t.Fatalf("reload store: %v", err)
		return
	}

	got, err := store2.Get(p.ID)
	if err != nil {
		t.Fatalf("Get from reloaded store: %v", err)
		return
	}
	if got.Status != ProposalPending {
		t.Errorf("reloaded status = %s, want pending", got.Status)
	}
}

func TestProposalStore_ListFilters(t *testing.T) {
	t.Parallel()

	taskA := uuid.New().String()
	taskB := uuid.New().String()
	pool := testPool(t, testTask(taskA, "task a"), testTask(taskB, "task b"))
	store := testStore(t, pool)

	payloadA := testPayload(t, "note for A "+uuid.New().String())
	pA, _ := NewProposal(ProposalAddNote, taskA, time.Now().UTC(), payloadA, "mcp:claude", "reason")
	_ = store.Create(pA)

	payloadB := testPayload(t, "note for B "+uuid.New().String())
	pB, _ := NewProposal(ProposalAddContext, taskB, time.Now().UTC(), payloadB, "mcp:cursor", "reason")
	_ = store.Create(pB)

	tests := []struct {
		name   string
		filter ProposalFilter
		want   int
	}{
		{"all", ProposalFilter{}, 2},
		{"by task A", ProposalFilter{TaskID: taskA}, 1},
		{"by task B", ProposalFilter{TaskID: taskB}, 1},
		{"by status pending", ProposalFilter{Status: ProposalPending}, 2},
		{"by status approved", ProposalFilter{Status: ProposalApproved}, 0},
		{"by type add-note", ProposalFilter{Type: ProposalAddNote}, 1},
		{"by type add-context", ProposalFilter{Type: ProposalAddContext}, 1},
		{"by source claude", ProposalFilter{Source: "mcp:claude"}, 1},
		{"by source cursor", ProposalFilter{Source: "mcp:cursor"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := store.List(tt.filter)
			if len(got) != tt.want {
				t.Errorf("List(%+v) = %d, want %d", tt.filter, len(got), tt.want)
			}
		})
	}
}

func TestIntakeChannel_LLM(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "intake task"))
	store := testStore(t, pool)

	ch := NewLLMIntakeChannel(store, "mcp:claude-desktop")

	if ch.Name() != "mcp:claude-desktop" {
		t.Errorf("Name() = %s, want mcp:claude-desktop", ch.Name())
	}

	health := ch.HealthCheck()
	if health.Status != HealthOK {
		t.Errorf("HealthCheck() = %s, want ok", health.Status)
	}

	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), testPayload(t, "intake note "+uuid.New().String()), "", "reason")
	if err := ch.Suggest(context.Background(), p); err != nil {
		t.Fatalf("Suggest: %v", err)
	}

	// Verify source was set.
	got, _ := store.Get(p.ID)
	if got.Source != "mcp:claude-desktop" {
		t.Errorf("Source = %s, want mcp:claude-desktop", got.Source)
	}
}

func TestIntakeChannel_LLM_CancelledContext(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "ctx task"))
	store := testStore(t, pool)

	ch := NewLLMIntakeChannel(store, "mcp:test")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), testPayload(t, "cancelled "+uuid.New().String()), "", "reason")
	err := ch.Suggest(ctx, p)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestIntakeChannel_NilStore(t *testing.T) {
	t.Parallel()

	ch := NewLLMIntakeChannel(nil, "mcp:test")
	health := ch.HealthCheck()
	if health.Status != HealthDown {
		t.Errorf("HealthCheck with nil store = %s, want down", health.Status)
	}
}

func TestProposalStore_NewFromNonexistentDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nested", "deep")
	path := filepath.Join(dir, "proposals.jsonl")
	store, err := NewProposalStore(path, nil)
	if err != nil {
		t.Fatalf("NewProposalStore: %v", err)
		return
	}

	// Verify the directory was created.
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("directory not created: %v", err)
	}

	// Store should work.
	all := store.List(ProposalFilter{})
	if len(all) != 0 {
		t.Errorf("expected empty store, got %d", len(all))
	}
}

func TestProposalToolsViaServer(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	task := testTask(taskID, "server test task")
	pool := testPool(t, task)

	store := testStore(t, pool)

	registry := core.NewRegistry()
	session := core.NewSessionTracker()

	server := NewMCPServer(registry, nil, pool, session, nil, "test")
	server.SetProposalStore(store)

	// Initialize to set client info.
	initParams, _ := json.Marshal(InitializeParams{
		ProtocolVersion: MCPVersion,
		ClientInfo:      EntityInfo{Name: "claude-desktop", Version: "1.0"},
	})
	initReq := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  initParams,
	}
	raw, _ := json.Marshal(initReq)
	_, _ = server.HandleRequest(raw)

	// Test propose_enrichment tool.
	enrichArgs, _ := json.Marshal(map[string]any{
		"task_id":   taskID,
		"type":      "add-note",
		"payload":   map[string]string{"text": "enrichment via tool " + uuid.New().String()},
		"rationale": "testing tool handler",
	})
	callParams, _ := json.Marshal(ToolCallParams{
		Name:      "propose_enrichment",
		Arguments: enrichArgs,
	})
	callReq := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/call",
		Params:  callParams,
	}
	raw, _ = json.Marshal(callReq)
	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
		return
	}

	// Verify proposal was created.
	pending := store.List(ProposalFilter{Status: ProposalPending})
	if len(pending) != 1 {
		t.Errorf("expected 1 pending proposal, got %d", len(pending))
	}
	if len(pending) > 0 && pending[0].Source != "mcp:claude-desktop" {
		t.Errorf("Source = %s, want mcp:claude-desktop", pending[0].Source)
	}
}

func TestProposalToolsViaServer_SuggestTask(t *testing.T) {
	t.Parallel()

	pool := testPool(t)
	store := testStore(t, pool)

	registry := core.NewRegistry()
	session := core.NewSessionTracker()

	server := NewMCPServer(registry, nil, pool, session, nil, "test")
	server.SetProposalStore(store)

	suggestArgs, _ := json.Marshal(map[string]any{
		"text":      "new task suggestion " + uuid.New().String(),
		"context":   "some context",
		"effort":    "quick-win",
		"rationale": "useful task",
	})
	callParams, _ := json.Marshal(ToolCallParams{
		Name:      "suggest_task",
		Arguments: suggestArgs,
	})
	callReq := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  callParams,
	}
	raw, _ := json.Marshal(callReq)
	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
		return
	}

	all := store.List(ProposalFilter{})
	if len(all) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(all))
	}
}

func TestProposalToolsViaServer_SuggestRelationship(t *testing.T) {
	t.Parallel()

	fromID := uuid.New().String()
	toID := uuid.New().String()
	pool := testPool(t, testTask(fromID, "from task"), testTask(toID, "to task"))
	store := testStore(t, pool)

	registry := core.NewRegistry()
	session := core.NewSessionTracker()

	server := NewMCPServer(registry, nil, pool, session, nil, "test")
	server.SetProposalStore(store)

	relArgs, _ := json.Marshal(map[string]any{
		"from_id":       fromID,
		"to_id":         toID,
		"relation_type": "blocks",
		"rationale":     "from blocks to",
	})
	callParams, _ := json.Marshal(ToolCallParams{
		Name:      "suggest_relationship",
		Arguments: relArgs,
	})
	callReq := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "tools/call",
		Params:  callParams,
	}
	raw, _ := json.Marshal(callReq)
	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
		return
	}

	all := store.List(ProposalFilter{})
	if len(all) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(all))
	}
}

func TestPendingProposalsResource(t *testing.T) {
	t.Parallel()

	taskID := uuid.New().String()
	pool := testPool(t, testTask(taskID, "resource test task"))
	store := testStore(t, pool)

	registry := core.NewRegistry()
	session := core.NewSessionTracker()

	server := NewMCPServer(registry, nil, pool, session, nil, "test")
	server.SetProposalStore(store)

	// Create a pending proposal.
	payload := testPayload(t, "resource note "+uuid.New().String())
	p, _ := NewProposal(ProposalAddNote, taskID, time.Now().UTC(), payload, "mcp:test", "reason")
	_ = store.Create(p)

	// Read the proposals/pending resource.
	readParams, _ := json.Marshal(ResourceReadParams{URI: "threedoors://proposals/pending"})
	req := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`5`),
		Method:  "resources/read",
		Params:  readParams,
	}
	raw, _ := json.Marshal(req)
	respBytes, err := server.HandleRequest(raw)
	if err != nil {
		t.Fatalf("HandleRequest: %v", err)
		return
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
		return
	}
}

func TestPayloadText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload json.RawMessage
		want    string
	}{
		{"text field", json.RawMessage(`{"text":"hello world"}`), "hello world"},
		{"description field", json.RawMessage(`{"description":"desc"}`), "desc"},
		{"note field", json.RawMessage(`{"note":"a note"}`), "a note"},
		{"no text field", json.RawMessage(`{"foo":"bar"}`), `{"foo":"bar"}`},
		{"invalid json", json.RawMessage(`not json`), "not json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := payloadText(tt.payload)
			if got != tt.want {
				t.Errorf("payloadText() = %q, want %q", got, tt.want)
			}
		})
	}
}
