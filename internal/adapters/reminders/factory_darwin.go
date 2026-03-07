//go:build darwin

package reminders

import (
	"github.com/arcaven/ThreeDoors/internal/core"
)

// NewFactory returns an AdapterFactory that creates a RemindersProvider
// configured from the provider settings. Only available on macOS.
func NewFactory() core.AdapterFactory {
	return func(config *core.ProviderConfig) (core.TaskProvider, error) {
		rc := ParseRemindersConfig(config)
		executor := &OSAScriptExecutor{}
		return NewRemindersProvider(executor, rc.Lists), nil
	}
}
