package mapreduce

import (
	"context"
)

type semaphore struct {
	ch chan struct{}
}

func newSemaphore(count int) *semaphore {
	ch := make(chan struct{}, count)

	for len(ch) < cap(ch) {
		ch <- struct{}{}
	}

	return &semaphore{
		ch: ch,
	}
}

func (s *semaphore) get() *localSem {
	return &localSem{
		ch: s.ch,
	}
}

// localSem will only unlock, if it obtained its lock.
type localSem struct{
	ch chan struct{}

	locked bool
}

func (s *localSem) wait(ctx context.Context) error {
	select {
	case <-s.ch:
		s.locked = true

	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (s *localSem) done() {
	if !s.locked {
		return
	}

	select{
	case s.ch <- struct{}{}:
	default:
		panic("too many semaphore unlocks")
	}
}
