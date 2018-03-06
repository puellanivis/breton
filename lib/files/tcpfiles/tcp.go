// Package datafiles implements the "tcp:" URL scheme, but throws away all errors except short writes.
package tcpfiles

import (
	"context"
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
	files.RegisterScheme(&handler{}, "tcp")
}

type Writer struct {
	mu sync.Mutex

	conn *net.TCPConn
	*wrapper.Info

	raddr, laddr *net.TCPAddr
	tos int
	ttl int
	bufferSize int
	bitrate int

	delay time.Duration
}

func (w *Writer) updateDelay() {
	if w.bitrate <= 0 {
		w.delay = 0
		return
	}

	// delay = nanoseconds per byte
	w.delay = (8 * time.Second) / time.Duration(w.bitrate)
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

func (w *Writer) SetMaxBitrate(bitrate int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	prev := w.bitrate

	w.bitrate = bitrate
	w.updateDelay()

	return prev
}

func (w *Writer) Sync() error {
	return nil
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.Close()
}

func (w *Writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.conn.Write(b)

	if w.delay > 0 {
		// Avoid a multiplication if we don’t have to do it.
		time.Sleep(time.Duration(len(b)) * w.delay)
	}

	return n, err
}

const (
	FieldBufferSize   = "buffer_size"
	FieldLocalAddress = "localaddr"
	FieldLocalPort    = "localport"
	FieldMaxBitrate   = "max_bitrate"
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

	if w.tos > 0 {
		q.Set(FieldTOS, "0x" + strconv.FormatInt(int64(w.tos), 16))
	}

	if w.ttl > 0 {
		setInt(q, FieldTTL, w.ttl)
	}

	return &url.URL{
		Scheme: "tcp",
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

	w.raddr, err = net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return nil, err
	}

	q := uri.Query()

	port := q.Get(FieldLocalPort)
	addr := q.Get(FieldLocalAddress)

	if port != "" || addr != "" {
		w.laddr = new(net.TCPAddr)

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

	w.conn, err = net.DialTCP("tcp", w.laddr, w.raddr)
	if err != nil {
		return nil, err
	}

	w.laddr = w.conn.LocalAddr().(*net.TCPAddr)

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

	if buf_size, ok := getInt(q, FieldBufferSize); ok {
		if _, err2 := w.SetWriteBuffer(buf_size); err == nil {
			err = err2
		}
	}

	w.Info = wrapper.NewInfo(w.uri(), 0, time.Now())

	return w, err
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
}
