package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// planningStep tracks which step the planning flow is on.
type planningStep int

const (
	planningStepGuidance planningStep = iota
	planningStepReview
	planningStepSelect
	planningStepConfirm
)

// ShowPlanningMsg is sent to open the planning view.
type ShowPlanningMsg struct{}

// PlanningCompleteMsg is sent when the planning session finishes.
type PlanningCompleteMsg struct {
	FocusTasks []*core.Task
	Timestamp  time.Time
}

// PlanningCancelledMsg is sent when the user exits planning without confirming.
type PlanningCancelledMsg struct{}

// planningGuidanceDismissMsg fires when the user dismisses first-time guidance.
type planningGuidanceDismissMsg struct{}

// PlanningView orchestrates the 3-step planning flow: Review → Select → Confirm.
type PlanningView struct {
	step          planningStep
	pool          *core.TaskPool
	provider      core.TaskProvider
	reviewView    *ReviewView
	selectView    *SelectView
	confirmView   *ConfirmView
	reviewMetrics ReviewCompleteMsg
	startTime     time.Time
	width         int
	height        int
	showGuidance  bool
	configPath    string
}

// NewPlanningView creates a PlanningView that orchestrates the planning flow.
func NewPlanningView(pool *core.TaskPool, provider core.TaskProvider) *PlanningView {
	configPath := ""
	if dir, err := core.GetConfigDirPath(); err == nil {
		configPath = dir
	}

	pv := &PlanningView{
		pool:       pool,
		provider:   provider,
		startTime:  time.Now().UTC(),
		configPath: configPath,
	}

	// Check if this is the first planning session
	if pv.isFirstPlanning() {
		pv.step = planningStepGuidance
		pv.showGuidance = true
	} else {
		pv.step = planningStepReview
		pv.initReviewStep()
	}

	return pv
}

// SetWidth sets terminal width on all sub-views.
func (pv *PlanningView) SetWidth(w int) {
	pv.width = w
	if pv.reviewView != nil {
		pv.reviewView.SetWidth(w)
	}
	if pv.selectView != nil {
		pv.selectView.SetWidth(w)
	}
	if pv.confirmView != nil {
		pv.confirmView.SetWidth(w)
	}
}

// SetHeight sets terminal height on sub-views.
func (pv *PlanningView) SetHeight(h int) {
	pv.height = h
	if pv.selectView != nil {
		pv.selectView.SetHeight(h)
	}
	if pv.confirmView != nil {
		pv.confirmView.SetHeight(h)
	}
}

// Init initializes the current step.
func (pv *PlanningView) Init() tea.Cmd {
	if pv.step == planningStepReview && pv.reviewView != nil {
		return pv.reviewView.Init()
	}
	return nil
}

// Update handles messages and step transitions.
func (pv *PlanningView) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case planningGuidanceDismissMsg:
		pv.step = planningStepReview
		pv.initReviewStep()
		return pv.reviewView.Init()
	}

	// Handle step-specific messages
	switch pv.step {
	case planningStepGuidance:
		return pv.updateGuidance(msg)
	case planningStepReview:
		return pv.updateReview(msg)
	case planningStepSelect:
		return pv.updateSelect(msg)
	case planningStepConfirm:
		return pv.updateConfirm(msg)
	}
	return nil
}

func (pv *PlanningView) updateGuidance(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEscape:
			return func() tea.Msg { return PlanningCancelledMsg{} }
		default:
			// Any key dismisses guidance
			return func() tea.Msg { return planningGuidanceDismissMsg{} }
		}
	}
	return nil
}

func (pv *PlanningView) updateReview(msg tea.Msg) tea.Cmd {
	// Check for ReviewCompleteMsg to transition to Select step
	if rcm, ok := msg.(ReviewCompleteMsg); ok {
		pv.reviewMetrics = rcm
		pv.step = planningStepSelect
		pv.initSelectStep()
		return nil
	}

	if pv.reviewView == nil {
		return nil
	}
	return pv.reviewView.Update(msg)
}

func (pv *PlanningView) updateSelect(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case SelectCompleteMsg:
		pv.step = planningStepConfirm
		pv.initConfirmStep(msg.FocusTasks, msg.EnergyLevel, msg.EnergyOverride)
		return pv.confirmView.Init()
	case SelectCancelMsg:
		// Go back to review
		pv.step = planningStepReview
		pv.initReviewStep()
		return pv.reviewView.Init()
	}

	if pv.selectView == nil {
		return nil
	}
	return pv.selectView.Update(msg)
}

func (pv *PlanningView) updateConfirm(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ConfirmCompleteMsg:
		return pv.finishSession(msg)
	case ConfirmCancelMsg:
		// Go back to select
		pv.step = planningStepSelect
		pv.initSelectStep()
		return nil
	}

	if pv.confirmView == nil {
		return nil
	}
	return pv.confirmView.Update(msg)
}

func (pv *PlanningView) initReviewStep() {
	// Get incomplete tasks from past 24 hours
	since := time.Now().UTC().Add(-24 * time.Hour)
	tasks := pv.pool.GetIncompleteTasks(since)
	pv.reviewView = NewReviewView(tasks)
	pv.reviewView.SetWidth(pv.width)
}

func (pv *PlanningView) initSelectStep() {
	energy := core.InferEnergyFromTime(time.Now().UTC())
	allTasks := pv.pool.GetAvailableForDoors()
	pv.selectView = NewSelectView(allTasks, energy)
	pv.selectView.SetWidth(pv.width)
	pv.selectView.SetHeight(pv.height)
}

func (pv *PlanningView) initConfirmStep(
	focusTasks []*core.Task,
	energy core.EnergyLevel,
	energyOverride bool,
) {
	pv.confirmView = NewConfirmView(focusTasks, pv.reviewMetrics, energy, energyOverride, pv.startTime)
	pv.confirmView.SetWidth(pv.width)
	pv.confirmView.SetHeight(pv.height)
}

func (pv *PlanningView) finishSession(msg ConfirmCompleteMsg) tea.Cmd {
	now := time.Now().UTC()

	// Clear old focus tags
	core.ClearFocusTags(pv.pool)

	// Apply new focus tags
	for _, t := range msg.FocusTasks {
		if !core.HasFocusTag(t) {
			t.Text = t.Text + " +focus"
			t.UpdatedAt = now
		}
	}

	// Log metrics
	if pv.configPath != "" {
		sessionsPath := filepath.Join(pv.configPath, "sessions.jsonl")
		event := core.PlanningSessionEvent{
			Timestamp:        now,
			DurationSeconds:  time.Since(pv.startTime).Seconds(),
			TasksReviewed:    pv.reviewMetrics.Reviewed,
			TasksContinued:   pv.reviewMetrics.Continued,
			TasksDeferred:    pv.reviewMetrics.Deferred,
			TasksDropped:     pv.reviewMetrics.Dropped,
			FocusTaskCount:   len(msg.FocusTasks),
			EnergyLevel:      msg.EnergyLevel,
			EnergyOverridden: msg.EnergyOverride,
		}
		if err := core.LogPlanningSession(sessionsPath, event); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to log planning session: %v\n", err)
		}
	}

	// Save planning timestamp
	if pv.configPath != "" {
		savePlanningTimestamp(pv.configPath, now)
	}

	// Mark first-run as done
	if pv.showGuidance {
		pv.markGuidanceShown()
	}

	// Save tasks (focus tags are in task text)
	if pv.provider != nil {
		allTasks := pv.pool.GetAllTasks()
		if err := pv.provider.SaveTasks(allTasks); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save tasks after planning: %v\n", err)
		}
	}

	focusTasks := msg.FocusTasks
	return func() tea.Msg {
		return PlanningCompleteMsg{
			FocusTasks: focusTasks,
			Timestamp:  now,
		}
	}
}

// View renders the current planning step.
func (pv *PlanningView) View() string {
	switch pv.step {
	case planningStepGuidance:
		return pv.renderGuidance()
	case planningStepReview:
		if pv.reviewView != nil {
			return pv.reviewView.View()
		}
	case planningStepSelect:
		if pv.selectView != nil {
			return pv.selectView.View()
		}
	case planningStepConfirm:
		if pv.confirmView != nil {
			return pv.confirmView.View()
		}
	}
	return ""
}

func (pv *PlanningView) renderGuidance() string {
	w := pv.width - 6
	if w < 40 {
		w = 40
	}

	content := confirmHeaderStyle.Render("Welcome to Daily Planning!") + "\n\n" +
		"This 3-step flow helps you focus your day:\n\n" +
		"  1. Review — Look at incomplete tasks from yesterday\n" +
		"  2. Select — Pick up to 5 tasks to focus on today\n" +
		"  3. Confirm — Review your choices and start your day\n\n" +
		helpStyle.Render("Press any key to continue...")

	return detailBorder.Width(w).Render(content)
}

// isFirstPlanning checks whether a planning session has ever been run.
func (pv *PlanningView) isFirstPlanning() bool {
	if pv.configPath == "" {
		return false
	}
	guidancePath := filepath.Join(pv.configPath, "planning_guidance_shown")
	_, err := os.Stat(guidancePath)
	return os.IsNotExist(err)
}

func (pv *PlanningView) markGuidanceShown() {
	if pv.configPath == "" {
		return
	}
	guidancePath := filepath.Join(pv.configPath, "planning_guidance_shown")
	f, err := os.OpenFile(guidancePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return
	}
	_ = f.Close()
}

// savePlanningTimestamp writes the planning session timestamp for focus expiry.
func savePlanningTimestamp(configDir string, t time.Time) {
	path := filepath.Join(configDir, "planning_timestamp")
	_ = os.WriteFile(path, []byte(t.Format(time.RFC3339)), 0o600)
}

// LoadPlanningTimestamp reads the last planning session timestamp.
// Returns nil if no timestamp file exists or it can't be parsed.
func LoadPlanningTimestamp(configDir string) *time.Time {
	path := filepath.Join(configDir, "planning_timestamp")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return nil
	}
	return &t
}
