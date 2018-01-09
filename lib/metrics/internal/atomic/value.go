package atomic

import (
	"sync/atomic"
)

type Value int64

func (v *Value) Inc() {
	v.Add(1)
}

func (v *Value) Dec() {
	v.Add(-1)
}

func (v *Value) Add(delta int64) {
	atomic.AddInt64((*int64)(v), delta)
}

func (v *Value) Sub(delta int64) {
	v.Add(-delta)
}

func (v *Value) Set(value int64) {
	atomic.StoreInt64((*int64)(v), value)
}

func (v *Value) Get() int64 {
	return atomic.LoadInt64((*int64)(v))
}
