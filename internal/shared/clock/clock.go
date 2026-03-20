package clock

import "time"

// Clock is an interface for time-based operations, making time injectable
// for testing.
type Clock interface {
	Now() time.Time
}

// RealClock is the production implementation that delegates to time.Now.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

// New returns a RealClock.
func New() Clock { return RealClock{} }
