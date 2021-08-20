package clock

import (
	"time"
)

type realClock struct{}

func NewRealClock() Clock {
	return realClock{}
}

// Now returns the current local time.
func (realClock) Now() time.Time {
	return time.Now()
}

func (realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (realClock) Tick(d time.Duration) func() <-chan time.Time {
	// nolint: staticcheck
	c := time.Tick(d)

	return func() <-chan time.Time { return c }
}

func (realClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

type realTimer struct {
	*time.Timer
}

func (timer realTimer) C() <-chan time.Time {
	return timer.Timer.C
}

func (realClock) AfterFunc(d time.Duration, f func()) Timer {
	return realTimer{
		Timer: time.AfterFunc(d, f),
	}
}

func (r realClock) NewTimer(d time.Duration) Timer {
	return realTimer{
		Timer: time.NewTimer(d),
	}
}

type realTicker struct {
	*time.Ticker
}

func (ticker realTicker) C() <-chan time.Time {
	return ticker.Ticker.C
}

func (r realClock) NewTicker(d time.Duration) Ticker {
	return realTicker{
		Ticker: time.NewTicker(d),
	}
}
