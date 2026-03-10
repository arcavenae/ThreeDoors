package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

func TestHelpView_NewHelpView(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	if hv == nil {
		t.Fatal("NewHelpView returned nil")
		return
	}
	if len(hv.content) == 0 {
		t.Error("help view should have pre-rendered content")
	}
	if hv.width != 80 {
		t.Errorf("default width = %d, want 80", hv.width)
	}
	if !hv.ready {
		t.Error("help view should be ready after creation")
	}
}

func TestHelpView_SetWidth(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	hv.SetWidth(120)
	if hv.width != 120 {
		t.Errorf("width = %d, want 120", hv.width)
	}
	if hv.viewport.Width != 120 {
		t.Errorf("viewport width = %d, want 120", hv.viewport.Width)
	}
	if len(hv.content) == 0 {
		t.Error("content should not be empty after SetWidth")
	}
}

func TestHelpView_SetHeight(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	hv.SetHeight(40)
	if hv.height != 40 {
		t.Errorf("height = %d, want 40", hv.height)
	}
	// viewport height = total - header - footer
	wantVPHeight := 40 - headerHeight - footerHeight
	if hv.viewport.Height != wantVPHeight {
		t.Errorf("viewport height = %d, want %d", hv.viewport.Height, wantVPHeight)
	}
}

func TestHelpView_SetHeight_MinimumClamp(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	hv.SetHeight(1)
	if hv.viewport.Height < 1 {
		t.Errorf("viewport height should be at least 1, got %d", hv.viewport.Height)
	}
}

func TestHelpView_Update_QuitKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "esc returns to doors",
			key:  tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name: "q returns to doors",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hv := NewHelpView()

			cmd := hv.Update(tt.key)
			if cmd == nil {
				t.Fatal("expected command, got nil")
				return
			}
			msg := cmd()
			if _, ok := msg.(ReturnToDoorsMsg); !ok {
				t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
			}
		})
	}
}

func TestHelpView_Update_ScrollKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		key        tea.KeyMsg
		wantScroll bool
	}{
		{
			name:       "j scrolls down",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantScroll: true,
		},
		{
			name:       "down arrow scrolls",
			key:        tea.KeyMsg{Type: tea.KeyDown},
			wantScroll: true,
		},
		{
			name:       "k at top stays at 0",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantScroll: false,
		},
		{
			name:       "pgdown pages forward",
			key:        tea.KeyMsg{Type: tea.KeyPgDown},
			wantScroll: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hv := NewHelpView()
			hv.SetHeight(10) // small height so content overflows

			initialOffset := hv.viewport.YOffset
			hv.Update(tt.key)

			scrolled := hv.viewport.YOffset != initialOffset
			if scrolled != tt.wantScroll {
				t.Errorf("scrolled = %v, want %v (offset: %d → %d)",
					scrolled, tt.wantScroll, initialOffset, hv.viewport.YOffset)
			}
		})
	}
}

func TestHelpView_Update_ViewportDelegation(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	hv.SetHeight(10) // small viewport so content overflows

	// Press down multiple times to scroll
	for range 5 {
		hv.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if hv.viewport.YOffset == 0 {
		t.Error("viewport should have scrolled after multiple down presses")
	}

	// Press up to scroll back
	offset := hv.viewport.YOffset
	hv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if hv.viewport.YOffset >= offset {
		t.Error("viewport should scroll up on up key")
	}
}

func TestHelpView_View_Contains80Columns(t *testing.T) {
	t.Parallel()

	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHelpView()
	hv.SetWidth(80)
	hv.SetHeight(100) // large enough to show all content
	view := hv.View()

	lines := strings.Split(view, "\n")
	for i, line := range lines {
		// Viewport right-pads lines to fill width; trim trailing spaces for check
		trimmed := strings.TrimRight(line, " ")
		if len(trimmed) > 80 {
			t.Errorf("line %d exceeds 80 columns (%d chars): %q", i+1, len(trimmed), trimmed)
		}
	}
}

func TestHelpView_View_ContainsSections(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	hv.SetHeight(100) // large enough to show all content

	content := hv.View()

	expectedSections := []string{"Navigation", "Task Actions", "Commands", "Search", "Global"}
	for _, section := range expectedSections {
		if !strings.Contains(content, section) {
			t.Errorf("help content missing section %q", section)
		}
	}
}

func TestHelpView_View_ContainsAllBindings(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	// Check the raw content rather than the viewport view
	expectedEntries := []string{
		":add", ":tag", ":theme", ":mood", ":stats", ":health",
		":synclog", ":devqueue", ":help", ":quit",
		"Complete task", "Mark task as blocked",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(hv.content, entry) {
			t.Errorf("help content missing entry %q", entry)
		}
	}
}

func TestHelpView_View_ScrollPercent(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	view := hv.View()

	if !strings.Contains(view, "%") {
		t.Error("view should contain scroll percentage indicator")
	}
}

func TestHelpView_View_Empty(t *testing.T) {
	t.Parallel()
	hv := &HelpView{width: 80, height: 24}
	view := hv.View()

	if !strings.Contains(view, "Help") {
		t.Error("view should contain 'Help' header")
	}
	if !strings.Contains(view, "No help content available") {
		t.Error("view should show empty message")
	}
}

func TestHelpView_MouseWheelEnabled(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()

	if !hv.viewport.MouseWheelEnabled {
		t.Error("viewport should have mouse wheel enabled")
	}
}

func TestGolden_HelpView(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHelpView()
	hv.SetWidth(80)
	hv.SetHeight(24)
	out := hv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestHelpCommand_ProducesShowHelpMsg(t *testing.T) {
	t.Parallel()
	pool := makePool("task1", "task2", "task3")
	sv := NewSearchView(pool, nil, nil, nil, nil)
	sv.textInput.SetValue(":help")
	sv.checkCommandMode()

	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal("expected command from :help")
		return
	}

	msg := cmd()
	if _, ok := msg.(ShowHelpMsg); !ok {
		t.Errorf(":help should produce ShowHelpMsg, got %T", msg)
	}
}

func TestWordWrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{"short text", "hello world", 80, "hello world"},
		{"exact fit", "hello", 5, "hello"},
		{"wraps at boundary", "hello world foo bar", 11, "hello world\nfoo bar"},
		{"long word", "superlongword", 5, "superlongword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wordWrap(tt.text, tt.maxWidth)
			if got != tt.want {
				t.Errorf("wordWrap(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}
