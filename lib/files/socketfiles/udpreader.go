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

type udpReader struct {
	*wrapper.Info
	sock *ipSocket

	*net.UDPConn
}

func (r *udpReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
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

	sock, err := ipReader(conn, uri.Query())
	if err != nil {
		conn.Close()
		return nil, files.PathError("open", uri.String(), err)
	}

	uri = &url.URL{
		Scheme:   laddr.Network(),
		Host:     laddr.String(),
		RawQuery: sock.uriQuery().Encode(),
	}

	return &udpReader{
		Info: wrapper.NewInfo(uri, 0, time.Now()),
		sock: sock,

		UDPConn: conn,
	}, nil
}
