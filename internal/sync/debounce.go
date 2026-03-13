package sync

import "time"

// Timer is the interface for a stoppable timer (matches *time.Timer).
type Timer interface {
	Stop() bool
}

// TimerFunc creates a new timer that calls f after duration d.
type TimerFunc func(d time.Duration, f func()) Timer

// Debouncer coalesces rapid events into a single callback after a quiet period.
type Debouncer struct {
	delay     time.Duration
	timerFunc TimerFunc
	callback  func()
	timer     Timer
}

// NewDebouncer creates a debouncer that fires callback after delay of quiet.
func NewDebouncer(delay time.Duration, callback func(), timerFunc TimerFunc) *Debouncer {
	if timerFunc == nil {
		timerFunc = func(d time.Duration, f func()) Timer {
			return time.AfterFunc(d, f)
		}
	}
	return &Debouncer{
		delay:     delay,
		timerFunc: timerFunc,
		callback:  callback,
	}
}

// Trigger resets the debounce timer. The callback fires after delay of quiet.
func (d *Debouncer) Trigger() {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = d.timerFunc(d.delay, d.callback)
}

// Stop cancels any pending callback.
func (d *Debouncer) Stop() {
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
