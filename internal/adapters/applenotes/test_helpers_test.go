package applenotes

import (
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

var baseTime = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

func newTestTask(id, text string, status core.TaskStatus, updatedAt time.Time) *core.Task {
	return &core.Task{
		ID:        id,
		Text:      text,
		Status:    status,
		Notes:     []core.TaskNote{},
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	}
}
