// Package socketfiles implements the "tcp:", "udp:", and "unix:" URL schemes.
package socketfiles

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strconv"
	"syscall"

	"golang.org/x/net/ipv4"
)

var (
	errInvalidURL = errors.New("invalid url")
	errInvalidIP  = errors.New("invalid ip")
)

// URL query field keys.
const (
	FieldBufferSize   = "buffer_size"
	FieldLocalAddress = "localaddr"
	FieldLocalPort    = "localport"
	FieldMaxBitrate   = "max_bitrate"
	FieldPacketSize   = "pkt_size"
	FieldTOS          = "tos"
	FieldTTL          = "ttl"
)

type socket struct {
	conn net.Conn

	addr, qaddr net.Addr

	bufferSize int
	packetSize int

	tos, ttl int

	throttler
}

func (s *socket) uri() *url.URL {
	q := s.uriQuery()

	switch qaddr := s.qaddr.(type) {
	case *net.TCPAddr:
		q.Set(FieldLocalAddress, qaddr.IP.String())
		q.Set(FieldLocalPort, strconv.Itoa(qaddr.Port))

	case *net.UDPAddr:
		q.Set(FieldLocalAddress, qaddr.IP.String())
		q.Set(FieldLocalPort, strconv.Itoa(qaddr.Port))

	case *net.UnixAddr:
		q.Set(FieldLocalAddress, qaddr.String())
	}

	host, path := s.addr.String(), ""

	switch s.addr.Network() {
	case "unix", "unixgram", "unixpacket":
		host, path = "", host
	}

	return &url.URL{
		Scheme:   s.addr.Network(),
		Host:     host,
		Path:     path,
		RawQuery: q.Encode(),
	}
}

func (s *socket) uriQuery() url.Values {
	q := make(url.Values)

	if s.bitrate > 0 {
		q.Set(FieldMaxBitrate, strconv.Itoa(s.bitrate))
	}

	if s.bufferSize > 0 {
		q.Set(FieldBufferSize, strconv.Itoa(s.bufferSize))
	}

	network := s.addr.Network()

	switch network {
	case "udp", "udp4", "udp6", "unixgram", "unixpacket":
		if s.packetSize > 0 {
			q.Set(FieldPacketSize, strconv.Itoa(s.packetSize))
		}
	}

	switch network {
	case "udp", "udp4", "tcp", "tcp4":
		if s.tos > 0 {
			q.Set(FieldTOS, "0x"+strconv.FormatInt(int64(s.tos), 16))
		}

		if s.ttl > 0 {
			q.Set(FieldTTL, strconv.Itoa(s.ttl))
		}
	}

	return q
}

func sockReader(conn net.Conn, q url.Values) (*socket, error) {
	bufferSize, err := getSize(q, FieldBufferSize)
	if err != nil {
		return nil, err
	}

	if bufferSize > 0 {
		type readBufferSetter interface {
			SetReadBuffer(int) error
		}

		conn, ok := conn.(readBufferSetter)
		if !ok {
			return nil, syscall.EINVAL
		}

		if err := conn.SetReadBuffer(bufferSize); err != nil {
			return nil, err
		}
	}

	return &socket{
		conn: conn,

		addr: conn.LocalAddr(),

		bufferSize: bufferSize,
	}, nil
}

func sockWriter(conn net.Conn, showLocalAddr bool, q url.Values) (*socket, error) {
	raddr := conn.RemoteAddr()

	bufferSize, err := getSize(q, FieldBufferSize)
	if err != nil {
		return nil, err
	}

	if bufferSize > 0 {
		type writeBufferSetter interface {
			SetWriteBuffer(int) error
		}

		conn, ok := conn.(writeBufferSetter)
		if !ok {
			return nil, syscall.EINVAL
		}

		if err := conn.SetWriteBuffer(bufferSize); err != nil {
			return nil, err
		}
	}

	var packetSize int
	switch raddr.Network() {
	case "udp", "udp4", "udp6", "unixgram", "unixpacket":
		packetSize, err = getSize(q, FieldPacketSize)
		if err != nil {
			return nil, err
		}
	}

	bitrate, err := getSize(q, FieldMaxBitrate)
	if err != nil {
		return nil, err
	}

	var t throttler
	if bitrate > 0 {
		t.setBitrate(bitrate, packetSize)
	}

	var tos, ttl int

	switch raddr.Network() {
	case "udp", "udp4", "tcp", "tcp4":
		var p *ipv4.Conn

		tos, err = getInt(q, FieldTOS)
		if err != nil {
			return nil, err
		}

		if tos > 0 {
			if p == nil {
				p = ipv4.NewConn(conn)
			}

			if err := p.SetTOS(tos); err != nil {
				return nil, err
			}

			tos, _ = p.TOS()
		}

		ttl, err = getInt(q, FieldTTL)
		if err != nil {
			return nil, err
		}

		if ttl > 0 {
			if p == nil {
				p = ipv4.NewConn(conn)
			}

			if err := p.SetTTL(ttl); err != nil {
				return nil, err
			}

			ttl, _ = p.TTL()
		}
	}

	var laddr net.Addr
	if showLocalAddr {
		laddr = conn.LocalAddr()
	}

	return &socket{
		conn: conn,

		addr:  raddr,
		qaddr: laddr,

		bufferSize: bufferSize,
		packetSize: packetSize,

		tos: tos,
		ttl: ttl,

		throttler: t,
	}, nil
}

var scales = map[byte]int{
	'G': 1000000000,
	'g': 1000000000,
	'M': 1000000,
	'm': 1000000,
	'K': 1000,
	'k': 1000,
}

func getSize(q url.Values, field string) (val int, err error) {
	value := q.Get(field)
	if value == "" {
		return 0, nil
	}

	suffix := value[len(value)-1]

	scale := 1
	if s := scales[suffix]; s > 0 {
		scale = s
		value = value[:len(value)-1]
	}

	i, err := strconv.ParseInt(value, 0, strconv.IntSize)
	if err != nil {
		return 0, err
	}

	return int(i) * scale, nil
}

func getInt(q url.Values, field string) (val int, err error) {
	value := q.Get(field)
	if value == "" {
		return 0, nil
	}

	i, err := strconv.ParseInt(value, 0, strconv.IntSize)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func do(ctx context.Context, fn func() error) error {
	done := make(chan struct{})

	var err error
	go func() {
		defer close(done)

		err = fn()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	return err
}
