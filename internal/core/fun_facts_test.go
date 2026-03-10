package core

import (
	"strings"
	"testing"
	"time"
)

// bannedWords are words that must never appear in fun facts (D-089 content rules).
var bannedWords = []string{
	"declined", "dropped", "worse", "less", "failed",
	"missing", "behind", "overdue",
}

func funFactFrozen(year int, month time.Month, day, hour int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	}
}

// richAnalyzer creates a PatternAnalyzer loaded with diverse session data
// suitable for generating a wide range of fun facts.
func richAnalyzer(t *testing.T, nowFunc func() time.Time) *PatternAnalyzer {
	t.Helper()
	pa := NewPatternAnalyzerWithNow(nowFunc)
	// Directly inject sessions for testing (avoid file I/O)
	pa.sessions = []SessionMetrics{
		{
			SessionID:      "s1",
			StartTime:      time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC),
			TasksCompleted: 3,
			DoorsViewed:    5,
			RefreshesUsed:  1,
			MoodEntries:    []MoodEntry{{Mood: "Focused", Timestamp: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)}},
			MoodEntryCount: 1,
			DoorSelections: []DoorSelectionRecord{
				{DoorPosition: 0, TaskText: "task-a"},
				{DoorPosition: 1, TaskText: "task-b"},
				{DoorPosition: 0, TaskText: "task-c"},
			},
		},
		{
			SessionID:      "s2",
			StartTime:      time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2026, 3, 2, 14, 45, 0, 0, time.UTC),
			TasksCompleted: 5,
			DoorsViewed:    8,
			RefreshesUsed:  2,
			MoodEntries:    []MoodEntry{{Mood: "Energized", Timestamp: time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)}},
			MoodEntryCount: 1,
			DoorSelections: []DoorSelectionRecord{
				{DoorPosition: 2, TaskText: "task-d"},
				{DoorPosition: 1, TaskText: "task-e"},
				{DoorPosition: 2, TaskText: "task-f"},
				{DoorPosition: 0, TaskText: "task-g"},
				{DoorPosition: 1, TaskText: "task-h"},
			},
		},
		{
			SessionID:      "s3",
			StartTime:      time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2026, 3, 3, 9, 20, 0, 0, time.UTC),
			TasksCompleted: 2,
			DoorsViewed:    4,
			RefreshesUsed:  0,
			MoodEntries:    []MoodEntry{{Mood: "Focused", Timestamp: time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC)}},
			MoodEntryCount: 1,
			DoorSelections: []DoorSelectionRecord{
				{DoorPosition: 0, TaskText: "task-i"},
				{DoorPosition: 2, TaskText: "task-j"},
			},
		},
		{
			SessionID:      "s4",
			StartTime:      time.Date(2026, 3, 5, 11, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2026, 3, 5, 11, 40, 0, 0, time.UTC),
			TasksCompleted: 4,
			DoorsViewed:    6,
			RefreshesUsed:  1,
			MoodEntries:    []MoodEntry{{Mood: "Focused", Timestamp: time.Date(2026, 3, 5, 11, 0, 0, 0, time.UTC)}},
			MoodEntryCount: 1,
			DoorSelections: []DoorSelectionRecord{
				{DoorPosition: 1, TaskText: "task-k"},
				{DoorPosition: 0, TaskText: "task-l"},
				{DoorPosition: 2, TaskText: "task-m"},
				{DoorPosition: 1, TaskText: "task-n"},
			},
		},
	}
	return pa
}

// richCounter creates a CompletionCounter with streak data.
func richCounter(t *testing.T, nowFunc func() time.Time) *CompletionCounter {
	t.Helper()
	cc := NewCompletionCounterWithNow(nowFunc)
	cc.dateCounts = map[string]int{
		"2026-03-01": 3,
		"2026-03-02": 5,
		"2026-03-03": 2,
		"2026-03-05": 4,
		"2026-03-06": 1,
		"2026-03-07": 3,
	}
	return cc
}

func TestNewFunFactGenerator(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	if gen == nil {
		t.Fatal("NewFunFactGenerator() returned nil")
		return
	}
}

func TestFunFactGenerator_Generate_ReturnsNonEmpty(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	fact := gen.Generate()
	if fact == "" {
		t.Error("Generate() returned empty string with rich data")
	}
}

func TestFunFactGenerator_Generate_DeterministicSameDay(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 10)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)
	gen := NewFunFactGenerator(pa, cc)

	fact1 := gen.Generate()

	// Same day, different hour
	now2 := funFactFrozen(2026, 3, 7, 22)
	pa2 := richAnalyzer(t, now2)
	cc2 := richCounter(t, now2)
	gen2 := NewFunFactGenerator(pa2, cc2)
	fact2 := gen2.Generate()

	if fact1 != fact2 {
		t.Errorf("same UTC day should produce same fact:\n  hour 10: %q\n  hour 22: %q", fact1, fact2)
	}
}

func TestFunFactGenerator_Generate_RotatesDifferentDay(t *testing.T) {
	// With a pool of multiple facts, different days should (usually) produce different facts.
	// We test several days to ensure at least one differs.
	now1 := funFactFrozen(2026, 3, 7, 14)
	pa1 := richAnalyzer(t, now1)
	cc1 := richCounter(t, now1)
	gen1 := NewFunFactGenerator(pa1, cc1)
	fact1 := gen1.Generate()

	differentFound := false
	for day := 8; day <= 20; day++ {
		nowN := funFactFrozen(2026, 3, day, 14)
		paN := richAnalyzer(t, nowN)
		ccN := richCounter(t, nowN)
		genN := NewFunFactGenerator(paN, ccN)
		factN := genN.Generate()
		if factN != fact1 {
			differentFound = true
			break
		}
	}
	if !differentFound {
		t.Error("different days should produce different facts (tested days 7-20)")
	}
}

func TestFunFactGenerator_Generate_EmptyPoolReturnsDefault(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	// Empty analyzer — no sessions
	pa := NewPatternAnalyzerWithNow(now)
	cc := NewCompletionCounterWithNow(now)

	gen := NewFunFactGenerator(pa, cc)
	fact := gen.Generate()
	expected := "Every door you open is a step forward!"
	if fact != expected {
		t.Errorf("empty pool should return default fact %q, got %q", expected, fact)
	}
}

func TestFunFactGenerator_BannedWords(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	pool := gen.BuildFactPool()

	if len(pool) == 0 {
		t.Fatal("buildFactPool() returned empty pool with rich data")
	}

	for i, fact := range pool {
		lower := strings.ToLower(fact)
		for _, word := range bannedWords {
			if strings.Contains(lower, word) {
				t.Errorf("fact[%d] contains banned word %q: %q", i, word, fact)
			}
		}
	}
}

func TestFunFactGenerator_BuildFactPool_NonEmpty(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	pool := gen.BuildFactPool()

	if len(pool) < 3 {
		t.Errorf("buildFactPool() should produce at least 3 facts with rich data, got %d", len(pool))
	}
}

func TestFunFactGenerator_BuildFactPool_FriendTest(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	pool := gen.BuildFactPool()

	// Facts should be positive/neutral observations, not prescriptions.
	prescriptiveWords := []string{"you should", "try to", "you need to", "you must"}
	for i, fact := range pool {
		lower := strings.ToLower(fact)
		for _, phrase := range prescriptiveWords {
			if strings.Contains(lower, phrase) {
				t.Errorf("fact[%d] is prescriptive (contains %q): %q", i, phrase, fact)
			}
		}
	}
}

func TestFunFactGenerator_BuildFactPool_ReferencesRealData(t *testing.T) {
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	pool := gen.BuildFactPool()

	// At least some facts should contain real numbers from the data
	hasNumber := false
	for _, fact := range pool {
		for _, ch := range fact {
			if ch >= '0' && ch <= '9' {
				hasNumber = true
				break
			}
		}
		if hasNumber {
			break
		}
	}
	if !hasNumber {
		t.Error("fact pool should contain at least one fact with real data (numbers)")
	}
}

func TestFunFactGenerator_SelectionAlgorithm(t *testing.T) {
	// Verify the selection uses yearDay + year*366 as documented
	now := funFactFrozen(2026, 3, 7, 14)
	pa := richAnalyzer(t, now)
	cc := richCounter(t, now)

	gen := NewFunFactGenerator(pa, cc)
	pool := gen.BuildFactPool()

	if len(pool) == 0 {
		t.Fatal("need non-empty pool for selection test")
	}

	// Calculate expected index
	nowTime := now()
	day := nowTime.YearDay() + nowTime.Year()*366
	expectedIdx := day % len(pool)
	expectedFact := pool[expectedIdx]

	actual := gen.Generate()
	if actual != expectedFact {
		t.Errorf("Generate() = %q, want pool[%d] = %q", actual, expectedIdx, expectedFact)
	}
}
