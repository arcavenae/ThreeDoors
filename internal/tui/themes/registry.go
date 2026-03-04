package themes

import "sort"

// Registry holds available door themes keyed by name.
type Registry struct {
	themes map[string]*DoorTheme
}

// NewRegistry creates an empty theme registry.
func NewRegistry() *Registry {
	return &Registry{
		themes: make(map[string]*DoorTheme),
	}
}

// Register adds a theme to the registry. If a theme with the same name
// already exists, it is replaced.
func (r *Registry) Register(theme *DoorTheme) {
	r.themes[theme.Name] = theme
}

// Get returns the theme with the given name, or false if not found.
func (r *Registry) Get(name string) (*DoorTheme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

// Names returns sorted names of all registered themes.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.themes))
	for name := range r.themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
