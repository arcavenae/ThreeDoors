package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arcavenae/ThreeDoors/internal/docaudit"
	"github.com/spf13/cobra"
)

func newDocAuditCmd() *cobra.Command {
	var projectRoot string
	var jsonlPath string

	cmd := &cobra.Command{
		Use:   "doc-audit",
		Short: "Audit planning doc consistency across story files, epic-list, epics-and-stories, and ROADMAP",
		Long: `Cross-check the four planning doc layers for drift:
  - docs/stories/*.story.md (story file status — authoritative for individual stories)
  - docs/prd/epic-list.md (epic-level status)
  - docs/prd/epics-and-stories.md (comprehensive story/epic listing)
  - ROADMAP.md (synced copy for merge-queue scope checks)

Detects status mismatches, orphaned stories, phantom stories, and epic status drift.
Used by the retrospector agent as part of its periodic deep analysis rotation.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectRoot == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				projectRoot = cwd
			}

			docs, err := docaudit.LoadDocSet(projectRoot)
			if err != nil {
				return fmt.Errorf("load planning docs: %w", err)
			}

			auditor := docaudit.NewAuditor(docs)
			result := auditor.Run()

			if isJSONOutput(cmd) {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if err := docaudit.FormatHumanSummary(os.Stdout, result); err != nil {
				return fmt.Errorf("format summary: %w", err)
			}

			// Write JSONL if path is provided.
			if jsonlPath != "" {
				if err := os.MkdirAll(filepath.Dir(jsonlPath), 0o755); err != nil {
					return fmt.Errorf("create JSONL dir: %w", err)
				}
				if err := docaudit.WriteJSONL(jsonlPath, result); err != nil {
					return fmt.Errorf("write JSONL: %w", err)
				}
			}

			if !result.Clean {
				return fmt.Errorf("found %d inconsistencies", len(result.Findings))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectRoot, "root", "", "project root directory (default: current directory)")
	cmd.Flags().StringVar(&jsonlPath, "jsonl", "", "path to JSONL findings log (appends entry)")

	return cmd
}
