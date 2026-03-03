package core

import (
	"strings"
	"testing"
)

func TestFormatAvoidanceInsights_NilReport(t *testing.T) {
	result := FormatAvoidanceInsights(nil)
	if !strings.Contains(result, "Not enough data") {
		t.Errorf("Expected encouragement message for nil report, got: %s", result)
	}
}

func TestFormatAvoidanceInsights_NoAvoidance(t *testing.T) {
	report := &PatternReport{
		SessionCount:  10,
		AvoidanceList: []AvoidanceEntry{},
	}
	result := FormatAvoidanceInsights(report)
	if !strings.Contains(result, "No significant avoidance") {
		t.Errorf("Expected no avoidance message, got: %s", result)
	}
}

func TestFormatAvoidanceInsights_WithAvoidance(t *testing.T) {
	report := &PatternReport{
		SessionCount: 10,
		AvoidanceList: []AvoidanceEntry{
			{TaskText: "Write tests", TimesBypassed: 7, TimesShown: 10, NeverSelected: false},
			{TaskText: "Clean garage", TimesBypassed: 12, TimesShown: 15, NeverSelected: true},
		},
	}
	result := FormatAvoidanceInsights(report)
	if !strings.Contains(result, "Tasks bypassed 5+ times: 2") {
		t.Errorf("Expected count of 2 tasks, got: %s", result)
	}
	if !strings.Contains(result, "Write tests") {
		t.Errorf("Expected task text in output, got: %s", result)
	}
	if !strings.Contains(result, "never selected") {
		t.Errorf("Expected 'never selected' label for task, got: %s", result)
	}
	if !strings.Contains(result, "Persistent avoidance") {
		t.Errorf("Expected persistent avoidance section for 10+ task, got: %s", result)
	}
}

func TestFormatAvoidanceInsights_BelowThreshold(t *testing.T) {
	report := &PatternReport{
		SessionCount: 10,
		AvoidanceList: []AvoidanceEntry{
			{TaskText: "Minor task", TimesBypassed: 3, TimesShown: 5},
		},
	}
	result := FormatAvoidanceInsights(report)
	if !strings.Contains(result, "No significant avoidance") {
		t.Errorf("Expected no avoidance for tasks with <5 bypasses, got: %s", result)
	}
}

func TestFormatAvoidanceInsights_WithMoodAvoidance(t *testing.T) {
	report := &PatternReport{
		SessionCount: 10,
		AvoidanceList: []AvoidanceEntry{
			{TaskText: "Tech task", TimesBypassed: 8, TimesShown: 12},
		},
		MoodCorrelations: []MoodCorrelation{
			{Mood: "stressed", SessionCount: 5, AvoidedType: "technical"},
			{Mood: "happy", SessionCount: 4, AvoidedType: ""},
		},
	}
	result := FormatAvoidanceInsights(report)
	if !strings.Contains(result, "When stressed, you tend to bypass technical tasks") {
		t.Errorf("Expected mood-avoidance pattern, got: %s", result)
	}
	// Should not show "happy" since no avoided type
	if strings.Contains(result, "happy") {
		t.Errorf("Should not include moods without avoided type, got: %s", result)
	}
}
