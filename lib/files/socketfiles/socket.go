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
	FieldBitrate      = "bitrate"
	FieldPacketSize   = "pkt_size"
	FieldTOS          = "tos"
	FieldTTL          = "ttl"
)

type common struct {
	laddr, raddr net.Addr

	bufferSize int
	bitrate    int

	tos, ttl int
}

func (c *common) uriQuery() url.Values {
	q := make(url.Values)

	if c.bitrate > 0 {
		setInt(q, FieldBitrate, c.bitrate)
	}

	if c.bufferSize > 0 {
		setInt(q, FieldBufferSize, c.bufferSize)
	}

	if c.tos > 0 {
		q.Set(FieldTOS, "0x"+strconv.FormatInt(int64(c.tos), 16))
	}

	if c.ttl > 0 {
		setInt(q, FieldTTL, c.ttl)
	}

	return q
}

func (c *common) setForWriter(conn net.Conn, q url.Values) error {
	c.laddr = conn.LocalAddr()
	c.raddr = conn.RemoteAddr()

	if bitrate, ok := getInt(q, FieldBitrate); ok {
		c.bitrate = bitrate
	}

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

		c.bufferSize = buffer_size
	}

	var p *ipv4.Conn

	if tos, ok := getInt(q, FieldTOS); ok {
		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err := p.SetTOS(tos); err != nil {
			return err
		}

		c.tos, _ = p.TOS()
	}

	if ttl, ok := getInt(q, FieldTTL); ok {
		if p == nil {
			p = ipv4.NewConn(conn)
		}

		if err := p.SetTTL(ttl); err != nil {
			return err
		}

		c.ttl, _ = p.TTL()
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
