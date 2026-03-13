package core

import "testing"

func TestVectorClock_Increment(t *testing.T) {
	t.Parallel()

	vc := NewVectorClock()
	vc.Increment("deviceA")
	if vc["deviceA"] != 1 {
		t.Errorf("expected 1, got %d", vc["deviceA"])
	}
	vc.Increment("deviceA")
	if vc["deviceA"] != 2 {
		t.Errorf("expected 2, got %d", vc["deviceA"])
	}
	vc.Increment("deviceB")
	if vc["deviceB"] != 1 {
		t.Errorf("expected 1, got %d", vc["deviceB"])
	}
}

func TestVectorClock_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    VectorClock
		b    VectorClock
		want VectorClock
	}{
		{
			name: "disjoint devices",
			a:    VectorClock{"A": 3},
			b:    VectorClock{"B": 5},
			want: VectorClock{"A": 3, "B": 5},
		},
		{
			name: "overlapping take max",
			a:    VectorClock{"A": 3, "B": 2},
			b:    VectorClock{"A": 1, "B": 5},
			want: VectorClock{"A": 3, "B": 5},
		},
		{
			name: "merge with empty",
			a:    VectorClock{"A": 3},
			b:    VectorClock{},
			want: VectorClock{"A": 3},
		},
		{
			name: "equal clocks",
			a:    VectorClock{"A": 3, "B": 2},
			b:    VectorClock{"A": 3, "B": 2},
			want: VectorClock{"A": 3, "B": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vc := tt.a.Copy()
			vc.Merge(tt.b)
			for device, expected := range tt.want {
				if vc[device] != expected {
					t.Errorf("device %s: expected %d, got %d", device, expected, vc[device])
				}
			}
			if len(vc) != len(tt.want) {
				t.Errorf("expected %d entries, got %d", len(tt.want), len(vc))
			}
		})
	}
}

func TestVectorClock_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    VectorClock
		b    VectorClock
		want Ordering
	}{
		{
			name: "equal clocks",
			a:    VectorClock{"A": 3, "B": 2},
			b:    VectorClock{"A": 3, "B": 2},
			want: Equal,
		},
		{
			name: "both empty",
			a:    VectorClock{},
			b:    VectorClock{},
			want: Equal,
		},
		{
			name: "a happened before b",
			a:    VectorClock{"A": 2, "B": 2},
			b:    VectorClock{"A": 3, "B": 2},
			want: HappenedBefore,
		},
		{
			name: "a happened after b",
			a:    VectorClock{"A": 3, "B": 2},
			b:    VectorClock{"A": 2, "B": 2},
			want: HappenedAfter,
		},
		{
			name: "concurrent - mixed ordering",
			a:    VectorClock{"A": 3, "B": 2},
			b:    VectorClock{"A": 2, "B": 4},
			want: Concurrent,
		},
		{
			name: "a has extra device all greater",
			a:    VectorClock{"A": 3, "B": 2, "C": 1},
			b:    VectorClock{"A": 3, "B": 2},
			want: HappenedAfter,
		},
		{
			name: "b has extra device",
			a:    VectorClock{"A": 3},
			b:    VectorClock{"A": 3, "B": 1},
			want: HappenedBefore,
		},
		{
			name: "concurrent with extra devices",
			a:    VectorClock{"A": 3, "C": 1},
			b:    VectorClock{"A": 2, "B": 1},
			want: Concurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.a.Compare(tt.b)
			if got != tt.want {
				t.Errorf("Compare: expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestVectorClock_Copy(t *testing.T) {
	t.Parallel()

	t.Run("deep copy independence", func(t *testing.T) {
		t.Parallel()
		orig := VectorClock{"A": 3, "B": 2}
		cp := orig.Copy()
		cp.Increment("A")
		if orig["A"] != 3 {
			t.Errorf("original modified: expected A=3, got A=%d", orig["A"])
		}
		if cp["A"] != 4 {
			t.Errorf("copy not incremented: expected A=4, got A=%d", cp["A"])
		}
	})

	t.Run("nil copy", func(t *testing.T) {
		t.Parallel()
		var vc VectorClock
		cp := vc.Copy()
		if cp != nil {
			t.Errorf("expected nil copy of nil clock, got %v", cp)
		}
	})
}
