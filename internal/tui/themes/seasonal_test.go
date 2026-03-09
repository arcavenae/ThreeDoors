package themes

import (
	"testing"
	"time"
)

func TestResolveSeason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"spring start March 1", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), "spring"},
		{"spring end May 31", time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC), "spring"},
		{"spring mid April 15", time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), "spring"},
		{"summer start June 1", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "summer"},
		{"summer end August 31", time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), "summer"},
		{"autumn start September 1", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), "autumn"},
		{"autumn end November 30", time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), "autumn"},
		{"winter start December 1", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), "winter"},
		{"winter wrap January 15", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), "winter"},
		{"winter wrap February 28", time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC), "winter"},
		{"leap year February 29", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), "winter"},
		{"boundary spring to summer May 31", time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC), "spring"},
		{"boundary summer start June 1", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "summer"},
		{"boundary autumn to winter Nov 30", time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), "autumn"},
		{"boundary winter start Dec 1", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), "winter"},
		{"winter December 31", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), "winter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveSeason(tt.date, DefaultSeasonRanges)
			if got != tt.want {
				t.Errorf("ResolveSeason(%v) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestResolveSeason_EmptyRanges(t *testing.T) {
	t.Parallel()
	got := ResolveSeason(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), nil)
	if got != "" {
		t.Errorf("ResolveSeason with nil ranges = %q, want empty", got)
	}
}

func TestResolveSeason_CustomRanges(t *testing.T) {
	t.Parallel()
	custom := []SeasonRange{
		{Name: "holiday", Start: MonthDay{12, 20}, End: MonthDay{1, 5}},
	}

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"in holiday Dec 25", time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC), "holiday"},
		{"in holiday Jan 1", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "holiday"},
		{"outside holiday Feb 1", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveSeason(tt.date, custom)
			if got != tt.want {
				t.Errorf("ResolveSeason(%v) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestRegistryGetBySeason(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	springTheme := &DoorTheme{Name: "spring-flowers", Season: "spring"}
	nonSeasonal := &DoorTheme{Name: "classic", Season: ""}
	r.Register(springTheme)
	r.Register(nonSeasonal)

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		got, ok := r.GetBySeason("spring")
		if !ok {
			t.Fatal("expected to find spring theme")
		}
		if got.Name != "spring-flowers" {
			t.Errorf("got theme %q, want %q", got.Name, "spring-flowers")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, ok := r.GetBySeason("winter")
		if ok {
			t.Error("expected no winter theme")
		}
	})

	t.Run("empty season", func(t *testing.T) {
		t.Parallel()
		_, ok := r.GetBySeason("")
		if ok {
			t.Error("expected no match for empty season")
		}
	})
}

func TestMonthDay_ZeroValue(t *testing.T) {
	t.Parallel()
	theme := &DoorTheme{Name: "classic"}
	if theme.Season != "" {
		t.Errorf("zero-value Season = %q, want empty", theme.Season)
	}
	if theme.SeasonStart != (MonthDay{}) {
		t.Errorf("zero-value SeasonStart = %v, want zero", theme.SeasonStart)
	}
}

func BenchmarkResolveSeason(b *testing.B) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for b.Loop() {
		ResolveSeason(now, DefaultSeasonRanges)
	}
}
