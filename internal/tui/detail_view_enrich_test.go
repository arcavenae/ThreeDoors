package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/intelligence/services"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDetailView_EnrichKeyNoEnricher(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("vague task")
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("expected a command for flash message")
	}
	msg := cmd()
	flash, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(flash.Text, "LLM not configured") {
		t.Errorf("unexpected flash: %s", flash.Text)
	}
}

func TestDetailView_EnrichKeyWithEnricher(t *testing.T) {
	t.Parallel()
	task := core.NewTask("vague task")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	dv.SetEnricher(services.NewTaskEnricher(&stubBackend{}))

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("expected EnrichStartMsg command")
	}
	msg := cmd()
	enrichMsg, ok := msg.(EnrichStartMsg)
	if !ok {
		t.Fatalf("expected EnrichStartMsg, got %T", msg)
	}
	if enrichMsg.TaskText != "vague task" {
		t.Errorf("TaskText = %q, want %q", enrichMsg.TaskText, "vague task")
	}
	if dv.mode != DetailModeEnrichLoading {
		t.Errorf("mode = %v, want DetailModeEnrichLoading", dv.mode)
	}
}

func TestDetailView_EnrichLoadingIgnoresKeys(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("task")
	dv.mode = DetailModeEnrichLoading

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd != nil {
		t.Error("loading mode should ignore key presses")
	}
}

func TestDetailView_EnrichResultRendering(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("taxes")
	dv.SetWidth(80)
	dv.mode = DetailModeEnrichResult
	dv.enrichResult = &services.EnrichedTask{
		OriginalText: "taxes",
		EnrichedText: "Gather 2025 tax documents and schedule accountant appointment",
		Tags:         []string{"finance", "personal"},
		Effort:       3,
		Context:      "Tax deadline April 15",
	}

	view := dv.View()
	if !strings.Contains(view, "Original:") {
		t.Error("should show Original: label")
	}
	if !strings.Contains(view, "taxes") {
		t.Error("should show original text")
	}
	if !strings.Contains(view, "Enriched:") {
		t.Error("should show Enriched: label")
	}
	if !strings.Contains(view, "Gather 2025 tax documents") {
		t.Error("should show enriched text")
	}
	if !strings.Contains(view, "finance, personal") {
		t.Error("should show tags")
	}
	if !strings.Contains(view, "3/5") {
		t.Error("should show effort")
	}
	if !strings.Contains(view, "Tax deadline April 15") {
		t.Error("should show context")
	}
	if !strings.Contains(view, "[A]ccept") {
		t.Error("should show accept hint")
	}
	if !strings.Contains(view, "[D]iscard") {
		t.Error("should show discard hint")
	}
}

func TestDetailView_EnrichAccept(t *testing.T) {
	t.Parallel()
	task := core.NewTask("taxes")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	dv.mode = DetailModeEnrichResult
	dv.enrichResult = &services.EnrichedTask{
		OriginalText: "taxes",
		EnrichedText: "Gather tax docs",
		Tags:         []string{"finance"},
		Effort:       3,
		Context:      "April 15 deadline",
	}

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected command for EnrichAcceptMsg")
	}
	msg := cmd()
	acceptMsg, ok := msg.(EnrichAcceptMsg)
	if !ok {
		t.Fatalf("expected EnrichAcceptMsg, got %T", msg)
	}
	if acceptMsg.EnrichedText != "Gather tax docs" {
		t.Errorf("EnrichedText = %q, want %q", acceptMsg.EnrichedText, "Gather tax docs")
	}
	if acceptMsg.Effort != 3 {
		t.Errorf("Effort = %d, want 3", acceptMsg.Effort)
	}
	if dv.mode != DetailModeView {
		t.Errorf("mode = %v, want DetailModeView", dv.mode)
	}
	if dv.enrichResult != nil {
		t.Error("enrichResult should be cleared after accept")
	}
}

func TestDetailView_EnrichDiscard(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("taxes")
	dv.mode = DetailModeEnrichResult
	dv.enrichResult = &services.EnrichedTask{
		OriginalText: "taxes",
		EnrichedText: "Gather tax docs",
		Tags:         []string{"finance"},
		Effort:       3,
		Context:      "April 15 deadline",
	}

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected command for flash message")
	}
	msg := cmd()
	flash, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(flash.Text, "discarded") {
		t.Errorf("unexpected flash: %s", flash.Text)
	}
	if dv.mode != DetailModeView {
		t.Errorf("mode = %v, want DetailModeView", dv.mode)
	}
	if dv.enrichResult != nil {
		t.Error("enrichResult should be cleared after discard")
	}
}

func TestDetailView_EnrichDiscardEsc(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("taxes")
	dv.mode = DetailModeEnrichResult
	dv.enrichResult = &services.EnrichedTask{
		OriginalText: "taxes",
		EnrichedText: "Gather tax docs",
	}

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if dv.mode != DetailModeView {
		t.Errorf("mode = %v, want DetailModeView", dv.mode)
	}
}

func TestDetailView_EnrichResultMsg_Success(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test")
	dv.mode = DetailModeEnrichLoading

	result := &services.EnrichedTask{
		OriginalText: "test",
		EnrichedText: "Test enriched",
		Tags:         []string{"testing"},
		Effort:       2,
		Context:      "Context here",
	}
	dv.Update(EnrichResultMsg{TaskID: "abc", Result: result})

	if dv.mode != DetailModeEnrichResult {
		t.Errorf("mode = %v, want DetailModeEnrichResult", dv.mode)
	}
	if dv.enrichResult != result {
		t.Error("enrichResult should be set to the result")
	}
}

func TestDetailView_EnrichResultMsg_Error(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test")
	dv.mode = DetailModeEnrichLoading

	cmd := dv.Update(EnrichResultMsg{TaskID: "abc", Err: errTestEnrich})
	if dv.mode != DetailModeView {
		t.Errorf("mode = %v, want DetailModeView on error", dv.mode)
	}
	if cmd == nil {
		t.Fatal("expected flash command on error")
	}
	msg := cmd()
	flash, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(flash.Text, "Enrich failed") {
		t.Errorf("unexpected flash: %s", flash.Text)
	}
}

func TestDetailView_EnrichHintShown(t *testing.T) {
	t.Parallel()
	task := core.NewTask("test")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	dv.SetEnricher(services.NewTaskEnricher(&stubBackend{}))

	view := dv.View()
	if !strings.Contains(view, "enrich") && !strings.Contains(view, "Enrich") && !strings.Contains(view, "N]") {
		t.Error("should show enrich hint when enricher is set")
	}
}

func TestDetailView_EnrichHintHidden(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test")
	dv.SetWidth(80)

	view := dv.View()
	if strings.Contains(view, "[N]enrich") {
		t.Error("should NOT show enrich hint when enricher is nil")
	}
}

func TestDetailView_EnrichLoadingView(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test")
	dv.SetWidth(80)
	dv.mode = DetailModeEnrichLoading

	view := dv.View()
	if !strings.Contains(view, "Enriching task") {
		t.Error("loading view should show enriching message")
	}
}

// stubBackend implements llm.LLMBackend for testing.
type stubBackend struct{}

func (s *stubBackend) Name() string                                         { return "stub" }
func (s *stubBackend) Complete(_ context.Context, _ string) (string, error) { return "{}", nil }
func (s *stubBackend) Available(_ context.Context) bool                     { return true }

var errTestEnrich = errEnrichTest("test enrichment error")

type errEnrichTest string

func (e errEnrichTest) Error() string { return string(e) }
