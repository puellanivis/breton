package wrapper

import (
	"bytes"
	"context"
	"net/url"
	"sync"
	"time"
)

// Writer implements the files.Writer interface, that buffers all writes until a Sync or Close, before committing.
type Writer struct {
	sync.Mutex

	*Info
	b bytes.Buffer

	flush chan bool
	done  chan struct{}
	errch chan error
}

// WriteFn is a function that is intended to write the given byte slice to some
// underlying source returning any error that should be returned during the
// Sync or Close call which is committing the file.
type WriteFn func([]byte) error

// NewWriter returns a Writer that is setup to call the given WriteFn with
// the underlying buffer on every Sync, and Close.
func NewWriter(ctx context.Context, uri *url.URL, f WriteFn) *Writer {
	wr := &Writer{
		Info:  NewInfo(uri, 0, time.Now()),
		done:  make(chan struct{}),
		errch: make(chan error),
		flush: make(chan bool),
	}

	go func() {
		for {
			select {
			case <-wr.flush:
				wr.errch <- f(wr.b.Bytes())

			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		defer close(wr.errch)

		select {
		case <-wr.done:
		case <-ctx.Done():
			return
		}

		close(wr.flush)
	}()

	return wr
}

// Write performs a thread-safe Write to the underlying buffer.
func (w *Writer) Write(b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()

	return w.b.Write(b)
}

// Sync calls the defined WriteFn for the Writer with the entire underlying buffer.
func (w *Writer) Sync() error {
	w.Lock()
	defer w.Unlock()

	w.flush <- true
	return <-w.errch
}

// Close performs a marks the Writer as complete, which also causes a Sync.
func (w *Writer) Close() error {
	w.Lock()
	defer w.Unlock()

	select {
	case <-w.done:
		return nil
	default:
	}

	close(w.done)
	return <-w.errch
}
