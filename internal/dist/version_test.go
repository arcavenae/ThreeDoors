package dist

import (
	"strings"
	"testing"
)

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "dev version", version: "dev", want: "ThreeDoors dev"},
		{name: "semver", version: "0.1.0", want: "ThreeDoors 0.1.0"},
		{name: "alpha with date", version: "0.1.0-alpha.20260302.d5e99a1", want: "ThreeDoors 0.1.0-alpha.20260302.d5e99a1"},
		{name: "major release", version: "1.0.0", want: "ThreeDoors 1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatVersion(tt.version)
			if got != tt.want {
				t.Errorf("FormatVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestFormatVersion_StartsWithPrefix(t *testing.T) {
	// The Homebrew test block uses: assert_match "ThreeDoors"
	// This test ensures all version strings start with "ThreeDoors"
	versions := []string{"dev", "0.1.0", "1.0.0-beta.1", "2.3.4"}

	for _, v := range versions {
		result := FormatVersion(v)
		if !strings.HasPrefix(result, VersionPrefix) {
			t.Errorf("FormatVersion(%q) = %q, does not start with %q", v, result, VersionPrefix)
		}
	}
}

func TestVersionPrefix_IsThreeDoors(t *testing.T) {
	// Ensure the constant matches what Homebrew formula expects
	if VersionPrefix != "ThreeDoors" {
		t.Errorf("VersionPrefix = %q, want %q", VersionPrefix, "ThreeDoors")
	}
}
