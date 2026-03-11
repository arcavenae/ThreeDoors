package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestStdioTransportRoundTrip(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	initReq := Request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}
	reqBytes, _ := json.Marshal(initReq)

	input := bytes.NewReader(append(reqBytes, '\n'))
	var output bytes.Buffer

	transport := NewStdioTransport(input, &output)

	ctx := context.Background()
	if err := transport.Serve(ctx, server); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 response line, got %d: %q", len(lines), output.String())
	}

	var resp Response
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}
}

func TestStdioTransportMultipleMessages(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	var inputBuf bytes.Buffer
	for i := 1; i <= 3; i++ {
		req := Request{
			JSONRPC: jsonRPCVersion,
			ID:      json.RawMessage(fmt.Sprintf(`%d`, i)),
			Method:  "resources/list",
		}
		b, _ := json.Marshal(req)
		inputBuf.Write(b)
		inputBuf.WriteByte('\n')
	}

	var output bytes.Buffer
	transport := NewStdioTransport(&inputBuf, &output)

	if err := transport.Serve(context.Background(), server); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 response lines, got %d", len(lines))
	}
}

func TestStdioTransportSkipsEmptyLines(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	req := Request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage(`1`),
		Method:  "tools/list",
	}
	reqBytes, _ := json.Marshal(req)

	input := bytes.NewReader([]byte("\n\n" + string(reqBytes) + "\n\n"))
	var output bytes.Buffer

	transport := NewStdioTransport(input, &output)

	if err := transport.Serve(context.Background(), server); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 response line, got %d", len(lines))
	}
}

func TestStdioTransportNotificationNoOutput(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	notif := Request{
		JSONRPC: jsonRPCVersion,
		Method:  "notifications/initialized",
	}
	notifBytes, _ := json.Marshal(notif)

	input := bytes.NewReader(append(notifBytes, '\n'))
	var output bytes.Buffer

	transport := NewStdioTransport(input, &output)

	if err := transport.Serve(context.Background(), server); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	if output.Len() != 0 {
		t.Errorf("expected no output for notification, got %q", output.String())
	}
}

func TestStdioTransportContextCancellation(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")

	r, _ := io.Pipe()
	var output bytes.Buffer

	transport := NewStdioTransport(r, &output)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := transport.Serve(ctx, server)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSSETransportMessageEndpoint(t *testing.T) {
	t.Parallel()

	server := NewMCPServer(core.NewRegistry(), nil, core.NewTaskPool(), core.NewSessionTracker(), nil, "test")
	transport := NewSSETransport(":0")

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", transport.handleSSE(server))
	mux.HandleFunc("/message", transport.handleMessage(server))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
		return
	}

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	baseURL := fmt.Sprintf("http://%s", ln.Addr().String())

	// GET /message is rejected
	resp, err := http.Get(baseURL + "/message?sessionId=test")
	if err != nil {
		t.Fatalf("GET /message: %v", err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /message status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}

	// POST /message without sessionId is rejected
	resp, err = http.Post(baseURL+"/message", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /message no session: %v", err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /message no session status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// POST /message with unknown sessionId is rejected
	resp, err = http.Post(baseURL+"/message?sessionId=unknown", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /message unknown session: %v", err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("POST /message unknown session status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestTransportFromFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		transportType string
		wantType      string
	}{
		{"default is stdio", "", "*mcp.StdioTransport"},
		{"stdio explicit", "stdio", "*mcp.StdioTransport"},
		{"sse", "sse", "*mcp.SSETransport"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transport := TransportFromFlags(tt.transportType, ":8080", nil, nil)
			got := fmt.Sprintf("%T", transport)
			if got != tt.wantType {
				t.Errorf("type = %s, want %s", got, tt.wantType)
			}
		})
	}
}
