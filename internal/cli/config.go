package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/spf13/cobra"
)

// validConfigKeys lists the config keys that can be read/written via CLI.
var validConfigKeys = map[string]bool{
	"provider":             true,
	"note_title":           true,
	"theme":                true,
	"dev_dispatch_enabled": true,
	"schema_version":       true,
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and modify configuration",
	}
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display the full configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd)
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a single config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet(cmd, args[0])
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(cmd, args[0], args[1])
		},
	}
}

func loadConfig() (string, *core.ProviderConfig, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return "", nil, fmt.Errorf("config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return "", nil, fmt.Errorf("load config: %w", err)
	}
	return configPath, cfg, nil
}

func runConfigShow(cmd *cobra.Command) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if isJSON {
		return formatter.WriteJSON("config show", configToMap(cfg), nil)
	}

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "KEY\tVALUE\n")
	_, _ = fmt.Fprintf(tw, "schema_version\t%d\n", cfg.SchemaVersion)
	_, _ = fmt.Fprintf(tw, "provider\t%s\n", cfg.Provider)
	_, _ = fmt.Fprintf(tw, "note_title\t%s\n", cfg.NoteTitle)
	_, _ = fmt.Fprintf(tw, "theme\t%s\n", cfg.Theme)
	_, _ = fmt.Fprintf(tw, "dev_dispatch_enabled\t%t\n", cfg.DevDispatchEnabled)
	return tw.Flush()
}

func runConfigGet(cmd *cobra.Command, key string) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	if !validConfigKeys[key] {
		if isJSON {
			_ = formatter.WriteJSONError("config get", ExitValidation, fmt.Sprintf("unknown config key: %s", key), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: unknown config key: %s\n", key)
		}
		os.Exit(ExitValidation)
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	value := getConfigValue(cfg, key)

	if isJSON {
		return formatter.WriteJSON("config get", map[string]string{"key": key, "value": value}, nil)
	}
	return formatter.Writef("%s\n", value)
}

func runConfigSet(cmd *cobra.Command, key, value string) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	if !validConfigKeys[key] {
		if isJSON {
			_ = formatter.WriteJSONError("config set", ExitValidation, fmt.Sprintf("unknown config key: %s", key), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: unknown config key: %s\n", key)
		}
		os.Exit(ExitValidation)
	}

	configPath, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := setConfigValue(cfg, key, value); err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("config set", ExitValidation, err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(ExitValidation)
	}

	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	if isJSON {
		return formatter.WriteJSON("config set", map[string]string{"key": key, "value": value}, nil)
	}
	return formatter.Writef("Set %s = %s\n", key, value)
}

func getConfigValue(cfg *core.ProviderConfig, key string) string {
	switch key {
	case "provider":
		return cfg.Provider
	case "note_title":
		return cfg.NoteTitle
	case "theme":
		return cfg.Theme
	case "dev_dispatch_enabled":
		return strconv.FormatBool(cfg.DevDispatchEnabled)
	case "schema_version":
		return strconv.Itoa(cfg.SchemaVersion)
	default:
		return ""
	}
}

func setConfigValue(cfg *core.ProviderConfig, key, value string) error {
	switch key {
	case "provider":
		cfg.Provider = value
	case "note_title":
		cfg.NoteTitle = value
	case "theme":
		cfg.Theme = value
	case "dev_dispatch_enabled":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for %s: %w", key, err)
		}
		cfg.DevDispatchEnabled = b
	case "schema_version":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for %s: %w", key, err)
		}
		cfg.SchemaVersion = v
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func configToMap(cfg *core.ProviderConfig) map[string]string {
	return map[string]string{
		"schema_version":       strconv.Itoa(cfg.SchemaVersion),
		"provider":             cfg.Provider,
		"note_title":           cfg.NoteTitle,
		"theme":                cfg.Theme,
		"dev_dispatch_enabled": strconv.FormatBool(cfg.DevDispatchEnabled),
	}
}
