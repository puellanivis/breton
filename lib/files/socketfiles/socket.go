// Package socketfiles implements the "tcp:", "udp:", and "unix:" URL schemes.
package socketfiles

import (
	"net"
	"net/url"
	"strconv"
	"syscall"

	"golang.org/x/net/ipv4"
)

const (
	FieldBufferSize   = "buffer_size"
	FieldLocalAddress = "localaddr"
	FieldLocalPort    = "localport"
	FieldBitrate      = "max_bitrate"
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
		setInt(q, FieldBitrate, s.bitrate)
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

func (s *ipSocket) setForWriter(conn net.Conn, q url.Values) error {
	s.laddr = conn.LocalAddr()
	s.raddr = conn.RemoteAddr()

	s.throttler.set(q)

	type bufferSizeSetter interface {
		SetWriteBuffer(int) error
	}
	if buffer_size, ok := getInt(q, FieldBufferSize); ok {
		conn, ok := conn.(bufferSizeSetter)
		if !ok {
			return syscall.EINVAL
		}

		if err := conn.SetWriteBuffer(buffer_size); err != nil {
			return err
		}

		s.bufferSize = buffer_size
	}

	var p *ipv4.Conn

	if tos, ok := getInt(q, FieldTOS); ok {
		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err := p.SetTOS(tos); err != nil {
			return err
		}

		s.tos, _ = p.TOS()
	}

	if ttl, ok := getInt(q, FieldTTL); ok {
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
