package main

import (
	"testing"
)

func TestParseCheckbox(t *testing.T) {
	tests := []struct {
		input      string
		wantText   string
		wantStatus string
	}{
		{"- [ ] Buy groceries", "Buy groceries", "todo"},
		{"- [x] Set up dev env", "Set up dev env", "complete"},
		{"- [X] Done task", "Done task", "complete"},
		{"* [ ] Star checkbox", "Star checkbox", "todo"},
		{"* [x] Star complete", "Star complete", "complete"},
		{"Plain text line", "Plain text line", "todo"},
		{"  - [ ] Indented task", "  - [ ] Indented task", "todo"}, // note: parseCheckbox doesn't trim; parseTasks trims first
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotText, gotStatus := parseCheckbox(tt.input)
			if gotText != tt.wantText {
				t.Errorf("parseCheckbox(%q) text = %q, want %q", tt.input, gotText, tt.wantText)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("parseCheckbox(%q) status = %q, want %q", tt.input, gotStatus, tt.wantStatus)
			}
		})
	}
}

func TestParseTasks(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst string
	}{
		{
			name:      "standard checklist",
			input:     "- [ ] Task one\n- [x] Task two\n- [ ] Task three",
			wantCount: 3,
			wantFirst: "Task one",
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "whitespace only",
			input:     "   \n  \n  ",
			wantCount: 0,
		},
		{
			name:      "mixed with empty lines",
			input:     "- [ ] Task one\n\n- [x] Task two\n\n",
			wantCount: 2,
			wantFirst: "Task one",
		},
		{
			name:      "plain text lines",
			input:     "Some heading\nAnother line",
			wantCount: 2,
			wantFirst: "Some heading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := parseTasks(tt.input)
			if len(tasks) != tt.wantCount {
				t.Errorf("parseTasks() returned %d tasks, want %d", len(tasks), tt.wantCount)
			}
			if tt.wantCount > 0 && len(tasks) > 0 && tasks[0].Text != tt.wantFirst {
				t.Errorf("first task text = %q, want %q", tasks[0].Text, tt.wantFirst)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	tests := []struct {
		name string
		p    float64
		want float64
	}{
		{"p50", 0.50, 55},
		{"p0", 0.0, 10},
		{"p100", 1.0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(sorted, tt.p)
			if got != tt.want {
				t.Errorf("percentile(sorted, %v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestPercentileEmpty(t *testing.T) {
	got := percentile([]float64{}, 0.5)
	if got != 0 {
		t.Errorf("percentile(empty, 0.5) = %v, want 0", got)
	}
}
