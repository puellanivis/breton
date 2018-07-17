// Package atomic (DO NOT USE) is a baseline implementation to replace float64 metrics used internally by prometheus.
package atomic

import (
	"sync/atomic"
)

// Counter is a simple atomic monotonously-increasing counter.
type Counter uint64

// Inc increments the Counter by one (1).
func (c *Counter) Inc() {
	c.Add(1)
}

// Add increments the Counter by a given integer (must be non-negative).
func (c *Counter) Add(delta uint64) {
	atomic.AddUint64((*uint64)(c), delta)
}

// Get returns a momentary value of the Counter.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64((*uint64)(c))
}
