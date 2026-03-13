package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
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
