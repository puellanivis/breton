package atomic

import (
	"math"
	"sync/atomic"
)

// Float is a replica of the prometheus spinlock counter based on float64.
type Float uint64

// Inc increments the Float by one (1).
func (f *Float) Inc() {
	f.Add(1)
}

// Dec decrements the Float by one (1).
func (f *Float) Dec() {
	f.Add(-1)
}

// Add increments the Float by a given floating-point value (may be negative).
func (f *Float) Add(delta float64) {
	var o, n uint64
	for {
		o = atomic.LoadUint64((*uint64)(f))
		n = math.Float64bits(math.Float64frombits(o) + delta)
		if atomic.CompareAndSwapUint64((*uint64)(f), o, n) {
			return
		}
	}
}

// Sub decrements the Float by a given floating-point value (may be negative).
func (f *Float) Sub(delta float64) {
	f.Add(-delta)
}

// Set assigns a value into the Float, overriding any previous value.
func (f *Float) Set(value float64) {
	atomic.StoreUint64((*uint64)(f), math.Float64bits(value))
}

// Get returns a momentary value of the Float.
func (f *Float) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}
