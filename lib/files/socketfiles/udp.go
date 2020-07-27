package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
)

type udpHandler struct{}

func init() {
	files.RegisterScheme(udpHandler{}, "udp")
}

func (udpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host == "" {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  errInvalidURL,
		}
	}

	laddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	// Maybe we asked for an arbitrary port,
	// so, refresh our address to the one we’re actually listening on.
	laddr = conn.LocalAddr().(*net.UDPAddr)

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
}

func (udpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host == "" {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  errInvalidURL,
		}
	}

	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	q := uri.Query()

	var laddr *net.UDPAddr

	host := q.Get(FieldLocalAddress)
	port := q.Get(FieldLocalPort)
	if host != "" || port != "" {
		laddr, err = net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, &os.PathError{
				Op:   "create",
				Path: uri.String(),
				Err:  err,
			}
		}
	}

	var conn *net.UDPConn
	dial := func() error {
		var err error

		conn, err = net.DialUDP("udp", laddr, raddr)

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

	return newDatagramWriter(ctx, sock), nil
}
