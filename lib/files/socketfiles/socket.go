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

type ipSocket struct {
	laddr, raddr net.Addr

	bufferSize int

	tos, ttl int

	throttler
}

func (s *ipSocket) uriQuery() url.Values {
	q := make(url.Values)

	if s.bitrate > 0 {
		setInt(q, FieldMaxBitrate, s.bitrate)
	}

	if s.bufferSize > 0 {
		setInt(q, FieldBufferSize, s.bufferSize)
	}

	if s.tos > 0 {
		q.Set(FieldTOS, "0x"+strconv.FormatInt(int64(s.tos), 16))
	}

	if s.ttl > 0 {
		setInt(q, FieldTTL, s.ttl)
	}

	return q
}

func (s *ipSocket) setForReader(conn net.Conn, q url.Values) error {
	s.laddr = conn.LocalAddr()

	type bufferSizeSetter interface {
		SetReadBuffer(int) error
	}
	if bufferSize, ok, err := getSize(q, FieldBufferSize); ok || err != nil {
		if err != nil {
			return err
		}

		conn, ok := conn.(bufferSizeSetter)
		if !ok {
			return syscall.EINVAL
		}

		if err := conn.SetReadBuffer(bufferSize); err != nil {
			return err
		}

		s.bufferSize = bufferSize
	}

	return nil
}

func (s *ipSocket) setForWriter(conn net.Conn, q url.Values) error {
	s.laddr = conn.LocalAddr()
	s.raddr = conn.RemoteAddr()

	s.throttler.set(q)

	type bufferSizeSetter interface {
		SetWriteBuffer(int) error
	}
	if bufferSize, ok, err := getSize(q, FieldBufferSize); ok || err != nil {
		if err != nil {
			return err
		}

		conn, ok := conn.(bufferSizeSetter)
		if !ok {
			return syscall.EINVAL
		}

		if err := conn.SetWriteBuffer(bufferSize); err != nil {
			return err
		}

		s.bufferSize = bufferSize
	}

	var p *ipv4.Conn

	if tos, ok, err := getInt(q, FieldTOS); ok || err != nil {
		if err != nil {
			return err
		}

		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err := p.SetTOS(tos); err != nil {
			return err
		}

		s.tos, _ = p.TOS()
	}

	if ttl, ok, err := getInt(q, FieldTTL); ok || err != nil {
		if err != nil {
			return err
		}

		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err := p.SetTTL(ttl); err != nil {
			return err
		}

		s.ttl, _ = p.TTL()
	}

	return nil
}

func setInt(q url.Values, field string, val int) {
	q.Set(field, strconv.Itoa(val))
}

var scales = map[byte]int{
	'G': 1000000000,
	'g': 1000000000,
	'M': 1000000,
	'm': 1000000,
	'K': 1000,
	'k': 1000,
}

func getSize(q url.Values, field string) (val int, specified bool, err error) {
	s := q.Get(field)
	if s == "" {
		return 0, false, nil
	}

	suffix := s[len(s)-1]

	scale := 1
	if val, ok := scales[suffix]; ok {
		scale = val
		s = s[:len(s)-1]
	}

	i, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return 0, true, err
	}

	return int(i) * scale, true, nil
}

func getInt(q url.Values, field string) (val int, specified bool, err error) {
	s := q.Get(field)
	if s == "" {
		return 0, false, nil
	}

	i, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return 0, true, err
	}

	return int(i), true, nil
}

func buildAddr(addr, portString string) (ip net.IP, port int, err error) {
	if addr != "" {
		ip = net.ParseIP(addr)
	}

	if portString != "" {
		p, err := strconv.ParseInt(portString, 10, strconv.IntSize)
		if err != nil {
			return nil, 0, err
		}

		port = int(p)
	}

	return ip, port, nil
}

func withContext(ctx context.Context, fn func() error) (err error) {
	done := make(chan struct{})

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
