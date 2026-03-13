package device

// MachineIDReader reads a platform-specific machine identifier.
type MachineIDReader interface {
	ReadMachineID() (string, error)
}

// NewPlatformMachineIDReader returns the platform-appropriate MachineIDReader.
func NewPlatformMachineIDReader() MachineIDReader {
	return &platformMachineIDReader{}
}
