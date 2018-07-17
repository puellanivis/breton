package bufpipe

// Option defines a function that will apply a specific value or feature to a given Pipe.
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

// WithMaxOutstanding sets the maximal size of the internal buffer before a Read is forced.
// If this value is set to be greater than zero, if any Write would cause the internal buffer to exceed this value,
// then that Write will block until a Read is performed that empties the internal buffer.
func WithMaxOutstanding(size int) Option {
	return func(p *Pipe) Option {
		p.mu.Lock()
		defer p.mu.Unlock()

		save := p.maxOutstanding

		p.maxOutstanding = size

		return WithMaxOutstanding(save)
	}
}
