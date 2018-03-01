// Package datafiles implements the "unix:" URL scheme, by reading/writing to a raw unix socket.
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
	*net.UnixConn
	*wrapper.Info
}

func (w *writer) Sync() error { return nil }

const (
	FieldLocalAddress = "local_addr"
)

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	raddr, err := net.ResolveUnixAddr("unix", uri.Path)
	if err != nil {
		return nil, err
	}

	fixURL := *uri
	fixURL.Path = raddr.String()

	var laddr *net.UnixAddr

	q := uri.Query()
	if addr := q.Get(FieldLocalAddress); addr != "" {
		laddr, err = net.ResolveUnixAddr("unix", addr)
		if err != nil {
			return nil, err
		}
		q.Set(FieldLocalAddress, laddr.String())
		fixURL.RawQuery = q.Encode()
	}

	conn, err := net.DialUnix("unix", laddr, raddr)
	if err != nil {
		return nil, err
	}

	if err := conn.CloseRead(); err != nil {
		conn.Close()
		return nil, err
	}

	w := &writer{
		UnixConn: conn,
		Info:    wrapper.NewInfo(&fixURL, 0, time.Now()),
	}

	return w, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{ "readdir", uri.String(), os.ErrInvalid }
}
