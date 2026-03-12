package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/spf13/cobra"
)

// sourcesConnJSON is the JSON representation of a connection in list output.
type sourcesConnJSON struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	LastSync string `json:"last_sync"`
	Tasks    int    `json:"tasks"`
}

// sourcesStatusJSON is the JSON representation of detailed connection status.
type sourcesStatusJSON struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	Status       string `json:"status"`
	Server       string `json:"server,omitempty"`
	LastSync     string `json:"last_sync"`
	TasksActive  int    `json:"tasks_active"`
	Filter       string `json:"filter,omitempty"`
	SyncMode     string `json:"sync_mode"`
	PollInterval string `json:"poll_interval"`
	LastError    string `json:"last_error,omitempty"`
}

// sourcesTestCheckJSON is the JSON representation of a single health check.
type sourcesTestCheckJSON struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

// sourcesTestJSON is the JSON representation of test results.
type sourcesTestJSON struct {
	Name    string                 `json:"name"`
	Healthy bool                   `json:"healthy"`
	Checks  []sourcesTestCheckJSON `json:"checks"`
}

// newSourcesCmd creates the "sources" command group with list, status, and test subcommands.
func newSourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "List and manage data source connections",
		Long: `List all configured data source connections. Use subcommands
to view detailed status or test connection health.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, _ := sourcesManager(ctx)
			return runSourcesListTo(cmd, manager, os.Stdout, isJSONOutput(cmd))
		},
	}

	cmd.AddCommand(newSourcesStatusCmd())
	cmd.AddCommand(newSourcesTestCmd())
	cmd.AddCommand(newSourcesLogCmd())

	return cmd
}

func newSourcesStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <name>",
		Short: "Show detailed status of a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, _ := sourcesManager(ctx)
			return runSourcesStatusTo(cmd, manager, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <name>",
		Short: "Test connection health",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			code := runSourcesTestTo(cmd, manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
			if code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
}

// sourcesManager extracts the ConnectionManager and Service from a cliContext.
// Returns a bare manager with no connections if none were resolved.
func sourcesManager(ctx *cliContext) (*connection.ConnectionManager, *connection.ConnectionService) {
	if ctx.resolved != nil {
		return ctx.resolved.Manager, ctx.resolved.Service
	}
	return connection.NewConnectionManager(nil), nil
}

func runSourcesListTo(_ *cobra.Command, manager *connection.ConnectionManager, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conns := manager.List()

	if jsonMode {
		items := make([]sourcesConnJSON, 0, len(conns))
		for _, c := range conns {
			items = append(items, sourcesConnJSON{
				Name:     c.Label,
				Provider: c.ProviderName,
				Status:   c.State.String(),
				LastSync: formatSyncTime(c.LastSync),
				Tasks:    c.TaskCount,
			})
		}
		return formatter.WriteJSON("sources", items, nil)
	}

	if len(conns) == 0 {
		return formatter.Writef("No connections configured.\n")
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "NAME\tPROVIDER\tSTATUS\tLAST SYNC\tTASKS\n")
	for _, c := range conns {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
			c.Label, c.ProviderName, c.State.String(),
			formatSyncTime(c.LastSync), c.TaskCount)
	}
	return tw.Flush()
}

func runSourcesStatusTo(_ *cobra.Command, manager *connection.ConnectionManager, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources status", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if jsonMode {
		data := sourcesStatusJSON{
			Name:         conn.Label,
			Provider:     conn.ProviderName,
			Status:       conn.State.String(),
			Server:       conn.Settings["server"],
			LastSync:     formatSyncTime(conn.LastSync),
			TasksActive:  conn.TaskCount,
			Filter:       conn.Settings["filter"],
			SyncMode:     conn.SyncMode,
			PollInterval: conn.PollInterval.String(),
			LastError:    conn.LastError,
		}
		return formatter.WriteJSON("sources status", data, nil)
	}

	_ = formatter.Writef("Name:          %s\n", conn.Label)
	_ = formatter.Writef("Provider:      %s\n", conn.ProviderName)
	_ = formatter.Writef("Status:        %s\n", conn.State.String())
	if server := conn.Settings["server"]; server != "" {
		_ = formatter.Writef("Server:        %s\n", server)
	}
	_ = formatter.Writef("Last Sync:     %s\n", formatSyncTime(conn.LastSync))
	_ = formatter.Writef("Tasks Active:  %d\n", conn.TaskCount)
	if filter := conn.Settings["filter"]; filter != "" {
		_ = formatter.Writef("Filter:        %s\n", filter)
	}
	_ = formatter.Writef("Sync Mode:     %s\n", conn.SyncMode)
	_ = formatter.Writef("Poll Interval: %s\n", conn.PollInterval.String())
	if conn.LastError != "" {
		_ = formatter.Writef("Last Error:    %s\n", conn.LastError)
	}

	return nil
}

// runSourcesTestTo runs health checks and writes results. Returns the exit code
// (0 for healthy, ExitNotFound for missing connection, ExitGeneralError for unhealthy).
func runSourcesTestTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) int {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			_ = formatter.WriteJSONError("sources test", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		} else {
			_ = formatter.Writef("Error: connection %q not found\n", name)
		}
		return ExitNotFound
	}

	if svc == nil {
		if jsonMode {
			_ = formatter.WriteJSONError("sources test", ExitGeneralError,
				"no connection service configured", "")
		} else {
			_ = formatter.Writef("Error: no connection service configured\n")
		}
		return ExitGeneralError
	}

	result, err := svc.TestConnection(conn.ID)
	if err != nil {
		if jsonMode {
			_ = formatter.WriteJSONError("sources test", ExitGeneralError,
				fmt.Sprintf("health check failed: %v", err), "")
		} else {
			_ = formatter.Writef("Error: health check failed: %v\n", err)
		}
		return ExitGeneralError
	}

	checks := []struct {
		name   string
		passed bool
	}{
		{"DNS resolution", result.APIReachable},
		{"TLS", result.APIReachable},
		{"Authentication", result.TokenValid},
		{"Authorization", result.TokenValid},
		{"Rate limit", result.RateLimitOK},
	}

	healthy := result.Healthy()

	if jsonMode {
		jsonChecks := make([]sourcesTestCheckJSON, 0, len(checks))
		for _, c := range checks {
			jsonChecks = append(jsonChecks, sourcesTestCheckJSON{
				Name:   c.name,
				Passed: c.passed,
			})
		}
		data := sourcesTestJSON{
			Name:    name,
			Healthy: healthy,
			Checks:  jsonChecks,
		}
		_ = formatter.WriteJSON("sources test", data, nil)
		if !healthy {
			return ExitGeneralError
		}
		return 0
	}

	_ = formatter.Writef("Health check: %s\n", name)
	for _, c := range checks {
		icon := "✓"
		if !c.passed {
			icon = "✗"
		}
		_ = formatter.Writef("  %s %s\n", icon, c.name)
	}

	if !healthy {
		return ExitGeneralError
	}

	return 0
}

// formatSyncTime formats a sync timestamp for display.
// Returns "never" for zero times.
func formatSyncTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02 15:04")
}
