package device

import "errors"

// ErrMachineIDUnavailable is returned when the platform machine ID cannot be read.
var ErrMachineIDUnavailable = errors.New("machine-id unavailable")

// ErrInvalidDeviceID is returned when a string is not a valid UUID.
var ErrInvalidDeviceID = errors.New("invalid device ID")

// ErrDeviceNotFound is returned when a device ID doesn't exist in the registry.
var ErrDeviceNotFound = errors.New("device not found")

// ErrDeviceAlreadyExists is returned when attempting to register a device that already exists.
var ErrDeviceAlreadyExists = errors.New("device already exists")
