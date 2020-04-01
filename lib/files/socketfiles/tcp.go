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

type tcpWriter struct {
	*wrapper.Info
	conn *net.TCPConn

	mu sync.Mutex

	closed chan struct{}

	sock *ipSocket
}

func (w *tcpWriter) SetBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.sock.setBitrate(bitrate, 1)
}

func (w *tcpWriter) Sync() error {
	return nil
}

func (w *tcpWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.closed:
	default:
		close(w.closed)
	}

	return w.conn.Close()
}

func (w *tcpWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.sock.throttle(len(b))

	return w.conn.Write(b)
}

func (w *tcpWriter) uri() *url.URL {
	return w.sock.uri()
}

func (h *tcpHandler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	if uri.Host == "" {
		return nil, files.PathError("create", uri.String(), errInvalidURL)
	}

	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	q := uri.Query()

	var laddr *net.TCPAddr

	host := q.Get(FieldLocalAddress)
	port := q.Get(FieldLocalPort)
	if host != "" || port != "" {
		laddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, files.PathError("create", uri.String(), err)
		}
	}

	var conn *net.TCPConn
	dial := func() error {
		var err error

		conn, err = net.DialTCP("tcp", laddr, raddr)

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

	w := &tcpWriter{
		Info: wrapper.NewInfo(sock.uri(), 0, time.Now()),
		conn: conn,

		closed: make(chan struct{}),
		sock:   sock,
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

func (h *tcpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}
