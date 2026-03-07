package dispatch

import (
	"context"
	"time"
)

// WorkerInfo represents the state of a multiclaude worker.
type WorkerInfo struct {
	Name   string
	Status string
	Branch string
	Task   string
}

// HistoryEntry represents a completed worker from multiclaude history.
type HistoryEntry struct {
	WorkerName  string
	Status      string
	PRNumber    int
	PRURL       string
	Summary     string
	CompletedAt *time.Time
}

// Dispatcher defines the interface for dispatching work to multiclaude workers.
type Dispatcher interface {
	CreateWorker(ctx context.Context, task string) (workerName string, err error)
	ListWorkers(ctx context.Context) ([]WorkerInfo, error)
	GetHistory(ctx context.Context, limit int) ([]HistoryEntry, error)
	RemoveWorker(ctx context.Context, name string) error
	CheckAvailable(ctx context.Context) error
}
