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
	q := r.ipSocket.uriQuery()

	return &url.URL{
		Scheme:   "udp",
		Host:     r.laddr.String(),
		RawQuery: q.Encode(),
	}
}

func (h *udpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host == "" {
		return nil, files.PathError("open", uri.String(), errInvalidURL)
	}

	r := new(udpReader)

	laddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	q := uri.Query()

	r.conn, err = net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	if err := r.ipSocket.setForReader(r.conn, q); err != nil {
		r.conn.Close()
		return nil, files.PathError("open", uri.String(), err)
	}

	r.Info = wrapper.NewInfo(r.uri(), 0, time.Now())

	return r, nil
}
