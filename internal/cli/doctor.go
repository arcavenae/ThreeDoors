package cli

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// doctorCategoryJSON is the JSON representation of a category result.
type doctorCategoryJSON struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Checks []doctorCheckJSON `json:"checks"`
}

// doctorCheckJSON is the JSON representation of a single check result.
type doctorCheckJSON struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// doctorResultJSON is the JSON envelope data for the doctor command.
type doctorResultJSON struct {
	Categories []doctorCategoryJSON `json:"categories"`
}

// newDoctorCmd creates the "doctor" command with "health" as an alias.
func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"health"},
		Short:   "Run system diagnostics",
		Long: `Run comprehensive system diagnostics and display results
with category-based output. The 'health' command is an alias for 'doctor'.

Use --fix to automatically repair safe, reversible issues.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd)
		},
	}
	cmd.Flags().Bool("fix", false, "Auto-repair safe, reversible issues")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed sub-check information")
	cmd.Flags().String("category", "", "Run only specific categories (comma-separated: env,tasks,providers,sessions,sync,db,version)")
	cmd.Flags().Bool("skip-version", false, "Skip the version check category")
	return cmd
}

func runDoctor(cmd *cobra.Command) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	// Parse doctor-specific flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	categoryFlag, _ := cmd.Flags().GetString("category")
	skipVersion, _ := cmd.Flags().GetBool("skip-version")

	opts := core.DoctorOptions{Verbose: verbose}

	// Build category filter list
	if categoryFlag != "" {
		cats := strings.Split(categoryFlag, ",")
		for i := range cats {
			cats[i] = strings.TrimSpace(cats[i])
		}
		// Validate category names
		for _, c := range cats {
			if _, ok := core.ValidCategoryKeys[c]; !ok {
				validKeys := validCategoryKeyList()
				if isJSON {
					_ = formatter.WriteJSONError("doctor", ExitValidation,
						fmt.Sprintf("unknown category %q", c),
						fmt.Sprintf("valid categories: %s", strings.Join(validKeys, ", ")))
				} else {
					fmt.Fprintf(os.Stderr, "Error: unknown category %q\nValid categories: %s\n",
						c, strings.Join(validKeys, ", "))
				}
				os.Exit(ExitValidation)
			}
		}
		opts.Categories = cats
	}

	// --skip-version removes "version" from the enabled set (or adds all-except-version)
	if skipVersion {
		if len(opts.Categories) == 0 {
			// No explicit categories — enable all except version
			for key := range core.ValidCategoryKeys {
				if key != "version" {
					opts.Categories = append(opts.Categories, key)
				}
			}
		} else {
			// Remove version from explicit list
			filtered := opts.Categories[:0]
			for _, c := range opts.Categories {
				if c != "version" {
					filtered = append(filtered, c)
				}
			}
			opts.Categories = filtered
		}
	}

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("doctor", ExitProviderError, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitProviderError)
	}

	fix, _ := cmd.Flags().GetBool("fix")

	dc := core.NewDoctorChecker(configDir)
	dc.SetFix(fix)

	// Detect terminal capabilities
	dc.SetTerminalInfo(detectTerminalInfo())
	dc.SetVersionInfo(Version, Channel, nil, "")
	dc.SetRegistry(core.DefaultRegistry())
	result := dc.RunWithOptions(opts)

	if isJSON {
		return writeDoctorJSON(formatter, result)
	}
	if err := writeDoctorHuman(formatter, result, opts.Verbose); err != nil {
		return err
	}
	if code := doctorExitCode(result); code != ExitSuccess {
		os.Exit(code) //nolint:gocritic // intentional exit with doctor-specific code for scripting
	}
	return nil
}

// validCategoryKeyList returns sorted valid category keys for error messages.
func validCategoryKeyList() []string {
	keys := make([]string, 0, len(core.ValidCategoryKeys))
	for k := range core.ValidCategoryKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeDoctorJSON(formatter *OutputFormatter, result core.DoctorResult) error {
	cats := make([]doctorCategoryJSON, 0, len(result.Categories))
	for _, cat := range result.Categories {
		checks := make([]doctorCheckJSON, 0, len(cat.Checks))
		for _, check := range cat.Checks {
			checks = append(checks, doctorCheckJSON{
				Name:       check.Name,
				Status:     check.Status.String(),
				Message:    check.Message,
				Suggestion: check.Suggestion,
			})
		}
		cats = append(cats, doctorCategoryJSON{
			Name:   cat.Name,
			Status: cat.Status.String(),
			Checks: checks,
		})
	}
	data := doctorResultJSON{Categories: cats}
	return formatter.WriteJSON("doctor", data, nil)
}

func writeDoctorHuman(formatter *OutputFormatter, result core.DoctorResult, verbose bool) error {
	// Header
	_ = formatter.Writef("ThreeDoors Doctor (%s • %s/%s)\n\n", Version, runtime.GOOS, runtime.GOARCH)

	// Category results with icons
	for _, cat := range result.Categories {
		icon := statusIcon(cat.Status)
		_ = formatter.Writef("%s %s\n", icon, cat.Name)

		if cat.Status == core.CheckSkip {
			_ = formatter.Writef("    %s Skipped (not selected)\n", statusIcon(core.CheckSkip))
			_ = formatter.Writef("\n")
			continue
		}

		for _, check := range cat.Checks {
			checkIcon := statusIcon(check.Status)
			_ = formatter.Writef("    %s %s\n", checkIcon, check.Message)
			if check.Suggestion != "" {
				_ = formatter.Writef("      → %s\n", check.Suggestion)
			}
			if verbose && check.Name != "" {
				_ = formatter.Writef("      Name: %s\n", check.Name)
			}
		}
		_ = formatter.Writef("\n")
	}

	// Summary line
	fixedCount := result.FixedCount()
	manualCount := result.ManualCount()

	if fixedCount == 0 && manualCount == 0 {
		_ = formatter.Writef("No issues found. Your system is ready to use.\n")
	} else if fixedCount > 0 && manualCount == 0 {
		_ = formatter.Writef("Fixed %d %s.\n",
			fixedCount, pluralize("issue", fixedCount))
	} else if fixedCount > 0 {
		_ = formatter.Writef("Fixed %d %s. %d %s require manual intervention.\n",
			fixedCount, pluralize("issue", fixedCount),
			manualCount, pluralize("issue", manualCount))
	} else {
		catCount := result.CategoryIssueCount()
		_ = formatter.Writef("Doctor found issues in %d %s.\n",
			catCount, pluralize("category", catCount))
		if result.HasFixableIssues() {
			_ = formatter.Writef("Run 'threedoors doctor --fix' to auto-repair fixable issues.\n")
		}
	}

	return nil
}

// doctorExitCode returns the appropriate exit code for a doctor result.
// 0 = no issues, 1 = warnings only, 2 = errors found.
func doctorExitCode(result core.DoctorResult) int {
	_, errors := result.IssueCount()
	warnings, _ := result.IssueCount()
	switch {
	case errors > 0:
		return ExitDoctorError
	case warnings > 0:
		return ExitDoctorWarning
	default:
		return ExitSuccess
	}
}

// statusIcon returns a styled icon string for a check status.
func statusIcon(status core.CheckStatus) string {
	icon := status.Icon()

	// Apply color based on status (respects NO_COLOR via lipgloss)
	switch status {
	case core.CheckOK:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(icon)
	case core.CheckInfo:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(icon)
	case core.CheckFixed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(icon)
	case core.CheckSkip:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(icon)
	case core.CheckWarn:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(icon)
	case core.CheckFail:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(icon)
	default:
		return icon
	}
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if word == "category" {
		return "categories"
	}
	return word + "s"
}

// detectTerminalInfo gathers terminal size and color profile for doctor checks.
func detectTerminalInfo() core.TerminalInfo {
	info := core.TerminalInfo{}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		info.Width = width
		info.Height = height
	}

	info.ColorProfile = lipgloss.ColorProfile().Name()

	return info
}
