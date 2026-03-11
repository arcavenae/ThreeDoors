package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBreadcrumbTrail_Record_Empty(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	result := trail.Format()
	if result != "" {
		t.Errorf("empty trail Format() = %q, want empty string", result)
	}
}

func TestBreadcrumbTrail_Record_Single(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	trail.Record("Doors", "key:Enter")

	result := trail.Format()
	if !strings.Contains(result, "key:Enter") {
		t.Errorf("Format() = %q, want to contain %q", result, "key:Enter")
	}
	if !strings.Contains(result, "Doors") {
		t.Errorf("Format() = %q, want to contain %q", result, "Doors")
	}
}

func TestBreadcrumbTrail_RingBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entries   int
		wantCount int
		wantFirst string
		wantLast  string
	}{
		{
			name:      "partial fill",
			entries:   10,
			wantCount: 10,
			wantFirst: "action-0",
			wantLast:  "action-9",
		},
		{
			name:      "exact capacity",
			entries:   BreadcrumbCapacity,
			wantCount: BreadcrumbCapacity,
			wantFirst: "action-0",
			wantLast:  fmt.Sprintf("action-%d", BreadcrumbCapacity-1),
		},
		{
			name:      "overflow by one",
			entries:   BreadcrumbCapacity + 1,
			wantCount: BreadcrumbCapacity,
			wantFirst: "action-1",
			wantLast:  fmt.Sprintf("action-%d", BreadcrumbCapacity),
		},
		{
			name:      "double capacity",
			entries:   BreadcrumbCapacity * 2,
			wantCount: BreadcrumbCapacity,
			wantFirst: fmt.Sprintf("action-%d", BreadcrumbCapacity),
			wantLast:  fmt.Sprintf("action-%d", BreadcrumbCapacity*2-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			trail := NewBreadcrumbTrail()
			for i := range tt.entries {
				trail.Record("view", fmt.Sprintf("action-%d", i))
			}

			result := trail.Format()
			lines := strings.Split(strings.TrimSpace(result), "\n")
			if len(lines) != tt.wantCount {
				t.Errorf("got %d lines, want %d", len(lines), tt.wantCount)
			}

			if !strings.Contains(lines[0], tt.wantFirst) {
				t.Errorf("first line %q does not contain %q", lines[0], tt.wantFirst)
			}
			if !strings.Contains(lines[len(lines)-1], tt.wantLast) {
				t.Errorf("last line %q does not contain %q", lines[len(lines)-1], tt.wantLast)
			}
		})
	}
}

func TestBreadcrumbTrail_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries int
		want    int
	}{
		{"49th entry", 49, 49},
		{"50th entry", 50, 50},
		{"51st entry", 51, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			trail := NewBreadcrumbTrail()
			for i := range tt.entries {
				trail.Record("view", fmt.Sprintf("e%d", i))
			}
			result := trail.Format()
			lines := strings.Split(strings.TrimSpace(result), "\n")
			if len(lines) != tt.want {
				t.Errorf("got %d entries, want %d", len(lines), tt.want)
			}
		})
	}
}

func TestBreadcrumbTrail_ChronologicalOrder(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()

	// Fill past capacity to force wrap-around
	for i := range BreadcrumbCapacity + 10 {
		trail.Record("view", fmt.Sprintf("action-%d", i))
	}

	result := trail.Format()
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Verify chronological ordering: each line's action number should increase
	prev := -1
	for _, line := range lines {
		var num int
		// Extract action number from line
		for _, part := range strings.Fields(line) {
			if strings.HasPrefix(part, "action-") {
				if _, err := fmt.Sscanf(part, "action-%d", &num); err == nil {
					break
				}
			}
		}
		if num <= prev {
			t.Errorf("non-chronological: action-%d after action-%d", num, prev)
		}
		prev = num
	}
}

func TestBreadcrumbTrail_FormatIncludesTimestamp(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	trail.Record("Doors", "key:Enter")

	result := trail.Format()
	// Should contain UTC timestamp in some form
	now := time.Now().UTC()
	yearStr := fmt.Sprintf("%d", now.Year())
	if !strings.Contains(result, yearStr) {
		t.Errorf("Format() = %q, expected to contain year %q", result, yearStr)
	}
}

func TestBreadcrumbTrail_FormatIncludesViewMode(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	trail.Record("Doors", "key:Enter")
	trail.Record("Detail", "key:Esc")

	result := trail.Format()
	if !strings.Contains(result, "Doors") {
		t.Errorf("Format() missing view mode 'Doors'")
	}
	if !strings.Contains(result, "Detail") {
		t.Errorf("Format() missing view mode 'Detail'")
	}
}

func TestBreadcrumbTrail_ConstantMemory(t *testing.T) {
	t.Parallel()
	// BreadcrumbTrail uses a fixed array — verify the capacity constant.
	if BreadcrumbCapacity != 50 {
		t.Errorf("BreadcrumbCapacity = %d, want 50", BreadcrumbCapacity)
	}
}

func TestBreadcrumbTrail_Entries(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	trail.Record("Doors", "key:Enter")
	trail.Record("Detail", "view:Detail")

	entries := trail.Entries()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Action != "key:Enter" {
		t.Errorf("entries[0].Action = %q, want %q", entries[0].Action, "key:Enter")
	}
	if entries[1].Action != "view:Detail" {
		t.Errorf("entries[1].Action = %q, want %q", entries[1].Action, "view:Detail")
	}
	if entries[0].ViewMode != "Doors" {
		t.Errorf("entries[0].ViewMode = %q, want %q", entries[0].ViewMode, "Doors")
	}
}

func TestBreadcrumbTrail_EntriesChronologicalAfterWrap(t *testing.T) {
	t.Parallel()
	trail := NewBreadcrumbTrail()
	for i := range BreadcrumbCapacity + 5 {
		trail.Record("v", fmt.Sprintf("a-%d", i))
	}

	entries := trail.Entries()
	if len(entries) != BreadcrumbCapacity {
		t.Fatalf("got %d entries, want %d", len(entries), BreadcrumbCapacity)
	}
	// First entry should be a-5 (oldest after 5 were overwritten)
	if entries[0].Action != "a-5" {
		t.Errorf("entries[0].Action = %q, want %q", entries[0].Action, "a-5")
	}
}

func TestViewModeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode ViewMode
		want string
	}{
		{ViewDoors, "Doors"},
		{ViewDetail, "Detail"},
		{ViewSearch, "Search"},
		{ViewMood, "Mood"},
		{ViewHealth, "Health"},
		{ViewHelp, "Help"},
		{ViewPlanning, "Planning"},
		{ViewMode(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("ViewMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}
