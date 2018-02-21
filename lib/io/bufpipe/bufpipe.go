// Package bufpipe implements a buffered in-memory pipe where reads will block until data is available.
package bufpipe

import (
	"bytes"
	"context"
	"io"
	"sync"
)

// Pipe defines an io.Reader and io.Writer where data given to Write will be buffered until a corresponding Read.
type Pipe struct {
	once sync.Once
	mu sync.Mutex

	closed chan struct{}
	ready  chan struct{}

	noAutoFlush bool
	b bytes.Buffer
}

func (p *Pipe) init() {
	p.closed = make(chan struct{})
	p.ready = make(chan struct{})
}

// New returns a new Pipe with the given Options, and will be closed if the given context.Context is canceled.
// If a nil context is given, then no context-dependent closing will be done.
func New(ctx context.Context, opts ...Option) *Pipe {
	p := new(Pipe)
	p.once.Do(p.init)

	for _, opt := range opts {
		_ = opt(p)
	}

	if ctx != nil {
		p.CloseOnContext(ctx)
	}

	return p
}

// CloseOnContext will close the Pipe if the given context.Context is canceled.
// A single Pipe can be setup to close on multiple different and even independent contexts.
// With great power, comes great responsibility. Use wisely.
func (p *Pipe) CloseOnContext(ctx context.Context) {
	go func() {
		// Watch the context, if it closes, then Close this pipe.
		select {
		case <-ctx.Done():
			p.Close()
		case <-p.closed:
		}
	}()
}

func (p *Pipe) doEmptyBuffer() error {
	select {
	case <-p.closed:
		return io.EOF
	default:
	}

	// We have to check these separately,
	// because by Go standards, a select picks between ready channels randomly.
	// And we need to ensure these are tested sequentially.

	select {
	case <-p.ready:
		// The ready channel is closed, so remake it so future Readers will block.
		p.ready = make(chan struct{})

	default:
		// The ready channel is already reopened, so don't open it again.
	}

	return nil
}

// Read blocks until data is available on the buffer, then performs a locked Read on the underlying buffer.
func (p *Pipe) Read(b []byte) (n int, err error) {
	p.once.Do(p.init)

	// We want to block here outside of the mutex lock,
	// so we can block waiting for data while at the same time also not holding the mutex.
	// Otherwise, we could not Write to the Pipe!
	<-p.ready

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.b.Len() == 0 {
		return 0, p.doEmptyBuffer()
	}

	n, err = p.b.Read(b)

	if p.b.Len() == 0 {
		_ = p.doEmptyBuffer()

		return n, nil
	}

	return n, err
}

func (p *Pipe) prewrite() error {
	select {
	case <-p.closed:
		// One cannot write/flush a closed pipe.
		return io.ErrClosedPipe
	default:
	}

	return nil
}

// Write performs an locked Write to the underlying buffer.
// If AutoFlush is enabled (the default), it will also unblock any blocked Readers.
func (p *Pipe) Write(b []byte) (n int, err error) {
	p.once.Do(p.init)
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.prewrite(); err != nil {
		return 0, err
	}

	n, err = p.b.Write(b)

	if !p.noAutoFlush && n > 0 {
		p.flush()
	}

	return n, err
}

func (p *Pipe) flush() {
	select {
	case <-p.ready:
		// Pipe is already marked as ready, so we don't need to mark it ready again.
	default:
		close(p.ready)
	}
}

// Flush will unblock any blocked Readers.
// This could cause them to read zero bytes of data.
func (p *Pipe) Flush() error {
	p.once.Do(p.init)
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.prewrite(); err != nil {
		return err
	}

	p.flush()

	return nil
}

func (p *Pipe) close() {
	select {
	case <-p.closed:
	default:
		close(p.closed)
	}
}

// Close will close the Pipe, and unblock any blocked Readers.
func (p *Pipe) Close() error {
	p.once.Do(p.init)
	p.mu.Lock()
	defer p.mu.Unlock()

	// Order is vitally important here: close then flush.
	p.close()
	p.flush()

	return nil
}
