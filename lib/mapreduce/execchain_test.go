package mapreduce

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/puellanivis/breton/lib/sync/edge"
)

const count = 42

func TestExecChainOrdered(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(true)

	var accum []int
	var wg sync.WaitGroup

	wg.Add(count)
	for i := 0; i < count; i++ {
		me := i
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			accum = append(accum, me)
		}()
	}

	wg.Wait()

	if got := len(accum); got != count {
		t.Fatalf("unexpected accumulator count: got %d, expected %d", got, count)
	}

	for i, n := range accum {
		if expected := i; n != expected {
			t.Errorf("got %d, expected %d", n, expected)
		}
	}
}

func TestExecChainOrdered_OddsFailFast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(true)

	errch := make(chan error, count)
	wait := make(chan struct{})

	var accum []int
	var wg sync.WaitGroup
	var e  edge.Edge

	wg.Add(count)
	for i := 0; i < count; i++ {
		me := i
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if me & 1 == 1 {
				// If we are odd, return early
				// the evens should still complete in order.
				return
			}

			if err := link.wait(ctx); err != nil {
				return
			}

			if !e.Up() {
				errch <- fmt.Errorf("duplicate execution: edge already up: goroutine %d", me)
			}

			<-wait
			accum = append(accum, me)

			if !e.Down() {
				errch <- fmt.Errorf("duplicate execution: edge already down: goroutine %d", me)
			}
		}()
	}

	go func() {
		time.Sleep(1 * time.Millisecond)
		close(wait)

		wg.Wait()
		close(errch)
	}()

	for err := range errch {
		t.Error(err)
	}

	if got, expected := len(accum), count/2; got != expected {
		t.Fatalf("unexpected accumulator count: got %d, expected %d", got, expected)
	}

	for i, n := range accum {
		if expected := i*2; n != expected {
			t.Errorf("got %d, expected %d", n, expected)
		}
	}
}

func TestExecChainOrderedStall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(true)

	wait := make(chan struct{})

	var wg sync.WaitGroup
	var ran int

	wg.Add(count)
	for i := 0; i < count; i++ {
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			ran++

			close(wait)
			<-ctx.Done()
		}()
	}

	<-wait
	cancel()

	wg.Wait()

	if ran != 1 {
		t.Error("only one thread should run, but", ran, "did")
	}
}

func TestExecChainUnordered(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(false)

	var accum []int
	var wg sync.WaitGroup

	wg.Add(count)
	for i := 0; i < count; i++ {
		me := i
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			accum = append(accum, me)
		}()
	}

	wg.Wait()

	sort.Ints(accum)

	if got := len(accum); got != count {
		t.Fatalf("unexpected accumulator count: got %d, expected %d", got, count)
	}

	for i, n := range accum {
		if expected := i; n != expected {
			t.Errorf("got %d, expected %d", n, expected)
		}
	}
}

func TestExecChainUnordered_OddsFailFast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(false)

	errch := make(chan error, count)
	wait := make(chan struct{})

	var accum []int
	var wg sync.WaitGroup
	var e edge.Edge

	wg.Add(count)
	for i := 0; i < count; i++ {
		me := i
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if me & 1 == 1 {
				// If we are odd, return early
				// the evens should still complete.
				return
			}

			if err := link.wait(ctx); err != nil {
				return
			}

			if !e.Up() {
				errch <- fmt.Errorf("duplicate execution: edge already up: goroutine %d", me)
			}

			<-wait
			accum = append(accum, me)

			if !e.Down() {
				errch <- fmt.Errorf("duplicate execution: edge already down: goroutine %d", me)
			}
		}()
	}

	go func() {
		time.Sleep(1 * time.Millisecond)
		close(wait)

		wg.Wait()
		close(errch)
	}()

	for err := range errch {
		t.Error(err)
	}

	sort.Ints(accum)

	if got, expected := len(accum), count/2; got != expected {
		t.Fatalf("unexpected accumulator count: got %d, expected %d", got, expected)
	}

	for i, n := range accum {
		if expected := i*2; n != expected {
			t.Errorf("got %d, expected %d", n, expected)
		}
	}
}

func TestExecChainUnorderedStall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(false)

	wait := make(chan struct{})

	var wg sync.WaitGroup
	var ran int

	wg.Add(count)
	for i := 0; i < count; i++ {
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			ran++

			close(wait)
			<-ctx.Done()
		}()
	}

	<-wait
	cancel()

	wg.Wait()

	if ran != 1 {
		t.Error("only one thread should run, but", ran, "did")
	}
}
