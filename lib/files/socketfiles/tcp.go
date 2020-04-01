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
	files.RegisterScheme(&tcpHandler{}, "tcp")
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

	return newStreamReader(ctx, l)
}

func (h *tcpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host == "" {
		return nil, files.PathError("create", uri.String(), errInvalidURL)
	}

	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	q := uri.Query()

	var laddr *net.TCPAddr

	host := q.Get(FieldLocalAddress)
	port := q.Get(FieldLocalPort)
	if host != "" || port != "" {
		laddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, files.PathError("create", uri.String(), err)
		}
	}

	var conn *net.TCPConn
	dial := func() error {
		var err error

		conn, err = net.DialTCP("tcp", laddr, raddr)

		return err
	}

	if err := do(ctx, dial); err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	sock, err := sockWriter(conn, laddr != nil, q)
	if err != nil {
		conn.Close()
		return nil, files.PathError("create", uri.String(), err)
	}

	return newStreamWriter(ctx, sock), nil
}

func (h *tcpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}
