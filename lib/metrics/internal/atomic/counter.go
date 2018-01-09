// Package atomic (DO NOT USE) is a baseline implementation to replace float64 metrics used internally by prometheus.
package atomic

import (
	"sync/atomic"
)

type Counter uint64

func (c *Counter) Inc() {
	c.Add(1)
}

func (c *Counter) Add(delta uint64) {
	atomic.AddUint64((*uint64)(c), delta)
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64((*uint64)(c))
}
