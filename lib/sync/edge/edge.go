// Package edge implements a simple atomic up/down edge/condition-change detector.
//
// When in the Down state, only the first call in any series of calls to Up will return true.
// When in the Up state, only the first call in any series of calls to Down will return true.
//
// So, after Up has been called at least once, only the first Down will return true, and vice-versa.
//
// Usage is pretty basic:
//	// a zero-value Edge starts in state Down
//	var e edge.Edge
//
//	var wg sync.WaitGroup
//	wg.Add(len(values))
//
//	for _, val := range values {
//		val := val		// shadow this value to loop-scope for the goroutine.
//
//		go func(){
//			defer wg.Done()
//
//			if someTestDoneInParallel(val) {
//				if e.Up() {
//					fmt.Println("called at most once")
//				}
//			}
//		}()
//	}
//
//	wg.Wait()
//
//	if e.Up() {
//		fmt.Println("called if and only if e.Up was not called")
//	}
//
// The functionality of Down is symmetric to Up.
//
// This is also useful for marking a data-set as dirty, while only locking
// individual data elements:
//	var dirty edge.Edge
//	t := time.NewTicker(5 * time.Second)
//
//	go func() {
//		for range t.C {
//			if dirty.Down() {
//				lockAndCommitDataSet()
//			}
//		}
//	}()
//
package edge // import "github.com/puellanivis/breton/lib/sync/edge"

import (
	"sync/atomic"
)

// Edge defines a synchronization object that defines two states Up and Down
type Edge struct {
	// hide this value from callers, so that they cannot manipulate it.
	e int32
}

var states = []string{"down", "up"}

// String returns a momentary state of the Edge.
func (e *Edge) String() string {
	v := atomic.LoadInt32(&e.e)
	return states[v]
}

// Up will ensure that the Edge is in state Up, and returns true only if state changed.
func (e *Edge) Up() bool {
	return atomic.CompareAndSwapInt32(&e.e, 0, 1)
}

// Down will ensure that the Edge is in state Down, and returns true only if state changed.
func (e *Edge) Down() bool {
	return atomic.CompareAndSwapInt32(&e.e, 1, 0)
}
