package cli

import (
	"testing"
)

func TestExitCodes_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		want int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitGeneralError", ExitGeneralError, 1},
		{"ExitNotFound", ExitNotFound, 2},
		{"ExitValidation", ExitValidation, 3},
		{"ExitProviderError", ExitProviderError, 4},
		{"ExitAmbiguousInput", ExitAmbiguousInput, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.code != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestExitCodes_Unique(t *testing.T) {
	t.Parallel()

	codes := []int{ExitSuccess, ExitGeneralError, ExitNotFound, ExitValidation, ExitProviderError, ExitAmbiguousInput}
	seen := make(map[int]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate exit code: %d", code)
		}
		seen[code] = true
	}
}
