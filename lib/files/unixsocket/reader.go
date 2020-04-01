package unixsocket

import (
	"context"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type reader struct {
	*wrapper.Info

	loading <-chan struct{}

	err  error
	conn *net.UnixConn
}

func (r *reader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.conn.Read(b)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *reader) Close() error {
	for range r.loading {
	}

	// Never connected, so just return nil.
	if r.conn == nil {
		return nil
	}

	// Ignore the r.err, as it is a request-scope error, and not relevant to closing.

	return r.conn.Close()
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	laddr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenUnix("unix", laddr)
	if err != nil {
		return nil, err
	}

	// Make sure we are setting our file name to the actual address we’re listening on.
	laddr = l.Addr().(*net.UnixAddr)

	uri = &url.URL{
		Scheme: laddr.Network(),
		Path:   laddr.String(),
	}

	loading := make(chan struct{})
	r := &reader{
		loading: loading,
		Info:    wrapper.NewInfo(uri, 0, time.Now()),
	}

	go func() {
		defer close(loading)
		defer l.Close()

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = files.PathError("open", uri.String(), ctx.Err())
			return
		}

		var conn *net.UnixConn
		accept := func() error {
			var err error

			conn, err = l.AcceptUnix()

			return err
		}

		if err := do(ctx, accept); err != nil {
			r.err = files.PathError("accept", uri.String(), err)
			return
		}

		r.conn = conn
	}()

	return r, nil
}
