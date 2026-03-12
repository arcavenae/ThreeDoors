package tui

import (
	"strings"
	"testing"
)

func TestLayoutFull_ExactHeight(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		header  string
		content string
		footer  string
		height  int
	}{
		{"simple", "header", "content", "footer", 20},
		{"tall terminal", "h", "c", "f", 50},
		{"minimal", "h", "c", "f", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := layoutFull(tt.header, tt.content, tt.footer, tt.height)
			lines := strings.Count(result, "\n") + 1
			if lines != tt.height {
				t.Errorf("layoutFull() produced %d lines, want %d", lines, tt.height)
			}
		})
	}
}

func TestLayoutFull_ZeroHeight(t *testing.T) {
	t.Parallel()
	result := layoutFull("header", "content", "footer", 0)
	if !strings.Contains(result, "header") || !strings.Contains(result, "content") || !strings.Contains(result, "footer") {
		t.Error("zero height should still include all parts")
	}
}

func TestLayoutFull_NegativeHeight(t *testing.T) {
	t.Parallel()
	result := layoutFull("h", "c", "f", -1)
	if !strings.Contains(result, "h") || !strings.Contains(result, "c") || !strings.Contains(result, "f") {
		t.Error("negative height should still include all parts")
	}
}

func TestLayoutFull_ContentExceedsHeight(t *testing.T) {
	t.Parallel()
	bigContent := strings.Repeat("line\n", 30) + "last"
	result := layoutFull("header", bigContent, "footer", 10)
	// Should not truncate — just return as-is
	if !strings.Contains(result, "header") || !strings.Contains(result, "last") || !strings.Contains(result, "footer") {
		t.Error("content exceeding height should not be truncated")
	}
}

func TestLayoutFull_EmptyFooter(t *testing.T) {
	t.Parallel()
	result := layoutFull("header", "content", "", 10)
	lines := strings.Count(result, "\n") + 1
	if lines != 10 {
		t.Errorf("layoutFull with empty footer produced %d lines, want 10", lines)
	}
	if !strings.Contains(result, "header") || !strings.Contains(result, "content") {
		t.Error("should contain header and content")
	}
}

func TestLayoutFull_EmptyHeader(t *testing.T) {
	t.Parallel()
	result := layoutFull("", "content", "footer", 10)
	lines := strings.Count(result, "\n") + 1
	if lines != 10 {
		t.Errorf("layoutFull with empty header produced %d lines, want 10", lines)
	}
}

func TestLayoutFull_PaddingBetweenContentAndFooter(t *testing.T) {
	t.Parallel()
	result := layoutFull("header", "content", "footer", 20)
	// Footer should appear at the end
	footerIdx := strings.LastIndex(result, "footer")
	contentIdx := strings.Index(result, "content")
	if footerIdx <= contentIdx {
		t.Error("footer should appear after content")
	}
}

func TestJoinNonEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{"all non-empty", []string{"a", "b", "c"}, "a\nb\nc"},
		{"some empty", []string{"a", "", "c"}, "a\nc"},
		{"all empty", []string{"", "", ""}, ""},
		{"single", []string{"hello"}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := joinNonEmpty(tt.parts...)
			if got != tt.want {
				t.Errorf("joinNonEmpty(%v) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}
