package wrapper

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"sync"
	"time"
)

// Writer implements the files.Writer interface, that buffers all writes until a Sync or Close, before committing.
type Writer struct {
	mu sync.Mutex

	*Info
	b bytes.Buffer

	flush chan struct{}
	done  chan struct{}
	errch chan error
}

// WriteFunc is a function that is intended to write the given byte slice to some
// underlying source returning any error that should be returned during the
// Sync or Close call which is committing the file.
type WriteFunc func([]byte) error

// NewWriter returns a Writer that is setup to call the given WriteFunc with
// the underlying buffer on every Sync, and Close.
func NewWriter(ctx context.Context, uri *url.URL, f WriteFunc) *Writer {
	wr := &Writer{
		Info:  NewInfo(uri, 0, time.Now()),
		flush: make(chan struct{}),
		done:  make(chan struct{}),
		errch: make(chan error),
	}

	doWrite := func() error {
		wr.mu.Lock()
		defer wr.mu.Unlock()

		// Update ModTime to now.
		wr.Info.SetModTime(time.Now())
		return f(wr.b.Bytes())
	}

	go func() {
		defer func() {
			close(wr.errch)
			close(wr.flush)
		}()

		for {
			select {
			case <-wr.done:
				// For done, we only send a non-nil err,
				// When we close the errch, it will then return nil errors.
				if err := doWrite(); err != nil {
					wr.errch <- err
				}
				return

			case <-wr.flush:
				// For flush, we send even nil errors,
				// Otherwise, the Sync() routine would block forever waiting on an errch.
				wr.errch <- doWrite()

			case <-ctx.Done():
				return
			}
		}
	}()

	return wr
}

// Write performs a thread-safe Write to the underlying buffer.
func (w *Writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.b.Write(b)

	w.Info.SetSize(w.b.Len())

	return n, err
}

func (w *Writer) signalSync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		// cannot flush a closed Writer.
		return io.ErrClosedPipe
	default:
	}

	w.flush <- struct{}{}
	return nil
}

// Sync calls the defined WriteFunc for the Writer with the entire underlying buffer.
func (w *Writer) Sync() error {
	if err := w.signalSync(); err != nil {
		return err
	}

	// We cannot wait here under Lock, because the sync process requires the Lock.
	return <-w.errch
}

func (w *Writer) markDone() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		// already closed
		return os.ErrClosed
	default:
	}

	close(w.done)
	return nil
}

// Close performs a marks the Writer as complete, which also causes a Sync.
func (w *Writer) Close() error {
	if err := w.markDone(); err != nil {
		return err
	}

	// We cannot wait here under Lock, because the sync process requires the Lock.
	return <-w.errch
}
