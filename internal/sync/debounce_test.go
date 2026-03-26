package sync_test

import (
	"sync/atomic"
	"testing"
	"time"

	gosync "github.com/arcavenae/ThreeDoors/internal/sync"
)

// fakeTimer is a test timer that can be fired manually.
type fakeTimer struct {
	stopped  bool
	callback func()
}

func (f *fakeTimer) Stop() bool {
	wasPending := !f.stopped
	f.stopped = true
	return wasPending
}

func (f *fakeTimer) Fire() {
	if !f.stopped {
		f.callback()
	}
}

func TestDebouncer_SingleTrigger(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	var lastTimer *fakeTimer

	timerFunc := func(d time.Duration, f func()) gosync.Timer {
		ft := &fakeTimer{callback: f}
		lastTimer = ft
		return ft
	}

	d := gosync.NewDebouncer(30*time.Second, func() {
		callCount.Add(1)
	}, timerFunc)
	_ = d

	d.Trigger()
	if lastTimer == nil {
		t.Fatal("Trigger() should create a timer")
	}

	// Fire the timer
	lastTimer.Fire()

	if callCount.Load() != 1 {
		t.Errorf("callback called %d times, want 1", callCount.Load())
	}
}

func TestDebouncer_MultipleTriggers_CoalescesToOne(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	var timers []*fakeTimer

	timerFunc := func(d time.Duration, f func()) gosync.Timer {
		ft := &fakeTimer{callback: f}
		timers = append(timers, ft)
		return ft
	}

	d := gosync.NewDebouncer(30*time.Second, func() {
		callCount.Add(1)
	}, timerFunc)

	// Trigger 5 times rapidly
	for i := 0; i < 5; i++ {
		d.Trigger()
	}

	// Only the last timer should fire
	if len(timers) < 2 {
		t.Fatalf("expected multiple timer creations, got %d", len(timers))
	}

	// All timers except the last should be stopped
	for i := 0; i < len(timers)-1; i++ {
		if !timers[i].stopped {
			t.Errorf("timer[%d] should be stopped", i)
		}
	}

	// Fire the last timer
	timers[len(timers)-1].Fire()

	if callCount.Load() != 1 {
		t.Errorf("callback called %d times, want 1 (debounce should coalesce)", callCount.Load())
	}
}

func TestDebouncer_Stop_CancelsPending(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	var lastTimer *fakeTimer

	timerFunc := func(d time.Duration, f func()) gosync.Timer {
		ft := &fakeTimer{callback: f}
		lastTimer = ft
		return ft
	}

	d := gosync.NewDebouncer(30*time.Second, func() {
		callCount.Add(1)
	}, timerFunc)

	d.Trigger()
	d.Stop()

	// Fire should be a no-op after stop
	if lastTimer != nil {
		lastTimer.Fire()
	}

	if callCount.Load() != 0 {
		t.Errorf("callback called %d times after Stop(), want 0", callCount.Load())
	}
}

func TestDebouncer_TimerDuration(t *testing.T) {
	t.Parallel()

	var receivedDuration time.Duration

	timerFunc := func(d time.Duration, f func()) gosync.Timer {
		receivedDuration = d
		return &fakeTimer{callback: f}
	}

	d := gosync.NewDebouncer(30*time.Second, func() {}, timerFunc)
	d.Trigger()

	if receivedDuration != 30*time.Second {
		t.Errorf("timer duration = %v, want 30s", receivedDuration)
	}
}
