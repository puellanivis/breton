package atomic

import (
	"sync/atomic"
)

// Value is a atomic fully variable integer Value.
type Value int64

// Inc increments the Value by one (1).
func (v *Value) Inc() {
	v.Add(1)
}

// Dec decrements the Value by one (1).
func (v *Value) Dec() {
	v.Add(-1)
}

// Add increments the Value by a given integer (may be negative).
func (v *Value) Add(delta int64) {
	atomic.AddInt64((*int64)(v), delta)
}

// Sub decrements the Value by a given integer (may be negative).
func (v *Value) Sub(delta int64) {
	v.Add(-delta)
}

// Set assigns a value into the Value, overriding any previous value.
func (v *Value) Set(value int64) {
	atomic.StoreInt64((*int64)(v), value)
}

// Get returns a momentary value of the Value.
func (v *Value) Get() int64 {
	return atomic.LoadInt64((*int64)(v))
}
