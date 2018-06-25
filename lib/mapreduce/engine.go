package mapreduce

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type engine struct {
	m    Mapper
	r    Reducer
	conf config
}

func (e *engine) run(ctx context.Context, rng Range) <-chan error {
	width := rng.Width()

	if width <= 0 {
		errch := make(chan error, 1)

		if width < 0 {
			errch <- errors.New("bad range")
		}

		close(errch)
		return errch
	}

	errch := make(chan error)

	threads := e.conf.threadCount
	if threads <= 0 {
		threads = DefaultThreadCount

		if threads < 1 {
			// If the default was set to less than one, we want to ensure it is at least one.
			threads = 1
		}
	}
	pool := make(chan struct{}, threads)

	mappers := e.conf.mapperCount
	if mappers <= 0 {
		mappers = threads
	}

	stripe := width / mappers

	// extraWork is how many mappers need one more element in order to cover the whole width.
	extraWork := width % mappers

	if e.conf.stripeSize > 0 && stripe > e.conf.stripeSize {
		// If the number of mappers we have already makes a stripe size of less than the configured value,
		// then we do not need to recalculate the mapper count.
		stripe = e.conf.stripeSize
		mappers = width / stripe
		if width%stripe > 0 {
			mappers++
		}

		extraWork = 0 // do not add any extra work for any mappers.
	}

	var mu sync.Mutex
	chain := make(chan struct{})
	close(chain)

	unordered := make(chan struct{})
	if !e.conf.ordered {
		close(unordered)
	}

	var wg sync.WaitGroup
	wg.Add(mappers)

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

		ready := chain
		chain = make(chan struct{})
		next := chain

		go func() {
			defer func() {
				wg.Done()
				close(next)
			}()

			rng := Range{
				Start: start,
				End:   end,
			}

			<-pool

			out, err := e.m.Map(ctx, rng)
			if err != nil {
				errch <- err
			}

			pool <- struct{}{}

			if out == nil || e.r == nil {
				return
			}

			select {
			case <-unordered:
			case <-ready:
			case <-ctx.Done():
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if err := e.r.Reduce(ctx, out); err != nil {
				errch <- err
			}
		}()
	}

	if last != rng.End {
		panic(fmt.Errorf("dropped entries! %d != %d", last, rng.End))
	}

	go func() {
		defer close(errch)

		for i := 0; i < threads; i++ {
			pool <- struct{}{}
		}

		wg.Wait()
	}()

	return errch
}
