package themes

import "time"

// SeasonRange maps a season name to its start and end dates (inclusive).
type SeasonRange struct {
	Name  string
	Start MonthDay
	End   MonthDay
}

// DefaultSeasonRanges defines meteorological seasons.
var DefaultSeasonRanges = []SeasonRange{
	{Name: "spring", Start: MonthDay{3, 1}, End: MonthDay{5, 31}},
	{Name: "summer", Start: MonthDay{6, 1}, End: MonthDay{8, 31}},
	{Name: "autumn", Start: MonthDay{9, 1}, End: MonthDay{11, 30}},
	{Name: "winter", Start: MonthDay{12, 1}, End: MonthDay{2, 29}},
}

// ResolveSeason returns the season name for the given time based on the
// provided ranges. It handles year-boundary wrapping (e.g. winter spans
// December through February). Returns "" if no range matches.
func ResolveSeason(now time.Time, ranges []SeasonRange) string {
	m := int(now.Month())
	d := now.Day()

	for _, r := range ranges {
		if inRange(m, d, r.Start, r.End) {
			return r.Name
		}
	}
	return ""
}

// inRange checks whether month/day falls within [start, end] inclusive,
// handling year-boundary wrapping (start > end means the range crosses
// December→January).
func inRange(month, day int, start, end MonthDay) bool {
	cur := monthDayOrd(month, day)
	s := monthDayOrd(start.Month, start.Day)
	e := monthDayOrd(end.Month, end.Day)

	if s <= e {
		// Normal range (e.g. March 1 – May 31)
		return cur >= s && cur <= e
	}
	// Wrapping range (e.g. December 1 – February 28)
	return cur >= s || cur <= e
}

// monthDayOrd converts a month/day pair to an ordinal for comparison.
// Uses month*100+day so March 1 = 301, December 1 = 1201, etc.
func monthDayOrd(month, day int) int {
	return month*100 + day
}
