package bufpipe

import (
)

type Option func(*Pipe) Option

func WithAutoFlush(state bool) Option {
	return func(p *Pipe) Option {
		p.mu.Lock()
		defer p.mu.Unlock()

		save := !p.noAutoFlush

		p.noAutoFlush = !state

		return WithAutoFlush(save)
	}
}
