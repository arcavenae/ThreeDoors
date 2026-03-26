package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/intelligence/llm"
	"github.com/arcavenae/ThreeDoors/internal/intelligence/services"
	"github.com/spf13/cobra"
)

// extractResultJSON is the JSON representation of extract command output.
type extractResultJSON struct {
	Tasks    []extractedTaskJSON `json:"tasks"`
	Imported int                 `json:"imported"`
	Source   string              `json:"source"`
}

// extractedTaskJSON is the JSON representation of a single extracted task.
type extractedTaskJSON struct {
	Text       string   `json:"text"`
	Effort     int      `json:"effort"`
	Tags       []string `json:"tags"`
	Confidence float64  `json:"confidence,omitempty"`
}

// extractDeps holds injectable dependencies for the extract command.
type extractDeps struct {
	stdin         io.Reader
	stdinStatFunc func() (os.FileInfo, error)
	configLoader  func() (llm.Config, error)
	backendFunc   func(ctx context.Context, cfg llm.Config) (llm.LLMBackend, error)
	importFunc    func(tasks []services.ExtractedTask) (int, error)
	promptFunc    func(reader io.Reader, prompt string) (string, error)
}

func newExtractCmd() *cobra.Command {
	var (
		filePath  string
		clipboard bool
		autoYes   bool
	)

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract tasks from text using LLM",
		Long: `Extract actionable tasks from unstructured text using an LLM backend.

Input sources (in priority order):
  --file <path>      Read from a file
  --clipboard        Read from the system clipboard
  (stdin)            Auto-detected when piped

Output:
  Default: human-readable numbered list with confirmation prompt
  --json:  JSON array for scripting (no prompts)
  --yes:   Auto-import without confirmation`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps := defaultExtractDeps()
			return runExtract(cmd, filePath, clipboard, autoYes, deps)
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to text file to extract from")
	cmd.Flags().BoolVarP(&clipboard, "clipboard", "c", false, "extract from system clipboard")
	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "auto-import all extracted tasks without confirmation")

	return cmd
}

// defaultExtractDeps returns production dependencies.
func defaultExtractDeps() extractDeps {
	return extractDeps{
		stdin:         os.Stdin,
		stdinStatFunc: os.Stdin.Stat,
		configLoader:  loadLLMConfig,
		backendFunc: func(ctx context.Context, cfg llm.Config) (llm.LLMBackend, error) {
			backend, _, err := llm.DiscoverBackend(ctx, cfg)
			return backend, err
		},
		importFunc: defaultImportFunc,
		promptFunc: defaultPromptFunc,
	}
}

func runExtract(cmd *cobra.Command, filePath string, clipboard, autoYes bool, deps extractDeps) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	// Discover LLM backend.
	cfg, err := deps.configLoader()
	if err != nil {
		return writeExtractError(formatter, isJSON, "load config: "+err.Error(),
			"Run 'threedoors doctor' to check your configuration.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	backend, err := deps.backendFunc(ctx, cfg)
	if err != nil {
		return writeExtractError(formatter, isJSON, "no LLM backend available",
			"Install an LLM backend: claude CLI, gemini CLI, or ollama.\nSee 'threedoors doctor' for details.")
	}

	extractor := services.NewTaskExtractor(backend, services.WithRunner(&llm.ExecRunner{}))

	// Resolve input source.
	var tasks []services.ExtractedTask
	switch {
	case filePath != "":
		tasks, err = extractor.ExtractFromFile(ctx, filePath)
		if err != nil {
			return writeExtractError(formatter, isJSON, err.Error(), "")
		}
	case clipboard:
		tasks, err = extractor.ExtractFromClipboard(ctx)
		if err != nil {
			return writeExtractError(formatter, isJSON, err.Error(), "")
		}
	default:
		// Check for stdin pipe.
		text, stdinErr := readStdin(deps.stdin, deps.stdinStatFunc)
		if stdinErr != nil {
			return writeExtractError(formatter, isJSON,
				"no input provided",
				"Usage: threedoors extract --file <path>\n       threedoors extract --clipboard\n       cat notes.txt | threedoors extract")
		}
		tasks, err = extractor.ExtractFromText(ctx, text)
		if err != nil {
			return writeExtractError(formatter, isJSON, err.Error(), "")
		}
	}

	if len(tasks) == 0 {
		if isJSON {
			return formatter.WriteJSON("extract", extractResultJSON{
				Tasks:    []extractedTaskJSON{},
				Imported: 0,
				Source:   resolveSource(filePath, clipboard),
			}, nil)
		}
		_ = formatter.Writef("No tasks found in input.\n")
		return nil
	}

	// JSON output mode: emit and exit (no prompts).
	if isJSON {
		jsonTasks := toJSONTasks(tasks)
		if autoYes {
			imported, importErr := deps.importFunc(tasks)
			if importErr != nil {
				return writeExtractError(formatter, isJSON, "import failed: "+importErr.Error(), "")
			}
			return formatter.WriteJSON("extract", extractResultJSON{
				Tasks:    jsonTasks,
				Imported: imported,
				Source:   resolveSource(filePath, clipboard),
			}, nil)
		}
		return formatter.WriteJSON("extract", extractResultJSON{
			Tasks:    jsonTasks,
			Imported: 0,
			Source:   resolveSource(filePath, clipboard),
		}, nil)
	}

	// Human-readable output.
	_ = formatter.Writef("Extracted %d task(s):\n\n", len(tasks))
	for i, t := range tasks {
		effortLabel := effortToLabel(t.Effort)
		tags := ""
		if len(t.Tags) > 0 {
			tags = fmt.Sprintf(" [%s]", strings.Join(t.Tags, ", "))
		}
		_ = formatter.Writef("  %d. %s  (%s)%s\n", i+1, t.Text, effortLabel, tags)
	}
	_ = formatter.Writef("\n")

	if autoYes {
		imported, importErr := deps.importFunc(tasks)
		if importErr != nil {
			return fmt.Errorf("import tasks: %w", importErr)
		}
		_ = formatter.Writef("Imported %d task(s).\n", imported)
		return nil
	}

	// Interactive confirmation.
	response, promptErr := deps.promptFunc(deps.stdin, "Import all? [y/N/select] ")
	if promptErr != nil {
		return nil // EOF or no terminal — skip import
	}

	response = strings.TrimSpace(strings.ToLower(response))
	switch response {
	case "y", "yes":
		imported, importErr := deps.importFunc(tasks)
		if importErr != nil {
			return fmt.Errorf("import tasks: %w", importErr)
		}
		_ = formatter.Writef("Imported %d task(s).\n", imported)
	case "select", "s":
		imported, selectErr := selectiveImport(formatter, tasks, deps)
		if selectErr != nil {
			return selectErr
		}
		_ = formatter.Writef("Imported %d task(s).\n", imported)
	default:
		_ = formatter.Writef("Import cancelled.\n")
	}

	return nil
}

// selectiveImport prompts the user to enter task numbers to import.
func selectiveImport(formatter *OutputFormatter, tasks []services.ExtractedTask, deps extractDeps) (int, error) {
	_ = formatter.Writef("Enter task numbers to import (comma-separated, e.g. 1,3,5): ")

	response, err := deps.promptFunc(deps.stdin, "")
	if err != nil {
		return 0, nil
	}

	var selected []services.ExtractedTask
	for _, part := range strings.Split(response, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var idx int
		if _, scanErr := fmt.Sscanf(part, "%d", &idx); scanErr != nil {
			continue
		}
		if idx >= 1 && idx <= len(tasks) {
			selected = append(selected, tasks[idx-1])
		}
	}

	if len(selected) == 0 {
		_ = formatter.Writef("No valid tasks selected.\n")
		return 0, nil
	}

	return deps.importFunc(selected)
}

// readStdin reads all content from stdin if it's being piped/redirected.
func readStdin(stdin io.Reader, statFunc func() (os.FileInfo, error)) (string, error) {
	info, err := statFunc()
	if err != nil {
		return "", fmt.Errorf("stat stdin: %w", err)
	}

	if (info.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("stdin is a terminal, not piped")
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("stdin is empty")
	}

	return text, nil
}

// loadLLMConfig loads the LLM configuration from the standard config path.
func loadLLMConfig() (llm.Config, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return llm.Config{}, fmt.Errorf("get config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return llm.Config{}, fmt.Errorf("load config: %w", err)
	}

	return cfg.LLM, nil
}

// defaultImportFunc imports extracted tasks by creating core.Task objects
// and saving them via the configured provider.
func defaultImportFunc(tasks []services.ExtractedTask) (int, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return 0, fmt.Errorf("get config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return 0, fmt.Errorf("load config: %w", err)
	}

	provider := core.NewProviderFromConfig(cfg)

	existing, err := provider.LoadTasks()
	if err != nil {
		return 0, fmt.Errorf("load existing tasks: %w", err)
	}

	for _, et := range tasks {
		t := core.NewTask(et.Text)
		t.Effort = mapEffort(et.Effort)
		existing = append(existing, t)
	}

	if err := provider.SaveTasks(existing); err != nil {
		return 0, fmt.Errorf("save tasks: %w", err)
	}

	return len(tasks), nil
}

// defaultPromptFunc reads a line from the reader after printing a prompt.
func defaultPromptFunc(reader io.Reader, prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(os.Stderr, prompt)
	}
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return scanner.Text(), nil
}

// mapEffort converts a 1-5 effort score to a TaskEffort category.
func mapEffort(effort int) core.TaskEffort {
	switch {
	case effort <= 2:
		return core.EffortQuickWin
	case effort <= 3:
		return core.EffortMedium
	default:
		return core.EffortDeepWork
	}
}

// effortToLabel converts a 1-5 effort score to a human-readable label.
func effortToLabel(effort int) string {
	switch {
	case effort <= 1:
		return "trivial"
	case effort == 2:
		return "quick"
	case effort == 3:
		return "medium"
	case effort == 4:
		return "significant"
	default:
		return "major"
	}
}

func toJSONTasks(tasks []services.ExtractedTask) []extractedTaskJSON {
	result := make([]extractedTaskJSON, len(tasks))
	for i, t := range tasks {
		tags := t.Tags
		if tags == nil {
			tags = []string{}
		}
		result[i] = extractedTaskJSON{
			Text:       t.Text,
			Effort:     t.Effort,
			Tags:       tags,
			Confidence: t.Confidence,
		}
	}
	return result
}

func resolveSource(filePath string, clipboard bool) string {
	switch {
	case filePath != "":
		return "file:" + filePath
	case clipboard:
		return "clipboard"
	default:
		return "stdin"
	}
}

func writeExtractError(formatter *OutputFormatter, isJSON bool, message, detail string) error {
	if isJSON {
		_ = formatter.WriteJSONError("extract", ExitGeneralError, message, detail)
		return fmt.Errorf("%s", message)
	}
	if detail != "" {
		return fmt.Errorf("%s\n%s", message, detail)
	}
	return fmt.Errorf("%s", message)
}
