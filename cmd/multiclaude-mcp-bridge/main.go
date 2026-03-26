package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcavenae/ThreeDoors/internal/mcpbridge"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "multiclaude-mcp-bridge: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	server := mcpbridge.NewBridgeServer(&mcpbridge.ExecRunner{}, version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return serveStdio(ctx, server)
}

func serveStdio(ctx context.Context, server *mcpbridge.BridgeServer) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			return nil
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp, err := server.HandleRequest(line)
		if err != nil {
			return fmt.Errorf("handle request: %w", err)
		}
		if resp == nil {
			continue
		}

		resp = append(resp, '\n')
		if _, err := os.Stdout.Write(resp); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
}
