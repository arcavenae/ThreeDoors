package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

func setupLinkTest(t *testing.T) (*LinkView, *tasks.Task, *tasks.Task, *enrichment.DB) {
	t.Helper()
	source := tasks.NewTask("Source task")
	target := tasks.NewTask("Target task")

	pool := tasks.NewTaskPool()
	pool.AddTask(source)
	pool.AddTask(target)

	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	db, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close db: %v", closeErr)
		}
		_ = os.Remove(dbPath)
	})

	lv := NewLinkView(source, pool, db)
	return lv, source, target, db
}

func TestLinkView_InitialState(t *testing.T) {
	lv, source, _, _ := setupLinkTest(t)

	if lv.sourceTask != source {
		t.Error("source task should be set")
	}
	if lv.step != linkStepSelectTarget {
		t.Error("initial step should be target selection")
	}
	if lv.selectedIndex != -1 {
		t.Error("initial selected index should be -1")
	}
}

func TestLinkView_SearchFiltersTargets(t *testing.T) {
	lv, _, _, _ := setupLinkTest(t)

	// Type "Targ" to search
	for _, r := range "Targ" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if len(lv.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(lv.results))
	}
	if lv.results[0].Text != "Target task" {
		t.Errorf("expected 'Target task', got %q", lv.results[0].Text)
	}
}

func TestLinkView_ExcludesSourceTask(t *testing.T) {
	lv, source, _, _ := setupLinkTest(t)

	// Type "Source" to search
	for _, r := range "Source" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	for _, r := range lv.results {
		if r.ID == source.ID {
			t.Error("source task should be excluded from results")
		}
	}
}

func TestLinkView_EscCancels(t *testing.T) {
	lv, _, _, _ := setupLinkTest(t)

	cmd := lv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
	}
	msg := cmd()
	if _, ok := msg.(LinkCancelledMsg); !ok {
		t.Errorf("expected LinkCancelledMsg, got %T", msg)
	}
}

func TestLinkView_SelectTargetAndRelationship(t *testing.T) {
	lv, _, _, db := setupLinkTest(t)

	// Type to filter
	for _, r := range "Target" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Enter to select target
	lv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if lv.step != linkStepSelectRelationship {
		t.Fatal("should advance to relationship selection")
	}

	// Default is "related" (index 0); press Enter to confirm
	cmd := lv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirming should produce a command")
	}
	msg := cmd()
	linkMsg, ok := msg.(LinkCreatedMsg)
	if !ok {
		t.Fatalf("expected LinkCreatedMsg, got %T", msg)
	}
	if linkMsg.Relationship != "related" {
		t.Errorf("expected relationship 'related', got %q", linkMsg.Relationship)
	}

	// Verify the cross-reference was stored
	refs, err := db.GetCrossReferences(linkMsg.SourceTaskID)
	if err != nil {
		t.Fatalf("failed to get cross references: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 cross reference, got %d", len(refs))
	}
	if refs[0].Relationship != "related" {
		t.Errorf("stored relationship should be 'related', got %q", refs[0].Relationship)
	}
}

func TestLinkView_RelationshipNavigation(t *testing.T) {
	lv, _, _, _ := setupLinkTest(t)

	// Type to filter and select
	for _, r := range "Target" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	lv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if lv.relIndex != 0 {
		t.Fatal("default relationship index should be 0")
	}

	// Navigate down
	lv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lv.relIndex != 1 {
		t.Errorf("expected relIndex 1, got %d", lv.relIndex)
	}

	lv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lv.relIndex != 2 {
		t.Errorf("expected relIndex 2, got %d", lv.relIndex)
	}

	// Can't go past end
	lv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lv.relIndex != 2 {
		t.Errorf("should stay at 2, got %d", lv.relIndex)
	}

	// Navigate up
	lv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if lv.relIndex != 1 {
		t.Errorf("expected relIndex 1 after up, got %d", lv.relIndex)
	}
}

func TestLinkView_EscFromRelationshipReturnsToTarget(t *testing.T) {
	lv, _, _, _ := setupLinkTest(t)

	for _, r := range "Target" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	lv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if lv.step != linkStepSelectRelationship {
		t.Fatal("should be on relationship step")
	}

	lv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if lv.step != linkStepSelectTarget {
		t.Error("esc from relationship should return to target selection")
	}
}

func TestLinkView_ViewRendering(t *testing.T) {
	lv, _, _, _ := setupLinkTest(t)
	lv.SetWidth(80)

	view := lv.View()
	if !strings.Contains(view, "LINK TASK") {
		t.Error("view should contain header")
	}
	if !strings.Contains(view, "Source task") {
		t.Error("view should show source task")
	}
}

func TestDetailView_LinkedTasksDisplay(t *testing.T) {
	source := tasks.NewTask("Source task")
	target := tasks.NewTask("Target task")

	pool := tasks.NewTaskPool()
	pool.AddTask(source)
	pool.AddTask(target)

	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	db, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close db: %v", closeErr)
		}
		_ = os.Remove(dbPath)
	})

	// Create a cross-reference
	ref := &enrichment.CrossReference{
		SourceTaskID: source.ID,
		TargetTaskID: target.ID,
		SourceSystem: "user",
		Relationship: "related",
	}
	if addErr := db.AddCrossReference(ref); addErr != nil {
		t.Fatalf("failed to add cross reference: %v", addErr)
	}

	dv := NewDetailView(source, nil)
	dv.SetEnrichDB(db, pool)
	dv.SetWidth(80)

	view := dv.View()
	if !strings.Contains(view, "Linked Tasks") {
		t.Error("detail view should show 'Linked Tasks' section")
	}
	if !strings.Contains(view, "Target task") {
		t.Error("detail view should show linked task text")
	}
	if !strings.Contains(view, "related") {
		t.Error("detail view should show relationship type")
	}
}

func TestDetailView_NoLinkedTasks_NoSection(t *testing.T) {
	source := tasks.NewTask("Source task")
	pool := tasks.NewTaskPool()
	pool.AddTask(source)

	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	db, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close db: %v", closeErr)
		}
		_ = os.Remove(dbPath)
	})

	dv := NewDetailView(source, nil)
	dv.SetEnrichDB(db, pool)
	dv.SetWidth(80)

	view := dv.View()
	if strings.Contains(view, "Linked Tasks") {
		t.Error("detail view should NOT show 'Linked Tasks' when none exist")
	}
}

func TestDetailView_NavigateToLinkedTask(t *testing.T) {
	source := tasks.NewTask("Source task")
	target := tasks.NewTask("Target task")

	pool := tasks.NewTaskPool()
	pool.AddTask(source)
	pool.AddTask(target)

	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	db, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close db: %v", closeErr)
		}
		_ = os.Remove(dbPath)
	})

	ref := &enrichment.CrossReference{
		SourceTaskID: source.ID,
		TargetTaskID: target.ID,
		SourceSystem: "user",
		Relationship: "blocks",
	}
	if addErr := db.AddCrossReference(ref); addErr != nil {
		t.Fatalf("failed to add cross reference: %v", addErr)
	}

	dv := NewDetailView(source, nil)
	dv.SetEnrichDB(db, pool)

	// Press "1" to navigate to first linked task
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd == nil {
		t.Fatal("pressing '1' with a linked task should produce a command")
	}
	msg := cmd()
	navMsg, ok := msg.(NavigateToLinkedTaskMsg)
	if !ok {
		t.Fatalf("expected NavigateToLinkedTaskMsg, got %T", msg)
	}
	if navMsg.TaskID != target.ID {
		t.Errorf("expected target task ID %s, got %s", target.ID, navMsg.TaskID)
	}
}

func TestMainModel_LinkCommandFlow(t *testing.T) {
	source := tasks.NewTask("Source task")
	target := tasks.NewTask("Target task")

	pool := tasks.NewTaskPool()
	pool.AddTask(source)
	pool.AddTask(target)

	dbPath := filepath.Join(t.TempDir(), "test-enrichment.db")
	db, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open enrichment db: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close db: %v", closeErr)
		}
		_ = os.Remove(dbPath)
	})

	tracker := tasks.NewSessionTracker()
	m := NewMainModel(pool, tracker, &testProvider{}, nil, false, db)

	// Open detail view for source task
	m.detailView = NewDetailView(source, tracker)
	m.detailView.SetEnrichDB(db, pool)
	m.viewMode = ViewDetail

	// Send ShowLinkViewMsg
	m.Update(ShowLinkViewMsg{SourceTask: source})

	if m.viewMode != ViewLink {
		t.Errorf("expected ViewLink mode, got %d", m.viewMode)
	}
	if m.linkView == nil {
		t.Fatal("linkView should be created")
	}
}

func TestMainModel_ShowLinkViewMsg_NoEnrichDB(t *testing.T) {
	pool := tasks.NewTaskPool()
	pool.AddTask(tasks.NewTask("task"))
	tracker := tasks.NewSessionTracker()
	m := NewMainModel(pool, tracker, &testProvider{}, nil, false, nil)

	m.Update(ShowLinkViewMsg{SourceTask: pool.GetAllTasks()[0]})

	if m.viewMode == ViewLink {
		t.Error("should NOT enter ViewLink when enrichDB is nil")
	}
	if m.flash == "" {
		t.Error("should show flash message about unavailable DB")
	}
}
