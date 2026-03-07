package cli

import (
	"fmt"
	"os"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// newTaskCmd creates the "task" command group.
func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}
	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskCompleteCmd())
	return cmd
}

// newTaskAddCmd creates the "task add" subcommand.
func newTaskAddCmd() *cobra.Command {
	var (
		context  string
		taskType string
		effort   string
	)

	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskAdd(cmd, args[0], context, taskType, effort)
		},
	}

	cmd.Flags().StringVar(&context, "context", "", "why this task matters")
	cmd.Flags().StringVar(&taskType, "type", "", "task type (creative, administrative, technical, physical)")
	cmd.Flags().StringVar(&effort, "effort", "", "effort level (quick-win, medium, deep-work)")

	return cmd
}

func runTaskAdd(_ *cobra.Command, text, context, taskType, effort string) error {
	formatter := NewOutputFormatter(os.Stdout, jsonOutput)

	var task *core.Task
	if context != "" {
		task = core.NewTaskWithContext(text, context)
	} else {
		task = core.NewTask(text)
	}

	if taskType != "" {
		task.Type = core.TaskType(taskType)
	}
	if effort != "" {
		task.Effort = core.TaskEffort(effort)
	}

	if err := task.Validate(); err != nil {
		if jsonOutput {
			_ = formatter.WriteJSONError("task add", ExitValidation, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitValidation)
	}

	ctx, err := bootstrap()
	if err != nil {
		if jsonOutput {
			_ = formatter.WriteJSONError("task add", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	ctx.pool.AddTask(task)
	if err := ctx.provider.SaveTask(task); err != nil {
		if jsonOutput {
			_ = formatter.WriteJSONError("task add", ExitProviderError, fmt.Sprintf("save task: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: save task: %v\n", err)
		}
		os.Exit(ExitProviderError)
	}

	shortID := task.ID[:8]
	if jsonOutput {
		return formatter.WriteJSON("task add", task, nil)
	}
	return formatter.Writef("Created task %s: %s\n", shortID, task.Text)
}

// completeResult tracks the outcome of completing a single task.
type completeResult struct {
	ID       string `json:"id"`
	ShortID  string `json:"short_id"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// newTaskCompleteCmd creates the "task complete" subcommand.
func newTaskCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <id> [id...]",
		Short: "Mark tasks as complete",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskComplete(cmd, args)
		},
	}
	return cmd
}

func runTaskComplete(_ *cobra.Command, ids []string) error {
	formatter := NewOutputFormatter(os.Stdout, jsonOutput)

	ctx, err := bootstrap()
	if err != nil {
		if jsonOutput {
			_ = formatter.WriteJSONError("task complete", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	results := make([]completeResult, 0, len(ids))
	worstExit := ExitSuccess

	for _, idPrefix := range ids {
		result := completeOneTask(ctx, idPrefix)
		results = append(results, result)
		if result.ExitCode > worstExit {
			worstExit = result.ExitCode
		}
	}

	if jsonOutput {
		_ = formatter.WriteJSON("task complete", results, nil)
	} else {
		for _, r := range results {
			if r.Success {
				_ = formatter.Writef("Completed task %s\n", r.ShortID)
			} else {
				fmt.Fprintf(os.Stderr, "Error completing %s: %s\n", r.ShortID, r.Error)
			}
		}
	}

	if worstExit != ExitSuccess {
		os.Exit(worstExit)
	}
	return nil
}

func completeOneTask(ctx *cliContext, idPrefix string) completeResult {
	matches := ctx.pool.FindByPrefix(idPrefix)

	if len(matches) == 0 {
		return completeResult{
			ID:       idPrefix,
			ShortID:  shortID(idPrefix),
			Success:  false,
			Error:    "task not found",
			ExitCode: ExitNotFound,
		}
	}

	if len(matches) > 1 {
		return completeResult{
			ID:       idPrefix,
			ShortID:  shortID(idPrefix),
			Success:  false,
			Error:    fmt.Sprintf("ambiguous prefix, matches %d tasks", len(matches)),
			ExitCode: ExitAmbiguousInput,
		}
	}

	task := matches[0]
	if err := task.UpdateStatus(core.StatusComplete); err != nil {
		return completeResult{
			ID:       task.ID,
			ShortID:  shortID(task.ID),
			Success:  false,
			Error:    err.Error(),
			ExitCode: ExitValidation,
		}
	}

	if err := ctx.provider.SaveTask(task); err != nil {
		return completeResult{
			ID:       task.ID,
			ShortID:  shortID(task.ID),
			Success:  false,
			Error:    fmt.Sprintf("save: %v", err),
			ExitCode: ExitProviderError,
		}
	}

	return completeResult{
		ID:       task.ID,
		ShortID:  shortID(task.ID),
		Success:  true,
		ExitCode: ExitSuccess,
	}
}
