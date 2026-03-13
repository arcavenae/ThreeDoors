package device

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ThreeDoorsNamespace is the UUID v5 namespace for device ID generation.
var ThreeDoorsNamespace = uuid.NewSHA1(uuid.NameSpaceDNS, []byte("threedoors.app"))

// GenerateDeviceID creates a deterministic device ID from machine ID and config directory,
// or falls back to a random UUID v4 when the machine ID is unavailable.
func GenerateDeviceID(reader MachineIDReader, configDir string) (DeviceID, error) {
	machineID, err := reader.ReadMachineID()
	if err != nil {
		if !errors.Is(err, ErrMachineIDUnavailable) {
			return "", fmt.Errorf("read machine ID: %w", err)
		}
		// Fallback to UUID v4
		id := uuid.New()
		return DeviceID(id.String()), nil
	}

	// UUID v5: namespace + machine-id + config dir path
	input := machineID + ":" + configDir
	id := uuid.NewSHA1(ThreeDoorsNamespace, []byte(input))
	return DeviceID(id.String()), nil
}
