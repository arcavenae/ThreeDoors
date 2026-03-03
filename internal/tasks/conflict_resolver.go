package tasks

import "fmt"

// ConflictChoice represents a user's resolution choice for a sync conflict.
type ConflictChoice string

const (
	ChoiceKeepLocal  ConflictChoice = "local"
	ChoiceKeepRemote ConflictChoice = "remote"
	ChoiceKeepBoth   ConflictChoice = "both"
)

// InteractiveConflict wraps a Conflict with display metadata for the TUI.
type InteractiveConflict struct {
	Conflict   Conflict
	Resolved   bool
	Choice     ConflictChoice
	ResultTask *Task // the winning task (or both for "keep both")
}

// ConflictSet holds a batch of conflicts from a single sync operation.
type ConflictSet struct {
	Provider  string
	Conflicts []InteractiveConflict
	Current   int // index of the conflict being viewed
}

// NewConflictSet creates a ConflictSet from detected conflicts.
func NewConflictSet(provider string, conflicts []Conflict) *ConflictSet {
	interactive := make([]InteractiveConflict, len(conflicts))
	for i, c := range conflicts {
		interactive[i] = InteractiveConflict{
			Conflict: c,
		}
	}
	return &ConflictSet{
		Provider:  provider,
		Conflicts: interactive,
	}
}

// CurrentConflict returns the conflict currently being viewed.
// Returns nil if all conflicts have been resolved or the set is empty.
func (cs *ConflictSet) CurrentConflict() *InteractiveConflict {
	if cs.Current >= len(cs.Conflicts) {
		return nil
	}
	return &cs.Conflicts[cs.Current]
}

// Resolve applies a user choice to the current conflict and advances to the next.
func (cs *ConflictSet) Resolve(choice ConflictChoice) error {
	if cs.Current >= len(cs.Conflicts) {
		return fmt.Errorf("no conflict to resolve")
	}

	ic := &cs.Conflicts[cs.Current]
	ic.Choice = choice
	ic.Resolved = true

	switch choice {
	case ChoiceKeepLocal:
		ic.ResultTask = ic.Conflict.LocalTask
	case ChoiceKeepRemote:
		ic.ResultTask = ic.Conflict.RemoteTask
	case ChoiceKeepBoth:
		// "Keep both" keeps the local task and creates a copy of remote
		ic.ResultTask = ic.Conflict.LocalTask
	default:
		return fmt.Errorf("unknown conflict choice: %s", choice)
	}

	cs.Current++
	return nil
}

// AllResolved returns true if every conflict in the set has been resolved.
func (cs *ConflictSet) AllResolved() bool {
	return cs.Current >= len(cs.Conflicts)
}

// Resolutions returns Resolution values for applying to the TaskPool.
func (cs *ConflictSet) Resolutions() []Resolution {
	var resolutions []Resolution
	for _, ic := range cs.Conflicts {
		if !ic.Resolved {
			continue
		}
		r := Resolution{
			TaskID:      ic.Conflict.LocalTask.ID,
			WinningTask: ic.ResultTask,
		}
		switch ic.Choice {
		case ChoiceKeepLocal:
			r.Winner = "local"
			r.LocalOverridden = false
			r.Message = fmt.Sprintf("Kept local version of '%s'", ic.Conflict.LocalTask.Text)
		case ChoiceKeepRemote:
			r.Winner = "remote"
			r.LocalOverridden = true
			r.Message = fmt.Sprintf("Accepted remote version of '%s'", ic.Conflict.RemoteTask.Text)
		case ChoiceKeepBoth:
			r.Winner = "both"
			r.LocalOverridden = false
			r.Message = fmt.Sprintf("Kept both versions of '%s'", ic.Conflict.LocalTask.Text)
		}
		resolutions = append(resolutions, r)
	}
	return resolutions
}

// UnresolvedCount returns the number of conflicts not yet resolved.
func (cs *ConflictSet) UnresolvedCount() int {
	count := 0
	for _, ic := range cs.Conflicts {
		if !ic.Resolved {
			count++
		}
	}
	return count
}
