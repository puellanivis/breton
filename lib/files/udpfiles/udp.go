// Package datafiles implements the "udp:" URL scheme, but throws away all errors except short writes.
package udpfiles

import (
	"context"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "udp")
}

type writer struct {
	mu sync.Mutex

	*net.UDPConn
	*wrapper.Info

	off int
	buf []byte
}

func (w *writer) Sync() error { return nil }

func (w *writer) Write(b []byte) (n int, err error) {
	if len(w.buf) < 1 {
		n, _ = w.UDPConn.Write(b)
		return n, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.off > 0 {
		n2 := copy(w.buf[w.off:], b)
		w.off += n2
		n += n2

		if w.off < len(w.buf) {
			return n, nil
		}

		w.off = 0
		b = b[n2:]

		n3, _ := w.UDPConn.Write(w.buf)
		if n3 == 0 {
			// Drop the error.
			return n, nil
		}

		if n3 != len(w.buf) {
			w.off = copy(b, b[n3:])
			return n, io.ErrShortWrite
		}
	}

	for len(b) > len(w.buf) {
		n2, _ := w.UDPConn.Write(b[:len(w.buf)])
		n += n2

		if n2 != len(w.buf) {
			return n, io.ErrShortWrite
		}

		b = b[len(w.buf):]
	}

	n2 := copy(w.buf, b)
	w.off += n2
	n += n2

	return n, nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, err
	}

	var laddr *net.UDPAddr

	q := uri.Query()
	if addr := q.Get("local_addr"); addr != "" {
		laddr, err = net.ResolveUDPAddr("udp", addr)
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	w := &writer{
		UDPConn: conn,
		Info:    wrapper.NewInfo(uri, 0, time.Now()),
	}

	if pkt_size := q.Get("pkt_size"); pkt_size != "" {
		if sz, err := strconv.ParseInt(pkt_size, 0, strconv.IntSize); err == nil {
			w.buf = make([]byte, sz)
		}
	}

	return w, nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	return nil, os.ErrInvalid
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
