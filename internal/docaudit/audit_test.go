package docaudit

import (
	"testing"
)

func TestAudit_Clean(t *testing.T) {
	t.Parallel()

	docs := DocSet{
		StoryFiles: map[string]StoryEntry{
			"1.1": {ID: "1.1", Status: "Done", Source: "story_file"},
			"1.2": {ID: "1.2", Status: "Not Started", Source: "story_file"},
		},
		EpicsAndStories: map[string]StoryEntry{
			"1.1": {ID: "1.1", Status: "Referenced", Source: "epics_and_stories"},
			"1.2": {ID: "1.2", Status: "Referenced", Source: "epics_and_stories"},
		},
		Roadmap: map[string]StoryEntry{
			"1.1": {ID: "1.1", Status: "Done", Source: "roadmap"},
			"1.2": {ID: "1.2", Status: "Not Started", Source: "roadmap"},
		},
		EpicList: map[string]EpicEntry{
			"1": {ID: "1", Status: "Complete", Source: "epic_list"},
		},
		RoadmapEpics: map[string]EpicEntry{
			"1": {ID: "1", Status: "Complete", Source: "roadmap"},
		},
		EpicListEpics: map[string]EpicEntry{
			"1": {ID: "1", Status: "Complete", Source: "epic_list"},
		},
	}

	auditor := NewAuditor(docs)
	result := auditor.Run()

	if !result.Clean {
		t.Errorf("expected clean audit, got %d findings", len(result.Findings))
		for _, f := range result.Findings {
			t.Logf("  finding: %s — %s", f.Type, f.Description)
		}
	}
}

func TestAudit_StatusMismatch(t *testing.T) {
	t.Parallel()

	docs := DocSet{
		StoryFiles: map[string]StoryEntry{
			"1.1": {ID: "1.1", Status: "Done", Source: "story_file"},
		},
		EpicsAndStories: map[string]StoryEntry{
			"1.1": {ID: "1.1", Source: "epics_and_stories"},
		},
		Roadmap: map[string]StoryEntry{
			"1.1": {ID: "1.1", Status: "Not Started", Source: "roadmap"},
		},
		EpicList:      map[string]EpicEntry{},
		RoadmapEpics:  map[string]EpicEntry{},
		EpicListEpics: map[string]EpicEntry{},
	}

	auditor := NewAuditor(docs)
	result := auditor.Run()

	if result.Clean {
		t.Fatal("expected findings, got clean audit")
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == FindingStatusMismatch && f.StoryID == "1.1" {
			found = true
			if f.Expected != "Done" {
				t.Errorf("expected = %q, want %q", f.Expected, "Done")
			}
			if f.Actual != "Not Started" {
				t.Errorf("actual = %q, want %q", f.Actual, "Not Started")
			}
			if f.Authority != "story_file" {
				t.Errorf("authority = %q, want %q", f.Authority, "story_file")
			}
		}
	}
	if !found {
		t.Error("expected status mismatch finding for story 1.1")
	}
}

func TestAudit_OrphanedStory(t *testing.T) {
	t.Parallel()

	docs := DocSet{
		StoryFiles: map[string]StoryEntry{
			"99.1": {ID: "99.1", Status: "Done", Source: "story_file"},
		},
		EpicsAndStories: map[string]StoryEntry{},
		Roadmap:         map[string]StoryEntry{},
		EpicList:        map[string]EpicEntry{},
		RoadmapEpics:    map[string]EpicEntry{},
		EpicListEpics:   map[string]EpicEntry{},
	}

	auditor := NewAuditor(docs)
	result := auditor.Run()

	if result.Clean {
		t.Fatal("expected findings, got clean audit")
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == FindingOrphanedStory && f.StoryID == "99.1" {
			found = true
		}
	}
	if !found {
		t.Error("expected orphaned story finding for 99.1")
	}
}

func TestAudit_PhantomStory(t *testing.T) {
	t.Parallel()

	docs := DocSet{
		StoryFiles: map[string]StoryEntry{},
		EpicsAndStories: map[string]StoryEntry{
			"99.1": {ID: "99.1", Source: "epics_and_stories"},
		},
		Roadmap: map[string]StoryEntry{
			"99.2": {ID: "99.2", Source: "roadmap"},
		},
		EpicList:      map[string]EpicEntry{},
		RoadmapEpics:  map[string]EpicEntry{},
		EpicListEpics: map[string]EpicEntry{},
	}

	auditor := NewAuditor(docs)
	result := auditor.Run()

	if result.Clean {
		t.Fatal("expected findings, got clean audit")
	}

	phantomCount := 0
	for _, f := range result.Findings {
		if f.Type == FindingPhantomStory {
			phantomCount++
		}
	}
	if phantomCount != 2 {
		t.Errorf("got %d phantom findings, want 2", phantomCount)
	}
}

func TestAudit_EpicStatusMismatch(t *testing.T) {
	t.Parallel()

	docs := DocSet{
		StoryFiles:      map[string]StoryEntry{},
		EpicsAndStories: map[string]StoryEntry{},
		Roadmap:         map[string]StoryEntry{},
		EpicList: map[string]EpicEntry{
			"1": {ID: "1", Status: "Complete", Source: "epic_list"},
		},
		RoadmapEpics: map[string]EpicEntry{
			"1": {ID: "1", Status: "In Progress", Source: "roadmap"},
		},
		EpicListEpics: map[string]EpicEntry{
			"1": {ID: "1", Status: "Complete", Source: "epic_list"},
		},
	}

	auditor := NewAuditor(docs)
	result := auditor.Run()

	if result.Clean {
		t.Fatal("expected findings, got clean audit")
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == FindingEpicMismatch && f.EpicID == "1" {
			found = true
		}
	}
	if !found {
		t.Error("expected epic mismatch finding for epic 1")
	}
}

func TestStatusesMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		want bool
	}{
		{"Done", "Done", true},
		{"Done (PR #123)", "Done", true},
		{"Not Started", "Not Started", true},
		{"Done", "Not Started", false},
		{"In Progress", "In Review", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			t.Parallel()
			got := statusesMatch(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("statusesMatch(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
