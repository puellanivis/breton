package wrapper

import (
	"context"
	"bytes"
	"net/url"
	"sync"
	"time"
)

type Writer struct {
	sync.Mutex

	*Info
	b bytes.Buffer

	flush chan bool
	done chan struct{}
	errch chan error
}

type WriteFn func([]byte) error

func NewWriter(ctx context.Context, uri *url.URL, f WriteFn) *Writer {
	wr := &Writer{
		Info: NewInfo(uri, 0, time.Now()),
		done: make(chan struct{}),
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

func (w *Writer) Write(b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()

	return w.b.Write(b)
}

func (w *Writer) Sync() error {
	<-w.flush
	return <-w.errch
}

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
