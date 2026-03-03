package tasks

import (
	"fmt"
	"log"
	"sync"
)

// AdapterFactory creates a TaskProvider from the given configuration.
// If initialization fails, it should return an error rather than panicking.
type AdapterFactory func(config *ProviderConfig) (TaskProvider, error)

// Registry manages adapter registration and runtime discovery of TaskProvider implementations.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]AdapterFactory
	active    map[string]TaskProvider
}

// NewRegistry creates a new empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]AdapterFactory),
		active:    make(map[string]TaskProvider),
	}
}

// Register adds an adapter factory to the registry under the given name.
// Returns an error if the name is empty or already registered.
func (r *Registry) Register(name string, factory AdapterFactory) error {
	if name == "" {
		return fmt.Errorf("register adapter: name must not be empty")
	}
	if factory == nil {
		return fmt.Errorf("register adapter %q: factory must not be nil", name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("register adapter %q: already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// ListProviders returns the names of all registered adapters in no particular order.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// GetProvider returns the active provider instance for the given name.
// If the provider has not been initialized yet, it returns an error.
func (r *Registry) GetProvider(name string) (TaskProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.active[name]
	if !ok {
		return nil, fmt.Errorf("get provider %q: not active", name)
	}
	return provider, nil
}

// ActiveProviders returns all currently initialized and active provider instances.
func (r *Registry) ActiveProviders() []TaskProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]TaskProvider, 0, len(r.active))
	for _, p := range r.active {
		providers = append(providers, p)
	}
	return providers
}

// InitProvider initializes a registered adapter with the given config and marks it as active.
// If initialization fails, a warning is logged and the error is returned,
// but the registry remains usable for other adapters.
func (r *Registry) InitProvider(name string, config *ProviderConfig) (TaskProvider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("init provider %q: not registered", name)
	}

	provider, err := factory(config)
	if err != nil {
		log.Printf("Warning: adapter %q failed to initialize: %v", name, err)
		return nil, fmt.Errorf("init provider %q: %w", name, err)
	}

	r.active[name] = provider
	return provider, nil
}

// InitAll attempts to initialize all registered adapters with the given config.
// Adapters that fail to initialize are logged as warnings but do not prevent
// other adapters from initializing. Returns the number of successfully initialized adapters.
func (r *Registry) InitAll(config *ProviderConfig) int {
	r.mu.Lock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	r.mu.Unlock()

	initialized := 0
	for _, name := range names {
		if _, err := r.InitProvider(name, config); err != nil {
			log.Printf("Warning: skipping adapter %q: %v", name, err)
			continue
		}
		initialized++
	}
	return initialized
}

// IsRegistered returns true if an adapter with the given name has been registered.
func (r *Registry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.factories[name]
	return ok
}

// defaultRegistry is the global adapter registry used by the application.
var defaultRegistry = NewRegistry()

// DefaultRegistry returns the global adapter registry.
func DefaultRegistry() *Registry {
	return defaultRegistry
}
