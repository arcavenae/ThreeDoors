package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// --- Keybinding Bar Toggle (h key) ---

func TestHKey_TogglesKeyHints(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.doorsView.SetWidth(80)
	m.doorsView.SetHeight(24)
	m.doorsView.SetThemeByName("classic")

	// Key hints on by default.
	if !m.showKeyHints {
		t.Fatal("expected key hints visible by default")
	}

	// In doors view, hints should appear on doors.
	view := m.View()
	if !strings.Contains(view, "[a]") {
		t.Error("expected door hints visible when key hints enabled")
	}

	// Press 'h' to toggle off.
	m.Update(keyMsg("h"))
	if m.showKeyHints {
		t.Error("expected key hints hidden after pressing h")
	}

	view = m.View()
	if strings.Contains(view, "[a]") {
		t.Error("expected no door hints when key hints disabled")
	}

	// Press 'h' again to toggle back on.
	m.Update(keyMsg("h"))
	if !m.showKeyHints {
		t.Error("expected key hints visible after pressing h again")
	}
}

func TestHKey_SuppressedDuringTextInput(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")

	// Enter search mode (text input active).
	m.searchView = m.newSearchView()
	m.viewMode = ViewSearch

	barBefore := m.showKeyHints
	m.Update(keyMsg("h"))
	if m.showKeyHints != barBefore {
		t.Error("'h' should not toggle bar during text input")
	}
}

func TestHKey_SuppressedWhenOverlayVisible(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true

	barBefore := m.showKeyHints
	m.Update(keyMsg("h"))
	// Overlay intercepts all keys — bar should not toggle.
	if m.showKeyHints != barBefore {
		t.Error("'h' should not toggle bar when overlay is visible")
	}
}

// --- Keybinding Overlay Toggle (? key) ---

func TestQuestionMark_OpensOverlay(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24

	if m.showKeybindingOverlay {
		t.Fatal("overlay should be hidden initially")
	}

	m.Update(keyMsg("?"))
	if !m.showKeybindingOverlay {
		t.Error("expected overlay to open after pressing ?")
	}
	if m.keybindingOverlay == nil {
		t.Fatal("expected keybindingOverlay to be created")
		return
	}
	if m.keybindingOverlay.state.ViewMode != ViewDoors {
		t.Errorf("expected overlay ViewMode to be ViewDoors, got %d", m.keybindingOverlay.state.ViewMode)
	}
}

func TestQuestionMark_ClosesOverlay(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true
	m.keybindingOverlay = NewKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)

	m.Update(keyMsg("?"))
	if m.showKeybindingOverlay {
		t.Error("expected overlay to close after pressing ?")
	}
	if m.keybindingOverlay != nil {
		t.Error("expected keybindingOverlay to be nil on overlay dismiss")
	}
}

func TestEsc_ClosesOverlay(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true
	m.keybindingOverlay = NewKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)

	m.Update(keyMsg("esc"))
	if m.showKeybindingOverlay {
		t.Error("expected overlay to close after pressing esc")
	}
	if m.keybindingOverlay != nil {
		t.Error("expected keybindingOverlay to be nil on overlay dismiss")
	}
}

func TestQuestionMark_SuppressedDuringTextInput(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")

	// Enter add task mode (text input active).
	m.addTaskView = NewAddTaskView()
	m.viewMode = ViewAddTask

	m.Update(keyMsg("?"))
	if m.showKeybindingOverlay {
		t.Error("'?' should not open overlay during text input")
	}
}

// --- Overlay Key Interception ---

func TestOverlay_ScrollKeys(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true
	m.keybindingOverlay = NewKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)

	initialOffset := m.keybindingOverlay.viewport.YOffset

	// Scroll down.
	m.Update(keyMsg("j"))
	if m.keybindingOverlay.viewport.YOffset <= initialOffset {
		t.Error("expected viewport to scroll down after j")
	}

	offset1 := m.keybindingOverlay.viewport.YOffset
	m.Update(keyMsg("down"))
	if m.keybindingOverlay.viewport.YOffset <= offset1 {
		t.Error("expected viewport to scroll down after down")
	}

	// Scroll up.
	offset2 := m.keybindingOverlay.viewport.YOffset
	m.Update(keyMsg("k"))
	if m.keybindingOverlay.viewport.YOffset >= offset2 {
		t.Error("expected viewport to scroll up after k")
	}

	offset3 := m.keybindingOverlay.viewport.YOffset
	m.Update(keyMsg("up"))
	if m.keybindingOverlay.viewport.YOffset >= offset3 {
		t.Error("expected viewport to scroll up after up")
	}
}

func TestOverlay_ConsumesOtherKeys(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true
	m.keybindingOverlay = NewKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)

	// These keys should be consumed — no view change.
	for _, key := range []string{"a", "w", "d", "s", "q", "n"} {
		m.Update(keyMsg(key))
		if !m.showKeybindingOverlay {
			t.Errorf("overlay should still be visible after pressing %q", key)
		}
		if m.viewMode != ViewDoors {
			t.Errorf("view should not change while overlay is visible, changed to %d after %q", m.viewMode, key)
		}
	}
}

// --- View() Integration ---

func TestView_OverlayReplacesNormalView(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true
	m.keybindingOverlay = NewKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)

	view := m.View()
	if !strings.Contains(view, "KEYBINDING REFERENCE") {
		t.Error("expected overlay content in view output when overlay is visible")
	}
}

func TestView_BarContentChangesWithView(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24

	// DoorsView bar should show door-related bindings.
	doorsView := m.View()

	// Navigate to detail view.
	m.Update(keyMsg("a"))     // select door
	m.Update(keyMsg("enter")) // enter detail

	if m.viewMode != ViewDetail {
		t.Fatalf("expected ViewDetail, got %d", m.viewMode)
	}

	detailView := m.View()

	// The bar content should differ between views.
	// Doors shows "select door", detail shows "complete".
	if doorsView == detailView {
		t.Error("expected different bar content for DoorsView vs DetailView")
	}
}

func TestView_BarHiddenInDoorsView(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeyHints = true

	// In doors view, the keybinding bar should never appear (inline hints instead).
	view := m.View()
	lines := strings.Split(view, "\n")
	lastLine := lines[len(lines)-1]
	if strings.Contains(lastLine, "? help") || strings.Contains(lastLine, "? Help") {
		t.Error("keybinding bar should not appear in doors view")
	}
}

func TestView_BarHiddenWhenHintsDisabled(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeyHints = false

	// Navigate to detail view where bar would normally show.
	m.Update(keyMsg("a"))
	m.Update(keyMsg("enter"))

	view := m.View()
	lines := strings.Split(view, "\n")
	lastLine := lines[len(lines)-1]
	if strings.Contains(lastLine, "? help") || strings.Contains(lastLine, "? Help") {
		t.Error("bar content should not appear when hints are disabled")
	}
}

// --- Height Adjustment ---

func TestContentHeight_AdjustsWhenBarVisible(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.height = 24
	m.showKeyHints = true

	// Doors view: no bar, full height even when hints are on
	m.viewMode = ViewDoors
	if h := m.contentHeight(); h != 24 {
		t.Errorf("expected content height 24 in doors view (no bar), got %d", h)
	}

	// Non-door view: bar deducted
	m.viewMode = ViewDetail
	if h := m.contentHeight(); h != 22 {
		t.Errorf("expected content height 22 in detail view with bar, got %d", h)
	}

	m.showKeyHints = false
	if h := m.contentHeight(); h != 24 {
		t.Errorf("expected content height 24 without hints, got %d", h)
	}
}

func TestContentHeight_NoReductionForSmallTerminals(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.height = 8 // below barHeightHidden (10)
	m.showKeyHints = true
	m.viewMode = ViewDetail // non-door view

	if h := m.contentHeight(); h != 8 {
		t.Errorf("expected no height reduction for small terminal, got %d", h)
	}
}

// --- Config Persistence ---

func TestSaveKeyHintsCmd_PersistsToConfig(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write an initial config.
	cfg := &core.ProviderConfig{Provider: "textfile"}
	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	m := makeModel("task1", "task2", "task3")
	m.configPath = configPath

	// Generate the save command and execute it.
	show := false
	cmd := m.saveKeyHintsCmd(show)
	if cmd == nil {
		t.Fatal("expected non-nil command")
		return
	}

	// Execute the command (it returns a tea.Msg).
	cmd()

	// Verify config was updated.
	loaded, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
		return
	}
	if loaded.ShowKeyHints == nil {
		t.Fatal("expected ShowKeyHints to be set in config")
		return
	}
	if *loaded.ShowKeyHints != false {
		t.Error("expected ShowKeyHints to be false")
	}
}

func TestSaveKeyHintsCmd_NilWhenNoConfigPath(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	// configPath is empty — command should be a no-op.
	cmd := m.saveKeyHintsCmd(true)
	if cmd == nil {
		t.Fatal("expected non-nil command")
		return
	}
	// Execute should return nil (no-op).
	result := cmd()
	if result != nil {
		t.Error("expected nil result when configPath is empty")
	}
}

// --- SetShowKeyHints ---

func TestSetShowKeyHints(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	if !m.showKeyHints {
		t.Fatal("expected default true")
	}

	m.SetShowKeyHints(false)
	if m.showKeyHints {
		t.Error("expected false after SetShowKeyHints(false)")
	}
}

// --- Config ShowKeyHints field and migration ---

func TestProviderConfig_ShowKeyHintsField(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test default: absent field returns nil pointer.
	if err := os.WriteFile(configPath, []byte("provider: textfile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
		return
	}
	if cfg.ShowKeyHints != nil {
		t.Error("expected nil ShowKeyHints when absent from config")
	}

	// Test migration: old show_keybinding_bar migrates to show_key_hints.
	if err := os.WriteFile(configPath, []byte("provider: textfile\nshow_keybinding_bar: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
		return
	}
	if cfg.ShowKeyHints == nil || *cfg.ShowKeyHints != false {
		t.Error("expected ShowKeyHints to be false after migration from show_keybinding_bar")
	}

	// Test explicit show_key_hints true.
	if err := os.WriteFile(configPath, []byte("provider: textfile\nshow_key_hints: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
		return
	}
	if cfg.ShowKeyHints == nil || *cfg.ShowKeyHints != true {
		t.Error("expected ShowKeyHints to be true")
	}
}

// --- Overlay ViewMode Context ---

func TestOverlay_OpensWithCurrentViewContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(m *MainModel)
		wantMode ViewMode
	}{
		{
			name:     "doors view",
			setup:    func(m *MainModel) {},
			wantMode: ViewDoors,
		},
		{
			name: "detail view",
			setup: func(m *MainModel) {
				m.Update(keyMsg("a"))
				m.Update(keyMsg("enter"))
			},
			wantMode: ViewDetail,
		},
		{
			name: "health view",
			setup: func(m *MainModel) {
				m.healthView = NewHealthView(core.HealthCheckResult{})
				m.viewMode = ViewHealth
			},
			wantMode: ViewHealth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := makeModel("task1", "task2", "task3")
			m.width = 80
			m.height = 24
			tt.setup(m)
			m.Update(keyMsg("?"))
			if m.keybindingOverlay == nil {
				t.Fatal("expected keybindingOverlay to be created")
				return
			}
			if m.keybindingOverlay.state.ViewMode != tt.wantMode {
				t.Errorf("expected overlay ViewMode %d, got %d", tt.wantMode, m.keybindingOverlay.state.ViewMode)
			}
		})
	}
}

// --- Ensure 'h' handler works in all non-text-input views ---

func TestHKey_WorksInMultipleViews(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		setup func(m *MainModel)
	}{
		{
			name:  "doors view",
			setup: func(m *MainModel) {},
		},
		{
			name: "detail view",
			setup: func(m *MainModel) {
				m.Update(keyMsg("a"))
				m.Update(keyMsg("enter"))
			},
		},
		{
			name: "health view",
			setup: func(m *MainModel) {
				m.healthView = NewHealthView(core.HealthCheckResult{})
				m.viewMode = ViewHealth
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := makeModel("task1", "task2", "task3")
			m.width = 80
			m.height = 24
			tt.setup(m)

			if !m.showKeyHints {
				t.Fatal("expected key hints visible initially")
			}
			m.Update(keyMsg("h"))
			if m.showKeyHints {
				t.Error("expected key hints hidden after pressing h")
			}
		})
	}
}

// --- Race Detector Validation ---

func TestKeybindingIntegration_NoRaceConditions(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24

	// Simulate a rapid sequence of operations.
	m.Update(keyMsg("?"))
	_ = m.View()
	m.Update(keyMsg("j"))
	_ = m.View()
	m.Update(keyMsg("?"))
	_ = m.View()
	m.Update(keyMsg("h"))
	_ = m.View()
	m.Update(keyMsg("h"))
	_ = m.View()
}
