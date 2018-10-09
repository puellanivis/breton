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

type udpWriter struct {
	mu sync.Mutex

	closed chan struct{}

	conn *net.UDPConn
	*wrapper.Info
	ipSocket

	noerrs bool

	off int
	buf []byte
}

func (w *udpWriter) IgnoreErrors(state bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.noerrs

	w.noerrs = state

	return prev
}

func (w *udpWriter) err(err error) error {
	if w.noerrs && err != io.ErrShortWrite {
		return nil
	}

	return err
}

func (w *udpWriter) SetPacketSize(size int) int {
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

func (w *udpWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bitrate

	w.bitrate = bitrate
	w.updateDelay(len(w.buf))

	return prev
}

func (w *udpWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.err(w.sync())
}

func (w *udpWriter) sync() error {
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

func (w *udpWriter) mustWrite(b []byte) (n int, err error) {
	// We should have already prescaled the delay, so scale=1 here.
	w.throttle(1)

	n, err = w.conn.Write(b)
	if n != len(b) {
		if (w.noerrs && n > 0) || err == nil {
			err = io.ErrShortWrite
		}
	}

	return n, err
}

func (w *udpWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.sync()

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	if err2 := w.conn.Close(); err == nil {
		err = err2
	}

	return err
}

func (w *udpWriter) Write(b []byte) (n int, err error) {
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

func (w *udpWriter) uri() *url.URL {
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
	if uri.Host == "" {
		return nil, &os.PathError{"create", uri.String(), errInvalidURL}
	}

	w := &udpWriter{
		closed: make(chan struct{}),
	}

	raddr, err := net.ResolveUDPAddr("udp", uri.Host)
	if err != nil {
		return nil, &os.PathError{"create", uri.String(), err}
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	var laddr *net.UDPAddr

	if port != "" || addr != "" {
		laddr = new(net.UDPAddr)

		laddr.IP, laddr.Port, err = buildAddr(addr, port)
		if err != nil {
			return nil, &os.PathError{"create", uri.String(), err}
		}
	}

	dail := func() error {
		var err error

		w.conn, err = net.DialUDP("udp", laddr, raddr)

		return err
	}

	if err := withContext(ctx, dail); err != nil {
		return nil, &os.PathError{"create", uri.String(), err}
	}

	go func() {
		select {
		case <-w.closed:
		case <-ctx.Done():
			w.Close()
		}
	}()

	if err := w.ipSocket.setForWriter(w.conn, q); err != nil {
		w.Close()
		return nil, &os.PathError{"create", uri.String(), err}
	}

	if pktSize, ok, err := getSize(q, FieldPacketSize); ok || err != nil {
		if err != nil {
			w.Close()
			return nil, &os.PathError{"create", uri.String(), err}
		}

		w.buf = make([]byte, pktSize)
	}

	w.updateDelay(len(w.buf))
	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, nil
}

func (h *udpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
