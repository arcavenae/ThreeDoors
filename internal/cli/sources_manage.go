package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/spf13/cobra"
)

// sourcesActionJSON is the JSON output for management commands that perform
// a single action on a connection (pause, resume, sync, disconnect, reauth, edit).
type sourcesActionJSON struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	Status string `json:"status"`
}

func newSourcesPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause sync for a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesPauseTo(manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume sync for a paused connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			return runSourcesResumeTo(manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
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
			return runSourcesSyncTo(manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesDisconnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <name>",
		Short: "Remove a data source connection",
		Long: `Remove a data source connection and delete its credentials.

By default, prompts for how to handle synced tasks. Use --keep-tasks to
preserve tasks without prompting.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)
			keepTasks, _ := cmd.Flags().GetBool("keep-tasks")
			jsonMode := isJSONOutput(cmd)
			return runSourcesDisconnectTo(manager, svc, args[0], keepTasks, cmd.Flags().Changed("keep-tasks"), os.Stdin, os.Stdout, jsonMode)
		},
	}
	cmd.Flags().Bool("keep-tasks", false, "Preserve synced tasks after disconnecting")
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
			return runSourcesReauthTo(manager, svc, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

func newSourcesEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit connection settings (re-opens connect wizard)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, _ := sourcesManager(ctx)
			return runSourcesEditTo(manager, args[0], os.Stdout, isJSONOutput(cmd))
		},
	}
}

// runSourcesPauseTo pauses sync for a named connection.
func runSourcesPauseTo(manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
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
			return formatter.WriteJSONError("sources pause", ExitGeneralError,
				fmt.Sprintf("cannot pause %q: %v", name, err), stateHint(conn.State, "pause"))
		}
		return fmt.Errorf("cannot pause %q: %w\n%s", name, err, stateHint(conn.State, "pause"))
	}

	if jsonMode {
		return formatter.WriteJSON("sources pause", sourcesActionJSON{
			Name:   name,
			Action: "pause",
			Status: "paused",
		}, nil)
	}

	return formatter.Writef("Paused sync for %q.\n", name)
}

// runSourcesResumeTo resumes sync for a paused connection.
func runSourcesResumeTo(manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
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
			return formatter.WriteJSONError("sources resume", ExitGeneralError,
				fmt.Sprintf("cannot resume %q: %v", name, err), stateHint(conn.State, "resume"))
		}
		return fmt.Errorf("cannot resume %q: %w\n%s", name, err, stateHint(conn.State, "resume"))
	}

	if jsonMode {
		return formatter.WriteJSON("sources resume", sourcesActionJSON{
			Name:   name,
			Action: "resume",
			Status: "connected",
		}, nil)
	}

	return formatter.Writef("Resumed sync for %q.\n", name)
}

// runSourcesSyncTo triggers an immediate sync for a connection.
func runSourcesSyncTo(manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
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
			return formatter.WriteJSONError("sources sync", ExitGeneralError,
				fmt.Sprintf("sync failed for %q: %v", name, err), stateHint(conn.State, "sync"))
		}
		return fmt.Errorf("sync failed for %q: %w\n%s", name, err, stateHint(conn.State, "sync"))
	}

	if jsonMode {
		return formatter.WriteJSON("sources sync", sourcesActionJSON{
			Name:   name,
			Action: "sync",
			Status: "connected",
		}, nil)
	}

	return formatter.Writef("Sync complete for %q.\n", name)
}

// runSourcesDisconnectTo removes a connection.
// When keepTasksFlag is false and flagExplicit is false (i.e., --keep-tasks was
// not passed), prompts interactively via reader for how to handle tasks.
func runSourcesDisconnectTo(manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, keepTasks, flagExplicit bool, reader io.Reader, w io.Writer, jsonMode bool) error {
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

	// If --keep-tasks was not explicitly set, prompt interactively.
	if !flagExplicit && !jsonMode {
		_ = formatter.Writef("Disconnecting %q will remove the connection and delete credentials.\n", name)
		_ = formatter.Writef("Keep synced tasks? [y/N] ")
		scanner := bufio.NewScanner(reader)
		if scanner.Scan() {
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			keepTasks = answer == "y" || answer == "yes"
		}
	}

	if err := svc.Remove(conn.ID, keepTasks); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources disconnect", ExitGeneralError,
				fmt.Sprintf("disconnect failed for %q: %v", name, err), "")
		}
		return fmt.Errorf("disconnect failed for %q: %w", name, err)
	}

	taskAction := "removed"
	if keepTasks {
		taskAction = "preserved"
	}

	if jsonMode {
		return formatter.WriteJSON("sources disconnect", sourcesActionJSON{
			Name:   name,
			Action: "disconnect",
			Status: fmt.Sprintf("disconnected (tasks %s)", taskAction),
		}, nil)
	}

	return formatter.Writef("Disconnected %q. Credentials deleted, tasks %s.\n", name, taskAction)
}

// runSourcesReauthTo re-authenticates a connection by transitioning through Connecting.
func runSourcesReauthTo(manager *connection.ConnectionManager, svc *connection.ConnectionService, name string, w io.Writer, jsonMode bool) error {
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

	// Re-auth: transition to Connecting, then test, then Connected.
	// Only valid from AuthExpired or Error states.
	if conn.State != connection.StateAuthExpired && conn.State != connection.StateError {
		msg := fmt.Sprintf("cannot reauth %q: connection is %s, must be auth_expired or error", name, conn.State)
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError, msg, "")
		}
		return fmt.Errorf("%s", msg)
	}

	if err := manager.Transition(conn.ID, connection.StateConnecting); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError,
				fmt.Sprintf("reauth transition failed for %q: %v", name, err), "")
		}
		return fmt.Errorf("reauth transition failed for %q: %w", name, err)
	}

	// Test the connection to verify credentials.
	result, testErr := svc.TestConnection(conn.ID)
	if testErr != nil || !result.Healthy() {
		_ = manager.TransitionWithError(conn.ID, connection.StateError, "re-authentication failed")
		msg := fmt.Sprintf("re-authentication failed for %q", name)
		if testErr != nil {
			msg = fmt.Sprintf("re-authentication failed for %q: %v", name, testErr)
		}
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError, msg, "")
		}
		return fmt.Errorf("%s", msg)
	}

	if err := manager.Transition(conn.ID, connection.StateConnected); err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources reauth", ExitGeneralError,
				fmt.Sprintf("reauth complete transition failed for %q: %v", name, err), "")
		}
		return fmt.Errorf("reauth complete transition failed for %q: %w", name, err)
	}

	if jsonMode {
		return formatter.WriteJSON("sources reauth", sourcesActionJSON{
			Name:   name,
			Action: "reauth",
			Status: "connected",
		}, nil)
	}

	return formatter.Writef("Re-authenticated %q. Connection is now active.\n", name)
}

// runSourcesEditTo prints instructions for editing a connection.
// Full wizard re-entry requires TUI support; for now this prints the connect
// command the user would run to reconfigure.
func runSourcesEditTo(manager *connection.ConnectionManager, name string, w io.Writer, jsonMode bool) error {
	formatter := NewOutputFormatter(w, jsonMode)

	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources edit", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	hint := fmt.Sprintf("threedoors connect %s --label %q", conn.ProviderName, conn.Label)

	if jsonMode {
		return formatter.WriteJSON("sources edit", struct {
			Name     string `json:"name"`
			Action   string `json:"action"`
			Provider string `json:"provider"`
			Command  string `json:"command"`
		}{
			Name:     name,
			Action:   "edit",
			Provider: conn.ProviderName,
			Command:  hint,
		}, nil)
	}

	_ = formatter.Writef("To reconfigure %q, disconnect and reconnect:\n", name)
	_ = formatter.Writef("  threedoors sources disconnect %q --keep-tasks\n", name)
	return formatter.Writef("  %s\n", hint)
}

// stateHint returns a helpful message explaining why an operation may have
// failed for the given state.
func stateHint(state connection.ConnectionState, action string) string {
	switch action {
	case "pause":
		if state == connection.StatePaused {
			return "Connection is already paused."
		}
		if state != connection.StateConnected {
			return fmt.Sprintf("Connection must be connected to pause (currently %s).", state)
		}
	case "resume":
		if state == connection.StateConnected {
			return "Connection is already active."
		}
		if state != connection.StatePaused {
			return fmt.Sprintf("Connection must be paused to resume (currently %s).", state)
		}
	case "sync":
		if state != connection.StateConnected {
			return fmt.Sprintf("Connection must be connected to sync (currently %s).", state)
		}
	}
	return ""
}
