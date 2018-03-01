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

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "tcp")
}

type writer struct {
	*net.TCPConn
	*wrapper.Info
}

func (w *writer) Sync() error { return nil }

const (
	FieldLocalAddress = "local_addr"
)

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, err
	}

	fixURL := *uri
	fixURL.Host = raddr.String()

	var laddr *net.TCPAddr

	q := uri.Query()
	if addr := q.Get(FieldLocalAddress); addr != "" {
		laddr, err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
		q.Set(FieldLocalAddress, laddr.String())
		fixURL.RawQuery = q.Encode()
	}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	if err := conn.CloseRead(); err != nil {
		return nil, err
	}

	w := &writer{
		TCPConn: conn,
		Info:    wrapper.NewInfo(&fixURL, 0, time.Now()),
	}

	return w, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
