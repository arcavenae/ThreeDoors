package core

import (
	"testing"
	"time"
)

func TestAnalyzeMoodCorrelations_AvoidedType(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.SetTaskCategories(map[string]TaskCategoryInfo{
		"Write code":   {Type: TypeTechnical, Effort: "deep-work"},
		"Send emails":  {Type: TypeAdministrative, Effort: "quick-win"},
		"Draw mockups": {Type: TypeCreative, Effort: "medium"},
	})

	// Create sessions where "stressed" mood leads to bypassing technical tasks
	sessions := []SessionMetrics{}
	for i := 0; i < 4; i++ {
		sessions = append(sessions, SessionMetrics{
			StartTime:      time.Now().Add(time.Duration(-i) * time.Hour),
			TasksCompleted: 2,
			MoodEntries:    []MoodEntry{{Mood: "stressed"}},
			DoorSelections: []DoorSelectionRecord{{TaskText: "Send emails", DoorPosition: 0}},
			TaskBypasses:   [][]string{{"Write code", "Draw mockups"}},
		})
	}

	// Cold start: need 5 sessions, so first analyze returns nil report
	report, _ := pa.Analyze(sessions)
	if report != nil {
		t.Log("Note: report returned with < 5 sessions (cold start guard may have different threshold)")
	}

	// Add one more to reach 5
	sessions = append(sessions, SessionMetrics{
		StartTime:      time.Now().Add(-5 * time.Hour),
		TasksCompleted: 1,
		MoodEntries:    []MoodEntry{{Mood: "stressed"}},
		DoorSelections: []DoorSelectionRecord{{TaskText: "Send emails", DoorPosition: 1}},
		TaskBypasses:   [][]string{{"Write code"}},
	})

	var err error
	report, err = pa.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Find the stressed correlation
	found := false
	for _, mc := range report.MoodCorrelations {
		if mc.Mood == "stressed" {
			found = true
			// Technical tasks should be most bypassed
			if mc.AvoidedType != "technical" {
				t.Errorf("Expected AvoidedType 'technical' for stressed, got %q", mc.AvoidedType)
			}
			break
		}
	}
	if !found {
		t.Error("Expected stressed mood correlation in report")
	}
}

func TestAnalyzeMoodCorrelations_NoBypassData(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.SetTaskCategories(map[string]TaskCategoryInfo{
		"Write code": {Type: TypeTechnical},
	})

	sessions := make([]SessionMetrics, 5)
	for i := range sessions {
		sessions[i] = SessionMetrics{
			StartTime:      time.Now().Add(time.Duration(-i) * time.Hour),
			TasksCompleted: 1,
			MoodEntries:    []MoodEntry{{Mood: "happy"}},
			DoorSelections: []DoorSelectionRecord{{TaskText: "Write code", DoorPosition: 0}},
			TaskBypasses:   nil, // no bypasses
		}
	}

	report, err := pa.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, mc := range report.MoodCorrelations {
		if mc.Mood == "happy" && mc.AvoidedType != "" {
			t.Errorf("Expected empty AvoidedType when no bypasses, got %q", mc.AvoidedType)
		}
	}
}
