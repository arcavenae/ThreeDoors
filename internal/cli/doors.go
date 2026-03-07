package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// doorEntry represents a single door for JSON output.
type doorEntry struct {
	Door int        `json:"door"`
	Task *core.Task `json:"task"`
}

// NewDoorsCmd creates the "doors" subcommand that presents three randomly selected tasks.
func NewDoorsCmd() *cobra.Command {
	var (
		pick        int
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "doors",
		Short: "Present three randomly selected tasks",
		Long:  "Display three doors — randomly selected tasks from your task pool. Use --pick to select a door and mark it in-progress, or --interactive for a prompted selection.",
		RunE: func(cmd *cobra.Command, args []string) error {
			formatter := NewOutputFormatter(cmd.OutOrStdout(), isJSONOutput(cmd))

			pool, provider, err := loadTaskPool()
			if err != nil {
				if isJSONOutput(cmd) {
					_ = formatter.WriteJSONError("doors", ExitProviderError, "failed to load tasks", err.Error())
				}
				return err
			}

			doors := core.SelectDoors(pool, 3)
			totalAvailable := len(pool.GetAvailableForDoors())

			if len(doors) == 0 {
				if isJSONOutput(cmd) {
					_ = formatter.WriteJSONError("doors", ExitGeneralError, "no tasks available", "")
				} else {
					_ = formatter.Writef("No tasks available.\n")
				}
				return exitError{code: ExitGeneralError}
			}

			if pick > 0 {
				return handlePick(cmd, formatter, doors, pick, provider)
			}

			if err := displayDoors(formatter, doors, totalAvailable); err != nil {
				return err
			}

			if interactive && !isJSONOutput(cmd) {
				if !stdoutIsTerminal() {
					return nil
				}

				picked, promptErr := promptDoorSelection(os.Stdin, cmd.OutOrStdout(), len(doors))
				if promptErr != nil {
					return fmt.Errorf("interactive selection: %w", promptErr)
				}
				return handlePick(cmd, formatter, doors, picked, provider)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&pick, "pick", 0, "select the Nth door (1, 2, or 3) and mark it in-progress")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "prompt to pick a door interactively")

	return cmd
}

// exitError carries an exit code without printing an extra Cobra error message.
type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

// shortID returns the first 8 characters of a task ID.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// displayDoors renders the selected doors in human-readable or JSON format.
func displayDoors(formatter *OutputFormatter, doors []*core.Task, totalAvailable int) error {
	if formatter.IsJSON() {
		entries := make([]doorEntry, len(doors))
		for i, t := range doors {
			entries[i] = doorEntry{Door: i + 1, Task: t}
		}
		meta := map[string]interface{}{
			"total_available":  totalAvailable,
			"selection_method": "diversity",
		}
		return formatter.WriteJSON("doors", entries, meta)
	}

	if len(doors) < 3 {
		_ = formatter.Writef("Note: Only %d task(s) available.\n\n", len(doors))
	}

	tw := formatter.TableWriter()
	if _, err := fmt.Fprintln(tw, "DOOR\tID\tTEXT\tSTATUS\tTYPE\tEFFORT"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for i, t := range doors {
		taskType := string(t.Type)
		if taskType == "" {
			taskType = "-"
		}
		effort := string(t.Effort)
		if effort == "" {
			effort = "-"
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
			i+1, shortID(t.ID), t.Text, t.Status, taskType, effort); err != nil {
			return fmt.Errorf("write door %d: %w", i+1, err)
		}
	}
	return tw.Flush()
}

// handlePick selects the Nth door and marks the task in-progress.
func handlePick(cmd *cobra.Command, formatter *OutputFormatter, doors []*core.Task, pick int, provider core.TaskProvider) error {
	if pick < 1 || pick > len(doors) {
		msg := fmt.Sprintf("--pick must be between 1 and %d", len(doors))
		if formatter.IsJSON() {
			_ = formatter.WriteJSONError("doors.pick", ExitValidation, msg, "")
		}
		return fmt.Errorf("%s", msg)
	}

	task := doors[pick-1]

	if err := task.UpdateStatus(core.StatusInProgress); err != nil {
		if formatter.IsJSON() {
			_ = formatter.WriteJSONError("doors.pick", ExitValidation, "status transition failed", err.Error())
		}
		return fmt.Errorf("status transition: %w", err)
	}

	if err := provider.SaveTask(task); err != nil {
		if formatter.IsJSON() {
			_ = formatter.WriteJSONError("doors.pick", ExitProviderError, "failed to save task", err.Error())
		}
		return fmt.Errorf("save task: %w", err)
	}

	if formatter.IsJSON() {
		entry := doorEntry{Door: pick, Task: task}
		return formatter.WriteJSON("doors.pick", entry, nil)
	}

	_ = formatter.Writef("Selected door %d: %s\n", pick, task.Text)
	_ = formatter.Writef("Task %s marked as in-progress.\n", shortID(task.ID))
	return nil
}

// loadTaskPool initializes a provider, loads tasks, and returns a populated TaskPool.
func loadTaskPool() (*core.TaskPool, core.TaskProvider, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return nil, nil, fmt.Errorf("config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	var provider core.TaskProvider
	if len(cfg.Providers) > 1 {
		agg, aggErr := core.ResolveAllProviders(cfg, core.DefaultRegistry())
		if aggErr != nil {
			return nil, nil, fmt.Errorf("init providers: %w", aggErr)
		}
		provider = agg
	} else {
		provider = core.NewProviderFromConfig(cfg)
	}

	tasks, err := provider.LoadTasks()
	if err != nil {
		return nil, nil, fmt.Errorf("load tasks: %w", err)
	}

	pool := core.NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}

	return pool, provider, nil
}
