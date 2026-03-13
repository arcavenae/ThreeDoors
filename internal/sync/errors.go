package sync

import "errors"

// ErrGitNotFound is returned when the git binary is not in PATH.
var ErrGitNotFound = errors.New("git binary not found in PATH")

// ErrRemoteUnreachable is returned when the Git remote cannot be contacted.
var ErrRemoteUnreachable = errors.New("git remote unreachable")

// ErrNotInitialized is returned when sync operations are attempted before init.
var ErrNotInitialized = errors.New("sync not initialized — run 'threedoors sync init' first")

// ErrSyncInProgress is returned when a sync is already running.
var ErrSyncInProgress = errors.New("sync already in progress")

// ErrRebaseConflict is returned when git pull --rebase encounters conflicts.
var ErrRebaseConflict = errors.New("rebase conflict during sync")
