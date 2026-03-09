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
	}
	if len(hv.lines) == 0 {
		t.Error("help view should have pre-rendered lines")
	}
	if hv.offset != 0 {
		t.Errorf("initial offset = %d, want 0", hv.offset)
	}
	if hv.width != 80 {
		t.Errorf("default width = %d, want 80", hv.width)
	}
}

func TestHelpView_SetWidth(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	originalLines := len(hv.lines)

	hv.SetWidth(120)
	if hv.width != 120 {
		t.Errorf("width = %d, want 120", hv.width)
	}
	// Lines should be re-rendered (may differ in count due to word wrap changes)
	if len(hv.lines) == 0 {
		t.Error("lines should not be empty after SetWidth")
	}
	_ = originalLines // word wrap may change line count
}

func TestHelpView_Update_KeyHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		key        tea.KeyMsg
		initial    int
		totalLines int
		wantOffset int
		wantCmd    bool
	}{
		{
			name:       "j scrolls down",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			initial:    0,
			totalLines: 50,
			wantOffset: 1,
		},
		{
			name:       "k scrolls up",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			initial:    5,
			totalLines: 50,
			wantOffset: 4,
		},
		{
			name:       "k does not go below 0",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			initial:    0,
			totalLines: 50,
			wantOffset: 0,
		},
		{
			name:       "j does not exceed line count",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			initial:    49,
			totalLines: 50,
			wantOffset: 49,
		},
		{
			name:       "down arrow scrolls down",
			key:        tea.KeyMsg{Type: tea.KeyDown},
			initial:    0,
			totalLines: 50,
			wantOffset: 1,
		},
		{
			name:       "up arrow scrolls up",
			key:        tea.KeyMsg{Type: tea.KeyUp},
			initial:    3,
			totalLines: 50,
			wantOffset: 2,
		},
		{
			name:       "pgdown pages forward",
			key:        tea.KeyMsg{Type: tea.KeyPgDown},
			initial:    0,
			totalLines: 50,
			wantOffset: helpPageSize,
		},
		{
			name:       "pgdown clamps to max",
			key:        tea.KeyMsg{Type: tea.KeyPgDown},
			initial:    45,
			totalLines: 50,
			wantOffset: 49,
		},
		{
			name:       "pgup pages backward",
			key:        tea.KeyMsg{Type: tea.KeyPgUp},
			initial:    helpPageSize + 5,
			totalLines: 50,
			wantOffset: 5,
		},
		{
			name:       "pgup clamps to 0",
			key:        tea.KeyMsg{Type: tea.KeyPgUp},
			initial:    3,
			totalLines: 50,
			wantOffset: 0,
		},
		{
			name:       "space pages forward",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}},
			initial:    0,
			totalLines: 50,
			wantOffset: helpPageSize,
		},
		{
			name:       "esc returns to doors",
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			initial:    0,
			totalLines: 50,
			wantOffset: 0,
			wantCmd:    true,
		},
		{
			name:       "q returns to doors",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			initial:    0,
			totalLines: 50,
			wantOffset: 0,
			wantCmd:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hv := &HelpView{width: 80}
			// Create exact number of lines for deterministic testing
			hv.lines = make([]string, tt.totalLines)
			for i := range hv.lines {
				hv.lines[i] = "test line"
			}
			hv.offset = tt.initial

			cmd := hv.Update(tt.key)

			if tt.wantCmd {
				if cmd == nil {
					t.Fatal("expected command, got nil")
				}
				msg := cmd()
				if _, ok := msg.(ReturnToDoorsMsg); !ok {
					t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
				}
			} else {
				if hv.offset != tt.wantOffset {
					t.Errorf("offset = %d, want %d", hv.offset, tt.wantOffset)
				}
			}
		})
	}
}

func TestHelpView_View_Contains80Columns(t *testing.T) {
	t.Parallel()

	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHelpView()
	hv.SetWidth(80)
	view := hv.View()

	lines := strings.Split(view, "\n")
	for i, line := range lines {
		// Strip ANSI codes for width check (in Ascii mode there shouldn't be any,
		// but let's be safe)
		if len(line) > 80 {
			t.Errorf("line %d exceeds 80 columns (%d chars): %q", i+1, len(line), line)
		}
	}
}

func TestHelpView_View_ContainsSections(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	// View the full content by scrolling through
	var fullContent strings.Builder
	for offset := 0; offset < len(hv.lines); offset += helpPageSize {
		hv.offset = offset
		fullContent.WriteString(hv.View())
	}

	content := fullContent.String()

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
	// Collect all lines
	var allText strings.Builder
	for _, line := range hv.lines {
		allText.WriteString(line)
		allText.WriteString("\n")
	}
	content := allText.String()

	expectedEntries := []string{
		":add", ":tag", ":theme", ":mood", ":stats", ":health",
		":synclog", ":devqueue", ":help", ":quit",
		"Complete task", "Mark task as blocked",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(content, entry) {
			t.Errorf("help content missing entry %q", entry)
		}
	}
}

func TestHelpView_View_ScrollIndicator(t *testing.T) {
	t.Parallel()
	hv := NewHelpView()
	view := hv.View()

	if !strings.Contains(view, "Showing lines") {
		t.Error("view should contain scroll position indicator")
	}
}

func TestHelpView_View_Empty(t *testing.T) {
	t.Parallel()
	hv := &HelpView{width: 80}
	view := hv.View()

	if !strings.Contains(view, "Help") {
		t.Error("view should contain 'Help' header")
	}
	if !strings.Contains(view, "No help content available") {
		t.Error("view should show empty message")
	}
}

func TestGolden_HelpView(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hv := NewHelpView()
	hv.SetWidth(80)
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
