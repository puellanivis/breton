package socketfiles

import (
	"context"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type udpHandler struct{}

func init() {
	files.RegisterScheme(&udpHandler{}, "udp")
}

type UDPWriter struct {
	mu sync.Mutex

	conn *net.UDPConn
	*wrapper.Info
	ipSocket

	noerrs bool

	off int
	buf []byte
}

func (w *UDPWriter) IgnoreErrors(state bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.noerrs

	w.noerrs = state

	return prev
}

func (w *UDPWriter) err(err error) error {
	if w.noerrs && err != io.ErrShortWrite {
		return nil
	}

	return err
}

func (w *UDPWriter) SetPacketSize(size int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := len(w.buf)

	w.buf = nil
	if size > 0 {
		w.buf = make([]byte, size)
	}

	w.updateDelay(len(w.buf))

	return prev
}

func (w *UDPWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bitrate

	w.bitrate = bitrate
	w.updateDelay(len(w.buf))

	return prev
}

func (w *UDPWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.err(w.sync())
}

func (w *UDPWriter) sync() error {
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

func (w *UDPWriter) mustWrite(b []byte) (n int, err error) {
	w.throttle(0)

	n, err = w.conn.Write(b)
	if n != len(b) {
		if (w.noerrs && n > 0) || err == nil {
			err = io.ErrShortWrite
		}
	}

	return n, err
}

func (w *UDPWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.sync()

	if err := w.conn.Close(); err != nil {
		return err
	}

	return err
}

func (w *UDPWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		w.throttle(len(b))

		n, err = w.conn.Write(b)
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

func (w *UDPWriter) uri() *url.URL {
	q := w.ipSocket.uriQuery()

	if w.laddr != nil {
		laddr := w.laddr.(*net.UDPAddr)

		q.Set(FieldLocalAddress, laddr.IP.String())
		setInt(q, FieldLocalPort, laddr.Port)
	}

	if len(w.buf) > 0 {
		setInt(q, FieldPacketSize, len(w.buf))
	}

	return &url.URL{
		Scheme:   "udp",
		Host:     w.raddr.String(),
		RawQuery: q.Encode(),
	}
}

func (h *udpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	w := new(UDPWriter)

	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, err
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	var laddr *net.UDPAddr

	if port != "" || addr != "" {
		laddr = new(net.UDPAddr)

		laddr.IP, laddr.Port, err = buildAddr(addr, port)
		if err != nil {
			return nil, err
		}
	}

	w.conn, err = net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	w.ipSocket.setForWriter(w.conn, q)

	if pkt_size, ok := getInt(q, FieldPacketSize); ok {
		w.buf = make([]byte, pkt_size)
	}

	w.updateDelay(len(w.buf))
	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, err
}

func (h *udpHandler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	return nil, &os.PathError{"open", uri.String(), os.ErrInvalid}
}

func (h *udpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
