//go:build darwin

package device

import (
	"fmt"
	"os/exec"
	"strings"
)

type platformMachineIDReader struct{}

func (r *platformMachineIDReader) ReadMachineID() (string, error) {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "", fmt.Errorf("%w: ioreg failed: %v", ErrMachineIDUnavailable, err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, `"`)
			if len(parts) >= 4 {
				return parts[3], nil
			}
		}
	}

	return "", fmt.Errorf("%w: IOPlatformUUID not found in ioreg output", ErrMachineIDUnavailable)
}
