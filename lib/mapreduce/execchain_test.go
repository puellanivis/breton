package mapreduce

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestExecChainOrdered(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(true)

	var accum []int

	var wg sync.WaitGroup

	wg.Add(42)
	for i := 0; i < 42; i++ {
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

	for i, n := range accum {
		if i != n {
			t.Error(i, "!=", n)
		}
	}
}

func TestExecChainOrderedStall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(true)

	var wg sync.WaitGroup
	var ran int

	wg.Add(42)
	for i := 0; i < 42; i++ {
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			ran++

			<-ctx.Done()
		}()
	}

	wg.Wait()

	if ran != 1 {
		t.Error("only one thread should run, but", ran, "did")
	}
}

func TestExecChainUnordered(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(false)

	ran := make(map[int]bool)

	var wg sync.WaitGroup

	wg.Add(42)
	for i := 0; i < 42; i++ {
		me := i
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			ran[me] = true
		}()
	}

	wg.Wait()

	var accum []int
	for n := range ran {
		accum = append(accum, n)
	}
	sort.Ints(accum)

	for i, n := range accum {
		if i != n {
			t.Error(i, "!=", n)
		}
	}
}

func TestExecChainUnorderedStall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	chain := newExecChain(false)

	var wg sync.WaitGroup

	var ran int

	wg.Add(42)
	for i := 0; i < 42; i++ {
		link := chain.next()

		go func() {
			defer wg.Done()
			defer link.done()

			if err := link.wait(ctx); err != nil {
				return
			}

			ran++

			<-ctx.Done()
		}()
	}

	wg.Wait()

	if ran != 1 {
		t.Error("only one thread should run, but", ran, "did")
	}
}
