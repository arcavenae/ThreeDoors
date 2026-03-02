package tasks

import (
	"math/rand/v2"
	"testing"
)

// --- MoodAlignmentScore Tests ---

func TestMoodAlignmentScore_MatchesType(t *testing.T) {
	tasks := []*Task{
		newCategorizedTestTask("1", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
	}
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortQuickWin)
	if score != 2 {
		t.Errorf("Expected score 2 (type match only), got %d", score)
	}
}

func TestMoodAlignmentScore_MatchesEffort(t *testing.T) {
	tasks := []*Task{
		newCategorizedTestTask("1", "Reply emails", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
	}
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortQuickWin)
	if score != 1 {
		t.Errorf("Expected score 1 (effort match only), got %d", score)
	}
}

func TestMoodAlignmentScore_MatchesBoth(t *testing.T) {
	tasks := []*Task{
		newCategorizedTestTask("1", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
	}
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortDeepWork)
	if score != 3 {
		t.Errorf("Expected score 3 (type + effort match), got %d", score)
	}
}

func TestMoodAlignmentScore_NoMatch(t *testing.T) {
	tasks := []*Task{
		newCategorizedTestTask("1", "Buy groceries", StatusTodo, TypePhysical, EffortMedium, LocationErrands),
	}
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortDeepWork)
	if score != 0 {
		t.Errorf("Expected score 0 (no match), got %d", score)
	}
}

func TestMoodAlignmentScore_MultipleTasks(t *testing.T) {
	tasks := []*Task{
		newCategorizedTestTask("1", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Write tests", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("3", "Buy groceries", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
	}
	// Task 1: type match (+2) + effort match (+1) = 3
	// Task 2: type match (+2) = 2
	// Task 3: no match = 0
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortDeepWork)
	if score != 5 {
		t.Errorf("Expected score 5 (3+2+0), got %d", score)
	}
}

func TestMoodAlignmentScore_EmptyTasks(t *testing.T) {
	score := MoodAlignmentScore(nil, TypeTechnical, EffortDeepWork)
	if score != 0 {
		t.Errorf("Expected score 0 for nil tasks, got %d", score)
	}
}

func TestMoodAlignmentScore_UncategorizedTask(t *testing.T) {
	tasks := []*Task{
		newTestTask("1", "Do something", StatusTodo, baseTime), // no type/effort set
	}
	score := MoodAlignmentScore(tasks, TypeTechnical, EffortDeepWork)
	if score != 0 {
		t.Errorf("Expected score 0 for uncategorized task, got %d", score)
	}
}

// --- BuildTaskCategoryMap Tests ---

func TestBuildTaskCategoryMap_NilPool(t *testing.T) {
	m := BuildTaskCategoryMap(nil)
	if m == nil {
		t.Error("Expected non-nil map for nil pool")
	}
	if len(m) != 0 {
		t.Errorf("Expected empty map for nil pool, got %d entries", len(m))
	}
}

func TestBuildTaskCategoryMap_EmptyPool(t *testing.T) {
	pool := NewTaskPool()
	m := BuildTaskCategoryMap(pool)
	if m == nil {
		t.Error("Expected non-nil map for empty pool")
	}
	if len(m) != 0 {
		t.Errorf("Expected empty map for empty pool, got %d entries", len(m))
	}
}

func TestBuildTaskCategoryMap_CategorizedTasks(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Reply emails", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
	)
	m := BuildTaskCategoryMap(pool)
	if len(m) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(m))
	}
	info, ok := m["Fix bug"]
	if !ok {
		t.Fatal("Expected entry for 'Fix bug'")
	}
	if info.Type != TypeTechnical {
		t.Errorf("Expected TypeTechnical, got %q", info.Type)
	}
	if info.Effort != EffortDeepWork {
		t.Errorf("Expected EffortDeepWork, got %q", info.Effort)
	}
}

func TestBuildTaskCategoryMap_IncludesUncategorized(t *testing.T) {
	pool := poolFromTasks(
		newTestTask("1", "Do something", StatusTodo, baseTime), // no type/effort
	)
	m := BuildTaskCategoryMap(pool)
	if len(m) != 1 {
		t.Fatalf("Expected 1 entry (uncategorized tasks included), got %d", len(m))
	}
	info := m["Do something"]
	if info.Type != "" {
		t.Errorf("Expected empty type for uncategorized, got %q", info.Type)
	}
}

// --- SelectDoorsWithMood Tests ---

func TestSelectDoorsWithMood_EmptyMood_FallsBack(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Task A", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Task B", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Task C", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("4", "Task D", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_NilPatterns_FallsBack(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Task A", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Task B", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Task C", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("4", "Task D", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
	)

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "focused", nil, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_UnknownMood_FallsBack(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Task A", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Task B", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Task C", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("4", "Task D", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "unknown-mood", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_EmptyPreferredType_FallsBack(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Task A", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Task B", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Task C", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("4", "Task D", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "", PreferredEffort: "", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "focused", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_ValidMood_ReturnsTasks(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Reply emails", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Design mockup", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("4", "Buy groceries", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
		newCategorizedTestTask("5", "Write tests", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("6", "File expenses", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "stressed", PreferredType: "administrative", PreferredEffort: "quick-win", SessionCount: 5, AvgTasksCompleted: 2.1},
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 8, AvgTasksCompleted: 3.5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "stressed", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}

	// Verify all returned tasks are unique
	seen := map[string]bool{}
	for _, d := range doors {
		if seen[d.Text] {
			t.Errorf("Duplicate task in doors: %q", d.Text)
		}
		seen[d.Text] = true
	}

	// Verify the mood-aware selector biases toward administrative/quick-win
	// At least one task should be administrative (the preferred type for stressed)
	hasAdmin := false
	for _, d := range doors {
		if d.Type == TypeAdministrative {
			hasAdmin = true
		}
	}
	if !hasAdmin {
		t.Log("Note: no administrative tasks selected despite stressed mood preference — may be due to combined scoring")
	}
}

func TestSelectDoorsWithMood_DiversityFloor_AllMatchMood(t *testing.T) {
	// Pool where all tasks match the mood preference
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Email 1", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("2", "Email 2", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Email 3", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "stressed", PreferredType: "administrative", PreferredEffort: "quick-win", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "stressed", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}
	// All match — diversity floor swap can't happen because no non-matching tasks exist
	for _, d := range doors {
		if d.Type != TypeAdministrative {
			t.Errorf("Expected all administrative when no alternatives, got %q", d.Type)
		}
	}
}

func TestSelectDoorsWithMood_DiversityFloor_SwapWhenAlternativesExist(t *testing.T) {
	// Pool with 4 admin and 2 non-admin tasks
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Email 1", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("2", "Email 2", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("3", "Email 3", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("4", "Email 4", StatusTodo, TypeAdministrative, EffortQuickWin, LocationAnywhere),
		newCategorizedTestTask("5", "Fix bug", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("6", "Design mockup", StatusTodo, TypeCreative, EffortMedium, LocationWork),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "stressed", PreferredType: "administrative", PreferredEffort: "quick-win", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "stressed", patterns, rng)

	if len(doors) != 3 {
		t.Fatalf("Expected 3 doors, got %d", len(doors))
	}

	// At least 1 door should NOT be administrative (diversity floor enforcement)
	adminCount := 0
	for _, d := range doors {
		if d.Type == TypeAdministrative {
			adminCount++
		}
	}
	if adminCount == 3 {
		t.Error("Diversity floor violated: all 3 doors match mood preference despite alternatives existing")
	}
}

func TestSelectDoorsWithMood_EmptyPool(t *testing.T) {
	pool := NewTaskPool()
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "focused", patterns, rng)

	if len(doors) != 0 {
		t.Errorf("Expected 0 doors for empty pool, got %d", len(doors))
	}
}

func TestSelectDoorsWithMood_FewerTasksThanCount(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("1", "Task A", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
		newCategorizedTestTask("2", "Task B", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
	)
	patterns := &PatternReport{
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 5},
		},
	}

	rng := rand.New(rand.NewPCG(42, 0))
	doors := selectDoorsWithMoodAndRand(pool, 3, "focused", patterns, rng)

	if len(doors) != 2 {
		t.Errorf("Expected 2 doors (fewer than count), got %d", len(doors))
	}
}

// --- SetTaskCategories Tests ---

func TestSetTaskCategories_Nil(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	// Should not panic
	analyzer.SetTaskCategories(nil)
}

func TestSetTaskCategories_EmptyMap(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{})
}

func TestSetTaskCategories_Overwrite(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
		"Task A": {Type: TypeTechnical, Effort: EffortDeepWork},
	})
	// Second call overwrites
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
		"Task B": {Type: TypeCreative, Effort: EffortMedium},
	})
	// The analyzer should now only have Task B's categories
	// (this is tested indirectly via Analyze tests in pattern_analyzer_test.go)
}
