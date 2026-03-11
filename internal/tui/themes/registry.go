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

// NewDefaultRegistry creates a registry pre-populated with all built-in themes.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewClassicTheme())
	r.Register(NewModernTheme())
	r.Register(NewSciFiTheme())
	r.Register(NewShojiTheme())
	r.Register(NewWinterTheme())
	r.Register(NewSpringTheme())
	r.Register(NewSummerTheme())
	r.Register(NewAutumnTheme())
	return r
}

// GetBySeason returns the first theme matching the given season, or false
// if no seasonal theme is registered for that season.
func (r *Registry) GetBySeason(season string) (*DoorTheme, bool) {
	if season == "" {
		return nil, false
	}
	for _, t := range r.themes {
		if t.Season == season {
			return t, true
		}
	}
	return nil, false
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

// SeasonalNames returns sorted names of themes where Season is non-empty.
func (r *Registry) SeasonalNames() []string {
	var names []string
	for name, theme := range r.themes {
		if theme.Season != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// NonSeasonalNames returns sorted names of themes where Season is empty.
func (r *Registry) NonSeasonalNames() []string {
	var names []string
	for name, theme := range r.themes {
		if theme.Season == "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
