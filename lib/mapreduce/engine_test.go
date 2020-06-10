package mapreduce

import (
	"context"
	"fmt"
	"runtime"
	"sync"
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

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMR{}
	e := &engine{
		MapReduce: MapReduce{
			m: mr,
			r: mr,
		},
	}

	WithOrdering(true)(&e.MapReduce)
	WithThreadCount(1)(&e.MapReduce)

	for n := 0; n <= rng.Width(); n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			WithMapperCount(n)(&e.MapReduce)

			mr.reset()

			for err := range e.run(ctx, rng) {
				t.Errorf("%d mappers: %+v", n, err)
			}

			t.Log(n, len(mr.widths), mr.widths)

			if n > 0 && len(mr.ranges) != n {
				t.Log(mr.ranges)

				t.Errorf("wrong number of mappers ran: got %d, but expected %d", len(mr.ranges), n)
			}
		})
	}
}

func TestEngineMaxSliceSize(t *testing.T) {
	ctx := context.Background()

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMR{}

	e := &engine{
		MapReduce: MapReduce{
			m: mr,
			r: mr,
		},
	}

	testWidth := 7

	WithOrdering(true)(&e.MapReduce)
	WithThreadCount(1)(&e.MapReduce)
	WithMaxStripeSize(testWidth)(&e.MapReduce)

	for n := 0; n <= rng.Width(); n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			WithMapperCount(n)(&e.MapReduce)

			mr.reset()

			for err := range e.run(ctx, rng) {
				t.Errorf("unexpected error: %+v", err)
			}

			t.Log(n, len(mr.widths), mr.widths)

			for _, width := range mr.widths {
				if width > testWidth {
					t.Log(mr.ranges)
					t.Fatalf("range was greater than maximum: got %d, but expected not greater than %d", width, testWidth)
				}
			}
		})
	}
}

func TestEngineMinSliceSize(t *testing.T) {
	ctx := context.Background()

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	mr := &TestMR{}

	e := &engine{
		MapReduce: MapReduce{
			m: mr,
			r: mr,
		},
	}

	testWidth := 7

	WithOrdering(true)(&e.MapReduce)
	WithThreadCount(1)(&e.MapReduce)
	WithMinStripeSize(testWidth)(&e.MapReduce)

	for n := 0; n <= rng.Width(); n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			WithMapperCount(n)(&e.MapReduce)

			mr.reset()

			for err := range e.run(ctx, rng) {
				t.Errorf("unexpected error: %+v", err)
			}

			t.Log(n, len(mr.widths), mr.widths)

			for _, width := range mr.widths {
				if width < testWidth {
					t.Log(mr.ranges)
					t.Errorf("range was less than minimum: got %d, but expected not less than %d", width, testWidth)
					break
				}
			}
		})
	}
}

type TestMRBlock struct {
	wg *sync.WaitGroup

	reduces      int
	duringReduce bool
}

func (mr *TestMRBlock) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	out = struct{}{}

	if !mr.duringReduce {
		// Mark WaitGroup done here, so we can fast-cancel the context.
		mr.wg.Done()

		<-ctx.Done()
		return out, ctx.Err()
	}

	return out, nil
}

func (mr *TestMRBlock) Reduce(ctx context.Context, in interface{}) error {
	mr.reduces++

	if mr.duringReduce {
		// Mark WaitGroup done here, so we can fast-cancel the context.
		mr.wg.Done()

		<-ctx.Done()
		return ctx.Err()
	}

	return nil
}

func TestEngineStallInMapper(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	var wg sync.WaitGroup
	mr := &TestMRBlock{
		wg: &wg,
	}

	e := &engine{
		MapReduce: MapReduce{
			m: mr,
			r: mr,
		},
	}

	n := 4

	WithThreadCount(n)(&e.MapReduce)
	WithMapperCount(n)(&e.MapReduce)
	wg.Add(n)

	go func() {
		wg.Wait()
		cancel()
	}()

	var errCount int
	for err := range e.run(ctx, rng) {
		t.Logf("%+v", err)
		errCount++
	}

	if errCount != n {
		t.Errorf("got %d errors, but expected %d", errCount, n)
	}

	expectedReduces := 0
	if mr.reduces != expectedReduces {
		t.Errorf("wrong number of reducer ran: got %d, but expected %d", mr.reduces, expectedReduces)
	}
}

func TestEngineStallInReducer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	rng := Range{
		Start: 42,
		End:   42 + 53, // Give this a width of 53, a prime number.
	}

	var wg sync.WaitGroup
	mr := &TestMRBlock{
		wg: &wg,

		duringReduce: true,
	}

	e := &engine{
		MapReduce: MapReduce{
			m: mr,
			r: mr,
		},
	}

	n := 4

	WithThreadCount(n)(&e.MapReduce)
	WithMapperCount(n)(&e.MapReduce)
	wg.Add(1)

	go func() {
		wg.Wait()
		cancel()
	}()

	var errCount int
	for err := range e.run(ctx, rng) {
		t.Logf("%+v", err)
		errCount++
	}

	if errCount != n {
		t.Errorf("got %d errors, but expected %d", errCount, n)
	}

	expectedReduces := 1
	if mr.reduces != expectedReduces {
		t.Errorf("wrong number of reducer ran: got %d, but expected %d", mr.reduces, expectedReduces)
	}
}
