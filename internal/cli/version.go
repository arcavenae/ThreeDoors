package cli

import (
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time variables set via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// versionData holds structured version info for JSON output.
type versionData struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// newVersionCmd creates the "version" subcommand.
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVersion(cmd)
		},
	}
	return cmd
}

func runVersion(cmd *cobra.Command) error {
	return writeVersion(os.Stdout, isJSONOutput(cmd))
}

func writeVersion(w io.Writer, isJSON bool) error {
	formatter := NewOutputFormatter(w, isJSON)

	data := versionData{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
	}

	if isJSON {
		return formatter.WriteJSON("version", data, nil)
	}

	_ = formatter.Writef("ThreeDoors %s\n", data.Version)
	_ = formatter.Writef("Commit:     %s\n", data.Commit)
	_ = formatter.Writef("Built:      %s\n", data.BuildDate)
	return formatter.Writef("Go version: %s\n", data.GoVersion)
}
