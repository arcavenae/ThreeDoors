package sync_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createBareRepo creates a bare Git repository in a temp directory.
func createBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bareDir := filepath.Join(dir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", bareDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %v\n%s", err, out)
	}
	return bareDir
}

// cloneRepo clones a bare repo into a working directory.
func cloneRepo(t *testing.T, bareDir string) string {
	t.Helper()
	dir := t.TempDir()
	workDir := filepath.Join(dir, "work")
	cmd := exec.Command("git", "clone", bareDir, workDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\n%s", err, out)
	}
	// Configure user for commits
	configGitUser(t, workDir)
	return workDir
}

// commitFile writes a file, stages it, and commits.
func commitFile(t *testing.T, repoDir, filename, content string) {
	t.Helper()
	path := filepath.Join(repoDir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	runGit(t, repoDir, "add", filename)
	runGit(t, repoDir, "commit", "-m", "test: "+filename)
}

// pushRepo pushes to origin.
func pushRepo(t *testing.T, repoDir string) {
	t.Helper()
	runGit(t, repoDir, "push")
}

// configGitUser sets git user.name and user.email for commits.
func configGitUser(t *testing.T, repoDir string) {
	t.Helper()
	runGit(t, repoDir, "config", "user.name", "Test User")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
}

// runGit runs a git command in the given directory.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// initRepoWithFirstCommit creates a bare repo and a clone with an initial commit.
// Returns (bareDir, workDir).
func initRepoWithFirstCommit(t *testing.T) (string, string) {
	t.Helper()
	bareDir := createBareRepo(t)
	workDir := cloneRepo(t, bareDir)
	// Need an initial commit so we have a main branch
	commitFile(t, workDir, "README.md", "# sync test repo")
	pushRepo(t, workDir)
	return bareDir, workDir
}
