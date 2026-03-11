package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// clearCIEnvVars unsets CI environment variables so version check tests
// actually exercise the check logic instead of short-circuiting.
// Must be called BEFORE t.Parallel() since t.Setenv doesn't work with parallel.
func clearCIEnvVars(t *testing.T) {
	t.Helper()
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
	}
	t.Setenv("THREEDOORS_NO_UPDATE_CHECK", "")
}

func TestCompareSemver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"major greater", "2.0.0", "1.0.0", 1},
		{"major less", "1.0.0", "2.0.0", -1},
		{"minor greater", "1.2.0", "1.1.0", 1},
		{"minor less", "1.1.0", "1.2.0", -1},
		{"patch greater", "1.0.2", "1.0.1", 1},
		{"patch less", "1.0.1", "1.0.2", -1},
		{"with v prefix", "v1.2.0", "v1.1.0", 1},
		{"mixed v prefix", "v1.2.0", "1.1.0", 1},
		{"with pre-release suffix", "1.2.0-alpha", "1.1.0", 1},
		{"pre-release same base", "1.2.0-alpha", "1.2.0", 0},
		{"empty string", "", "1.0.0", -1},
		{"both empty", "", "", 0},
		{"partial version", "1.2", "1.2.0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CompareSemver(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestBaseSemver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"1.2.0", "1.2.0"},
		{"1.2.0-alpha", "1.2.0"},
		{"1.2.0-alpha.20260308.abc1234", "1.2.0"},
		{"v1.2.0", "1.2.0"},
		{"v1.2.0-beta", "1.2.0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := baseSemver(tt.input)
			if got != tt.want {
				t.Errorf("baseSemver(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClassifyChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    string
	}{
		{"1.0.0", "stable"},
		{"1.0.0-alpha", "alpha"},
		{"1.0.0-alpha.20260308", "alpha"},
		{"1.0.0-beta", "beta"},
		{"1.0.0-beta.1", "beta"},
		{"2.0.0-rc.1", "stable"}, // rc is not alpha or beta
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			got := classifyChannel(tt.version)
			if got != tt.want {
				t.Errorf("classifyChannel(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestVersionCheckCache_IsFresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkedAt time.Time
		want      bool
	}{
		{"just now", time.Now().UTC(), true},
		{"1 hour ago", time.Now().UTC().Add(-1 * time.Hour), true},
		{"23 hours ago", time.Now().UTC().Add(-23 * time.Hour), true},
		{"25 hours ago", time.Now().UTC().Add(-25 * time.Hour), false},
		{"48 hours ago", time.Now().UTC().Add(-48 * time.Hour), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := &VersionCheckCache{CheckedAt: tt.checkedAt}
			if got := cache.IsFresh(); got != tt.want {
				t.Errorf("IsFresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadWriteVersionCache(t *testing.T) {
	tmpDir := t.TempDir()

	original := &VersionCheckCache{
		CheckedAt: time.Now().UTC().Truncate(time.Second),
		LatestVersions: map[string]string{
			"stable": "1.3.0",
			"alpha":  "1.4.0-alpha.1",
		},
	}

	if err := WriteVersionCache(tmpDir, original); err != nil {
		t.Fatalf("WriteVersionCache: %v", err)
	}

	got, err := ReadVersionCache(tmpDir)
	if err != nil {
		t.Fatalf("ReadVersionCache: %v", err)
	}

	if !got.CheckedAt.Equal(original.CheckedAt) {
		t.Errorf("CheckedAt = %v, want %v", got.CheckedAt, original.CheckedAt)
	}
	if got.LatestVersions["stable"] != "1.3.0" {
		t.Errorf("stable = %q, want %q", got.LatestVersions["stable"], "1.3.0")
	}
	if got.LatestVersions["alpha"] != "1.4.0-alpha.1" {
		t.Errorf("alpha = %q, want %q", got.LatestVersions["alpha"], "1.4.0-alpha.1")
	}
}

func TestReadVersionCache_NotFound(t *testing.T) {
	_, err := ReadVersionCache(t.TempDir())
	if err == nil {
		t.Error("expected error for missing cache, got nil")
	}
}

func TestReadVersionCache_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, versionCheckCacheFile), []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ReadVersionCache(tmpDir)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestVersionChecker_FetchLatestVersions(t *testing.T) {
	t.Parallel()

	releases := []githubRelease{
		{TagName: "v1.3.0", Prerelease: false},
		{TagName: "v1.2.0", Prerelease: false},
		{TagName: "v1.4.0-alpha.1", Prerelease: true},
		{TagName: "v1.3.1-beta.1", Prerelease: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		Channel:        "",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	latest, err := vc.FetchLatestVersions()
	if err != nil {
		t.Fatalf("FetchLatestVersions: %v", err)
	}

	if latest["stable"] != "1.3.0" {
		t.Errorf("stable = %q, want %q", latest["stable"], "1.3.0")
	}
	if latest["alpha"] != "1.4.0-alpha.1" {
		t.Errorf("alpha = %q, want %q", latest["alpha"], "1.4.0-alpha.1")
	}
	if latest["beta"] != "1.3.1-beta.1" {
		t.Errorf("beta = %q, want %q", latest["beta"], "1.3.1-beta.1")
	}
}

func TestVersionChecker_FetchLatestVersions_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	_, err := vc.FetchLatestVersions()
	if err == nil {
		t.Error("expected error for 403 response, got nil")
	}
}

func TestVersionChecker_Check_DevBuild(t *testing.T) {
	t.Parallel()

	vc := &VersionChecker{
		CurrentVersion: "dev",
		ConfigDir:      t.TempDir(),
	}

	results := vc.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CheckInfo {
		t.Errorf("status = %v, want %v", results[0].Status, CheckInfo)
	}
	if results[0].Message != "Running dev build" {
		t.Errorf("message = %q, want %q", results[0].Message, "Running dev build")
	}
}

func TestVersionChecker_Check_StableUpdateAvailable(t *testing.T) {
	clearCIEnvVars(t)

	releases := []githubRelease{
		{TagName: "v1.3.0"},
		{TagName: "v1.1.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.1.0",
		Channel:        "",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()

	// Should have: current version + update available
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d: %+v", len(results), results)
	}

	// First: current version OK
	if results[0].Status != CheckOK {
		t.Errorf("current version status = %v, want %v", results[0].Status, CheckOK)
	}

	// Second: update available
	if results[1].Status != CheckInfo {
		t.Errorf("update status = %v, want %v", results[1].Status, CheckInfo)
	}
	if results[1].Message != "Update available: v1.3.0" {
		t.Errorf("update message = %q, want %q", results[1].Message, "Update available: v1.3.0")
	}
	if results[1].Suggestion != "brew upgrade threedoors" {
		t.Errorf("suggestion = %q, want %q", results[1].Suggestion, "brew upgrade threedoors")
	}
}

func TestVersionChecker_Check_AlphaSeesNewerStable(t *testing.T) {
	clearCIEnvVars(t)

	releases := []githubRelease{
		{TagName: "v1.3.0"},
		{TagName: "v1.4.0-alpha.1", Prerelease: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.2.0-alpha",
		Channel:        "alpha",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()

	// Should have: current version + alpha update + newer stable
	if len(results) < 3 {
		t.Fatalf("expected at least 3 results, got %d: %+v", len(results), results)
	}

	// Find the newer stable result
	var foundStable bool
	for _, r := range results {
		if r.Name == "Newer stable release" {
			foundStable = true
			if r.Status != CheckInfo {
				t.Errorf("stable notice status = %v, want %v", r.Status, CheckInfo)
			}
		}
	}
	if !foundStable {
		t.Error("expected 'Newer stable release' result, not found")
	}
}

func TestVersionChecker_Check_AlphaIgnoresOlderStable(t *testing.T) {
	clearCIEnvVars(t)

	releases := []githubRelease{
		{TagName: "v1.1.0"},
		{TagName: "v1.2.0-alpha", Prerelease: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.2.0-alpha",
		Channel:        "alpha",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()

	// Should NOT have a "Newer stable release" entry
	for _, r := range results {
		if r.Name == "Newer stable release" {
			t.Errorf("should not suggest older stable v1.1.0 to alpha v1.2.0 user")
		}
	}
}

func TestVersionChecker_Check_UpToDate(t *testing.T) {
	clearCIEnvVars(t)

	releases := []githubRelease{
		{TagName: "v1.3.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.3.0",
		Channel:        "",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()

	// Only current version — no update available
	if len(results) != 1 {
		t.Fatalf("expected 1 result (up to date), got %d: %+v", len(results), results)
	}
	if results[0].Status != CheckOK {
		t.Errorf("status = %v, want %v", results[0].Status, CheckOK)
	}
}

func TestVersionChecker_Check_UsesFreshCache(t *testing.T) {
	clearCIEnvVars(t)
	tmpDir := t.TempDir()

	// Write a fresh cache
	cache := &VersionCheckCache{
		CheckedAt: time.Now().UTC(),
		LatestVersions: map[string]string{
			"stable": "2.0.0",
		},
	}
	if err := WriteVersionCache(tmpDir, cache); err != nil {
		t.Fatal(err)
	}

	// Server should NOT be called (cache is fresh)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("server was called despite fresh cache")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		Channel:        "",
		ConfigDir:      tmpDir,
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results (cached), got %d", len(results))
	}
}

func TestVersionChecker_Check_StaleCache_FetchesFresh(t *testing.T) {
	clearCIEnvVars(t)
	tmpDir := t.TempDir()

	// Write a stale cache
	cache := &VersionCheckCache{
		CheckedAt: time.Now().UTC().Add(-25 * time.Hour),
		LatestVersions: map[string]string{
			"stable": "1.2.0",
		},
	}
	if err := WriteVersionCache(tmpDir, cache); err != nil {
		t.Fatal(err)
	}

	// Server returns newer version
	called := false
	releases := []githubRelease{{TagName: "v1.5.0"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		Channel:        "",
		ConfigDir:      tmpDir,
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()
	if !called {
		t.Error("expected server to be called for stale cache")
	}

	// Should report v1.5.0 (from fresh fetch, not stale cache v1.2.0)
	var found bool
	for _, r := range results {
		if r.Name == "Update available" && r.Message == "Update available: v1.5.0" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected update to v1.5.0 from fresh fetch, got: %+v", results)
	}
}

func TestVersionChecker_Check_NetworkError_UsesStaleCache(t *testing.T) {
	clearCIEnvVars(t)
	tmpDir := t.TempDir()

	// Write a stale cache with valid data
	cache := &VersionCheckCache{
		CheckedAt: time.Now().UTC().Add(-25 * time.Hour),
		LatestVersions: map[string]string{
			"stable": "1.5.0",
		},
	}
	if err := WriteVersionCache(tmpDir, cache); err != nil {
		t.Fatal(err)
	}

	// Server is unreachable
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		Channel:        "",
		ConfigDir:      tmpDir,
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()

	// Should fall back to stale cache
	var found bool
	for _, r := range results {
		if r.Name == "Update available" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected update from stale cache fallback, got: %+v", results)
	}
}

func TestVersionChecker_Check_NetworkError_NoCache(t *testing.T) {
	clearCIEnvVars(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		Channel:        "",
		ConfigDir:      t.TempDir(),
		HTTPClient:     server.Client(),
		ReleasesURL:    server.URL,
	}

	results := vc.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result (skip), got %d", len(results))
	}
	if results[0].Status != CheckSkip {
		t.Errorf("status = %v, want %v", results[0].Status, CheckSkip)
	}
}

func TestIsUpdateCheckDisabled(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"no env vars", nil, false},
		{"THREEDOORS_NO_UPDATE_CHECK=1", map[string]string{"THREEDOORS_NO_UPDATE_CHECK": "1"}, true},
		{"CI=true", map[string]string{"CI": "true"}, true},
		{"GITHUB_ACTIONS=true", map[string]string{"GITHUB_ACTIONS": "true"}, true},
		{"GITLAB_CI=true", map[string]string{"GITLAB_CI": "true"}, true},
		{"JENKINS_URL set", map[string]string{"JENKINS_URL": "http://jenkins"}, true},
		{"BUILDKITE set", map[string]string{"BUILDKITE": "true"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars (empty string = not set for our check)
			for _, v := range ciEnvVars {
				t.Setenv(v, "")
			}
			t.Setenv("THREEDOORS_NO_UPDATE_CHECK", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got := IsUpdateCheckDisabled()
			if got != tt.want {
				t.Errorf("IsUpdateCheckDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConfigUpdateCheckDisabled(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   bool
	}{
		{"no config file", "", false},
		{"update_check true", "update_check: true\n", false},
		{"update_check false", "update_check: false\n", true},
		{"update_check false with other fields", "schema_version: 2\nupdate_check: false\nprovider: textfile\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.config != "" {
				if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(tt.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := IsConfigUpdateCheckDisabled(tmpDir)
			if got != tt.want {
				t.Errorf("IsConfigUpdateCheckDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionChecker_Check_OptOutEnvVar(t *testing.T) {
	t.Setenv("THREEDOORS_NO_UPDATE_CHECK", "1")

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		ConfigDir:      t.TempDir(),
	}

	results := vc.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CheckSkip {
		t.Errorf("status = %v, want %v", results[0].Status, CheckSkip)
	}
}

func TestVersionChecker_Check_OptOutConfig(t *testing.T) {
	// Clear env vars that might interfere
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
	}
	t.Setenv("THREEDOORS_NO_UPDATE_CHECK", "")

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("update_check: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		ConfigDir:      tmpDir,
	}

	results := vc.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CheckSkip {
		t.Errorf("status = %v, want %v", results[0].Status, CheckSkip)
	}
}

func TestVersionChecker_Check_CIDetection(t *testing.T) {
	t.Setenv("CI", "true")

	vc := &VersionChecker{
		CurrentVersion: "1.0.0",
		ConfigDir:      t.TempDir(),
	}

	results := vc.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CheckSkip {
		t.Errorf("status = %v, want %v", results[0].Status, CheckSkip)
	}
}

func TestBuildResults_CrossChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentVersion string
		channel        string
		latest         map[string]string
		wantStable     bool
	}{
		{
			name:           "alpha sees newer stable",
			currentVersion: "1.2.0-alpha",
			channel:        "alpha",
			latest:         map[string]string{"stable": "1.3.0", "alpha": "1.2.0-alpha"},
			wantStable:     true,
		},
		{
			name:           "alpha ignores older stable",
			currentVersion: "1.2.0-alpha",
			channel:        "alpha",
			latest:         map[string]string{"stable": "1.1.0", "alpha": "1.2.0-alpha"},
			wantStable:     false,
		},
		{
			name:           "alpha ignores same stable",
			currentVersion: "1.2.0-alpha",
			channel:        "alpha",
			latest:         map[string]string{"stable": "1.2.0", "alpha": "1.2.0-alpha"},
			wantStable:     false,
		},
		{
			name:           "stable never shows cross-channel",
			currentVersion: "1.2.0",
			channel:        "",
			latest:         map[string]string{"stable": "1.2.0", "alpha": "1.5.0-alpha"},
			wantStable:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vc := &VersionChecker{
				CurrentVersion: tt.currentVersion,
				Channel:        tt.channel,
			}
			results := vc.buildResults(tt.latest)

			var foundStable bool
			for _, r := range results {
				if r.Name == "Newer stable release" {
					foundStable = true
				}
			}
			if foundStable != tt.wantStable {
				t.Errorf("found 'Newer stable release' = %v, want %v; results: %+v", foundStable, tt.wantStable, results)
			}
		})
	}
}

func TestWriteVersionCache_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()

	cache := &VersionCheckCache{
		CheckedAt:      time.Now().UTC(),
		LatestVersions: map[string]string{"stable": "1.0.0"},
	}

	if err := WriteVersionCache(tmpDir, cache); err != nil {
		t.Fatal(err)
	}

	// Verify no .tmp file remains
	tmpFile := filepath.Join(tmpDir, versionCheckCacheFile+".tmp")
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("tmp file should not exist after successful write")
	}

	// Verify cache file exists and is valid JSON
	data, err := os.ReadFile(filepath.Join(tmpDir, versionCheckCacheFile))
	if err != nil {
		t.Fatalf("read cache file: %v", err)
	}

	var readBack VersionCheckCache
	if err := json.Unmarshal(data, &readBack); err != nil {
		t.Fatalf("unmarshal cache: %v", err)
	}
}

func TestVersionChecker_FetchLatestVersions_PicksHighest(t *testing.T) {
	t.Parallel()

	// Multiple releases per channel — should pick highest
	releases := []githubRelease{
		{TagName: "v1.1.0"},
		{TagName: "v1.3.0"},
		{TagName: "v1.2.0"},
		{TagName: "v0.9.0-alpha"},
		{TagName: "v1.0.0-alpha.2"},
		{TagName: "v1.0.0-alpha.1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	vc := &VersionChecker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
	}

	latest, err := vc.FetchLatestVersions()
	if err != nil {
		t.Fatal(err)
	}

	if latest["stable"] != "1.3.0" {
		t.Errorf("stable = %q, want %q", latest["stable"], "1.3.0")
	}
	if latest["alpha"] != "1.0.0-alpha.2" {
		t.Errorf("alpha = %q, want %q", latest["alpha"], "1.0.0-alpha.2")
	}
}

func TestDoctorChecker_VersionCategory(t *testing.T) {
	// Create a mock server
	releases := []githubRelease{
		{TagName: "v1.5.0"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(server.Close)

	// Clear env vars
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
	}
	t.Setenv("THREEDOORS_NO_UPDATE_CHECK", "")

	tmpDir := t.TempDir()

	// Write config so Environment category passes
	configContent := fmt.Sprintf("schema_version: %d\nprovider: textfile\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	dc.SetVersionInfo("1.0.0", "", server.Client(), server.URL)
	result := dc.Run()

	// Should have Environment + Version categories
	if len(result.Categories) < 2 {
		t.Fatalf("expected at least 2 categories, got %d", len(result.Categories))
	}

	var versionCat *CategoryResult
	for i := range result.Categories {
		if result.Categories[i].Name == "Version" {
			versionCat = &result.Categories[i]
			break
		}
	}
	if versionCat == nil {
		t.Fatal("Version category not found")
	}

	// Should report update available (1.0.0 → 1.5.0)
	var foundUpdate bool
	for _, check := range versionCat.Checks {
		if check.Name == "Update available" {
			foundUpdate = true
			if check.Status != CheckInfo {
				t.Errorf("update status = %v, want %v", check.Status, CheckInfo)
			}
		}
	}
	if !foundUpdate {
		t.Errorf("expected 'Update available' check, got: %+v", versionCat.Checks)
	}
}
