package mapreduce

import (
	"context"
	"testing"
	"runtime"
)

type TestMR struct {
	ranges []Range
}

func (mr *TestMR) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	rng := in.(Range)

	runtime.Gosched()

	return rng, nil
}

func (mr *TestMR) Reduce(ctx context.Context, in interface{}) error {
	rng := in.(Range)

	mr.ranges = append(mr.ranges, rng)

	return nil
}

func TestEngine(t *testing.T) {
	DefaultThreadCount = -1

	rng := Range{
		Start: 42,
		End:   69,
	}

	mr := &TestMR{}

	e := &engine{
		m: mr,
		r: mr,
	}

	WithOrdering(true)(&e.conf)
	WithThreadCount(1)(&e.conf)

	f := func(n int) {
		WithMapperCount(n)(&e.conf)

		mr.ranges = nil

		for err := range e.run(context.Background(), rng) {
			t.Error(err)
		}

		t.Log(n, mr.ranges)
	}

	for i := 0; i <= rng.Width(); i++ {
		f(i)
	}
}
