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

	conn *net.TCPConn
	*wrapper.Info

	common

	delay time.Duration
	next  *time.Timer
}

func (w *TCPWriter) updateDelay() {
	if w.bitrate <= 0 {
		w.delay = 0
		w.next = nil
		return
	}

	// delay = nanoseconds per byte
	w.delay = (8 * time.Second) / time.Duration(w.bitrate)
	w.next = time.NewTimer(0)
}

func (w *TCPWriter) SetMaxBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bitrate

	w.bitrate = bitrate
	w.updateDelay()

	return prev
}

func (w *TCPWriter) Sync() error {
	return nil
}

func (w *TCPWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.Close()
}

func (w *TCPWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.next != nil {
		<-w.next.C
		w.next.Reset(time.Duration(len(b)) * w.delay)
	}

	return w.conn.Write(b)
}

func (w *TCPWriter) uri() *url.URL {
	q := w.common.uriQuery()

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
	w := new(TCPWriter)

	raddr, err := net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, err
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	var laddr *net.TCPAddr

	if port != "" || addr != "" {
		laddr = new(net.TCPAddr)

		laddr.IP, laddr.Port, err = buildAddr(addr, port)
		if err != nil {
			return nil, err
		}
	}

	w.conn, err = net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	w.common.setForWriter(w.conn, q)

	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, err
}

func (h *tcpHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
