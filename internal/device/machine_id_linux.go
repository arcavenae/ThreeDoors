//go:build linux

package device

import (
	"fmt"
	"os"
	"strings"
)

type platformMachineIDReader struct{}

func (r *platformMachineIDReader) ReadMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMachineIDUnavailable, err)
	}

	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("%w: /etc/machine-id is empty", ErrMachineIDUnavailable)
	}

	return id, nil
}
