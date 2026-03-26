package cli

import (
	"fmt"
	"os"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newMergeDriverCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "merge-driver <base> <local> <remote>",
		Short:  "Git custom merge driver for tasks.yaml",
		Long:   "Three-way merge driver for task files. Registered in .gitattributes and invoked by Git during merge operations.",
		Args:   cobra.ExactArgs(3),
		Hidden: true, // internal command, not user-facing
		RunE:   runMergeDriver,
	}
}

func runMergeDriver(_ *cobra.Command, args []string) error {
	basePath := args[0]
	localPath := args[1]
	remotePath := args[2]

	baseTasks, err := loadTasksFromYAML(basePath)
	if err != nil {
		return fmt.Errorf("load base (%s): %w", basePath, err)
	}

	localTasks, err := loadTasksFromYAML(localPath)
	if err != nil {
		return fmt.Errorf("load local (%s): %w", localPath, err)
	}

	remoteTasks, err := loadTasksFromYAML(remotePath)
	if err != nil {
		return fmt.Errorf("load remote (%s): %w", remotePath, err)
	}

	outcome := core.ThreeWayMergeTaskLists(baseTasks, localTasks, remoteTasks)

	if err := writeTasksToYAML(localPath, outcome.MergedTasks); err != nil {
		return fmt.Errorf("write merged result: %w", err)
	}

	// Log conflicts if any
	if len(outcome.Conflicts) > 0 {
		configDir, err := core.GetConfigDirPath()
		if err == nil {
			cl, logErr := core.NewConflictLog(configDir)
			if logErr == nil {
				_ = cl.LogConflicts(outcome.Conflicts)
			}
		}
	}

	return nil
}

// taskFileFormat wraps tasks for YAML serialization.
type taskFileFormat struct {
	Tasks []*core.Task `yaml:"tasks"`
}

func loadTasksFromYAML(path string) ([]*core.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file taskFileFormat
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	return file.Tasks, nil
}

func writeTasksToYAML(path string, tasks []*core.Task) error {
	file := taskFileFormat{Tasks: tasks}
	data, err := yaml.Marshal(&file)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	// Atomic write: tmp → sync → rename
	tmpPath := path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
