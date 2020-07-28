package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
)

type tcpHandler struct{}

func init() {
	files.RegisterScheme(tcpHandler{}, "tcp")
}

func (tcpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host == "" {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  files.ErrURLHostRequired,
		}
	}

	laddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	return newStreamReader(ctx, l)
}

func (tcpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host == "" {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  files.ErrURLHostRequired,
		}
	}

	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	q := uri.Query()

	var laddr *net.TCPAddr

	host := q.Get(FieldLocalAddress)
	port := q.Get(FieldLocalPort)
	if host != "" || port != "" {
		laddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, &os.PathError{
				Op:   "create",
				Path: uri.String(),
				Err:  err,
			}
		}
	}

	var conn *net.TCPConn
	dial := func() error {
		var err error

		conn, err = net.DialTCP("tcp", laddr, raddr)

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

	return newStreamWriter(ctx, sock), nil
}
