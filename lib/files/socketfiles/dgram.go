package socketfiles

import (
	"context"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files/wrapper"
)

type datagramWriter struct {
	*wrapper.Info

	mu sync.Mutex
	closed chan struct{}

	noerrs bool
	off int
	buf []byte

	sock *socket
}

func (w *datagramWriter) IgnoreErrors(state bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.noerrs

	w.noerrs = state

	return prev
}

func (w *datagramWriter) err(err error) error {
	if w.noerrs && err != io.ErrShortWrite {
		return nil
	}

	return err
}

func (w *datagramWriter) SetPacketSize(size int) int {
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

func (w *datagramWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.sock.setBitrate(bitrate, len(w.buf))
}

func (w *datagramWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.sync()
	return w.err(err)
}

func (w *datagramWriter) sync() (n int, err error) {
	if w.off < 1 {
		return 0, nil
	}

	// zero out the end of the buffer.
	b := w.buf[w.off:]
	for i := range b {
		b[i] = 0
	}

	w.off = 0
	return w.write(w.buf)
}

func (w *datagramWriter) write(b []byte) (n int, err error) {
	// We should have already prescaled the delay, so scale=1 here.
	w.sock.throttle(1)

	n, err = w.sock.conn.Write(b)
	if n != len(b) {
		if (w.noerrs && n > 0) || err == nil {
			err = io.ErrShortWrite
		}
	}

	return n, err
}

func (w *datagramWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	_, err := w.sync()

	if err2 := w.sock.conn.Close(); err == nil {
		err = err2
	}

	return err
}

func (w *datagramWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) < 1 {
		w.sock.throttle(len(b))

		n, err = w.sock.conn.Write(b)
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

		_, err2 := w.sync()
		if err = w.err(err2); err != nil {
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

		// skip the whole packet size, even if n2 < sz
		b = b[sz:]
	}

	if len(b) > 0 {
		w.off = copy(w.buf, b)
		n += w.off
	}

	return n, nil
}

func newDatagramWriter(ctx context.Context, sock *socket) *datagramWriter {
	var buf []byte
	if sock.packetSize > 0 {
		buf = make([]byte, sock.packetSize)
	}

	w := &datagramWriter{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		sock: sock,

		closed: make(chan struct{}),
		buf:    buf,
	}

	go func() {
		select {
		case <-w.closed:
		case <-ctx.Done():
			w.Close()
		}
	}()

	return w
}

type datagramReader struct {
	*wrapper.Info
	net.Conn
}

func (r *datagramReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func newDatagramReader(ctx context.Context, sock *socket) *datagramReader {
	return &datagramReader{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		Conn: sock.conn,
	}
}
