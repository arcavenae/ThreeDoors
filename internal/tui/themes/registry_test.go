package themes

import "testing"

func TestRegisterAndGet(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	theme := &DoorTheme{
		Name:        "test-theme",
		Description: "A test theme",
		Render: func(content string, width int, selected bool) string {
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
		Render: func(content string, width int, selected bool) string {
			return "v1"
		},
	}
	theme2 := &DoorTheme{
		Name:        "dup",
		Description: "second",
		Render: func(content string, width int, selected bool) string {
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

func TestNames(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	reg.Register(&DoorTheme{
		Name: "alpha",
		Render: func(content string, width int, selected bool) string {
			return content
		},
	})
	reg.Register(&DoorTheme{
		Name: "beta",
		Render: func(content string, width int, selected bool) string {
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
