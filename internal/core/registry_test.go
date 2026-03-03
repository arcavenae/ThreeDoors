package core

import (
	"fmt"
	"sort"
	"testing"
)

// stubFactory returns a factory that creates a TextFileProvider.
func stubFactory() AdapterFactory {
	return func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	}
}

// failingFactory returns a factory that always returns an error.
func failingFactory() AdapterFactory {
	return func(config *ProviderConfig) (TaskProvider, error) {
		return nil, fmt.Errorf("init failed")
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		adapter string
		factory AdapterFactory
		wantErr bool
	}{
		{"valid registration", "test-adapter", stubFactory(), false},
		{"empty name", "", stubFactory(), true},
		{"nil factory", "nil-factory", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg := NewRegistry()

			err := reg.Register(tt.adapter, tt.factory)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_Register_DuplicateName(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("dup", stubFactory()); err != nil {
		t.Fatalf("first Register() failed: %v", err)
	}
	if err := reg.Register("dup", stubFactory()); err == nil {
		t.Error("second Register() should have failed for duplicate name")
	}
}

func TestRegistry_ListProviders(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("alpha", stubFactory()); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("beta", stubFactory()); err != nil {
		t.Fatal(err)
	}

	names := reg.ListProviders()
	sort.Strings(names)

	if len(names) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("expected [alpha beta], got %v", names)
	}
}

func TestRegistry_ListProviders_Empty(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	names := reg.ListProviders()
	if len(names) != 0 {
		t.Errorf("expected 0 providers, got %d", len(names))
	}
}

func TestRegistry_GetProvider_NotActive(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("inactive", stubFactory()); err != nil {
		t.Fatal(err)
	}

	_, err := reg.GetProvider("inactive")
	if err == nil {
		t.Error("GetProvider() should fail for uninitialized provider")
	}
}

func TestRegistry_GetProvider_Active(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("active", stubFactory()); err != nil {
		t.Fatal(err)
	}

	cfg := defaultProviderConfig()
	if _, err := reg.InitProvider("active", cfg); err != nil {
		t.Fatalf("InitProvider() failed: %v", err)
	}

	provider, err := reg.GetProvider("active")
	if err != nil {
		t.Fatalf("GetProvider() failed: %v", err)
	}
	if provider == nil {
		t.Error("GetProvider() returned nil provider")
	}
}

func TestRegistry_ActiveProviders(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("one", stubFactory()); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("two", stubFactory()); err != nil {
		t.Fatal(err)
	}

	cfg := defaultProviderConfig()
	if _, err := reg.InitProvider("one", cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.InitProvider("two", cfg); err != nil {
		t.Fatal(err)
	}

	active := reg.ActiveProviders()
	if len(active) != 2 {
		t.Errorf("expected 2 active providers, got %d", len(active))
	}
}

func TestRegistry_ActiveProviders_Empty(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	active := reg.ActiveProviders()
	if len(active) != 0 {
		t.Errorf("expected 0 active providers, got %d", len(active))
	}
}

func TestRegistry_InitProvider_NotRegistered(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_, err := reg.InitProvider("missing", defaultProviderConfig())
	if err == nil {
		t.Error("InitProvider() should fail for unregistered provider")
	}
}

func TestRegistry_InitProvider_FactoryError(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("broken", failingFactory()); err != nil {
		t.Fatal(err)
	}

	_, err := reg.InitProvider("broken", defaultProviderConfig())
	if err == nil {
		t.Error("InitProvider() should fail when factory returns error")
	}

	// Should not be in active providers
	active := reg.ActiveProviders()
	if len(active) != 0 {
		t.Errorf("failed adapter should not be active, got %d active", len(active))
	}
}

func TestRegistry_InitAll_PartialFailure(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("good", stubFactory()); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("bad", failingFactory()); err != nil {
		t.Fatal(err)
	}

	count := reg.InitAll(defaultProviderConfig())
	if count != 1 {
		t.Errorf("expected 1 successful init, got %d", count)
	}

	active := reg.ActiveProviders()
	if len(active) != 1 {
		t.Errorf("expected 1 active provider, got %d", len(active))
	}
}

func TestRegistry_IsRegistered(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	if err := reg.Register("exists", stubFactory()); err != nil {
		t.Fatal(err)
	}

	if !reg.IsRegistered("exists") {
		t.Error("IsRegistered() should return true for registered adapter")
	}
	if reg.IsRegistered("nope") {
		t.Error("IsRegistered() should return false for unregistered adapter")
	}
}
