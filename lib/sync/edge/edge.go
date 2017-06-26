package edge

import (
	"sync/atomic"
)

type Edge int64

func (e *Edge) String() string {
	v := atomic.LoadInt64((*int64)(e))
	if v == 0 {
		return "down"
	}

	return "up"
}

func (e *Edge) Up() bool {
	return atomic.CompareAndSwapInt64((*int64)(e), 0, 1)
}

func (e *Edge) Down() bool {
	return atomic.CompareAndSwapInt64((*int64)(e), 1, 0)
}
