package tasks

// TaskProvider defines the interface for task storage backends.
// Implementations include TextFileProvider (wraps file_manager.go) and
// future AppleNotesProvider.
type TaskProvider interface {
	LoadTasks() ([]*Task, error)
	SaveTask(task *Task) error
	SaveTasks(tasks []*Task) error
	DeleteTask(taskID string) error
	MarkComplete(taskID string) error
}
