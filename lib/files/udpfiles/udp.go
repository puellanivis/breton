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

	noerrs bool

	off int
	buf []byte
}

func (w *writer) IgnoreErrors(state bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.noerrs

	w.noerrs = state

	return prev
}

func (w *writer) err(err error) error {
	if w.noerrs {
		return nil
	}

	return err
}

func (w *writer) SetPacketSize(sz int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := len(w.buf)

	w.buf = make([]byte, sz)

	return prev
}

func (w *writer) Sync() error { return nil }

func (w *writer) mustWrite(b []byte) (n int, err error) {
	n, err = w.UDPConn.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	return n, err
}

func (w *writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		n, err = w.UDPConn.Write(b)
		return n, w.err(err)
	}

	if w.off > 0 {
		n = copy(w.buf[w.off:], b)
		w.off += n

		if w.off < len(w.buf) {
			// The full length of b was copied into buffer,
			// and we haven’t filled the buffer.
			// So, we’re done.
			return n, nil
		}

		w.off = 0
		b = b[n:]

		n2, err2 := w.mustWrite(w.buf)
		if err = w.err(err2); err != nil {
			if n2 > 0 {
				w.off = copy(w.buf, w.buf[n2:])
			}

			n -= len(w.buf) - n2
			if n < 0 {
				n = 0
			}

			return n, err
		}
	}

	sz := len(w.buf)

	for len(b) > sz {
		n2, err2 := w.mustWrite(b[:sz])
		n += n2

		if err = w.err(err2); err != nil {
			return n, err
		}

		// skip the whole packet size, even on a short write.
		b = b[sz:]
	}

	if len(b) > 0 {
		n2 := copy(w.buf, b)
		w.off += n2
		n += n2
	}

	return n, nil
}

const (
	FieldLocalAddress = "local_addr"
	FieldPacketSize = "pkt_size"
)

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, err
	}

	var laddr *net.UDPAddr

	q := uri.Query()
	if addr := q.Get(FieldLocalAddress); addr != "" {
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

	if pkt_size := q.Get(FieldPacketSize); pkt_size != "" {
		sz, err := strconv.ParseInt(pkt_size, 0, strconv.IntSize)
		if err != nil {
			return w, err
		}

		w.SetPacketSize(int(sz))
	}

	return w, nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	return nil, os.ErrInvalid
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
