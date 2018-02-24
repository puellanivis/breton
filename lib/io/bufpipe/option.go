package bufpipe

import (
)

type Option func(*Pipe) Option

// WithAutoFlush sets the internal buffer size which shall trigger an Automatic Flush.
// If this size is negative, then no automatic flushing will ever trigger.
// Otherwise, after any successful Write, if the buffer size is greater than this value,
// a Flush will be triggered automatically before returning to the calling program.
func WithAutoFlush(size int) Option {
	return func(p *Pipe) Option {
		p.mu.Lock()
		defer p.mu.Unlock()

		save := p.autoFlush

		p.autoFlush = size

		return WithAutoFlush(save)
	}
}

// WithNoAutoFlush is a more readable version of WithAutoFlush(-1).
func WithNoAutoFlush() Option {
	return WithAutoFlush(-1)
}
