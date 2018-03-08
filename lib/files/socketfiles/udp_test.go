package socketfiles

import (
	"net"
	"testing"
)

func TestUDPName(t *testing.T) {
	w := &UDPWriter{
		ipSocket: ipSocket{
			laddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65535,
			},
			raddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 80,
			},
			bufferSize: 1024,
			ttl:        100,
			tos:        0x80,

			throttler: throttler{
				bitrate: 2048,
			},
		},
		buf: make([]byte, 188),
	}

	uri := w.uri()
	expected := "udp://127.0.0.1:80?buffer_size=1024&localaddr=127.0.0.1&localport=65535&max_bitrate=2048&pkt_size=188&tos=0x80&ttl=100"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

	w = &UDPWriter{
		ipSocket: ipSocket{
			laddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65534,
			},
			raddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 443,
			},
		},
	}

	uri = w.uri()
	expected = "udp://127.0.0.1:443?localaddr=127.0.0.1&localport=65534"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

}