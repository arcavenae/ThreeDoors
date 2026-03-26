package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	"github.com/spf13/cobra"
)

// sourcesLogEventJSON is the JSON representation of a sync event.
type sourcesLogEventJSON struct {
	Timestamp  time.Time `json:"timestamp"`
	Connection string    `json:"connection"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Summary    string    `json:"summary"`
}

const defaultLogLimit = 20

func newSourcesLogCmd() *cobra.Command {
	var last int
	var errorsOnly bool

	cmd := &cobra.Command{
		Use:   "log [name]",
		Short: "View sync log events",
		Long: `View recent sync events for a specific connection or across all connections.
Use --errors to filter to only error and conflict events.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := bootstrap()
			if err != nil {
				return err
			}
			manager, svc := sourcesManager(ctx)

			var eventLog *connection.SyncEventLog
			if svc != nil {
				eventLog = svc.EventLog()
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			limit := last
			if limit <= 0 {
				limit = defaultLogLimit
			}

			return runSourcesLogTo(cmd, manager, eventLog, name, limit, errorsOnly, os.Stdout, isJSONOutput(cmd))
		},
	}

	cmd.Flags().IntVar(&last, "last", 0, "Number of events to show (default 20)")
	cmd.Flags().BoolVar(&errorsOnly, "errors", false, "Show only error and conflict events")

	return cmd
}

// runSourcesLogTo implements the log command logic, writing to w.
func runSourcesLogTo(
	_ *cobra.Command,
	manager *connection.ConnectionManager,
	eventLog *connection.SyncEventLog,
	name string,
	limit int,
	errorsOnly bool,
	w io.Writer,
	jsonMode bool,
) error {
	formatter := NewOutputFormatter(w, jsonMode)

	if eventLog == nil {
		if jsonMode {
			return formatter.WriteJSON("sources log", []sourcesLogEventJSON{}, nil)
		}
		return formatter.Writef("No sync log available.\n")
	}

	if name != "" {
		return sourcesLogForConnection(formatter, manager, eventLog, name, limit, errorsOnly, jsonMode)
	}

	return sourcesLogAllConnections(formatter, manager, eventLog, limit, errorsOnly, jsonMode)
}

// sourcesLogForConnection shows log events for a single named connection.
func sourcesLogForConnection(
	formatter *OutputFormatter,
	manager *connection.ConnectionManager,
	eventLog *connection.SyncEventLog,
	name string,
	limit int,
	errorsOnly bool,
	jsonMode bool,
) error {
	conn, err := manager.GetByLabel(name)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("sources log", ExitNotFound,
				fmt.Sprintf("connection %q not found", name), "")
		}
		return fmt.Errorf("connection %q not found", name)
	}

	events, err := fetchEvents(eventLog, conn.ID, limit, errorsOnly)
	if err != nil {
		return fmt.Errorf("read sync log for %q: %w", name, err)
	}

	return renderEvents(formatter, events, conn.Label, jsonMode)
}

// sourcesLogAllConnections shows log events across all connections.
func sourcesLogAllConnections(
	formatter *OutputFormatter,
	manager *connection.ConnectionManager,
	eventLog *connection.SyncEventLog,
	limit int,
	errorsOnly bool,
	jsonMode bool,
) error {
	conns := manager.List()
	if len(conns) == 0 {
		if jsonMode {
			return formatter.WriteJSON("sources log", []sourcesLogEventJSON{}, nil)
		}
		return formatter.Writef("No connections configured.\n")
	}

	var allEvents []labeledEvent
	for _, c := range conns {
		events, err := fetchEvents(eventLog, c.ID, 0, errorsOnly)
		if err != nil {
			continue
		}
		for _, e := range events {
			allEvents = append(allEvents, labeledEvent{label: c.Label, event: e})
		}
	}

	// Sort by timestamp descending (most recent first).
	sortLabeledEvents(allEvents)

	if limit > 0 && len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}

	if jsonMode {
		items := make([]sourcesLogEventJSON, 0, len(allEvents))
		for _, le := range allEvents {
			items = append(items, sourcesLogEventJSON{
				Timestamp:  le.event.Timestamp,
				Connection: le.label,
				Type:       string(le.event.Type),
				Status:     eventStatus(le.event),
				Summary:    le.event.Summary,
			})
		}
		return formatter.WriteJSON("sources log", items, nil)
	}

	if len(allEvents) == 0 {
		return formatter.Writef("No events found.\n")
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "TIMESTAMP\tCONNECTION\tSTATUS\tDESCRIPTION\n")
	for _, le := range allEvents {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			le.event.Timestamp.Format("15:04:05"),
			le.label,
			eventStatus(le.event),
			le.event.Summary,
		)
	}
	return tw.Flush()
}

type labeledEvent struct {
	label string
	event connection.SyncEvent
}

// sortLabeledEvents sorts events by timestamp descending (most recent first).
func sortLabeledEvents(events []labeledEvent) {
	for i := 1; i < len(events); i++ {
		for j := i; j > 0 && events[j].event.Timestamp.After(events[j-1].event.Timestamp); j-- {
			events[j], events[j-1] = events[j-1], events[j]
		}
	}
}

// fetchEvents retrieves events from the log, optionally filtering to errors/conflicts.
func fetchEvents(eventLog *connection.SyncEventLog, connectionID string, limit int, errorsOnly bool) ([]connection.SyncEvent, error) {
	if errorsOnly {
		errors, err := eventLog.EventsByType(connectionID, connection.EventSyncError, 0)
		if err != nil {
			return nil, err
		}
		conflicts, err := eventLog.EventsByType(connectionID, connection.EventConflict, 0)
		if err != nil {
			return nil, err
		}

		// Merge and sort by timestamp descending.
		all := append(errors, conflicts...)
		for i := 1; i < len(all); i++ {
			for j := i; j > 0 && all[j].Timestamp.After(all[j-1].Timestamp); j-- {
				all[j], all[j-1] = all[j-1], all[j]
			}
		}
		if limit > 0 && len(all) > limit {
			all = all[:limit]
		}
		return all, nil
	}

	return eventLog.SyncLog(connectionID, limit)
}

// renderEvents outputs events for a single connection.
func renderEvents(formatter *OutputFormatter, events []connection.SyncEvent, connLabel string, jsonMode bool) error {
	if jsonMode {
		items := make([]sourcesLogEventJSON, 0, len(events))
		for _, e := range events {
			items = append(items, sourcesLogEventJSON{
				Timestamp:  e.Timestamp,
				Connection: connLabel,
				Type:       string(e.Type),
				Status:     eventStatus(e),
				Summary:    e.Summary,
			})
		}
		return formatter.WriteJSON("sources log", items, nil)
	}

	if len(events) == 0 {
		return formatter.Writef("No events found for %q.\n", connLabel)
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "TIMESTAMP\tSTATUS\tDESCRIPTION\n")
	for _, e := range events {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n",
			e.Timestamp.Format("15:04:05"),
			eventStatus(e),
			e.Summary,
		)
	}
	return tw.Flush()
}

// eventStatus returns a status indicator for a sync event.
func eventStatus(e connection.SyncEvent) string {
	switch e.Type {
	case connection.EventSyncComplete:
		return "✓"
	case connection.EventSyncError, connection.EventReauthRequired:
		return "✗"
	case connection.EventConflict:
		return "⚠"
	case connection.EventStateChange:
		return "→"
	case connection.EventSyncStart:
		return "▶"
	default:
		return "·"
	}
}
