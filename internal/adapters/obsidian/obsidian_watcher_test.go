package obsidian

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"

	"github.com/fsnotify/fsnotify"
)

func TestObsidianWatcher_StartStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	watcher := NewObsidianWatcher(adapter, nil)

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if err := watcher.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}

func TestObsidianWatcher_DetectsFileChange(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	var mu sync.Mutex
	var received []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	// Write a file with a task
	taskFile := filepath.Join(dir, "test.md")
	content := "- [ ] Buy groceries <!-- td:abc-123 -->\n"
	if err := os.WriteFile(taskFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Wait for debounce + processing
	deadline := time.After(3 * time.Second)
	for {
		mu.Lock()
		count := len(received)
		mu.Unlock()
		if count > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for file change event")
		case <-time.After(50 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) == 0 {
		t.Fatal("expected at least one change event")
	}

	event := received[0]
	if len(event.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(event.Tasks))
	}
	if event.Tasks[0].Text != "Buy groceries" {
		t.Errorf("expected task text 'Buy groceries', got %q", event.Tasks[0].Text)
	}
}

func TestObsidianWatcher_IgnoresSelfWrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	var mu sync.Mutex
	var received []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	if os.Getenv("CI") != "" {
		t.Skip("Skipping flaky file watcher test in CI — fsnotify timing is unreliable on GitHub Actions runners")
	}

	// Allow watcher goroutine to fully start
	time.Sleep(100 * time.Millisecond)

	// Record a self-write before writing the file
	taskFile := filepath.Join(dir, "test.md")
	watcher.RecordSelfWrite(taskFile)

	content := "- [ ] Self-written task <!-- td:self-123 -->\n"
	if err := os.WriteFile(taskFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Wait a reasonable amount of time — no event should arrive
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 0 {
		t.Errorf("expected no events for self-write, got %d", len(received))
	}
}

func TestObsidianWatcher_DebounceCoalescesRapidWrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	var mu sync.Mutex
	var received []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	taskFile := filepath.Join(dir, "rapid.md")

	// Rapid writes simulating an editor saving
	for i := 0; i < 5; i++ {
		content := "- [ ] Task version " + string(rune('A'+i)) + " <!-- td:rapid-123 -->\n"
		if err := os.WriteFile(taskFile, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for debounce to settle
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should have coalesced into 1-2 events, not 5
	if len(received) > 2 {
		t.Errorf("expected debouncing to coalesce events, got %d events", len(received))
	}
	if len(received) == 0 {
		t.Error("expected at least one event after debounce")
	}
}

func TestObsidianWatcher_IgnoresNonMarkdownFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	var mu sync.Mutex
	var received []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	// Write a non-markdown file
	txtFile := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(txtFile, []byte("not markdown"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 0 {
		t.Errorf("expected no events for .txt file, got %d", len(received))
	}
}

func TestObsidianWatcher_IncrementalParse(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	// Create two files — only modify one
	file1 := filepath.Join(dir, "file1.md")
	file2 := filepath.Join(dir, "file2.md")
	if err := os.WriteFile(file1, []byte("- [ ] Task A <!-- td:a-123 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	if err := os.WriteFile(file2, []byte("- [ ] Task B <!-- td:b-456 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var mu sync.Mutex
	var received []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		received = append(received, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	// Only modify file1
	if err := os.WriteFile(file1, []byte("- [x] Task A done <!-- td:a-123 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		mu.Lock()
		count := len(received)
		mu.Unlock()
		if count > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for change event")
		case <-time.After(50 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// Check that only file1 was re-parsed
	for _, event := range received {
		if filepath.Base(event.FilePath) == "file2.md" {
			t.Error("file2.md should not have been re-parsed")
		}
	}

	// The event for file1 should contain the updated task
	lastEvent := received[len(received)-1]
	if len(lastEvent.Tasks) != 1 {
		t.Fatalf("expected 1 task in event, got %d", len(lastEvent.Tasks))
	}
	if lastEvent.Tasks[0].Status != core.StatusComplete {
		t.Errorf("expected task status Complete, got %s", lastEvent.Tasks[0].Status)
	}
}

func TestObsidianWatcher_SyncLatencyUnder2Seconds(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	eventCh := make(chan time.Time, 1)
	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		select {
		case eventCh <- event.Timestamp:
		default:
		}
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	taskFile := filepath.Join(dir, "latency.md")
	writeTime := time.Now().UTC()
	if err := os.WriteFile(taskFile, []byte("- [ ] Latency test <!-- td:lat-1 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	select {
	case eventTime := <-eventCh:
		latency := eventTime.Sub(writeTime)
		if latency > 2*time.Second {
			t.Errorf("sync latency %v exceeds 2s requirement", latency)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out — no event within 3 seconds")
	}
}

func TestObsidianWatcher_ConflictLogging(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	// Create initial file
	taskFile := filepath.Join(dir, "conflict.md")
	if err := os.WriteFile(taskFile, []byte("- [ ] Original task <!-- td:conf-1 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var mu sync.Mutex
	var events []ObsidianChangeEvent

	watcher := NewObsidianWatcher(adapter, func(event ObsidianChangeEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	t.Cleanup(func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})

	// External write (simulating Obsidian user edit)
	if err := os.WriteFile(taskFile, []byte("- [x] Original task done <!-- td:conf-1 -->\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Wait for event
	deadline := time.After(3 * time.Second)
	for {
		mu.Lock()
		count := len(events)
		mu.Unlock()
		if count > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for event")
		case <-time.After(50 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// Last-write-wins: the external change should be reflected
	lastEvent := events[len(events)-1]
	if len(lastEvent.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(lastEvent.Tasks))
	}
	if lastEvent.Tasks[0].Status != core.StatusComplete {
		t.Errorf("expected external change to be reflected (Complete), got %s", lastEvent.Tasks[0].Status)
	}
}

func TestObsidianWatcher_StartErrorInvalidDir(t *testing.T) {
	t.Parallel()

	adapter := NewObsidianAdapter("/nonexistent/path/that/should/not/exist", "", "")
	watcher := NewObsidianWatcher(adapter, nil)

	err := watcher.Start()
	if err == nil {
		_ = watcher.Stop()
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestIsRelevantEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    fsnotify.Event
		relevant bool
	}{
		{
			name:     "md write",
			event:    fsnotify.Event{Name: "tasks.md", Op: fsnotify.Write},
			relevant: true,
		},
		{
			name:     "md create",
			event:    fsnotify.Event{Name: "new.md", Op: fsnotify.Create},
			relevant: true,
		},
		{
			name:     "md rename",
			event:    fsnotify.Event{Name: "old.md", Op: fsnotify.Rename},
			relevant: true,
		},
		{
			name:     "txt file",
			event:    fsnotify.Event{Name: "notes.txt", Op: fsnotify.Write},
			relevant: false,
		},
		{
			name:     "tmp file",
			event:    fsnotify.Event{Name: "tasks.md.tmp", Op: fsnotify.Write},
			relevant: false,
		},
		{
			name:     "md chmod only",
			event:    fsnotify.Event{Name: "tasks.md", Op: fsnotify.Chmod},
			relevant: false,
		},
		{
			name:     "md remove",
			event:    fsnotify.Event{Name: "tasks.md", Op: fsnotify.Remove},
			relevant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isRelevantEvent(tt.event)
			if got != tt.relevant {
				t.Errorf("isRelevantEvent(%v) = %v, want %v", tt.event, got, tt.relevant)
			}
		})
	}
}

func TestObsidianWatcher_RecordSelfWriteExpiry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	watcher := NewObsidianWatcher(adapter, nil)
	// Use very short TTL for testing
	watcher.selfWriteTTL = 50 * time.Millisecond

	path := filepath.Join(dir, "test.md")
	watcher.RecordSelfWrite(path)

	// Should be self-write immediately
	if !watcher.isSelfWrite(path) {
		t.Error("expected isSelfWrite=true immediately after recording")
	}

	// Record again and wait for expiry
	watcher.RecordSelfWrite(path)
	time.Sleep(100 * time.Millisecond)

	if watcher.isSelfWrite(path) {
		t.Error("expected isSelfWrite=false after TTL expiry")
	}
}
