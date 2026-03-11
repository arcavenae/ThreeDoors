package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func echoHandler(req *Request) *Response {
	return NewResponse(req.ID, map[string]string{"method": req.Method})
}

func TestRateLimiter_GlobalLimit(t *testing.T) {
	t.Parallel()

	cfg := RateLimitConfig{
		GlobalPerMin:    5,
		ProposalsPerMin: 100,
		QueriesPerMin:   100,
		MaxPending:      5,
		BurstAllowance:  5,
	}
	rl := NewRateLimiter(cfg)
	handler := rl.Middleware()(echoHandler)

	reqID := json.RawMessage(`1`)

	// First 5 should succeed (burst allowance).
	for i := 0; i < 5; i++ {
		resp := handler(&Request{ID: reqID, Method: "test"})
		if resp.Error != nil {
			t.Fatalf("request %d: unexpected error: %v", i, resp.Error.Message)
			return
		}
	}

	// 6th should be rate limited.
	resp := handler(&Request{ID: reqID, Method: "test"})
	if resp.Error == nil {
		t.Fatal("expected rate limit error")
		return
	}
	if resp.Error.Code != CodeRateLimited {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeRateLimited)
	}
}

func TestRateLimiter_RetryAfterHint(t *testing.T) {
	t.Parallel()

	cfg := RateLimitConfig{
		GlobalPerMin:    10,
		ProposalsPerMin: 100,
		QueriesPerMin:   100,
		MaxPending:      5,
		BurstAllowance:  1,
	}
	rl := NewRateLimiter(cfg)
	handler := rl.Middleware()(echoHandler)

	reqID := json.RawMessage(`1`)

	// Exhaust burst.
	handler(&Request{ID: reqID, Method: "test"})

	resp := handler(&Request{ID: reqID, Method: "test"})
	if resp.Error == nil {
		t.Fatal("expected rate limit error")
		return
	}

	data, ok := resp.Error.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data map in error response")
	}
	retryAfter, ok := data["retry_after_seconds"]
	if !ok {
		t.Fatal("expected retry_after_seconds in error data")
	}
	if retryAfter.(float64) <= 0 {
		t.Errorf("retry_after_seconds should be positive, got %v", retryAfter)
	}
}

func TestRateLimiter_QueryLimit(t *testing.T) {
	t.Parallel()

	cfg := RateLimitConfig{
		GlobalPerMin:    1000,
		ProposalsPerMin: 100,
		QueriesPerMin:   3,
		MaxPending:      5,
		BurstAllowance:  3,
	}
	rl := NewRateLimiter(cfg)
	handler := rl.Middleware()(echoHandler)

	reqID := json.RawMessage(`1`)

	for i := 0; i < 3; i++ {
		resp := handler(&Request{ID: reqID, Method: "resources/list"})
		if resp.Error != nil {
			t.Fatalf("request %d: unexpected error: %v", i, resp.Error.Message)
			return
		}
	}

	resp := handler(&Request{ID: reqID, Method: "tools/list"})
	if resp.Error == nil {
		t.Fatal("expected query rate limit error")
		return
	}
	if resp.Error.Code != CodeRateLimited {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeRateLimited)
	}
}

func TestRateLimiter_NonQueryPassesQueryLimit(t *testing.T) {
	t.Parallel()

	// Use rate.Inf-like global rate (very high) so only query limiter matters.
	cfg := RateLimitConfig{
		GlobalPerMin:    60000,
		ProposalsPerMin: 60000,
		QueriesPerMin:   60000,
		MaxPending:      5,
		BurstAllowance:  3,
	}
	rl := NewRateLimiter(cfg)
	// Override the query limiter to be very restrictive.
	rl.queries = rate.NewLimiter(rate.Limit(1.0/60.0), 1)

	handler := rl.Middleware()(echoHandler)
	reqID := json.RawMessage(`1`)

	// Exhaust query burst (1 token).
	handler(&Request{ID: reqID, Method: "resources/list"})

	// Next query should fail.
	qResp := handler(&Request{ID: reqID, Method: "resources/list"})
	if qResp.Error == nil {
		t.Fatal("expected query rate limit error")
		return
	}

	// Non-query should still pass.
	resp := handler(&Request{ID: reqID, Method: "initialize"})
	if resp.Error != nil {
		t.Errorf("non-query request should not be query-limited: %v", resp.Error.Message)
	}
}

func TestAuditLogger_WritesEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	fixedTime := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	al.nowFunc = func() time.Time { return fixedTime }

	handler := al.Middleware()(echoHandler)
	reqID := json.RawMessage(`42`)
	handler(&Request{ID: reqID, Method: "tools/list"})

	logFile := filepath.Join(dir, "mcp-audit-2026-03-07.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
		return
	}

	var entry AuditEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal entry: %v", err)
	}

	if entry.RequestID != "42" {
		t.Errorf("req_id = %q, want %q", entry.RequestID, "42")
	}
	if entry.Tool != "tools/list" {
		t.Errorf("tool = %q, want %q", entry.Tool, "tools/list")
	}
	if entry.Result != "ok" {
		t.Errorf("result = %q, want %q", entry.Result, "ok")
	}
}

func TestAuditLogger_HashChain(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	fixedTime := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	al.nowFunc = func() time.Time { return fixedTime }

	handler := al.Middleware()(echoHandler)
	reqID := json.RawMessage(`1`)

	handler(&Request{ID: reqID, Method: "tools/list"})
	handler(&Request{ID: reqID, Method: "resources/list"})

	logFile := filepath.Join(dir, "mcp-audit-2026-03-07.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(lines))
	}

	var entry1, entry2 AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("unmarshal entry1: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("unmarshal entry2: %v", err)
	}

	// First entry has empty prev_hash.
	if entry1.PrevHash != "" {
		t.Errorf("first entry prev_hash should be empty, got %q", entry1.PrevHash)
	}

	// Second entry has non-empty prev_hash.
	if entry2.PrevHash == "" {
		t.Error("second entry prev_hash should not be empty")
	}
	if len(entry2.PrevHash) != 64 {
		t.Errorf("prev_hash length = %d, want 64 (SHA-256 hex)", len(entry2.PrevHash))
	}
}

func TestAuditLogger_RecordsErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	fixedTime := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	al.nowFunc = func() time.Time { return fixedTime }

	errorHandler := func(_ *Request) *Response {
		return NewErrorResponse(json.RawMessage(`1`), CodeInternalError, "something broke")
	}

	handler := al.Middleware()(errorHandler)
	handler(&Request{ID: json.RawMessage(`1`), Method: "test"})

	logFile := filepath.Join(dir, "mcp-audit-2026-03-07.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
		return
	}

	var entry AuditEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if entry.Result != "error" {
		t.Errorf("result = %q, want %q", entry.Result, "error")
	}
	if entry.Error != "something broke" {
		t.Errorf("error = %q, want %q", entry.Error, "something broke")
	}
}

func TestAuditLogger_RecordsRateLimited(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	fixedTime := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	al.nowFunc = func() time.Time { return fixedTime }

	rateLimitedHandler := func(_ *Request) *Response {
		return NewErrorResponseWithData(json.RawMessage(`1`), CodeRateLimited, "rate limited", nil)
	}

	handler := al.Middleware()(rateLimitedHandler)
	handler(&Request{ID: json.RawMessage(`1`), Method: "test"})

	logFile := filepath.Join(dir, "mcp-audit-2026-03-07.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
		return
	}

	var entry AuditEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if entry.Result != "rate_limited" {
		t.Errorf("result = %q, want %q", entry.Result, "rate_limited")
	}
}

func TestAuditLogger_RotateLogs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	now := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	al.nowFunc = func() time.Time { return now }

	// Create old and new log files.
	oldFile := filepath.Join(dir, "mcp-audit-2026-01-01.jsonl")
	newFile := filepath.Join(dir, "mcp-audit-2026-03-06.jsonl")
	nonAuditFile := filepath.Join(dir, "other-file.txt")

	for _, f := range []string{oldFile, newFile, nonAuditFile} {
		if err := os.WriteFile(f, []byte("test\n"), 0o600); err != nil {
			t.Fatalf("create file %s: %v", f, err)
		}
	}

	if err := al.RotateLogs(); err != nil {
		t.Fatalf("RotateLogs: %v", err)
	}

	// Old file should be removed (> 30 days).
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old audit file should have been removed")
	}

	// New file should still exist.
	if _, err := os.Stat(newFile); err != nil {
		t.Error("recent audit file should still exist")
	}

	// Non-audit file should still exist.
	if _, err := os.Stat(nonAuditFile); err != nil {
		t.Error("non-audit file should still exist")
	}
}

func TestSchemaValidator_ValidUUID(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	params, _ := json.Marshal(map[string]string{
		"task_id": "550e8400-e29b-41d4-a716-446655440000",
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error != nil {
		t.Errorf("valid UUID should pass: %v", resp.Error.Message)
	}
}

func TestSchemaValidator_InvalidUUID(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	params, _ := json.Marshal(map[string]string{
		"task_id": "not-a-uuid",
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error == nil {
		t.Fatal("expected error for invalid UUID")
		return
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestSchemaValidator_TextTooLong(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	longText := strings.Repeat("a", 501)
	params, _ := json.Marshal(map[string]string{
		"text": longText,
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error == nil {
		t.Fatal("expected error for text too long")
		return
	}
	if !strings.Contains(resp.Error.Message, "500 character limit") {
		t.Errorf("error message should mention limit: %q", resp.Error.Message)
	}
}

func TestSchemaValidator_TextWithinLimit(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	params, _ := json.Marshal(map[string]string{
		"text": strings.Repeat("a", 500),
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error != nil {
		t.Errorf("text within limit should pass: %v", resp.Error.Message)
	}
}

func TestSchemaValidator_ValidStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
		valid  bool
	}{
		{"todo", "todo", true},
		{"blocked", "blocked", true},
		{"in-progress", "in-progress", true},
		{"complete", "complete", true},
		{"deferred", "deferred", true},
		{"archived", "archived", true},
		{"in-review", "in-review", true},
		{"invalid", "nonsense", false},
		{"empty", "", false},
	}

	mw := SchemaValidator()
	handler := mw(echoHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, _ := json.Marshal(map[string]string{"status": tt.status})
			resp := handler(&Request{
				ID:     json.RawMessage(`1`),
				Method: "tools/call",
				Params: params,
			})

			if tt.valid && resp.Error != nil {
				t.Errorf("status %q should be valid: %v", tt.status, resp.Error.Message)
			}
			if !tt.valid && resp.Error == nil {
				t.Errorf("status %q should be invalid", tt.status)
			}
		})
	}
}

func TestSchemaValidator_FutureTimestamp(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	futureTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	params, _ := json.Marshal(map[string]string{
		"timestamp": futureTime,
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error == nil {
		t.Fatal("expected error for future timestamp")
		return
	}
	if !strings.Contains(resp.Error.Message, "future") {
		t.Errorf("error should mention future: %q", resp.Error.Message)
	}
}

func TestSchemaValidator_ValidTimestamp(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	pastTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	params, _ := json.Marshal(map[string]string{
		"timestamp": pastTime,
	})

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error != nil {
		t.Errorf("valid timestamp should pass: %v", resp.Error.Message)
	}
}

func TestSchemaValidator_NoParams(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	resp := handler(&Request{ID: json.RawMessage(`1`), Method: "test"})
	if resp.Error != nil {
		t.Errorf("nil params should pass through: %v", resp.Error.Message)
	}
}

func TestReadOnlyEnforcer_BlocksDirectWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tool    string
		blocked bool
	}{
		{"save task", "save_task", true},
		{"delete task", "delete_task", true},
		{"create task", "create_task", true},
		{"update task", "update_task", true},
		{"remove task", "remove_task", true},
		{"write data", "write_data", true},
		{"propose change", "propose_change", false},
		{"create proposal", "create_proposal", false},
		{"list tasks", "list_tasks", false},
		{"get task", "get_task", false},
		{"query doors", "query_doors", false},
	}

	mw := ReadOnlyEnforcer()
	handler := mw(echoHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, _ := json.Marshal(map[string]string{"name": tt.tool})
			resp := handler(&Request{
				ID:     json.RawMessage(`1`),
				Method: "tools/call",
				Params: params,
			})

			if tt.blocked && resp.Error == nil {
				t.Errorf("tool %q should be blocked", tt.tool)
			}
			if !tt.blocked && resp.Error != nil {
				t.Errorf("tool %q should not be blocked: %v", tt.tool, resp.Error.Message)
			}
		})
	}
}

func TestReadOnlyEnforcer_AllowsNonToolsMethods(t *testing.T) {
	t.Parallel()

	mw := ReadOnlyEnforcer()
	handler := mw(echoHandler)

	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/list",
	})

	if resp.Error != nil {
		t.Errorf("non-tools method should not be blocked: %v", resp.Error.Message)
	}
}

func TestReadOnlyEnforcer_ErrorCode(t *testing.T) {
	t.Parallel()

	mw := ReadOnlyEnforcer()
	handler := mw(echoHandler)

	params, _ := json.Marshal(map[string]string{"name": "save_task"})
	resp := handler(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/call",
		Params: params,
	})

	if resp.Error == nil {
		t.Fatal("expected error")
		return
	}
	if resp.Error.Code != CodeReadOnly {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeReadOnly)
	}
}

func TestMiddlewareChainOrder(t *testing.T) {
	t.Parallel()

	var order []string

	mw1 := func(next Handler) Handler {
		return func(req *Request) *Response {
			order = append(order, "ReadOnlyEnforcer")
			return next(req)
		}
	}
	mw2 := func(next Handler) Handler {
		return func(req *Request) *Response {
			order = append(order, "RateLimiter")
			return next(req)
		}
	}
	mw3 := func(next Handler) Handler {
		return func(req *Request) *Response {
			order = append(order, "AuditLogger")
			return next(req)
		}
	}
	mw4 := func(next Handler) Handler {
		return func(req *Request) *Response {
			order = append(order, "SchemaValidator")
			return next(req)
		}
	}

	// Simulate the same order as ApplyDefaultMiddleware:
	// Use in order: SchemaValidator, AuditLogger, RateLimiter, ReadOnlyEnforcer
	// Because buildHandler reverses, the execution order becomes:
	// ReadOnlyEnforcer -> RateLimiter -> AuditLogger -> SchemaValidator -> handler
	h := echoHandler
	for _, mw := range []Middleware{mw4, mw3, mw2, mw1} {
		h = mw(h)
	}

	h(&Request{ID: json.RawMessage(`1`), Method: "test"})

	expected := []string{"ReadOnlyEnforcer", "RateLimiter", "AuditLogger", "SchemaValidator"}
	if len(order) != len(expected) {
		t.Fatalf("order length = %d, want %d", len(order), len(expected))
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, order[i], name)
		}
	}
}

func TestApplyDefaultMiddleware(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	server := &MCPServer{}
	server.handler = echoHandler

	al := ApplyDefaultMiddleware(server, dir)
	if al == nil {
		t.Fatal("ApplyDefaultMiddleware returned nil audit logger")
		return
	}

	if len(server.middleware) != 4 {
		t.Errorf("expected 4 middleware, got %d", len(server.middleware))
	}
}

func TestAuditLogger_DailyFilename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	al := NewAuditLogger(dir)

	day1 := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)

	al.nowFunc = func() time.Time { return day1 }
	handler1 := al.Middleware()(echoHandler)
	handler1(&Request{ID: json.RawMessage(`1`), Method: "test"})

	al.nowFunc = func() time.Time { return day2 }
	handler2 := al.Middleware()(echoHandler)
	handler2(&Request{ID: json.RawMessage(`2`), Method: "test"})

	file1 := filepath.Join(dir, "mcp-audit-2026-03-07.jsonl")
	file2 := filepath.Join(dir, "mcp-audit-2026-03-08.jsonl")

	if _, err := os.Stat(file1); err != nil {
		t.Errorf("expected day 1 log file: %v", err)
	}
	if _, err := os.Stat(file2); err != nil {
		t.Errorf("expected day 2 log file: %v", err)
	}
}

func TestSchemaValidator_MultipleIDFields(t *testing.T) {
	t.Parallel()

	mw := SchemaValidator()
	handler := mw(echoHandler)

	tests := []struct {
		name  string
		field string
	}{
		{"task_id", "task_id"},
		{"taskId", "taskId"},
		{"id", "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, _ := json.Marshal(map[string]string{
				tt.field: "not-a-uuid",
			})
			resp := handler(&Request{
				ID:     json.RawMessage(`1`),
				Method: "test",
				Params: params,
			})

			if resp.Error == nil {
				t.Errorf("field %q with invalid UUID should fail", tt.field)
			}
		})
	}
}
