package mapreduce

import (
	"context"
	"runtime"
	"testing"
	"time"
)

type TestMR struct {
	ranges []Range
	widths []int
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
	mr.widths = append(mr.widths, rng.Width())

	return nil
}

func (mr *TestMR) reset() {
	mr.ranges = nil
	mr.widths = nil
}

func TestEngine(t *testing.T) {
	ctx := context.Background()
	DefaultThreadCount = -1

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
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

		mr.reset()

		for err := range e.run(ctx, rng) {
			t.Errorf("%d mappers: %+v", n, err)
		}

		t.Log(n, len(mr.widths), mr.widths)

		if n > 0 && len(mr.ranges) != n {
			t.Log(mr.ranges)

			t.Errorf("wrong number of mappers ran, expected %d, but got %d", n, len(mr.ranges))
		}
	}

	for i := 0; i <= rng.Width(); i++ {
		f(i)
	}
}

func TestEngineMaxSliceSize(t *testing.T) {
	ctx := context.Background()
	DefaultThreadCount = -1

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMR{}

	e := &engine{
		m: mr,
		r: mr,
	}

	testWidth := 7

	WithOrdering(true)(&e.conf)
	WithThreadCount(1)(&e.conf)
	WithMaxStripeSize(testWidth)(&e.conf)

	f := func(n int) {
		WithMapperCount(n)(&e.conf)

		mr.reset()

		for err := range e.run(ctx, rng) {
			t.Errorf("%d mappers: %+v", n, err)
		}

		t.Log(n, len(mr.widths), mr.widths)

		for _, width := range mr.widths {
			if width > testWidth {
				t.Log(mr.ranges)
				t.Errorf("range was greater than maximum, expected %d, but got %d", width, testWidth)
				break
			}
		}
	}

	for i := 0; i <= rng.Width(); i++ {
		f(i)
	}
}

func TestEngineMinSliceSize(t *testing.T) {
	ctx := context.Background()
	DefaultThreadCount = -1

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMR{}

	e := &engine{
		m: mr,
		r: mr,
	}

	testWidth := 7

	WithOrdering(true)(&e.conf)
	WithThreadCount(1)(&e.conf)
	WithMinStripeSize(testWidth)(&e.conf)

	f := func(n int) {
		WithMapperCount(n)(&e.conf)

		mr.reset()

		for err := range e.run(ctx, rng) {
			t.Errorf("%d mappers: %+v", n, err)
		}

		t.Log(n, len(mr.widths), mr.widths)

		for _, width := range mr.widths {
			if width < testWidth {
				t.Log(mr.ranges)
				t.Errorf("range was less than minimum, expected %d, but got %d", width, testWidth)
				break
			}
		}
	}

	for i := 0; i <= rng.Width(); i++ {
		f(i)
	}
}

type TestMRBlock struct {
	reduces      int
	duringReduce bool
}

func (mr *TestMRBlock) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	out = struct{}{}

	if !mr.duringReduce {
		<-ctx.Done()
		return out, ctx.Err()
	}

	return out, nil
}

func (mr *TestMRBlock) Reduce(ctx context.Context, in interface{}) error {
	mr.reduces++

	if mr.duringReduce {
		<-ctx.Done()
		return ctx.Err()
	}

	return nil
}

func TestEngineStall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	DefaultThreadCount = -1

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMRBlock{}

	e := &engine{
		m: mr,
		r: mr,
	}

	n := 4

	WithThreadCount(n)(&e.conf)
	WithMapperCount(n)(&e.conf)

	var errCount int
	for err := range e.run(ctx, rng) {
		t.Logf("%+v", err)
		errCount++
	}

	if errCount != n {
		t.Errorf("expected %d errors, but got %d", n, errCount)
	}

	expectedReduces := 0
	if mr.reduces != expectedReduces {
		t.Errorf("wrong number of mappers got to reduce phase, expected %d, but got %d", expectedReduces, mr.reduces)
	}

	errCount = 0
	mr.duringReduce = true

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for err := range e.run(ctx, rng) {
		t.Logf("%+v", err)
		errCount++
	}

	if errCount != n {
		t.Errorf("expected %d errors, but got %d", n, errCount)
	}

	expectedReduces = 1
	if mr.reduces != expectedReduces {
		t.Errorf("wrong number of mappers got to reduce phase, expected %d, but got %d", expectedReduces, mr.reduces)
	}

}
