# Clock

This package contains a simple clock type that implements the global methods of the `time` package.

There are two implementations of the clock.
- A real implementation backed by the `time` package (`clock.NewRealClock()`).
- A fake implementation for mocking and testing (`clock.NewFakeClock()`).

## `C()`

The one big different between the `time` package and this `clock` package is that the timer and ticker objects return their channel by an interface method (`timer.C()`) instead of a struct field (`timer.C`).

## `Until(n)`

The fake clock keeps track of how many goroutines are waiting on the clock. This allows tests to start background routines and block until those routines are loaded and waiting on the clock. See the `BlockUntil(n)` and `Until(n)` methods for more details.

A goroutine is considered blocking on a timer or ticker once it's called `C()`. Before that, it's not blocking.

**Note. It's best recommended that the calling process save output of `C()` instead of calling it every time.**

Example Timer Usage:
```go
func loop(clock *clock.Clock, ...) {
	timer := clock.NewTimer(d)
	defer timer.Stop()
	
	c := timer.C()
	for {
		select {
		case <-c:
			...
		}
	}
}
```

Example Ticker Usage:
```go
func loop(clock *clock.Clock, ...) {
    ticker := clock.NewTicker(d)
	defer ticker.Stop()
	
	c := ticker.C()
	for {
		select {
		case <-c:
		    c = ticker.C()
			...
		}
	}
}
```

## Influences

This package was influenced by other clocks available for go.

- https://github.com/facebookarchive/clock
- https://github.com/jonboulle/clockwork
- https://github.com/benbjohnson/clock

This package takes a slightly different approach (see the sections above about `C()` and `Until(n)`). This difference allows non-racing usage in tests not possible with the other clocks.
