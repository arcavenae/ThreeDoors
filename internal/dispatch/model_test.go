package dispatch

import "testing"

func TestQueueItemStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status QueueItemStatus
		want   string
	}{
		{"pending", QueueItemPending, "pending"},
		{"dispatched", QueueItemDispatched, "dispatched"},
		{"completed", QueueItemCompleted, "completed"},
		{"failed", QueueItemFailed, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.String(); got != tt.want {
				t.Errorf("QueueItemStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
