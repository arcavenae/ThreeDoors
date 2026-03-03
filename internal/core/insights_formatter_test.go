package core

import (
	"strings"
	"testing"
)

func TestFormatMoodInsights_NilReport(t *testing.T) {
	result := FormatMoodInsights(nil)
	expected := "Not enough data yet — need at least 5 sessions for insights."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatMoodInsights_EmptyCorrelations(t *testing.T) {
	report := &PatternReport{
		SessionCount:     10,
		MoodCorrelations: []MoodCorrelation{},
	}
	result := FormatMoodInsights(report)
	expected := "No mood correlation data yet. Log moods during sessions (press M) to build patterns. Need at least 3 sessions with the same mood."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatMoodInsights_PopulatedCorrelations(t *testing.T) {
	report := &PatternReport{
		SessionCount: 25,
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", SessionCount: 12, PreferredType: "technical", PreferredEffort: "deep-work", AvgTasksCompleted: 3.2},
			{Mood: "stressed", SessionCount: 8, PreferredType: "administrative", PreferredEffort: "quick-win", AvgTasksCompleted: 2.1},
		},
	}
	result := FormatMoodInsights(report)

	// Check header
	if !strings.Contains(result, "Mood Correlation Analysis") {
		t.Error("Expected header 'Mood Correlation Analysis'")
	}

	// Check mood entries
	if !strings.Contains(result, "focused") {
		t.Error("Expected 'focused' mood in output")
	}
	if !strings.Contains(result, "stressed") {
		t.Error("Expected 'stressed' mood in output")
	}
	if !strings.Contains(result, "technical") {
		t.Error("Expected 'technical' type in output")
	}
	if !strings.Contains(result, "administrative") {
		t.Error("Expected 'administrative' type in output")
	}
	if !strings.Contains(result, "deep-work") {
		t.Error("Expected 'deep-work' effort in output")
	}
	if !strings.Contains(result, "quick-win") {
		t.Error("Expected 'quick-win' effort in output")
	}
}

func TestFormatMoodInsights_PartialData_EmptyPreferredType(t *testing.T) {
	report := &PatternReport{
		SessionCount: 15,
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", SessionCount: 8, PreferredType: "technical", PreferredEffort: "deep-work", AvgTasksCompleted: 3.2},
			{Mood: "tired", SessionCount: 4, PreferredType: "", PreferredEffort: "", AvgTasksCompleted: 1.0},
		},
	}
	result := FormatMoodInsights(report)

	// Empty PreferredType should show "-"
	if !strings.Contains(result, "focused") {
		t.Error("Expected 'focused' mood in output")
	}
	if !strings.Contains(result, "tired") {
		t.Error("Expected 'tired' mood in output")
	}
	// The "-" placeholder should appear for empty fields
	if !strings.Contains(result, "-") {
		t.Error("Expected '-' placeholder for empty PreferredType")
	}
}

func TestFormatMoodInsights_IdentifiesMostProductive(t *testing.T) {
	report := &PatternReport{
		SessionCount: 20,
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", SessionCount: 10, PreferredType: "technical", PreferredEffort: "deep-work", AvgTasksCompleted: 4.5},
			{Mood: "stressed", SessionCount: 8, PreferredType: "administrative", PreferredEffort: "quick-win", AvgTasksCompleted: 1.5},
		},
	}
	result := FormatMoodInsights(report)

	// Should identify most productive mood
	if !strings.Contains(result, "focused") {
		t.Error("Expected insight about most productive mood 'focused'")
	}
}

func TestFormatMoodInsights_SingleCorrelation(t *testing.T) {
	report := &PatternReport{
		SessionCount: 10,
		MoodCorrelations: []MoodCorrelation{
			{Mood: "calm", SessionCount: 5, PreferredType: "creative", PreferredEffort: "medium", AvgTasksCompleted: 2.0},
		},
	}
	result := FormatMoodInsights(report)

	if !strings.Contains(result, "Mood Correlation Analysis") {
		t.Error("Expected header")
	}
	if !strings.Contains(result, "calm") {
		t.Error("Expected 'calm' mood")
	}
	if !strings.Contains(result, "creative") {
		t.Error("Expected 'creative' type")
	}
}
