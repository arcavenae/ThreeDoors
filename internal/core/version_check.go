package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// versionCheckCacheFile is the name of the version check cache file.
	versionCheckCacheFile = "version-check.json"

	// versionCheckTTL is how long a cached version check remains fresh.
	versionCheckTTL = 24 * time.Hour

	// versionCheckTimeout is the HTTP timeout for GitHub API requests.
	versionCheckTimeout = 5 * time.Second

	// githubReleasesURL is the GitHub Releases API endpoint.
	githubReleasesURL = "https://api.github.com/repos/arcaven/ThreeDoors/releases"
)

// ciEnvVars are environment variables that indicate a CI environment.
var ciEnvVars = []string{
	"CI",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"JENKINS_URL",
	"BUILDKITE",
}

// VersionCheckCache holds cached version check results.
type VersionCheckCache struct {
	CheckedAt      time.Time         `json:"checked_at"`
	LatestVersions map[string]string `json:"latest_versions"`
	Error          string            `json:"error,omitempty"`
}

// IsFresh returns true if the cache is less than 24 hours old.
func (c *VersionCheckCache) IsFresh() bool {
	return time.Since(c.CheckedAt) < versionCheckTTL
}

// ReadVersionCache reads the cached version check from disk.
func ReadVersionCache(configDir string) (*VersionCheckCache, error) {
	path := filepath.Join(configDir, versionCheckCacheFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read version cache: %w", err)
	}

	var cache VersionCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parse version cache: %w", err)
	}

	return &cache, nil
}

// WriteVersionCache writes the version check cache to disk atomically.
func WriteVersionCache(configDir string, cache *VersionCheckCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal version cache: %w", err)
	}

	path := filepath.Join(configDir, versionCheckCacheFile)
	tmpPath := path + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create version cache tmp: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write version cache: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync version cache: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close version cache: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename version cache: %w", err)
	}

	return nil
}

// githubRelease is a minimal representation of a GitHub release.
type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
}

// VersionChecker checks for new versions via the GitHub Releases API.
type VersionChecker struct {
	CurrentVersion string
	Channel        string
	ConfigDir      string
	HTTPClient     *http.Client
	ReleasesURL    string
}

// NewVersionChecker creates a VersionChecker with sensible defaults.
func NewVersionChecker(currentVersion, channel, configDir string) *VersionChecker {
	return &VersionChecker{
		CurrentVersion: currentVersion,
		Channel:        channel,
		ConfigDir:      configDir,
		HTTPClient: &http.Client{
			Timeout: versionCheckTimeout,
		},
		ReleasesURL: githubReleasesURL,
	}
}

// IsUpdateCheckDisabled returns true if the user has opted out of update checks
// via environment variable or CI detection.
func IsUpdateCheckDisabled() bool {
	if os.Getenv("THREEDOORS_NO_UPDATE_CHECK") == "1" {
		return true
	}
	for _, v := range ciEnvVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// IsConfigUpdateCheckDisabled checks the config file for update_check: false.
func IsConfigUpdateCheckDisabled(configDir string) bool {
	configPath := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	// Simple string check — avoids importing yaml just for one field.
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "update_check: false" {
			return true
		}
	}
	return false
}

// FetchLatestVersions fetches releases from GitHub and returns the latest
// version per channel (stable, alpha, beta).
func (vc *VersionChecker) FetchLatestVersions() (map[string]string, error) {
	req, err := http.NewRequest(http.MethodGet, vc.ReleasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := vc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode releases: %w", err)
	}

	latest := make(map[string]string)
	for _, rel := range releases {
		tag := strings.TrimPrefix(rel.TagName, "v")
		ch := classifyChannel(tag)
		existing, ok := latest[ch]
		if !ok || CompareSemver(tag, existing) > 0 {
			latest[ch] = tag
		}
	}

	return latest, nil
}

// Check performs a version check, using cache when fresh.
// Returns the check results for the Version doctor category.
func (vc *VersionChecker) Check() []CheckResult {
	// Dev build — skip check entirely
	if vc.CurrentVersion == "dev" {
		return []CheckResult{{
			Name:    "Current version",
			Status:  CheckInfo,
			Message: "Running dev build",
		}}
	}

	// Opt-out checks
	if IsUpdateCheckDisabled() {
		return []CheckResult{{
			Name:    "Update check",
			Status:  CheckSkip,
			Message: "Update check disabled (env var or CI)",
		}}
	}
	if IsConfigUpdateCheckDisabled(vc.ConfigDir) {
		return []CheckResult{{
			Name:    "Update check",
			Status:  CheckSkip,
			Message: "Update check disabled (config)",
		}}
	}

	// Try cache first
	cache, cacheErr := ReadVersionCache(vc.ConfigDir)
	if cacheErr == nil && cache.IsFresh() && cache.Error == "" {
		return vc.buildResults(cache.LatestVersions)
	}

	// Fetch fresh data
	latest, fetchErr := vc.FetchLatestVersions()
	if fetchErr != nil {
		// If we have stale cache, use it
		if cacheErr == nil && cache.Error == "" {
			return vc.buildResults(cache.LatestVersions)
		}
		return []CheckResult{{
			Name:    "Update check",
			Status:  CheckSkip,
			Message: fmt.Sprintf("Cannot check for updates: %v", fetchErr),
		}}
	}

	// Save to cache
	newCache := &VersionCheckCache{
		CheckedAt:      time.Now().UTC(),
		LatestVersions: latest,
	}
	_ = WriteVersionCache(vc.ConfigDir, newCache)

	return vc.buildResults(latest)
}

// buildResults constructs CheckResult entries from the latest version map.
func (vc *VersionChecker) buildResults(latest map[string]string) []CheckResult {
	var results []CheckResult

	currentChannel := vc.Channel
	if currentChannel == "" {
		currentChannel = "stable"
	}

	// Show current version
	results = append(results, CheckResult{
		Name:    "Current version",
		Status:  CheckOK,
		Message: fmt.Sprintf("v%s (%s)", vc.CurrentVersion, currentChannel),
	})

	currentBase := baseSemver(vc.CurrentVersion)

	// Check same-channel update
	if latestInChannel, ok := latest[currentChannel]; ok {
		latestBase := baseSemver(latestInChannel)
		if CompareSemver(latestBase, currentBase) > 0 {
			suggestion := "Visit https://github.com/arcaven/ThreeDoors/releases"
			if currentChannel == "stable" {
				suggestion = "brew upgrade threedoors"
			}
			results = append(results, CheckResult{
				Name:       "Update available",
				Status:     CheckInfo,
				Message:    fmt.Sprintf("Update available: v%s", latestInChannel),
				Suggestion: suggestion,
			})
		}
	}

	// Cross-channel: alpha/beta users see stable if stable base is higher
	if currentChannel != "stable" {
		if latestStable, ok := latest["stable"]; ok {
			stableBase := baseSemver(latestStable)
			if CompareSemver(stableBase, currentBase) > 0 {
				results = append(results, CheckResult{
					Name:    "Newer stable release",
					Status:  CheckInfo,
					Message: fmt.Sprintf("Stable release v%s is available (newer than your %s base)", latestStable, currentChannel),
				})
			}
		}
	}

	return results
}

// classifyChannel determines the release channel from a version string.
// Versions with "-alpha" are alpha, "-beta" are beta, everything else is stable.
func classifyChannel(version string) string {
	if strings.Contains(version, "-alpha") {
		return "alpha"
	}
	if strings.Contains(version, "-beta") {
		return "beta"
	}
	return "stable"
}

// baseSemver extracts the major.minor.patch portion from a version string,
// stripping any pre-release suffix (e.g., "1.2.0-alpha" → "1.2.0").
func baseSemver(version string) string {
	version = strings.TrimPrefix(version, "v")
	if idx := strings.IndexByte(version, '-'); idx != -1 {
		return version[:idx]
	}
	return version
}

// CompareSemver compares two semver strings (major.minor.patch).
// Returns -1 if a < b, 0 if equal, 1 if a > b.
// Non-parseable versions compare as 0.0.0.
func CompareSemver(a, b string) int {
	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

// parseSemverParts extracts [major, minor, patch] from a version string.
func parseSemverParts(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	// Strip pre-release suffix
	if idx := strings.IndexByte(v, '-'); idx != -1 {
		v = v[:idx]
	}

	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err == nil {
			result[i] = n
		}
	}
	return result
}
