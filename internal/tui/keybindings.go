package tui

// KeyBinding represents a single key binding with display metadata.
type KeyBinding struct {
	Key         string
	Description string
	Priority    int // 1=always show in bar, 2=show if space, 3=overlay only
}

// KeyBindingGroup is a named collection of related key bindings.
type KeyBindingGroup struct {
	Name     string
	Bindings []KeyBinding
}

// Priority constants for keybinding display.
const (
	PriorityAlways  = 1 // Always show in the bar
	PriorityIfSpace = 2 // Show in bar if space permits
	PriorityOverlay = 3 // Overlay only
)

// BarContext carries all context needed for keybinding bar rendering.
// This allows the bar to adapt to sub-modes (e.g., DetailView text input)
// without growing function parameter lists.
type BarContext struct {
	Mode         ViewMode
	DoorSelected bool
	DetailMode   DetailViewMode
	CommandMode  bool
}

// viewKeyBindings returns categorized bindings for the given view mode.
// The doorSelected parameter only matters for ViewDoors — when true, it
// returns the door-selected binding set instead of the default set.
func viewKeyBindings(mode ViewMode, doorSelected bool) []KeyBindingGroup {
	switch mode {
	case ViewDoors:
		if doorSelected {
			return doorsSelectedBindings()
		}
		return doorsBindings()
	case ViewDetail:
		return detailBindings()
	case ViewMood:
		return moodBindings()
	case ViewSearch:
		return searchBindings()
	case ViewHealth:
		return healthBindings()
	case ViewAddTask:
		return addTaskBindings()
	case ViewValuesGoals:
		return valuesBindings()
	case ViewFeedback:
		return feedbackBindings()
	case ViewImprovement:
		return improvementBindings()
	case ViewNextSteps:
		return nextStepsBindings()
	case ViewAvoidancePrompt:
		return avoidanceBindings()
	case ViewInsights:
		return insightsBindings()
	case ViewOnboarding:
		return onboardingBindings()
	case ViewConflict:
		return conflictBindings()
	case ViewSyncLog:
		return syncLogBindings()
	case ViewThemePicker:
		return themePickerBindings()
	case ViewDevQueue:
		return devQueueBindings()
	case ViewProposals:
		return proposalsBindings()
	case ViewHelp:
		return helpBindings()
	case ViewDeferred:
		return deferredBindings()
	case ViewSnooze:
		return snoozeBindings()
	default:
		return nil
	}
}

// barBindings returns only priority-1 bindings for the given view mode.
func barBindings(mode ViewMode, doorSelected bool) []KeyBinding {
	groups := viewKeyBindings(mode, doorSelected)
	var result []KeyBinding
	for _, g := range groups {
		for _, b := range g.Bindings {
			if b.Priority == PriorityAlways {
				result = append(result, b)
			}
		}
	}
	return result
}

// contextBarBindings returns priority-1 bindings using full BarContext,
// supporting sub-mode awareness (e.g., DetailView text input modes).
func contextBarBindings(ctx BarContext) []KeyBinding {
	// Sub-mode overrides for specific view modes.
	switch ctx.Mode {
	case ViewDetail:
		switch ctx.DetailMode {
		case DetailModeBlockerInput, DetailModeExpandInput:
			return detailTextInputBarBindings()
		case DetailModeDispatchConfirm:
			return detailConfirmBarBindings()
		case DetailModeLinkSelect, DetailModeDepAdd:
			return detailSelectBarBindings()
		case DetailModeLinkBrowse, DetailModeDepBrowse:
			return detailBrowseBarBindings()
		}
	case ViewSearch:
		if ctx.CommandMode {
			return commandModeBarBindings()
		}
	}
	return barBindings(ctx.Mode, ctx.DoorSelected)
}

// allKeyBindingGroups returns all bindings across all views, organized
// by category for the full overlay display.
func allKeyBindingGroups() []KeyBindingGroup {
	seen := make(map[string]map[string]bool) // group name → "key:desc" → seen
	var groups []KeyBindingGroup

	addUnique := func(groupName string, bindings []KeyBinding) {
		if seen[groupName] == nil {
			seen[groupName] = make(map[string]bool)
		}
		var unique []KeyBinding
		for _, b := range bindings {
			key := b.Key + ":" + b.Description
			if !seen[groupName][key] {
				seen[groupName][key] = true
				unique = append(unique, b)
			}
		}
		// Find existing group or create new.
		found := false
		for i := range groups {
			if groups[i].Name == groupName {
				groups[i].Bindings = append(groups[i].Bindings, unique...)
				found = true
				break
			}
		}
		if !found && len(unique) > 0 {
			groups = append(groups, KeyBindingGroup{Name: groupName, Bindings: unique})
		}
	}

	// Collect from all views (both door states).
	allModes := []struct {
		mode         ViewMode
		doorSelected bool
	}{
		{ViewDoors, false},
		{ViewDoors, true},
		{ViewDetail, false},
		{ViewMood, false},
		{ViewSearch, false},
		{ViewHealth, false},
		{ViewAddTask, false},
		{ViewValuesGoals, false},
		{ViewFeedback, false},
		{ViewImprovement, false},
		{ViewNextSteps, false},
		{ViewAvoidancePrompt, false},
		{ViewInsights, false},
		{ViewOnboarding, false},
		{ViewConflict, false},
		{ViewSyncLog, false},
		{ViewThemePicker, false},
		{ViewDevQueue, false},
		{ViewProposals, false},
		{ViewHelp, false},
		{ViewDeferred, false},
		{ViewSnooze, false},
	}

	for _, m := range allModes {
		for _, g := range viewKeyBindings(m.mode, m.doorSelected) {
			addUnique(g.Name, g.Bindings)
		}
	}

	// Append commands section — always last.
	cmdGroup := commandBindingGroup()
	addUnique(cmdGroup.Name, cmdGroup.Bindings)

	return groups
}

// --- Per-view binding definitions ---

func doorsBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "a/w/d", Description: "select door", Priority: PriorityAlways},
			{Key: "s", Description: "re-roll", Priority: PriorityAlways},
			{Key: "←/↑/→", Description: "select door", Priority: PriorityOverlay},
			{Key: "↓", Description: "re-roll", Priority: PriorityOverlay},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "n", Description: "add task", Priority: PriorityAlways},
			{Key: ":", Description: "command", Priority: PriorityAlways},
			{Key: "/", Description: "search", Priority: PriorityIfSpace},
			{Key: "m", Description: "mood", Priority: PriorityIfSpace},
			{Key: "S", Description: "proposals", Priority: PriorityOverlay},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
			{Key: "q", Description: "quit", Priority: PriorityAlways},
		}},
	}
}

func doorsSelectedBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "enter", Description: "confirm", Priority: PriorityAlways},
			{Key: "space", Description: "confirm", Priority: PriorityOverlay},
			{Key: "a/w/d", Description: "change door", Priority: PriorityAlways},
			{Key: "esc", Description: "deselect", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "n", Description: "feedback", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func detailBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "c", Description: "complete", Priority: PriorityAlways},
			{Key: "b", Description: "blocked", Priority: PriorityAlways},
			{Key: "i", Description: "in progress", Priority: PriorityIfSpace},
			{Key: "e", Description: "expand", Priority: PriorityAlways},
			{Key: "f", Description: "fork", Priority: PriorityIfSpace},
			{Key: "p", Description: "procrastinate", Priority: PriorityOverlay},
			{Key: "r", Description: "rework", Priority: PriorityOverlay},
			{Key: "l", Description: "link", Priority: PriorityOverlay},
			{Key: "m", Description: "mood", Priority: PriorityOverlay},
			{Key: "g", Description: "decompose", Priority: PriorityOverlay},
			{Key: "d", Description: "dismiss dup", Priority: PriorityOverlay},
			{Key: "y", Description: "merge dup", Priority: PriorityOverlay},
			{Key: "x", Description: "dispatch", Priority: PriorityOverlay},
			{Key: "z", Description: "snooze", Priority: PriorityOverlay},
			{Key: "u", Description: "undo complete", Priority: PriorityOverlay},
			{Key: "+", Description: "add dependency", Priority: PriorityOverlay},
			{Key: "-", Description: "remove dependency", Priority: PriorityOverlay},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func moodBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "1-6", Description: "select mood", Priority: PriorityAlways},
			{Key: "7", Description: "custom mood", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func searchBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "esc", Description: "close", Priority: PriorityAlways},
			{Key: "↑/↓", Description: "navigate", Priority: PriorityAlways},
			{Key: "enter", Description: "select", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func healthBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func addTaskBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "enter", Description: "submit", Priority: PriorityAlways},
			{Key: "esc", Description: "cancel", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func valuesBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "esc", Description: "save & back", Priority: PriorityAlways},
			{Key: "↑/↓", Description: "navigate", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "a", Description: "add value", Priority: PriorityAlways},
			{Key: "d", Description: "delete", Priority: PriorityIfSpace},
			{Key: "enter", Description: "edit", Priority: PriorityIfSpace},
			{Key: "K/J", Description: "reorder", Priority: PriorityOverlay},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func feedbackBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "1", Description: "blocked", Priority: PriorityAlways},
			{Key: "2", Description: "not now", Priority: PriorityAlways},
			{Key: "3", Description: "needs breakdown", Priority: PriorityIfSpace},
			{Key: "4", Description: "custom", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func improvementBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "enter", Description: "submit", Priority: PriorityAlways},
			{Key: "esc", Description: "skip", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func nextStepsBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "dismiss", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func avoidanceBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "r", Description: "reconsider", Priority: PriorityAlways},
			{Key: "b", Description: "breakdown", Priority: PriorityAlways},
			{Key: "d", Description: "defer", Priority: PriorityIfSpace},
			{Key: "a", Description: "archive", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func insightsBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func onboardingBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "enter", Description: "continue", Priority: PriorityAlways},
			{Key: "a/w/d", Description: "select door", Priority: PriorityIfSpace},
			{Key: "s", Description: "re-roll", Priority: PriorityOverlay},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func conflictBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "l", Description: "keep local", Priority: PriorityAlways},
			{Key: "r", Description: "keep remote", Priority: PriorityAlways},
			{Key: "b", Description: "keep both", Priority: PriorityAlways},
			{Key: "esc", Description: "cancel", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func syncLogBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "esc", Description: "back", Priority: PriorityAlways},
			{Key: "j/k", Description: "scroll", Priority: PriorityAlways},
			{Key: "pgdn/pgup", Description: "page", Priority: PriorityIfSpace},
			{Key: "space", Description: "page down", Priority: PriorityOverlay},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func themePickerBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "←/→", Description: "browse", Priority: PriorityAlways},
			{Key: "enter", Description: "select", Priority: PriorityAlways},
			{Key: "q/esc", Description: "cancel", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func devQueueBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
			{Key: "j/k", Description: "navigate", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "y", Description: "approve", Priority: PriorityAlways},
			{Key: "n", Description: "reject", Priority: PriorityAlways},
			{Key: "K", Description: "kill", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func proposalsBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "j/k", Description: "navigate", Priority: PriorityAlways},
			{Key: "esc", Description: "back", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "enter", Description: "approve", Priority: PriorityAlways},
			{Key: "del", Description: "reject", Priority: PriorityIfSpace},
			{Key: "tab", Description: "skip", Priority: PriorityIfSpace},
			{Key: "ctrl+a", Description: "approve all", Priority: PriorityOverlay},
			{Key: "p", Description: "preview", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func helpBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
			{Key: "j/k", Description: "scroll", Priority: PriorityAlways},
			{Key: "pgdn/pgup", Description: "page", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func deferredBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "q/esc", Description: "back", Priority: PriorityAlways},
			{Key: "j/k", Description: "navigate", Priority: PriorityAlways},
		}},
		{Name: "Actions", Bindings: []KeyBinding{
			{Key: "u", Description: "unsnooze", Priority: PriorityAlways},
			{Key: "e", Description: "edit date", Priority: PriorityIfSpace},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

func snoozeBindings() []KeyBindingGroup {
	return []KeyBindingGroup{
		{Name: "Navigation", Bindings: []KeyBinding{
			{Key: "↑/↓", Description: "navigate", Priority: PriorityAlways},
			{Key: "enter", Description: "select", Priority: PriorityAlways},
			{Key: "esc", Description: "cancel", Priority: PriorityAlways},
		}},
		{Name: "Display", Bindings: []KeyBinding{
			{Key: "?", Description: "help", Priority: PriorityAlways},
		}},
	}
}

// --- Sub-mode bar bindings for DetailView ---
// These replace the normal detail bindings when the view is in a sub-mode.

func detailTextInputBarBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Description: "submit", Priority: PriorityAlways},
		{Key: "esc", Description: "cancel", Priority: PriorityAlways},
	}
}

func detailConfirmBarBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "y", Description: "confirm", Priority: PriorityAlways},
		{Key: "n", Description: "cancel", Priority: PriorityAlways},
	}
}

func detailSelectBarBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/↓", Description: "navigate", Priority: PriorityAlways},
		{Key: "enter", Description: "select", Priority: PriorityAlways},
		{Key: "esc", Description: "cancel", Priority: PriorityAlways},
	}
}

func detailBrowseBarBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/↓", Description: "navigate", Priority: PriorityAlways},
		{Key: "esc", Description: "back", Priority: PriorityAlways},
	}
}

func commandModeBarBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Description: "run command", Priority: PriorityAlways},
		{Key: "esc", Description: "cancel", Priority: PriorityAlways},
	}
}

// --- Commands section for overlay ---

// commandBindingGroup returns a KeyBindingGroup listing all : commands.
func commandBindingGroup() KeyBindingGroup {
	return KeyBindingGroup{
		Name: "Commands",
		Bindings: []KeyBinding{
			{Key: ":add", Description: "add task", Priority: PriorityOverlay},
			{Key: ":add-ctx", Description: "add task with context", Priority: PriorityOverlay},
			{Key: ":mood", Description: "capture mood", Priority: PriorityOverlay},
			{Key: ":stats", Description: "session statistics", Priority: PriorityOverlay},
			{Key: ":health", Description: "health check", Priority: PriorityOverlay},
			{Key: ":dashboard", Description: "insights dashboard", Priority: PriorityOverlay},
			{Key: ":goals", Description: "values & goals", Priority: PriorityOverlay},
			{Key: ":synclog", Description: "sync log", Priority: PriorityOverlay},
			{Key: ":theme", Description: "theme picker", Priority: PriorityOverlay},
			{Key: ":deferred", Description: "deferred tasks", Priority: PriorityOverlay},
			{Key: ":devqueue", Description: "dev queue", Priority: PriorityOverlay},
			{Key: ":suggestions", Description: "task proposals", Priority: PriorityOverlay},
			{Key: ":help", Description: "help view", Priority: PriorityOverlay},
			{Key: ":quit", Description: "exit app", Priority: PriorityOverlay},
		},
	}
}

// Audit notes — intentionally omitted bindings:
// - ctrl+c: handled by Bubbletea framework as universal quit, not view-specific
// - h: global bar toggle, not a view action — toggling the bar from within the bar
//   would be circular; users discover 'h' from the overlay or documentation
// - backspace: standard text editing key in input modes, not a keybinding per se
// - single-character typing in input modes: not keybindings, just text entry
