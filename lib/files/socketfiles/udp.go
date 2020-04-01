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
	files.RegisterScheme(&udpHandler{}, "udp")
}

func (h *udpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host == "" {
		return nil, files.PathError("open", uri.String(), errInvalidURL)
	}

	laddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	// Maybe we asked for an arbitrary port,
	// so, refresh our address to the one we’re actually listening on.
	laddr = conn.LocalAddr().(*net.UDPAddr)

	sock, err := sockReader(conn, uri.Query())
	if err != nil {
		conn.Close()
		return nil, files.PathError("open", uri.String(), err)
	}

	return newDatagramReader(ctx, sock), nil
}

func (h *udpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host == "" {
		return nil, files.PathError("create", uri.String(), errInvalidURL)
	}

	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	q := uri.Query()

	var laddr *net.UDPAddr

	host := q.Get(FieldLocalAddress)
	port := q.Get(FieldLocalPort)
	if host != "" || port != "" {
		laddr, err = net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, files.PathError("create", uri.String(), err)
		}
	}

	var conn *net.UDPConn
	dial := func() error {
		var err error

		conn, err = net.DialUDP("udp", laddr, raddr)

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

	return newDatagramWriter(ctx, sock), nil
}

func (h *udpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}
