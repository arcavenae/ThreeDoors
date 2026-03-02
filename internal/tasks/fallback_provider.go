package tasks

import (
	"errors"
	"fmt"
	"os"
)

// FallbackProvider wraps a primary TaskProvider with a fallback.
// If the primary fails on LoadTasks, the fallback is used instead.
type FallbackProvider struct {
	primary        TaskProvider
	fallback       TaskProvider
	usedFallback   bool
	fallbackReason string
}

// NewFallbackProvider creates a FallbackProvider with the given primary and fallback providers.
func NewFallbackProvider(primary, fallback TaskProvider) *FallbackProvider {
	return &FallbackProvider{
		primary:  primary,
		fallback: fallback,
	}
}

// LoadTasks tries the primary provider first. On error, falls back to the fallback provider.
func (fp *FallbackProvider) LoadTasks() ([]*Task, error) {
	tasks, err := fp.primary.LoadTasks()
	if err == nil {
		return tasks, nil
	}

	// Primary failed — try fallback
	fp.usedFallback = true
	fp.fallbackReason = err.Error()
	fmt.Fprintf(os.Stderr, "Warning: primary provider failed: %v. Falling back to text file.\n", err)

	fallbackTasks, fallbackErr := fp.fallback.LoadTasks()
	if fallbackErr != nil {
		return nil, fmt.Errorf("both providers failed: primary: %v, fallback: %w", err, fallbackErr)
	}

	return fallbackTasks, nil
}

// SaveTask delegates to the active provider. If primary returns ErrReadOnly, delegates to fallback.
func (fp *FallbackProvider) SaveTask(task *Task) error {
	if fp.usedFallback {
		return fp.fallback.SaveTask(task)
	}

	err := fp.primary.SaveTask(task)
	if errors.Is(err, ErrReadOnly) {
		return fp.fallback.SaveTask(task)
	}
	return err
}

// SaveTasks delegates to the active provider. If primary returns ErrReadOnly, delegates to fallback.
func (fp *FallbackProvider) SaveTasks(tasks []*Task) error {
	if fp.usedFallback {
		return fp.fallback.SaveTasks(tasks)
	}

	err := fp.primary.SaveTasks(tasks)
	if errors.Is(err, ErrReadOnly) {
		return fp.fallback.SaveTasks(tasks)
	}
	return err
}

// DeleteTask delegates to the active provider. If primary returns ErrReadOnly, delegates to fallback.
func (fp *FallbackProvider) DeleteTask(taskID string) error {
	if fp.usedFallback {
		return fp.fallback.DeleteTask(taskID)
	}

	err := fp.primary.DeleteTask(taskID)
	if errors.Is(err, ErrReadOnly) {
		return fp.fallback.DeleteTask(taskID)
	}
	return err
}

// MarkComplete delegates to the active provider. If primary returns ErrReadOnly, delegates to fallback.
func (fp *FallbackProvider) MarkComplete(taskID string) error {
	if fp.usedFallback {
		return fp.fallback.MarkComplete(taskID)
	}

	err := fp.primary.MarkComplete(taskID)
	if errors.Is(err, ErrReadOnly) {
		return fp.fallback.MarkComplete(taskID)
	}
	return err
}

// IsFallback returns true if the fallback provider is currently active.
func (fp *FallbackProvider) IsFallback() bool {
	return fp.usedFallback
}

// FallbackReason returns the reason the fallback was activated.
func (fp *FallbackProvider) FallbackReason() string {
	return fp.fallbackReason
}
