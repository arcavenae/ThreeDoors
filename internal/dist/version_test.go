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

func TestFormatVersionWithChannel(t *testing.T) {
	tests := []struct {
		name    string
		version string
		channel string
		want    string
	}{
		{
			name:    "alpha channel with date version",
			version: "0.1.0-alpha.20260308.abc1234",
			channel: "alpha",
			want:    "ThreeDoors (alpha) v0.1.0-alpha.20260308.abc1234",
		},
		{
			name:    "empty channel (stable)",
			version: "1.0.0",
			channel: "",
			want:    "ThreeDoors 1.0.0",
		},
		{
			name:    "dev version no channel",
			version: "dev",
			channel: "",
			want:    "ThreeDoors dev",
		},
		{
			name:    "beta channel",
			version: "0.2.0-beta.1",
			channel: "beta",
			want:    "ThreeDoors (beta) v0.2.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatVersionWithChannel(tt.version, tt.channel)
			if got != tt.want {
				t.Errorf("FormatVersionWithChannel(%q, %q) = %q, want %q", tt.version, tt.channel, got, tt.want)
			}
		})
	}
}

func TestFormatVersionWithChannel_StartsWithPrefix(t *testing.T) {
	cases := []struct {
		version string
		channel string
	}{
		{"0.1.0-alpha.20260308.abc1234", "alpha"},
		{"1.0.0", ""},
		{"dev", ""},
	}

	for _, c := range cases {
		result := FormatVersionWithChannel(c.version, c.channel)
		if !strings.HasPrefix(result, VersionPrefix) {
			t.Errorf("FormatVersionWithChannel(%q, %q) = %q, does not start with %q",
				c.version, c.channel, result, VersionPrefix)
		}
	}
}

func TestFormatVersionWithChannel_AlphaContainsChannelLabel(t *testing.T) {
	result := FormatVersionWithChannel("0.1.0-alpha.20260308.abc1234", "alpha")
	if !strings.Contains(result, "(alpha)") {
		t.Errorf("alpha version output %q does not contain '(alpha)'", result)
	}
}

func TestFormatVersionWithChannel_StableNoParens(t *testing.T) {
	result := FormatVersionWithChannel("1.0.0", "")
	if strings.Contains(result, "(") || strings.Contains(result, ")") {
		t.Errorf("stable version output %q should not contain parentheses", result)
	}
}
