package clock

import "time"

type Clock interface {
	// Now returns the current local time.
	Now() time.Time

	// Since returns the time elapsed since t.
	// It is shorthand for clock.Now().Sub(t).
	Since(t time.Time) time.Duration

	// NewTimer creates a new Timer that will send
	// the current time on its channel after at least duration d.
	NewTimer(d time.Duration) Timer

	// Sleep pauses the current goroutine for at least the duration d.
	// A negative or zero duration causes Sleep to return immediately.
	Sleep(d time.Duration)

	// After waits for the duration to elapse and then sends the current time
	// on the returned channel.
	// It is equivalent to NewTimer(d).C.
	// The underlying Timer is not recovered by the garbage collector
	// until the timer fires. If efficiency is a concern, use NewTimer
	// instead and call Timer.Stop if the timer is no longer needed.
	After(d time.Duration) <-chan time.Time

	// AfterFunc waits for the duration to elapse and then calls f
	// in its own goroutine. It returns a Timer that can
	// be used to cancel the call using its Stop method.
	AfterFunc(d time.Duration, f func()) Timer

	// NewTicker returns a new Ticker containing a channel that will send
	// the time on the channel after each tick. The period of the ticks is
	// specified by the duration argument. The ticker will adjust the time
	// interval or drop ticks to make up for slow receivers.
	// The duration d must be greater than zero; if not, NewTicker will
	// panic. Stop the ticker to release associated resources.
	NewTicker(d time.Duration) Ticker

	// Tick is a convenience wrapper for NewTicker providing access to the ticking
	// channel only. While Tick is useful for clients that have no need to shut down
	// the Ticker, be aware that without a way to shut it down the underlying
	// Ticker cannot be recovered by the garbage collector; it "leaks".
	// Unlike NewTicker, Tick will return nil if d <= 0.
	Tick(d time.Duration) func() <-chan time.Time
}

type FakeClock interface {
	Clock

	// Advance increments the time in the clock by d.
	// If d < 0, this call is a noop.
	// Time travel is not allowed.
	Advance(d time.Duration)

	// Until waits until n goroutines are blocked on the clock.
	// The returned channel is then closed
	Until(n int) <-chan struct{}

	// BlockUntil blocks until n goroutines are blocked on the clock.
	// It's a convenience method for `<-clock.Until(n)`.
	BlockUntil(n int)
}

// The Timer type represents a single event.
// When the Timer expires, the current time will be sent on C,
// unless the Timer was created by AfterFunc.
// A Timer must be created with clock.NewTimer or clock.AfterFunc.
type Timer interface {
	// C returns the channel on which the time is delivered.
	C() <-chan time.Time

	// Stop prevents the Timer from firing.
	// It returns true if the call stops the timer, false if the timer has already
	// expired or been stopped.
	// Stop does not close the channel, to prevent a read from the channel succeeding
	// incorrectly.
	//
	// To ensure the channel is empty after a call to Stop, check the
	// return value and drain the channel.
	// For example, assuming the program has not received from t.C already:
	//
	// 	if !t.Stop() {
	// 		<-t.C
	// 	}
	//
	// This cannot be done concurrent to other receives from the Timer's
	// channel or other calls to the Timer's Stop method.
	//
	// For a timer created with AfterFunc(d, f), if t.Stop returns false, then the timer
	// has already expired and the function f has been started in its own goroutine;
	// Stop does not wait for f to complete before returning.
	// If the caller needs to know whether f is completed, it must coordinate
	// with f explicitly.
	Stop() bool

	// Reset changes the timer to expire after duration d.
	// It returns true if the timer had been active, false if the timer had
	// expired or been stopped.
	//
	// For a Timer created with NewTimer, Reset should be invoked only on
	// stopped or expired timers with drained channels.
	//
	// If a program has already received a value from t.C, the timer is known
	// to have expired and the channel drained, so t.Reset can be used directly.
	// If a program has not yet received a value from t.C, however,
	// the timer must be stopped and—if Stop reports that the timer expired
	// before being stopped—the channel explicitly drained:
	//
	// 	if !t.Stop() {
	// 		<-t.C
	// 	}
	// 	t.Reset(d)
	//
	// This should not be done concurrent to other receives from the Timer's
	// channel.
	//
	// Note that it is not possible to use Reset's return value correctly, as there
	// is a race condition between draining the channel and the new timer expiring.
	// Reset should always be invoked on stopped or expired channels, as described above.
	// The return value exists to preserve compatibility with existing programs.
	//
	// For a Timer created with AfterFunc(d, f), Reset either reschedules
	// when f will run, in which case Reset returns true, or schedules f
	// to run again, in which case it returns false.
	// When Reset returns false, Reset neither waits for the prior f to
	// complete before returning nor does it guarantee that the subsequent
	// goroutine running f does not run concurrently with the prior
	// one. If the caller needs to know whether the prior execution of
	// f is completed, it must coordinate with f explicitly.
	Reset(d time.Duration) bool
}

// A Ticker holds a channel that delivers ``ticks'' of a clock at intervals.
type Ticker interface {
	// C returns the channel on which the ticks are delivered.
	// Note. The caller must save the output of C instead of calling it repeatedly.
	// It's not guaranteed that subsequent calls will return the same channel.
	// Re-calling C before recieing a tick will result in lost ticks.
	C() <-chan time.Time

	// Stop turns off a ticker. After Stop, no more ticks will be sent.
	// Stop does not close the channel, to prevent a concurrent goroutine
	// reading from the channel from seeing an erroneous "tick".
	Stop()

	// Reset stops a ticker and resets its period to the specified duration.
	// The next tick will arrive after the new period elapses.
	Reset(d time.Duration)
}
