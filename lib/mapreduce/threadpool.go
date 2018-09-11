package mapreduce

import (
	"context"
)

type threadPool struct {
	ch chan struct{}
}

func newThreadPool(count int) *threadPool {
	ch := make(chan struct{}, count)

	for i := 0; i < count; i++ {
		ch <- struct{}{}
	}

	return &threadPool{
		ch: ch,
	}
}

func (p *threadPool) wait(ctx context.Context) error {
	select {
	case <-p.ch:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (p *threadPool) done(ctx context.Context) error {
	select {
	case p.ch <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
