package wrapper

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"sync"
	"time"
)

// Writer implements the files.Writer interface, that buffers all writes until a Sync or Close, before committing.
type Writer struct {
	mu sync.Mutex

	*Info
	b  *bytes.Buffer
	do func([]byte) error // must be called with lock.
}

// WriteFunc is a function that is intended to write the given byte slice to some
// underlying source returning any error that should be returned during the
// Sync or Close call which is committing the file.
type WriteFunc func([]byte) error

// NewWriter returns a Writer that is setup to call the given WriteFunc with
// the underlying buffer on every Sync, and Close.
func NewWriter(ctx context.Context, uri *url.URL, f WriteFunc) *Writer {
	info := NewInfo(uri, 0, time.Now())

	wr := &Writer{
		Info: info,
		b:    new(bytes.Buffer),
		do: func(b []byte) error {
			// Update ModTime to now.
			info.SetModTime(time.Now())
			return f(b)
		},
	}

	return wr
}

// Write performs a thread-safe Write to the underlying buffer.
func (w *Writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.b == nil {
		// cannot write to closed Writer
		return 0, os.ErrClosed
	}

	n, err = w.b.Write(b)

	w.Info.SetSize(w.b.Len())

	return n, err
}

// Sync calls the defined WriteFunc for the Writer with the entire underlying buffer.
func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.b == nil {
		// cannot sync a closed Writer
		return os.ErrClosed
	}

	return w.do(w.b.Bytes())
}

// Close performs a marks the Writer as complete, which also causes a Sync.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.b == nil {
		// cannot sync a closed Writer
		return os.ErrClosed
	}
	data := w.b.Bytes()
	w.b = nil

	return w.do(data)
}
