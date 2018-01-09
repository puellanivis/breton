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
	sync.Mutex

	closed chan struct{}
	ready  chan struct{}

	b bytes.Buffer
}

// New returns a new Pipe that will close if the context.Context given is canceled.
func New(ctx context.Context) *Pipe {
	// initial state is not-closed, and not-ready
	p := &Pipe{
		closed: make(chan struct{}),
		ready:  make(chan struct{}),
	}

	go func() {
		// watch the context, if it closes, then close this pipe.
		select {
		case <-ctx.Done():
			p.Close()
		case <-p.closed:
		}
	}()

	return p
}

// Read blocks until data is available on the buffer, then performs a locked Read on the underlying buffer.
func (p *Pipe) Read(b []byte) (n int, err error) {
	// we want to block here outside of the mutex lock, so we can block waiting for data
	// while at the same time also not holding the mutex.
	<-p.ready

	p.Lock()
	defer p.Unlock()

	if p.b.Len() == 0 {
		// no data on the pipe, can happen when two Readers block on the mutex at the same time.

		select {
		case <-p.closed:
			// we're closed, so return EOF now
			return 0, io.EOF
		default:
		}

		select {
		case <-p.ready:
			// the ready channel is closed, so remake it so we later block
			p.ready = make(chan struct{})

		default:
			// the ready channel is already reopened, so don't open it again
		}

		// report that we intentionally read 0-bytes, with no error.
		return 0, nil
	}

	n, err = p.b.Read(b)

	if p.b.Len() == 0 {
		// no data on the pipe

		select {
		case <-p.closed:
			// we're closed, so don't try and reopen ready
			return n, err
		default:
		}

		select {
		case <-p.ready:
			// the ready channel is closed, so remake it so we later block
			p.ready = make(chan struct{})

		default:
			// the ready channel is already reopened, so don't open it again
		}
	}

	return n, err
}

// Write performs an locked Write to the underlying buffer, and potentially unblocks any Read waiting on data.
func (p *Pipe) Write(b []byte) (n int, err error) {
	p.Lock()
	defer p.Unlock()

	select {
	case <-p.closed:
		// cannot write to a closed pipe
		return 0, io.ErrClosedPipe
	default:
	}

	n, err = p.b.Write(b)

	if err == nil && n > 0 {
		// we wrote data to the buffer, so there is data now, notify any Read that is blocking.

		select {
		case <-p.ready:
			// already marked as ready, so we don't need to mark it ready again
		default:
			close(p.ready)
		}
	}

	return n, err
}

// Close locks the Pipe, then marks it as closed, then unblocks any Readers waiting on data.
func (p *Pipe) Close() error {
	p.Lock()
	defer p.Unlock()

	// ensure we are closed _before_ notifying readers that it is ready
	select {
	case <-p.closed:
	default:
		close(p.closed)
	}

	select {
	case <-p.ready:
	default:
		close(p.ready)
	}

	return nil
}
