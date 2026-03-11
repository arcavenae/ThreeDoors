package core

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// ErrSymlink is returned when a path is a symbolic link.
var ErrSymlink = fmt.Errorf("path is a symbolic link")

// ErrOwnerMismatch is returned when a directory's owner does not match the current user.
var ErrOwnerMismatch = fmt.Errorf("directory owner does not match current user")

// ValidatePath checks that an existing file at path is not a symbolic link.
// If the path does not exist, it returns nil (no validation needed for new files).
// If the path exists and is a symlink, it returns ErrSymlink.
func ValidatePath(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("lstat %s: %w", path, err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s: %w (refusing to follow symlink for security)", path, ErrSymlink)
	}

	return nil
}

// ValidateDir checks that a directory is not a symbolic link and that its
// owner matches the current user. Returns ErrSymlink if the directory is a
// symlink. Returns ErrOwnerMismatch (as a warning-level error) if ownership
// does not match on supported platforms.
func ValidateDir(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("lstat %s: %w", path, err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s: %w (refusing to use symlinked data directory)", path, ErrSymlink)
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s: expected directory, got file", path)
	}

	if err := checkOwnership(fi, path); err != nil {
		return err
	}

	return nil
}

// checkOwnership verifies that the file/directory owner matches the current user.
// Only supported on darwin and linux; returns nil on other platforms.
func checkOwnership(fi os.FileInfo, path string) error {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil
	}

	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}

	currentUID := uint32(os.Getuid())
	if stat.Uid != currentUID {
		return fmt.Errorf("%s: %w (owned by uid %d, running as uid %d)",
			path, ErrOwnerMismatch, stat.Uid, currentUID)
	}

	return nil
}
