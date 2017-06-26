package edge

import (
	"sync/atomic"
)

// Edge defines a synchronization object that deifnes two states Up and Down
type Edge int64

// String returns a momentary state of the Edge.
func (e *Edge) String() string {
	v := atomic.LoadInt64((*int64)(e))
	if v == 0 {
		return "down"
	}

	return "up"
}

// Up will ensure that the Edge is in state Up, and returns true only if state changed.
func (e *Edge) Up() bool {
	return atomic.CompareAndSwapInt64((*int64)(e), 0, 1)
}

// Down will ensure that the Edge is in state Down, and returns true only if state changed.
func (e *Edge) Down() bool {
	return atomic.CompareAndSwapInt64((*int64)(e), 1, 0)
}
