package device

import (
	"fmt"

	"github.com/google/uuid"
)

// DeviceID is a named string type that holds a validated UUID (v4 or v5).
type DeviceID string

// NewDeviceID validates the input as a UUID and returns a DeviceID.
func NewDeviceID(s string) (DeviceID, error) {
	if s == "" {
		return "", fmt.Errorf("%w: empty string", ErrInvalidDeviceID)
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidDeviceID, s)
	}
	return DeviceID(parsed.String()), nil
}

// String returns the string representation of the DeviceID.
func (d DeviceID) String() string {
	return string(d)
}

// IsValid returns true if the DeviceID contains a valid UUID.
func (d DeviceID) IsValid() bool {
	if d == "" {
		return false
	}
	_, err := uuid.Parse(string(d))
	return err == nil
}
