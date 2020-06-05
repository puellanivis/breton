package mapreduce

import (
	"context"
)

type waiter interface{
	wait(ctx context.Context) error
	done()
}

type link struct{
	prev <-chan struct{}
	next chan struct{}
}

func (l *link) wait(ctx context.Context) error {
	select {
	case <-l.prev:
	case <-ctx.Done():
		return ctx.Err()
	}

	// If both select cases were ready,
	// Go will complete one at random,
	// so, we have to check for possible context completion again.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (l *link) done() {
	close(l.next)
}

type chain interface{
	next() waiter
}

type orderedChain struct {
	ch chan struct{}
}

func (c *orderedChain) next() waiter {
	prev, next := c.ch, make(chan struct{})
	c.ch = next

	return &link{
		prev: prev,
		next: next,
	}
}

type unorderedChain struct{
	sem *semaphore
}

func (c *unorderedChain) next() waiter {
	return c.sem.get()
}

func newExecChain(ordered bool) chain {
	if ordered {
		// setup for an ordered chain of chan struct{}s
		ch := make(chan struct{})
		close(ch)

		return &orderedChain{
			ch: ch,
		}
	}

	return &unorderedChain{
		sem: newSemaphore(1),
	}
}
