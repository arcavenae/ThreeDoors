//go:build !darwin && !linux

package device

type platformMachineIDReader struct{}

func (r *platformMachineIDReader) ReadMachineID() (string, error) {
	return "", ErrMachineIDUnavailable
}
