package adapters_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters/applenotes"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// noteStore provides an in-memory Apple Notes simulation for contract testing.
// It stores plaintext note content keyed by note title and handles the
// read-modify-write cycle that AppleNotesProvider uses via osascript.
type noteStore struct {
	mu      sync.Mutex
	content map[string]string
}

func newNoteStore() *noteStore {
	return &noteStore{
		content: make(map[string]string),
	}
}

func (ns *noteStore) executor(title string) applenotes.CommandExecutor {
	return func(_ context.Context, script string) (string, error) {
		ns.mu.Lock()
		defer ns.mu.Unlock()

		if strings.Contains(script, "get plaintext text") {
			content, ok := ns.content[title]
			if !ok {
				return "", nil
			}
			return content, nil
		}

		if strings.Contains(script, "set body") {
			ns.content[title] = htmlToPlaintext(script)
			return "", nil
		}

		return "", fmt.Errorf("unexpected script: %s", script)
	}
}

// htmlToPlaintext extracts plaintext from the AppleScript set body command.
// This reverses the plaintextToHTML conversion for testing purposes.
func htmlToPlaintext(script string) string {
	idx := strings.Index(script, "set body")
	if idx == -1 {
		return ""
	}

	rest := script[idx:]
	toIdx := strings.Index(rest, `to "`)
	if toIdx == -1 {
		return ""
	}

	html := rest[toIdx+4:]
	html = strings.TrimSuffix(html, `"`)

	// Unescape AppleScript escaping
	html = strings.ReplaceAll(html, `\"`, `"`)
	html = strings.ReplaceAll(html, `\\`, `\`)

	// Convert HTML divs back to plaintext lines
	html = strings.ReplaceAll(html, "<div><br></div>", "\n")
	html = strings.ReplaceAll(html, "<div>", "")
	html = strings.ReplaceAll(html, "</div>", "\n")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&#34;", `"`)

	return strings.TrimRight(html, "\n")
}

// TestAppleNotesProvider_ContractSubset runs contract-compatible tests against
// AppleNotesProvider using an in-memory note store simulation.
//
// IMPORTANT: AppleNotesProvider uses position-based SHA-1 IDs (noteTitle:lineIndex),
// not stored IDs. This means core.NewTask() UUIDs are NOT preserved across
// save/load cycles — the provider regenerates deterministic IDs from line position.
// This is by design: Apple Notes has no metadata storage, so IDs must be derived.
//
// Because of this, the standard RunContractTests suite (which expects ID round-tripping)
// is NOT applicable. Instead, we validate the subset of contract behaviors that
// are compatible with position-based IDs.
func TestAppleNotesProvider_ContractSubset(t *testing.T) {
	t.Run("SaveAndLoad_PreservesTextAndCount", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-save-load"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		original := []*core.Task{
			core.NewTask("Task Alpha"),
			core.NewTask("Task Beta"),
		}

		if err := provider.SaveTasks(original); err != nil {
			t.Fatalf("SaveTasks() error: %v", err)
		}

		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() error: %v", err)
		}

		if len(loaded) != len(original) {
			t.Fatalf("LoadTasks() returned %d tasks, want %d", len(loaded), len(original))
		}

		// Verify text is preserved (IDs will differ — position-based)
		for i, orig := range original {
			if loaded[i].Text != orig.Text {
				t.Errorf("task[%d].Text = %q, want %q", i, loaded[i].Text, orig.Text)
			}
		}
	})

	t.Run("SaveTask_AppendsNew", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-save-new"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		task := core.NewTask("New individual task")
		if err := provider.SaveTask(task); err != nil {
			t.Fatalf("SaveTask() error: %v", err)
		}

		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() error: %v", err)
		}

		found := false
		for _, lt := range loaded {
			if lt.Text == task.Text {
				found = true
			}
		}
		if !found {
			t.Error("saved task text not found in loaded tasks")
		}
	})

	t.Run("SaveTask_UpdateExistingByPositionID", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-update"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		// Save initial task
		task := core.NewTask("Original text")
		if err := provider.SaveTask(task); err != nil {
			t.Fatalf("SaveTask() setup error: %v", err)
		}

		// Load to get position-based ID
		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() error: %v", err)
		}
		if len(loaded) == 0 {
			t.Fatal("no tasks loaded")
		}

		// Update using the position-based ID
		loaded[0].Status = core.StatusComplete
		if err := provider.SaveTask(loaded[0]); err != nil {
			t.Fatalf("SaveTask() update error: %v", err)
		}

		// Verify update persisted
		reloaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() after update error: %v", err)
		}
		if reloaded[0].Status != core.StatusComplete {
			t.Errorf("task status = %q, want %q", reloaded[0].Status, core.StatusComplete)
		}
	})

	t.Run("DeleteTask_ByPositionID", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-delete"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		batch := []*core.Task{
			core.NewTask("Keep this"),
			core.NewTask("Delete this"),
		}
		if err := provider.SaveTasks(batch); err != nil {
			t.Fatalf("SaveTasks() error: %v", err)
		}

		// Load to get position-based IDs
		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() error: %v", err)
		}
		if len(loaded) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(loaded))
		}

		deleteID := loaded[1].ID
		if err := provider.DeleteTask(deleteID); err != nil {
			t.Fatalf("DeleteTask() error: %v", err)
		}

		reloaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() after delete error: %v", err)
		}
		if len(reloaded) != 1 {
			t.Fatalf("expected 1 task after delete, got %d", len(reloaded))
		}
		if reloaded[0].Text != "Keep this" {
			t.Errorf("remaining task text = %q, want %q", reloaded[0].Text, "Keep this")
		}
	})

	t.Run("MarkComplete_ReturnsReadOnly", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-mark"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		err := provider.MarkComplete("any-id")
		if err == nil {
			t.Fatal("MarkComplete() should return error")
			return
		}
		if err.Error() != "provider is read-only" {
			t.Errorf("MarkComplete() error = %q, want 'provider is read-only'", err.Error())
		}
	})

	t.Run("SaveTasks_Batch", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-batch"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		batch := make([]*core.Task, 10)
		for i := range batch {
			batch[i] = core.NewTask(fmt.Sprintf("Batch task %d", i))
		}

		if err := provider.SaveTasks(batch); err != nil {
			t.Fatalf("SaveTasks() batch error: %v", err)
		}

		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("LoadTasks() error: %v", err)
		}
		if len(loaded) != 10 {
			t.Errorf("LoadTasks() returned %d tasks, want 10", len(loaded))
		}
	})

	t.Run("ConcurrentReads", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-concurrent-read"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		seed := []*core.Task{
			core.NewTask("Read test 1"),
			core.NewTask("Read test 2"),
		}
		if err := provider.SaveTasks(seed); err != nil {
			t.Fatalf("SaveTasks() setup error: %v", err)
		}

		var wg sync.WaitGroup
		errCh := make(chan error, 20)

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := provider.LoadTasks(); err != nil {
					errCh <- err
				}
			}()
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Errorf("concurrent read error: %v", err)
		}
	})

	t.Run("ConcurrentWrites", func(t *testing.T) {
		store := newNoteStore()
		title := "contract-concurrent-write"
		provider := applenotes.NewAppleNotesProviderWithExecutor(title, store.executor(title))

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				task := core.NewTask("concurrent write task")
				_ = provider.SaveTask(task)
			}()
		}

		wg.Wait()

		// Verify no panics — provider remains usable
		_, _ = provider.LoadTasks()
	})
}
