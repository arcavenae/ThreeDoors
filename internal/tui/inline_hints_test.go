package tui

import (
	"strings"
	"testing"
)

func TestRenderInlineHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		enabled bool
		want    string // substring expected; empty means result should be empty
	}{
		{
			name:    "enabled returns styled key",
			key:     "a",
			enabled: true,
			want:    "[a]",
		},
		{
			name:    "disabled returns empty",
			key:     "a",
			enabled: false,
			want:    "",
		},
		{
			name:    "multi-char key",
			key:     "Enter",
			enabled: true,
			want:    "[Enter]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := renderInlineHint(tt.key, tt.enabled)

			if tt.want == "" {
				if got != "" {
					t.Errorf("expected empty string, got %q", got)
				}
				return
			}

			if !strings.Contains(got, tt.want) {
				t.Errorf("expected output to contain %q, got %q", tt.want, got)
			}
		})
	}
}
