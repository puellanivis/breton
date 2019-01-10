package mapreduce

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type engine struct {
	MapReduce
}

// threadCount returns the valid threadCount value to use based on configuration.
// It guards against invalid values.
func (e *engine) threadCount() int {
	n := e.conf.threadCount

	if n < 1 {
		n = DefaultThreadCount

		if n < 1 {
			// Even if the package-level Default was set to less than one,
			// we need to ensure it is at least one.
			n = 1
		}

		e.conf.threadCount = n
	}

	return n
}

func quickError(err error) <-chan error {
	errch := make(chan error, 1)

	if err != nil {
		errch <- err
	}

	close(errch)
	return errch
}

func (e *engine) run(ctx context.Context, rng Range) <-chan error {
	width := rng.Width()
	if width < 1 {
		return quickError(errors.New("bad range"))
	}

	threads := e.threadCount()

	mappers := e.conf.mapperCount
	if mappers < 1 {
		mappers = threads
	}

	stripe := width / mappers
	extraWork := width % mappers // How many mappers need one more element in order to cover the whole width.

	switch {
	case e.conf.stripeSize > 0:
		maxSize := e.conf.stripeSize

		// We need to calculate the stripe size for an extra-work mapper, if there are extra-work mappers.
		maxWorkSize := stripe
		if extraWork > 0 {
			maxWorkSize++
		}

		if maxWorkSize > maxSize {
			// We only recalculate mapper count if the stripe size is greater than the max stripe size.
			stripe = maxSize
			extraWork = 0

			// Here, the math is simple, but the code is complex.
			//
			// Our mapper count is ⌈width ÷ stripe⌉,
			// but integer math on computers gives ⌊width ÷ stripe⌋.
			mappers = width / stripe

			if width%stripe > 0 {
				// So, if the work does not split up exactly, so we need another mapper.
				mappers++

				// And now, we may as well just recalculate the whole coverage anew… just to be sure.
				stripe = width / mappers
				extraWork = width % mappers
			}
		}

	case e.conf.stripeSize < 0:
		minSize := -e.conf.stripeSize

		// strip is already the smallest work size.

		if stripe < minSize {
			// We only recalculate mapper count if the stripe size is less than the min stripe size.
			stripe = minSize

			// Here, the math is simple, and the code is simple.
			//
			// Our mapper count is ⌊width ÷ stripe⌋.
			mappers = width / stripe

			// Now we just need to recalculate the extra coverage.
			extraWork = width % mappers
		}
	}

	var reducerMutex sync.Mutex
	pool := newThreadPool(threads)
	chain := newExecChain(e.conf.ordered)

	var wg sync.WaitGroup
	wg.Add(mappers)
	errch := make(chan error, mappers)

	go func() {
		wg.Wait()
		close(errch)
	}()

	last := rng.Start
	for i := 0; i < mappers; i++ {
		start := last
		end := start + stripe

		if i < extraWork {
			end++
		}

		if end > rng.End {
			end = rng.End
		}
		last = end

		ready, next := chain.next()

		go func() {
			defer func() {
				wg.Done()
				close(next)
			}()

			rng := Range{
				Start: start,
				End:   end,
			}

			if err := pool.wait(ctx); err != nil {
				errch <- err
				return
			}

			out, err := e.m.Map(ctx, rng)
			if err != nil {
				errch <- err
				return
			}

			if err := pool.done(ctx); err != nil {
				errch <- err
				return
			}

			if out == nil || e.r == nil {
				return
			}

			select {
			case <-ready:
			case <-ctx.Done():
				errch <- ctx.Err()
				return
			}

			reducerMutex.Lock()
			defer reducerMutex.Unlock()

			// Our context may have expired waiting for mutex, so check again.
			select {
			case <-ctx.Done():
				errch <- ctx.Err()
				return
			default:
			}

			if err := e.r.Reduce(ctx, out); err != nil {
				errch <- err
				return
			}
		}()
	}

	if last != rng.End {
		panic(fmt.Errorf("dropped entries! %d != %d", last, rng.End))
	}

	return errch
}
