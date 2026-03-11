package docaudit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// WriteJSONL appends a single JSONL entry for this audit result to the given
// findings log file. Creates the file if it does not exist.
func WriteJSONL(path string, result AuditResult) error {
	var entry JSONLEntry
	entry.Timestamp = result.Timestamp.Format(time.RFC3339)
	entry.Repo = "ThreeDoors"

	if result.Clean {
		entry.Type = "doc_audit_clean"
		entry.Summary = "Doc consistency audit passed — no inconsistencies found"
	} else {
		entry.Type = "doc_inconsistency"
		entry.Findings = result.Findings
		entry.Summary = fmt.Sprintf("Doc consistency audit found %d inconsistencies", len(result.Findings))
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal JSONL entry: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open JSONL file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write JSONL entry: %w", err)
	}
	return nil
}

// FormatHumanSummary writes a human-readable summary of the audit result.
func FormatHumanSummary(w io.Writer, result AuditResult) error {
	if result.Clean {
		_, err := fmt.Fprintf(w, "Doc consistency audit: CLEAN (%s)\n", result.Timestamp.Format(time.RFC3339))
		return err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Doc consistency audit: %d inconsistencies found (%s)\n\n",
		len(result.Findings), result.Timestamp.Format(time.RFC3339))

	// Group findings by type.
	groups := make(map[FindingType][]Finding)
	for _, f := range result.Findings {
		groups[f.Type] = append(groups[f.Type], f)
	}

	typeLabels := map[FindingType]string{
		FindingStatusMismatch:  "Status Mismatches",
		FindingOrphanedStory:   "Orphaned Stories (file exists, no planning doc reference)",
		FindingPhantomStory:    "Phantom Stories (planning doc reference, no file)",
		FindingEpicMismatch:    "Epic Status Mismatches",
		FindingMissingStoryRef: "Missing Story References",
	}

	order := []FindingType{
		FindingStatusMismatch,
		FindingEpicMismatch,
		FindingOrphanedStory,
		FindingPhantomStory,
		FindingMissingStoryRef,
	}

	for _, ft := range order {
		findings, ok := groups[ft]
		if !ok {
			continue
		}
		label := typeLabels[ft]
		fmt.Fprintf(&sb, "### %s (%d)\n", label, len(findings))
		for _, f := range findings {
			id := f.StoryID
			if id == "" {
				id = "Epic " + f.EpicID
			}
			fmt.Fprintf(&sb, "  - %s: %s\n", id, f.Description)
			fmt.Fprintf(&sb, "    Expected: %s | Actual: %s\n", f.Expected, f.Actual)
			fmt.Fprintf(&sb, "    Fix: %s\n", f.Fix)
		}
		fmt.Fprintln(&sb)
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// FormatSupervisorMessage creates a concise message for the supervisor.
func FormatSupervisorMessage(result AuditResult) string {
	if result.Clean {
		return ""
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "DOC AUDIT: %d inconsistencies found.\n", len(result.Findings))

	// Count by type.
	counts := make(map[FindingType]int)
	for _, f := range result.Findings {
		counts[f.Type]++
	}

	if n := counts[FindingStatusMismatch]; n > 0 {
		fmt.Fprintf(&sb, "- %d status mismatches (story file vs ROADMAP)\n", n)
	}
	if n := counts[FindingEpicMismatch]; n > 0 {
		fmt.Fprintf(&sb, "- %d epic status mismatches\n", n)
	}
	if n := counts[FindingOrphanedStory]; n > 0 {
		fmt.Fprintf(&sb, "- %d orphaned stories (file exists, no planning doc ref)\n", n)
	}
	if n := counts[FindingPhantomStory]; n > 0 {
		fmt.Fprintf(&sb, "- %d phantom stories (planning doc ref, no file)\n", n)
	}

	fmt.Fprintf(&sb, "Run `threedoors doc-audit` for details.")
	return sb.String()
}
