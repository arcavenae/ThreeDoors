package obsidian

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"

	"github.com/fsnotify/fsnotify"
)

// ObsidianChangeEvent represents a detected change in the Obsidian vault.
type ObsidianChangeEvent struct {
	FilePath  string
	Tasks     []*core.Task
	Timestamp time.Time
}

// ObsidianWatcher monitors an Obsidian vault directory for file changes and
// triggers incremental re-parsing of modified files.
type ObsidianWatcher struct {
	adapter  *ObsidianAdapter
	watcher  *fsnotify.Watcher
	onChange func(ObsidianChangeEvent)

	// debounce tracks pending file events to coalesce rapid writes
	debounce   map[string]*time.Timer
	debounceMu sync.Mutex

	// selfWrites tracks files we recently wrote to, so we can ignore our own events
	selfWrites   map[string]time.Time
	selfWritesMu sync.Mutex

	debounceInterval time.Duration
	selfWriteTTL     time.Duration

	done chan struct{}
	wg   sync.WaitGroup
}

// NewObsidianWatcher creates a watcher that monitors the adapter's vault directory.
// The onChange callback is invoked with the re-parsed tasks when a file changes.
func NewObsidianWatcher(adapter *ObsidianAdapter, onChange func(ObsidianChangeEvent)) *ObsidianWatcher {
	return &ObsidianWatcher{
		adapter:          adapter,
		onChange:         onChange,
		debounce:         make(map[string]*time.Timer),
		selfWrites:       make(map[string]time.Time),
		debounceInterval: 100 * time.Millisecond,
		selfWriteTTL:     2 * time.Second,
		done:             make(chan struct{}),
	}
}

// Start begins watching the vault directory for file changes.
func (w *ObsidianWatcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("obsidian watcher create: %w", err)
	}
	w.watcher = watcher

	dir := w.adapter.taskDir()
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("obsidian watcher add %q: %w", dir, err)
	}

	w.wg.Add(1)
	go w.eventLoop()

	return nil
}

// Stop cleanly shuts down the file watcher.
func (w *ObsidianWatcher) Stop() error {
	close(w.done)
	w.wg.Wait()

	w.debounceMu.Lock()
	for _, t := range w.debounce {
		t.Stop()
	}
	w.debounce = make(map[string]*time.Timer)
	w.debounceMu.Unlock()

	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}

// RecordSelfWrite marks a file as recently written by us, so the watcher
// ignores the resulting fsnotify event.
func (w *ObsidianWatcher) RecordSelfWrite(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	w.selfWritesMu.Lock()
	w.selfWrites[abs] = time.Now().UTC()
	w.selfWritesMu.Unlock()
}

// isSelfWrite checks if a file event was caused by our own write.
func (w *ObsidianWatcher) isSelfWrite(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	w.selfWritesMu.Lock()
	defer w.selfWritesMu.Unlock()

	t, ok := w.selfWrites[abs]
	if !ok {
		return false
	}

	if time.Since(t) < w.selfWriteTTL {
		delete(w.selfWrites, abs)
		return true
	}

	// Expired entry — clean it up
	delete(w.selfWrites, abs)
	return false
}

// eventLoop processes fsnotify events with debouncing.
func (w *ObsidianWatcher) eventLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("obsidian watcher error: %v", err)
		}
	}
}

// handleEvent processes a single fsnotify event with debouncing and filtering.
func (w *ObsidianWatcher) handleEvent(event fsnotify.Event) {
	// Only care about write/create/rename events on .md files
	if !isRelevantEvent(event) {
		return
	}

	// Ignore our own writes
	if w.isSelfWrite(event.Name) {
		return
	}

	// Debounce: reset timer for this file
	w.debounceMu.Lock()
	if existing, ok := w.debounce[event.Name]; ok {
		existing.Stop()
	}
	w.debounce[event.Name] = time.AfterFunc(w.debounceInterval, func() {
		w.processFileChange(event.Name)
	})
	w.debounceMu.Unlock()
}

// isRelevantEvent checks if an fsnotify event should trigger a re-parse.
func isRelevantEvent(event fsnotify.Event) bool {
	if !strings.HasSuffix(event.Name, ".md") {
		return false
	}
	// Ignore temp files from atomic writes
	if strings.HasSuffix(event.Name, ".tmp") {
		return false
	}
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename)
}

// processFileChange re-parses a single changed file and notifies via callback.
func (w *ObsidianWatcher) processFileChange(filePath string) {
	w.debounceMu.Lock()
	delete(w.debounce, filePath)
	w.debounceMu.Unlock()

	now := time.Now().UTC()

	// Use the adapter's parseFile method for incremental re-parse
	w.adapter.mu.Lock()
	tasks, err := w.adapter.parseFile(filePath, now)
	w.adapter.mu.Unlock()

	if err != nil {
		log.Printf("obsidian watcher: re-parse %q: %v", filepath.Base(filePath), err)
		return
	}

	if w.onChange != nil {
		w.onChange(ObsidianChangeEvent{
			FilePath:  filePath,
			Tasks:     tasks,
			Timestamp: now,
		})
	}
}
