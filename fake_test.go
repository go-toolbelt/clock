package clock_test

import (
	"testing"
	"time"

	"github.com/go-toolbelt/clock"
)

func TestNow(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	assertClockAt(t, start, clock)
}

func TestAdvance(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	clock.Advance(1 * time.Second)
	assertClockAt(t, start.Add(1*time.Second), clock)
}

func TestSince_Positive(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	expected := 1 * time.Second
	actual := clock.Since(start.Add(-expected))
	if actual != expected {
		t.Errorf("expected %s got %s", expected, actual)
	}
}

func TestSince_Negative(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	expected := -1 * time.Second
	actual := clock.Since(start.Add(-expected))
	if actual != expected {
		t.Errorf("expected %s got %s", expected, actual)
	}
}

func TestNewTimer(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(2 * time.Second)
	c := timer.C()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertNotSent(t, c)

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(2*time.Second), c)
}

func TestNewTimer_CallCTwice(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)
	c := timer.C()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(1*time.Second), c)

	c = timer.C()
	assertNotSent(t, c)
}

func TestNewTimer_Stop(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(2 * time.Second)
	c := timer.C()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertNotSent(t, c)

	if !timer.Stop() {
		t.Error("expected stop to return true")
	}
	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	clock.Advance(1 * time.Second)

	assertNotSent(t, c)
}

func TestNewTimer_Stop_NeverCalledC(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(2 * time.Second)

	clock.Advance(1 * time.Second)

	if !timer.Stop() {
		t.Error("expected stop to return true")
	}
	if timer.Stop() {
		t.Error("expected stop to return false")
	}
}

func TestNewTimer_Stop_AfterFired(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)

	c := timer.C()

	clock.BlockUntil(1)
	clock.Advance(1 * time.Second)

	if timer.Stop() {
		t.Error("expected stop to return false")
	}
	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	assertSent(t, start.Add(1*time.Second), c)
}

func TestNewTimer_Reset_Positive(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)

	c := timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestNewTimer_Reset_Zero(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)

	c := timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(0) {
		t.Error("expected stop to return false")
	}

	c = timer.C()
	assertSent(t, start.Add(1*time.Second), c)
}

func TestNewTimer_Reset_Negative(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)

	c := timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(-1) {
		t.Error("expected stop to return false")
	}

	c = timer.C()
	assertSent(t, start.Add(1*time.Second), c)
}

func TestNewTimer_Reset_Stopped(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(2 * time.Second)
	c := timer.C()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)

	if !timer.Stop() {
		t.Error("expected stop to return true")
	}

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestNewTimer_Reset_Stopped_AfterFired(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	timer := clock.NewTimer(1 * time.Second)
	c := timer.C()

	clock.BlockUntil(1)
	clock.Advance(1 * time.Second)

	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = timer.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestSleep_Positive(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	woke := make(chan struct{})
	go func() {
		defer close(woke)
		clock.Sleep(1 * time.Second)
	}()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertClosed(t, woke)
}

func TestSleep_Positive_TwoSleepers(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	woke0 := make(chan struct{})

	go func() {
		defer close(woke0)
		clock.Sleep(1 * time.Second)
	}()

	woke1 := make(chan struct{})

	go func() {
		defer close(woke1)
		clock.Sleep(2 * time.Second)
	}()

	assertClockUntil(t, 2, clock)
	clock.Advance(1 * time.Second)
	assertClosed(t, woke0)
	assertNotClosed(t, woke1)

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertClosed(t, woke1)
}

func TestSleep_Zero(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	clock.Sleep(0) // should return immediately
}

func TestSleep_Negative(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	clock.Sleep(-1) // should return immediately
}

func TestAfter_Positive(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	after := clock.After(1 * time.Second)

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(1*time.Second), after)
}

func TestAfter_Positive_TwoSleepers(t *testing.T) {
	start := time.Unix(2, 0)
	clock := clock.NewFakeClockAt(start)

	after0 := clock.After(1 * time.Second)

	after1 := clock.After(2 * time.Second)

	assertClockUntil(t, 2, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(1*time.Second), after0)

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(2*time.Second), after1)
}

func TestAfter_Zero(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	after := clock.After(0)
	assertSent(t, start, after)
}

func TestAfter_Negative(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	after := clock.After(-1)
	assertSent(t, start, after)
}

func TestAfterFunc(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	clock.AfterFunc(2*time.Second, func() {
		c <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(2*time.Second), c)
}

func TestAfterFunc_Stop(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	timer := clock.AfterFunc(2*time.Second, func() {
		c <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertNotSent(t, c)

	if !timer.Stop() {
		t.Error("expected stop to return true")
	}
	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	clock.Advance(1 * time.Second)

	assertNotSent(t, c)
}

func TestAfterFunc_Stop_AfterFired(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	timer := clock.AfterFunc(1*time.Second, func() {
		c <- clock.Now()
	})

	clock.BlockUntil(1)
	clock.Advance(1 * time.Second)

	if timer.Stop() {
		t.Error("expected stop to return false")
	}
	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	assertSent(t, start.Add(1*time.Second), c)
}

func TestAfterFunc_Reset_Positive(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	getter := func() chan time.Time {
		return c
	}
	timer := clock.AfterFunc(1*time.Second, func() {
		getter() <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = make(chan time.Time)
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestAfterFunc_Reset_Zero(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	getter := func() chan time.Time {
		return c
	}
	timer := clock.AfterFunc(1*time.Second, func() {
		getter() <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	c = make(chan time.Time)
	if timer.Reset(0) {
		t.Error("expected stop to return false")
	}
	assertSent(t, start.Add(1*time.Second), c)
}

func TestAfterFunc_Reset_Negative(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	getter := func() chan time.Time {
		return c
	}
	timer := clock.AfterFunc(1*time.Second, func() {
		getter() <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	assertSent(t, start.Add(1*time.Second), c)

	c = make(chan time.Time)
	if timer.Reset(-1) {
		t.Error("expected stop to return false")
	}
	assertSent(t, start.Add(1*time.Second), c)
}

func TestAfterFunc_Reset_Stopped(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	getter := func() chan time.Time {
		return c
	}
	timer := clock.AfterFunc(2*time.Second, func() {
		getter() <- clock.Now()
	})

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)

	if !timer.Stop() {
		t.Error("expected stop to return true")
	}

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = make(chan time.Time)
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestAfterFunc_Reset_Stopped_AfterFired(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	c := make(chan time.Time)
	getter := func() chan time.Time {
		return c
	}
	timer := clock.AfterFunc(1*time.Second, func() {
		getter() <- clock.Now()
	})

	clock.BlockUntil(1)
	clock.Advance(1 * time.Second)

	if timer.Stop() {
		t.Error("expected stop to return false")
	}

	assertSent(t, start.Add(1*time.Second), c)

	if timer.Reset(2 * time.Second) {
		t.Error("expected stop to return false")
	}

	c = make(chan time.Time)
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)

	assertSent(t, start.Add(3*time.Second), c)
}

func TestNewTicker(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	ticker := clock.NewTicker(2 * time.Second)

	c := ticker.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(2*time.Second), c)

	c = ticker.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(4*time.Second), c)
}

func TestNewTicker_Double(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	ticker := clock.NewTicker(1 * time.Second)

	c := ticker.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)
	assertSent(t, start.Add(1*time.Second), c)
	// c = ticker.C()
	// assertClockUntil(t, 1, clock)
	// assertSent(t, start.Add(2 * time.Second), c)
}

func TestNewTicker_Stop(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	ticker := clock.NewTicker(1 * time.Second)

	c := ticker.C()
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(1*time.Second), c)

	ticker.Stop()

	c = ticker.C()
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)
}

func TestNewTicker_Stop_NeverCalledC(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	ticker := clock.NewTicker(1 * time.Second)

	clock.Advance(1 * time.Second)

	ticker.Stop()

	c := ticker.C()
	clock.Advance(1 * time.Second)
	assertNotSent(t, c)
}

func TestTick_Positive(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)

	tick := clock.Tick(2 * time.Second)

	c := tick()

	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertClockUntil(t, 1, clock)
	clock.Advance(1 * time.Second)
	assertSent(t, start.Add(2*time.Second), c)

	c = tick()

	assertClockUntil(t, 1, clock)
	clock.Advance(2 * time.Second)
	assertSent(t, start.Add(4*time.Second), c)
}

func TestTick_Zero(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	tick := clock.Tick(0)
	if tick() != nil {
		t.Error("expected tick to return nil")
	}
}

func TestTick_Negative(t *testing.T) {
	start := time.Unix(1, 0)
	clock := clock.NewFakeClockAt(start)
	tick := clock.Tick(-1)
	if tick() != nil {
		t.Error("expected tick to return nil")
	}
}

func assertClockAt(t *testing.T, expected time.Time, clock clock.FakeClock) {
	if actual := clock.Now(); actual != expected {
		t.Errorf("expected %s got %s", expected, actual)
	}
}

const untilTimeout = 100 * time.Millisecond

func assertClockUntil(t *testing.T, n int, clock clock.FakeClock) {
	timer := time.NewTimer(untilTimeout)
	defer timer.Stop()

	select {
	case <-clock.Until(n):
	case <-timer.C:
		t.Errorf("timeout: after %s waiting for %d", untilTimeout, n)
	}
}

const closedTimeout = 100 * time.Millisecond

func assertClosed(t *testing.T, c chan struct{}) {
	timer := time.NewTimer(closedTimeout)
	defer timer.Stop()

	select {
	case _, open := <-c:
		if open {
			t.Error("channel open")
		}
	case <-timer.C:
		t.Errorf("timeout: after %s", closedTimeout)
	}
}

const notClosedTimeout = 100 * time.Millisecond

func assertNotClosed(t *testing.T, c <-chan struct{}) {
	timer := time.NewTimer(notClosedTimeout)
	defer timer.Stop()

	select {
	case <-c:
		t.Error("channel closed unexpectedly")
	case <-timer.C:
	}
}

const sentTimeout = 100 * time.Millisecond

func assertSent(t *testing.T, expected time.Time, c <-chan time.Time) {
	timer := time.NewTimer(sentTimeout)
	defer timer.Stop()

	select {
	case actual := <-c:
		if actual != expected {
			t.Errorf("expected %s got %s", expected, actual)
		}
	case <-timer.C:
		t.Errorf("timeout: after %s", notClosedTimeout)
	}
}

const notSentTimeout = 100 * time.Millisecond

func assertNotSent(t *testing.T, c <-chan time.Time) {
	timer := time.NewTimer(notSentTimeout)
	defer timer.Stop()

	select {
	case <-c:
		t.Error("time sent unexpectedly")
	case <-timer.C:
	}
}
