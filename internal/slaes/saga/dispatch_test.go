package saga

import (
	"strings"
	"testing"
	"time"
)

func TestNewDispatchTracker(t *testing.T) {
	t.Parallel()
	dt := NewDispatchTracker(4 * time.Hour)
	if dt == nil {
		t.Fatal("expected non-nil tracker")
	}
	if len(dt.Records()) != 0 {
		t.Errorf("expected 0 records, got %d", len(dt.Records()))
	}
}

func TestDispatchTracker_AddRecord(t *testing.T) {
	t.Parallel()
	dt := NewDispatchTracker(4 * time.Hour)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	dt.AddRecord(WorkerRecord{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: now})
	dt.AddRecord(WorkerRecord{Name: "w2", Branch: "fix/ci", Status: "active", Timestamp: now.Add(1 * time.Hour)})

	records := dt.Records()
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].Name != "w1" || records[1].Name != "w2" {
		t.Errorf("unexpected record names: %s, %s", records[0].Name, records[1].Name)
	}
}

func TestDispatchTracker_RecordsReturnsCopy(t *testing.T) {
	t.Parallel()
	dt := NewDispatchTracker(4 * time.Hour)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	dt.AddRecord(WorkerRecord{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: now})

	records := dt.Records()
	records[0].Name = "modified"

	if dt.Records()[0].Name != "w1" {
		t.Error("Records() should return a copy, not a reference")
	}
}

func TestDispatchTracker_DetectOverlaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		window     time.Duration
		records    []WorkerRecord
		wantCount  int
		wantBranch string
	}{
		{
			name:   "two workers same branch within window",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
				{Name: "w2", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)},
			},
			wantCount:  1,
			wantBranch: "fix/ci",
		},
		{
			name:   "two workers same branch outside window",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)},
				{Name: "w2", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)},
			},
			wantCount: 0,
		},
		{
			name:   "two workers different branches",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
				{Name: "w2", Branch: "feat/new", Status: "active", Timestamp: time.Date(2026, 3, 10, 13, 0, 0, 0, time.UTC)},
			},
			wantCount: 0,
		},
		{
			name:   "same worker same branch not a saga",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 13, 0, 0, 0, time.UTC)},
			},
			wantCount: 0,
		},
		{
			name:   "three workers same branch escalation",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "failed", Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)},
				{Name: "w2", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC)},
				{Name: "w3", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
			},
			wantCount:  1,
			wantBranch: "fix/ci",
		},
		{
			name:      "no records",
			window:    4 * time.Hour,
			records:   nil,
			wantCount: 0,
		},
		{
			name:   "single record",
			window: 4 * time.Hour,
			records: []WorkerRecord{
				{Name: "w1", Branch: "fix/ci", Status: "active", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dt := NewDispatchTracker(tt.window)
			for _, r := range tt.records {
				dt.AddRecord(r)
			}

			overlaps := dt.DetectOverlaps()
			if len(overlaps) != tt.wantCount {
				t.Errorf("expected %d overlaps, got %d", tt.wantCount, len(overlaps))
			}
			if tt.wantCount > 0 && len(overlaps) > 0 && overlaps[0].Branch != tt.wantBranch {
				t.Errorf("expected branch %q, got %q", tt.wantBranch, overlaps[0].Branch)
			}
		})
	}
}

func TestParseWorkerList(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid output with header",
			input: `NAME	BRANCH	STATUS
w1	fix/ci-lint	active
w2	fix/ci-lint	active
w3	feat/new-thing	active`,
			wantCount: 3,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "only comments and blanks",
			input:     "# workers\n\n# end\n",
			wantCount: 0,
		},
		{
			name:    "invalid line too few fields",
			input:   "w1 fix/ci",
			wantErr: true,
		},
		{
			name: "space separated",
			input: `w1 fix/ci active
w2 fix/ci completed`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			records, err := ParseWorkerList(strings.NewReader(tt.input), now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && len(records) != tt.wantCount {
				t.Errorf("expected %d records, got %d", tt.wantCount, len(records))
			}
			if !tt.wantErr {
				for _, r := range records {
					if !r.Timestamp.Equal(now) {
						t.Errorf("expected timestamp %v, got %v", now, r.Timestamp)
					}
				}
			}
		})
	}
}

func TestParseWorkerList_FieldValues(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	input := "kind-panda	work/kind-panda	active\n"

	records, err := ParseWorkerList(strings.NewReader(input), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	r := records[0]
	if r.Name != "kind-panda" {
		t.Errorf("name: got %q, want %q", r.Name, "kind-panda")
	}
	if r.Branch != "work/kind-panda" {
		t.Errorf("branch: got %q, want %q", r.Branch, "work/kind-panda")
	}
	if r.Status != "active" {
		t.Errorf("status: got %q, want %q", r.Status, "active")
	}
}
