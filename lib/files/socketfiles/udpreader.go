package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type udpReader struct {
	mu sync.Mutex

	conn *net.UDPConn
	*wrapper.Info
	ipSocket
}

func (r *udpReader) Read(b []byte) (n int, err error) {
	return r.conn.Read(b)
}

func (r *udpReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *udpReader) Close() error {
	return r.conn.Close()
}

func (r *udpReader) uri() *url.URL {
	q := w.ipSocket.uriQuery()

	return &url.URL{
		Scheme:   "udp",
		Host:     w.laddr.String(),
		RawQuery: q.Encode(),
	}
}

func (h *udpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	r := new(udpReader)

	laddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	q := uri.Query()

	r.conn, err = net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	if err := r.ipSocket.setForReader(r.conn, q); err != nil {
		r.conn.Close()
		return nil, &os.PathError{"open", uri.String(), err}
	}

	r.Info = wrapper.NewInfo(r.uri(), 0, time.Now())

	return r, nil
}
