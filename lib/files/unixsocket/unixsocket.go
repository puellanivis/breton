// Package unixsocket implements the "unix:" URL scheme, by reading/writing to a raw unix socket.
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

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "unix")
}

type writer struct {
	*wrapper.Info
	*net.UnixConn
}

func (w *writer) Sync() error { return nil }

// URL query field keys.
const (
	FieldLocalAddress = "local_addr"
)

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	raddr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}

	var laddr *net.UnixAddr

	q := uri.Query()
	if addr := q.Get(FieldLocalAddress); addr != "" {
		laddr, err = net.ResolveUnixAddr("unix", addr)
		if err != nil {
			return nil, err
		}
	}

	var conn *net.UnixConn
	dial := func() error {
		var err error

		conn, err = net.DialUnix("unix", laddr, raddr)

		return err
	}

	if err := do(ctx, dial); err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	q = make(url.Values)
	if laddr != nil {
		q.Set(FieldLocalAddress, laddr.String())
	}

	uri = &url.URL{
		Scheme:   raddr.Network(),
		Path:     raddr.String(),
		RawQuery: q.Encode(),
	}

	w := &writer{
		Info:     wrapper.NewInfo(uri, 0, time.Now()),
		UnixConn: conn,
	}

	return w, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}

func do(ctx context.Context, fn func() error) error {
	done := make(chan struct{})

	var err error
	go func() {
		defer close(done)

		err = fn()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	return err
}
