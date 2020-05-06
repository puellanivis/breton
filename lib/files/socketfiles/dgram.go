package socketfiles

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files/wrapper"
)

type datagramWriter struct {
	*wrapper.Info

	mu     sync.Mutex
	closed chan struct{}

	noerrs bool

	buf []byte
	off int

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

	// Update filename.
	w.Info.SetName(w.sock.uri())

	return prev
}

func (w *datagramWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.sock.setBitrate(bitrate, len(w.buf))

	// Update filename.
	w.Info.SetName(w.sock.uri())

	return prev
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
	w := &datagramWriter{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		sock: sock,

		closed: make(chan struct{}),
	}

	w.SetPacketSize(sock.packetSize)

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
	sock *socket

	mu sync.Mutex

	buf []byte
	cnt int
}

// defaultMaxPacketSize is the maximum size of an IPv4 payload, and non-Jumbogram IPv6 payload.
// This is an overly safe default.
const defaultMaxPacketSize = 64 * 1024

func (r *datagramReader) SetPacketSize(size int) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	prev := len(r.buf)

	if size <= 0 {
		size = defaultMaxPacketSize
	}

	switch {
	case size <= len(r.buf):
		r.buf = r.buf[:size]

	default:
		r.buf = append(r.buf, make([]byte, size-len(r.buf))...)
	}

	if r.cnt > len(r.buf) {
		r.cnt = len(r.buf)
	}

	r.sock.maxPacketSize = len(r.buf)

	// Update filename.
	r.Info.SetName(r.sock.uri())

	return prev
}

func (r *datagramReader) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (r *datagramReader) Close() error {
	// Do not attempt to acquire the Mutex.
	// Doing so will deadlock with a concurrent blocking Read(),
	// and prevent read cancellation.
	return r.sock.conn.Close()
}

// ReadPacket reads a single packet from a data source.
// It is up to the caller to ensure that the given buffer is sufficient to read a full packet.
func (r *datagramReader) ReadPacket(b []byte) (n int, err error) {
	return r.sock.conn.Read(b)
}

// Read performs reads from a datagram source into a continuous stream.
//
// It does this by ensuring that each read on the datagram socket is to a sufficiently sized buffer.
// If the given buffer is too small, it will read to an internal buffer with length set from max_pkt_size,
// and following reads will read from that buffer until it is empty.
//
// Properly, a datagram source should know it is reading packets,
// and ensure each given buffer is large enough to read the maximum packet size expected.
// Unfortunately, some APIs in Go can expect Read()s to operate as a continuous stream instead of packets,
// and that a short read buffer, will just leave the rest of the unread data ready to read, not dropped on the floor.
func (r *datagramReader) Read(b []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cnt <= 0 {
		// Nothing is buffered.

		if len(b) >= len(r.buf) {
			// The read can be done directly.
			return r.ReadPacket(b)
		}

		// Given buffer is too small, use internal buffer.
		r.cnt, err = r.ReadPacket(r.buf)
	}

	n = copy(b, r.buf[:r.cnt])
	r.cnt = copy(r.buf, r.buf[n:r.cnt])
	return n, nil
}

func newDatagramReader(ctx context.Context, sock *socket) *datagramReader {
	r := &datagramReader{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		sock: sock,
	}

	r.SetPacketSize(sock.maxPacketSize)

	return r
}
