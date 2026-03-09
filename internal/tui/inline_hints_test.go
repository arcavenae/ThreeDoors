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
		fade    bool
		want    string // substring expected; empty means result should be empty
	}{
		{
			name:    "enabled normal returns styled key",
			key:     "a",
			enabled: true,
			fade:    false,
			want:    "[a]",
		},
		{
			name:    "enabled fade returns styled key",
			key:     "d",
			enabled: true,
			fade:    true,
			want:    "[d]",
		},
		{
			name:    "disabled returns empty",
			key:     "a",
			enabled: false,
			fade:    false,
			want:    "",
		},
		{
			name:    "disabled with fade returns empty",
			key:     "a",
			enabled: false,
			fade:    true,
			want:    "",
		},
		{
			name:    "multi-char key",
			key:     "Enter",
			enabled: true,
			fade:    false,
			want:    "[Enter]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := renderInlineHint(tt.key, tt.enabled, tt.fade)

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

func TestRenderInlineHintFadeVsNormalDiffer(t *testing.T) {
	t.Parallel()

	normal := renderInlineHint("a", true, false)
	fade := renderInlineHint("a", true, true)

	if normal == fade {
		t.Error("expected fade and normal styles to produce different output")
	}

	// Both should contain the key text
	if !strings.Contains(normal, "[a]") {
		t.Errorf("normal hint missing key text, got %q", normal)
	}
	if !strings.Contains(fade, "[a]") {
		t.Errorf("fade hint missing key text, got %q", fade)
	}
}
