package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/metrics"
	"github.com/spf13/cobra"
)

// validMoods lists the predefined mood values accepted by the CLI.
var validMoods = []string{
	"focused",
	"energized",
	"tired",
	"stressed",
	"neutral",
	"calm",
	"distracted",
}

// isValidMood checks if a mood string is a known predefined mood.
func isValidMood(mood string) bool {
	lower := strings.ToLower(mood)
	for _, v := range validMoods {
		if lower == v {
			return true
		}
	}
	return false
}

// newMoodCmd creates the "mood" command group.
func newMoodCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mood",
		Short: "Track and view mood entries",
	}
	cmd.AddCommand(newMoodSetCmd())
	cmd.AddCommand(newMoodHistoryCmd())
	return cmd
}

// newMoodSetCmd creates the "mood set" subcommand.
func newMoodSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <mood> [custom-text]",
		Short: "Record a mood entry",
		Long: `Record a mood entry. Accepted moods: focused, energized, tired, stressed, neutral, calm, distracted.
Use "custom" as the mood value followed by a custom text string for free-form moods.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMoodSet(cmd, args)
		},
	}
	return cmd
}

func runMoodSet(cmd *cobra.Command, args []string) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)
	mood := strings.ToLower(args[0])
	customText := ""

	if mood == "custom" {
		if len(args) < 2 {
			msg := "custom mood requires a text argument"
			if isJSON {
				_ = formatter.WriteJSONError("mood set", ExitValidation, msg, "")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
			}
			os.Exit(ExitValidation)
		}
		customText = args[1]
	} else if !isValidMood(mood) {
		msg := fmt.Sprintf("invalid mood %q; valid moods: %s, custom", mood, strings.Join(validMoods, ", "))
		if isJSON {
			_ = formatter.WriteJSONError("mood set", ExitValidation, msg, "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		}
		os.Exit(ExitValidation)
	}

	// Record mood via a temporary session tracker and persist
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("mood set", ExitGeneralError, fmt.Sprintf("config dir: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: config dir: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	tracker := core.NewSessionTracker()
	tracker.RecordMood(mood, customText)
	sessionMetrics := tracker.Finalize()

	writer := core.NewMetricsWriter(configDir)
	if err := writer.AppendSession(sessionMetrics); err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("mood set", ExitGeneralError, fmt.Sprintf("save mood: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: save mood: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	entry := core.MoodEntry{
		Timestamp:  time.Now().UTC(),
		Mood:       mood,
		CustomText: customText,
	}

	if isJSON {
		return formatter.WriteJSON("mood set", entry, nil)
	}
	if customText != "" {
		return formatter.Writef("Recorded mood: custom (%s)\n", customText)
	}
	return formatter.Writef("Recorded mood: %s\n", mood)
}

// newMoodHistoryCmd creates the "mood history" subcommand.
func newMoodHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent mood entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMoodHistory(cmd)
		},
	}
	return cmd
}

func runMoodHistory(cmd *cobra.Command) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("mood history", ExitGeneralError, fmt.Sprintf("config dir: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: config dir: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	reader := metrics.NewReader(sessionsPath)
	sessions, err := reader.ReadAll()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("mood history", ExitGeneralError, fmt.Sprintf("read sessions: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: read sessions: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	// Collect all mood entries from all sessions
	var entries []core.MoodEntry
	for _, s := range sessions {
		entries = append(entries, s.MoodEntries...)
	}

	if isJSON {
		return formatter.WriteJSON("mood history", entries, map[string]int{"total": len(entries)})
	}

	if len(entries) == 0 {
		return formatter.Writef("No mood entries found.\n")
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "TIME\tMOOD\tCUSTOM\n")
	for _, e := range entries {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02 15:04"),
			e.Mood,
			e.CustomText,
		)
	}
	_ = tw.Flush()
	return formatter.Writef("%d mood entries\n", len(entries))
}
