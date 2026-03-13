package core

// Ordering represents the causal relationship between two vector clocks.
type Ordering int

const (
	// Equal means both clocks have identical entries.
	Equal Ordering = iota
	// HappenedBefore means the first clock causally precedes the second.
	HappenedBefore
	// HappenedAfter means the first clock causally follows the second.
	HappenedAfter
	// Concurrent means neither clock causally precedes the other.
	Concurrent
)

// VectorClock tracks logical timestamps per device for causal ordering.
// Keys are device IDs (strings), values are monotonically increasing counters.
type VectorClock map[string]uint64

// NewVectorClock creates an empty vector clock.
func NewVectorClock() VectorClock {
	return make(VectorClock)
}

// Increment advances the clock for the given device by one.
func (vc VectorClock) Increment(deviceID string) {
	vc[deviceID]++
}

// Merge takes the element-wise maximum of this clock and another,
// updating this clock in place.
func (vc VectorClock) Merge(other VectorClock) {
	for device, ts := range other {
		if ts > vc[device] {
			vc[device] = ts
		}
	}
}

// Compare determines the causal ordering between this clock and another.
// Returns HappenedBefore if every entry in vc is <= the corresponding entry
// in other AND at least one is strictly less.
// Returns HappenedAfter if every entry in other is <= vc with at least one less.
// Returns Equal if all entries match.
// Returns Concurrent if neither happened-before the other.
func (vc VectorClock) Compare(other VectorClock) Ordering {
	// Collect all device IDs from both clocks.
	devices := make(map[string]struct{})
	for d := range vc {
		devices[d] = struct{}{}
	}
	for d := range other {
		devices[d] = struct{}{}
	}

	hasLess := false
	hasGreater := false

	for d := range devices {
		a := vc[d]
		b := other[d]
		if a < b {
			hasLess = true
		}
		if a > b {
			hasGreater = true
		}
		if hasLess && hasGreater {
			return Concurrent
		}
	}

	switch {
	case !hasLess && !hasGreater:
		return Equal
	case hasLess && !hasGreater:
		return HappenedBefore
	default:
		return HappenedAfter
	}
}

// Copy returns a deep copy of the vector clock.
func (vc VectorClock) Copy() VectorClock {
	if vc == nil {
		return nil
	}
	cp := make(VectorClock, len(vc))
	for k, v := range vc {
		cp[k] = v
	}
	return cp
}
