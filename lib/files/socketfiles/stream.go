package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files/wrapper"
)

type streamWriter struct {
	*wrapper.Info

	mu     sync.Mutex
	closed chan struct{}

	sock *socket
}

func (w *streamWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.sock.setBitrate(bitrate, 1)

	// Update filename.
	w.Info.SetNameFromURL(w.sock.uri())

	return prev
}

func (w *streamWriter) Sync() error {
	return nil
}

func (w *streamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	return w.sock.conn.Close()
}

func (w *streamWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.sock.throttle(len(b))

	return w.sock.conn.Write(b)
}

func (w *streamWriter) uri() *url.URL {
	return w.sock.uri()
}

func newStreamWriter(ctx context.Context, sock *socket) *streamWriter {
	w := &streamWriter{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		sock: sock,

		closed: make(chan struct{}),
	}

	go func() {
		select {
		case <-w.closed:
		case <-ctx.Done():
			w.Close()
		}
	}()

	return w
}

type streamReader struct {
	*wrapper.Info

	loading <-chan struct{}

	err  error
	conn net.Conn
}

func (r *streamReader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.conn.Read(b)
}

func (r *streamReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *streamReader) Close() error {
	for range r.loading {
	}

	// Never connected, so just return nil.
	if r.conn == nil {
		return nil
	}

	// Ignore the r.err, as it is a request-scope error, and not relevant to closing.

	return r.conn.Close()
}

func newStreamReader(ctx context.Context, l net.Listener) (*streamReader, error) {
	// Maybe we asked for an arbitrary port,
	// so, refresh our address to the one we’re actually listening on.
	laddr := l.Addr()

	host, path := laddr.String(), ""
	switch laddr.Network() {
	case "unix":
		host, path = "", host
	}

	uri := &url.URL{
		Scheme: laddr.Network(),
		Host:   host,
		Path:   path,
	}

	loading := make(chan struct{})
	r := &streamReader{
		Info: wrapper.NewInfo(uri, 0, time.Now()),

		loading: loading,
	}

	go func() {
		defer close(loading)
		defer l.Close()

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = &os.PathError{
				Op:   "accept",
				Path: uri.String(),
				Err:  ctx.Err(),
			}
			return
		}

		var conn net.Conn
		accept := func() error {
			var err error

			conn, err = l.Accept()

			return err
		}

		if err := do(ctx, accept); err != nil {
			r.err = &os.PathError{
				Op:   "accept",
				Path: uri.String(),
				Err:  err,
			}
			return
		}

		// TODO: make the a configurable option?
		/* if err := conn.CloseWrite(); err != nil {
			conn.Close()
			r.err = err
			return
		} */

		r.conn = conn
	}()

	return r, nil
}
