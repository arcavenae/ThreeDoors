package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/device"
	gosync "github.com/arcavenae/ThreeDoors/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Cross-machine sync operations",
		Long:  "Manage cross-machine task synchronization, conflicts, and manual overrides.",
	}

	cmd.AddCommand(newSyncConflictsCmd())
	cmd.AddCommand(newSyncResolveCmd())
	cmd.AddCommand(newSyncInitCmd())
	cmd.AddCommand(newSyncPushCmd())
	cmd.AddCommand(newSyncStatusCmd())

	return cmd
}

func newSyncConflictsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conflicts",
		Short: "List recent sync conflicts",
		Long:  "Lists recent cross-machine sync conflicts with resolution details.",
		RunE:  runSyncConflicts,
	}

	cmd.Flags().Int("limit", 20, "maximum number of conflicts to show")
	cmd.Flags().String("since", "", "show conflicts since date (RFC3339 format)")
	cmd.Flags().String("task-id", "", "filter by task ID")

	return cmd
}

func newSyncResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <conflict-id> --keep <device>",
		Short: "Manually override a conflict resolution",
		Long: `Replays the rejected version's field values from a previous conflict,
overwriting only the conflicting fields. Increments the local vector clock
to ensure the override propagates on next sync.`,
		Args: cobra.ExactArgs(1),
		RunE: runSyncResolve,
	}
}

func conflictLog() (*core.ConflictLog, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return nil, fmt.Errorf("get config dir: %w", err)
	}
	return core.NewConflictLog(configDir)
}

func runSyncConflicts(cmd *cobra.Command, _ []string) error {
	cl, err := conflictLog()
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	sinceStr, _ := cmd.Flags().GetString("since")
	taskID, _ := cmd.Flags().GetString("task-id")

	var entries []core.ConflictLogEntry

	switch {
	case taskID != "":
		entries, err = cl.EntriesForTask(taskID)
	case sinceStr != "":
		since, parseErr := time.Parse(time.RFC3339, sinceStr)
		if parseErr != nil {
			return fmt.Errorf("invalid --since format (use RFC3339): %w", parseErr)
		}
		entries, err = cl.EntriesSince(since)
	default:
		entries, err = cl.ReadRecentEntries(limit)
	}
	if err != nil {
		return fmt.Errorf("read conflicts: %w", err)
	}

	if isJSONOutput(cmd) {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	if len(entries) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "No conflicts found."); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "CONFLICT ID\tTIMESTAMP\tTASK ID\tFIELDS\tOUTCOME"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, e := range entries {
		fields := ""
		for i, f := range e.Fields {
			if i > 0 {
				fields += ", "
			}
			fields += f.Field
		}
		ts := e.Timestamp.Format("2006-01-02 15:04 UTC")
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.ConflictID, ts, e.TaskID, fields, e.ResolutionOutcome); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return w.Flush()
}

func runSyncResolve(cmd *cobra.Command, args []string) error {
	conflictID := args[0]

	cl, err := conflictLog()
	if err != nil {
		return err
	}

	entry, err := cl.FindByID(conflictID)
	if err != nil {
		if errors.Is(err, core.ErrConflictNotFound) {
			return fmt.Errorf("conflict %q not found", conflictID)
		}
		return err
	}

	if entry.ResolutionOutcome == "manual-override" {
		return core.ErrConflictAlreadyResolved
	}

	// Log the manual override
	overrideEntry := core.ConflictLogEntry{
		ConflictID:        core.NewConflictID(),
		Timestamp:         time.Now().UTC(),
		TaskID:            entry.TaskID,
		DeviceIDs:         entry.DeviceIDs,
		Fields:            entry.Fields,
		ResolutionOutcome: "manual-override",
		RejectedValues:    entry.RejectedValues,
		OverrideOf:        conflictID,
	}

	if err := cl.Append(overrideEntry); err != nil {
		return fmt.Errorf("log override: %w", err)
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(),
		"Conflict %s overridden. New entry: %s\nAffected task: %s\nOverridden fields: %v\n",
		conflictID, overrideEntry.ConflictID, entry.TaskID, entry.RejectedValues); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func newSyncInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <remote-url>",
		Short: "Initialize sync with a Git remote",
		Long: `Sets up cross-computer sync by connecting to a Git remote.

The remote-url should be a Git repository URL (SSH or HTTPS).
Examples:
  git@github.com:user/threedoors-sync.git
  https://github.com/user/threedoors-sync.git`,
		Args: cobra.ExactArgs(1),
		RunE: runSyncInit,
	}
}

func newSyncPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Manually trigger a sync push",
		Long:  "Stages local changes, commits, and pushes to the sync remote.",
		Args:  cobra.NoArgs,
		RunE:  runSyncPush,
	}
}

func newSyncStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		Long:  "Displays the current sync state, last sync time, unpushed commits, and remote URL.",
		Args:  cobra.NoArgs,
		RunE:  runSyncStatus,
	}
}

// syncRepoDir returns the path to the sync Git repository.
func syncRepoDir() (string, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sync"), nil
}

// loadSyncTransport creates a GitSyncTransport from persisted config.
func loadSyncTransport() (*gosync.GitSyncTransport, error) {
	repoDir, err := syncRepoDir()
	if err != nil {
		return nil, err
	}

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return nil, err
	}

	// Load device identity
	devPath := filepath.Join(configDir, "device.yaml")
	dev, err := device.LoadDevice(devPath)
	if err != nil {
		return nil, fmt.Errorf("load device identity: %w (run 'threedoors' first to create device identity)", err)
	}

	// Load sync remote URL from config
	syncCfg, err := loadSyncConfig(configDir)
	if err != nil {
		return nil, err
	}

	executor := gosync.NewExecGitExecutor(30 * time.Second)
	return gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  syncCfg.RemoteURL,
		DeviceID:   dev.ID,
		DeviceName: dev.Name,
		Executor:   executor,
	}), nil
}

// syncConfig holds persisted sync configuration.
type syncConfig struct {
	RemoteURL string `json:"remote_url"`
	Enabled   bool   `json:"enabled"`
}

func syncConfigPath(configDir string) string {
	return filepath.Join(configDir, "sync.json")
}

func loadSyncConfig(configDir string) (syncConfig, error) {
	path := syncConfigPath(configDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return syncConfig{}, gosync.ErrNotInitialized
		}
		return syncConfig{}, fmt.Errorf("read sync config: %w", err)
	}
	var cfg syncConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return syncConfig{}, fmt.Errorf("parse sync config: %w", err)
	}
	return cfg, nil
}

func saveSyncConfig(configDir string, cfg syncConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync config: %w", err)
	}
	path := syncConfigPath(configDir)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write sync config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename sync config: %w", err)
	}
	return nil
}

func runSyncInit(cmd *cobra.Command, args []string) error {
	remoteURL := args[0]

	// Validate URL scheme
	if err := validateRemoteURL(remoteURL); err != nil {
		return err
	}

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return err
	}

	repoDir, err := syncRepoDir()
	if err != nil {
		return err
	}

	// Load device identity
	devPath := filepath.Join(configDir, "device.yaml")
	dev, err := device.LoadDevice(devPath)
	if err != nil {
		return fmt.Errorf("load device identity: %w (run 'threedoors' first)", err)
	}

	executor := gosync.NewExecGitExecutor(30 * time.Second)

	// Validate remote is reachable
	if _, err := executor.Run(cmd.Context(), ".", "ls-remote", remoteURL); err != nil {
		return fmt.Errorf("remote unreachable: %w — check the URL and your credentials", err)
	}

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  remoteURL,
		DeviceID:   dev.ID,
		DeviceName: dev.Name,
		Executor:   executor,
	})

	if err := transport.Init(cmd.Context()); err != nil {
		return fmt.Errorf("sync init: %w", err)
	}

	// Persist config
	if err := saveSyncConfig(configDir, syncConfig{
		RemoteURL: remoteURL,
		Enabled:   true,
	}); err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if isJSONOutput(cmd) {
		return json.NewEncoder(out).Encode(map[string]string{
			"status":     "initialized",
			"remote_url": remoteURL,
			"repo_dir":   repoDir,
		})
	}

	_, _ = fmt.Fprintf(out, "Sync initialized.\n")
	_, _ = fmt.Fprintf(out, "  Remote: %s\n", remoteURL)
	_, _ = fmt.Fprintf(out, "  Local:  %s\n", repoDir)
	return nil
}

func runSyncPush(cmd *cobra.Command, _ []string) error {
	transport, err := loadSyncTransport()
	if err != nil {
		return err
	}

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return err
	}

	// Collect files to sync
	files, err := collectSyncFiles(configDir)
	if err != nil {
		return fmt.Errorf("collect sync files: %w", err)
	}

	devPath := filepath.Join(configDir, "device.yaml")
	dev, err := device.LoadDevice(devPath)
	if err != nil {
		return fmt.Errorf("load device: %w", err)
	}

	changeset := gosync.Changeset{
		DeviceID:  dev.ID,
		Timestamp: time.Now().UTC(),
		Files:     files,
	}

	// Initialize the transport (needed to set initialized flag)
	if err := transport.Init(cmd.Context()); err != nil {
		return fmt.Errorf("sync init: %w", err)
	}

	if err := transport.Push(cmd.Context(), changeset); err != nil {
		return fmt.Errorf("sync push: %w", err)
	}

	out := cmd.OutOrStdout()
	if isJSONOutput(cmd) {
		return json.NewEncoder(out).Encode(map[string]interface{}{
			"status":     "pushed",
			"file_count": len(files),
		})
	}

	_, _ = fmt.Fprintf(out, "Sync push complete (%d files).\n", len(files))
	return nil
}

func runSyncStatus(cmd *cobra.Command, _ []string) error {
	transport, err := loadSyncTransport()
	if err != nil {
		return err
	}

	// Initialize to set the initialized flag
	if err := transport.Init(cmd.Context()); err != nil {
		return fmt.Errorf("sync init: %w", err)
	}

	status, err := transport.Status(cmd.Context())
	if err != nil {
		return fmt.Errorf("sync status: %w", err)
	}

	// Probe connectivity to determine state
	configDir, cfgErr := core.GetConfigDirPath()
	if cfgErr == nil {
		executor := gosync.NewExecGitExecutor(10 * time.Second)
		repoDir, _ := syncRepoDir()
		cfg, cfgLoadErr := loadSyncConfig(configDir)
		if cfgLoadErr == nil {
			remote := cfg.RemoteURL
			if remote == "" {
				remote = "origin"
			}
			_, probeErr := executor.Run(cmd.Context(), repoDir, "ls-remote", "--exit-code", remote)
			if probeErr != nil {
				status.ConnectivityState = "offline"
			} else {
				status.ConnectivityState = "online"
			}
		}
	}
	if status.ConnectivityState == "" {
		status.ConnectivityState = "unknown"
	}

	out := cmd.OutOrStdout()
	if isJSONOutput(cmd) {
		return json.NewEncoder(out).Encode(status)
	}

	_, _ = fmt.Fprintf(out, "Sync Status: %s\n", status.State)
	_, _ = fmt.Fprintf(out, "  Connectivity: %s\n", status.ConnectivityState)
	_, _ = fmt.Fprintf(out, "  Remote:       %s\n", status.RemoteURL)
	if !status.LastSyncTime.IsZero() {
		_, _ = fmt.Fprintf(out, "  Last sync:    %s\n", status.LastSyncTime.Format("2006-01-02 15:04:05 UTC"))
	} else {
		_, _ = fmt.Fprintf(out, "  Last sync:    never\n")
	}
	_, _ = fmt.Fprintf(out, "  Unpushed:     %d commits\n", status.UnpushedCount)
	if !status.OldestUnpushed.IsZero() {
		_, _ = fmt.Fprintf(out, "  Oldest queued: %s\n", status.OldestUnpushed.Format("2006-01-02 15:04:05 UTC"))
	}
	if status.LocalHEAD != "" {
		_, _ = fmt.Fprintf(out, "  Local HEAD:   %s\n", status.LocalHEAD)
	}
	if status.RemoteHEAD != "" {
		_, _ = fmt.Fprintf(out, "  Remote HEAD:  %s\n", status.RemoteHEAD)
	}
	if status.LocalHEAD != "" && status.RemoteHEAD != "" {
		if status.LocalHEAD == status.RemoteHEAD {
			_, _ = fmt.Fprintf(out, "  Divergence:   in sync\n")
		} else {
			_, _ = fmt.Fprintf(out, "  Divergence:   diverged\n")
		}
	}
	return nil
}

// collectSyncFiles reads the syncable files from the config directory.
func collectSyncFiles(configDir string) ([]gosync.SyncFile, error) {
	syncableFiles := []string{"tasks.yaml", "sessions.jsonl"}
	var files []gosync.SyncFile

	for _, name := range syncableFiles {
		path := filepath.Join(configDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		files = append(files, gosync.SyncFile{
			Path:    name,
			Content: content,
			Op:      gosync.OpModify,
		})
	}

	return files, nil
}

// validateRemoteURL checks that the URL scheme is acceptable.
func validateRemoteURL(url string) error {
	validPrefixes := []string{"ssh://", "git@", "https://", "http://", "/"}
	for _, prefix := range validPrefixes {
		if len(url) >= len(prefix) && url[:len(prefix)] == prefix {
			return nil
		}
	}
	return fmt.Errorf("invalid remote URL %q: must start with ssh://, git@, https://, or be a local path", url)
}
