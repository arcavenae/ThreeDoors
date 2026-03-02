package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// tagViewState tracks which step of the tag editing flow we're on.
type tagViewState int

const (
	tagSelectingField tagViewState = iota
	tagSelectingValue
)

// tagField identifies which categorization field is being edited.
type tagField int

const (
	tagFieldType tagField = iota
	tagFieldEffort
	tagFieldLocation
	tagFieldDone
)

// TagView handles editing task categorization fields.
type TagView struct {
	task         *tasks.Task
	state        tagViewState
	fieldIndex   int
	valueIndex   int
	currentField tagField
	width        int

	// Snapshot of original values for cancel
	origType     tasks.TaskType
	origEffort   tasks.TaskEffort
	origLocation tasks.TaskLocation
}

var fieldLabels = []string{"Type", "Effort", "Location", "Done"}

var typeOptions = []struct {
	label string
	value tasks.TaskType
}{
	{"(none)", ""},
	{"Creative", tasks.TypeCreative},
	{"Administrative", tasks.TypeAdministrative},
	{"Technical", tasks.TypeTechnical},
	{"Physical", tasks.TypePhysical},
}

var effortOptions = []struct {
	label string
	value tasks.TaskEffort
}{
	{"(none)", ""},
	{"Quick Win", tasks.EffortQuickWin},
	{"Medium", tasks.EffortMedium},
	{"Deep Work", tasks.EffortDeepWork},
}

var locationOptions = []struct {
	label string
	value tasks.TaskLocation
}{
	{"(none)", ""},
	{"Home", tasks.LocationHome},
	{"Work", tasks.LocationWork},
	{"Errands", tasks.LocationErrands},
	{"Anywhere", tasks.LocationAnywhere},
}

// NewTagView creates a new TagView for the given task.
// Returns nil if task is nil.
func NewTagView(task *tasks.Task) *TagView {
	if task == nil {
		return nil
	}
	return &TagView{
		task:         task,
		state:        tagSelectingField,
		origType:     task.Type,
		origEffort:   task.Effort,
		origLocation: task.Location,
	}
}

// SetWidth sets the terminal width for rendering.
func (tv *TagView) SetWidth(w int) {
	tv.width = w
}

// Update handles messages for the tag view.
func (tv *TagView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			// Restore original values
			tv.task.Type = tv.origType
			tv.task.Effort = tv.origEffort
			tv.task.Location = tv.origLocation
			return func() tea.Msg { return TagCancelledMsg{} }

		case tea.KeyUp:
			if tv.state == tagSelectingField && tv.fieldIndex > 0 {
				tv.fieldIndex--
			} else if tv.state == tagSelectingValue && tv.valueIndex > 0 {
				tv.valueIndex--
			}
			return nil

		case tea.KeyDown:
			if tv.state == tagSelectingField && tv.fieldIndex < len(fieldLabels)-1 {
				tv.fieldIndex++
			} else if tv.state == tagSelectingValue {
				max := tv.maxValueIndex()
				if tv.valueIndex < max {
					tv.valueIndex++
				}
			}
			return nil

		case tea.KeyEnter:
			if tv.state == tagSelectingField {
				tv.currentField = tagField(tv.fieldIndex)
				if tv.currentField == tagFieldDone {
					return func() tea.Msg { return TagUpdatedMsg{Task: tv.task} }
				}
				tv.state = tagSelectingValue
				tv.valueIndex = tv.currentValueIndex()
				return nil
			}
			// tagSelectingValue — apply selection
			tv.applyValue()
			tv.state = tagSelectingField
			return nil
		}
	}
	return nil
}

func (tv *TagView) maxValueIndex() int {
	switch tv.currentField {
	case tagFieldType:
		return len(typeOptions) - 1
	case tagFieldEffort:
		return len(effortOptions) - 1
	case tagFieldLocation:
		return len(locationOptions) - 1
	}
	return 0
}

func (tv *TagView) currentValueIndex() int {
	switch tv.currentField {
	case tagFieldType:
		for i, opt := range typeOptions {
			if opt.value == tv.task.Type {
				return i
			}
		}
	case tagFieldEffort:
		for i, opt := range effortOptions {
			if opt.value == tv.task.Effort {
				return i
			}
		}
	case tagFieldLocation:
		for i, opt := range locationOptions {
			if opt.value == tv.task.Location {
				return i
			}
		}
	}
	return 0
}

func (tv *TagView) applyValue() {
	switch tv.currentField {
	case tagFieldType:
		tv.task.Type = typeOptions[tv.valueIndex].value
	case tagFieldEffort:
		tv.task.Effort = effortOptions[tv.valueIndex].value
	case tagFieldLocation:
		tv.task.Location = locationOptions[tv.valueIndex].value
	}
}

// View renders the tag view.
func (tv *TagView) View() string {
	s := strings.Builder{}
	s.WriteString(headerStyle.Render("ThreeDoors - Edit Tags"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render(fmt.Sprintf("Task: %s", tv.task.Text)))
	s.WriteString("\n\n")

	if tv.state == tagSelectingField {
		s.WriteString("Select field to edit:\n\n")
		for i, label := range fieldLabels {
			cursor := "  "
			if i == tv.fieldIndex {
				cursor = "> "
			}
			current := ""
			switch tagField(i) {
			case tagFieldType:
				if tv.task.Type != "" {
					current = fmt.Sprintf(" [%s]", tv.task.Type)
				}
			case tagFieldEffort:
				if tv.task.Effort != "" {
					current = fmt.Sprintf(" [%s]", tv.task.Effort)
				}
			case tagFieldLocation:
				if tv.task.Location != "" {
					current = fmt.Sprintf(" [%s]", tv.task.Location)
				}
			}
			fmt.Fprintf(&s, "%s%s%s\n", cursor, label, current)
		}
	} else {
		fmt.Fprintf(&s, "Select %s:\n\n", fieldLabels[tv.currentField])
		var opts []string
		switch tv.currentField {
		case tagFieldType:
			for _, opt := range typeOptions {
				opts = append(opts, opt.label)
			}
		case tagFieldEffort:
			for _, opt := range effortOptions {
				opts = append(opts, opt.label)
			}
		case tagFieldLocation:
			for _, opt := range locationOptions {
				opts = append(opts, opt.label)
			}
		}
		for i, label := range opts {
			cursor := "  "
			if i == tv.valueIndex {
				cursor = "> "
			}
			fmt.Fprintf(&s, "%s%s\n", cursor, label)
		}
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓ navigate | Enter select | Esc cancel"))
	return s.String()
}
