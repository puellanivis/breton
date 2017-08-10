// Package edge implements a simple atomic up/down edge/condition-change detector.
package edge

import (
	"sync/atomic"
)

// Edge defines a synchronization object that deifnes two states Up and Down
type Edge int32

// String returns a momentary state of the Edge.
func (e *Edge) String() string {
	v := atomic.LoadInt64((*int32)(e))
	if v == 0 {
		return "down"
	}

	return "up"
}

// Up will ensure that the Edge is in state Up, and returns true only if state changed.
func (e *Edge) Up() bool {
	return atomic.CompareAndSwapInt32((*int32)(e), 0, 1)
}

// Down will ensure that the Edge is in state Down, and returns true only if state changed.
func (e *Edge) Down() bool {
	return atomic.CompareAndSwapInt32((*int32)(e), 1, 0)
}
