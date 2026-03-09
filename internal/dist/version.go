package dist

import "fmt"

// VersionPrefix is the expected prefix in version output.
const VersionPrefix = "ThreeDoors"

// FormatVersion returns the formatted version string.
// Format: "ThreeDoors <version>" (e.g., "ThreeDoors 0.1.0")
func FormatVersion(version string) string {
	return fmt.Sprintf("%s %s", VersionPrefix, version)
}

// FormatVersionWithChannel returns the formatted version string with an
// optional channel label. When channel is non-empty, the output includes
// the channel in parentheses: "ThreeDoors (alpha) v0.1.0-alpha.20260308.abc1234".
// When channel is empty (stable builds), output is unchanged: "ThreeDoors v1.0.0".
func FormatVersionWithChannel(version, channel string) string {
	if channel != "" {
		return fmt.Sprintf("%s (%s) v%s", VersionPrefix, channel, version)
	}
	return fmt.Sprintf("%s %s", VersionPrefix, version)
}
