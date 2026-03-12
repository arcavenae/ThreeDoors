package themes

import "testing"

func TestRegisterAndGet(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	theme := &DoorTheme{
		Name:        "test-theme",
		Description: "A test theme",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
		MinWidth: 10,
	}

	reg.Register(theme)

	got, ok := reg.Get("test-theme")
	if !ok {
		t.Fatal("expected to find registered theme")
	}
	if got.Name != "test-theme" {
		t.Errorf("got name %q, want %q", got.Name, "test-theme")
	}
	if got.Description != "A test theme" {
		t.Errorf("got description %q, want %q", got.Description, "A test theme")
	}
	if got.MinWidth != 10 {
		t.Errorf("got MinWidth %d, want 10", got.MinWidth)
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("expected not-found for unregistered theme")
	}
}

func TestRegisterOverwrite(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	theme1 := &DoorTheme{
		Name:        "dup",
		Description: "first",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return "v1"
		},
	}
	theme2 := &DoorTheme{
		Name:        "dup",
		Description: "second",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return "v2"
		},
	}

	reg.Register(theme1)
	reg.Register(theme2)

	got, ok := reg.Get("dup")
	if !ok {
		t.Fatal("expected to find theme")
	}
	if got.Description != "second" {
		t.Errorf("got description %q, want %q (latest registration wins)", got.Description, "second")
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	t.Parallel()

	reg := NewDefaultRegistry()

	// All built-in themes should be registered
	expected := []string{"autumn", "classic", "modern", "scifi", "shoji", "spring", "summer", "winter"}
	names := reg.Names()
	if len(names) != len(expected) {
		t.Fatalf("got %d themes, want %d: %v", len(names), len(expected), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}

	// Each theme should have a working Render function
	for _, name := range expected {
		theme, ok := reg.Get(name)
		if !ok {
			t.Fatalf("theme %q not found", name)
		}
		output := theme.Render("Test", 30, 0, false, "", 0.0)
		if output == "" {
			t.Errorf("theme %q rendered empty output", name)
		}
	}
}

func TestSeasonalNames(t *testing.T) {
	t.Parallel()

	reg := NewDefaultRegistry()
	names := reg.SeasonalNames()

	expected := []string{"autumn", "spring", "summer", "winter"}
	if len(names) != len(expected) {
		t.Fatalf("got %d seasonal names, want %d: %v", len(names), len(expected), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestNonSeasonalNames(t *testing.T) {
	t.Parallel()

	reg := NewDefaultRegistry()
	names := reg.NonSeasonalNames()

	expected := []string{"classic", "modern", "scifi", "shoji"}
	if len(names) != len(expected) {
		t.Fatalf("got %d non-seasonal names, want %d: %v", len(names), len(expected), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestSeasonalNamesEmptyRegistry(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	reg.Register(&DoorTheme{
		Name: "plain",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
	})

	names := reg.SeasonalNames()
	if len(names) != 0 {
		t.Errorf("expected no seasonal names, got %v", names)
	}

	nonSeasonal := reg.NonSeasonalNames()
	if len(nonSeasonal) != 1 || nonSeasonal[0] != "plain" {
		t.Errorf("expected [plain], got %v", nonSeasonal)
	}
}

func TestNames(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	reg.Register(&DoorTheme{
		Name: "alpha",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
	})
	reg.Register(&DoorTheme{
		Name: "beta",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
	})

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}

	// Names should be sorted
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("got names %v, want [alpha beta]", names)
	}
}
