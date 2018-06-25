package mapreduce

import (
	"context"
	"runtime"
	"testing"
)

type TestMR struct {
	ranges []Range
}

func (mr *TestMR) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	rng := in.(Range)

	for i := rng.Start; i < rng.End; i++ {
		runtime.Gosched()
	}

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

		conf: config{
			ordered:     true,
			threadCount: 1,
		},
	}

	f := func(n int) {
		e.conf.mapperCount = n

		mr.ranges = nil

		errs := false

		for err := range e.run(context.Background(), rng) {
			t.Errorf("%d mappers: %+v", n, err)
			errs = true
		}

		if n > 0 && len(mr.ranges) != n {
			t.Errorf("wrong number of mappers ran, expected %d, but got %d", n, len(mr.ranges))
		}

		if errs {
			t.Log(n, mr.ranges)
		}
	}

	for i := 0; i <= rng.Width(); i++ {
		f(i)
	}
}
