package cli

import (
	"fmt"
	"os"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// statusResult tracks the outcome of a status change for a single task.
type statusResult struct {
	ID        string `json:"id"`
	ShortID   string `json:"short_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
}

// newTaskBlockCmd creates the "task block" subcommand.
func newTaskBlockCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "block <id>",
		Short: "Block a task with a reason",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reason == "" {
				return fmt.Errorf("--reason is required")
			}
			return runTaskBlock(cmd, args[0], reason)
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "why the task is blocked (required)")

	return cmd
}

func runTaskBlock(cmd *cobra.Command, idPrefix, reason string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task block", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	result := blockOneTask(ctx, idPrefix, reason)

	if isJSONOutput(cmd) {
		_ = formatter.WriteJSON("task block", result, nil)
	} else if result.Success {
		_ = formatter.Writef("Task %s status: %s -> %s\n", result.ShortID, result.OldStatus, result.NewStatus)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
	}

	if result.ExitCode != ExitSuccess {
		os.Exit(result.ExitCode)
	}
	return nil
}

func blockOneTask(ctx *cliContext, idPrefix, reason string) statusResult {
	task, result := resolveTask(ctx, idPrefix)
	if task == nil {
		return *result
	}

	oldStatus := string(task.Status)
	if err := task.UpdateStatus(core.StatusBlocked); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: "blocked",
			Error:     fmt.Sprintf("invalid transition from %q to %q", oldStatus, "blocked"),
			ExitCode:  ExitValidation,
		}
	}

	if err := task.SetBlocker(reason); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: "blocked",
			Error:     fmt.Sprintf("set blocker: %v", err),
			ExitCode:  ExitValidation,
		}
	}

	if err := ctx.provider.SaveTask(task); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: "blocked",
			Error:     fmt.Sprintf("save: %v", err),
			ExitCode:  ExitProviderError,
		}
	}

	return statusResult{
		ID:        task.ID,
		ShortID:   shortID(task.ID),
		OldStatus: oldStatus,
		NewStatus: "blocked",
		Success:   true,
		ExitCode:  ExitSuccess,
	}
}

// newTaskUnblockCmd creates the "task unblock" subcommand.
func newTaskUnblockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblock <id>",
		Short: "Unblock a blocked task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskUnblock(cmd, args[0])
		},
	}
	return cmd
}

func runTaskUnblock(cmd *cobra.Command, idPrefix string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task unblock", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	result := unblockOneTask(ctx, idPrefix)

	if isJSONOutput(cmd) {
		_ = formatter.WriteJSON("task unblock", result, nil)
	} else if result.Success {
		_ = formatter.Writef("Task %s status: %s -> %s\n", result.ShortID, result.OldStatus, result.NewStatus)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
	}

	if result.ExitCode != ExitSuccess {
		os.Exit(result.ExitCode)
	}
	return nil
}

func unblockOneTask(ctx *cliContext, idPrefix string) statusResult {
	task, result := resolveTask(ctx, idPrefix)
	if task == nil {
		return *result
	}

	if task.Status != core.StatusBlocked {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: string(task.Status),
			NewStatus: "todo",
			Error:     fmt.Sprintf("task is not blocked (current status: %q)", task.Status),
			ExitCode:  ExitValidation,
		}
	}

	oldStatus := string(task.Status)
	if err := task.UpdateStatus(core.StatusTodo); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: "todo",
			Error:     fmt.Sprintf("invalid transition from %q to %q", oldStatus, "todo"),
			ExitCode:  ExitValidation,
		}
	}

	if err := ctx.provider.SaveTask(task); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: "todo",
			Error:     fmt.Sprintf("save: %v", err),
			ExitCode:  ExitProviderError,
		}
	}

	return statusResult{
		ID:        task.ID,
		ShortID:   shortID(task.ID),
		OldStatus: oldStatus,
		NewStatus: "todo",
		Success:   true,
		ExitCode:  ExitSuccess,
	}
}

// newTaskStatusCmd creates the "task status" subcommand.
func newTaskStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <id> [id...] <new-status>",
		Short: "Change task status",
		Long:  "Change the status of one or more tasks. The last argument is the target status.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetStatus := args[len(args)-1]
			ids := args[:len(args)-1]
			return runTaskStatus(cmd, ids, targetStatus)
		},
	}
	return cmd
}

func runTaskStatus(cmd *cobra.Command, ids []string, targetStatus string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	if err := core.ValidateStatus(targetStatus); err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task status", ExitValidation, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitValidation)
	}

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task status", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	results := make([]statusResult, 0, len(ids))
	worstExit := ExitSuccess

	for _, idPrefix := range ids {
		result := changeOneTaskStatus(ctx, idPrefix, core.TaskStatus(targetStatus))
		results = append(results, result)
		if result.ExitCode > worstExit {
			worstExit = result.ExitCode
		}
	}

	if isJSONOutput(cmd) {
		_ = formatter.WriteJSON("task status", results, nil)
	} else {
		for _, r := range results {
			if r.Success {
				_ = formatter.Writef("Task %s status: %s -> %s\n", r.ShortID, r.OldStatus, r.NewStatus)
			} else {
				fmt.Fprintf(os.Stderr, "Error changing %s: %s\n", r.ShortID, r.Error)
			}
		}
	}

	if worstExit != ExitSuccess {
		os.Exit(worstExit)
	}
	return nil
}

func changeOneTaskStatus(ctx *cliContext, idPrefix string, targetStatus core.TaskStatus) statusResult {
	task, result := resolveTask(ctx, idPrefix)
	if task == nil {
		return *result
	}

	oldStatus := string(task.Status)
	if err := task.UpdateStatus(targetStatus); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: string(targetStatus),
			Error:     fmt.Sprintf("invalid transition from %q to %q", oldStatus, targetStatus),
			ExitCode:  ExitValidation,
		}
	}

	if err := ctx.provider.SaveTask(task); err != nil {
		return statusResult{
			ID:        task.ID,
			ShortID:   shortID(task.ID),
			OldStatus: oldStatus,
			NewStatus: string(targetStatus),
			Error:     fmt.Sprintf("save: %v", err),
			ExitCode:  ExitProviderError,
		}
	}

	return statusResult{
		ID:        task.ID,
		ShortID:   shortID(task.ID),
		OldStatus: oldStatus,
		NewStatus: string(targetStatus),
		Success:   true,
		ExitCode:  ExitSuccess,
	}
}

// resolveTask finds a single task by prefix. Returns (nil, result) on error.
func resolveTask(ctx *cliContext, idPrefix string) (*core.Task, *statusResult) {
	matches := ctx.pool.FindByPrefix(idPrefix)

	if len(matches) == 0 {
		return nil, &statusResult{
			ID:       idPrefix,
			ShortID:  shortID(idPrefix),
			Error:    "task not found",
			ExitCode: ExitNotFound,
		}
	}

	if len(matches) > 1 {
		return nil, &statusResult{
			ID:       idPrefix,
			ShortID:  shortID(idPrefix),
			Error:    fmt.Sprintf("ambiguous prefix, matches %d tasks", len(matches)),
			ExitCode: ExitAmbiguousInput,
		}
	}

	return matches[0], nil
}
