package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newTaskCmd creates the "task" command group.
func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}
	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskCompleteCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskShowCmd())
	return cmd
}

// stdinDetector abstracts TTY detection for testing.
var stdinDetector = func() bool {
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

// stdinReader abstracts stdin reading for testing.
var stdinReader io.Reader = os.Stdin

// newTaskAddCmd creates the "task add" subcommand.
func newTaskAddCmd() *cobra.Command {
	var (
		context  string
		taskType string
		effort   string
		useStdin bool
	)

	cmd := &cobra.Command{
		Use:   "add [text]",
		Short: "Add a new task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if useStdin {
				return runTaskAddFromStdin(cmd, stdinReader, context, taskType, effort)
			}
			if len(args) == 0 && stdinDetector() {
				return runTaskAddSingleStdin(cmd, stdinReader, context, taskType, effort)
			}
			if len(args) == 0 {
				return fmt.Errorf("requires task text as argument or pipe input via stdin")
			}
			return runTaskAdd(cmd, args[0], context, taskType, effort)
		},
	}

	cmd.Flags().StringVar(&context, "context", "", "why this task matters")
	cmd.Flags().StringVar(&taskType, "type", "", "task type (creative, administrative, technical, physical)")
	cmd.Flags().StringVar(&effort, "effort", "", "effort level (quick-win, medium, deep-work)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read multiple tasks from stdin (one per line)")

	return cmd
}

func runTaskAdd(cmd *cobra.Command, text, context, taskType, effort string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

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
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task add", ExitValidation, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitValidation)
	}

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task add", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	ctx.pool.AddTask(task)
	if err := ctx.provider.SaveTask(task); err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task add", ExitProviderError, fmt.Sprintf("save task: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: save task: %v\n", err)
		}
		os.Exit(ExitProviderError)
	}

	shortID := task.ID[:8]
	if isJSONOutput(cmd) {
		return formatter.WriteJSON("task add", task, nil)
	}
	return formatter.Writef("Created task %s: %s\n", shortID, task.Text)
}

// runTaskAddSingleStdin reads a single task text from stdin (auto-detected, no flag).
func runTaskAddSingleStdin(_ *cobra.Command, r io.Reader, context, taskType, effort string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return fmt.Errorf("no task text provided via stdin")
	}
	return runTaskAdd(nil, text, context, taskType, effort)
}

// runTaskAddFromStdin reads multiple tasks from stdin, one per line (--stdin flag).
func runTaskAddFromStdin(cmd *cobra.Command, r io.Reader, context, taskType, effort string) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	ctx, err := bootstrap()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("task add", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	scanner := bufio.NewScanner(r)
	var created []*core.Task
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

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
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid task %q: %v\n", text, err)
			continue
		}

		ctx.pool.AddTask(task)
		if err := ctx.provider.SaveTask(task); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save task %q: %v\n", text, err)
			continue
		}

		created = append(created, task)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	if isJSON {
		return formatter.WriteJSON("task add", created, map[string]int{"count": len(created)})
	}

	for _, t := range created {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", t.ID[:8])
	}
	return nil
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

func runTaskComplete(cmd *cobra.Command, ids []string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
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

	if isJSONOutput(cmd) {
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

// listMetadata holds metadata for JSON list output.
type listMetadata struct {
	Total    int               `json:"total"`
	Filtered int               `json:"filtered"`
	Filters  map[string]string `json:"filters"`
}

// newTaskListCmd creates the "task list" subcommand.
func newTaskListCmd() *cobra.Command {
	var (
		statusFilter string
		typeFilter   string
		effortFilter string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskList(cmd, statusFilter, typeFilter, effortFilter)
		},
	}

	cmd.Flags().StringVar(&statusFilter, "status", "", "filter by status (todo, in-progress, blocked, etc.)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "filter by type (creative, administrative, technical, physical)")
	cmd.Flags().StringVar(&effortFilter, "effort", "", "filter by effort (quick-win, medium, deep-work)")

	return cmd
}

func runTaskList(cmd *cobra.Command, statusFilter, typeFilter, effortFilter string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task list", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	allTasks := ctx.pool.GetAllTasks()
	sort.Slice(allTasks, func(i, j int) bool {
		return allTasks[i].CreatedAt.Before(allTasks[j].CreatedAt)
	})

	filtered := filterTasks(allTasks, statusFilter, typeFilter, effortFilter)

	if isJSONOutput(cmd) {
		filters := make(map[string]string)
		if statusFilter != "" {
			filters["status"] = statusFilter
		}
		if typeFilter != "" {
			filters["type"] = typeFilter
		}
		if effortFilter != "" {
			filters["effort"] = effortFilter
		}
		meta := listMetadata{
			Total:    len(allTasks),
			Filtered: len(filtered),
			Filters:  filters,
		}
		return formatter.WriteJSON("task list", filtered, meta)
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "ID\tSTATUS\tTYPE\tEFFORT\tTEXT\n")
	for _, t := range filtered {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			shortID(t.ID),
			t.Status,
			t.Type,
			t.Effort,
			t.Text,
		)
	}
	_ = tw.Flush()
	return formatter.Writef("%d tasks found\n", len(filtered))
}

func filterTasks(tasks []*core.Task, status, taskType, effort string) []*core.Task {
	result := make([]*core.Task, 0, len(tasks))
	for _, t := range tasks {
		if status != "" && string(t.Status) != status {
			continue
		}
		if taskType != "" && string(t.Type) != taskType {
			continue
		}
		if effort != "" && string(t.Effort) != effort {
			continue
		}
		result = append(result, t)
	}
	return result
}

// newTaskShowCmd creates the "task show" subcommand.
func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskShow(cmd, args[0])
		},
	}
	return cmd
}

func runTaskShow(cmd *cobra.Command, idPrefix string) error {
	formatter := NewOutputFormatter(os.Stdout, isJSONOutput(cmd))

	ctx, err := bootstrap()
	if err != nil {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task show", ExitGeneralError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	matches := ctx.pool.FindByPrefix(idPrefix)

	if len(matches) == 0 {
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task show", ExitNotFound, "task not found", idPrefix)
		} else {
			fmt.Fprintf(os.Stderr, "Error: no task found with prefix %q\n", idPrefix)
		}
		os.Exit(ExitNotFound)
	}

	if len(matches) > 1 {
		msg := fmt.Sprintf("ambiguous prefix %q matches %d tasks", idPrefix, len(matches))
		if isJSONOutput(cmd) {
			_ = formatter.WriteJSONError("task show", ExitAmbiguousInput, msg, "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		}
		os.Exit(ExitAmbiguousInput)
	}

	task := matches[0]

	if isJSONOutput(cmd) {
		return formatter.WriteJSON("task show", task, nil)
	}

	_ = formatter.Writef("ID:        %s\n", task.ID)
	_ = formatter.Writef("Text:      %s\n", task.Text)
	_ = formatter.Writef("Status:    %s\n", task.Status)
	if task.Context != "" {
		_ = formatter.Writef("Context:   %s\n", task.Context)
	}
	if task.Type != "" {
		_ = formatter.Writef("Type:      %s\n", task.Type)
	}
	if task.Effort != "" {
		_ = formatter.Writef("Effort:    %s\n", task.Effort)
	}
	if task.Location != "" {
		_ = formatter.Writef("Location:  %s\n", task.Location)
	}
	if task.Blocker != "" {
		_ = formatter.Writef("Blocker:   %s\n", task.Blocker)
	}
	_ = formatter.Writef("Created:   %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	_ = formatter.Writef("Updated:   %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
	if task.CompletedAt != nil {
		_ = formatter.Writef("Completed: %s\n", task.CompletedAt.Format("2006-01-02 15:04:05"))
	}
	if len(task.Notes) > 0 {
		_ = formatter.Writef("Notes:\n")
		for _, n := range task.Notes {
			_ = formatter.Writef("  [%s] %s\n", n.Timestamp.Format("2006-01-02 15:04"), n.Text)
		}
	}

	return nil
}
