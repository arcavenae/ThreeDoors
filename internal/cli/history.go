package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// historyRecordJSON is the JSON representation of a completion record.
type historyRecordJSON struct {
	Title       string `json:"title"`
	CompletedAt string `json:"completed_at"`
	Source      string `json:"source"`
	TaskID      string `json:"task_id"`
}

// newHistoryCmd creates the "history" command that displays completed tasks.
func newHistoryCmd() *cobra.Command {
	var (
		today bool
		week  bool
		month bool
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show completed tasks",
		Long: `Display completed tasks from your history. By default, shows today's completions.
Use --week, --month, or --all to change the time range.
Use --json to output as a JSON array for programmatic consumption.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configDir, err := core.GetConfigDirPath()
			if err != nil {
				return fmt.Errorf("config dir: %w", err)
			}
			reader := core.NewCompletionReader(configDir)
			return runHistoryTo(cmd.Context(), reader, cmd.OutOrStdout(), isJSONOutput(cmd), today, week, month, all)
		},
	}

	cmd.Flags().BoolVar(&today, "today", false, "show today's completions (default)")
	cmd.Flags().BoolVar(&week, "week", false, "show this week's completions")
	cmd.Flags().BoolVar(&month, "month", false, "show this month's completions")
	cmd.Flags().BoolVar(&all, "all", false, "show all completions")

	return cmd
}

// runHistoryTo executes the history command, writing output to w.
// Extracted for testability — callers provide the CompletionReader and writer.
func runHistoryTo(ctx context.Context, reader *core.CompletionReader, w io.Writer, jsonMode, today, week, month, all bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	records, err := fetchRecords(ctx, reader, today, week, month, all)
	if err != nil {
		if jsonMode {
			_ = formatter.WriteJSONError("history", ExitGeneralError, fmt.Sprintf("read completions: %v", err), "")
		}
		return fmt.Errorf("read completions: %w", err)
	}

	if len(records) == 0 {
		if jsonMode {
			return formatter.WriteJSON("history", []historyRecordJSON{}, nil)
		}
		return formatter.Writef("No completed tasks found.\n")
	}

	if jsonMode {
		return writeHistoryJSON(formatter, records)
	}

	return writeHistoryHuman(formatter, records)
}

// fetchRecords selects the appropriate CompletionReader method based on flags.
// Last flag wins when multiple are set. Default is today.
func fetchRecords(ctx context.Context, reader *core.CompletionReader, today, week, month, all bool) ([]core.CompletionRecord, error) {
	switch {
	case all:
		return reader.Read(ctx)
	case month:
		return reader.ThisMonth(ctx)
	case week:
		return reader.ThisWeek(ctx)
	default:
		return reader.Today(ctx)
	}
}

// writeHistoryJSON writes completion records as a JSON array.
func writeHistoryJSON(formatter *OutputFormatter, records []core.CompletionRecord) error {
	items := make([]historyRecordJSON, len(records))
	for i, rec := range records {
		items[i] = historyRecordJSON{
			Title:       rec.Title,
			CompletedAt: rec.CompletedAt.Format(time.RFC3339),
			Source:      rec.Source,
			TaskID:      rec.TaskID,
		}
	}
	return formatter.WriteJSON("history", items, nil)
}

// writeHistoryHuman writes completion records grouped by day in human-readable format.
func writeHistoryHuman(formatter *OutputFormatter, records []core.CompletionRecord) error {
	type dayGroup struct {
		date    time.Time
		records []core.CompletionRecord
	}

	var groups []dayGroup
	var currentDate string

	for _, rec := range records {
		local := rec.CompletedAt.Local()
		dateKey := local.Format("2006-01-02")
		if dateKey != currentDate {
			groups = append(groups, dayGroup{date: local})
			currentDate = dateKey
		}
		groups[len(groups)-1].records = append(groups[len(groups)-1].records, rec)
	}

	for i, g := range groups {
		if i > 0 {
			_ = formatter.Writef("\n")
		}
		_ = formatter.Writef("%s\n", formatDayHeader(g.date))
		for _, rec := range g.records {
			local := rec.CompletedAt.Local()
			_ = formatter.Writef("  %-36s %s\n", rec.Title, local.Format("15:04"))
		}
	}

	return nil
}

// formatDayHeader formats a date as "Weekday — Month Day" (e.g., "Sunday — March 15").
func formatDayHeader(t time.Time) string {
	return fmt.Sprintf("%s — %s %d", t.Weekday(), t.Month(), t.Day())
}
