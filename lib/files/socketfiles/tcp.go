package socketfiles

import (
	"context"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type tcpHandler struct{}

func init() {
	files.RegisterScheme(&tcpHandler{}, "tcp")
}

type TCPWriter struct {
	mu sync.Mutex

	closed chan struct{}

	conn *net.TCPConn
	*wrapper.Info
	ipSocket
}

func (w *TCPWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bitrate

	w.bitrate = bitrate
	w.updateDelay(1)

	return prev
}

func (w *TCPWriter) Sync() error {
	return nil
}

func (w *TCPWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	return w.conn.Close()
}

func (w *TCPWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.throttle(len(b))

	return w.conn.Write(b)
}

func (w *TCPWriter) uri() *url.URL {
	q := w.ipSocket.uriQuery()

	if w.laddr != nil {
		laddr := w.laddr.(*net.TCPAddr)

		q.Set(FieldLocalAddress, laddr.IP.String())
		setInt(q, FieldLocalPort, laddr.Port)
	}

	return &url.URL{
		Scheme:   "tcp",
		Host:     w.raddr.String(),
		RawQuery: q.Encode(),
	}
}

func (h *tcpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	w := &TCPWriter{
		closed: make(chan struct{}),
	}

	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, &os.PathError{"create", uri.String(), err}
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	var laddr *net.TCPAddr

	if port != "" || addr != "" {
		laddr = new(net.TCPAddr)

		laddr.IP, laddr.Port, err = buildAddr(addr, port)
		if err != nil {
			return nil, &os.PathError{"create", uri.String(), err}
		}
	}

	dail := func() error {
		var err error

		w.conn, err = net.DialTCP("tcp", laddr, raddr)

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

	w.updateDelay(1)
	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, nil
}

func (h *tcpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
