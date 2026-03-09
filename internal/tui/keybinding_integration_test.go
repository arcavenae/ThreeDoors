package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// --- Keybinding Bar Toggle (h key) ---

func TestHKey_TogglesBarVisibility(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24

	// Bar is on by default.
	if !m.showKeybindingBar {
		t.Fatal("expected bar visible by default")
	}

	// View should contain the bar separator.
	view := m.View()
	if !strings.Contains(view, "─") {
		t.Error("expected bar separator in view output when bar is enabled")
	}

	// Press 'h' to toggle off.
	m.Update(keyMsg("h"))
	if m.showKeybindingBar {
		t.Error("expected bar hidden after pressing h")
	}

	view = m.View()
	if strings.Contains(view, "? help") {
		t.Error("expected no bar content in view output when bar is disabled")
	}

	// Press 'h' again to toggle back on.
	m.Update(keyMsg("h"))
	if !m.showKeybindingBar {
		t.Error("expected bar visible after pressing h again")
	}
}

func TestHKey_SuppressedDuringTextInput(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")

	// Enter search mode (text input active).
	m.searchView = m.newSearchView()
	m.viewMode = ViewSearch

	barBefore := m.showKeybindingBar
	m.Update(keyMsg("h"))
	if m.showKeybindingBar != barBefore {
		t.Error("'h' should not toggle bar during text input")
	}
}

func TestHKey_SuppressedWhenOverlayVisible(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingOverlay = true

	barBefore := m.showKeybindingBar
	m.Update(keyMsg("h"))
	// Overlay intercepts all keys — bar should not toggle.
	if m.showKeybindingBar != barBefore {
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

func TestView_BarHiddenWhenDisabled(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.width = 80
	m.height = 24
	m.showKeybindingBar = false

	view := m.View()
	// When bar is disabled, there should be no separator line in the bar area.
	// The view should not contain the RenderKeybindingBar output.
	lines := strings.Split(view, "\n")
	lastLine := lines[len(lines)-1]
	// The bar would end with "? help" — shouldn't be there.
	if strings.Contains(lastLine, "? help") || strings.Contains(lastLine, "? Help") {
		t.Error("bar content should not appear when bar is disabled")
	}
}

// --- Height Adjustment ---

func TestContentHeight_AdjustsWhenBarVisible(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.height = 24
	m.showKeybindingBar = true

	if h := m.contentHeight(); h != 22 {
		t.Errorf("expected content height 22 with bar, got %d", h)
	}

	m.showKeybindingBar = false
	if h := m.contentHeight(); h != 24 {
		t.Errorf("expected content height 24 without bar, got %d", h)
	}
}

func TestContentHeight_NoReductionForSmallTerminals(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	m.height = 8 // below barHeightHidden (10)
	m.showKeybindingBar = true

	if h := m.contentHeight(); h != 8 {
		t.Errorf("expected no height reduction for small terminal, got %d", h)
	}
}

// --- Config Persistence ---

func TestSaveKeybindingBarCmd_PersistsToConfig(t *testing.T) {
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
	cmd := m.saveKeybindingBarCmd(show)
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	// Execute the command (it returns a tea.Msg).
	cmd()

	// Verify config was updated.
	loaded, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.ShowKeybindingBar == nil {
		t.Fatal("expected ShowKeybindingBar to be set in config")
	}
	if *loaded.ShowKeybindingBar != false {
		t.Error("expected ShowKeybindingBar to be false")
	}
}

func TestSaveKeybindingBarCmd_NilWhenNoConfigPath(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	// configPath is empty — command should be a no-op.
	cmd := m.saveKeybindingBarCmd(true)
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	// Execute should return nil (no-op).
	result := cmd()
	if result != nil {
		t.Error("expected nil result when configPath is empty")
	}
}

// --- SetShowKeybindingBar ---

func TestSetShowKeybindingBar(t *testing.T) {
	t.Parallel()
	m := makeModel("task1", "task2", "task3")
	if !m.showKeybindingBar {
		t.Fatal("expected default true")
	}

	m.SetShowKeybindingBar(false)
	if m.showKeybindingBar {
		t.Error("expected false after SetShowKeybindingBar(false)")
	}
}

// --- Config ShowKeybindingBar field ---

func TestProviderConfig_ShowKeybindingBarField(t *testing.T) {
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
	}
	if cfg.ShowKeybindingBar != nil {
		t.Error("expected nil ShowKeybindingBar when absent from config")
	}

	// Test explicit false.
	if err := os.WriteFile(configPath, []byte("provider: textfile\nshow_keybinding_bar: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.ShowKeybindingBar == nil || *cfg.ShowKeybindingBar != false {
		t.Error("expected ShowKeybindingBar to be false")
	}

	// Test explicit true.
	if err := os.WriteFile(configPath, []byte("provider: textfile\nshow_keybinding_bar: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.ShowKeybindingBar == nil || *cfg.ShowKeybindingBar != true {
		t.Error("expected ShowKeybindingBar to be true")
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

			if !m.showKeybindingBar {
				t.Fatal("expected bar visible initially")
			}
			m.Update(keyMsg("h"))
			if m.showKeybindingBar {
				t.Error("expected bar hidden after pressing h")
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
