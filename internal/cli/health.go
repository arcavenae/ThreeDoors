package cli

import (
	"fmt"
	"os"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// healthCheckJSON is the JSON representation of a single health check item.
type healthCheckJSON struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// healthResultJSON is the JSON envelope data for the health command.
type healthResultJSON struct {
	Overall    string            `json:"overall"`
	DurationMs int64             `json:"duration_ms"`
	Checks     []healthCheckJSON `json:"checks"`
}

// newHealthCmd creates the "health" subcommand.
func newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run system health checks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHealth(cmd)
		},
	}
	return cmd
}

func runHealth(cmd *cobra.Command) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	ctx, err := bootstrap()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("health", ExitProviderError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitProviderError)
	}

	hc := core.NewHealthChecker(ctx.provider)
	result := hc.RunAll()

	if isJSON {
		checks := make([]healthCheckJSON, 0, len(result.Items))
		for _, item := range result.Items {
			checks = append(checks, healthCheckJSON{
				Name:    item.Name,
				Status:  string(item.Status),
				Message: item.Message,
			})
		}
		data := healthResultJSON{
			Overall:    string(result.Overall),
			DurationMs: result.Duration.Milliseconds(),
			Checks:     checks,
		}
		_ = formatter.WriteJSON("health", data, nil)
	} else {
		tw := formatter.TableWriter()
		_, _ = fmt.Fprintf(tw, "CHECK\tSTATUS\tMESSAGE\n")
		for _, item := range result.Items {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Name, item.Status, item.Message)
		}
		_ = tw.Flush()
		_ = formatter.Writef("\nOverall: %s (%dms)\n", result.Overall, result.Duration.Milliseconds())
	}

	if result.Overall == core.HealthFail {
		os.Exit(ExitProviderError)
	}
	return nil
}
