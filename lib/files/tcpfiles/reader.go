// Package datafiles implements the "tcp:" URL scheme, by reading/writing to a raw tcp socket.
package udpfiles

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
	conn *net.TCPConn
	*wrapper.Info

	err error
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

	if r.err != nil {
		return r.err
	}

	return r.conn.Close()
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	laddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}

	// Maybe we asked for an arbitrary port,
	// so we need to update the uri’s Host value with the actual address from the listener.
	listenURL := *uri
	listenURL.Host = l.Addr().String()

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

		conn, err := l.AcceptTCP()
		if err != nil {
			r.err = err
			return
		}

		if err := conn.CloseWrite(); err != nil {
			conn.Close()
			r.err = err
			return
		}

		r.conn = conn
	}()

	return r, nil
}
