package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type tcpReader struct {
	*wrapper.Info

	loading <-chan struct{}

	err  error
	conn *net.TCPConn
}

func (r *tcpReader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.conn.Read(b)
}

func (r *tcpReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *tcpReader) Close() error {
	for range r.loading {
	}

	// Never connected, so just return nil.
	if r.conn == nil {
		return nil
	}

	// Ignore the r.err, as it is a request-scope error, and not relevant to closing.

	return r.conn.Close()
}

func (h *tcpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host == "" {
		return nil, files.PathError("open", uri.String(), errInvalidURL)
	}

	laddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	// Maybe we asked for an arbitrary port,
	// so, refresh our address to the one we’re actually listening on.
	laddr = l.Addr().(*net.TCPAddr)

	uri = &url.URL{
		Scheme: laddr.Network(),
		Host:   laddr.String(),
	}

	loading := make(chan struct{})
	r := &tcpReader{
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

		var conn *net.TCPConn
		accept := func() error {
			var err error

			conn, err = l.AcceptTCP()

			return err
		}

		if err := do(ctx, accept); err != nil {
			r.err = files.PathError("accept", uri.String(), err)
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
