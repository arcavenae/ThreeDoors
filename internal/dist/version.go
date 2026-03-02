package dist

import "fmt"

// VersionPrefix is the expected prefix in version output.
const VersionPrefix = "ThreeDoors"

// FormatVersion returns the formatted version string.
// Format: "ThreeDoors <version>" (e.g., "ThreeDoors 0.1.0")
func FormatVersion(version string) string {
	return fmt.Sprintf("%s %s", VersionPrefix, version)
}
