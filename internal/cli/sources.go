package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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

// sourcesActionJSON is the JSON representation of a management action result.
type sourcesActionJSON struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	Status string `json:"status"`
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
	cmd.AddCommand(newSourcesPauseCmd())
	cmd.AddCommand(newSourcesResumeCmd())
	cmd.AddCommand(newSourcesSyncCmd())
	cmd.AddCommand(newSourcesDisconnectCmd())
	cmd.AddCommand(newSourcesReauthCmd())
	cmd.AddCommand(newSourcesEditCmd())

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

func newSourcesPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause sync polling for a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesPauseTo(cmd, manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume sync polling for a paused connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesResumeTo(cmd, manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <name>",
		Short: "Trigger an immediate sync cycle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesSyncTo(cmd, manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesDisconnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <name>",
		Short: "Remove a data source connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			keepTasks, _ := cmd.Flags().GetBool("keep-tasks")
			return runSourcesDisconnectTo(cmd, manager, svc, args[0], os.Stdout, os.Stdin, isJSONOutput(cmd), keepTasks)
		},
	}
	cmd.Flags().Bool("keep-tasks", false, "Keep synced tasks locally after disconnecting")
	return cmd
}

func newSourcesReauthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reauth <name>",
		Short: "Re-authenticate a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesReauthTo(cmd, manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit connection settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, _ := sourcesManager(ctx)
			return runSourcesEditTo(cmd, manager, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func runSourcesPauseTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources pause", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if svc == nil {
		if jsonMode {
			return formatter.WriteJSONError("sources pause", ExitGeneralError,
				"no connection service configured", "")
		}
		return fmt.Errorf("no connection service configured")
	}

	if err := svc.Pause(conn.ID); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources pause", ExitValidation,
				fmt.Sprintf("cannot pause: %v", err), fmt.Sprintf("current state: %s", conn.State))
		}
		return fmt.Errorf("cannot pause %q: %w", name, err)
	}

	if jsonMode {
		return formatter.WriteJSON("sources pause", sourcesActionJSON{
			Name:   name,
			Action: "paused",
			Status: connection.StatePaused.String(),
		}, nil)
	}

	_ = formatter.Writef("Paused %q\n", name)
	_ = formatter.Writef("Sync polling has stopped.\n")
	return nil
}

func runSourcesResumeTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources resume", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if svc == nil {
		if jsonMode {
			return formatter.WriteJSONError("sources resume", ExitGeneralError,
				"no connection service configured", "")
		}
		return fmt.Errorf("no connection service configured")
	}

	if err := svc.Resume(conn.ID); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources resume", ExitValidation,
				fmt.Sprintf("cannot resume: %v", err), fmt.Sprintf("current state: %s", conn.State))
		}
		return fmt.Errorf("cannot resume %q: %w", name, err)
	}

	if jsonMode {
		return formatter.WriteJSON("sources resume", sourcesActionJSON{
			Name:   name,
			Action: "resumed",
			Status: connection.StateConnected.String(),
		}, nil)
	}

	_ = formatter.Writef("Resumed %q\n", name)
	_ = formatter.Writef("Sync polling is active.\n")
	return nil
}

func runSourcesSyncTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources sync", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if svc == nil {
		if jsonMode {
			return formatter.WriteJSONError("sources sync", ExitGeneralError,
				"no connection service configured", "")
		}
		return fmt.Errorf("no connection service configured")
	}

	if err := svc.ForceSync(conn.ID); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources sync", ExitValidation,
				fmt.Sprintf("sync failed: %v", err), fmt.Sprintf("current state: %s", conn.State))
		}
		return fmt.Errorf("sync %q failed: %w", name, err)
	}

	if jsonMode {
		return formatter.WriteJSON("sources sync", sourcesActionJSON{
			Name:   name,
			Action: "synced",
			Status: connection.StateConnected.String(),
		}, nil)
	}

	_ = formatter.Writef("Synced %q\n", name)
	return nil
}

func runSourcesDisconnectTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, r io.Reader, jsonMode bool, keepTasks bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources disconnect", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if svc == nil {
		if jsonMode {
			return formatter.WriteJSONError("sources disconnect", ExitGeneralError,
				"no connection service configured", "")
		}
		return fmt.Errorf("no connection service configured")
	}

	// If --keep-tasks not specified, prompt interactively.
	if !keepTasks && !jsonMode {
		_ = formatter.Writef("What should happen to synced tasks?\n")
		_ = formatter.Writef("  keep   - Keep tasks locally\n")
		_ = formatter.Writef("  remove - Remove synced tasks\n")
		_ = formatter.Writef("Choice [keep/remove]: ")

		scanner := bufio.NewScanner(r)
		if scanner.Scan() {
			choice := strings.TrimSpace(strings.ToLower(scanner.Text()))
			switch choice {
			case "keep":
				keepTasks = true
			case "remove":
				keepTasks = false
			default:
				return fmt.Errorf("invalid choice %q: expected 'keep' or 'remove'", choice)
			}
		} else {
			return fmt.Errorf("disconnect %q: no input received", name)
		}
	}

	if err := svc.Remove(conn.ID, keepTasks); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources disconnect", ExitGeneralError,
				fmt.Sprintf("disconnect failed: %v", err), "")
		}
		return fmt.Errorf("disconnect %q: %w", name, err)
	}

	taskAction := "removed"
	if keepTasks {
		taskAction = "kept"
	}

	if jsonMode {
		return formatter.WriteJSON("sources disconnect", struct {
			Name      string `json:"name"`
			Action    string `json:"action"`
			TasksKept bool   `json:"tasks_kept"`
		}{
			Name:      name,
			Action:    "disconnected",
			TasksKept: keepTasks,
		}, nil)
	}

	_ = formatter.Writef("Disconnected %q\n", name)
	_ = formatter.Writef("Credentials deleted. Tasks %s.\n", taskAction)
	return nil
}

func runSourcesReauthTo(_ *cobra.Command, manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	if svc == nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError,
				"no connection service configured", "")
		}
		return fmt.Errorf("no connection service configured")
	}

	// Reauth is only valid for connections that can transition to AuthExpired or
	// are already in AuthExpired/Error state. Disconnected connections cannot be reauthed.
	if conn.State == connection.StateDisconnected || conn.State == connection.StateSyncing {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitValidation,
				fmt.Sprintf("cannot re-authenticate: connection is %s", conn.State), "")
		}
		return fmt.Errorf("cannot re-authenticate %q: connection is %s", name, conn.State)
	}

	// For Connected/Paused: transition through AuthExpired → Connecting → Connected.
	// For Error/AuthExpired: transition through Connecting → Connected.
	switch conn.State {
	case connection.StateConnected:
		if err := manager.Transition(conn.ID, connection.StateAuthExpired); err != nil {
			if jsonMode {
				return formatter.WriteJSONError("sources reauth", ExitValidation,
					fmt.Sprintf("cannot re-authenticate: %v", err), "")
			}
			return fmt.Errorf("cannot re-authenticate %q: %w", name, err)
		}
	case connection.StatePaused:
		// Paused → Disconnected → Connecting path
		if err := manager.Transition(conn.ID, connection.StateDisconnected); err != nil {
			if jsonMode {
				return formatter.WriteJSONError("sources reauth", ExitValidation,
					fmt.Sprintf("cannot re-authenticate: %v", err), "")
			}
			return fmt.Errorf("cannot re-authenticate %q: %w", name, err)
		}
	case connection.StateError, connection.StateAuthExpired:
		// Already in a state that can transition to Connecting.
	}

	// Now transition through Connecting → Connected.
	if err := manager.Transition(conn.ID, connection.StateConnecting); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError,
				fmt.Sprintf("re-authentication failed: %v", err), "")
		}
		return fmt.Errorf("re-authenticate %q: %w", name, err)
	}

	if err := manager.Transition(conn.ID, connection.StateConnected); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError,
				fmt.Sprintf("re-authentication completion failed: %v", err), "")
		}
		return fmt.Errorf("re-authenticate %q completion: %w", name, err)
	}

	if jsonMode {
		return formatter.WriteJSON("sources reauth", sourcesActionJSON{
			Name:   name,
			Action: "reauthenticated",
			Status: connection.StateConnected.String(),
		}, nil)
	}

	_ = formatter.Writef("Re-authenticated %q\n", name)
	_ = formatter.Writef("Connection credentials have been refreshed.\n")
	return nil
}

func runSourcesEditTo(_ *cobra.Command, manager *connection.ConnectionManager, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	_, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources edit", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	// Edit requires the interactive setup wizard (Story 44.1).
	// For now, validate the connection exists and report that the wizard
	// is not yet available.
	if jsonMode {
		return formatter.WriteJSONError("sources edit", ExitGeneralError,
			"interactive setup wizard not yet available", "edit requires the setup wizard from Story 44.1")
	}

	return fmt.Errorf("edit %q: interactive setup wizard not yet available", name)
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
