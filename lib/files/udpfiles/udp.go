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
	"golang.org/x/net/ipv4"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "udp")
}

type Writer struct {
	mu sync.Mutex

	w *net.UDPConn
	*wrapper.Info

	noerrs bool

	br int
	delay time.Duration

	off int
	buf []byte
}

func (w *Writer) IgnoreErrors(state bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.noerrs

	w.noerrs = state

	return prev
}

func (w *Writer) err(err error) error {
	if w.noerrs && err != io.ErrShortWrite {
		return nil
	}

	return err
}

func (w *Writer) updateDelay() {
	if w.br <= 0 {
		w.delay = 0
		return
	}

	// delay = nanoseconds per byte
	w.delay = (8 * time.Second) / time.Duration(w.br)

	if len(w.buf) > 0 {
		// If we have a fixed packet size, then we can pre-calculate our delay.
		// delay = (nanoseconds per byte) * x bytes = just nanoseconds
		w.delay *= time.Duration(len(w.buf))
	}
}

func (w *Writer) SetPacketSize(size int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := len(w.buf)

	w.buf = nil
	if size > 0 {
		w.buf = make([]byte, size)
	}

	w.updateDelay()

	return prev
}

func (w *Writer) SetMaxBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.br

	w.br = bitrate
	w.updateDelay()

	return prev
}

func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.err(w.sync())
}

func (w *Writer) sync() error {
	if w.off < 1 {
		return nil
	}

	// zero out the end of the buffer.
	for i := w.off; i < len(w.buf); i++ {
		w.buf[i] = 0
	}

	w.off = 0
	_, err := w.mustWrite(w.buf)
	return err
}

func (w *Writer) mustWrite(b []byte) (n int, err error) {
	n, err = w.w.Write(b)
	if n != len(b) {
		if (w.noerrs && n > 0) || err == nil {
			err = io.ErrShortWrite
		}
	}

	// The time package defines that Sleep will immediately return if w.delay is less than or equal to zero.
	time.Sleep(w.delay)

	return n, err
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.sync()

	if err := w.w.Close(); err != nil {
		return err
	}

	return err
}

func (w *Writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		n, err = w.w.Write(b)

		if w.delay > 0 {
			// Avoid a multiplication if we don’t have to do it.
			time.Sleep(time.Duration(len(b)) * w.delay)
		}

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

			/*n -= len(w.buf) - n2
			if n < 0 {
				n = 0
			} */

			return n, err
		}
	}

	sz := len(w.buf)

	for len(b) >= sz {
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
	FieldMaxBitrate   = "max_bitrate"
	FieldLocalAddress = "localaddr"
	FieldBufferSize   = "buf_size"
	FieldPacketSize   = "pkt_size"
	FieldTOS          = "tos"
	FieldTTL          = "ttl"
)

func getInt(q url.Values, field string) (int, bool) {
	s := q.Get(field)
	if s == "" {
		return 0, false
	}

	i, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return 0, false
	}

	return int(i), true
}

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

	w := &Writer{
		w:    conn,
		Info: wrapper.NewInfo(uri, 0, time.Now()),
	}

	var p *ipv4.Conn

	if tos, ok := getInt(q, FieldTOS); ok {
		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err2 := p.SetTOS(tos); err == nil {
			err = err2
		}
	}

	if ttl, ok := getInt(q, FieldTTL); ok {
		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err2 := p.SetTTL(ttl); err == nil {
			err = err2
		}
	}

	if buf_size, ok := getInt(q, FieldBufferSize); ok {
		if err2 := conn.SetWriteBuffer(buf_size); err == nil {
			err = err2
		}
	}

	if pkt_size, ok := getInt(q, FieldPacketSize); ok {
		w.SetPacketSize(pkt_size)
	}

	if max_bitrate, ok := getInt(q, FieldMaxBitrate); ok {
		w.SetMaxBitrate(max_bitrate)
	}

	return w, err
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	return nil, &os.PathError{"open", uri.String(), os.ErrInvalid}
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
