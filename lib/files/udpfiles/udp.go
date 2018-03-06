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

	conn *net.UDPConn
	*wrapper.Info

	noerrs bool

	raddr, laddr *net.UDPAddr
	tos int
	ttl int
	bufferSize int
	bitrate int

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
	if w.bitrate <= 0 {
		w.delay = 0
		return
	}

	// delay = nanoseconds per byte
	w.delay = (8 * time.Second) / time.Duration(w.bitrate)

	if len(w.buf) > 0 {
		// If we have a fixed packet size, then we can pre-calculate our delay.
		// delay = (nanoseconds per byte) * x bytes = just nanoseconds
		w.delay *= time.Duration(len(w.buf))
	}
}

func (w *Writer) SetWriteBuffer(size int) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bufferSize

	err := w.conn.SetWriteBuffer(size)
	if err == nil {
		w.bufferSize = size
	}

	return prev, err
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

	prev := w.bitrate

	w.bitrate = bitrate
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
	n, err = w.conn.Write(b)
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

	if err := w.conn.Close(); err != nil {
		return err
	}

	return err
}

func (w *Writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		n, err = w.conn.Write(b)

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
	FieldBufferSize   = "buffer_size"
	FieldLocalAddress = "localaddr"
	FieldLocalPort    = "localport"
	FieldMaxBitrate   = "max_bitrate"
	FieldPacketSize   = "pkt_size"
	FieldTOS          = "tos"
	FieldTTL          = "ttl"
)

func (w *Writer) uri() *url.URL {
	q := make(url.Values)

	if w.laddr != nil {
		q.Set(FieldLocalAddress, w.laddr.IP.String())
		setInt(q, FieldLocalPort, w.laddr.Port)
	}

	if w.bitrate > 0 {
		setInt(q, FieldMaxBitrate, w.bitrate)
	}

	if w.bufferSize > 0 {
		setInt(q, FieldBufferSize, w.bufferSize)
	}

	if len(w.buf) > 0 {
		setInt(q, FieldPacketSize, len(w.buf))
	}

	if w.tos > 0 {
		q.Set(FieldTOS, "0x" + strconv.FormatInt(int64(w.tos), 16))
	}

	if w.ttl > 0 {
		setInt(q, FieldTTL, w.ttl)
	}

	return &url.URL{
		Scheme: "udp",
		Host: w.raddr.String(),
		RawQuery: q.Encode(),
	}
}


func setInt(q url.Values, field string, val int) {
	q.Set(field, strconv.Itoa(val))
}

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
	var err error
	w := new(Writer)

	w.raddr, err = net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, err
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	if port != "" || addr != "" {
		w.laddr = new(net.UDPAddr)

		if addr != "" {
			w.laddr.IP = net.ParseIP(addr)
		}

		if port != "" {
			p, err := strconv.ParseInt(port, 0, strconv.IntSize)
			if err != nil {
				return nil, err
			}

			w.laddr.Port = int(p)
		}
	}

	w.conn, err = net.DialUDP("udp", w.laddr, w.raddr)
	if err != nil {
		return nil, err
	}

	w.laddr = w.conn.LocalAddr().(*net.UDPAddr)

	var p *ipv4.Conn

	if max_bitrate, ok := getInt(q, FieldMaxBitrate); ok {
		w.SetMaxBitrate(max_bitrate)
	}

	if tos, ok := getInt(q, FieldTOS); ok {
		if p == nil {
			p = ipv4.NewConn(w.conn)
		}

		if err2 := p.SetTOS(tos); err == nil {
			err = err2
		}

		w.tos, _ = p.TOS()
	}

	if ttl, ok := getInt(q, FieldTTL); ok {
		if p == nil {
			p = ipv4.NewConn(w.conn)
		}

		if err2 := p.SetTTL(ttl); err == nil {
			err = err2
		}

		w.ttl, _ = p.TTL()
	}

	if pkt_size, ok := getInt(q, FieldPacketSize); ok {
		w.SetPacketSize(pkt_size)
	}

	if buf_size, ok := getInt(q, FieldBufferSize); ok {
		if _, err2 := w.SetWriteBuffer(buf_size); err == nil {
			err = err2
		}
	}

	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, err
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	return nil, &os.PathError{"open", uri.String(), os.ErrInvalid}
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
