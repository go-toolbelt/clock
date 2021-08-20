package clock

import (
	"errors"
	"sync"
	"time"
)

type sleeper struct {
	i     int
	until time.Time
	woke  bool
	c     chan time.Time
	f     func()
}

func (s *sleeper) wake() {
	if s.woke {
		return
	}
	s.woke = true

	// if c is set, send the current time
	if s.c != nil {
		s.c <- s.until
	}

	// if f is set, call it in separate goroutine
	if s.f != nil {
		s.f()
	}
}

type blocker struct {
	n    int
	done chan struct{}
}

type fakeClock struct {
	mutex    sync.RWMutex
	at       time.Time
	sleepers []*sleeper
	blockers []blocker
}

func NewFakeClock() FakeClock {
	return NewFakeClockAt(time.Unix(1, 0))
}

func NewFakeClockAt(at time.Time) FakeClock {
	return &fakeClock{
		at: at,
	}
}

func (clock *fakeClock) Now() time.Time {
	clock.mutex.RLock()
	defer clock.mutex.RUnlock()

	return clock.at
}

func (clock *fakeClock) Since(t time.Time) time.Duration {
	return clock.Now().Sub(t)
}

func (clock *fakeClock) Sleep(d time.Duration) {
	<-clock.After(d)
}

func (clock *fakeClock) After(d time.Duration) <-chan time.Time {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	if d < 0 {
		d = 0
	}

	c := make(chan time.Time, 1)
	clock.appendSleeper(&sleeper{
		until: clock.at.Add(d),
		c:     c,
	})
	return c
}

func (clock *fakeClock) AfterFunc(d time.Duration, f func()) Timer {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	timer := &fakeTimer{
		clock: clock,
		sleeper: sleeper{
			until: clock.at.Add(d),
			f:     func() { go f() },
		},
	}
	clock.appendSleeper(&timer.sleeper)

	return timer
}

type fakeTimer struct {
	clock   *fakeClock
	stopped bool
	sleeper sleeper
}

func (clock *fakeClock) NewTimer(d time.Duration) Timer {
	return &fakeTimer{
		clock: clock,
		sleeper: sleeper{
			i:     -1,
			until: clock.Now().Add(d),
			c:     make(chan time.Time, 1),
		},
	}
}

func (timer *fakeTimer) C() <-chan time.Time {
	clock := timer.clock

	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	sleeper := &timer.sleeper

	if sleeper.i < 0 {
		clock.appendSleeper(sleeper)
	}

	return sleeper.c
}

func (timer *fakeTimer) Stop() bool {
	clock := timer.clock

	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	defer func() { timer.stopped = true }()
	if timer.stopped {
		return false
	}

	sleeper := &timer.sleeper

	if clock.removeSleeper(sleeper) {
		return true
	}

	return sleeper.until.After(clock.at)
}

func (timer *fakeTimer) Reset(d time.Duration) bool {
	clock := timer.clock

	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	sleeper := &timer.sleeper

	if d < 0 {
		d = 0
	}

	sleeper.until = timer.clock.at.Add(d)
	sleeper.woke = false
	sleeper.c = make(chan time.Time, 1)

	defer func() {
		if sleeper.f != nil {
			clock.appendSleeper(&timer.sleeper)
		}
	}()

	return clock.removeSleeper(sleeper)
}

type fakeTicker struct {
	clock    *fakeClock
	interval time.Duration
	next     time.Time
	stopped  bool
	sleeper  *sleeper
}

var errNonPositiveInterval = errors.New("non-positive interval for NewTicker")

func (clock *fakeClock) NewTicker(d time.Duration) Ticker {
	if d <= 0 {
		panic(errNonPositiveInterval)
	}

	return &fakeTicker{
		clock:    clock,
		interval: d,
		next:     clock.Now().Add(d),
		sleeper: &sleeper{
			i: -1,
		},
	}
}

func (ticker *fakeTicker) C() <-chan time.Time {
	clock := ticker.clock

	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	c := make(chan time.Time, 1)
	if ticker.stopped {
		return c
	}

	ticker.sleeper = &sleeper{

		until: ticker.next,
		c:     c,
	}
	clock.appendSleeper(ticker.sleeper)
	ticker.next = ticker.next.Add(ticker.interval)

	return c
}

func (ticker *fakeTicker) Stop() {
	clock := ticker.clock

	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	ticker.stopped = true
	clock.removeSleeper(ticker.sleeper)
}

func (ticker *fakeTicker) Reset(d time.Duration) {
	ticker.Stop()
	ticker.stopped = false
	ticker.interval = d
	ticker.sleeper.until = ticker.clock.Now().Add(d)
}

func (clock *fakeClock) Tick(d time.Duration) func() <-chan time.Time {
	if d <= 0 {
		return func() <-chan time.Time { return nil }
	}

	return clock.NewTicker(d).C
}

func (clock *fakeClock) Advance(d time.Duration) {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	// time travel is not allowed
	if d <= 0 {
		return
	}

	clock.at = clock.at.Add(d)
	clock.checkSleepers()
}

func (clock *fakeClock) Until(n int) <-chan struct{} {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	done := make(chan struct{})
	if len(clock.sleepers) >= n {
		close(done)
		return done
	}

	clock.appendBlocker(blocker{
		n:    n,
		done: done,
	})
	return done
}

func (clock *fakeClock) BlockUntil(n int) {
	<-clock.Until(n)
}

func (clock *fakeClock) appendSleeper(s *sleeper) {
	if !clock.at.Before(s.until) {
		s.i = -1
		s.wake()
		return
	}

	s.i = len(clock.sleepers)
	clock.sleepers = append(clock.sleepers, s)
	clock.checkBlockers()
}

func (clock *fakeClock) removeSleeper(s *sleeper) bool {
	i := s.i

	if i < 0 {
		return false
	}

	// Replace the sleeper with the last sleeper
	clock.sleepers[i] = clock.sleepers[len(clock.sleepers)-1]
	// Update the replacing sleeper's i
	clock.sleepers[i].i = i
	// nil out the last reference
	clock.sleepers[len(clock.sleepers)-1] = nil
	// make the sleeper index negative
	s.i = -1
	// Shrink the sleeper slice
	clock.sleepers = clock.sleepers[:len(clock.sleepers)-1]

	return true
}

func (clock *fakeClock) checkSleepers() {
	oldSleepers := clock.sleepers
	clock.sleepers = clock.sleepers[:0]
	for _, sleeper := range oldSleepers {
		clock.appendSleeper(sleeper)
	}
}

func (clock *fakeClock) appendBlocker(b blocker) {
	clock.blockers = append(clock.blockers, b)
}

func (clock *fakeClock) checkBlockers() {
	n := 0
	for _, blocker := range clock.blockers {
		if len(clock.sleepers) < blocker.n {
			clock.blockers[n] = blocker
			n++
		} else {
			close(blocker.done)
		}
	}
	clock.blockers = clock.blockers[:n]
}
