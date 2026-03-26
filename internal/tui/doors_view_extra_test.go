package tui

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestDoorsView_SetTimeContext(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	tc := &core.TimeContext{}
	dv.SetTimeContext(tc)

	if dv.TimeContext() != tc {
		t.Error("expected TimeContext to return the set value")
	}
}

func TestDoorsView_TimeContext_Nil(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	if dv.TimeContext() != nil {
		t.Error("expected nil TimeContext by default")
	}
}

func TestDoorsView_SetPendingConflicts(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	dv.SetPendingConflicts(3)
	if dv.pendingConflicts != 3 {
		t.Errorf("expected 3, got %d", dv.pendingConflicts)
	}
}

func TestDoorsView_SetSyncTracker(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	tracker := core.NewSyncStatusTracker()
	dv.SetSyncTracker(tracker)
	if dv.syncTracker != tracker {
		t.Error("expected sync tracker to be set")
	}
}

func TestDoorsView_SetInsightsData(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	cc := core.NewCompletionCounter()
	dv.SetInsightsData(nil, cc)

	if dv.completionCounter != cc {
		t.Error("expected completion counter to be set")
	}
}

func TestDoorsView_SetAvoidanceData_Nil(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	dv.SetAvoidanceData(nil)
	if len(dv.avoidanceMap) != 0 {
		t.Error("expected empty avoidance map for nil report")
	}
}

func TestDoorsView_SetAvoidanceData_WithReport(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	report := &core.PatternReport{
		AvoidanceList: []core.AvoidanceEntry{
			{TaskText: "avoided task", TimesBypassed: 5, TimesShown: 10},
		},
	}
	dv.SetAvoidanceData(report)
	if dv.avoidanceMap["avoided task"] != 5 {
		t.Errorf("expected bypass count 5, got %d", dv.avoidanceMap["avoided task"])
	}
	if dv.avoidanceShown["avoided task"] != 10 {
		t.Errorf("expected shown count 10, got %d", dv.avoidanceShown["avoided task"])
	}
}
