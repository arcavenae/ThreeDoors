package reminders

import (
	"runtime"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestNewFactory_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	factory := NewFactory()
	if factory == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestNewFactory_NonDarwinReturnsError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "darwin" {
		t.Skip("this test validates non-darwin behavior")
	}

	factory := NewFactory()
	cfg := &core.ProviderConfig{
		Providers: []core.ProviderEntry{
			{Name: "reminders"},
		},
	}

	_, err := factory(cfg)
	if err == nil {
		t.Fatal("expected error on non-darwin platform")
		return
	}

	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestNewFactory_DarwinCreatesProvider(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "darwin" {
		t.Skip("this test requires macOS")
	}

	factory := NewFactory()
	cfg := &core.ProviderConfig{
		Providers: []core.ProviderEntry{
			{
				Name:     "reminders",
				Settings: map[string]string{"lists": "Work,Personal"},
			},
		},
	}

	provider, err := factory(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if provider == nil {
		t.Fatal("factory returned nil provider")
		return
	}

	rp, ok := provider.(*RemindersProvider)
	if !ok {
		t.Fatalf("expected *RemindersProvider, got %T", provider)
	}
	if rp.Name() != "reminders" {
		t.Errorf("Name() = %q, want %q", rp.Name(), "reminders")
	}
}
