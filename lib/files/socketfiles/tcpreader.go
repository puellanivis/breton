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

type TCPReader struct {
	conn *net.TCPConn
	*wrapper.Info

	err     error
	loading <-chan struct{}
}

func (r *TCPReader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.conn.Read(b)
}

func (r *TCPReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *TCPReader) Close() error {
	for range r.loading {
	}

	if r.err != nil {
		return r.err
	}

	return r.conn.Close()
}

func (h *tcpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	laddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	// Maybe we asked for an arbitrary port,
	// so, we build our own copy of the URL, and use that.
	lurl := &url.URL{
		Host: l.Addr().String(),
	}

	loading := make(chan struct{})
	r := &TCPReader{
		loading: loading,
		Info:    wrapper.NewInfo(lurl, 0, time.Now()),
	}

	go func() {
		defer close(loading)
		defer l.Close()

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = &os.PathError{"open", uri.String(), ctx.Err()}
			return
		}

		conn, err := l.AcceptTCP()
		if err != nil {
			r.err = &os.PathError{"accept", uri.String(), err}
			return
		}

		/* if err := conn.CloseWrite(); err != nil {
			conn.Close()
			r.err = err
			return
		} */

		r.conn = conn
	}()

	return r, nil
}
