package mapreduce

import (
	"context"
	"runtime"
)

var DefaultThreadCount = runtime.NumCPU()

// Map defines a function to be called on each Stripe of data in a given MapReduce.
// This is not critical section code and each Map will be run in a separate goroutine in parallel to the thread count.
type MapFunc func(ctx context.Context, in interface{}) (out interface{}, err error)

func (m MapFunc) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	return m(ctx, in)
}

// Reduce defines a function that recieves the output of a single Map.
// This is critical section code, and only one Reduce goroutine will ever be running at a time.
type ReduceFunc func(ctx context.Context, in interface{}) error

func (r ReduceFunc) Reduce(ctx context.Context, in interface{}) error {
	return r(ctx, in)
}

type MapReducer interface {
	Map(ctx context.Context, in interface{}) (out interface{}, err error)
	Reduce(ctx context.Context, in interface{}) error
}
