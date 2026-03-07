package dispatch

import "time"

// DevDispatch tracks dev dispatch state orthogonal to task lifecycle.
// A nil pointer means "never dispatched".
type DevDispatch struct {
	Queued      bool       `yaml:"queued,omitempty" json:"queued,omitempty"`
	QueuedAt    *time.Time `yaml:"queued_at,omitempty" json:"queued_at,omitempty"`
	WorkerName  string     `yaml:"worker_name,omitempty" json:"worker_name,omitempty"`
	PRNumber    int        `yaml:"pr_number,omitempty" json:"pr_number,omitempty"`
	PRStatus    string     `yaml:"pr_status,omitempty" json:"pr_status,omitempty"`
	DispatchErr string     `yaml:"dispatch_err,omitempty" json:"dispatch_err,omitempty"`
}

// QueueItemStatus represents the lifecycle state of a queue item.
type QueueItemStatus string

const (
	QueueItemPending    QueueItemStatus = "pending"
	QueueItemDispatched QueueItemStatus = "dispatched"
	QueueItemCompleted  QueueItemStatus = "completed"
	QueueItemFailed     QueueItemStatus = "failed"
)

// String returns the string representation of the status.
func (s QueueItemStatus) String() string {
	return string(s)
}

// QueueItem represents a single entry in the dev dispatch queue.
type QueueItem struct {
	ID                 string          `yaml:"id" json:"id"`
	TaskID             string          `yaml:"task_id" json:"task_id"`
	TaskText           string          `yaml:"task_text" json:"task_text"`
	Context            string          `yaml:"context,omitempty" json:"context,omitempty"`
	Status             QueueItemStatus `yaml:"status" json:"status"`
	Priority           int             `yaml:"priority,omitempty" json:"priority,omitempty"`
	Scope              string          `yaml:"scope,omitempty" json:"scope,omitempty"`
	AcceptanceCriteria []string        `yaml:"acceptance_criteria,omitempty" json:"acceptance_criteria,omitempty"`
	QueuedAt           *time.Time      `yaml:"queued_at,omitempty" json:"queued_at,omitempty"`
	DispatchedAt       *time.Time      `yaml:"dispatched_at,omitempty" json:"dispatched_at,omitempty"`
	CompletedAt        *time.Time      `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	WorkerName         string          `yaml:"worker_name,omitempty" json:"worker_name,omitempty"`
	PRNumber           int             `yaml:"pr_number,omitempty" json:"pr_number,omitempty"`
	PRURL              string          `yaml:"pr_url,omitempty" json:"pr_url,omitempty"`
	Error              string          `yaml:"error,omitempty" json:"error,omitempty"`
}
