package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	tea "github.com/charmbracelet/bubbletea"
)

func openTestEnrichDB(t *testing.T) *enrichment.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	edb, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := edb.Close(); closeErr != nil {
			t.Errorf("failed to close test enrichment db: %v", closeErr)
		}
	})
	return edb
}

func makeTestPool(texts ...string) *core.TaskPool {
	pool := core.NewTaskPool()
	for _, text := range texts {
		pool.AddTask(core.NewTask(text))
	}
	return pool
}

// --- Cross-Reference Display ---

func TestDetailView_NoCrossRefs_NoLinkedSection(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A")
	allTasks := pool.GetAllTasks()
	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "Linked") {
		t.Error("should not show Linked section when no cross-references exist")
	}
}

func TestDetailView_WithCrossRefs_ShowsLinkedSection(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "Linked") {
		t.Error("should show Linked section when cross-references exist")
	}
	if !strings.Contains(view, "Task B") {
		t.Error("should show linked task text")
	}
	if !strings.Contains(view, "[related]") {
		t.Error("should show relationship type")
	}
}

func TestDetailView_WithCrossRefs_ShowsCount(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B", "Task C")
	allTasks := pool.GetAllTasks()

	for _, target := range allTasks[1:] {
		ref := &enrichment.CrossReference{
			SourceTaskID: allTasks[0].ID,
			TargetTaskID: target.ID,
			SourceSystem: "local",
			Relationship: "related",
		}
		if err := edb.AddCrossReference(ref); err != nil {
			t.Fatalf("failed to add cross-reference: %v", err)
		}
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "(2)") {
		t.Error("should show count of linked tasks")
	}
}

func TestDetailView_ShowsXrefsHint_WhenCrossRefsExist(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "[X]refs") {
		t.Error("should show [X]refs hint when cross-references exist")
	}
}

func TestDetailView_ShowsLinkHint_WhenEnrichDBAvailable(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "[L]ink") {
		t.Error("should show [L]ink hint when enrichment DB is available")
	}
}

func TestDetailView_NoLinkHint_WhenNoEnrichDB(t *testing.T) {
	pool := makeTestPool("Task A")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, nil, pool)
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "[L]ink") {
		t.Error("should not show [L]ink hint when enrichment DB is nil")
	}
}

// --- Link Mode ---

func TestDetailView_LKey_EntersLinkSelectMode(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	if dv.mode != DetailModeLinkSelect {
		t.Errorf("expected DetailModeLinkSelect, got %d", dv.mode)
	}
	if len(dv.linkCandidates) != 1 {
		t.Errorf("expected 1 link candidate (Task B), got %d", len(dv.linkCandidates))
	}
}

func TestDetailView_LKey_NoEnrichDB_FlashesError(t *testing.T) {
	pool := makeTestPool("Task A")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, nil, pool)
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if cmd == nil {
		t.Fatal("expected a flash command")
		return
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if !strings.Contains(fm.Text, "not available") {
		t.Errorf("expected 'not available' message, got %q", fm.Text)
	}
}

func TestDetailView_LinkSelect_ExcludesCurrentTask(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B", "Task C")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	for _, c := range dv.linkCandidates {
		if c.ID == allTasks[0].ID {
			t.Error("link candidates should not include the current task")
		}
	}
}

func TestDetailView_LinkSelect_ExcludesAlreadyLinked(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B", "Task C")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	if len(dv.linkCandidates) != 1 {
		t.Errorf("expected 1 candidate (Task C only), got %d", len(dv.linkCandidates))
	}
}

func TestDetailView_LinkSelect_Enter_CreatesLink(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
		return
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if fm.Text != "Linked!" {
		t.Errorf("expected 'Linked!' message, got %q", fm.Text)
	}
	if dv.mode != DetailModeView {
		t.Errorf("should return to DetailModeView after linking")
	}
	if len(dv.crossRefs) != 1 {
		t.Errorf("expected 1 cross-reference after linking, got %d", len(dv.crossRefs))
	}
}

func TestDetailView_LinkSelect_Esc_Cancels(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if dv.mode != DetailModeView {
		t.Errorf("Esc should return to DetailModeView, got %d", dv.mode)
	}
	if dv.linkCandidates != nil {
		t.Error("link candidates should be nil after cancel")
	}
}

func TestDetailView_LinkSelect_Navigation(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B", "Task C", "Task D")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	if dv.linkSelectedIndex != 0 {
		t.Errorf("initial index should be 0, got %d", dv.linkSelectedIndex)
	}

	dv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if dv.linkSelectedIndex != 1 {
		t.Errorf("after down, index should be 1, got %d", dv.linkSelectedIndex)
	}

	dv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if dv.linkSelectedIndex != 0 {
		t.Errorf("after up, index should be 0, got %d", dv.linkSelectedIndex)
	}
}

// --- Link Browse Mode ---

func TestDetailView_XKey_EntersLinkBrowseMode(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if dv.mode != DetailModeLinkBrowse {
		t.Errorf("expected DetailModeLinkBrowse, got %d", dv.mode)
	}
}

func TestDetailView_XKey_NoCrossRefs_StaysInViewMode(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if dv.mode != DetailModeView {
		t.Errorf("should stay in DetailModeView when no cross-refs, got %d", dv.mode)
	}
}

func TestDetailView_LinkBrowse_Enter_NavigatesToLinkedTask(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter in link browse should return a command")
		return
	}
	msg := cmd()
	navMsg, ok := msg.(NavigateToLinkedMsg)
	if !ok {
		t.Fatalf("expected NavigateToLinkedMsg, got %T", msg)
	}
	if navMsg.Task.ID != allTasks[1].ID {
		t.Errorf("expected to navigate to Task B, got task with ID %s", navMsg.Task.ID)
	}
}

func TestDetailView_LinkBrowse_UKey_Unlinks(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})

	if cmd == nil {
		t.Fatal("U should return a command")
		return
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if fm.Text != "Unlinked" {
		t.Errorf("expected 'Unlinked' message, got %q", fm.Text)
	}
	if len(dv.crossRefs) != 0 {
		t.Errorf("expected 0 cross-refs after unlink, got %d", len(dv.crossRefs))
	}
	if dv.mode != DetailModeView {
		t.Errorf("should return to DetailModeView when last link removed, got %d", dv.mode)
	}
}

func TestDetailView_LinkBrowse_Esc_ReturnsToView(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if dv.mode != DetailModeView {
		t.Errorf("Esc should return to DetailModeView, got %d", dv.mode)
	}
}

// --- Bidirectional Cross-Reference ---

func TestDetailView_CrossRef_BidirectionalDisplay(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	// Link A -> B
	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	// View from B's perspective — should still show the link
	dv := NewDetailView(allTasks[1], nil, edb, pool)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "Linked") {
		t.Error("B should show Linked section (bidirectional query)")
	}
	if !strings.Contains(view, "Task A") {
		t.Error("B should show Task A as linked")
	}
}

// --- Link Select View Rendering ---

func TestDetailView_LinkSelectMode_ShowsCandidateList(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B", "Task C")
	allTasks := pool.GetAllTasks()

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	view := dv.View()

	if !strings.Contains(view, "Select task to link") {
		t.Error("should show link select prompt")
	}
	if !strings.Contains(view, "Task B") {
		t.Error("should show Task B as candidate")
	}
	if !strings.Contains(view, "Task C") {
		t.Error("should show Task C as candidate")
	}
}

// --- Link Browse View Rendering ---

func TestDetailView_LinkBrowseMode_ShowsBrowseHelp(t *testing.T) {
	edb := openTestEnrichDB(t)
	pool := makeTestPool("Task A", "Task B")
	allTasks := pool.GetAllTasks()

	ref := &enrichment.CrossReference{
		SourceTaskID: allTasks[0].ID,
		TargetTaskID: allTasks[1].ID,
		SourceSystem: "local",
		Relationship: "related",
	}
	if err := edb.AddCrossReference(ref); err != nil {
		t.Fatalf("failed to add cross-reference: %v", err)
	}

	dv := NewDetailView(allTasks[0], nil, edb, pool)
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	view := dv.View()

	if !strings.Contains(view, "[U]nlink") {
		t.Error("browse mode should show [U]nlink hint")
	}
	if !strings.Contains(view, "[Enter] Navigate") {
		t.Error("browse mode should show [Enter] Navigate hint")
	}
}
