//go:build !darwin

package reminders

import (
	"fmt"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// NewFactory returns an AdapterFactory that always returns an error on
// non-macOS platforms, since Apple Reminders requires macOS and osascript.
func NewFactory() core.AdapterFactory {
	return func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return nil, fmt.Errorf("reminders adapter requires macOS (darwin): Apple Reminders is not available on this platform")
	}
}
