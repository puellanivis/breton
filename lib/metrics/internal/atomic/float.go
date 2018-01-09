package atomic

import (
	"math"
	"sync/atomic"
)

type Float uint64

func (f *Float) Inc() {
	f.Add(1)
}

func (f *Float) Dec() {
	f.Add(-1)
}

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

func (f *Float) Sub(delta float64) {
	f.Add(-delta)
}

func (f *Float) Set(value float64) {
	atomic.StoreUint64((*uint64)(f), math.Float64bits(value))
}

func (f *Float) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}
