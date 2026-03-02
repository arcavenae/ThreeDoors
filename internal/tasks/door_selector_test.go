package tasks

import (
	"math/rand/v2"
	"testing"
)

func TestDiversityScore(t *testing.T) {
	tests := []struct {
		name  string
		tasks []*Task
		want  int
	}{
		{"nil", nil, 0},
		{"empty", []*Task{}, 0},
		{
			"single task uncategorized",
			[]*Task{newTestTask("1", "t", StatusTodo, baseTime)},
			3, // 1 unique type + 1 unique effort + 1 unique location
		},
		{
			"single task categorized",
			[]*Task{newCategorizedTestTask("1", "t", StatusTodo, TypeTechnical, EffortMedium, LocationWork)},
			3,
		},
		{
			"two tasks all different",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeTechnical, EffortDeepWork, LocationWork),
			},
			6, // 2+2+2
		},
		{
			"perfect diversity score 9",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeAdministrative, EffortMedium, LocationWork),
				newCategorizedTestTask("3", "t3", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
			},
			9,
		},
		{
			"all same categories",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
				newCategorizedTestTask("3", "t3", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
			},
			3, // 1+1+1
		},
		{
			"mixed categorized and uncategorized",
			[]*Task{
				newCategorizedTestTask("1", "t1", StatusTodo, TypeCreative, EffortQuickWin, ""),
				newCategorizedTestTask("2", "t2", StatusTodo, TypeCreative, EffortDeepWork, ""),
				newCategorizedTestTask("3", "t3", StatusTodo, TypeTechnical, EffortQuickWin, ""),
			},
			5, // 2 types + 2 efforts + 1 location
		},
		{
			"all uncategorized",
			[]*Task{
				newTestTask("1", "t1", StatusTodo, baseTime),
				newTestTask("2", "t2", StatusTodo, baseTime),
				newTestTask("3", "t3", StatusTodo, baseTime),
			},
			3, // 1+1+1 (each "" is one unique value)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiversityScore(tt.tasks)
			if got != tt.want {
				t.Errorf("DiversityScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSelectDoorsWithRand_PrefersDiversity(t *testing.T) {
	// Pool with perfect diversity possible among 3 tasks
	pool := poolFromTasks(
		newCategorizedTestTask("t1", "Creative task", StatusTodo, TypeCreative, EffortQuickWin, LocationHome),
		newCategorizedTestTask("t2", "Admin task", StatusTodo, TypeAdministrative, EffortMedium, LocationWork),
		newCategorizedTestTask("t3", "Technical task", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
		newCategorizedTestTask("t4", "Technical task 2", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("t5", "Technical task 3", StatusTodo, TypeTechnical, EffortQuickWin, LocationWork),
	)

	rng := rand.New(rand.NewPCG(42, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	if len(selected) != 3 {
		t.Fatalf("expected 3 doors, got %d", len(selected))
	}

	score := DiversityScore(selected)
	if score < 7 {
		t.Errorf("expected high diversity score (>= 7), got %d", score)
	}
}

func TestSelectDoorsWithRand_FallbackFewTasks(t *testing.T) {
	pool := poolFromTasks(
		newTestTask("t1", "Task 1", StatusTodo, baseTime),
		newTestTask("t2", "Task 2", StatusTodo, baseTime),
	)

	rng := rand.New(rand.NewPCG(42, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	if len(selected) != 2 {
		t.Errorf("expected 2 doors (all available), got %d", len(selected))
	}
}

func TestSelectDoorsWithRand_EmptyPool(t *testing.T) {
	pool := NewTaskPool()
	rng := rand.New(rand.NewPCG(42, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	if selected != nil {
		t.Errorf("expected nil for empty pool, got %v", selected)
	}
}

func TestSelectDoorsWithRand_AllSameCategories(t *testing.T) {
	pool := poolFromTasks(
		newCategorizedTestTask("t1", "Task 1", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("t2", "Task 2", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("t3", "Task 3", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("t4", "Task 4", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
	)

	rng := rand.New(rand.NewPCG(42, 0))
	selected := selectDoorsWithRand(pool, 3, rng)

	if len(selected) != 3 {
		t.Fatalf("expected 3 doors, got %d", len(selected))
	}
	// All same categories → score = 3
	if score := DiversityScore(selected); score != 3 {
		t.Errorf("expected diversity score 3 for homogeneous pool, got %d", score)
	}
}

func TestSelectDoors_Integration_DiversePool(t *testing.T) {
	// Pool of 10 tasks with 4 types
	pool := poolFromTasks(
		newCategorizedTestTask("t1", "Design logo", StatusTodo, TypeCreative, EffortMedium, LocationWork),
		newCategorizedTestTask("t2", "File reports", StatusTodo, TypeAdministrative, EffortQuickWin, LocationWork),
		newCategorizedTestTask("t3", "Refactor auth", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
		newCategorizedTestTask("t4", "Buy groceries", StatusTodo, TypePhysical, EffortQuickWin, LocationErrands),
		newCategorizedTestTask("t5", "Write tests", StatusTodo, TypeTechnical, EffortMedium, LocationWork),
		newCategorizedTestTask("t6", "Plan sprint", StatusTodo, TypeAdministrative, EffortMedium, LocationWork),
		newCategorizedTestTask("t7", "Sketch wireframes", StatusTodo, TypeCreative, EffortDeepWork, LocationHome),
		newCategorizedTestTask("t8", "Clean desk", StatusTodo, TypePhysical, EffortQuickWin, LocationHome),
		newCategorizedTestTask("t9", "Debug API", StatusTodo, TypeTechnical, EffortDeepWork, LocationAnywhere),
		newCategorizedTestTask("t10", "Review PR", StatusTodo, TypeTechnical, EffortQuickWin, LocationAnywhere),
	)

	// Run multiple seeded iterations, check diversity is consistently good
	for seed := uint64(0); seed < 20; seed++ {
		// Reset pool recently shown buffer
		testPool := poolFromTasks(pool.GetAllTasks()...)

		rng := rand.New(rand.NewPCG(seed, 0))
		selected := selectDoorsWithRand(testPool, 3, rng)

		if len(selected) != 3 {
			t.Fatalf("seed %d: expected 3 doors, got %d", seed, len(selected))
		}

		score := DiversityScore(selected)
		if score < 5 {
			types := make(map[TaskType]bool)
			for _, s := range selected {
				types[s.Type] = true
			}
			t.Errorf("seed %d: low diversity score %d (types: %v)", seed, score, types)
		}
	}
}
