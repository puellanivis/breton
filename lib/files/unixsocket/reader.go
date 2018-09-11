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
	conn *net.UnixConn
	*wrapper.Info

	err     error
	loading <-chan struct{}
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

	// Ignore the r.err, as it is a request-scope error, and not relevant to closing.

	return r.conn.Close()
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	laddr, err := net.ResolveUnixAddr("unix", uri.Path)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenUnix("unix", laddr)
	if err != nil {
		return nil, err
	}

	// Maybe we asked for an arbitrary port,
	// so we need to update the uri’s Path value with the actual address from the listener.
	listenURL := *uri
	listenURL.Path = l.Addr().String()

	loading := make(chan struct{})
	r := &reader{
		loading: loading,
		Info:    wrapper.NewInfo(&listenURL, 0, time.Now()),
	}

	go func() {
		defer close(loading)
		defer l.Close()

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = ctx.Err()
			return
		}

		conn, err := l.AcceptUnix()
		if err != nil {
			r.err = err
			return
		}
		r.conn = conn
	}()

	return r, nil
}
