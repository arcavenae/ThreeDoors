package tasks

// TextFileProvider wraps the existing file_manager.go functions
// to implement the TaskProvider interface.
type TextFileProvider struct{}

// NewTextFileProvider creates a new TextFileProvider.
func NewTextFileProvider() *TextFileProvider {
	return &TextFileProvider{}
}

// LoadTasks reads tasks from the YAML file via file_manager.go.
func (p *TextFileProvider) LoadTasks() ([]*Task, error) {
	return LoadTasks()
}

// SaveTask saves a single task by loading all tasks, updating, and saving.
func (p *TextFileProvider) SaveTask(task *Task) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}
	found := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			found = true
			break
		}
	}
	if !found {
		tasks = append(tasks, task)
	}
	return SaveTasks(tasks)
}

// SaveTasks persists all tasks via file_manager.go.
func (p *TextFileProvider) SaveTasks(tasks []*Task) error {
	return SaveTasks(tasks)
}

// DeleteTask removes a task by ID.
func (p *TextFileProvider) DeleteTask(taskID string) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}
	filtered := make([]*Task, 0, len(tasks))
	for _, t := range tasks {
		if t.ID != taskID {
			filtered = append(filtered, t)
		}
	}
	return SaveTasks(filtered)
}
