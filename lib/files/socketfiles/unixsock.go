package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
)

type unixHandler struct{}

func init() {
	files.RegisterScheme(unixHandler{}, "unix", "unixgram")
}

func (h unixHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, files.ErrURLCannotHaveAuthority
	}

	path := uri.Path
	if path == "" {
		var err error
		path, err = url.PathUnescape(uri.Opaque)
		if err != nil {
			return nil, files.ErrURLInvalid
		}
	}
	network := uri.Scheme

	laddr, err := net.ResolveUnixAddr(network, path)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	switch laddr.Network() {
	case "unixgram":
		conn, err := net.ListenUnixgram(network, laddr)
		if err != nil {
			return nil, &os.PathError{
				Op:   "open",
				Path: uri.String(),
				Err:  err,
			}
		}

		sock, err := sockReader(conn, uri.Query())
		if err != nil {
			conn.Close()
			return nil, &os.PathError{
				Op:   "open",
				Path: uri.String(),
				Err:  err,
			}
		}

		return newDatagramReader(ctx, sock), nil

	case "unix":
		l, err := net.ListenUnix(network, laddr)
		if err != nil {
			return nil, &os.PathError{
				Op:   "open",
				Path: uri.String(),
				Err:  err,
			}
		}

		return newStreamReader(ctx, l)
	}

	return nil, &os.PathError{
		Op:   "open",
		Path: uri.String(),
		Err:  files.ErrNotSupported,
	}
}

func (h unixHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, files.ErrURLCannotHaveAuthority
	}

	path := uri.Path
	if path == "" {
		var err error
		path, err = url.PathUnescape(uri.Opaque)
		if err != nil {
			return nil, files.ErrURLInvalid
		}
	}
	network := uri.Scheme

	raddr, err := net.ResolveUnixAddr(network, path)
	if err != nil {
		return nil, err
	}

	q := uri.Query()

	var laddr *net.UnixAddr

	addr := q.Get(FieldLocalAddress)
	if addr != "" {
		laddr, err = net.ResolveUnixAddr(network, addr)
		if err != nil {
			return nil, &os.PathError{
				Op:   "create",
				Path: uri.String(),
				Err:  err,
			}
		}
	}

	var conn *net.UnixConn
	dial := func() error {
		var err error

		conn, err = net.DialUnix(network, laddr, raddr)

		return err
	}

	if err := do(ctx, dial); err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	sock, err := sockWriter(conn, laddr != nil, q)
	if err != nil {
		conn.Close()
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	switch network {
	case "unix":
		return newStreamWriter(ctx, sock), nil

	case "unixgram", "unixpacket":
		return newDatagramWriter(ctx, sock), nil
	}

	conn.Close()
	return nil, &os.PathError{
		Op:   "create",
		Path: uri.String(),
		Err:  files.ErrNotSupported,
	}
}
