package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/device"
	"github.com/spf13/cobra"
)

func newDevicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List known devices",
		Long:  "Lists all known ThreeDoors device identities from the local registry.",
		RunE:  runDevicesList,
	}

	cmd.AddCommand(newDevicesRenameCmd())

	return cmd
}

func newDevicesRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <device-id> <new-name>",
		Short: "Rename a device",
		Long:  "Updates the human-readable name of a device in both local device.yaml and the registry.",
		Args:  cobra.ExactArgs(2),
		RunE:  runDevicesRename,
	}
}

func devicesRegistry() (*device.LocalDeviceRegistry, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return nil, fmt.Errorf("get config dir: %w", err)
	}
	devicesDir := filepath.Join(configDir, "devices")
	return device.NewLocalDeviceRegistry(devicesDir), nil
}

func runDevicesList(cmd *cobra.Command, _ []string) error {
	reg, err := devicesRegistry()
	if err != nil {
		return err
	}

	devices, err := reg.List()
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}

	if isJSONOutput(cmd) {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(devices)
	}

	if len(devices) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "No devices registered."); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tDEVICE ID\tFIRST SEEN\tLAST SYNC"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, dev := range devices {
		lastSync := "never"
		if !dev.LastSync.IsZero() {
			lastSync = dev.LastSync.Format("2006-01-02 15:04 UTC")
		}
		firstSeen := dev.FirstSeen.Format("2006-01-02 15:04 UTC")
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dev.Name, dev.ID, firstSeen, lastSync); err != nil {
			return fmt.Errorf("write device row: %w", err)
		}
	}
	return w.Flush()
}

func runDevicesRename(cmd *cobra.Command, args []string) error {
	rawID := args[0]
	newName := args[1]

	id, err := device.NewDeviceID(rawID)
	if err != nil {
		return err
	}

	reg, err := devicesRegistry()
	if err != nil {
		return err
	}

	dev, err := reg.Get(id)
	if err != nil {
		return err
	}

	dev.Name = newName
	if err := reg.Update(dev); err != nil {
		return fmt.Errorf("update registry: %w", err)
	}

	// Also update local device.yaml if this is the current device
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return fmt.Errorf("get config dir: %w", err)
	}

	localPath := filepath.Join(configDir, "device.yaml")
	localDev, err := device.LoadDevice(localPath)
	if err == nil && localDev.ID == id {
		localDev.Name = newName
		if err := device.SaveDevice(localDev, localPath); err != nil {
			return fmt.Errorf("update local device.yaml: %w", err)
		}
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Device %s renamed to %q\n", id, newName); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
