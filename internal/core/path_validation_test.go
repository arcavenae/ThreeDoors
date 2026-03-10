package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath_RegularFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "regular.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ValidatePath(path); err != nil {
		t.Errorf("ValidatePath(regular file) = %v, want nil", err)
	}
}

func TestValidatePath_NonExistent(t *testing.T) {
	t.Parallel()
	if err := ValidatePath("/nonexistent/path/file.txt"); err != nil {
		t.Errorf("ValidatePath(nonexistent) = %v, want nil", err)
	}
}

func TestValidatePath_Symlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	err := ValidatePath(link)
	if err == nil {
		t.Fatal("ValidatePath(symlink) = nil, want error")
	}
	if !errors.Is(err, ErrSymlink) {
		t.Errorf("ValidatePath(symlink) error = %v, want ErrSymlink", err)
	}
}

func TestValidateDir_RegularDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := ValidateDir(dir); err != nil {
		t.Errorf("ValidateDir(regular dir) = %v, want nil", err)
	}
}

func TestValidateDir_NonExistent(t *testing.T) {
	t.Parallel()
	if err := ValidateDir("/nonexistent/path/dir"); err != nil {
		t.Errorf("ValidateDir(nonexistent) = %v, want nil", err)
	}
}

func TestValidateDir_Symlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o700); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, link); err != nil {
		t.Fatal(err)
	}

	err := ValidateDir(link)
	if err == nil {
		t.Fatal("ValidateDir(symlink) = nil, want error")
	}
	if !errors.Is(err, ErrSymlink) {
		t.Errorf("ValidateDir(symlink) error = %v, want ErrSymlink", err)
	}
}

func TestValidateDir_NotADir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := ValidateDir(file)
	if err == nil {
		t.Fatal("ValidateDir(file) = nil, want error")
	}
}

func TestValidateDir_OwnershipCurrentUser(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := ValidateDir(dir); err != nil {
		t.Errorf("ValidateDir(own dir) = %v, want nil (current user owns temp dir)", err)
	}
}

func TestValidatePath_SymlinkToNonExistent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	link := filepath.Join(dir, "dangling")
	if err := os.Symlink("/nonexistent/target", link); err != nil {
		t.Fatal(err)
	}

	err := ValidatePath(link)
	if err == nil {
		t.Fatal("ValidatePath(dangling symlink) = nil, want error")
	}
	if !errors.Is(err, ErrSymlink) {
		t.Errorf("ValidatePath(dangling symlink) error = %v, want ErrSymlink", err)
	}
}
