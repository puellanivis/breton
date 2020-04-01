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
	*wrapper.Info
	conn *net.UDPConn

	mu sync.Mutex

	closed chan struct{}

	sock *ipSocket

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

	switch {
	case size <= 0:
		w.buf = nil

	case size <= len(w.buf):
		w.buf = w.buf[:size]

	default:
		w.buf = append(w.buf, make([]byte, size-len(w.buf))...)
	}

	if w.off > len(w.buf) {
		w.off = len(w.buf)
	}

	w.sock.packetSize = len(w.buf)
	w.sock.updateDelay(len(w.buf))

	return prev
}

func (w *udpWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.sock.setBitrate(bitrate, len(w.buf))
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

	_, err := w.writeBuffer()
	return err
}

func (w *udpWriter) writeBuffer() (n int, err error) {
	w.off = 0
	return w.write(w.buf)
}

func (w *udpWriter) write(b []byte) (n int, err error) {
	// We should have already prescaled the delay, so scale=1 here.
	w.sock.throttle(1)

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

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	err := w.sync()

	if err2 := w.conn.Close(); err == nil {
		err = err2
	}

	return err
}

func (w *udpWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		w.sock.throttle(len(b))

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

		n2, err2 := w.writeBuffer()
		if err = w.err(err2); err != nil {
			if n2 > 0 {
				// Should we?
				// This could cause loss of packet-alignment from writers?
				w.off = copy(w.buf, w.buf[n2:])
			}

			return n, err
		}

		b = b[n:]
	}

	sz := len(w.buf)

	for len(b) >= sz {
		n2, err2 := w.write(b[:sz])
		n += n2

		if err = w.err(err2); err != nil {
			return n, err
		}

		// skip the whole packet size, even on a short write.
		b = b[sz:]
	}

	if len(b) > 0 {
		w.off = copy(w.buf, b)
		n += w.off
	}

	return n, nil
}

func (w *udpWriter) uri() *url.URL {
	return w.sock.uri()
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

	sock, err := ipWriter(conn, laddr != nil, q)
	if err != nil {
		conn.Close()
		return nil, files.PathError("create", uri.String(), err)
	}

	var buf []byte
	if sock.packetSize > 0 {
		buf = make([]byte, sock.packetSize)
	}

	w := &udpWriter{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		conn: conn,

		closed: make(chan struct{}),
		sock:   sock,

		buf: buf,
	}

	go func() {
		select {
		case <-w.closed:
		case <-ctx.Done():
			w.Close()
		}
	}()

	return w, nil
}

func (h *udpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}
