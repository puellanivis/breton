package mapreduce

import (
	"context"
	"testing"
	//"math/rand"
	//"time"
)

type TestMR struct{
	ranges []Range
}

func (mr *TestMR) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	rng := in.(Range)

	//<-time.After(time.Duration(rand.Intn(int(1 * time.Second))))

	return rng, nil
}

func (mr *TestMR) Reduce(ctx context.Context, in interface{}) error {
	rng := in.(Range)

	mr.ranges = append(mr.ranges, rng)

	return nil
}

func TestEngine(t *testing.T) {
	rng := Range{
		Start: 42,
		End: 69,
	}

	DefaultThreadCount = 16

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
